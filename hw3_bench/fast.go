package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

//easyjson:json
type User struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Browsers []string `json:"browsers"`
}

// результаты
//BenchmarkSlow
//BenchmarkSlow-8               51          22440963 ns/op        18949445 B/op     195828 allocs/op
//BenchmarkFast
//BenchmarkFast-8             1130           1053090 ns/op          316394 B/op       3998 allocs/op
//PASS

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()

	user := User{}
	seenBrowsers := make(map[string]bool, 20)
	foundUsers := make([]string, 0, 100)
	androidSubstr := "Android"
	msieSubstr := "MSIE"

	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan(); i++ {
		if !(bytes.Contains(scanner.Bytes(), []byte(androidSubstr)) ||
			bytes.Contains(scanner.Bytes(), []byte(msieSubstr))) {
			continue
		}

		user.UnmarshalJSON(scanner.Bytes())

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {
			if strings.Index(browser, androidSubstr) != -1 {
				isAndroid = true
				seenBrowsers[browser] = true
			} else if strings.Index(browser, msieSubstr) != -1 {
				isMSIE = true
				seenBrowsers[browser] = true
			}
		}
		if !(isAndroid && isMSIE) {
			continue
		}
		email := strings.Replace(user.Email, "@", " [at] ", 1)
		foundUsers = append(foundUsers, fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email))

	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	fmt.Fprintln(out, "found users:\n"+strings.Join(foundUsers, ""))
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}

//---------------------------------------------------------------------------------------
// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson7da3ae25DecodeCourseraGolangHomeworks(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "name":
			out.Name = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson7da3ae25EncodeCourseraGolangHomeworks(out *jwriter.Writer, in User) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"Name\":"
		out.RawString(prefix[1:])
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"Email\":"
		out.RawString(prefix)
		out.String(string(in.Email))
	}
	{
		const prefix string = ",\"Browsers\":"
		out.RawString(prefix)
		if in.Browsers == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Browsers {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v User) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson7da3ae25EncodeCourseraGolangHomeworks(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v User) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson7da3ae25EncodeCourseraGolangHomeworks(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson7da3ae25DecodeCourseraGolangHomeworks(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *User) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson7da3ae25DecodeCourseraGolangHomeworks(l, v)
}
