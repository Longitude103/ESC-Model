package qc

import (
	"github.com/Longitude103/ESC-Model/Utils"
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/paulmach/orb/geojson"
	"github.com/pterm/pterm"
)

type Node struct {
	CellNode int `db:"cell_node"`
}

type NodeData struct {
	GJson []byte `db:"geojson"`
	Node  int    `db:"node"`
}

type GroupedResult struct {
	Node   int     `db:"cell_node"`
	Result float64 `db:"result"`
}

var (
	//go:embed sql/nodeQuery.sql
	nodeQrySubString string
	//go:embed sql/welQueries.sql
	welQueries string
)

func (q *QC) WellsAnnPumping() error {
	// get all the nodes that there is output data for from sqlite in any year
	p, _ := pterm.DefaultSpinner.Start("Getting data")
	formattedQueries := Utils.SplitQueries(welQueries)

	var Nodes []Node
	if err := q.v.SlDb.Select(&Nodes, formattedQueries[0]); err != nil {
		return err
	}

	// get the centroid location of all nodes from postgis as geojson
	q.v.Logger.Info(fmt.Sprintf("Using grid: %d", q.grid))
	nodeQry := fmt.Sprintf(nodeQrySubString, q.grid)
	var NodeLocs []NodeData
	if err := q.v.PgDb.Select(&NodeLocs, nodeQry); err != nil {
		return err
	}

	var mapgResults = make(map[int][]GroupedResult)
	// make a map here for each year
	for i := q.SYear; i < q.EYear+1; i++ {
		// get annual amount of pumping at each node in sqlite
		groupResultsQry := fmt.Sprintf(formattedQueries[1], i)
		var gResults []GroupedResult
		if err := q.v.SlDb.Select(&gResults, groupResultsQry); err != nil {
			return err
		}

		mapgResults[i] = gResults
	}
	p.Success()

	pterm.Info.Print("Writing AllWells.geojson file")
	// attribute the nodes with pumping data and save geojson file
	fn := "AllWells.geojson"
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

	pbar, _ := pterm.DefaultProgressbar.WithTotal(len(Nodes)).WithTitle("Populating All Wells").WithRemoveWhenDone(true).Start()
	firstWrittenRecord := true
	for i := 0; i < len(Nodes); i++ {
		pbar.Increment()
		gj := findGeoJson(NodeLocs, Nodes[i].CellNode)
		fc, err := geojson.UnmarshalFeature(gj)
		if err != nil {
			return err
		}

		for y := q.SYear; y < q.EYear+1; y++ {
			yKey := fmt.Sprintf("%d Ann AF", y)
			r := findGResult(mapgResults[y], Nodes[i].CellNode)

			fc.Properties[yKey] = r
		}

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

	_, _ = w.WriteString("]}")
	_ = w.Flush()
	_ = writeFile.Close()
	pterm.Success.Println("Check Output Files for AllWells.geojson")

	return nil
}

func findGeoJson(nd []NodeData, node int) []byte {
	for _, n := range nd {
		if node == n.Node {
			return n.GJson
		}
	}

	return nil
}

func findGResult(gr []GroupedResult, node int) float64 {
	for _, g := range gr {
		if node == g.Node {
			return g.Result
		}
	}

	return 0.0
}
