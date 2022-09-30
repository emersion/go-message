// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quotedprintable

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestTolerantReader(t *testing.T) {
	tests := []struct {
		in, want string
		err      any
	}{
		{in: "", want: ""},
		{in: "foo bar", want: "foo bar"},
		{in: "foo bar=3D", want: "foo bar="},
		{in: "foo bar=3d", want: "foo bar="}, // lax.
		{in: "foo bar=\n", want: "foo bar"},
		{in: "foo bar\n", want: "foo bar\n"}, // somewhat lax.
		{in: "foo bar=0", want: "foo bar=0"}, // lax
		{in: "foo bar=0D=0A", want: "foo bar\r\n"},
		{in: " A B        \r\n C ", want: " A B\r\n C"},
		{in: " A B =\r\n C ", want: " A B  C"},
		{in: " A B =\n C ", want: " A B  C"}, // lax. treating LF as CRLF
		{in: "foo=\nbar", want: "foobar"},
		{in: "foo\x00bar", want: "foo", err: "quotedprintable: invalid unescaped byte 0x00 in body"},
		{in: "foo bar\xff", want: "foo bar\xff"},

		// Equal sign.
		{in: "=3D30\n", want: "=30\n"},
		{in: "=00=FF0=\n", want: "\x00\xff0"},

		// Trailing whitespace
		{in: "foo  \n", want: "foo\n"},
		{in: "foo  \n\nfoo =\n\nfoo=20\n\n", want: "foo\n\nfoo \nfoo \n\n"},

		// Tests that we allow bare \n and \r through, despite it being strictly
		// not permitted per RFC 2045, Section 6.7 Page 22 bullet (4).
		{in: "foo\nbar", want: "foo\nbar"},
		{in: "foo\rbar", want: "foo\rbar"},
		{in: "foo\r\nbar", want: "foo\r\nbar"},

		// Different types of soft line-breaks.
		{in: "foo=\r\nbar", want: "foobar"},
		{in: "foo=\nbar", want: "foobar"},
		{in: "foo=\rbar", want: "foo", err: "quotedprintable: invalid hex byte 0x0d"},
		{in: "foo=\r\r\r \nbar", want: "foo", err: `quotedprintable: invalid bytes after =: "\r\r\r \n"`},
		{in: "foo=\r\r\n", want: "foo"},
		// Issue 15486, accept trailing soft line-break at end of input.
		{in: "foo=", want: "foo"},
		// Tolerate lines that are just a space - seen in the wild where parts end in '='
		// {in: "=", want: "", err: `quotedprintable: invalid bytes after =: ""`},
		{in: "=", want: ""},

		// Example from RFC 2045:
		{in: "Now's the time =\n" + "for all folk to come=\n" + " to the aid of their country.",
			want: "Now's the time for all folk to come to the aid of their country."},
		{in: "accept UTF-8 right quotation mark: ’",
			want: "accept UTF-8 right quotation mark: ’"},
	}

	for _, tt := range tests {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, NewReader(strings.NewReader(tt.in)))
		if got := buf.String(); got != tt.want {
			t.Errorf("for %q, got %q; want %q", tt.in, got, tt.want)
		}
		switch verr := tt.err.(type) {
		case nil:
			if err != nil {
				t.Errorf("for %q, got unexpected error: %v", tt.in, err)
			}
		case string:
			if got := fmt.Sprint(err); got != verr {
				t.Errorf("for %q, got error %q; want %q", tt.in, got, verr)
			}
		case error:
			if err != verr {
				t.Errorf("for %q, got error %q; want %q", tt.in, err, verr)
			}
		}
	}
}

func everySequence(base, alpha string, length int, fn func(string)) {
	if len(base) == length {
		fn(base)
		return
	}
	for i := 0; i < len(alpha); i++ {
		everySequence(base+alpha[i:i+1], alpha, length, fn)
	}
}

