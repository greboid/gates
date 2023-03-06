package main

import (
	"machine"
	"time"
)

const (
	gateOpenTimeout  = 50   //In number of ticks, so varies depending on sleep interval
	gateCloseTimeout = 1200 //In number of ticks, so varies depending on sleep interval
	SleepInterval    = 100  // In milliseconds
)

type Gate struct {
	Name              string
	openRequestInput  machine.Pin
	closedInput       machine.Pin
	closedOutput      machine.Pin
	openRequestOutput machine.Pin
	workingOutput     machine.Pin
}

type Controller struct {
	stuckStarted      bool
	stuckGate         *Gate
	inbound           bool
	outbound          bool
	gate1Opened       bool
	gate1Closed       bool
	gate2Opened       bool
	gate2Closed       bool
	outerGate         *Gate
	innerGate         *Gate
	inboundCycling    machine.Pin
	outboundCycling   machine.Pin
	gate1             *Gate
	gate2             *Gate
	cycleTicks        int
	stuckRequestInput machine.Pin
}

func main() {
	outerGate := &Gate{
		Name:              "Outer",
		openRequestInput:  machine.D2,
		closedInput:       machine.D4,
		openRequestOutput: machine.ADC0,
		closedOutput:      machine.D10,
		workingOutput:     machine.D8,
	}
	innerGate := &Gate{
		Name:              "Inner",
		openRequestInput:  machine.D3,
		closedInput:       machine.D5,
		openRequestOutput: machine.ADC1,
		closedOutput:      machine.D11,
		workingOutput:     machine.D9,
	}
	gates := &Controller{
		outerGate:         outerGate,
		innerGate:         innerGate,
		inboundCycling:    machine.D6,
		outboundCycling:   machine.D7,
		stuckRequestInput: machine.D12,
	}
	gates.Init()
	for {
		gates.setStatusLEDs()
		if gates.stuckGate != nil {
			gates.handleStuckGates()
		} else if !gates.stuckRequestInput.Get() {
			gates.handleStuckRequest()
		} else if gates.gate1 != nil && gates.gate2 != nil {
			if !gates.gate1Opened {
				if gates.cycleTicks == 0 {
					println("Opening gate1")
					gates.gate1.openRequestOutput.High()
					gates.gate1.workingOutput.High()
				} else if gates.gate1.isOpen() {
					println("Gate1 open")
					gates.gate1Opened = true
					gates.gate1.openRequestOutput.Low()
				} else if gates.cycleTicks >= gateOpenTimeout {
					println("Timing out gate1 opening")
					gates.Reset()
				}
			} else if !gates.gate1Closed {
				if gates.gate1.isClosed() {
					println("Gate1 closed")
					gates.gate1Closed = true
					gates.cycleTicks = -1
					gates.gate1.workingOutput.Low()
				} else if gates.cycleTicks >= gateCloseTimeout {
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
				} else if gates.cycleTicks >= gateOpenTimeout {
					println("Timing out gate2 opening")
					gates.Reset()
				}
			} else if !gates.gate2Closed {
				if gates.gate2.isClosed() {
					println("Gate2 closed")
					gates.gate2Closed = true
					gates.cycleTicks = -1
					gates.gate2.workingOutput.Low()
				} else if gates.cycleTicks >= gateCloseTimeout {
					println("Timing out gate2 closing")
					gates.Reset()
				}
			} else if gates.gate2Closed {
				println("Cycle finished")
				gates.Reset()
			}
			gates.cycleTicks++
		} else if outerGate.checkOpenRequestInput() && innerGate.isClosed() {
			gates.startInboundCycle()
		} else if innerGate.checkOpenRequestInput() && outerGate.isClosed() {
			gates.startOutboundCycle()
		}
		time.Sleep(SleepInterval * time.Millisecond)
	}
}

func (gates *Controller) startInboundCycle() {
	println("Starting inbound cycle")
	gates.Reset()
	gates.inbound = true
	gates.gate1 = gates.outerGate
	gates.gate2 = gates.innerGate
}

func (gates *Controller) startOutboundCycle() {
	println("Starting outbound cycle")
	gates.Reset()
	gates.outbound = true
	gates.gate1 = gates.innerGate
	gates.gate2 = gates.outerGate
}

func (gates *Controller) handleStuckRequest() {
	println("Stuck request received")
	gates.Reset()
	if gates.innerGate.isOpen() {
		println("Stuck: innergate open, opening inner")
		gates.stuckGate = gates.innerGate
		gates.innerGate.workingOutput.High()
		gates.innerGate.openRequestOutput.High()
	} else if gates.outerGate.isOpen() {
		println("Stuck: outergate open, opening outer")
		gates.stuckGate = gates.outerGate
		gates.outerGate.workingOutput.High()
		gates.outerGate.openRequestOutput.High()
	} else {
		println("Stuck: Opening outer")
		gates.stuckGate = gates.outerGate
		gates.outerGate.workingOutput.High()
		gates.outerGate.openRequestOutput.High()
	}
}

func (gates *Controller) handleStuckGates() {
	if !gates.stuckStarted && gates.cycleTicks > gateOpenTimeout {
		println("Stuck: Timeout opening, resetting")
		gates.Reset()
	} else if !gates.stuckStarted {
		if gates.stuckGate.isOpen() {
			println("Stuck: Gate opening, waiting for close")
			gates.stuckStarted = true
			gates.stuckGate.openRequestOutput.Low()
		}
	} else if gates.stuckStarted && gates.stuckGate.isClosed() {
		println("Stuck complete")
		gates.Reset()
	} else if gates.stuckStarted && gates.cycleTicks > gateCloseTimeout {
		println("Stuck: Timeout closing, resetting")
		gates.Reset()
	}
	gates.cycleTicks++
}

func (gates *Controller) Init() {
	gates.stuckRequestInput.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	gates.inboundCycling.Configure(machine.PinConfig{Mode: machine.PinOutput})
	gates.outboundCycling.Configure(machine.PinConfig{Mode: machine.PinOutput})
	gates.inboundCycling.Low()
	gates.outboundCycling.Low()
	gates.innerGate.Init()
	gates.outerGate.Init()
}

func (gates *Controller) Reset() {
	gates.stuckStarted = false
	gates.stuckGate = nil
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
	gates.innerGate.closedOutput.Set(gates.innerGate.isClosed())
	gates.outerGate.closedOutput.Set(gates.outerGate.isClosed())
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
