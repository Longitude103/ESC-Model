package wells

import (
	"fmt"
	"testing"
)

func TestSSWells(t *testing.T) {
	sR := &sqlResults{}

	if err := SteadyStateWells(dbConnection(t), sR); err != nil {
		t.Error("Error in function")
	}

	for i, d := range sR.data {
		fmt.Printf("%+v\n", d)
		if i > 10 {
			break
		}
	}
}
