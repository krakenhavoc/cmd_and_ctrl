package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b659_helpers.go — the Invasion Apprentices and Masters (#5616):
// each pairs two colour-activated {T} abilities, and the same ability turns
// up on more than one of them.

// S82OldB659DrainOne is "{B}, {T}: Target player loses 1 life." (Thunderscape
// Apprentice, Stormscape Apprentice). Life loss, not damage.
func S82OldB659DrainOne() ActivatedAbility {
	return ActivatedAbility{
		Label:   "{B}, {T}: Target player loses 1 life.",
		Cost:    Plus(ManaCost("{B}"), TapCost()),
		Targets: TargetPlayer("target player"),
		Effect:  DrainTargetedOpponent(1),
	}
}

// S82OldB659DrainTwoGainTwo is "{B}{B}, {T}: Target player loses 2 life and
// you gain 2 life." (Thunderscape Master, Stormscape Master).
func S82OldB659DrainTwoGainTwo() ActivatedAbility {
	return ActivatedAbility{
		Label:   "{B}{B}, {T}: Target player loses 2 life and you gain 2 life.",
		Cost:    Plus(ManaCost("{B}{B}"), TapCost()),
		Targets: TargetPlayer("target player"),
		Effect: func(g *game.Game, item *game.StackItem) error {
			return TargetPlayerLosesLifeYouGain(2)(item, NewContext(g, item))
		},
	}
}

// S82OldB659LifeBidders is "you and target spell's controller": the
// controller and the controller of the spell the first target names.
func S82OldB659LifeBidders(ctx *Context) []uuid.UUID {
	out := []uuid.UUID{ctx.Controller()}
	if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
		if spell := ctx.Game.StackItemForEffect(t.ID); spell != nil && spell.Controller != ctx.Controller() {
			out = append(out, spell.Controller)
		}
	}
	return out
}
