package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angel of Indemnity — Creature — Angel Warrior {5}{W}, 5/5:
//
//	"Flying, lifelink
//	 When this creature enters, return target permanent card with mana
//	 value 4 or less from your graveyard to the battlefield.
//	 Encore {6}{W}{W} (…)"
//
// The enters trigger is Karmic Guide's reanimation over a permanent card
// of mana value 4 or less, chosen as the trigger goes on the stack and
// re-checked at resolution (CR 608.2b); the card comes back under your
// control. Encore is the shared keyword (encore.go, CR 702.141a): each
// token copy has the enters trigger too, so a table of three opponents
// returns up to three more cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4f57a320-b6b5-49de-af0a-f2e2649758a7",
		Name:            "Angel of Indemnity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Angel of Indemnity — return target permanent card with mana value 4 or less from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					effects.ReanimateSingleTarget(effects.NewContext(g, item), item.Controller)
					return nil
				}),
				effects.TargetCardInGraveyard("target permanent card with mana value 4 or less from your graveyard", effects.YouOwn(), effects.Permanent(), effects.ManaValueLE(4))),
		},
		Activated: []effects.ActivatedAbility{effects.Encore("{6}{W}{W}")},
	})
}
