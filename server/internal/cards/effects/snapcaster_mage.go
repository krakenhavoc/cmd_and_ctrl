package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snapcaster Mage — Creature — Human Wizard {1}{U}, 2/1:
//
//	"Flash
//	 When this creature enters, target instant or sorcery card in your
//	 graveyard gains flashback until end of turn. The flashback cost is
//	 equal to its mana cost."
//
// The card ADR 0066 exists for. Flashback (CR 702.34) had been a
// printed keyword since S29 and nothing could give it to a card that
// does not print it, because every cast permission and every
// alternative cost was keyed by oracle ID.
//
// Three halves of the printed text, and each is one field on the
// granted permission:
//
//   - "target instant or sorcery card in your graveyard" — the
//     permission names that card OBJECT. CR 400.7: exile it and return
//     it and the new object has no flashback, which the engine
//     enforces with the card's object epoch rather than with a sweep.
//   - "the flashback cost is equal to its mana cost" — the permission
//     leaves its Cost empty, which reads as the printed mana cost of
//     whatever card it ends up covering. Nothing had to be computed at
//     resolution.
//   - "gains flashback", not "you may cast it" — so the granted offer
//     carries CR 702.34a's exile replacement, and a Snapcaster'd
//     Brainstorm that is countered is exiled rather than returned to
//     the graveyard to be flashed back again.
//
// Flash is a printed keyword and rides PrintedKeywords like every
// other, so the creature can be cast on an opponent's turn and the
// trigger can answer a spell — which is the whole card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2bb2eda7-3b38-4c56-870f-c3218a1056f5",
		Name:            "Snapcaster Mage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Snapcaster Mage — target card in your graveyard gains flashback", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return GrantFlashbackToCard{
					Target: ts[0].ID,
					Label:  "Flashback — its mana cost (Snapcaster Mage)",
				}.Apply(ctx)
			}),
			TargetCardInGraveyard("target instant or sorcery card in your graveyard",
				YouOwn(), Or(Instant(), Sorcery())),
		)},
	})
}
