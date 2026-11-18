package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brazen Upstart — Creature — Elf Shaman {R}{G}{W}, 4/2 (Streets of New
// Capenna):
//
//	"Vigilance
//	 When this creature dies, look at the top five cards of your library.
//	 You may reveal a creature card from among them and put it into your
//	 hand. Put the rest on the bottom of your library in a random order."
//
// Commune with Nature's dig on a dies trigger, the rest going to the bottom
// in a random order. The library is the controller's as the trigger
// resolves, so the Upstart leaving does not change who digs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e0a9c4be-74fe-4984-9028-60380db1575b",
		Name:            "Brazen Upstart",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Brazen Upstart — look at the top five cards; you may reveal a creature card and put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return TakeFromLibraryToHand{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 5),
						Match:    Creature(),
						Max:      1,
						Optional: true,
						Label:    "Brazen Upstart — reveal a creature card and put it into your hand",
						Key:      takeRevealRestBottomRandom,
					}.Apply(ctx)
				}),
		},
	})
}
