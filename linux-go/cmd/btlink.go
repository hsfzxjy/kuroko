package cmd

import (
	"fmt"
	"kuroko-linux/cmd/tui"
	"kuroko-linux/internal/discovery"
	"kuroko-linux/internal/rpc"
	"kuroko-linux/util"

	"github.com/hsfzxjy/go-srpc"
)

type btlinkCommand struct{}

type ProbeResultData struct {
	result discovery.ProbeResult
	state  discovery.LinkState
	Err    error
	render func()
}

func (d *ProbeResultData) SetRenderFunc(render func()) {
	d.render = render
}

func (d *ProbeResultData) GetTexts() (string, string) {
	name := d.result.DisplayName()
	if len(name) == 0 {
		name = "<UNKNOWN>"
	}
	mainText := " + [darkgrey::bu]Device[-:-:-]: " + name
	secText := fmt.Sprintf(
		"     [darkgrey::bu]Address[-:-:-]: [lightpink]%s[-:-:-] [darkgrey::bu]State[-:-:-]: %s",
		d.result.DisplayAddr(),
		d.state.String(),
	)
	return mainText, secText
}

func (d *ProbeResultData) StartLink(client *srpc.Client) {
	hh, err := client.CallStream("LinkerManager.StartLink", &discovery.LinkerArg{
		Typ:    "bluetooth",
		Result: d.result,
	})
	if err != nil {
		panic(err)
	}
	for st := range hh.C() {
		d.state = st.(discovery.LinkState)
		d.render()
	}
}

var stopApp func()

type App struct {
	sa *tui.SelectorApp
}

func (c *btlinkCommand) Execute(args []string) error {
	client, err := rpc.Client()
	if err != nil {
		return err
	}
	h, _ := client.CallStream("ProberManager.Start", "bluetooth")
	
	app := tui.NewSelectorApp()
	stopApp = app.Stop
	go util.OnInterrupt(func() { h.Cancel(); app.Stop() })
	defer h.Cancel()

	go func() {
		for res := range h.C() {
			var res = res.(discovery.ProbeResult)
			var data = &ProbeResultData{result: res, state: res.LinkState()}
			app.AddItem(
				data,
				func() {
					h.Cancel()
					go data.StartLink(client)
				},
			)
		}
		h.Cancel()
	}()

	app.Run()
	return h.CancelAndResult()
}

func init() {
	CLIParser.AddCommand("btlink", "Link a bluetooth peer", "", &btlinkCommand{})
}
