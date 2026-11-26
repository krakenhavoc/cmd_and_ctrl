package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b546_helpers.go — shared bodies for OLD batch 546 (#5593, Lorwyn).

// S82OldB546DrawIfYouControl is "If you control a <subtype>, draw a card.",
// checked as the instruction runs (CR 608.2c), after the spell's earlier
// sentences. The type is a word of the spell's text (CR 612.2) and a
// changeling you control has it (CR 702.73a).
func S82OldB546DrawIfYouControl(subtype string) Instruction {
	return func(ctx *Context) error {
		if !ControlsAtLeast(1, func(c game.Card) bool { return c.HasSubtypeWord(subtype) })(ctx.Game, ctx.Controller(), ctx.Source()) {
			return nil
		}
		return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}
}

// S82OldB546AnyTargetDealtDamageThisTurn is "any target that was dealt damage
// this turn" (Needle Drop): a creature, planeswalker or battle that was
// dealt damage this turn (DealtDamageThisTurn), or a player who was.
func S82OldB546AnyTargetDealtDamageThisTurn() *game.TargetSpec {
	spec := TargetAny()
	spec.Label = "any target that was dealt damage this turn"
	creatureOK := spec.CardOK
	pred := DealtDamageThisTurn()
	spec.CardOK = func(g *game.Game, caster uuid.UUID, c game.Card, z game.ZoneKind) bool {
		return creatureOK(g, caster, c, z) && pred(g, caster, c)
	}
	spec.PlayerOK = func(g *game.Game, _ uuid.UUID, p *game.Player) bool {
		return g.DamageTakenThisTurn(p.ID) > 0
	}
	return spec
}
