package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hymn of the Wilds — Conspiracy (Conspiracy: Take the Crown, S82 OLD batch
// 269, #5540):
//
//	"(Start the game with this conspiracy face up in the command zone.)
//	 The first creature spell you cast each turn costs {1} less to cast.
//	 You can't cast instant or sorcery spells."
//
// A face-up conspiracy (CR 315.5, ADR 0227). The discount is Radagast of
// Rhosgobel's first-creature-spell predicate, and the restriction is a cast
// gate on the conspiracy's controller (Moderation's shape): an instant or
// sorcery face of the card being cast is refused, from any zone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfa62414-48c6-4e5c-8609-1ce507b41505",
		Name:         "Hymn of the Wilds",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "The first creature spell you cast each turn costs {1} less to cast.",
				effects.YourSpell(), effects.CreatureSpell(), effects.NoCreatureSpellCastYetThisTurn()),
		},
		CastRestrictions: []game.CastRestriction{{
			Label: "Hymn of the Wilds — you can't cast instant or sorcery spells.",
			Forbids: func(q game.CastQuery) bool {
				return q.Source.Controller == q.Controller && (q.Card.IsInstant() || q.Card.IsSorcery())
			},
		}},
	})
}
