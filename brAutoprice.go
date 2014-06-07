package main

import (
	"bufio"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/toorop/go-betarigs"
	"github.com/wsxiaoys/terminal"
	"math"
	"os"
	"time"
)

const (
	version          = "0.8"
	NL               = "\r\n"
	logTimeLayout    = "3:04:05"
	mailLoopDuration = 30
)

type display struct {
	Header []string
	Body   []string
	Footer []string
}

func (d *display) ResetBody() {
	d.Body = []string{}
}

func (d *display) BodyAddLine(line string) {
	d.Body = append(d.Body, line)
}

// Refresh refresh the display
func (d *display) Refresh() {
	terminal.Stdout.Clear().Move(0, 0)
	for _, h := range d.Header {
		fmt.Println(h)
	}
	for _, b := range d.Body {
		fmt.Println(b)
	}
	fmt.Println("")
	for _, f := range d.Footer {
		fmt.Println(f)
	}
}

func getTimeStamp() string {
	return time.Now().Local().Format(logTimeLayout)
}

// listenForUserCmd wait for a user input (char) and execute the correspondig command
func waitForExit() {
	for {
		_, _, _ = bufio.NewReader(os.Stdin).ReadLine()
		os.Exit(0)
	}
}

// Betarigs api returns an average price of the 20 cheapest rigs for a given algo
// It not optimal for our usage
// We will get the cheapest price in a first time and see if it's better
func getMarketPrice(algo uint32, rigId uint32, btr *betarigs.Betarigs) (maketPrice float64, err error) {
	rigs, err := btr.GetRigs(algo, "available", 1)
	if err != nil {
		return
	}
	if len(rigs) == 0 {
		maketPrice = 0
		return
	}
	// Preventing to return our rig price.If it's the case rig price never goes up
	for _, rig := range rigs {
		if rig.Id != rigId {
			maketPrice = rig.Price.PerSpeedUnit.Value
			break
		}
	}
	return
}

func init() {
	go waitForExit()
}

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} [options] [arguments...]

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}
OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
`
	app.Name = "brAutoprice"
	app.Usage = "brAutoprice is a tool that helps you to rent your rig at the best price."
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "apiKey", Value: "", Usage: "Your Betarig API key. (required)"},
		cli.IntFlag{Name: "rigId", Value: 0, Usage: "Your rig ID. (required)"},
		cli.Float64Flag{Name: "minPrice", Value: 0, Usage: "The min price per speed unit for your rig. brAutoprice will never set rental price below this limit. (required)"},
		cli.Float64Flag{Name: "priceDiff", Value: 0, Usage: "The diff between the market price and the price you want to apply. Ex: if you set 10, your price will be 10 percent higher than market price, if you set -10 it will be 10 percent lower."},
	}

	app.Action = func(c *cli.Context) {
		// check options
		flagOptionErrors := false
		apiKey := c.String("apiKey")
		if apiKey == "" {
			flagOptionErrors = true
			fmt.Println("ERROR : Option --apiKey is missing")
		}
		rigId := c.Int("rigId")
		if rigId == 0 {
			flagOptionErrors = true
			fmt.Println("ERROR : Option --rigId is missing")
		}
		minPrice := c.Float64("minPrice")
		if minPrice == 0 {
			flagOptionErrors = true
			fmt.Println("ERROR : Option --minPrice is missing")
		}
		if flagOptionErrors {
			fmt.Println(NL + "Usage:" + NL)
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
		priceDiff := c.Float64("priceDiff")

		// Go go go
		btr := betarigs.New(apiKey)
		dsp := display{}
		dsp.Header = []string{
			"brAutoPrice V" + version + " - (c)Toorop https://twitter.com/poroot",
			fmt.Sprintf("Rig: %d minPrice: %f  priceDiff: %g%%", rigId, minPrice, priceDiff),
			"--------------------------------------------------------",
			"",
		}
		dsp.Footer = []string{
			"--------------------------------------------------------",
			`Press "enter" to quit`,
			"Donation BTC: 1HgpsmxV52eAjDcoNpVGpYEhGfgN7mM1JB",
		}
		dsp.Refresh()

		// start main loop
		var algo betarigs.Algorithm
		var rig betarigs.Rig
		var err error
		for {
			dsp.ResetBody()
			rig, err = btr.GetRig(uint32(rigId))
			if err != nil {
				dsp.BodyAddLine(fmt.Sprintf("%s - ERR: Unable to fetch Rig lgoritm for rig %d. %v", getTimeStamp(), rigId, err))
				dsp.Refresh()
				time.Sleep(mailLoopDuration * time.Second)
				continue
			}
			dsp.BodyAddLine(fmt.Sprintf("%s - Current rig price: %f %s", getTimeStamp(), rig.Price.PerSpeedUnit.Value, rig.Price.PerSpeedUnit.Unit))
			// Get current market price
			algo, err = btr.GetAlgorithm(rig.Algorithm.Id)
			if err != nil {
				dsp.BodyAddLine(fmt.Sprintf("%s - ERR: Unable to fetch Algoritm %d. %v", getTimeStamp(), rig.Algorithm.Id, err))
				dsp.Refresh()
				time.Sleep(mailLoopDuration * time.Second)
				continue
			}
			marketPrice, err := getMarketPrice(rig.Algorithm.Id, uint32(rigId), btr)
			if err != nil {
				dsp.BodyAddLine(fmt.Sprintf("%s - ERR: Unable to get Market Price. %v", getTimeStamp(), err))
				time.Sleep(mailLoopDuration * time.Second)
				continue
			}
			dsp.BodyAddLine(fmt.Sprintf("%s - Current market price: %f %s", getTimeStamp(), marketPrice, algo.MarketPrice.Unit))

			// Change price ?
			// if market price == 0 or (me==me)
			if marketPrice != 0 && marketPrice != rig.Price.PerSpeedUnit.Value {
				newPrice := marketPrice + (priceDiff * marketPrice / 100)
				//fmt.Println(int(newPrice * 1000000))
				//fmt.Println(int(rig.Price.PerSpeedUnit.Value * 1000000))
				// Faire un diff sur la valeur abs des prix -0.1%
				if newPrice > minPrice && math.Abs(newPrice*1000000-rig.Price.PerSpeedUnit.Value*1000000) > 1 {
					success, err := btr.UpdateRigPricePerSpeedUnit(uint32(rigId), newPrice)
					if err != nil || !success {
						dsp.BodyAddLine(fmt.Sprintf("%s - ERR: Unable to update rig Price. %v", getTimeStamp(), err))
					} else {
						dsp.BodyAddLine(fmt.Sprintf("%s - Rig prince changed to newPrice: %f %s", getTimeStamp(), newPrice, algo.MarketPrice.Unit))
					}
				} else if rig.Price.PerSpeedUnit.Value != minPrice && marketPrice < minPrice {
					success, err := btr.UpdateRigPricePerSpeedUnit(uint32(rigId), minPrice)
					if err != nil || !success {
						dsp.BodyAddLine(fmt.Sprintf("%s - ERR: Unable to update rig Price. %v", getTimeStamp(), err))
					} else {
						dsp.BodyAddLine(fmt.Sprintf("%s - : Unable to update rig Price. %v", getTimeStamp(), err))
						dsp.BodyAddLine(fmt.Sprintf("%s - Rig price changed to minPrice: %f %s", getTimeStamp(), newPrice, algo.MarketPrice.Unit))
					}
				}
			}
			dsp.Refresh()
			time.Sleep(mailLoopDuration * time.Second)
		}
	}
	app.Run(os.Args)
}
