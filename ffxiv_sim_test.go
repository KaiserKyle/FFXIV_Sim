package main

import "testing"

func TestRates(t *testing.T) {
	critRate := calculateCritRate(364)
	directRate := calculateDirectRate(364)

	if critRate != 5.0 {
		t.Error(5.0, critRate)
	}
	if directRate != 0.0 {
		t.Error(0, directRate)
	}

	critRate = calculateCritRate(580)
	if critRate != 6.9 {
		t.Error(6.9, critRate)
	}

	critRate = calculateCritRate(581)
	if critRate != 7.0 {
		t.Error(7.0, critRate)
	}

	directRate = calculateDirectRate(612)
	if directRate != 6.2 {
		t.Error(6.2, directRate)
	}

	directRate = calculateDirectRate(613)
	if directRate != 6.3 {
		t.Error(6.3, directRate)
	}
}
