package main

import (
	"os"
	"reflect"
	"testing"
)

func TestParseWordFreq(t *testing.T) {
	wf, err := parseWordFreq("hello 42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wf.Word != "hello" || wf.Frequency != 42 {
		t.Fatalf("unexpected value: %#v", wf)
	}
	if _, err := parseWordFreq("invalid_line"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseCommand(t *testing.T) {
	cmd, prefix, err := parseCommand("get test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != "get" || prefix != "test" {
		t.Fatalf("unexpected: %s %s", cmd, prefix)
	}

	cases := []string{"", "get", "get 123", "set test", "get verylongprefixmorethan15"}
	for _, c := range cases {
		if _, _, err := parseCommand(c); err == nil {
			t.Errorf("expected error for %q", c)
		}
	}
}

func TestIsAlpha(t *testing.T) {
	if !isAlpha("abc") {
		t.Fatal("expected true")
	}
	if isAlpha("ab3") {
		t.Fatal("expected false")
	}
}

func TestLRUCache(t *testing.T) {
	cache := newLRUCache(2)
	a := []WordFreq{{Word: "a", Frequency: 1}}
	b := []WordFreq{{Word: "b", Frequency: 1}}
	c := []WordFreq{{Word: "c", Frequency: 1}}

	cache.Add("a", a)
	cache.Add("b", b)
	if _, ok := cache.Get("a"); !ok {
		t.Fatal("expected to find a")
	}
	cache.Add("c", c)
	if _, ok := cache.Get("b"); ok {
		t.Fatal("expected b to be evicted")
	}
}

func TestGetSuggestions(t *testing.T) {
	words := []WordFreq{{"app", 10}, {"apple", 5}, {"apples", 7}, {"banana", 3}}
	srv := NewServer(words)
	res := srv.getSuggestions("app")
	got := make([]string, len(res))
	for i, wf := range res {
		got[i] = wf.Word
	}
	expected := []string{"app", "apples", "apple"}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("got %v want %v", got, expected)
	}
}

func TestLoadWordFreq(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "wf.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	if _, err := tmp.WriteString("hello 1\nworld 2\n"); err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	res, err := loadWordFreq(tmp.Name())
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("unexpected length %d", len(res))
	}
}
