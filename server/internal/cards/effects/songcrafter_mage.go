package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Songcrafter Mage — Creature — Human Bard {G}{U}{R}, 3/2 (#3931):
//
//	"Flash
//	 When this creature enters, target instant or sorcery card in your
//	 graveyard gains harmonize until end of turn. Its harmonize cost is
//	 equal to its mana cost. (You may cast that card from your graveyard
//	 for its harmonize cost. You may tap a creature you control to reduce
//	 that cost by {X}, where X is its power. Then exile the spell.)"
//
// Snapcaster Mage's shape under the harmonize key (GrantHarmonizeToCard):
// the permission names the card OBJECT, so a card that leaves and returns
// has no harmonize (CR 400.7); its cost is the card's mana cost; and the
// key gives the granted offer harmonize's tap of up to one creature and
// its power reduction (CR 702.180a), so it prices exactly as a printed
// harmonize does. The spell is exiled as it leaves the stack. A card that
// already prints harmonize keeps its own offer beside the granted one
// (CR 113.2c, #3938).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "13f0863f-ae54-4bea-8f01-7f3891797577",
		Name:            "Songcrafter Mage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Songcrafter Mage — target card in your graveyard gains harmonize", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return GrantHarmonizeToCard{
					Target: ts[0].ID,
					Label:  "Harmonize — its mana cost (Songcrafter Mage)",
				}.Apply(ctx)
			}),
			TargetCardInGraveyard("target instant or sorcery card in your graveyard",
				YouOwn(), Or(Instant(), Sorcery())),
		)},
	})
}