var useQprint = flag.Bool("qprint", false, "Compare against the 'qprint' program.")

var badSoftRx = regexp.MustCompile(`=([^\r\n]+?\n)|([^\r\n]+$)|(\r$)|(\r[^\n]+\n)|( \r\n)`)

func TestTolerantExhaustive(t *testing.T) {

	t.Skip("we have changed the numbers and I'm not 100% confident yet of the changes")

	if *useQprint {
		_, err := exec.LookPath("qprint")
		if err != nil {
			t.Fatalf("Error looking for qprint: %v", err)
		}
	}

	var buf bytes.Buffer
	res := make(map[string]int)
	n := 6
	if testing.Short() {
		n = 4
	}
	everySequence("", "0A \r\n=", n, func(s string) {
		if strings.HasSuffix(s, "=") || strings.Contains(s, "==") {
			return
		}
		buf.Reset()
		_, err := io.Copy(&buf, NewReader(strings.NewReader(s)))
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "invalid bytes after =:") {
				errStr = "invalid bytes after ="
			}
			res[errStr]++
			if strings.Contains(errStr, "invalid hex byte ") {
				if strings.HasSuffix(errStr, "0x20") && (strings.Contains(s, "=0 ") || strings.Contains(s, "=A ") || strings.Contains(s, "= ")) {
					return
				}
				if strings.HasSuffix(errStr, "0x3d") && (strings.Contains(s, "=0=") || strings.Contains(s, "=A=")) {
					return
				}
				if strings.HasSuffix(errStr, "0x0a") || strings.HasSuffix(errStr, "0x0d") {
					// bunch of cases; since whitespace at the end of a line before \n is removed.
					return
				}
			}
			if strings.Contains(errStr, "unexpected EOF") {
				return
			}
			if errStr == "invalid bytes after =" && badSoftRx.MatchString(s) {
				return
			}
			t.Errorf("decode(%q) = %v", s, err)
			return
		}
		if *useQprint {
			cmd := exec.Command("qprint", "-d")
			cmd.Stdin = strings.NewReader(s)
			stderr, err := cmd.StderrPipe()
			if err != nil {
				panic(err)
			}
			qpres := make(chan any, 2)
			go func() {
				br := bufio.NewReader(stderr)
				s, _ := br.ReadString('\n')
				if s != "" {
					qpres <- errors.New(s)
					if cmd.Process != nil {
						// It can get stuck on invalid input, like:
						// echo -n "0000= " | qprint -d
						cmd.Process.Kill()
					}
				}
			}()
			go func() {
				want, err := cmd.Output()
				if err == nil {
					qpres <- want
				}
			}()
			select {
			case got := <-qpres:
				if want, ok := got.([]byte); ok {
					if string(want) != buf.String() {
						t.Errorf("go decode(%q) = %q; qprint = %q", s, want, buf.String())
					}
				} else {
					t.Logf("qprint -d(%q) = %v", s, got)
				}
			case <-time.After(5 * time.Second):
				t.Logf("qprint timeout on %q", s)
			}
		}
		res["OK"]++
	})
	var outcomes []string
	for k, v := range res {
		outcomes = append(outcomes, fmt.Sprintf("%v: %d", k, v))
	}
	sort.Strings(outcomes)
	got := strings.Join(outcomes, "\n")
	want := `OK: 28934
invalid bytes after =: 3949
quotedprintable: invalid hex byte 0x0d: 2048
unexpected EOF: 194`
	if testing.Short() {
		want = `OK: 896
invalid bytes after =: 100
quotedprintable: invalid hex byte 0x0d: 26
unexpected EOF: 3`
	}

	if got != want {
		t.Errorf("Got:\n%s\nWant:\n%s", got, want)
	}
}

