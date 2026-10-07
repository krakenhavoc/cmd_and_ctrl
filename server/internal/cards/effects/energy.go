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
