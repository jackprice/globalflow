package globalflow

import "testing"

func TestLamportClockIncrements(t *testing.T) {
	clock := NewClock()

	time1 := clock.Get()
	time2 := clock.Get()

	if time2 <= time1 {
		t.Errorf("expected time2 to be greater than time1")
	}
}
