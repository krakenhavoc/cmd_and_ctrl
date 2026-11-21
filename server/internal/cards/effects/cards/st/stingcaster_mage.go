package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stingcaster Mage — Creature — Human Wizard {1}{R}, 2/1:
//
//	"Haste
//	 When this creature enters, target instant or sorcery card in your
//	 graveyard gains flashback until end of turn. The flashback cost is
//	 equal to its mana cost."
//
// Snapcaster Mage's grant for red's two-drop; the engine's
// GrantFlashbackToCard ends at end of turn and prices the flashback at
// the card's own mana cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "056b651e-e0e2-4333-9235-d1ffe8fcca29",
		Name:            "Stingcaster Mage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WhenThisEnters("Stingcaster Mage — target instant or sorcery card in your graveyard gains flashback", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return effects.GrantFlashbackToCard{
					Target: ts[0].ID,
					Label:  "Flashback — its mana cost (Stingcaster Mage)",
				}.Apply(ctx)
			}),
			effects.TargetCardInGraveyard("target instant or sorcery card in your graveyard",
				effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
		)},
	})
}
