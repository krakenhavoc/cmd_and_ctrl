package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gapTargetsEnchantedCreature is "that target enchanted creature" for a
// cost modifier on an Aura: one of the spell's or ability's announced card
// targets is the permanent the modifier's own source is attached to
// (q.Source is the Aura). It reads q.Targets, so it is only meaningful on a
// modifier that sets ReadsTargets.
func GapTargetsEnchantedCreature() CostPredicate {
	return func(q game.CostQuery) bool {
		for _, t := range q.Targets {
			if t.Kind == game.TargetCard && q.Source.IsAttachedTo(t.ID) {
				return true
			}
		}
		return false
	}
}

// gapAuraSpell is "Aura spells": the spell being cast has the subtype Aura.
func GapAuraSpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.HasSubtype("Aura") }
}
