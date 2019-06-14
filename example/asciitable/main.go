package main

import (
	"flag"
	"github.com/peter-mount/escpos"
	"github.com/peter-mount/go-ipp"
	"log"
)

var (
	host    = flag.String("h", "localhost", "IPP Host")
	port    = flag.Int("p", 631, "IPP Port")
	printer = flag.String("d", "", "IPP Printer Name")
	user    = flag.String("user", "", "IPP Username")
	pass    = flag.String("pass", "", "IPP Password")
	useTls  = flag.Bool("tls", true, "Use TLS")
	jobName = flag.String("j", "asciitable", "Job name")
)

func main() {
	flag.Parse()

	p := escpos.NewBuffer().
		Init().
		SetSmooth(1).
		SetFontSize(2, 3).
		SetFont("A").
		Write("ASCII TABLE").
		Linefeed().
		Linefeed()

	repeatCount := (16-3)*2 + 1
	p.SetFont("A").
		SetLineSpacing(26).
		Linefeed().
		Write("  \xDA").
		WriteRepeat(repeatCount, '\xC4').
		WriteByte('\xBF').
		Linefeed().
		Write("  \xB3 3 4 5 6 7 8 9 A B C D E F \xB3").
		Linefeed().
		Write("  \xC3").
		WriteRepeat(repeatCount, '\xC4').
		WriteByte('\xB3').
		Linefeed()

	for y := 0; y < 16; y++ {
		p.Writef("%x \xB3 ", y)
		for x := 3; x < 16; x++ {
			p.WriteByte(byte((x*16)+y), ' ')
		}
		p.WriteByte('\xB3').
			Linefeed()
	}

	p.Write("  \xC0").
		WriteRepeat(repeatCount, '\xC4').
		WriteByte('\xD9').
		Linefeed().
		FormfeedN(2).
		Cut().
		End()

	client := ipp.NewIPPClient(*host, *port, *user, *pass, *useTls)

	buffer := p.Buffer()

	doc := ipp.Document{
		Document: buffer,
		Name:     *jobName,
		Size:     buffer.Len(),
		MimeType: ipp.MimeTypeOctetStream,
	}

	jobAttributes := make(map[string]interface{})
	jobAttributes[ipp.OperationAttributeJobName] = *jobName

	jobID, err := client.PrintJob(doc, *printer, jobAttributes)
	if err != nil {
		log.Fatal("Failed to print", err)
	}
	log.Println("Submitted job", jobID)
}
