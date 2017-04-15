package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

/*func TestMain(t *testing.T) {
	server := httptest.NewServer(main())
	defer server.Close()
	resp, _ := http.Get(server.URL)
	expected, _ := ioutil.ReadFile("index.html")
	actual, _ := ioutil.ReadAll(resp.Body)
	if string(actual) != string(expected) {
		t.Errorf("expecting default page to be loaded, got %s", resp)
	}
}*/

func TestDefaultHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	defaultHandler(w, r)
	expected := "Hello!"
	actual := w.Body.String()
	if !strings.Contains(actual, expected) {
		t.Errorf("expecting %s to contain %s", actual, expected)
	}
}

func TestSave(t *testing.T) {
	p := &Page{Title: "Test", Body: []byte("Article")}
	err := p.save()
	dirname := "data/"
	_, err = ioutil.ReadFile(dirname + p.Title + ".txt")
	if err != nil {
		t.Errorf("expecting file to be saved in %s, instead encountered %s error", dirname, err)
	}
	err = os.Remove("data/Test.txt")
}

func TestAddLinks(t *testing.T) {
	body := []byte("Pie is nice. So is dumpling.")
	p := &Page{Title: "Test", Body: body}
	keyword := "dumpling"
	p.addLinks(keyword)
	expected := []byte(`<a href="/view/dumpling">dumpling</a>`)
	if !bytes.Contains(p.Body, expected) {
		t.Errorf("expecting %s to contain %s", string(p.Body), string(expected))
	}
}

type MockPage struct {
	mockSave     func() error
	mockAddLinks func(string)
}

func (p *MockPage) addLinks(keyword string) {
	if p.mockAddLinks != nil {
		p.mockAddLinks(keyword)
	}
}

func (p *MockPage) save() error {
	if p.mockSave != nil {
		return p.mockSave()
	}
	return nil
}

func TestAddLinksToPages(t *testing.T) {
	arg1 := "not called"
	saved1 := false
	arg2 := "not called"
	err := "File not found"
	mp1 := &MockPage{
		mockAddLinks: func(kw string) {
			arg1 = kw
		},
		mockSave: func() error {
			saved1 = true
			return nil
		},
	}
	mp2 := &MockPage{
		mockAddLinks: func(kw string) {
			arg2 = kw
		},
		mockSave: func() error {
			return errors.New(err)
		},
	}
	mps := &Pages{
		All: []PageInterface{mp1, mp2},
	}
	keyword := "dumpling"
	errs := mps.addLinksToPages(keyword)

	if arg1 != keyword {
		t.Errorf("expecting addLinks to have been called with %s, was instead %s", keyword, arg1)
	}
	if !saved1 {
		t.Errorf("expecting save to have been called")
	}
	if !(len(errs) == 1) {
		t.Errorf("expecting 1 error, found %v", len(errs))
		t.Log(errs)
	}
	if errs[0].Error() != err {
		t.Errorf("expecting error %s to be %s", errs[0], err)
	}
}

type MockPages struct {
	mockAddLinksToPages func(string) []error
}

func (ps *MockPages) addLinksToPages(keyword string) []error {
	if ps.mockAddLinksToPages != nil {
		return ps.mockAddLinksToPages(keyword)
	}
	return nil
}

func TestSaveHandler(t *testing.T) {
	called := false
	pages := &MockPages{
		mockAddLinksToPages: func(keyword string) []error {
			called = true
			return nil
		},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	title := "dumpling"
	saveHandler(w, r, title, pages)
	if !called && pages != nil {
		t.Errorf("expecting addLinksToPages to have been called")
	}
}
