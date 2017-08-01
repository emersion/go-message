// Package charset provides functions to decode and encode charsets.
package charset

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

var charsets = map[string]encoding.Encoding{
	"big5":         traditionalchinese.Big5,
	"euc-jp":       japanese.EUCJP,
	"gbk":          simplifiedchinese.GBK,
	"gb2312":       simplifiedchinese.GBK, // as  GBK is a superset of HZGB2312,so just use GBK
	"iso-2022-jp":  japanese.ISO2022JP,
	"iso-8859-1":   charmap.ISO8859_1,
	"iso-8859-2":   charmap.ISO8859_2,
	"iso-8859-3":   charmap.ISO8859_3,
	"iso-8859-4":   charmap.ISO8859_4,
	"iso-8859-10":  charmap.ISO8859_10,
	"iso-8859-13":  charmap.ISO8859_13,
	"iso-8859-14":  charmap.ISO8859_14,
	"iso-8859-15":  charmap.ISO8859_15,
	"iso-8859-16":  charmap.ISO8859_16,
	"koi8-r":       charmap.KOI8R,
	"shift_jis":    japanese.ShiftJIS,
	"windows-1250": charmap.Windows1250,
	"windows-1251": charmap.Windows1251,
	"windows-1252": charmap.Windows1252,
}

// Reader returns an io.Reader that converts the provided charset to UTF-8.
func Reader(charset string, input io.Reader) (io.Reader, error) {
	charset = strings.ToLower(charset)
	if charset == "utf-8" || charset == "us-ascii" {
		return input, nil
	}
	if enc, ok := charsets[charset]; ok {
		return enc.NewDecoder().Reader(input), nil
	}
	return nil, fmt.Errorf("unhandled charset %q", charset)
}
