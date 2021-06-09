// https://github.com/codemicro/headers
// Copyright (c) 2021, codemicro and contributors
// SPDX-License-Identifier: MIT
// Filename: internal/headers/template.go

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
