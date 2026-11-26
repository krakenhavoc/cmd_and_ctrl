package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b438_helpers.go — Avacyn Restored, batch 438 (S82, #5570).

// SecondSpellCastThisTurn matches the spell that was the second spell cast
// this turn by any player (Second Guess). The turn's cast list is append-only
// (Game.SpellsCastBeforeThisTurn), so a spell's place is frozen by its own
// cast; a copy was never cast and matches nothing (CR 707.10). Caller must
// hold g.mu, as every target predicate's caller does.
func SecondSpellCastThisTurn() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return g.SpellsCastBeforeThisTurn(c.InstanceID) == 1
	}
}

// SpellTargetsAPlayer matches a spell on the stack with at least one player
// among its targets (Outwit). A spell that targets only permanents or other
// spells does not qualify.
func SpellTargetsAPlayer() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		item := g.StackItemForEffect(c.InstanceID)
		if item == nil {
			return false
		}
		for _, t := range item.Targets {
			if t.Kind == game.TargetPlayer {
				return true
			}
		}
		return false
	}
}

// YouControlExactlyOneCreature is Demonic Rising's intervening-if test.
func YouControlExactlyOneCreature(ctx *Context) bool {
	return len(PermanentsControlledByMatching(ctx.Game, ctx.Controller(), Creature())) == 1
}
