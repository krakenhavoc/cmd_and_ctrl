package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// soc_b16_helpers.go — shared by the sixteenth Secrets of Strixhaven
// Commander batch (#4834).

// socB16FirstXSpellYouCastThisTurn is a cost predicate for "the first spell
// you cast with {X} in its mana cost each turn" (Zimone, Infinite Analyst):
// the spell being priced has an {X} in its own printed mana cost (an {X}
// in an additional or alternative cost does not count, as
// socB9HasXInManaCost reads it), and the caster has cast no such spell
// yet this turn. The spell being priced is not in the tally yet (see
// NoArtifactSpellCastYetThisTurn), so a tally of none means this one is the
// first.
func socB16FirstXSpellYouCastThisTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil {
			return false
		}
		cost, err := game.ParseCost(q.Card.ManaCost)
		return err == nil && cost.XSlots > 0 && q.Game.CastTallyFor(q.Controller).XSpells == 0
	}
}
