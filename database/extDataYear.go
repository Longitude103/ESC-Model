package database

import (
	"database/sql"
	"fmt"
)

// lastExtDataYear returns the last year that has data in an external (outside the NRDs) table, ext_recharge or
// ext_pumping. Model years after it repeat that year's values, so adding newer rows to the table extends the model
// without a code change.
func lastExtDataYear(v *Setup, table string) (int, error) {
	if table != "ext_recharge" && table != "ext_pumping" {
		return 0, fmt.Errorf("unknown external data table %q", table)
	}

	var yr sql.NullInt64
	if err := v.PgDb.Get(&yr, fmt.Sprintf("select max(yr) from %s;", table)); err != nil {
		return 0, fmt.Errorf("error getting last year of %s: %w", table, err)
	}
	if !yr.Valid {
		return 0, fmt.Errorf("%s has no data", table)
	}

	return int(yr.Int64), nil
}
