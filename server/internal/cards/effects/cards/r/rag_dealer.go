package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rag Dealer — Creature — Human Rogue {B}, 1/1:
//
//	"{2}{B}, {T}: Exile up to three target cards from a single
//	 graveyard."
//
// Carrion Beetles' ability on a Human Rogue (#1807, ADR 0106 §5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "493acba2-92b0-43a4-9f4b-d49a91ca77f5",
		Name:         "Rag Dealer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.ExileFromASingleGraveyardAbility("{2}{B}, {T}", effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()), 3),
		},
	})
}
