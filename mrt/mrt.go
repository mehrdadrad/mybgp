// This is a project to find out how we can parse
// and export entire BGP table. it already exports to the json
// format and it's in active developement right now.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/osrg/gobgp/packet/mrt"
	"gopkg.in/cheggaaa/pb.v1"
	"gopkg.in/mgo.v2"
	"io"
	"log"
	"os"
	"sync"
)

var (
	mrtFile string
	format  string
	expFile string
	bar     *pb.ProgressBar
	pct     int = 0
)

// Export MTR func based on the GoBGP library
func exportMrt(filename string, output chan *mrt.MRTMessage) error {
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
		//fmt.Printf("%+v", msg)
		output <- msg

		for i := 0; i < ((int((bytesRead * 100) / totalBytes)) - pct); i++ {
			bar.Increment()
			pct++
		}

	}

	close(output)
	bar.FinishPrint("Exported!")
	return nil

}

// Init args and progress
func init() {
	flag.StringVar(&mrtFile, "mrtfile", "", "enter the full MRT path")
	flag.StringVar(&format, "format", "json", "export format")
	flag.StringVar(&expFile, "jsonfile", "export_mrt.json", "export file full path")
	flag.Parse()

	if mrtFile == "" {
		log.Fatal("The MRT file not specified!")
	}

	bar = pb.New(100)
	bar.SetWidth(80)
	bar.SetMaxWidth(80)
	bar.ShowTimeLeft = false
	bar.Start()
}

func main() {
	ch := make(chan *mrt.MRTMessage)
	go exportMrt(mrtFile, ch)

	switch format {
	case "json":
		var once sync.Once
		f, err := os.Create(expFile)
		if err != nil {
			fmt.Errorf("%s", err)
		}
		f.WriteString("[")
		for r := range ch {
			d, err := json.Marshal(r.Body)
			if err != nil {
				fmt.Errorf("marshal failed %s", err)
			}
			once.Do(func() {
				f.Write(d)
			})
			f.WriteString(",")
			f.Write(d)
			f.Sync()
		}
		f.WriteString("]")
		f.Close()
	case "mongo":
		session, err := mgo.Dial("localhost")
		if err != nil {
			panic(err)
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)
		c := session.DB("bgp").C("dump")

		for r := range ch {
			err = c.Insert(r.Body)
			if err != nil {
				log.Fatal(err)
			}
		}

	}

}
