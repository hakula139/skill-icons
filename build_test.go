package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestBuildIcons(t *testing.T) {
	want := map[string]string{
		"amber":       "<svg title=\"quoted\">`tick`\nsecond line</svg>",
		"birch-dark":  "<svg>dark</svg>",
		"birch-light": "<svg>light</svg>",
	}
	files := fstest.MapFS{
		"Amber.svg":          {Data: []byte(want["amber"])},
		"Birch-Dark.svg":     {Data: []byte(want["birch-dark"])},
		"Birch-Light.svg":    {Data: []byte(want["birch-light"])},
		"nested/ignored.svg": {Data: []byte("nested")},
	}

	first, err := buildIcons(files)
	if err != nil {
		t.Fatal(err)
	}
	second, err := buildIcons(files)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]string
	if err := json.Unmarshal(first, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("decoded icons = %#v, want %#v", got, want)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("repeated builds produced different JSON")
	}
}

func TestBuildIconsErrors(t *testing.T) {
	for _, name := range []string{".", "amber.svg"} {
		t.Run(name, func(t *testing.T) {
			files := failingFS{FS: fstest.MapFS{"amber.svg": {Data: []byte("<svg/>")}}, fail: name}

			_, err := buildIcons(files)

			if err == nil {
				t.Fatal("expected filesystem error")
			}
		})
	}
}

type failingFS struct {
	fs.FS
	fail string
}

func (f failingFS) Open(name string) (fs.File, error) {
	if name == f.fail {
		return nil, fs.ErrPermission
	}
	return f.FS.Open(name)
}
