package main

import (
	"fmt"
	"math"
)

func metricFormat(in int64) string {
	siUnits := []string{"k", "M", "G", "T", "P", "E"}
	for i := len(siUnits) - 1; i >= 0; i-- {
		p := int64(math.Pow(1000, float64(i+1)))
		if in >= p {
			return fmt.Sprintf("%d%s", in/p, siUnits[i])
		}
	}
	return fmt.Sprintf("%d", in)
}
