package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/pflag"
	"go.bug.st/serial"
)

func main() {
	inputfile := pflag.String("input", "", "input file")
	maxchunk := pflag.Int("chunk", 16, "maximum number of bytes to send in a chunk")
	port := pflag.String("port", "", "output serial port")
	baud := pflag.Int("baud", 9600, "baud rate")
	stopbits := pflag.String("stopbits", "1", "stop bits (1, 1.5 or 2)")
	bits := pflag.Int("bits", 8, "data bits")
	parity := pflag.String("parity", "N", "parity N/O/E (none, odd, even)")
	dtr := pflag.Bool("dtr", true, "set DTR hardware handshake signalling")
	rts := pflag.Bool("rts", true, "set RTS hardware handshake signalling")

	dsr := pflag.Bool("dsr", true, "use DSR hardware handshake signalling")
	cts := pflag.Bool("cts", true, "use CTS hardware handshake signalling")

	pflag.Parse()

	if *inputfile == "" {
		log.Print("No input file specified, listing available ports")
		ports, err := serial.GetPortsList()
		if err != nil {
			log.Fatal(err)
		}
		if len(ports) == 0 {
			log.Fatal("No serial ports found!")
		}
		for _, port := range ports {
			fmt.Printf("Found port: %v\n", port)
		}
		return
	}

	hpglraw, err := ioutil.ReadFile(*inputfile)
	if err != nil {
		panic(err)
	}

	var sb serial.StopBits

	switch *stopbits {
	case "1":
		sb = serial.OneStopBit
	case "1.5":
		sb = serial.OnePointFiveStopBits
	case "2":
		sb = serial.TwoStopBits
	default:
		log.Fatal("Invalid stop bits: ", *stopbits)
	}

	var pb serial.Parity
	switch *parity {
	case "N":
		pb = serial.NoParity
	case "O":
		pb = serial.OddParity
	case "E":
		pb = serial.EvenParity
	default:
		log.Fatal("Invalid parity: ", *parity)
	}

	mode := serial.Mode{
		BaudRate: *baud,
		StopBits: sb,
		DataBits: *bits,
		Parity:   pb,
	}

	output, err := serial.Open(*port, &mode)
	if err != nil {
		log.Fatal("Error opening serial port: ", err)
	}

	if err := output.SetDTR(*dtr); err != nil {
		log.Fatal("Error opening DTR state: ", err)
	}
	if err := output.SetRTS(*rts); err != nil {
		log.Fatal("Error setting RTS state: ", err)
	}

	progress := progressbar.NewOptions(len(hpglraw),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("Sending data ..."),
		progressbar.OptionSetItsString("bytes"),
		progressbar.OptionThrottle(time.Second),
	)

	for {
		if *dsr {
			for {
				status, _ := output.GetModemStatusBits()
				if status.DSR {
					break
				}
				time.Sleep(5 * time.Millisecond)
			}
		}
		if *cts {
			for {
				status, _ := output.GetModemStatusBits()
				if status.CTS {
					break
				}
				time.Sleep(5 * time.Millisecond)
			}
		}
		send := len(hpglraw)
		if send > *maxchunk {
			send = *maxchunk
		}
		output.Write(hpglraw[:send])
		hpglraw = hpglraw[send:]
		progress.Add(send)

		if len(hpglraw) == 0 {
			break // done
		}
	}
	output.Close()
}
