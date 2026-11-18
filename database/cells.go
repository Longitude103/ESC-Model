package database

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

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

	// NpFrac and SpFrac are the shares of the cell's area inside the NP and SP NRDs; the rest is outside both.
	NpFrac float64
	SpFrac float64
	// NpRemoved and SpRemoved are the acres of NP and SP parcels that a post-97 or total depletion run takes out of
	// the cell. That land is natural vegetation in the run and is filed under the NRD of the parcel it replaced.
	NpRemoved float64
	SpRemoved float64
}

// Natural vegetation recharge file types.
const (
	NaturalVegNP    = 126
	NaturalVegSP    = 127
	NaturalVegOther = 102
)

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
	cells, err = getCellAcres(v, y, v.Post97, v.TotalDep)
	if err != nil {
		return nil, err
	}

	if v.Post97 || v.TotalDep {
		// the same year without parcels removed, so the land each removed parcel left behind can be filed under its NRD
		baseCells, err := getCellAcres(v, y, false, false)
		if err != nil {
			return nil, err
		}

		setRemovedAreas(cells, baseCells)
	}

	if err := setCellNRDs(v, cells); err != nil {
		return nil, err
	}

	// if v.AppDebug {
	// 	return cells[6600:6800], nil
	// }

	return cells, nil
}

func getCellAcres(v *Setup, y int, post97, totalDep bool) (cells []CellIntersect, err error) {
	query := fmt.Sprintf(`select node, soil_code, coeff_zone, mtg, cell_area, pointx, pointy, nip_area, ndp_area, 
							sip_area, sdp_area from getCellAcres(%d, %d, %t, %t);`, y, v.CellType(), post97, totalDep)

	if err = v.PgDb.Select(&cells, query); err != nil {
		return nil, err
	}

	return cells, nil
}

type cellKey struct {
	node, soil, zone int
}

// setRemovedAreas sets NpRemoved and SpRemoved on each scenario cell to the parcel acres in that cell in the base run
// that are gone from the scenario run.
func setRemovedAreas(cells, baseCells []CellIntersect) {
	base := make(map[cellKey]CellIntersect, len(baseCells))
	for _, c := range baseCells {
		base[cellKey{c.Node, c.Soil, c.CZone}] = c
	}

	for i := range cells {
		b, ok := base[cellKey{cells[i].Node, cells[i].Soil, cells[i].CZone}]
		if !ok {
			continue
		}

		cells[i].NpRemoved = math.Max(b.npArea()-cells[i].npArea(), 0)
		cells[i].SpRemoved = math.Max(b.spArea()-cells[i].spArea(), 0)
	}
}

func (c CellIntersect) npArea() float64 { return returnF64(c.NpIrrArea) + returnF64(c.NpDryArea) }
func (c CellIntersect) spArea() float64 { return returnF64(c.SpIrrArea) + returnF64(c.SpDryArea) }

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

// VegAreasByFileType splits the natural vegetation area of the cell (VegArea) into the NP, SP and other natural
// vegetation file types. Land that a post-97 or total depletion run took out of parcels goes to the NRD of the parcel,
// and the rest of the natural vegetation is split by the share of the cell's area in each NRD. Only file types with
// area are returned and the areas add up to VegArea.
func (c CellIntersect) VegAreasByFileType() map[int]float64 {
	total := c.VegArea()
	if total <= 0 {
		return nil
	}

	np, sp := normalizeNRDFractions(c.NpFrac, c.SpFrac)
	other := 1 - np - sp

	npRemoved, spRemoved := math.Max(c.NpRemoved, 0), math.Max(c.SpRemoved, 0)
	base := math.Max(total-npRemoved-spRemoved, 0) // natural vegetation that is there without any parcels removed

	areas := [3]float64{base*np + npRemoved, base*sp + spRemoved, base * other}
	sum := areas[0] + areas[1] + areas[2]

	result := make(map[int]float64, 3)
	for i, ft := range [3]int{NaturalVegNP, NaturalVegSP, NaturalVegOther} {
		if sum > 0 && areas[i] > 0 {
			result[ft] = areas[i] * total / sum
		}
	}

	return result
}

