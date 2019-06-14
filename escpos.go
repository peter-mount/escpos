package escpos

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

const (
	// ASCII DLE (DataLinkEscape)
	DLE byte = 0x10

	// ASCII EOT (EndOfTransmission)
	EOT byte = 0x04

	// ASCII GS (Group Separator)
	GS byte = 0x1D
)

// text replacement map
var textReplaceMap = map[string]string{
	// horizontal tab
	"&#9;":  "\x09",
	"&#x9;": "\x09",

	// linefeed
	"&#10;": "\n",
	"&#xA;": "\n",

	// xml stuff
	"&apos;": "'",
	"&quot;": `"`,
	"&gt;":   ">",
	"&lt;":   "<",

	// ampersand must be last to avoid double decoding
	"&amp;": "&",
}

// replace text from the above map
func textReplace(data string) string {
	for k, v := range textReplaceMap {
		data = strings.Replace(data, k, v, -1)
	}
	return data
}

type Escpos struct {
	dst             io.ReadWriter
	buffer          bytes.Buffer
	width, height   uint8
	underline       uint8
	emphasize       uint8
	upsidedown      uint8
	rotate          uint8
	reverse, smooth uint8
}

// reset toggles
func (e *Escpos) reset() *Escpos {
	e.width = 1
	e.height = 1

	e.underline = 0
	e.emphasize = 0
	e.upsidedown = 0
	e.rotate = 0

	e.reverse = 0
	e.smooth = 0
	return e
}

// create Escpos printer
func New(dst io.ReadWriter) *Escpos {
	e := &Escpos{dst: dst}
	return e.reset()
}

// create Escpos printer using an internal buffer
func NewBuffer() *Escpos {
	e := &Escpos{}
	e.dst = &e.buffer
	return e.reset()
}

// Buffer returns the underlying buffer. Valid only if NewBufer() is used to create the printer
func (e *Escpos) Buffer() *bytes.Buffer {
	return &e.buffer
}

// write raw bytes to printer
func (e *Escpos) WriteRaw(data []byte) *Escpos {
	if len(data) > 0 {
		_, _ = e.dst.Write(data)
	}

	return e
}

func (e *Escpos) WriteByte(data ...byte) *Escpos {
	return e.WriteRaw(data)
}

// write a string to the printer
func (e *Escpos) Write(data string) *Escpos {
	return e.WriteRaw([]byte(data))
}

func (e *Escpos) Writef(format string, a ...interface{}) *Escpos {
	return e.Write(fmt.Sprintf(format, a...))
}

func (e *Escpos) Writeln(data string) *Escpos {
	return e.Write(data).Linefeed()
}

func (e *Escpos) WriteRepeat(count int, data ...byte) *Escpos {
	return e.WriteRaw(bytes.Repeat(data, count))
}

// init/reset printer settings
func (e *Escpos) Init() *Escpos {
	return e.reset().Write("\x1B@")
}

// end output
func (e *Escpos) End() *Escpos {
	return e.Write("\xFA")
}

// send cut
func (e *Escpos) Cut() *Escpos {
	return e.Write("\x1DVA0")
}

// send cut minus one point (partial cut)
func (e *Escpos) CutPartial() *Escpos {
	return e.WriteRaw([]byte{GS, 0x56, 1})
}

// send cash
func (e *Escpos) Cash() *Escpos {
	return e.Write("\x1B\x70\x00\x0A\xFF")
}

// send linefeed
func (e *Escpos) Linefeed() *Escpos {
	return e.Write("\n")
}

// send N formfeeds
func (e *Escpos) FormfeedN(n int) *Escpos {
	return e.Writef("\x1Bd%c", n)
}

// send formfeed
func (e *Escpos) Formfeed() *Escpos {
	return e.FormfeedN(1)
}

// set font
func (e *Escpos) SetFont(font string) *Escpos {
	f := 0

	switch font {
	case "A":
		f = 0
	case "B":
		f = 1
	case "C":
		f = 2
	default:
		f = 0
	}

	return e.Writef("\x1BM%c", f)
}

func (e *Escpos) SendFontSize() *Escpos {
	return e.Writef("\x1D!%c", ((e.width-1)<<4)|(e.height-1))
}

// set font size
func (e *Escpos) SetFontSize(width, height uint8) *Escpos {
	if width > 0 && height > 0 && width <= 8 && height <= 8 {
		e.width = width
		e.height = height
		e.SendFontSize()
	}
	return e
}

func (e *Escpos) SetLineSpacing(spacing uint8) *Escpos {
	return e.Writef("\x1B3%c", spacing)
}

// send underline
func (e *Escpos) SendUnderline() *Escpos {
	return e.Writef("\x1B-%c", e.underline)
}

// send emphasize / doublestrike
func (e *Escpos) SendEmphasize() *Escpos {
	return e.Writef("\x1BG%c", e.emphasize)
}

