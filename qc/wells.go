package qc

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Longitude103/ESC-Model/Utils"
	"github.com/paulmach/orb/geojson"
	"github.com/pterm/pterm"
)

type Well struct {
	Gjson  []byte `db:"geojson"`
	Wellid int    `db:"wellid"`
	Nrd    string `db:"nrd"`
}

type NodeCentroid struct {
	Gjson []byte `db:"geojson"`
	Node  int    `db:"node"`
}

type WellData struct {
	Wellid   int     `db:"well_id"`
	FileType int     `db:"file_type"`
	Result   float64 `db:"result"`
	Node     int     `db:"cell_node"`
}

var (
	//go:embed sql/wellsQuery.sql
	wellsQuery string
	//go:embed sql/nodeQuery.sql
	nodeQuery string
)

func (q *QC) WellPumpingGJson() error {
	pterm.DefaultSection.Println("Well GeoJSON Creation")
	spin, _ := pterm.DefaultSpinner.Start("Getting Wells from DB")
	formattedQueries := Utils.SplitQueries(welQueries) // welQueries is in wellsAnn.go

	q.v.Logger.Info(fmt.Sprintf("Using Grid: %d", q.grid))
	var wlls []Well
	if err := q.v.PgDb.Select(&wlls, wellsQuery); err != nil {
		return err
	}

	var nodes []NodeCentroid
	if err := q.v.PgDb.Select(&nodes, nodeQuery, q.grid); err != nil {
		return nil
	}

	spin.Success()

	spin, _ = pterm.DefaultSpinner.Start("Getting Result Data")
	rResMap := make(map[int][]WellData)
	ssResMap := make(map[int][]WellData)
	for m := 1; m < 13; m++ {
		var rResults []WellData
		var ssResults []WellData
		var mnString string
		if m < 10 {
			mnString = fmt.Sprintf("0%d", m)
		} else {
			mnString = fmt.Sprintf("%d", m)
		}

		q.v.Logger.Info(fmt.Sprintf("Qry: %s, Year: %d, mnString: %s", formattedQueries[2], q.Year, mnString))
		rQuery := fmt.Sprintf(formattedQueries[2], q.Year, mnString)
		if err := q.v.SlDb.Select(&rResults, rQuery); err != nil {
			return err
		}
		ssQuery := fmt.Sprintf(formattedQueries[3], q.Year, mnString)
		if err := q.v.SlDb.Select(&ssResults, ssQuery); err != nil {
			return err
		}
		q.v.Logger.Info(fmt.Sprintf("rResults len: %d", len(rResults)))

		rResMap[m] = rResults
		ssResMap[m] = ssResults
	}

	fn := fmt.Sprintf("%d_Wells.geojson", q.Year)
	path := q.fileName

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		fmt.Printf("Error in mkdir: %s", err)
		return err
	}
	writeFile, err := os.Create(filepath.Join(path, fn))
	if err != nil {
		fmt.Printf("Error in create file: %s", err)
		return err
	}

	spin.Success()

	// add data to wells
	// Follows format of https://datatracker.ietf.org/doc/html/rfc7946#section-1.5
	w := bufio.NewWriter(writeFile)

	header := `{"type":"FeatureCollection","features":[`

	_, _ = w.WriteString(header)
	cL := len(wlls)
	q.v.Logger.Info(fmt.Sprintf("Number of Wells: %d", cL))

	p, _ := pterm.DefaultProgressbar.WithTotal(cL).WithTitle("Irr Wells").WithRemoveWhenDone(true).Start()
	firstWrittenRecord := true
	for i := 0; i < cL; i++ {
		p.Increment()

		fc, err := geojson.UnmarshalFeature(wlls[i].Gjson)
		if err != nil {
			return err
		}

		q.v.Logger.Info(fmt.Sprintf("%+v\n", fc))

		// add property to them of the monthly result
		annTotal := 0.0
		ft := 0
		for m := 1; m < 13; m++ {
			mn := time.Month(m)
			q.v.Logger.Info(fmt.Sprintf("Wells in rResMap[%d]: %d", m, len(rResMap[m])))
			res, mft := findWellResult(rResMap[m], wlls[i].Wellid, wlls[i].Nrd)
			if q.Monthly {
				fc.Properties[mn.String()[:3]+"_AF"] = res
				fc.Properties[mn.String()[:3]+"_cft/d"] = res / float64(Utils.TimeExt{T: time.Date(q.Year, mn, 1, 0, 0, 0, 0, time.UTC)}.DaysInMonth()) * 43560
			}
			annTotal += res
			if mft > 0 {
				ft = mft
			}
		}

		delete(fc.Properties, "nrd")
		fc.Properties["FileType"] = ft
		fc.Properties["AnTl_AF"] = annTotal
		fc.Properties["AnTl_cf/d"] = annTotal * 43560

		q.v.Logger.Info(fmt.Sprintf("Wellid: %d, AnnTotal: %f", wlls[i].Wellid, annTotal))
		if annTotal > 0.0 {
			// marshal that item back to json
			d, err := fc.MarshalJSON()
			if err != nil {
				return err
			}

			if !firstWrittenRecord {
				_, _ = w.WriteString(", ")
			}

			if _, err := w.WriteString(string(d)); err != nil {
				return err
			}

			firstWrittenRecord = false
		}

	}

	ssL := len(nodes)
	p, _ = pterm.DefaultProgressbar.WithTotal(ssL).WithTitle("SS Wells").WithRemoveWhenDone(true).Start()
	// loop through 209 to 218 file types for other wells
	for i := 0; i < ssL; i++ {
		p.Increment()

		fc, err := geojson.UnmarshalFeature(nodes[i].Gjson)
		if err != nil {
			return err
		}

		// add property to them of the monthly result
		annTotal := 0.0
		for m := 1; m < 13; m++ {
			mn := time.Month(m)
			res := findSSWellResult(ssResMap[m], nodes[i].Node)
			if q.Monthly {
				fc.Properties[mn.String()[:3]+"_AF"] = res
				fc.Properties[mn.String()[:3]+"_cft/d"] = res / float64(Utils.TimeExt{T: time.Date(q.Year, mn, 1, 0, 0, 0, 0, time.UTC)}.DaysInMonth()) * 43560
			}
			annTotal += res
		}

		delete(fc.Properties, "node")
		fc.Properties["wellid"] = i
		fc.Properties["FileType"] = 220
		fc.Properties["AnTl_AF"] = annTotal
		fc.Properties["AnTl_cf/d"] = annTotal * 43560

		if annTotal > 0.0 {
			// marshal that item back to json
			d, err := fc.MarshalJSON()
			if err != nil {
				return err
			}

			_, _ = w.WriteString(", ")
			if _, err := w.WriteString(string(d)); err != nil {
				return err
			}

			firstWrittenRecord = false
		}

	}

	_, _ = w.WriteString("]}")
	_ = w.Flush()
	_ = writeFile.Close()
	pterm.Success.Println("Check Output Files for GeoJson")

	return nil
}

func findWellResult(wellData []WellData, wellid int, nrd string) (float64, int) {
	if nrd == "sp" {
		for _, wd := range wellData {
			if wd.Wellid == wellid && wd.FileType > 204 {
				return wd.Result, wd.FileType
			}
		}
	} else {
		for _, wd := range wellData {
			if wd.Wellid == wellid && wd.FileType < 205 {
				return wd.Result, wd.FileType
			}
		}
	}

	return 0.0, 0
}

func findSSWellResult(ssData []WellData, node int) float64 {
	for _, d := range ssData {
		if d.Node == node {
			return d.Result
		}
	}

	return 0.0
}