// normalizeNRDFractions keeps the NP and SP shares of a cell between 0 and 1 and scales them down if they add up to
// more than the whole cell, which overlapping zone polygons can cause.
func normalizeNRDFractions(np, sp float64) (float64, float64) {
	np = math.Min(math.Max(np, 0), 1)
	sp = math.Min(math.Max(sp, 0), 1)

	if s := np + sp; s > 1 {
		np, sp = np/s, sp/s
	}

	return np, sp
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

	return cells, setCellNRDs(v, cells)
}

// GetSSCellAreas2 is a function to return the amount of area within each model cell that is covered by parcels of surface water irrigated and
// dryland parcels. It returns the area, soil code, and zone of the cell in a slice of CellIntersect Struct.
func GetSSCellAreas2(v *Setup) (cells []CellIntersect, err error) {
	query := fmt.Sprintf(`select node, soil_code, coeff_zone, mtg, cell_area, pointx, pointy, nip_area, ndp_area, 
							sip_area, sdp_area from getcellss2acres(%d);`, v.CellType())

	if err = v.PgDb.Select(&cells, query); err != nil {
		return nil, err
	}

	return cells, setCellNRDs(v, cells)
}

// setCellNRDs fills in the share of each cell's area in the NP and SP NRDs so natural vegetation recharge can be split
// into the NP and SP file types.
func setCellNRDs(v *Setup, cells []CellIntersect) error {
	fracs, err := GetCellNRDFractions(v)
	if err != nil {
		return err
	}

	for i := range cells {
		f := fracs[cells[i].Node]
		cells[i].NpFrac, cells[i].SpFrac = f.NP, f.SP
	}

	return nil
}

// CellNRDFraction is the share of a model cell's area that lies in the NP and SP NRDs.
type CellNRDFraction struct {
	NP float64
	SP float64
}

type cellNRDArea struct {
	Node int     `db:"node"`
	Nrd  string  `db:"nrd"`
	Frac float64 `db:"frac"`
}

// GetCellNRDFractions returns the share of each model cell's area inside the NP and SP NRDs from the hsu_zones
// polygons. Cells that are not in any zone are left out (all of their area is outside both NRDs). The result is
// cached on the Setup for each cell type, because it does not change from year to year.
func GetCellNRDFractions(v *Setup) (map[int]CellNRDFraction, error) {
	ct := v.CellType()
	if f, ok := v.cellNRDFracs[ct]; ok {
		return f, nil
	}

	qry := fmt.Sprintf(`select node, nrd, sum(a) / max(cell_a) frac from (
    select c.node, lower(hsu.nrd) nrd, public.st_area(public.st_intersection(c.geom, hsu.geom)) a, 
           public.st_area(c.geom) cell_a
    from public.model_cells c inner join hsu_zones hsu on public.st_intersects(c.geom, hsu.geom) 
    where c.cell_type = %d) x where cell_a > 0 group by node, nrd;`, ct)

	var rows []cellNRDArea
	if err := v.PgDb.Select(&rows, qry); err != nil {
		return nil, err
	}

	fracs := cellNRDFractions(rows)

	if v.cellNRDFracs == nil {
		v.cellNRDFracs = make(map[int]map[int]CellNRDFraction)
	}
	v.cellNRDFracs[ct] = fracs

	return fracs, nil
}

func cellNRDFractions(rows []cellNRDArea) map[int]CellNRDFraction {
	fracs := make(map[int]CellNRDFraction)
	for _, r := range rows {
		f := fracs[r.Node]
		switch strings.ToLower(r.Nrd) {
		case "np":
			f.NP += r.Frac
		case "sp":
			f.SP += r.Frac
		default:
			continue
		}
		fracs[r.Node] = f
	}

	for n, f := range fracs {
		f.NP, f.SP = normalizeNRDFractions(f.NP, f.SP)
		fracs[n] = f
	}

	return fracs
}
