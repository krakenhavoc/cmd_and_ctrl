package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slickshot Lockpicker — Creature — Human Rogue {2}{U}, 2/3 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature enters, target instant or sorcery card in your
//	 graveyard gains flashback until end of turn. The flashback cost is
//	 equal to its mana cost.
//	 Plot {2}{U} (You may pay {2}{U} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// Snapcaster Mage's enters trigger (GrantFlashbackToCard, ADR 0066): the
// grant names that card object, ends at end of turn, and prices the
// flashback at the card's own mana cost; a card cast this way is exiled
// as it leaves the stack (CR 702.34a). Plot is the ordinary special
// action (CR 702.170).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "e6e4ad3c-079d-429b-9f29-b9e42a1a236f",
		Name:           "Slickshot Lockpicker",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{2}{U}")},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Slickshot Lockpicker — target instant or sorcery card in your graveyard gains flashback", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return GrantFlashbackToCard{
					Target: ts[0].ID,
					Label:  "Flashback — its mana cost (Slickshot Lockpicker)",
				}.Apply(ctx)
			}),
			TargetCardInGraveyard("target instant or sorcery card in your graveyard",
				YouOwn(), Or(Instant(), Sorcery())),
		)},
	})
}
