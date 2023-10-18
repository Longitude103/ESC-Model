package parcels

import (
	"bitbucket.org/longitude-103/wwumm-esc/database"
	"fmt"
)

func getSSParcels(v *database.Setup, yr int) []Parcel {
	query := fmt.Sprintf(`SELECT parcel_id, a.crop_int crop1, crop1_cov, b.crop_int crop2, crop2_cov, c.crop_int crop3, crop3_cov, d.crop_int crop4, crop4_cov, sw, gw, subarea, oa,
       irrig_type, sw_fac, first_irr, cert_num::varchar, model_id, sw_id, st_area(i.geom)/43560 area, 'np' nrd,
       st_x(st_transform(st_centroid(i.geom), 4326)) pointx, st_y(st_transform(st_centroid(i.geom), 4326)) pointy,
       sum(st_area(st_intersection(m.geom, i.geom))/43560) s_area, m.soil_code, m.coeff_zone
FROM np.t1953_irr i inner join public.model_cells m on st_intersects(i.geom, m.geom)
    LEFT join public.crops a on crop1 = a.crop_name
    LEFT join public.crops b on crop2 = b.crop_name
    LEFT join public.crops c on crop3 = c.crop_name
    LEFT join public.crops d on crop4 = d.crop_name
where m.cell_type = %d and sw = true
GROUP BY parcel_id, a.crop_int, parcel_id, crop1_cov, b.crop_int, crop2_cov, c.crop_int, crop3_cov, d.crop_int, crop4_cov, sw, gw, irrig_type, sw_fac, cert_num::varchar, model_id, st_area(i.geom)/43560, st_x(st_transform(st_centroid(i.geom), 4326)), st_y(st_transform(st_centroid(i.geom), 4326)), m.soil_code, crop1_cov, crop2, crop2_cov, crop3, crop3_cov, crop4, crop4_cov, sw, gw, subarea, oa, irrig_type,
    sw_fac, first_irr, cert_num::varchar, model_id, st_area(i.geom)/43560, st_x(st_transform(st_centroid(i.geom), 4326)),
    st_y(st_transform(st_centroid(i.geom), 4326)), nrd, m.soil_code, m.coeff_zone
UNION ALL
SELECT parcel_id, a.crop_int crop1, crop1_cov, b.crop_int crop2, crop2_cov, c.crop_int crop3, crop3_cov, d.crop_int crop4, crop4_cov, sw, gw, subarea, 0 oa,
       irr_type as irrig_type, sw_fac, first_irr, i.id as cert_num, null as model_id, sw_id, st_area(i.geom)/43560 area, 'sp' nrd,
       st_x(st_transform(st_centroid(i.geom), 4326)) pointx, st_y(st_transform(st_centroid(i.geom), 4326)) pointy,
       sum(st_area(st_intersection(m.geom, i.geom))/43560) s_area, m.soil_code, m.coeff_zone
FROM sp.t1953_irr i inner join public.model_cells m on st_intersects(i.geom, m.geom)
    LEFT join public.crops a on crop1 = a.crop_name
    LEFT join public.crops b on crop2 = b.crop_name
    LEFT join public.crops c on crop3 = c.crop_name
    LEFT join public.crops d on crop4 = d.crop_name
where m.cell_type = %d and sw = true
GROUP BY parcel_id, a.crop_int, parcel_id, crop1_cov, b.crop_int, crop2_cov, c.crop_int, crop3_cov, d.crop_int, crop4_cov, sw, gw, subarea, oa, irrig_type,
    sw_fac, first_irr, i.id, model_id, st_area(i.geom)/43560, st_x(st_transform(st_centroid(i.geom), 4326)),
    st_y(st_transform(st_centroid(i.geom), 4326)), nrd, m.soil_code, m.coeff_zone;`,
		v.CellType(), v.CellType())

	var parcels []Parcel
	err := v.PgDb.Select(&parcels, query)
	if err != nil {
		v.Logger.Errorf("Error in getting parcels for year %d, error: %s", yr, err)
	}

	for i := 0; i < len(parcels); i++ {
		parcels[i].Yr = yr
		parcels[i].ChangeFallow()
		parcels[i].NoCropCheck()
	}

	return parcels
}

