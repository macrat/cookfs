package main

import (
	"testing"
)

var (
	term_a = Term{1, ForceParseNode("http://localhost:8000")}
	term_b = Term{2, ForceParseNode("http://localhost:8001")}
	term_c = Term{1, ForceParseNode("http://localhost:8002")}
)

func Test_Term_String(t *testing.T) {
	if term_a.String() != "[Term 1](http://localhost:8000)" {
		t.Errorf("failed convert to string; got %s", term_a.String())
	}
	if term_b.String() != "[Term 2](http://localhost:8001)" {
		t.Errorf("failed convert to string; got %s", term_a.String())
	}
	if term_c.String() != "[Term 1](http://localhost:8002)" {
		t.Errorf("failed convert to string; got %s", term_a.String())
	}
}

func Test_Term_Equals(t *testing.T) {
	if term_a.Equals(term_b) {
		t.Errorf("saied %s and %s is equals", term_a, term_b)
	}
	if term_b.Equals(term_c) {
		t.Errorf("saied %s and %s is equals", term_b, term_c)
	}
	if term_a.Equals(term_c) {
		t.Errorf("saied %s and %s is equals", term_a, term_c)
	}
	if !term_a.Equals(term_a) {
		t.Errorf("saied %s and %s is not equals", term_a, term_a)
	}
}

func Test_Term_NewerThan(t *testing.T) {
	if term_a.NewerThan(term_b) {
		t.Errorf("saied %s is newer than %s", term_a, term_b)
	}
	if term_a.NewerThan(term_c) {
		t.Errorf("saied %s is newer than %s", term_a, term_c)
	}
	if !term_b.NewerThan(term_a) {
		t.Errorf("saied %s is not newer than %s", term_b, term_a)
	}
	if !term_b.NewerThan(term_c) {
		t.Errorf("saied %s is not newer than %s", term_b, term_c)
	}
}

func Test_Term_OlderThan(t *testing.T) {
	if !term_a.OlderThan(term_b) {
		t.Errorf("saied %s is not older than %s", term_a, term_b)
	}
	if term_a.OlderThan(term_c) {
		t.Errorf("saied %s is older than %s", term_a, term_c)
	}
	if term_b.OlderThan(term_a) {
		t.Errorf("saied %s is older than %s", term_b, term_a)
	}
	if term_b.OlderThan(term_c) {
		t.Errorf("saied %s is older than %s", term_b, term_c)
	}
}
