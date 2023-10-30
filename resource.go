package webcrawler

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var resourcePool = sync.Pool{
	New: func() any {
		return newResource()
	},
}

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

func newResource() *Resource {
	r := &Resource{
		rawBuffer: bytes.Buffer{},
	}
	return r
}

func NewResource() *Resource {
	r := resourcePool.Get().(*Resource)
	return r
}

func (r *Resource) Put() {
	r.ID = uuid.Nil
	r.URL = r.URL[:0]
	r.Title = r.Title[:0]
	r.Content = r.Content[:0]
	r.FoundURLs = r.FoundURLs[:0]
	r.retrieve = time.Time{}
	r.rawBuffer.Reset()
	resourcePool.Put(r)
}

func (r *Resource) Retrieved() {
	r.retrieve = time.Now().UTC()
}

func (r *Resource) String() string {
	return fmt.Sprintf("\ntitle: %s \ncontent: %s \nfoundURLs: %v", r.Title, r.Content, r.FoundURLs)
}
