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
	workingOutput     machine.Pin
	openOutput        machine.Pin
}

type Controller struct {
	inbound         bool
	outbound        bool
	gate1Opened     bool
	gate1Closed     bool
	gate2Opened     bool
	gate2Closed     bool
	outerGate       *Gate
	innerGate       *Gate
	inboundCycling  machine.Pin
	outboundCycling machine.Pin
	gate1           *Gate
	gate2           *Gate
	cycleTicks      int
}

func main() {
	outerGate := &Gate{
		Name:              "Outer",
		openRequestInput:  machine.D2,
		stuckRequestInput: machine.NoPin,
		closedInput:       machine.D4,
		closedOutput:      machine.D5,
		openRequestOutput: machine.D6,
		workingOutput:     machine.ADC3,
		openOutput:        machine.ADC1,
	}
	innerGate := &Gate{
		Name:              "Inner",
		openRequestInput:  machine.D8,
		stuckRequestInput: machine.NoPin,
		closedInput:       machine.D10,
		closedOutput:      machine.D11,
		openRequestOutput: machine.D12,
		workingOutput:     machine.ADC4,
		openOutput:        machine.ADC2,
	}
	gates := &Controller{
		gate1Opened:     false,
		gate1Closed:     false,
		gate2Opened:     false,
		gate2Closed:     false,
		outerGate:       outerGate,
		innerGate:       innerGate,
		inboundCycling:  machine.D7,
		outboundCycling: machine.ADC0,
	}
	gates.Init()
	for {
		//Set the status LEDs
		gates.setStatusLEDs()
		if gates.gate1 != nil && gates.gate2 != nil {
			if !gates.gate1Opened {
				if gates.cycleTicks == 0 {
					println("Opening gate1")
					gates.gate1.openRequestOutput.High()
					gates.gate1.workingOutput.High()
				} else if gates.gate1.isOpen() {
					println("Gate1 open")
					gates.gate1Opened = true
					gates.gate1.openRequestOutput.Low()
				} else if gates.cycleTicks >= 50 {
					println("Timing out gate1 opening")
					gates.Reset()
				}
			} else if !gates.gate1Closed {
				if gates.gate1.isClosed() {
					println("Gate1 closed")
					gates.gate1Closed = true
					gates.cycleTicks = -1
					gates.gate1.workingOutput.Low()
				} else if gates.cycleTicks >= 100 {
					println("Timing out gate1 closing")
					gates.Reset()
				}
			} else if gates.gate1Closed && !gates.gate2Opened {
				if gates.cycleTicks == 0 {
					println("Opening gate2")
					gates.gate2.openRequestOutput.High()
					gates.gate2.workingOutput.High()
				} else if gates.gate2.isOpen() {
					println("Gate2 open")
					gates.gate2Opened = true
					gates.gate2.openRequestOutput.Low()
				} else if gates.cycleTicks >= 50 {
					println("Timing out gate2 opening")
					gates.Reset()
				}
			} else if !gates.gate2Closed {
				if gates.gate2.isClosed() {
					println("Gate2 closed")
					gates.gate2Closed = true
					gates.cycleTicks = -1
					gates.gate2.workingOutput.Low()
				} else if gates.cycleTicks >= 100 {
					println("Timing out gate2 closing")
					gates.Reset()
				}
			} else if gates.gate2Closed {
				println("Cycle finished")
				gates.Reset()
			}
			gates.cycleTicks++
		} else if outerGate.checkOpenRequestInput() && innerGate.isClosed() {
			println("Starting inbound cycle")
			gates.Reset()
			gates.inbound = true
			gates.gate1 = outerGate
			gates.gate2 = innerGate
		} else if innerGate.checkOpenRequestInput() && outerGate.isClosed() {
			println("Starting outbound cycle")
			gates.Reset()
			gates.outbound = true
			gates.gate1 = innerGate
			gates.gate2 = outerGate
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (gates *Controller) Init() {
	gates.inboundCycling.Configure(machine.PinConfig{Mode: machine.PinOutput})
	gates.outboundCycling.Configure(machine.PinConfig{Mode: machine.PinOutput})
	gates.inboundCycling.Low()
	gates.outboundCycling.Low()
	gates.innerGate.Init()
	gates.outerGate.Init()
}

func (gates *Controller) Reset() {
	gates.gate2Opened = false
	gates.gate2Closed = false
	gates.gate1Opened = false
	gates.gate1Closed = false
	gates.gate1 = nil
	gates.gate2 = nil
	gates.inbound = false
	gates.outbound = false
	gates.cycleTicks = 0
	gates.innerGate.Reset()
	gates.outerGate.Reset()
}

func (g *Gate) Init() {
	g.openRequestInput.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	g.stuckRequestInput.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	g.closedInput.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	g.closedOutput.Configure(machine.PinConfig{Mode: machine.PinOutput})
	g.openRequestOutput.Configure(machine.PinConfig{Mode: machine.PinOutput})
	g.workingOutput.Configure(machine.PinConfig{Mode: machine.PinOutput})
	g.openRequestOutput.Low()
	g.workingOutput.Low()
}

func (g *Gate) Reset() {
	g.openRequestOutput.Low()
	g.workingOutput.Low()
}

func (gates *Controller) setStatusLEDs() {
	gates.innerGate.closedOutput.Set(gates.innerGate.closedInput.Get())
	gates.outerGate.closedOutput.Set(gates.outerGate.closedInput.Get())
	gates.innerGate.openOutput.Set(!gates.innerGate.closedInput.Get())
	gates.outerGate.openOutput.Set(!gates.outerGate.closedInput.Get())
	gates.inboundCycling.Set(gates.inbound)
	gates.outboundCycling.Set(gates.outbound)
}

func (g *Gate) isClosed() bool {
	return g.closedInput.Get()
}

func (g *Gate) isOpen() bool {
	return !g.closedInput.Get()
}

func (g *Gate) checkOpenRequestInput() bool {
	return !g.openRequestInput.Get()
}
