package dpi

import (
	"testing"

	"bufio"
	"bytes"
)

func TestParse(t *testing.T) {
	cases := []byte("<dpi cmd='open_url' url='test1'><dpi cmd='add_bookmark' url='test2' title='tit''l''''e1'>")

	br := bytes.NewReader(cases)
	r := bufio.NewReader(br)

	ret, err := parseCmd(r)
	if err != nil {
		t.Fatalf("failed to read cmd: %v", err)
	}

	if ret["cmd"] != "open_url" {
		t.Errorf("expected open_url, got '%s'", ret["cmd"])
	}
	if ret["url"] != "test1" {
		t.Errorf("expected test1, got '%s'", ret["url"])
	}

	ret, err = parseCmd(r)
	if err != nil {
		t.Fatalf("failed to read cmd: %v", err)
	}

	if ret["cmd"] != "add_bookmark" {
		t.Errorf("expected add_bookmark, got '%s'", ret["cmd"])
	}
	if ret["url"] != "test2" {
		t.Errorf("expected test2, got '%s'", ret["url"])
	}
	if ret["title"] != "tit'l''e1" {
		t.Errorf("expected test2, got '%s'", ret["url"])
	}

	ret, err = parseCmd(r)
	if err == nil {
		t.Errorf("expected to fail")
	}
}
