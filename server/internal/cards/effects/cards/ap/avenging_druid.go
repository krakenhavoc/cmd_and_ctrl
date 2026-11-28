package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avenging Druid — Creature — Human Druid {2}{G}, 1/3 (Exodus, S82 OLD batch 742, #5626):
//
//	"Whenever this creature deals damage to an opponent, you may reveal cards
//	 from the top of your library until you reveal a land card. If you do,
//	 put that card onto the battlefield and put all other cards revealed this
//	 way into your graveyard."
//
// Damage of any kind (combat or not) to an opponent triggers it; one trigger
// and one "you may" per damage event. On a yes the reveal runs to the first
// land card, which enters untapped, and every other card revealed goes to the
// graveyard. A library with no land card is revealed whole and all of it goes
// to the graveyard.
//
// No simplification.
func init() {
	const question = "Avenging Druid — reveal cards from the top of your library until you reveal a land card?"
	effects.Register(effects.Spec{
		OracleID:     "1f3fb367-60b5-4b36-9604-3064afdac0d9",
		Name:         "Avenging Druid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverThisDealsDamageToAnOpponent("Avenging Druid — reveal until a land card, put it onto the battlefield, the rest into your graveyard",
				effects.S82OldB742RevealUntilLandRestToGraveyard), question),
		},
	})
}
