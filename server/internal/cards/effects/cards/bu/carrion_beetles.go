package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carrion Beetles — Creature — Insect {B}, 1/1:
//
//	"{2}{B}, {T}: Exile up to three target cards from a single
//	 graveyard."
//
// A repeatable Decompose on a body (#1807, ADR 0106 §5). Rag Dealer
// prints the same ability, so both use the shared constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d64a51be-2dd1-485f-98fa-c93d4c296e68",
		Name:         "Carrion Beetles",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.ExileFromASingleGraveyardAbility("{2}{B}, {T}", effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()), 3),
		},
	})
}
