package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/osrg/gobgp/packet/mrt"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"os"
	"sync"
)

var (
	mrtFile string
	format  string
	bar     *pb.ProgressBar
	pct     int = 0
)

func exportMrt(filename string, output chan string) error {
	var bytesRead int64
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %s", err)
	}
	stat, _ := file.Stat()
	totalBytes := stat.Size()

	for {
		buf := make([]byte, mrt.MRT_COMMON_HEADER_LEN)
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Errorf("failed to read: %s", err)
		}

		bytesRead += int64(n)
		h := &mrt.MRTHeader{}
		err = h.DecodeFromBytes(buf)
		if err != nil {
			fmt.Errorf("failed to parse")
		}

		buf = make([]byte, h.Len)
		n, err = file.Read(buf)
		if err != nil {
			fmt.Errorf("failed to read")
		}

		bytesRead += int64(n)
		msg, err := mrt.ParseMRTBody(h, buf)
		if err != nil {
			fmt.Errorf("failed to parse: %s", err)
		}
		d, err := json.Marshal(msg)
		if err != nil {
			fmt.Errorf("marshal failed %s", err)
		}

		output <- string(d)

		for i := 0; i < ((int((bytesRead * 100) / totalBytes)) - pct); i++ {
			bar.Increment()
			pct++
		}

		//pct = int((bytesRead * 100) / totalBytes)
	}

	close(output)
	bar.FinishPrint("Exported!")
	return nil

}

func init() {
	flag.StringVar(&mrtFile, "mrtfile", "", "enter the full MRT path")
	flag.StringVar(&format, "format", "json", "export format")
	flag.Parse()

	bar = pb.New(100)
	bar.SetWidth(80)
	bar.SetMaxWidth(80)
	bar.ShowTimeLeft = false
	bar.Start()
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