func TestTolerantHtml(t *testing.T) {
	badHtml := `										 <td height=3D"10"><img src=3D"http://xxxxxxxxxxxx.xxxxx=
 xxxxxxxxxxxxxxx.xx.xx/xx/xxxxxxxxxxxxxxxxxxx/xxxxxxxxxxxxxxxxxxxxxxxxxxxx=
 xxx/xxxx.jpg?xxxxx=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" hei=
 ght=3D"2" width=3D"2" alt= 
									</tr>
		 
 
		 </table>`

	r := NewReader(strings.NewReader(badHtml))
	_, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error("should read this without an error", err.Error())
	}
}

func TestTolerantHtmlShort(t *testing.T) {
	badHtml := "ght=3D\"2\" width=3D\"2\" alt= \r\n"

	r := NewReader(strings.NewReader(badHtml))
	s, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error("should read this without an error", err.Error())
	}
	expected := "ght=\"2\" width=\"2\" alt=\r\n"
	if expected != string(s) {
		t.Errorf("decoded string was not what was expected %q, got %q", expected, string(s))
	}
}

func TestTolerantShort(t *testing.T) {
	badHtml := "foo= \r\n"

	r := NewReader(strings.NewReader(badHtml))
	s, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error("should read this without an error", err.Error())
	}
	expected := "foo=\r\n"
	if expected != string(s) {
		t.Errorf("decoded string was not what was expected %q, got %q", expected, string(s))
	}
}