// send upsidedown
func (e *Escpos) SendUpsidedown() *Escpos {
	return e.Writef("\x1B{%c", e.upsidedown)
}

// send rotate
func (e *Escpos) SendRotate() *Escpos {
	return e.Writef("\x1BR%c", e.rotate)
}

// send reverse
func (e *Escpos) SendReverse() *Escpos {
	return e.Writef("\x1DB%c", e.reverse)
}

// send smooth
func (e *Escpos) SendSmooth() *Escpos {
	return e.Writef("\x1Db%c", e.smooth)
}

// send move x
func (e *Escpos) SendMoveX(x uint16) *Escpos {
	return e.Write(string([]byte{0x1b, 0x24, byte(x % 256), byte(x / 256)}))
}

// send move y
func (e *Escpos) SendMoveY(y uint16) *Escpos {
	return e.Write(string([]byte{0x1d, 0x24, byte(y % 256), byte(y / 256)}))
}

// set underline
func (e *Escpos) SetUnderline(v uint8) *Escpos {
	e.underline = v
	return e.SendUnderline()
}

// set emphasize
func (e *Escpos) SetEmphasize(u uint8) *Escpos {
	e.emphasize = u
	return e.SendEmphasize()
}

// set upsidedown
func (e *Escpos) SetUpsidedown(v uint8) *Escpos {
	e.upsidedown = v
	return e.SendUpsidedown()
}

// set rotate
func (e *Escpos) SetRotate(v uint8) *Escpos {
	e.rotate = v
	return e.SendRotate()
}

// set reverse
func (e *Escpos) SetReverse(v uint8) *Escpos {
	e.reverse = v
	return e.SendReverse()
}

// set smooth
func (e *Escpos) SetSmooth(v uint8) *Escpos {
	e.smooth = v
	return e.SendSmooth()
}

// pulse (open the drawer)
func (e *Escpos) Pulse() *Escpos {
	// with t=2 -- meaning 2*2msec
	return e.Write("\x1Bp\x02")
}

// set alignment
func (e *Escpos) SetAlign(align string) *Escpos {
	a := 0
	switch align {
	case "left":
		a = 0
	case "center":
		a = 1
	case "right":
		a = 2
	default:
		log.Fatalf("Invalid alignment: %s", align)
	}
	return e.Writef("\x1Ba%c", a)
}

// set language -- ESC R
func (e *Escpos) SetLang(lang string) *Escpos {
	l := 0

	switch lang {
	case "en":
		l = 0
	case "fr":
		l = 1
	case "de":
		l = 2
	case "uk":
		l = 3
	case "da":
		l = 4
	case "sv":
		l = 5
	case "it":
		l = 6
	case "es":
		l = 7
	case "ja":
		l = 8
	case "no":
		l = 9
	default:
		log.Fatalf("Invalid language: %s", lang)
	}
	return e.Writef("\x1BR%c", l)
}

// do a block of text
func (e *Escpos) Text(params map[string]string, data string) *Escpos {

	// send alignment to printer
	if align, ok := params["align"]; ok {
		e.SetAlign(align)
	}

	// set lang
	if lang, ok := params["lang"]; ok {
		e.SetLang(lang)
	}

	// set smooth
	if smooth, ok := params["smooth"]; ok && (smooth == "true" || smooth == "1") {
		e.SetSmooth(1)
	}

	// set emphasize
	if em, ok := params["em"]; ok && (em == "true" || em == "1") {
		e.SetEmphasize(1)
	}

	// set underline
	if ul, ok := params["ul"]; ok && (ul == "true" || ul == "1") {
		e.SetUnderline(1)
	}

	// set reverse
	if reverse, ok := params["reverse"]; ok && (reverse == "true" || reverse == "1") {
		e.SetReverse(1)
	}

	// set rotate
	if rotate, ok := params["rotate"]; ok && (rotate == "true" || rotate == "1") {
		e.SetRotate(1)
	}

	// set font
	if font, ok := params["font"]; ok {
		e.SetFont(strings.ToUpper(font[5:6]))
	}

	// do dw (double font width)
	if dw, ok := params["dw"]; ok && (dw == "true" || dw == "1") {
		e.SetFontSize(2, e.height)
	}

	// do dh (double font height)
	if dh, ok := params["dh"]; ok && (dh == "true" || dh == "1") {
		e.SetFontSize(e.width, 2)
	}

	// do font width
	if width, ok := params["width"]; ok {
		if i, err := strconv.Atoi(width); err == nil {
			e.SetFontSize(uint8(i), e.height)
		} else {
			log.Fatalf("Invalid font width: %s", width)
		}
	}

	// do font height
	if height, ok := params["height"]; ok {
		if i, err := strconv.Atoi(height); err == nil {
			e.SetFontSize(e.width, uint8(i))
		} else {
			log.Fatalf("Invalid font height: %s", height)
		}
	}

	// do y positioning
	if x, ok := params["x"]; ok {
		if i, err := strconv.Atoi(x); err == nil {
			e.SendMoveX(uint16(i))
		} else {
			log.Fatalf("Invalid x param %s", x)
		}
	}

	// do y positioning
	if y, ok := params["y"]; ok {
		if i, err := strconv.Atoi(y); err == nil {
			e.SendMoveY(uint16(i))
		} else {
			log.Fatalf("Invalid y param %s", y)
		}
	}

	// do text replace, then write data
	data = textReplace(data)
	if len(data) > 0 {
		return e.Write(data)
	}

	return e
}

