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
	idleOutput        machine.Pin
}

type Gates struct {
	outerGate       *Gate
	innerGate       *Gate
	inboundCycling  machine.Pin
	outboundCycling machine.Pin
	debugLED        machine.Pin
}

func main() {
	outerGate := &Gate{
		Name:              "Outer",
		openRequestInput:  machine.D2,
		stuckRequestInput: machine.NoPin,
		closedInput:       machine.D4,
		closedOutput:      machine.D5,
		openRequestOutput: machine.D6,
		idleOutput:        machine.ADC3,
	}
	innerGate := &Gate{
		Name:              "Inner",
		openRequestInput:  machine.D8,
		stuckRequestInput: machine.NoPin,
		closedInput:       machine.D10,
		closedOutput:      machine.D11,
		openRequestOutput: machine.D12,
		idleOutput:        machine.ADC4,
	}
	gates := Gates{
		outerGate:       outerGate,
		innerGate:       innerGate,
		inboundCycling:  machine.D7,
		outboundCycling: machine.ADC0,
		debugLED:        machine.LED,
	}
	gates.Init()
	go func() {
		for {
			gates.setStatusLEDs()
			time.Sleep(100 * time.Millisecond)
		}
	}()
	for {
		if outerGate.isClosed() && innerGate.isClosed() {
			if outerGate.checkOpenRequestInput() {
				gates.Inbound()
			}
			if innerGate.checkOpenRequestInput() {
				gates.Outbound()
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (gates *Gates) Inbound() {
	gates.inboundCycling.High()
	gates.cycle(gates.outerGate, gates.innerGate)
	gates.inboundCycling.Low()
}

func (gates *Gates) Outbound() {
	gates.outboundCycling.High()
	gates.cycle(gates.innerGate, gates.outerGate)
	gates.outboundCycling.Low()
}

func (gates *Gates) cycle(gate1 *Gate, gate2 *Gate) {
	ticks := 0
	//Request first gate opens
	gate1.requestOpen()
	//Wait for the gate to be open
	for gate1.isClosed() {
		ticks++
		//Timeout after about 5 seconds
		if ticks > 50 {
			gate1.idleOutput.High()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	ticks = 0
	//The first gate has started opening, wait for it to close again
	for !gate1.isClosed() {
		ticks++
		//Timeout after about 2 minutes
		if ticks > 1200 {
			gate1.idleOutput.High()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	ticks = 0
	//The first gate has closed, open the second gate
	gate1.idleOutput.High()
	gate2.requestOpen()
	//Wait for the second gate to open
	for gate2.isClosed() {
		ticks++
		//Timeout after about 5 seconds
		if ticks > 50 {
			gate2.idleOutput.High()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	ticks = 0
	//The second gate has opened, wait for it to close
	for !gate2.isClosed() {
		ticks++
		//Timeout after about 2 minutes
		if ticks > 1200 {
			gate2.idleOutput.High()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	//Second gate has closed, cycle complete
	gate2.idleOutput.High()
	return
}

func (gates *Gates) Init() {
	gates.innerGate.Init()
	gates.outerGate.Init()
	gates.inboundCycling.Configure(machine.PinConfig{Mode: machine.PinOutput})
	gates.outboundCycling.Configure(machine.PinConfig{Mode: machine.PinOutput})
	gates.inboundCycling.Low()
	gates.outboundCycling.Low()
}

func (gates *Gates) setStatusLEDs() {
	gates.innerGate.closedOutput.Set(gates.innerGate.closedInput.Get())
	gates.outerGate.closedOutput.Set(gates.outerGate.closedInput.Get())
}

func (g *Gate) Init() {
	g.openRequestInput.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	g.stuckRequestInput.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	g.closedInput.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	g.closedOutput.Configure(machine.PinConfig{Mode: machine.PinOutput})
	g.openRequestOutput.Configure(machine.PinConfig{Mode: machine.PinOutput})
	g.idleOutput.Configure(machine.PinConfig{Mode: machine.PinOutput})
	g.openRequestOutput.Low()
	g.closedOutput.Low()
	g.idleOutput.High()
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
	g.idleOutput.Low()
}
