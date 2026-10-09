package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "056b651e-e0e2-4333-9235-d1ffe8fcca29",
		Name:            "Stingcaster Mage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Stingcaster Mage — target instant or sorcery card in your graveyard gains flashback", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return GrantFlashbackToCard{
					Target: ts[0].ID,
					Label:  "Flashback — its mana cost (Stingcaster Mage)",
				}.Apply(ctx)
			}),
			TargetCardInGraveyard("target instant or sorcery card in your graveyard",
				YouOwn(), Or(Instant(), Sorcery())),
		)},
	})
}
