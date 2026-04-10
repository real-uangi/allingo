package db

import (
	"context"
	"testing"
)

func TestJSONBValueUsesJSONString(t *testing.T) {
	value, err := NewJSONB([]string{"111", "222"}).Value()
	if err != nil {
		t.Fatal(err)
	}

	str, ok := value.(string)
	if !ok {
		t.Fatalf("expected driver.Value to be string, got %T", value)
	}

	if str != `["111","222"]` {
		t.Fatalf("unexpected json value: %s", str)
	}
}

func TestJSONBScanSupportsStringAndNil(t *testing.T) {
	var fromString JSONB[[]string]
	if err := fromString.Scan(`["111","222"]`); err != nil {
		t.Fatal(err)
	}

	got := fromString.Get()
	if len(got) != 2 || got[0] != "111" || got[1] != "222" {
		t.Fatalf("unexpected scan result: %#v", got)
	}

	if err := fromString.Scan(nil); err != nil {
		t.Fatal(err)
	}

	if fromString.Get() != nil {
		t.Fatalf("expected nil value after scanning nil, got %#v", fromString.Get())
	}
}

func TestJSONBGormValueCastsToJSONB(t *testing.T) {
	expr := NewJSONB(map[string]string{"k": "v"}).GormValue(context.Background(), nil)

	if expr.SQL != "?::jsonb" {
		t.Fatalf("unexpected sql: %s", expr.SQL)
	}

	if len(expr.Vars) != 1 {
		t.Fatalf("unexpected vars length: %d", len(expr.Vars))
	}

	str, ok := expr.Vars[0].(string)
	if !ok {
		t.Fatalf("expected cast var to be string, got %T", expr.Vars[0])
	}

	if str != `{"k":"v"}` {
		t.Fatalf("unexpected cast var: %s", str)
	}
}
