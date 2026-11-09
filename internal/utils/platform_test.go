package utils

import "testing"

func TestPrintOSInfo(t *testing.T) {
	info, err := GetOSInfo()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", info)
}
