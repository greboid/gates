package main

import (
	"machine"
)

type Gate struct {
	Name              string
	openRequestInput  machine.Pin
	stuckRequestInput machine.Pin
	closedInput       machine.Pin
	closedOutput      machine.Pin
	openRequestOutput machine.Pin
}

type Gates struct {
	Gate1        *Gate
	Gate2        *Gate
	startingGate machine.Pin
	endingGate   machine.Pin
}

func main() {
	outerGate := &Gate{
		Name:              "1",
		openRequestInput:  machine.D2,
		stuckRequestInput: machine.NoPin,
		closedInput:       machine.D4,
		closedOutput:      machine.D5,
		openRequestOutput: machine.D6,
	}
	innerGate := &Gate{
		Name:              "2",
		openRequestInput:  machine.D8,
		stuckRequestInput: machine.NoPin,
		closedInput:       machine.D10,
		closedOutput:      machine.D11,
		openRequestOutput: machine.D12,
	}
	gates := Gates{
		Gate1:        outerGate,
		Gate2:        innerGate,
		startingGate: machine.D7,
		endingGate:   machine.D13,
	}
	gates.Init()
	for {
		gates.setStatusLEDs()
		if outerGate.isClosed() && innerGate.isClosed() {
			if outerGate.checkOpenRequestInput() {
				inbound := Gates{
					Gate1:        outerGate,
					Gate2:        innerGate,
					startingGate: machine.D7,
					endingGate:   machine.D13,
				}
				inbound.Open()
			}
			if innerGate.checkOpenRequestInput() {
				outbound := Gates{
					Gate1:        innerGate,
					Gate2:        outerGate,
					startingGate: machine.D13,
					endingGate:   machine.D7,
				}
				outbound.Open()
			}
		}
	}
}

func (gates *Gates) Open() {
	for {
		gates.setStatusLEDs()
		gates.startingGate.High()
		gates.Gate1.openRequestOutput.High()
		for gates.Gate1.isClosed() {
			gates.setStatusLEDs()
			continue
		}
		for !gates.Gate1.isClosed() {
			gates.setStatusLEDs()
			gates.Gate1.openRequestOutput.Low()
		}
		gates.Gate2.openRequestOutput.High()
		for gates.Gate2.isClosed() {
			gates.setStatusLEDs()
			continue
		}
		for !gates.Gate2.isClosed() {
			gates.setStatusLEDs()
			gates.Gate2.openRequestOutput.Low()
		}
		gates.startingGate.Low()
		return
	}
}

func (gates *Gates) Init() {
	gates.Gate1.Init()
	gates.Gate2.Init()
	gates.startingGate.Configure(machine.PinConfig{Mode: machine.PinOutput})
	gates.endingGate.Configure(machine.PinConfig{Mode: machine.PinOutput})
	gates.startingGate.Low()
	gates.endingGate.Low()
}

func (gates *Gates) setStatusLEDs() {
	gates.Gate1.closedOutput.Set(gates.Gate1.closedInput.Get())
	gates.Gate2.closedOutput.Set(gates.Gate2.closedInput.Get())
}

func (g *Gate) Init() {
	g.openRequestInput.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	g.stuckRequestInput.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	g.closedInput.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	g.closedOutput.Configure(machine.PinConfig{Mode: machine.PinOutput})
	g.openRequestOutput.Configure(machine.PinConfig{Mode: machine.PinOutput})
	g.openRequestOutput.Low()
	g.closedOutput.Low()
}

func (g *Gate) isClosed() bool {
	return g.closedInput.Get()
}

func (g *Gate) checkOpenRequestInput() bool {
	return !g.openRequestInput.Get()
}
