package bot

import (
	"testing"
	"time"
)

func TestLookupIntervalFitsInsideRefresh(t *testing.T) {
	for _, n := range []int{0, 1, 100, 2700, 10000, 54000, 200000} {
		d := lookupInterval(n)
		if d < minLookupInterval || d > maxLookupInterval {
			t.Errorf("lookupInterval(%d) = %s, outside bounds", n, d)
		}
		// Up to the rate cap, every sweep must finish inside the refresh interval.
		if n <= int(sweepBudget/minLookupInterval) && d*time.Duration(n) >= refreshInterval {
			t.Errorf("lookupInterval(%d) = %s, sweep takes %s", n, d, d*time.Duration(n))
		}
	}
}
