package main

import "testing"

func TestMetricFormat(t *testing.T) {
	cases := map[int64]string{
		1000:    "1k",
		1234:    "1k",
		10023:   "10k",
		100000:  "100k",
		1000000: "1M",
		1023555: "1M",
		6012355: "6M",
	}
	for v, r := range cases {
		if cr := metricFormat(v); cr != r {
			t.Errorf("Metric format of number %d did not match '%s': '%s'", v, r, cr)
		}
	}
}
