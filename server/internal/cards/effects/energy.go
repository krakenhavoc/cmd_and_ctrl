package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy.go — ADR 0129: getting energy, the shared vocabulary of the
// energy cards. Paying it is a cost component (PayEnergy / PayXEnergy in
// activated.go, game.AbilityCost.Energy).
//
// Energy is a counter on the player (CR 107.14, CR 122.1). "You get
// {E}{E}" puts two energy counters on the controller through the
// ADR 0056 window, so a replacement on getting energy and a trigger on
// EventPlayerCounterPlaced both see it.

// GetEnergy is "you get N {E}": N energy counters on Player, or on the
// controller of the resolving spell or ability when Player is unset. The
// controller gives them (CR 120.3's placer). Zero or less gets nothing.
type GetEnergy struct {
	Player uuid.UUID
	N      int
}

// Apply puts the counters on.
func (e GetEnergy) Apply(ctx *Context) error {
	if e.N <= 0 {
		return nil
	}
	player := e.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.AddPlayerCounterByForEffect(ctx.Controller(), player, game.CounterEnergy, e.N)
}

// EnergySymbols is n energy symbols, "{E}{E}", for a stack label.
func EnergySymbols(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("{E}", n)
}

// WhenThisEntersYouGetEnergy is the line most energy cards print: "When
// this <permanent> enters, you get N {E}". A trigger on the stack like
// any ETB (CR 603.6a). `name` is the card's name, for the stack label.
func WhenThisEntersYouGetEnergy(name string, n int) game.TriggeredAbility {
	return WhenThisEnters(name+" — you get "+EnergySymbols(n), Do(GetEnergy{N: n}))
}

// YouGotEnergy is the When for "whenever you get one or more {E}"
// (ADR 0129 §6): one or more energy counters landed on the source's
// controller. One placement is one EventPlayerCounterPlaced, so "one or
// more" triggers once per placement, as printed (CR 603.2c). The event
// carries the delta that landed, after any replacement (Izzet
// Generatorium's "that many plus one"), and a payment is a negative
// delta, so paying energy never triggers it.
func YouGotEnergy(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Label == game.CounterEnergy && ev.Amount > 0 && ev.Target == source.Controller
}

// WheneverYouGetEnergy is "Whenever you get one or more {E}, <effect>".
// Set Targets on the returned ability for a targeted body; the effect
// reads how much was gotten with EnergyGotten.
func WheneverYouGetEnergy(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventPlayerCounterPlaced, YouGotEnergy, label, effect)
}

// EnergyGotten is "that much" in a WheneverYouGetEnergy body: the energy
// the triggering placement put on the player.
func EnergyGotten(item *game.StackItem) int {
	if item == nil || item.Trigger == nil || item.Trigger.Event.Amount < 0 {
		return 0
	}
	return item.Trigger.Event.Amount
}

// PaidOrLostEnergyThisTurn is "Activate only if you've paid or lost N or
// more {E} this turn" (Izzet Generatorium), read off the turn tally
// (ADR 0129 §6).
func PaidOrLostEnergyThisTurn(n int) ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		return g.EnergyPaidOrLostThisTurn(controller) >= n
	}
}

// CostsLessForEachEnergyPaidOrLost is "This spell costs {N} less to cast
// for each {E} you've paid or lost this turn" (Blaster Hulk), read at
// CR 601.2f through the one pricer. Generic mana only.
func CostsLessForEachEnergyPaidOrLost(n int, label string) game.CostModifier {
	return CostsLessEach(func(q game.CostQuery) int {
		if q.Game == nil {
			return 0
		}
		return n * q.Game.EnergyPaidOrLostThisTurn(q.Controller)
	}, label)
}
