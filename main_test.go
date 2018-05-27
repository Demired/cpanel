package main

import "testing"

func TestInstallDB(t *testing.T) {
	err := InstallDB()
	if err != nil {
		t.Log("install db ok")
	} else {
		t.Fatal("install db field")
	}
}