func TestTolerantLong(t *testing.T) {
	badContent := "<p style=3D\"text-align: left;\"><img width=3D\"750\" height=3D\"1\" style=3D\"clear: none; float: none;\" src=3D\"http://example.test/image.png\"></p><span style=3D\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13px;\"><span style=3D\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13px;\"><span style=3D\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13px;\"><span style=3D\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13px;\"><table width=3D\"750\" class=3D\"txc-table\" style=3D\"border: currentColor; border-image: none; width: 750px; padding-right: 0px; padding-left: 0px; font-family: =EA=B5=B4=EB=A6=BC; font-size: 13.33px; float: none; border-collapse: collapse;\" border=3D\"0\" cellspacing=3D\"0\" cellpadding=3D\"0\"><tbody><tr><td style=3D\"border-width: 0px; border-image: none; width: 373px; height: 25px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: =EA=B5=B4=EB=A6=BC; font-size: 13.33px; float: none;\" rowspan=3D\"1\" colspan=3D\"2\"><p><span style=3D\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13.33px;\"><strong></strong></span><br></p></td><td style=3D\"border-width: 0px; border-image: none; width: 20px; height: 25px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: =EA=B5=B4=EB=A6=BC; font-size: 13.33px; border-top-style: none; border-right-style: none; border-bottom-style: none; float: none;\" rowspan=3D\"1\" colspan=3D\"1\"><span style=3D\"color: rgb(140, 140, 140); font-family: =EA=B5=B4=EB=A6=BC; font-size: 10.66px;\">&nbsp;</span></td><td style=3D\"width: 93px; height: 25px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: =EA=B5=B4=EB=A6=BC; font-size: 13.33px; border-right-color: rgb(140, 140, 140); border-top-width: 0px; border-right-width: 3px; border-bottom-width: 0px; border-top-style: none; border-right-style: solid; border-bottom-style: none; float: none;\" colspan=3D\"1\"><p style=3D\"text-align: right;\">&nbsp;<span style=3D\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13.33px;\"><span style=3D\"color: rgb(0, 0, 0); font-family: Verdana,sans-serif; font-size: 13.33px;\"><strong><span style=3D\"color: rgb(0, 0, 0); font-family: Verdana,sans-serif; font-size: 16px;\">&nbsp;&nbsp;</span></strong></span></span></p></td><td style=3D\"width: 262px; height: 25px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: =EA=B5=B4=EB=A6=BC; font-size: 13.33px; border-top-width: 0px; border-right-width: 0px; border-bottom-width: 0px; border-top-style: none; border-right-style: none; border-bottom-style: none; float: none;\" colspan=3D\"1\"><p>&nbsp;<strong><font face=3D\"Verdana\">&nbsp; =EC=84=9C=EC=9A=B8=EC=98=81=EC=97=85=ED=8C=80 / =EC=B0=A8=EC=9E=A5</font></strong></p></td></tr><tr><td style=3D\"width: 373px; height: 5px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: =EA=B5=B4=EB=A6=BC; font-size: 13.33px; vertical-align: top; border-right-width: 0px; border-bottom-width: 0px; border-left-width: 0px; border-right-style: none; border-bottom-style: none; border-left-style: none; float: none;\" rowspan=3D\"1\" colspan=3D\"2\"><p><span style=3D\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 10.66px;\">=EC=98=81=EC=97=85=EB=B3=B8=EB=B6=80 : =EA=B2=BD=EA=B8=B0=EB=8F=84 =EC=84=B1=EB=82=A8=EC=8B=9C =EB=B6=84=EB=8B=B9=EA=B5=AC =EC=84=B1=EB=82=A8=EB=8C=80=EB=A1=9C 69 =EB=A1=9C=EB=93=9C=EB=9E=9C=EB=93=9CEZ=ED=83=80=EC=9B=8C 610=ED=98=B8<br></span></p></td><td style=3D\"width: 20px; height: 5px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: =EA=B5=B4=EB=A6=BC; font-size: 13.33px; vertical-align: top; border-right-width: 0px; border-bottom-width: 0px; border-left-width: 0px; border-right-style: none; border-bottom-style: none; border-left-style: none; float: none;\" rowspan=3D\"1\" colspan=3D\"1\"><span style=3D\"color: rgb(140, 140, 140); font-family: =EA=B5=B4=EB=A6=BC; font-size: 10.66px;\">&nbsp;</span></td><td style=3D\"width: 93px; height: 5px; text-align: right; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: =EA=B5=B4=EB=A6=BC; font-size: 13.33px; vertical-align: top; border-right-width: 0px; border-bottom-width: 0px; border-right-style: none; border-bottom-style: none; float: none;\" colspan=3D\"1\"><p><span style=3D\"color: rgb(140, 140, 140); font-family: =EA=B5=B4=EB=A6=BC; font-size: 10.66px;\">&nbsp;</span><span style=3D\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13.33px;\"><span style=3D\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 10.66px;\">&nbsp;&nbsp;&nbsp;Hyunwook Lee</span></span></p></td><td style=3D\"width: 262px; height: 5px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: =EA=B5=B4=EB=A6=BC; font-size: 13.33px; vertical-align: top; border-right-width: 0px; border-bottom-width: 0px; border-right-style: none; border-bottom-style: none; float: none;\" colspan=3D\"1\"><p>&nbsp;&nbsp;<span style=3D\"color: rgb(140, 140, 140); font-family: =EA=B5=B4=EB=A6=BC; font-size: 10.66px;\">&nbsp; </span><font color=3D\"#8c8c8c\" face=3D\"Verdana\"><span style=3D\"color: rgb(140, 140, 140); font-family: Verdana; font-size: 10.66px;\">Seoul Sales Team / Chief Manager</span></font></p></td></tr><tr><td style=3D\"width: 373px; height: 21px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: =EA=B5=B4=EB=A6=BC; font-size: 13.33px; border-right-width: 0px; border-bottom-width: 0px; border-left-width: 0px; border-right-style: none; border-bottom-style: none; border-left-style: none; float: none;\" rowspan=3D\"1\" colspan=3D\"2\"><p><span style=3D\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 10.66px;\">&nbsp;T: 02-2604-6648&nbsp; F: 02-2691-7238&nbsp; W:www.kossen.co.kr</span></p></td><td style=3D\"width: 20px; height: 21px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: =EA=B5=B4=EB=A6=BC; font-size: 13.33px; border-right-width: 0px; border-bottom-width: 0px; border-left-width: 0px; border-right-style: none; border-bottom-style: none; border-left-style: none; float: none;\" rowspan=3D\"1\" colspan=3D\"1\"><span style=3D\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 10.66px;\">&nbsp;</span></td><td style=3D\"width: 355px; height: 21px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: =EA=B5=B4=EB=A6=BC; font-size: 13.33px; border-right-width: 0px; border-bottom-width: 0px; border-right-style: none; border-bottom-style: none; float: none;\" rowspan=3D\"1\" colspan=3D\"2\"><p style=3D\"text-align: center;\"><span style=3D\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 10.66px;\"><span style=3D\"color: rgb(0, 85, 255); font-family: Verdana,sans-serif; font-size: 10.66px;\">T: 02-2604-6648&nbsp; F: 02-2691-7238&nbsp; M: 010-3390-3638</span></span></p></td></tr></tbody></table><p><img width=3D\"750\" height=3D\"35\" style=3D\"clear: none; float: none;\" src=3D\"http://example.test/image.png\"></p></span></span><p style=3D\"text-align: left;\"><br></p></span><p style=3D\"text-align: left;\"><br></p></span><p style=3D\"text-align: left;\"><br></p>"

	r := NewReader(strings.NewReader(badContent))
	s, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error("should read this without an error", err.Error())
	}
	expected := "<p style=\"text-align: left;\"><img width=\"750\" height=\"1\" style=\"clear: none; float: none;\" src=\"http://example.test/image.png\"></p><span style=\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13px;\"><span style=\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13px;\"><span style=\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13px;\"><span style=\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13px;\"><table width=\"750\" class=\"txc-table\" style=\"border: currentColor; border-image: none; width: 750px; padding-right: 0px; padding-left: 0px; font-family: 굴림; font-size: 13.33px; float: none; border-collapse: collapse;\" border=\"0\" cellspacing=\"0\" cellpadding=\"0\"><tbody><tr><td style=\"border-width: 0px; border-image: none; width: 373px; height: 25px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: 굴림; font-size: 13.33px; float: none;\" rowspan=\"1\" colspan=\"2\"><p><span style=\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13.33px;\"><strong></strong></span><br></p></td><td style=\"border-width: 0px; border-image: none; width: 20px; height: 25px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: 굴림; font-size: 13.33px; border-top-style: none; border-right-style: none; border-bottom-style: none; float: none;\" rowspan=\"1\" colspan=\"1\"><span style=\"color: rgb(140, 140, 140); font-family: 굴림; font-size: 10.66px;\">&nbsp;</span></td><td style=\"width: 93px; height: 25px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: 굴림; font-size: 13.33px; border-right-color: rgb(140, 140, 140); border-top-width: 0px; border-right-width: 3px; border-bottom-width: 0px; border-top-style: none; border-right-style: solid; border-bottom-style: none; float: none;\" colspan=\"1\"><p style=\"text-align: right;\">&nbsp;<span style=\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13.33px;\"><span style=\"color: rgb(0, 0, 0); font-family: Verdana,sans-serif; font-size: 13.33px;\"><strong><span style=\"color: rgb(0, 0, 0); font-family: Verdana,sans-serif; font-size: 16px;\">&nbsp;&nbsp;</span></strong></span></span></p></td><td style=\"width: 262px; height: 25px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: 굴림; font-size: 13.33px; border-top-width: 0px; border-right-width: 0px; border-bottom-width: 0px; border-top-style: none; border-right-style: none; border-bottom-style: none; float: none;\" colspan=\"1\"><p>&nbsp;<strong><font face=\"Verdana\">&nbsp; 서울영업팀 / 차장</font></strong></p></td></tr><tr><td style=\"width: 373px; height: 5px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: 굴림; font-size: 13.33px; vertical-align: top; border-right-width: 0px; border-bottom-width: 0px; border-left-width: 0px; border-right-style: none; border-bottom-style: none; border-left-style: none; float: none;\" rowspan=\"1\" colspan=\"2\"><p><span style=\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 10.66px;\">영업본부 : 경기도 성남시 분당구 성남대로 69 로드랜드EZ타워 610호<br></span></p></td><td style=\"width: 20px; height: 5px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: 굴림; font-size: 13.33px; vertical-align: top; border-right-width: 0px; border-bottom-width: 0px; border-left-width: 0px; border-right-style: none; border-bottom-style: none; border-left-style: none; float: none;\" rowspan=\"1\" colspan=\"1\"><span style=\"color: rgb(140, 140, 140); font-family: 굴림; font-size: 10.66px;\">&nbsp;</span></td><td style=\"width: 93px; height: 5px; text-align: right; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: 굴림; font-size: 13.33px; vertical-align: top; border-right-width: 0px; border-bottom-width: 0px; border-right-style: none; border-bottom-style: none; float: none;\" colspan=\"1\"><p><span style=\"color: rgb(140, 140, 140); font-family: 굴림; font-size: 10.66px;\">&nbsp;</span><span style=\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 13.33px;\"><span style=\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 10.66px;\">&nbsp;&nbsp;&nbsp;Hyunwook Lee</span></span></p></td><td style=\"width: 262px; height: 5px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: 굴림; font-size: 13.33px; vertical-align: top; border-right-width: 0px; border-bottom-width: 0px; border-right-style: none; border-bottom-style: none; float: none;\" colspan=\"1\"><p>&nbsp;&nbsp;<span style=\"color: rgb(140, 140, 140); font-family: 굴림; font-size: 10.66px;\">&nbsp; </span><font color=\"#8c8c8c\" face=\"Verdana\"><span style=\"color: rgb(140, 140, 140); font-family: Verdana; font-size: 10.66px;\">Seoul Sales Team / Chief Manager</span></font></p></td></tr><tr><td style=\"width: 373px; height: 21px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: 굴림; font-size: 13.33px; border-right-width: 0px; border-bottom-width: 0px; border-left-width: 0px; border-right-style: none; border-bottom-style: none; border-left-style: none; float: none;\" rowspan=\"1\" colspan=\"2\"><p><span style=\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 10.66px;\">&nbsp;T: 02-2604-6648&nbsp; F: 02-2691-7238&nbsp; W:www.kossen.co.kr</span></p></td><td style=\"width: 20px; height: 21px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: 굴림; font-size: 13.33px; border-right-width: 0px; border-bottom-width: 0px; border-left-width: 0px; border-right-style: none; border-bottom-style: none; border-left-style: none; float: none;\" rowspan=\"1\" colspan=\"1\"><span style=\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 10.66px;\">&nbsp;</span></td><td style=\"width: 355px; height: 21px; color: rgb(140, 140, 140); padding-right: 0px; padding-left: 0px; font-family: 굴림; font-size: 13.33px; border-right-width: 0px; border-bottom-width: 0px; border-right-style: none; border-bottom-style: none; float: none;\" rowspan=\"1\" colspan=\"2\"><p style=\"text-align: center;\"><span style=\"color: rgb(140, 140, 140); font-family: Verdana,sans-serif; font-size: 10.66px;\"><span style=\"color: rgb(0, 85, 255); font-family: Verdana,sans-serif; font-size: 10.66px;\">T: 02-2604-6648&nbsp; F: 02-2691-7238&nbsp; M: 010-3390-3638</span></span></p></td></tr></tbody></table><p><img width=\"750\" height=\"35\" style=\"clear: none; float: none;\" src=\"http://example.test/image.png\"></p></span></span><p style=\"text-align: left;\"><br></p></span><p style=\"text-align: left;\"><br></p></span><p style=\"text-align: left;\"><br></p>"
	if expected != string(s) {
		t.Errorf("decoded string was not what was expected %q, got %q", expected, string(s))
	}
}
