package qc

import (
	"github.com/Longitude103/ESC-Model/database"
)

type QC struct {
	v            *database.Setup
	fileName     string
	Graph        bool
	Year         int
	GJson        bool
	rechargeBal  bool
	Monthly      bool
	WellGJson    bool
	WellAnnGJson bool
	grid         int
	SYear        int
	EYear        int
}

type Option func(*QC)

func WithGraph() Option {
	return func(q *QC) { q.Graph = true }
}

func WithYear(year int) Option {
	return func(q *QC) { q.Year = year }
}

func WithGJson() Option {
	return func(q *QC) { q.GJson = true }
}

func WithMonthly() Option {
	return func(q *QC) { q.Monthly = true }
}

func WithWellGJson() Option {
	return func(q *QC) { q.WellGJson = true }
}

func WithWellAnnGJson() Option {
	return func(q *QC) { q.WellAnnGJson = true }
}

func WithRechargeBalance() Option {
	return func(q *QC) { q.rechargeBal = true }
}

func NewQC(v *database.Setup, fileName string, options ...Option) *QC {
	q := &QC{v: v, fileName: fileName, Year: 1997}
	for _, option := range options {
		option(q)
	}

	var err error
	q.grid, err = database.GetGrid(q.v.SlDb)
	if err != nil {
		panic(err)
	}

	q.SYear, q.EYear, err = database.GetStartEndYrs(q.v.SlDb)
	if err != nil {
		panic(err)
	}

	return q
}

func StartQcRMain(q *QC) error {
	//fmt.Printf("q: %+v\n", q)
	if q.rechargeBal {
		if err := q.rechargeBalance(); err != nil {
			return err
		}
	}

	if q.GJson {
		if err := q.rechargeGeoJson(); err != nil {
			return err
		}
	}

	if q.WellGJson {
		if err := q.WellPumpingGJson(); err != nil {
			return err
		}
	}

	if q.WellAnnGJson {
		if err := q.WellsAnnPumping(); err != nil {
			return err
		}
	}

	return nil
}

func (q *QC) getNodeDataFromSqlite(node int) {
	//q.v.SlDb.Select()
}
