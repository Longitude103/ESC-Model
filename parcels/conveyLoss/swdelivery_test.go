package conveyLoss

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func Test_GetSurfaceWaterDelivery(t *testing.T) {
	v := dbConnection(t)
	v.SYear = 1997
	v.EYear = 2020

	div, err := GetSurfaceWaterDelivery(v)
	if err != nil {
		t.Errorf("Error Getting SW Delivery: %s", err)
	}

	for _, d := range div[v.SYear] {
		if d.CanalId == 43 {
			v.Logger.Debug(d.printString())
		}
	}
}

func Test_GetSurfaceWaterDeliverySS(t *testing.T) {
	v := dbConnection(t)
	v.SYear = 1895
	v.EYear = 1905
	v.SteadyState = true

	div, err := GetSurfaceWaterDelivery(v)
	if err != nil {
		t.Errorf("Error Getting SW Delivery: %s", err)
	}

	for _, d := range div[v.SYear] {
		v.Logger.Debug(d.printString())
	}
}

func Test_filterCnl(t *testing.T) {
	var c1 = Canal{Id: 1, Name: "Canal", Eff: 0.5, Area: sql.NullFloat64{Float64: 100.25, Valid: true}, Yr: 1997}
	var c2 = Canal{Id: 2, Name: "Canal", Eff: 0.6, Area: sql.NullFloat64{Float64: 125.35, Valid: true}, Yr: 1997}
	var c3 = Canal{Id: 2, Name: "Canal", Eff: 0.7, Area: sql.NullFloat64{Float64: 501.11, Valid: true}, Yr: 1998}
	var c4 = Canal{Id: 1, Name: "Canal", Eff: 0.8, Area: sql.NullFloat64{Float64: 785.28, Valid: true}, Yr: 1999}
	var c5 = Canal{Id: 1, Name: "Canal", Eff: 0.4, Area: sql.NullFloat64{Float64: 532.81, Valid: true}, Yr: 2000}

	var canalSlice = []Canal{c1, c2, c3, c4, c5}

	c := filterCnl(canalSlice, 1, 1997)
	if c.Id != 1 || c.Yr != 1997 {
		t.Errorf("Should have returned canal with Id:1 and Yr:1997 but got: %+v", c)
	}
	//
	//c = filterCnl(canalSlice, 5, 1997)
	//fmt.Println(c)
}

func Test_FilterSWDeliveryByYear(t *testing.T) {
	d1 := Diversion{CanalId: 1, DivDate: sql.NullTime{Time: time.Now(), Valid: true}, DivAmount: sql.NullFloat64{Float64: 100.0, Valid: true}}
	d2 := Diversion{CanalId: 2, DivDate: sql.NullTime{Time: time.Now(), Valid: true}, DivAmount: sql.NullFloat64{Float64: 100.0, Valid: true}}
	d3 := Diversion{CanalId: 1, DivDate: sql.NullTime{Time: time.Now().Add(time.Hour * 24 * 365), Valid: true}, DivAmount: sql.NullFloat64{Float64: 100.0, Valid: true}}
	d4 := Diversion{CanalId: 3, DivDate: sql.NullTime{Time: time.Now().Add(time.Hour * 24 * 365), Valid: true}, DivAmount: sql.NullFloat64{Float64: 100.0, Valid: true}}

	divSlice := Diversions{d1, d2, d3, d4}

	d := divSlice.FilterSWDeliveryByYear(time.Now().Year())

	for _, diversion := range d {
		fmt.Printf("Divs: %+v\n", diversion)
	}

	for _, diversion := range d {
		if diversion.DivDate.Time.Year() != time.Now().Year() {
			t.Error("Diversions didn't filter correctly")
		}
	}

}
