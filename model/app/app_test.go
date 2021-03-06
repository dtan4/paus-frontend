package app

import (
	"testing"
)

func TestURL(t *testing.T) {
	uriScheme := "http"
	identifier := "dtan4-hoge"
	baseDomain := "pausapp.com"

	expected := "http://dtan4-hoge.pausapp.com"
	actual := URL(uriScheme, identifier, baseDomain)

	if expected != actual {
		t.Fatalf("Expected: %s, Actual: %s", expected, actual)
	}

	uriScheme = "HTTP"
	identifier = "Dtan4-Hoge"
	baseDomain = "Pausapp.com"

	expected = "http://dtan4-hoge.pausapp.com"
	actual = URL(uriScheme, identifier, baseDomain)

	if expected != actual {
		t.Fatalf("Expected: %s, Actual: %s", expected, actual)
	}
}

func TestLatestAppURLOfUser(t *testing.T) {
	uriScheme := "http"
	baseDomain := "pausapp.com"
	username := "dtan4"
	appName := "hoge"

	expected := "http://dtan4-hoge.pausapp.com"
	actual := LatestAppURLOfUser(uriScheme, baseDomain, username, appName)

	if expected != actual {
		t.Fatalf("Expected: %s, Actual: %s", expected, actual)
	}
}
