package main

import (
	"machine"
	"time"
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
	debugLED     machine.Pin
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
		endingGate:   machine.ADC0,
		debugLED:     machine.LED,
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
					endingGate:   machine.ADC0,
					debugLED:     machine.LED,
				}
				inbound.Open()
			}
			if innerGate.checkOpenRequestInput() {
				outbound := Gates{
					Gate1:        innerGate,
					Gate2:        outerGate,
					startingGate: machine.ADC0,
					endingGate:   machine.D7,
					debugLED:     machine.LED,
				}
				outbound.Open()
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (gates *Gates) Open() {
	ticks := 0
	gates.setStatusLEDs()
	gates.startingGate.High()
	//Request first gate opens
	gates.Gate1.requestOpen()
	//Wait for the gate to be open
	for gates.Gate1.isClosed() {
		gates.setStatusLEDs()
		ticks++
		//Timeout after about 5 seconds
		if ticks > 50 {
			gates.startingGate.Low()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	ticks = 0
	//The first gate has started opening, wait for it to close again
	for !gates.Gate1.isClosed() {
		gates.setStatusLEDs()
		ticks++
		//Timeout after about 2 minutes
		if ticks > 120000 {
			gates.startingGate.Low()
			return
		}
		gates.Gate1.openRequestOutput.Low()
		time.Sleep(100 * time.Millisecond)
	}
	ticks = 0
	//The first gate has closed, open the second gate
	gates.Gate2.requestOpen()
	//Wait for the second gate to open
	for gates.Gate2.isClosed() {
		gates.setStatusLEDs()
		ticks++
		//Timeout after about 5 seconds
		if ticks > 50 {
			gates.startingGate.Low()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	ticks = 0
	//The second gate has opened, wait for it to close
	for !gates.Gate2.isClosed() {
		gates.setStatusLEDs()
		ticks++
		//Timeout after about 2 minutes
		if ticks > 120000 {
			gates.startingGate.Low()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	//Second gate has closed, cycle complete
	gates.startingGate.Low()
	return
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

func (g *Gate) requestOpen() {
	g.openRequestOutput.High()
	time.Sleep(500 * time.Millisecond)
	g.openRequestOutput.Low()
}
