package webcrawler

import (
	"context"
	"testing"
)

func TestXxx(t *testing.T) {
	expected := []string{
		"https://example.com/home",
		"https://example.com/doc",
	}

	resource := newResource()
	resource.URL = "https://example.com"
	resource.rawBuffer.Write([]byte(htmlPage))
	res, err := ExtractURLs()(context.TODO(), resource)
	if err != nil {
		t.Fatal(err)
	}

	if len(expected) != len(res.FoundURLs) {
		t.Fatalf("len not matched got %v", len(res.FoundURLs))
	}

	for i, link := range res.FoundURLs {
		if link != expected[i] {
			t.Fatalf("\ngot %v\n expected %v", link, expected[i])
		}
	}
}

var htmlPage = `
<!DOCTYPE html>

<html lang="en">
	<head>
	</head>
	<body>
	<a href="https://example.com/doc.jpg"><a>
	<a href="/home/"><a>
	<a href="/doc.html"><a>
	<a href="/doc.jpg"><a>
	<a href="https://example.com/doc.jpg/"><a>
	<a href="/"><a>
	<a href=""><a>

	</body>
</html>
`

func BenchmarkXxx(b *testing.B) {
	resource := newResource()
	resource.URL = "https://example.com"

	for i := 0; i < b.N; i++ {

		resource.rawBuffer.Write([]byte(htmlPage))
		res, err := ExtractURLs()(context.TODO(), resource)
		if err != nil {
			b.Fatal(err)
		}
		if len(res.FoundURLs) != 2 {
			b.Fatal("got len", len(res.FoundURLs))
		}

		res.rawBuffer.Reset()
		res.FoundURLs = res.FoundURLs[:0]
	}
}
