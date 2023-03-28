package main

import (
	"machine"
	"time"
)

const (
	OpenTimeout   = 50   //In number of ticks, so varies depending on sleep interval
	CloseTimeout  = 1200 //In number of ticks, so varies depending on sleep interval
	SleepInterval = 100  // In milliseconds
)

func main() {
	context := Context{
		State: Idle,
		Outer: &Door{
			requestOpenPin: machine.D2,
			isOpenPin:      machine.D4,
			isEnabledPin:   machine.D6,
		},
		Inner: &Door{
			requestOpenPin: machine.D3,
			isOpenPin:      machine.D5,
			isEnabledPin:   machine.D7,
		},
		inboundPin:    machine.D8,
		outboundPin:   machine.D9,
		stuckCyclePin: machine.D10,
		isClosedPin:   machine.D11,
	}
	context.Inner.requestOpenPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	context.Inner.isOpenPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	context.Inner.isEnabledPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	context.Inner.requestOpenPin.Low()

	context.Outer.requestOpenPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	context.Outer.isOpenPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	context.Outer.isEnabledPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	context.Outer.requestOpenPin.Low()

	context.inboundPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	context.outboundPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	context.stuckCyclePin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	context.isClosedPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	context.isClosedPin.Low()

	for {
		context.Update()
		switch context.State {
		case Idle:
			context.Outer.SetOpenRequest(false)
			context.Inner.SetOpenRequest(false)
			if context.InboundRequest {
				context.ChangeState(InboundCycleStarted)
			} else if context.OutboundRequest {
				context.ChangeState(OutboundCycleStarted)
			} else if context.StuckRequest {
				context.ChangeState(StuckCycleStarted)
			}
			break
		case InboundCycleStarted:
			if context.Outer.Enabled {
				context.Outer.SetOpenRequest(true)
				context.Ticks = 0
				context.ChangeState(InboundCycleFirstWaiting)
			} else if context.Inner.Enabled {
				context.ChangeState(InboundCycleFirstClosed)
			} else {
				context.ChangeState(Idle)
			}
			break
		case InboundCycleFirstWaiting:
			if context.Outer.Open {
				if !context.Inner.Enabled {
					context.SetOpen(true)
				}
				context.Outer.SetOpenRequest(false)
				context.Ticks = 0
				context.ChangeState(InboundCycleFirstOpened)
			} else if context.Ticks > OpenTimeout {
				context.ChangeState(Idle)
			} else {
				context.Ticks++
			}
			break
		case InboundCycleFirstOpened:
			if !context.Outer.Open {
				context.ChangeState(InboundCycleFirstClosed)
			} else if context.Ticks > CloseTimeout {
				context.ChangeState(Idle)
			} else {
				context.Ticks++
			}
			break
		case InboundCycleFirstClosed:
			if context.Inner.Enabled {
				context.Inner.SetOpenRequest(true)
				context.Ticks = 0
				context.ChangeState(InboundCycleSecondWaiting)
			} else {
				context.SetOpen(false)
				context.ChangeState(InboundCycleCompleted)
			}
			break
		case InboundCycleSecondWaiting:
			if context.Inner.Open {
				context.Inner.SetOpenRequest(false)
				context.SetOpen(true)
				context.Ticks = 0
				context.ChangeState(InboundCycleSecondOpened)
			} else if context.Ticks > OpenTimeout {
				context.ChangeState(Idle)
			} else {
				context.Ticks++
			}
			break
		case InboundCycleSecondOpened:
			if !context.Inner.Open {
				context.SetOpen(false)
				context.ChangeState(InboundCycleCompleted)
			} else if context.Ticks > CloseTimeout {
				context.ChangeState(Idle)
			} else if context.InboundRequest {
				context.ChangeState(OutboundCycleStarted)
			} else {
				context.Ticks++
			}
			break
		case InboundCycleCompleted:
			context.ChangeState(Idle)
			break
		case OutboundCycleStarted:
			if context.Inner.Enabled {
				context.Inner.SetOpenRequest(true)
				context.Ticks = 0
				context.ChangeState(OutboundCycleFirstWaiting)
			} else if context.Outer.Enabled {
				context.ChangeState(OutboundCycleFirstClosed)
			} else {
				context.ChangeState(Idle)
			}
			break
		case OutboundCycleFirstWaiting:
			if context.Inner.Open {
				if !context.Outer.Enabled {
					context.SetOpen(true)
				}
				context.Inner.SetOpenRequest(false)
				context.Ticks = 0
				context.ChangeState(OutboundCycleFirstOpened)
			} else if context.Ticks > OpenTimeout {
				context.ChangeState(Idle)
			} else {
				context.Ticks++
			}
			break
		case OutboundCycleFirstOpened:
			if !context.Inner.Open {
				context.ChangeState(OutboundCycleFirstClosed)
			} else if context.Ticks > CloseTimeout {
				context.ChangeState(Idle)
			} else {
				context.Ticks++
			}
		case OutboundCycleFirstClosed:
			if context.Outer.Enabled {
				context.Outer.SetOpenRequest(true)
				context.Ticks = 0
				context.ChangeState(OutboundCycleSecondWaiting)
			} else {
				context.SetOpen(false)
				context.ChangeState(OutboundCycleCompleted)
			}
			break
		case OutboundCycleSecondWaiting:
			if context.Outer.Open {
				context.Outer.SetOpenRequest(false)
				context.Ticks = 0
				context.SetOpen(true)
				context.ChangeState(OutboundCycleSecondOpened)
			} else if context.Ticks > OpenTimeout {
				context.ChangeState(Idle)
			} else {
				context.Ticks++
			}
			break
		case OutboundCycleSecondOpened:
			if !context.Outer.Open {
				context.SetOpen(false)
				context.ChangeState(OutboundCycleCompleted)
			} else if context.Ticks > CloseTimeout {
				context.ChangeState(Idle)
			} else if context.InboundRequest {
				context.ChangeState(InboundCycleStarted)
			} else {
				context.Ticks++
			}
			break
		case OutboundCycleCompleted:
			context.ChangeState(Idle)
			break
		case StuckCycleStarted:
			if !context.Outer.Enabled || !context.Inner.Enabled {
				context.ChangeState(StuckCycleComplete)
			} else if context.Outer.Open {
				context.Outer.SetOpenRequest(true)
				context.Ticks = 0
				context.ChangeState(StuckCycleOuterWaiting)
			} else if context.Inner.Open {
				context.Inner.SetOpenRequest(true)
				context.Ticks = 0
				context.ChangeState(StuckCycleInnerWaiting)
			} else {
				context.Outer.SetOpenRequest(true)
				context.Ticks = 0
				context.ChangeState(StuckCycleOuterWaiting)
			}
			break
		case StuckCycleOuterWaiting:
			if context.Outer.Open {
				context.Outer.SetOpenRequest(false)
				context.Ticks = 0
				context.ChangeState(StuckCycleOuterOpened)
			} else if context.Ticks > OpenTimeout {
				context.ChangeState(Idle)
			} else {
				context.Ticks++
			}
			break
		case StuckCycleOuterOpened:
			if !context.Outer.Open {
				context.ChangeState(StuckCycleComplete)
			}
			break
		case StuckCycleInnerWaiting:
			context.Inner.SetOpenRequest(false)
			if context.Inner.Open {
				context.Ticks = 0
				context.ChangeState(StuckCycleInnerOpened)
			} else if context.Ticks > OpenTimeout {
				context.ChangeState(Idle)
			} else {
				context.Ticks++
			}
			break
		case StuckCycleInnerOpened:
			if !context.Inner.Open {
				context.ChangeState(StuckCycleComplete)
			}
			break
		case StuckCycleComplete:
			context.ChangeState(Idle)
			break
		}
		time.Sleep(SleepInterval * time.Millisecond)
	}
}

func (c *Context) ChangeState(newState State) {
	println("Moving from ", c.State.String(), "to ", newState.String())
	c.State = newState
}

func (c *Context) SetOpen(state bool) {
	c.isClosedPin.Set(state)
}

func (c *Context) Update() {
	c.Inner.Update()
	c.Outer.Update()
	c.InboundRequest = !c.inboundPin.Get()
	c.OutboundRequest = !c.outboundPin.Get()
	c.StuckRequest = !c.stuckCyclePin.Get()
}

func (d *Door) Update() {
	d.Open = !d.isOpenPin.Get()
	d.Enabled = !d.isEnabledPin.Get()
}

func (d *Door) SetOpenRequest(state bool) {
	d.requestOpenPin.Set(state)
}
