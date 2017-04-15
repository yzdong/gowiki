package main

import (
	"bytes"
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

/*func TestMakeHandler(t *testing.T) {

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

func TestAddLinksToPage(t *testing.T) {
	arg := "not called"
	saved := false
	mp := &MockPage{
		mockAddLinks: func(kw string) {
			arg = kw
		},
		mockSave: func() error {
			saved = true
			return nil
		},
	}
	keyword := "dumpling"
	addLinksToPage(mp, keyword)
	if arg != keyword {
		t.Errorf("expecting addLinks to have been called with %s, was instead %s", keyword, arg)
	}
	if !saved {
		t.Errorf("expecting save to have been called")
	}
}