// feed the printer
func (e *Escpos) Feed(params map[string]string) *Escpos {
	// handle lines (form feed X lines)
	if l, ok := params["line"]; ok {
		if i, err := strconv.Atoi(l); err == nil {
			e.FormfeedN(i)
		}
	}

	// handle units (dots)
	if u, ok := params["unit"]; ok {
		if i, err := strconv.Atoi(u); err == nil {
			e.SendMoveY(uint16(i))
		}
	}

	// send linefeed
	return e.Linefeed().
		// reset variables
		reset().
		// reset printer
		SendEmphasize().
		SendRotate().
		SendSmooth().
		SendReverse().
		SendUnderline().
		SendUpsidedown().
		SendFontSize().
		SendUnderline()
}

// feed and cut based on parameters
func (e *Escpos) FeedAndCut(params map[string]string) *Escpos {
	if t, ok := params["type"]; ok && t == "feed" {
		e.Formfeed()
	}

	return e.Cut()
}

// Barcode sends a barcode to the printer.
func (e *Escpos) Barcode(barcode string, format int) *Escpos {
	code := ""
	switch format {
	case 0:
		code = "\x00"
	case 1:
		code = "\x01"
	case 2:
		code = "\x02"
	case 3:
		code = "\x03"
	case 4:
		code = "\x04"
	case 73:
		code = "\x49"
	}

	// reset settings
	e.reset()

	// set align
	e.SetAlign("center")

	// write barcode
	if format > 69 {
		return e.Writef("\x1dk"+code+"%v%v", len(barcode), barcode)
	} else if format < 69 {
		return e.Writef("\x1dk"+code+"%v\x00", barcode)
	}
	return e.Writef("%v", barcode)
}

// used to send graphics headers
func (e *Escpos) gSend(m byte, fn byte, data []byte) *Escpos {
	l := len(data) + 2

	return e.Write("\x1b(L").
		WriteRaw([]byte{byte(l % 256), byte(l / 256), m, fn}).
		WriteRaw(data)
}

// write an image
func (e *Escpos) Image(params map[string]string, data string) *Escpos {
	// send alignment to printer
	if align, ok := params["align"]; ok {
		e.SetAlign(align)
	}

	// get width
	wstr, ok := params["width"]
	if !ok {
		return e
	}

	// get height
	hstr, ok := params["height"]
	if !ok {
		return e
	}

	// convert width
	_, err := strconv.Atoi(wstr)
	if err != nil {
		return e
	}

	// convert height
	_, err = strconv.Atoi(hstr)
	if err != nil {
		return e
	}

	// decode data frome b64 string
	dec, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return e
	}

	//log.Printf("Image len:%d w: %d h: %d\n", len(dec), width, height)

	// $imgHeader = self::dataHeader(array($img -> getWidth(), $img -> getHeight()), true);
	// $tone = '0';
	// $colors = '1';
	// $xm = (($size & self::IMG_DOUBLE_WIDTH) == self::IMG_DOUBLE_WIDTH) ? chr(2) : chr(1);
	// $ym = (($size & self::IMG_DOUBLE_HEIGHT) == self::IMG_DOUBLE_HEIGHT) ? chr(2) : chr(1);
	//
	// $header = $tone . $xm . $ym . $colors . $imgHeader;
	// $this -> graphicsSendData('0', 'p', $header . $img -> toRasterFormat());
	// $this -> graphicsSendData('0', '2');

	header := []byte{
		byte('0'), 0x01, 0x01, byte('1'),
	}

	a := append(header, dec...)

	return e.gSend(byte('0'), byte('p'), a).
		gSend(byte('0'), byte('2'), []byte{})

}

// write a "node" to the printer
func (e *Escpos) WriteNode(name string, params map[string]string, data string) *Escpos {
	switch name {
	case "text":
		return e.Text(params, data)
	case "feed":
		return e.Feed(params)
	case "cut":
		return e.FeedAndCut(params)
	case "pulse":
		return e.Pulse()
	case "image":
		return e.Image(params, data)
	default:
		return e
	}
}
