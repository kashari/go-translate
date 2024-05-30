package bread

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetWithClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><p>Hello, World!</p></body></html>"))
	}))
	defer server.Close()

	html, err := GetWithClient(server.URL, http.DefaultClient)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "<html><body><p>Hello, World!</p></body></html>"
	if html != expected {
		t.Fatalf("expected %s, got %s", expected, html)
	}
}

func TestPostWithClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>" + string(body) + "</body></html>"))
	}))
	defer server.Close()

	data := url.Values{}
	data.Set("name", "John")
	html, err := PostWithClient(server.URL, "application/x-www-form-urlencoded", data, http.DefaultClient)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "<html><body>name=John</body></html>"
	if html != expected {
		t.Fatalf("expected %s, got %s", expected, html)
	}
}

func TestHTMLParse(t *testing.T) {
	htmlContent := "<html><body><p>Hello, World!</p></body></html>"
	root := HTMLParse(htmlContent)
	if root.Error != nil {
		t.Fatalf("expected no error, got %v", root.Error)
	}

	if root.NodeValue != "html" {
		t.Fatalf("expected 'html', got %s", root.NodeValue)
	}

	body := root.Find("body")
	if body.NodeValue != "body" {
		t.Fatalf("expected 'body', got %s", body.NodeValue)
	}

	p := body.Find("p")
	if p.NodeValue != "p" {
		t.Fatalf("expected 'p', got %s", p.NodeValue)
	}

	text := p.Text()
	if text != "Hello, World!" {
		t.Fatalf("expected 'Hello, World!', got %s", text)
	}
}

func TestFindAll(t *testing.T) {
	htmlContent := "<html><body><p>First</p><p>Second</p></body></html>"
	root := HTMLParse(htmlContent)
	if root.Error != nil {
		t.Fatalf("expected no error, got %v", root.Error)
	}

	ps := root.FindAll("p")
	if len(ps) != 2 {
		t.Fatalf("expected 2 <p> elements, got %d", len(ps))
	}

	if ps[0].Text() != "First" {
		t.Fatalf("expected 'First', got %s", ps[0].Text())
	}

	if ps[1].Text() != "Second" {
		t.Fatalf("expected 'Second', got %s", ps[1].Text())
	}
}

func TestFindNextSibling(t *testing.T) {
	htmlContent := "<html><body><p>First</p><p>Second</p></body></html>"
	root := HTMLParse(htmlContent)
	if root.Error != nil {
		t.Fatalf("expected no error, got %v", root.Error)
	}

	firstP := root.Find("p")
	if firstP.Error != nil {
		t.Fatalf("expected no error, got %v", firstP.Error)
	}

	secondP := firstP.FindNextSibling()
	if secondP.NodeValue != "p" || secondP.Text() != "Second" {
		t.Fatalf("expected next sibling to be <p>Second</p>, got %s", secondP.NodeValue)
	}
}

func TestAttrs(t *testing.T) {
	htmlContent := `<html><body><p class="greeting">Hello, World!</p></body></html>`
	root := HTMLParse(htmlContent)
	if root.Error != nil {
		t.Fatalf("expected no error, got %v", root.Error)
	}

	p := root.Find("p")
	if p.Error != nil {
		t.Fatalf("expected no error, got %v", p.Error)
	}

	attrs := p.Attrs()
	if attrs["class"] != "greeting" {
		t.Fatalf("expected class='greeting', got class='%s'", attrs["class"])
	}
}

func TestChildren(t *testing.T) {
	htmlContent := `<html><body><div><p>First</p><p>Second</p></div></body></html>`
	root := HTMLParse(htmlContent)
	if root.Error != nil {
		t.Fatalf("expected no error, got %v", root.Error)
	}

	div := root.Find("div")
	if div.Error != nil {
		t.Fatalf("expected no error, got %v", div.Error)
	}

	children := div.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	if children[0].NodeValue != "p" || children[0].Text() != "First" {
		t.Fatalf("expected first child to be <p>First</p>, got %s", children[0].NodeValue)
	}

	if children[1].NodeValue != "p" || children[1].Text() != "Second" {
		t.Fatalf("expected second child to be <p>Second</p>, got %s", children[1].NodeValue)
	}
}
