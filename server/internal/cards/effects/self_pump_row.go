package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// self_pump_row.go — the whole activated-ability row "<cost>: This
// creature gets +P/+T until end of turn." with a mana cost and no tap:
// Shivan Dragon's firebreathing shape, which the Ice Age walls and
// spirits print at every size (#5636). The body is the shared
// thisGetsUntilEndOfTurn, so the boost lands on the ability's source and
// does nothing to a source that left and came back (CR 400.7). With no
// tap in the cost the row can be activated any number of times,
// summoning sickness or not (CR 302.6 is for the tap symbol).
//
// Append-only: add a builder, never change what one means.

// selfPumpRow builds the row for a non-negative power and toughness.
// `name` names the stack item ("Wall of Lava — +1/+1").
func selfPumpRow(name, cost string, power, toughness int) ActivatedAbility {
	size := fmt.Sprintf("%+d/%+d", power, toughness)
	return ActivatedAbility{
		Label:   cost + ": This creature gets " + size + " until end of turn.",
		Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: power, Toughness: toughness}},
		Cost:    ManaCost(cost),
		Effect:  thisGetsUntilEndOfTurn(power, toughness, name+" — "+size),
	}
}
