package parcels

import (
	"errors"
	"fmt"

	"bitbucket.org/longitude-103/wwumm-esc/database"
)

// EstimatePumping is a method that is called on parcels that shouldEstimate == true so that we can estimate
// the amount of pumping that was done at the parcel since a well is present, but not metered. It fills the Pump field of the Parcel struct
func (p *Parcel) EstimatePumping(v *database.Setup, cCrops []database.CoeffCrop) error {
	if se, err := p.shouldEstimate(v.Post97); err != nil || se {
		// nirAdj, err := adjustmentFactor(p, cCrops, database.NirEt)
		// if err != nil {
		// 	v.Logger.Errorf("Error in Parcel: %d Adjustment Factor: %s", p.ParcelNo, err)
		// }

		var swAvailableCU, nirRemaining [12]float64
		if p.IsSW() {
			for i := 0; i < 12; i++ {
				swAvailableCU[i] = p.SWDel[i] * p.AppEff
			}
		}

		// set nirRemaining to nir - swAvailableCU if positive, then divide by AppEff to arrive at pumping
		for m := 0; m < 12; m++ {
			nirRemaining[m] = p.Nir[m] - swAvailableCU[m]
			if nirRemaining[m] > 0 {
				// p.Pump[m] = nirRemaining[m] / p.AppEff
				p.Pump[m] = nirRemaining[m]
			}
		}

		return nil
	}

	// shouldn't estimate
	return nil
}

// shouldEstimate is a method that determines if the parcel should estimate pumping or if the pumping should not be estimated
// as it will have a pumping value assigned from the data.
func (p *Parcel) shouldEstimate(p97 bool) (bool, error) {
	if p97 {
		if p.IsGWO() {
			return true, nil
		}
	}

	if p.Nrd == "np" { // NPNRD parcel
		if p.Yr > 2016 {
			return false, nil
		}

		if p.Yr > 2008 {
			// have reads for OA areas
			if p.Oa.Valid {
				if p.Oa.Int64 == 1 {
					if p.Subarea.Valid {
						if p.Subarea.String == "North Platte" || p.Subarea.String == "Pumpkin Creek" {
							// areas not FA
							return false, nil
						}
					}
				}
			}

			// outside the OA Subareas
			return true, nil
		} else {
			// no reads, estimate
			return true, nil
		}
	}

	if p.Nrd == "sp" {
		// spnrd
		if p.Yr > 2009 {
			return false, nil
		} else {
			if len(p.Subarea.String) > 0 {
				if p.Yr > 2008 && (p.Subarea.String[2:] != "FA" || p.Subarea.String[3:] != "SPV") {
					return false, nil
				} else {
					return true, nil
				}
			}
		}
	}

	return true, errors.New("couldn't find if it should estimate pumping, will estimate")
}

// AdjustmentFactor function calculates the Parcel adjustment factor by weighting the crops and distribution of the
// crops in a Parcel by calling the nirFactor and then weighting it based on crop distribution
func AdjustmentFactor(p *Parcel, cCrops []database.CoeffCrop, adj database.Adjustment) (float64, error) {
	var (
		c2, c3, c4 float64
	)
	c1, err := AdjFactor(cCrops, p.CoeffZone, int(p.Crop1.Int64), adj)

	if p.Crop2.Valid {
		c2, err = AdjFactor(cCrops, p.CoeffZone, int(p.Crop2.Int64), adj)
	}
	if p.Crop3.Valid {
		c3, err = AdjFactor(cCrops, p.CoeffZone, int(p.Crop3.Int64), adj)
	}

	if p.Crop4.Valid {
		c4, err = AdjFactor(cCrops, p.CoeffZone, int(p.Crop4.Int64), adj)
	}
	if err != nil {
		return 0, err
	}

	return c1*p.Crop1Cov.Float64 + c2*p.Crop2Cov.Float64 + c3*p.Crop3Cov.Float64 + c4*p.Crop4Cov.Float64, nil
}

// AdjFactor is a filter function that returns the NirAdjFactor from the CoeffCrop slice and limits it to the zone of the
// Parcel and the crop type.
func AdjFactor(cCrops []database.CoeffCrop, zone int, crop int, adj database.Adjustment) (nf float64, err error) {
	for _, v := range cCrops {
		if v.Zone == zone && v.Crop == crop {
			switch adj {
			case database.NirEt:
				nf = v.NirAdjFactor
				return nf, nil
			case database.DryET:
				nf = v.DryEtAdj
				return nf, nil
			case database.IrrEt:
				nf = v.IrrEtAdj
				return nf, nil
			}
		}

		if crop == 15 {
			for _, v := range cCrops {
				if v.Zone == zone && v.Crop == 7 {
					switch adj {
					case database.NirEt:
						nf = v.NirAdjFactor
						return nf, nil
					case database.DryET:
						nf = v.DryEtAdj
						return nf, nil
					case database.IrrEt:
						nf = v.IrrEtAdj
						return nf, nil
					}
				}
			}
		}
	}

	errorText := fmt.Sprintf("zone %d and crop %d not found", zone, crop)
	return 0, errors.New(errorText)
}
