# binding

Bind request arguments easily in Golang.

[![test status](https://github.com/miclle/binding/workflows/tests/badge.svg?branch=master "test status")](https://github.com/miclle/binding/actions)

## Usage
```go
package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/miclle/binding"
)

// MixStruct bind target struct
type MixStruct struct {
	Page       int        `query:"page"`
	PageSize   int        `query:"page_size"`
	IDs        []int      `query:"ids[]"`
	Start      *time.Time `query:"start"         time_format:"unix"`
	Referer    string     `header:"referer"`
	XRequestID string     `header:"X-Request-Id"`
	Vary       []string   `header:"vary"`
	Name       string     `json:"name"`
	Content    *string    `json:"content"`
}

func main() {

	var (
		obj        MixStruct
		url        = "/?page=1&page_size=30&ids[]=1&ids[]=2&ids[]=3&ids[]=4&ids[]=5&start=1669732749"
		referer    = "http://domain.name/posts"
		varyHeader = []string{"X-PJAX, X-PJAX-Container, Turbo-Visit, Turbo-Frame", "Accept-Encoding, Accept, X-Requested-With"}
		XRequestID = "l4dCIsjENo3QsCoX"
	)

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Referer", referer)
	req.Header.Set("X-Request-Id", XRequestID)
	req.Header.Add("vary", varyHeader[0])
	req.Header.Add("vary", varyHeader[1])

	_ = binding.Bind(req, &obj)

	fmt.Printf("%#v \n", obj)
	// MixStruct{
	//   Page:1,
	//   PageSize:30,
	//   IDs:[]int{1, 2, 3, 4, 5},
	//   Start:time.Date(2022, time.November, 29, 22, 39, 9, 0, time.Local),
	//   Referer:"http://domain.name/posts",
	//   XRequestID:"l4dCIsjENo3QsCoX",
	//   Vary:[]string{"X-PJAX, X-PJAX-Container, Turbo-Visit, Turbo-Frame", "Accept-Encoding, Accept, X-Requested-With"},
	//   Name:"",
	//   Content:(*string)(nil)
	// }

	req, _ = http.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{"name": "Binder"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", referer)
	req.Header.Set("X-Request-Id", XRequestID)
	req.Header.Add("vary", varyHeader[0])
	req.Header.Add("vary", varyHeader[1])

	obj = MixStruct{}
	_ = binding.Bind(req, &obj)

	fmt.Printf("%#v \n", obj)
	// MixStruct{
	//   Page:1,
	//   PageSize:30,
	//   IDs:[]int{1, 2, 3, 4, 5},
	//   Start:time.Date(2022, time.November, 29, 22, 39, 9, 0, time.Local),
	//   Referer:"http://domain.name/posts",
	//   XRequestID:"l4dCIsjENo3QsCoX",
	//   Vary:[]string{"X-PJAX, X-PJAX-Container, Turbo-Visit, Turbo-Frame", "Accept-Encoding, Accept, X-Requested-With"},
	//   Name:"Binder",
	//   Content:(*string)(nil)
	// }
}
```