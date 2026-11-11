package qc

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Longitude103/ESC-Model/Utils"
	"github.com/paulmach/orb/geojson"
	"github.com/pterm/pterm"
)

type rchResults struct {
	FileType int     `db:"file_type"`
	Desc     string  `db:"description"`
	Recharge float64 `db:"recharge"`
}

func (q *QC) rechargeBalance() error {
	var rchResultsSlice []rchResults

	spin, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Getting %d Data", q.Year))
	query := "select file_type, description, sum(result) recharge from results inner join file_keys fk on fk.file_key = results.file_type where strftime('%Y', dt) = "
	query += fmt.Sprintf("'%d' group by file_type, description;", q.Year)

	if err := q.v.SlDb.Select(&rchResultsSlice, query); err != nil {
		return err
	}
	d := pterm.TableData{{"File Type", "Description", "Recharge (AF)"}}
	total := 0.0

	for _, r := range rchResultsSlice {
		d = append(d, []string{fmt.Sprintf("%d", r.FileType), r.Desc, fmt.Sprintf("%.0f", r.Recharge)})
		total += r.Recharge
	}

	d = append(d, []string{"TOTAL", "All File Types", fmt.Sprintf("%.0f", total)})
	spin.Success()

	pterm.DefaultSection.Println("Recharge Summary")
	if err := pterm.DefaultTable.WithHasHeader().WithRightAlignment().WithData(d).Render(); err != nil {
		return err
	}

	return nil
}

type modelCells struct {
	Gjson []byte  `db:"geojson"`
	Ac    float64 `db:"area_ac"`
	Node  int     `db:"node"`
	Rc    [2]int
}

type resultData struct {
	Node int     `db:"node"`
	Rslt float64 `db:"rslt"`
}

type location struct {
	Node int `db:"node"`
	Rw   int `db:"rw"`
	Clm  int `db:"clm"`
}

var (
	//go:embed sql/recharge.sql
	rechargeQrys string
)

func (q *QC) rechargeGeoJson() error {
	formattedQueries := Utils.SplitQueries(rechargeQrys)

	grid, err := findGrid(q)
	if err != nil {
		return err
	}

	// get a slice of model cells in geojson
	pterm.DefaultSection.Println("GeoJSON Creation")
	spin, _ := pterm.DefaultSpinner.Start("Getting Model Cells from DB")
	var mCells []modelCells
	qry := fmt.Sprintf(formattedQueries[0], grid)

	if err := q.v.PgDb.Select(&mCells, qry); err != nil {
		return err
	}
	spin.Success()

	if grid == 1 || grid == 3 {
		var locationInfo []location
		locQuery := formattedQueries[2]

		if err := q.v.SlDb.Select(&locationInfo, locQuery); err != nil {
			return err
		}

		for i := 0; i < len(mCells); i++ {
			mCells[i].Rc = findRC(locationInfo, mCells[i].Node)
		}
	}

	spin, _ = pterm.DefaultSpinner.Start("Getting Result Data")
	rResMap := make(map[int][]resultData)
	for m := 1; m < 13; m++ {
		var rResults []resultData
		var mnString string
		if m < 10 {
			mnString = fmt.Sprintf("0%d", m)
		} else {
			mnString = fmt.Sprintf("%d", m)
		}

		rqry := fmt.Sprintf(formattedQueries[1], q.Year, mnString)
		if err := q.v.SlDb.Select(&rResults, rqry); err != nil {
			return err
		}

		rResMap[m] = rResults
	}
	spin.Success()

	fn := fmt.Sprintf("%d_Recharge.geojson", q.Year)
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

	// Follows format of https://datatracker.ietf.org/doc/html/rfc7946#section-1.5
	w := bufio.NewWriter(writeFile)

	header := `{"type":"FeatureCollection","features":[`

	_, _ = w.WriteString(header)
	cL := len(mCells)

	// unmarshal each item
	p, _ := pterm.DefaultProgressbar.WithTotal(cL).WithTitle("Model Cells").WithRemoveWhenDone(true).Start()
	for i := 0; i < cL; i++ {
		p.Increment()
		fc, err := geojson.UnmarshalFeature(mCells[i].Gjson)
		if err != nil {
			return err
		}

		// add property to them of the monthly result
		annTotal := 0.0
		for m := 1; m < 13; m++ {
			mn := time.Month(m)
			res := findResult(rResMap[m], mCells[i].Node)
			if q.Monthly {
				fc.Properties[mn.String()[:3]+"_AF"] = res
				fc.Properties[mn.String()[:3]+"_Ft/m"] = res / mCells[i].Ac
				fc.Properties[mn.String()[:3]+"_Ft/d"] = res / mCells[i].Ac / float64(Utils.TimeExt{T: time.Date(q.Year, mn, 1, 0, 0, 0, 0, time.UTC)}.DaysInMonth())
			}
			annTotal += res
		}

		fc.Properties["AnTl_AF"] = annTotal
		fc.Properties["AnTl_Ft/y"] = annTotal / mCells[i].Ac
		fc.Properties["AnTl_Ft/d"] = annTotal / mCells[i].Ac / 365.25

		if grid == 1 || grid == 3 {
			fc.Properties["Row"] = mCells[i].Rc[0]
			fc.Properties["Column"] = mCells[i].Rc[1]
		}

		// marshal that item back to json
		d, err := fc.MarshalJSON()
		if err != nil {
			return err
		}

		if _, err := w.WriteString(string(d)); err != nil {
			return err
		}

		if i < cL-1 {
			_, _ = w.WriteString(", ")
		}
	}

	_, _ = w.WriteString("]}")
	_ = w.Flush()
	_ = writeFile.Close()
	pterm.Success.Println("Check Output Files for GeoJson")

	return nil
}

func findResult(rData []resultData, node int) float64 {
	for i := 0; i < len(rData); i++ {
		if rData[i].Node == node {
			return rData[i].Rslt
		}
	}

	return 0
}

type Note struct {
	Nt string `db:"note"`
}

func findGrid(q *QC) (int, error) {
	var notes []Note
	notesQry := "select note from results_notes;"

	q.v.SlDb.Select(&notes, notesQry)

	for _, n := range notes {
		if n.Nt[:4] == "grid" {
			g := strings.Split(n.Nt, "=")[1]

			grid, err := strconv.Atoi(g)
			if err != nil {
				return 0, err
			}

			return grid, nil
		}
	}

	return 0, errors.New("couldn't find grid type")
}

func findRC(mcData []location, node int) (rcData [2]int) {
	for _, mc := range mcData {
		if mc.Node == node {
			return [2]int{mc.Rw, mc.Clm}
		}
	}

	return [2]int{0, 0}
}