func getDrySSParcels(v *database.Setup, yr int) []Parcel {
	query := fmt.Sprintf(`SELECT i.parcel_id, a.crop_int crop1, crop1_cov, b.crop_int crop2, crop2_cov, c.crop_int crop3, crop3_cov, d.crop_int crop4, crop4_cov,
       st_area(i.geom)/43560 area, 'np' nrd, st_x(st_transform(st_centroid(i.geom), 4326)) pointx,
       st_y(st_transform(st_centroid(i.geom), 4326)) pointy, sum(st_area(st_intersection(m.geom, i.geom))/43560) s_area,
       m.soil_code, m.coeff_zone
FROM np.t1953_dry i inner join (select geom, node, soil_code, zone, coeff_zone, mtg, nat_veg from model_cells where cell_type = %d) m on st_intersects(i.geom, m.geom)
    LEFT join public.crops a on crop1 = a.crop_name
    LEFT join public.crops b on crop2 = b.crop_name
    LEFT join public.crops c on crop3 = c.crop_name
    LEFT join public.crops d on crop4 = d.crop_name
GROUP BY i.parcel_id, a.crop_int, parcel_id, crop1_cov, b.crop_int, crop2_cov, c.crop_int, crop3_cov, d.crop_int, crop4_cov,
    st_area(i.geom)/43560, st_x(st_transform(st_centroid(i.geom), 4326)), st_y(st_transform(st_centroid(i.geom), 4326)),
    m.soil_code, crop1_cov, crop2, crop2_cov, crop3, crop3_cov, crop4, crop4_cov, st_area(i.geom)/43560,
    st_x(st_transform(st_centroid(i.geom), 4326)), st_y(st_transform(st_centroid(i.geom), 4326)), nrd, m.soil_code, m.coeff_zone
UNION ALL
SELECT i.parcel_id, a.crop_int crop1, crop1_cov, b.crop_int crop2, crop2_cov, c.crop_int crop3, crop3_cov, d.crop_int crop4, crop4_cov,
       st_area(i.geom)/43560 area, 'sp' nrd, st_x(st_transform(st_centroid(i.geom), 4326)) pointx,
       st_y(st_transform(st_centroid(i.geom), 4326)) pointy, sum(st_area(st_intersection(m.geom, i.geom))/43560) s_area,
       m.soil_code, m.coeff_zone
FROM sp.t1953_dry i inner join (select geom, node, soil_code, zone, coeff_zone, mtg, nat_veg from model_cells where cell_type = %d) m on st_intersects(i.geom, m.geom)
    LEFT join public.crops a on crop1 = a.crop_name
    LEFT join public.crops b on crop2 = b.crop_name
    LEFT join public.crops c on crop3 = c.crop_name
    LEFT join public.crops d on crop4 = d.crop_name
GROUP BY i.parcel_id, a.crop_int, parcel_id, crop1_cov, b.crop_int, crop2_cov, c.crop_int, crop3_cov, d.crop_int, crop4_cov,
    st_area(i.geom)/43560, st_x(st_transform(st_centroid(i.geom), 4326)), st_y(st_transform(st_centroid(i.geom), 4326)),
    m.soil_code, crop1_cov, crop2, crop2_cov, crop3, crop3_cov, crop4, crop4_cov, st_area(i.geom)/43560,
    st_x(st_transform(st_centroid(i.geom), 4326)), st_y(st_transform(st_centroid(i.geom), 4326)), nrd, m.soil_code, m.coeff_zone;`, v.CellType(), v.CellType())

	var parcels []Parcel
	err := v.PgDb.Select(&parcels, query)
	if err != nil {
		v.Logger.Errorf("Error in getting parcels for year %d, error: %s", yr, err)
	}

	for i := 0; i < len(parcels); i++ {
		parcels[i].Yr = yr
		parcels[i].ChangeFallow()
		parcels[i].NoCropCheck()
	}

	return parcels
}
