package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Khan, Vengeful Heart — {2}{B}{B} Legendary Creature — Human Warrior, 5/3:
//
//	"When Khan enters, you may sacrifice another creature or artifact.
//	 If you do, draw two cards.
//	 You may cast this card from exile."
//
// The enters trigger asks the sacrifice as it resolves (maySacrificeOne,
// CR 608.2c), so it does not target, and the two cards follow only a
// sacrifice that really happened: Witherbloom Charm's keyed draw.
//
// "You may cast this card from exile" is the card's own text, so it opens
// exile for any face-up Khan its owner has there, however it got there:
// Swords to Plowshares, a blink that never came back, a warp. That is the
// declaration #659 refused for every other exile cast and the 2026-11-05
// amendment to ADR 0066 admits for this shape (CastsItselfFromExile). A Khan
// exiled face down has no text (CR 406.3a, 708.2) and is not opened. The
// cast is an ordinary one at the card's own timing and cost.
func init() {
	Register(Spec{
		OracleID:             "d2a86d0c-a4fe-434c-a45c-812cdec9169f",
		Name:                 "Khan, Vengeful Heart",
		Completeness:         CompletenessFull,
		CastableZones:        []game.ZoneKind{game.ZoneExile},
		CastsItselfFromExile: true,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Khan, Vengeful Heart — sacrifice another creature or artifact to draw two cards", func(g *game.Game, item *game.StackItem) error {
				return maySacrificeOne(g, item, "Khan, Vengeful Heart — sacrifice another creature or artifact to draw two cards?", Or(Creature(), Artifact()), true, witherbloomCharmDraw)
			}),
		},
	})
}
