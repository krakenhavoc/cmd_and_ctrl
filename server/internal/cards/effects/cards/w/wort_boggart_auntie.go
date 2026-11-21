package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wort, Boggart Auntie — Legendary Creature — Goblin Shaman {2}{B}{R}, 3/3
// (Lorwyn, #5593):
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)
//	 At the beginning of your upkeep, you may return target Goblin card
//	 from your graveyard to your hand."
//
// Fear is the printed keyword (CR 702.36b). The upkeep trigger targets as
// it goes on the stack (CR 603.3d), so with no Goblin card in your
// graveyard it is removed; "a Goblin card" is any card of the type, a
// Kindred one or a changeling card included. The return is the "you may"
// (CR 603.5), and a card gone by resolution stays where it is (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6f477e2e-220e-48d9-9236-9564478e796a",
		Name:            "Wort, Boggart Auntie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.Targeting(
					effects.AtYourUpkeep("Wort, Boggart Auntie — return target Goblin card from your graveyard to your hand",
						effects.ReturnFirstLegalGraveyardTargetToHand),
					effects.TargetCardInGraveyard("target Goblin card from your graveyard", effects.YouOwn(), effects.HasSubtype("Goblin"))),
				"Wort, Boggart Auntie — return target Goblin card from your graveyard to your hand?"),
		},
	})
}
