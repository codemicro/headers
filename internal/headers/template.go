package headers

import (
	"time"
)

type headerTemplate struct {
	Year     int
	Filename string
}

func newHeaderTemplate(filename string) *headerTemplate {
	return &headerTemplate{
		Year:     time.Now().Year(),
		Filename: filename,
	}
}
