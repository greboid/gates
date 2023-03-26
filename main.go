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
	working           bool
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
	debug             machine.Pin
}

func main() {
	outerGate := &Gate{
		Name:              "Outer",
		openRequestInput:  machine.D6,   //Terminal
		openRequestOutput: machine.D2,   //Terminal
		closedInput:       machine.D5,   //Terminal
		closedOutput:      machine.ADC1, //LED
		workingOutput:     machine.ADC2, //LED
		enabledInput:      machine.D10,  //Jumper
	}
	innerGate := &Gate{
		Name:              "Inner",
		openRequestInput:  machine.D7,   //Terminal
		openRequestOutput: machine.D3,   //Terminal
		closedInput:       machine.D8,   //Terminal
		closedOutput:      machine.ADC4, //LED
		workingOutput:     machine.ADC3, //LED
		enabledInput:      machine.D9,   //Jumper
	}
	gates := &Controller{
		debug:             machine.D12,
		outerGate:         outerGate,
		innerGate:         innerGate,
		inboundCycling:    machine.ADC0, //LED
		outboundCycling:   machine.ADC5, //LED
		stuckRequestInput: machine.D11,  //Terminal
		gatesOpenOutput:   machine.D4,   //Terminal
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
			g.working = true
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
			g.working = false
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
			gates.gate1.working = true
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
			gates.gate1.working = false
		} else if gates.cycleTicks >= gateCloseTimeout {
			println("Timing out closing ", gates.gate1.Name)
			gates.Reset()
		}
	} else if gates.gate1Closed && !gates.gate2Opened {
		if gates.cycleTicks == 0 {
			println("Opening ", gates.gate2.Name)
			gates.gate2.openRequestOutput.High()
			gates.gate2.working = true
		} else if gates.gate2.isOpen() {
			println("Opened ", gates.gate2.Name)
			gates.gate2Opened = true
			gates.gate2.working = false
		} else if gates.cycleTicks >= gateOpenTimeout {
			println("Timing out opening ", gates.gate2.Name)
			gates.Reset()
		}
	} else if !gates.gate2Closed {
		if gates.gate2.isClosed() {
			println("Closed ", gates.gate2.Name)
			gates.gate2Closed = true
			gates.cycleTicks = -1
			gates.gate2.working = false
		} else if gates.cycleTicks >= gateCloseTimeout {
			println("Timing out closing ", gates.gate2.Name)
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
		gates.innerGate.working = true
		gates.innerGate.openRequestOutput.High()
	} else if gates.outerGate.isOpen() {
		println("Stuck: outergate open, opening outer")
		gates.stuckGate = gates.outerGate
		gates.outerGate.working = true
		gates.outerGate.openRequestOutput.High()
	} else {
		println("Stuck: Opening outer")
		gates.stuckGate = gates.outerGate
		gates.outerGate.working = true
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
	gates.debug.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
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
	println("Gate ", g.Name, " enabled: ", g.isEnabled())
}

func (g *Gate) Reset() {
	g.openRequestOutput.Low()
	g.working = false
}

func (gates *Controller) isDebug() bool {
	return !gates.debug.Get()
}

func (gates *Controller) setStatusLEDs() {
	if gates.isDebug() {
		gates.innerGate.workingOutput.Set(gates.innerGate.working)
		gates.innerGate.closedOutput.Set(gates.innerGate.isClosed())
		gates.outerGate.closedOutput.Set(gates.outerGate.isClosed())
		gates.inboundCycling.Set(gates.inbound)
		gates.outboundCycling.Set(gates.outbound)
		gates.gatesOpenOutput.Set((gates.innerGate.isOpen() || gates.outerGate.isOpen()) || ((gates.inbound || gates.outbound) && (gates.gate1Opened && !gates.gate2Closed)))
	} else {
		gates.innerGate.workingOutput.Low()
		gates.innerGate.closedOutput.Low()
		gates.outerGate.closedOutput.Low()
		gates.inboundCycling.Low()
		gates.outboundCycling.Low()
		gates.gatesOpenOutput.Low()
	}
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
