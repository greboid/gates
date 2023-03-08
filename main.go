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
	enabledInput      machine.Pin
}

type Controller struct {
	stuckStarted      bool
	stuckGate         *Gate
	singleGate        *Gate
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
	gatesOpenOutput   machine.Pin
}

func main() {
	outerGate := &Gate{
		Name:              "Outer",
		openRequestInput:  machine.D2,
		closedInput:       machine.D4,
		openRequestOutput: machine.ADC0,
		closedOutput:      machine.D10,
		workingOutput:     machine.D8,
		enabledInput:      machine.ADC3,
	}
	innerGate := &Gate{
		Name:              "Inner",
		openRequestInput:  machine.D3,
		closedInput:       machine.D5,
		openRequestOutput: machine.ADC1,
		closedOutput:      machine.D11,
		workingOutput:     machine.D9,
		enabledInput:      machine.ADC4,
	}
	gates := &Controller{
		outerGate:         outerGate,
		innerGate:         innerGate,
		inboundCycling:    machine.D6,
		outboundCycling:   machine.D7,
		stuckRequestInput: machine.D12,
		gatesOpenOutput:   machine.ADC2,
	}
	gates.Init()
	for {
		gates.setStatusLEDs()
		if outerGate.isEnabled() && innerGate.isEnabled() {
			if gates.stuckGate != nil {
				gates.handleStuckGates()
			} else if !gates.stuckRequestInput.Get() {
				gates.handleStuckRequest()
			} else if gates.gate2Opened && !gates.gate2Closed && gates.outerGate.isOpen() && outerGate.checkOpenRequestInput() {
				gates.Reset()
				gates.startInboundCycle()
			} else if gates.gate2Opened && !gates.gate2Closed && gates.innerGate.isOpen() && innerGate.checkOpenRequestInput() {
				gates.Reset()
				gates.startOutboundCycle()
			} else if gates.gate1 != nil && gates.gate2 != nil {
				gates.cycleGates()
			} else if outerGate.checkOpenRequestInput() && innerGate.isClosed() {
				gates.startInboundCycle()
			} else if innerGate.checkOpenRequestInput() && outerGate.isClosed() {
				gates.startOutboundCycle()
			}
		} else if gates.singleGate != nil {
			gates.openGate(gates.singleGate)
		} else if !outerGate.isEnabled() && innerGate.isEnabled() {
			if (innerGate.checkOpenRequestInput() || outerGate.checkOpenRequestInput()) && innerGate.isClosed() {
				gates.startSingleGateCycle(gates.innerGate)
			}
		} else if !innerGate.isEnabled() && outerGate.isEnabled() {
			if (innerGate.checkOpenRequestInput() || outerGate.checkOpenRequestInput()) && outerGate.isClosed() {
				gates.startSingleGateCycle(gates.outerGate)
			}
		}
		time.Sleep(SleepInterval * time.Millisecond)
	}
}

func (gates *Controller) startSingleGateCycle(g *Gate) {
	println("Starting cycle: ", g.Name)
	gates.singleGate = g
	gates.cycleTicks = -1
}

func (gates *Controller) openGate(g *Gate) {
	if !gates.gate1Opened {
		if gates.cycleTicks == 0 {
			println("Opening ", g.Name)
			g.openRequestOutput.High()
			g.workingOutput.High()
		} else if g.isOpen() {
			gates.gate1Opened = true
			g.openRequestOutput.Low()
		} else if gates.cycleTicks >= gateOpenTimeout {
			println("Timing out opening ", g.Name)
			gates.Reset()
		}
	} else if !gates.gate1Closed {
		if g.isClosed() {
			println("Closed", g.Name)
			gates.gate1Closed = true
			gates.cycleTicks = -1
			g.workingOutput.Low()
		} else if gates.cycleTicks >= gateCloseTimeout {
			println("Timing out closing ", g.Name)
			gates.Reset()
		}
	} else if gates.gate1Closed {
		println("Cycle finished: ", g.Name)
		gates.Reset()
	}
	gates.cycleTicks++
}

func (gates *Controller) cycleGates() {
	if !gates.gate1Opened {
		if gates.cycleTicks == 0 {
			println("Opening ", gates.gate1.Name)
			gates.gate1.openRequestOutput.High()
			gates.gate1.workingOutput.High()
		} else if gates.gate1.isOpen() {
			println("Opened ", gates.gate1.Name)
			gates.gate1Opened = true
			gates.gate1.openRequestOutput.Low()
		} else if gates.cycleTicks >= gateOpenTimeout {
			println("Timing out opening ", gates.gate1.Name)
			gates.Reset()
		}
	} else if !gates.gate1Closed {
		if gates.gate1.isClosed() {
			println("Closed ", gates.gate1.Name)
			gates.gate1Closed = true
			gates.cycleTicks = -1
			gates.gate1.workingOutput.Low()
		} else if gates.cycleTicks >= gateCloseTimeout {
			println("Timing out closing ", gates.gate1.Name)
			gates.Reset()
		}
	} else if gates.gate1Closed && !gates.gate2Opened {
		if gates.cycleTicks == 0 {
			println("Opening ", gates.gate1.Name)
			gates.gate2.openRequestOutput.High()
			gates.gate2.workingOutput.High()
		} else if gates.gate2.isOpen() {
			println("Opened ", gates.gate1.Name)
			gates.gate2Opened = true
			gates.gate2.openRequestOutput.Low()
		} else if gates.cycleTicks >= gateOpenTimeout {
			println("Timing out opening ", gates.gate1.Name)
			gates.Reset()
		}
	} else if !gates.gate2Closed {
		if gates.gate2.isClosed() {
			println("Closed ", gates.gate1.Name)
			gates.gate2Closed = true
			gates.cycleTicks = -1
			gates.gate2.workingOutput.Low()
		} else if gates.cycleTicks >= gateCloseTimeout {
			println("Timing out closing ", gates.gate1.Name)
			gates.Reset()
		}
	} else if gates.gate2Closed {
		println("Cycle finished: ", gates.gate1.Name, "=>", gates.gate2.Name)
		gates.Reset()
	}
	gates.cycleTicks++
}

func (gates *Controller) startInboundCycle() {
	gates.Reset()
	gates.inbound = true
	gates.gate1 = gates.outerGate
	gates.gate2 = gates.innerGate
	println("Starting cycle: ", gates.gate1.Name, "=>", gates.gate2.Name)
}

func (gates *Controller) startOutboundCycle() {
	gates.Reset()
	gates.outbound = true
	gates.gate1 = gates.innerGate
	gates.gate2 = gates.outerGate
	println("Starting cycle: ", gates.gate1.Name, "=>", gates.gate2.Name)
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
	gates.singleGate = nil
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
	g.enabledInput.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
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
	gates.gatesOpenOutput.Set((gates.innerGate.isOpen() || gates.outerGate.isOpen()) || ((gates.inbound || gates.outbound) && (gates.gate1Opened && !gates.gate2Closed)))
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

func (g *Gate) isEnabled() bool {
	return !g.enabledInput.Get()
}
