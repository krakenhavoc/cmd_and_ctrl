package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tsabo's Web — Artifact {2}:
//
//	"When this artifact enters, draw a card.
//	 Each land with an activated ability that isn't a mana ability doesn't
//	 untap during its controller's untap step."
//
// The restriction narrows CR 502.3's untap set for every land on the
// battlefield, whoever controls it, that has an activated ability other
// than a mana ability right now: printed or granted, read through the
// layered view (game.ActivatedAbilitiesForCard lists the non-mana
// activated abilities, CR 605.1a). A land that loses its abilities, or one
// that has only mana abilities, untaps as usual. Only the untap step is
// touched: an untap effect still untaps such a land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96a4e515-2b4e-4c2d-aa1a-963597484723",
		Name:         "Tsabo's Web",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		UntapStepRestrictions: []game.UntapStepRestriction{{
			Label: "Tsabo's Web — each land with an activated ability that isn't a mana ability doesn't untap",
			Restricts: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsLand() && len(game.ActivatedAbilitiesForCard(*target)) > 0
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tsabo's Web — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
