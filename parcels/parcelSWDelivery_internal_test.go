package parcels

import (
	"database/sql"
	"math"
	"testing"
	"time"

	"github.com/Longitude103/ESC-Model/parcels/conveyLoss"
)

func div(canal int, month time.Month, afPerAcre float64) conveyLoss.Diversion {
	return conveyLoss.Diversion{CanalId: canal,
		DivDate:   sql.NullTime{Valid: true, Time: time.Date(2020, month, 1, 0, 0, 0, 0, time.UTC)},
		DivAmount: sql.NullFloat64{Valid: true, Float64: afPerAcre}}
}

// A 100 acre parcel split 60/40 across two soils must receive 100 acres worth of delivery in total, not 200.
func TestParcelSWDelivery_SplitBySoilArea(t *testing.T) {
	divs := []conveyLoss.Diversion{div(7, time.July, 0.5), div(7, time.July, 0.25), div(8, time.July, 9)}
	rows := []Parcel{
		{ParcelNo: 1, Area: 100, SoilArea: 60, SoilCode: 412, SwID: sql.NullInt64{Valid: true, Int64: 7}},
		{ParcelNo: 1, Area: 100, SoilArea: 40, SoilCode: 622, SwID: sql.NullInt64{Valid: true, Int64: 7}},
	}

	total := 0.0
	for i := range rows {
		rows[i].parcelSWDelivery(divs)
		total += rows[i].SWDel[6]
	}

	// two July records for canal 7 are both counted: (0.5 + 0.25) af/ac * 100 ac
	if math.Abs(total-75) > 1e-9 {
		t.Errorf("total July delivery = %f, want 75", total)
	}

	if math.Abs(rows[0].SWDel[6]-45) > 1e-9 {
		t.Errorf("60 acre soil row July delivery = %f, want 45", rows[0].SWDel[6])
	}
}
