package basic

import "testing"

func TestArith(t *testing.T) {
	if result, err := Mul(3, 2); err != nil || result != 6 {
		t.Error("Mul(3,2) expected 6, got: ", result)
	}
	if result, err := Mul(-5, 7); err != nil || result != -35 {
		t.Error("Mul(-5,7) expected -35, got:", result)
	}
}
