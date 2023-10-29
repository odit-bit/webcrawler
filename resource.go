package webcrawler

import (
	"bytes"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// represents HTTP URL as Resouces that will crawled
type Resource struct {
	ID  uuid.UUID
	URL string

	// buffer for http Response's body
	rawBuffer bytes.Buffer

	// founded urls in html
	FoundURLs []string

	// title is html title of resource's url
	Title []byte
	//content is html body content of resource's url
	Content []byte

	//retrieve indicate time that url is being crawled (retrieved)
	retrieve time.Time
}

func (r *Resource) Retrieved() {
	r.retrieve = time.Now().UTC()
}

func (r *Resource) String() string {
	return fmt.Sprintf("\ntitle: %s \ncontent: %s \nfoundURLs: %v", r.Title, r.Content, r.FoundURLs)
}