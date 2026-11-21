package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drifting Djinn — Creature — Djinn {4}{U}{U}, 5/5 (Urza's Saga):
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice this creature unless you
//	 pay {1}{U}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Flying is the printed keyword, the upkeep is Child of Gaea's
// UpkeepSacrificeUnlessYouPay and cycling is the shared Cycling row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fdea1945-0bb6-4a83-8cb1-5ba6a4dc39bd",
		Name:            "Drifting Djinn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.UpkeepSacrificeUnlessYouPay("Drifting Djinn", "{1}{U}"),
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
