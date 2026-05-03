package testdata

import "testing"

func TestNoErrorPattern(t *testing.T) {
	err := doSomething()
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}
}

func TestErrorPattern(t *testing.T) {
	err := doSomething()
	if err == nil {
		t.Fatal("Expected an error")
	}
}

func TestEqualPattern(t *testing.T) {
	result := calculate()
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

func TestNotEqualPattern(t *testing.T) {
	result := calculate()
	if result == 0 {
		t.Error("Result should not be zero")
	}
}

func TestTruePattern(t *testing.T) {
	condition := checkCondition()
	if !condition {
		t.Error("Expected condition to be true")
	}
}

func TestFalsePattern(t *testing.T) {
	condition := checkCondition()
	if condition {
		t.Error("Expected condition to be false")
	}
}

func TestNotNilPattern(t *testing.T) {
	ptr := getPointer()
	if ptr == nil {
		t.Fatal("Expected non-nil pointer")
	}
}

func TestNilPattern(t *testing.T) {
	ptr := getPointer()
	if ptr != nil {
		t.Error("Expected nil pointer")
	}
}

func TestLenPattern(t *testing.T) {
	items := getItems()
	if len(items) != 5 {
		t.Errorf("Expected 5 items, got %d", len(items))
	}
}

func TestGreaterPattern(t *testing.T) {
	value := getValue()
	if value <= 10 {
		t.Errorf("Expected value > 10, got %d", value)
	}
}

func TestGreaterOrEqualPattern(t *testing.T) {
	value := getValue()
	if value < 0 {
		t.Errorf("Expected value >= 0, got %d", value)
	}
}

func TestLessPattern(t *testing.T) {
	value := getValue()
	if value >= 100 {
		t.Errorf("Expected value < 100, got %d", value)
	}
}

func TestLessOrEqualPattern(t *testing.T) {
	value := getValue()
	if value > 255 {
		t.Errorf("Expected value <= 255, got %d", value)
	}
}

// Helper functions
func doSomething() error { return nil }
func calculate() int     { return 42 }
func checkCondition() bool { return true }
func getPointer() *int { return nil }
func getItems() []int { return nil }
func getValue() int { return 50 }
