package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/osrg/gobgp/packet/mrt"
	"io"
	"os"
	"sync"
)

var (
	mrtFile string
	format  string
)

func exportMrt(filename string, output chan string) error {

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %s", err)
	}

	for {
		buf := make([]byte, mrt.MRT_COMMON_HEADER_LEN)
		_, err := file.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Errorf("failed to read: %s", err)
		}

		h := &mrt.MRTHeader{}
		err = h.DecodeFromBytes(buf)
		if err != nil {
			fmt.Errorf("failed to parse")
		}

		buf = make([]byte, h.Len)
		_, err = file.Read(buf)
		if err != nil {
			fmt.Errorf("failed to read")
		}

		msg, err := mrt.ParseMRTBody(h, buf)
		if err != nil {
			fmt.Errorf("failed to parse: %s", err)
		}
		d, err := json.Marshal(msg)
		if err != nil {
			fmt.Errorf("marshal failed %s", err)
		}

		output <- string(d)
	}

	close(output)
	return nil

}

func init() {
	flag.StringVar(&mrtFile, "mrtfile", "", "enter the full MRT path")
	flag.StringVar(&format, "format", "json", "export format")
	flag.Parse()
}

func main() {
	ch := make(chan string)
	go exportMrt(mrtFile, ch)

	switch format {
	case "json":
		var once sync.Once
		f, err := os.Create("./mybgp.json")
		if err != nil {
			fmt.Errorf("%s", err)
		}
		f.WriteString("[")
		for r := range ch {
			once.Do(func() {
				f.WriteString(r)
			})
			f.WriteString("," + r)
			f.Sync()
		}
		f.WriteString("]")
	}

}
