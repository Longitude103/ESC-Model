package database

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"

	"bitbucket.org/longitude-103/gisUtils"
)

type ModelCell struct {
	Node      int     `db:"node"`
	SoilCode  int     `db:"soil_code"`
	CoeffZone int     `db:"coeff_zone"`
	Zone      int     `db:"zone"`
	Mtg       float64 `db:"mtg"`
	PointX    float64 `db:"pointx"`
	PointY    float64 `db:"pointy"`
	Rw        int     `db:"rw"`
	Clm       int     `db:"clm"`
}

type CellIntersect struct {
	Node      int             `db:"node"`
	Soil      int             `db:"soil_code"`
	CZone     int             `db:"coeff_zone"`
	CellArea  float64         `db:"cell_area"`
	Mtg       sql.NullFloat64 `db:"mtg"`
	NpIrrArea sql.NullFloat64 `db:"nip_area"`
	NpDryArea sql.NullFloat64 `db:"ndp_area"`
	SpIrrArea sql.NullFloat64 `db:"sip_area"`
	SpDryArea sql.NullFloat64 `db:"sdp_area"`
	PointX    float64         `db:"pointx"`
	PointY    float64         `db:"pointy"`
}

type StDistances struct {
	Station  string
	Distance float64
	Weight   float64
}

// GetXY is a method of ModelCell to return the XY Coordinates of the model cell.
func (m ModelCell) GetXY() (x float64, y float64) {
	return m.PointX, m.PointY
}

// GetNodeRC is a method of ModelCell to return the node, rw, clm of the model cell.
func (m ModelCell) GetNodeRC() (node, rw, clm int) {
	return m.Node, m.Rw, m.Clm
}

// GetCells is a function to retrieve the model cells from the database and return a struct of ModelCell. It also handles
// debug mode to only return a slice of 50 cells.
func GetCells(v *Setup) (cells []ModelCell, err error) {

	//query := fmt.Sprintf(`select node, st_x(st_transform(st_centroid(geom), 4326)) pointx,
	//			st_y(st_transform(st_centroid(geom), 4326)) pointy,
	//			soil_code, coeff_zone, zone, mtg from public.model_cells where cell_type = %d;`, v.CellType())

	query := "select * from noderc()"

	if err = v.PgDb.Select(&cells, query); err != nil {
		return nil, err
	}

	return
}

// GetCellAreas is a function to return the amount of area within each model cell that is covered by parcels of irrigated and
// dryland. It also returns the area, soil code, and zone of the cell in a slice of CellIntersect Struct. It implements the
// debug mode to only return 200 cells which were selected as having good data.
func GetCellAreas(v *Setup, y int) (cells []CellIntersect, err error) {
	query := fmt.Sprintf(`select node, soil_code, coeff_zone, mtg, cell_area, pointx, pointy, nip_area, ndp_area, 
							sip_area, sdp_area from getCellAcres(%d, %d);`, y, v.CellType())

	if err = v.PgDb.Select(&cells, query); err != nil {
		return nil, err
	}

	// if v.AppDebug {
	// 	return cells[6600:6800], nil
	// }

	return cells, nil
}

// XyPoints is an interface that uses the GetXY method and is used by the Distances function to enable different structs
// to be able to input to the Distances function.
type XyPoints interface {
	GetXY() (x float64, y float64)
}

// Distances is a function that that returns the top three weather stations from the list with the appropriate weighting
// factor. Used to make CSResults Distribution.
func Distances(points XyPoints, wStations []WeatherStation) (dist []StDistances, err error) {
	var lengths []float64
	for _, v := range wStations {
		var stDistance StDistances
		pX, pY := points.GetXY()
		d := gisUtils.Distance(pY, pX, v.PointY, v.PointX)
		lengths = append(lengths, d)
		stDistance.Distance = d
		stDistance.Station = v.Code
		dist = append(dist, stDistance)
	}

	sort.Slice(dist, func(i, j int) bool {
		return dist[i].Distance < dist[j].Distance
	})

	sort.Float64s(lengths)

	var idw []float64
	if len(lengths) >= 3 {
		idw, err = gisUtils.InverseDW(lengths[:3])
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("less then three stations present")
	}

	for i, v := range idw {
		dist[i].Weight = v
	}

	return dist[:3], nil
}

// VegArea is a method of the CellIntersect struct that returns the total area that isn't covered by a parcel (dry or irr)
// of a cell and returns an area. Checks for cropArea > cell area as small parcel overlaps can make this negative which cannot
// be introduced to equations.
func (c CellIntersect) VegArea() float64 {
	cropArea := returnF64(c.NpIrrArea) + returnF64(c.SpIrrArea) + returnF64(c.NpDryArea) + returnF64(c.SpDryArea)

	if cropArea > c.CellArea {
		return 0
	} else {
		return c.CellArea - cropArea
	}
}

// GetXY is a method of CellIntersect struct that returns the XY locations for use in the Distances function and is required
// by the XyPoints interface.
func (c CellIntersect) GetXY() (x float64, y float64) {
	return c.PointX, c.PointY
}

// returnF64 is a simple function that is used by the VegArea method to return a float64 value from a sql.NullFloat64 type
// and if it's invalid, then returns a zero.
func returnF64(v sql.NullFloat64) float64 {
	if v.Valid {
		return v.Float64
	}
	return 0.0
}

func (c CellIntersect) GetLossFactor() float64 {
	if !c.Mtg.Valid || c.Mtg.Float64 == 0 {
		return 0.5
	}

	return math.Min(1-math.Exp(-0.02*c.Mtg.Float64), 1)
}

func AddCellsToOutput(v *Setup) error {
	c, err := CellRCDB(v.SlDb)
	if err != nil {
		return err
	}

	mc, err := GetCells(v)
	if err != nil {
		return err
	}

	for _, m := range mc {
		n, r, clm := m.GetNodeRC()
		if err := c.Add(Cellrc{n, r, clm}); err != nil {
			return err
		}
	}

	if err := c.Flush(); err != nil {
		return err
	}

	return nil
}

// GetSSCellAreas1 is a function for the Steady State Run to return the amount of area within each model cell for the first two stress periods. NO parcels are included.
// It returns the area, soil code, and zone of the cell in a slice of CellIntersect Struct.
func GetSSCellAreas1(v *Setup) (cells []CellIntersect, err error) {
	query := fmt.Sprintf(`select node, soil_code, coeff_zone, mtg, cell_area, pointx, pointy from getcellss1acres(%d);`, v.CellType())

	if err = v.PgDb.Select(&cells, query); err != nil {
		return nil, err
	}

	return cells, nil
}

// GetSSCellAreas2 is a function to return the amount of area within each model cell that is covered by parcels of surface water irrigated and
// dryland parcels. It returns the area, soil code, and zone of the cell in a slice of CellIntersect Struct.
func GetSSCellAreas2(v *Setup) (cells []CellIntersect, err error) {
	query := fmt.Sprintf(`select node, soil_code, coeff_zone, mtg, cell_area, pointx, pointy, nip_area, ndp_area, 
							sip_area, sdp_area from getcellss2acres(%d);`, v.CellType())

	if err = v.PgDb.Select(&cells, query); err != nil {
		return nil, err
	}

	return cells, nil
}
