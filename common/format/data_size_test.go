package format

import "testing"

func TestDataSize(t *testing.T) {
	t.Log(DataSize(75293847523))
}

func TestDataSizeAsString(t *testing.T) {
	t.Log(DataSizeAsString(2*1024*1024*1024, SizeUnitGB))
	t.Log(DataSizeAsString(2048, SizeUnitKB))
	t.Log(DataSize(0))
	t.Log(DataSize(3000))
}
