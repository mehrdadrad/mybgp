package main

import (
	"flag"
	"fmt"
	"github.com/osrg/gobgp/packet/mrt"
	"io"
	"os"
)

func exportMrt(filename string) error {

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

		fmt.Printf("%+v", msg)
	}

	return nil

}
func main() {
	mrtFile := flag.String("mrtfile", "", "enter the full MRT path")
	flag.Parse()
	exportMrt(*mrtFile)
}
