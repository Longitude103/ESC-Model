package database

import (
	"database/sql"
	"fmt"
	"math"
)

// IrrCell is a struct to hold the data of each cell and parcel intersect, it includes the cert, crops, and other characteristics
// important to the calculations.
type IrrCell struct {
	Node     int             `db:"node"`
	CellArea float64         `db:"c_area"`
	IrrArea  float64         `db:"i_area"`
	ParcelId int             `db:"parcel_id"`
	Nrd      string          `db:"nrd"`
	Mtg      sql.NullFloat64 `db:"mtg"`
	Soil     int             `db:"soil_code"`
	CZone    int             `db:"coeff_zone"`
}

// GetCellsIrr gets the cells that have irrigation within them and splits them by parcel. If a cell has multiple parcels
// there will be multiples of the same cell listed. This includes both nrd irrigated acres.
func GetCellsIrr(v *Setup, yr int) ([]IrrCell, error) {
	if v.SteadyState {
		return GetSSCellsIrr(v)
	}

	if yr > 1997 {
		if v.Post97 {
			return GetCellsIrrPost97(v, yr)
		}
	}

	query := fmt.Sprintf(`SELECT node, mtg, c.soil_code, c.coeff_zone, st_area(c.geom)/43560 c_area, st_area(st_intersection(c.geom, i.geom))/43560 i_area, 
									parcel_id, nrd from public.model_cells c inner join (SELECT parcel_id, 'np' nrd, geom from np.t%d_irr UNION SELECT parcel_id, 'sp' nrd, geom from sp.t%d_irr) i
        on st_intersects(c.geom, i.geom) where c.cell_type = %d;`, yr, yr, v.CellType())

	var irrCells []IrrCell
	if err := v.PgDb.Select(&irrCells, query); err != nil {
		v.Logger.Errorf("Cannot Get Cells Parcel Split data: %s", err)
		return nil, err
	}

	if v.AppDebug && len(irrCells) > 100 {
		return irrCells[:100], nil
	}

	return irrCells, nil
}

func (i IrrCell) GetLossFactor() float64 {
	if !i.Mtg.Valid || i.Mtg.Float64 == 0 {
		return 0.5
	}

	return math.Min(1-math.Exp(-0.02*i.Mtg.Float64), 1)
}

// GetCellsIrrPost97 a Post 97 version of this function that gets the cells that have irrigation within them and
// splits them by parcel. If a cell has multiple parcels there will be multiples of the same cell listed.
// This includes both nrd irrigated acres. It limits to the actual parcels during that year without GWO, then goes
// to 1997 for GWO parcels, combines them and sends it off.
func GetCellsIrrPost97(v *Setup, yr int) ([]IrrCell, error) {
	query := fmt.Sprintf(`SELECT node, mtg, c.soil_code, c.coeff_zone, st_area(c.geom)/43560 c_area, st_area(st_intersection(c.geom, i.geom))/43560 i_area,
       parcel_id, nrd
from public.model_cells c
    inner join (SELECT parcel_id, 'np' nrd, geom
    from np.t%d_irr where (sw = true and gw = false) or (sw = true and gw = true)
    UNION
    SELECT parcel_id, 'sp' nrd, geom
    from sp.t%d_irr where (sw = true and gw = false) or (sw = true and gw = true)) i
    on st_intersects(c.geom, i.geom) where c.cell_type = %d order by parcel_id;`, yr, yr, v.CellType())

	var irrCells []IrrCell
	if err := v.PgDb.Select(&irrCells, query); err != nil {
		v.Logger.Errorf("Cannot Get Cells Parcel Split of Non-GWO parcel data in post97: %s", err)
		return nil, err
	}

	query97GWO := fmt.Sprintf(`SELECT node, mtg, c.soil_code, c.coeff_zone, st_area(c.geom)/43560 c_area, st_area(st_intersection(c.geom, i.geom))/43560 i_area,
       parcel_id + 30000 as parcel_id, nrd
from public.model_cells c
    inner join (SELECT parcel_id, 'np' nrd, geom
    from np.t1997_irr where sw = false and gw = true
    UNION
    SELECT parcel_id, 'sp' nrd, geom
    from sp.t1997_irr where sw = false and gw = true) i
    on st_intersects(c.geom, i.geom) where c.cell_type = %d order by parcel_id;`, v.CellType())

	var p97irrCells []IrrCell
	if err := v.PgDb.Select(&p97irrCells, query97GWO); err != nil {
		v.Logger.Errorf("Cannot Get Cells Parcel Split of GWO 1997 Parcels data: %s", err)
		return nil, err
	}

	irrCells = append(irrCells, p97irrCells...)

	return irrCells, nil
}

// GetSSCellsIrr is a function to return the irrigated cells for 1953 only where there are surface water parcels. Any GW only parcels
// are not included in this function.
func GetSSCellsIrr(v *Setup) ([]IrrCell, error) {
	query := fmt.Sprintf(`SELECT node, mtg, c.soil_code, c.coeff_zone, st_area(c.geom)/43560 c_area, st_area(st_intersection(c.geom, i.geom))/43560 i_area, 
									parcel_id, nrd from public.model_cells c inner join (SELECT parcel_id, 'np' nrd, 
									geom from np.t1953_irr where sw = true UNION SELECT parcel_id, 'sp' nrd, geom from sp.t1953_irr where sw = true) i
        on st_intersects(c.geom, i.geom) where c.cell_type = %d;`, v.CellType())

	var irrCells []IrrCell
	if err := v.PgDb.Select(&irrCells, query); err != nil {
		v.Logger.Errorf("Cannot Get Cells Parcel Split data in Steady State query: %s", err)
		return nil, err
	}

	return irrCells, nil
}
