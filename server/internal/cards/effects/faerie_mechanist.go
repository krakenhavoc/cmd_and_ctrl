package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faerie Mechanist — Artifact Creature — Faerie Artificer {3}{U}, 2/2
// (Conflux, #5587):
//
//	"Flying
//	 When this creature enters, look at the top three cards of your
//	 library. You may reveal an artifact card from among them and put it
//	 into your hand. Put the rest on the bottom of your library in any
//	 order."
//
// Adaptive Omnitool's look-and-take at three cards, with the rest
// ordered by their owner (takeRevealRestBottomAny) rather than at
// random. Only the looker sees the three; the card taken is revealed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fcbe14c4-1b0d-486f-bb9c-0fed452668cb",
		Name:            "Faerie Mechanist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Faerie Mechanist — look at the top three cards, you may reveal an artifact card and put it into your hand, the rest on the bottom in any order",
				func(g *game.Game, item *game.StackItem) error {
					return TakeFromLibraryToHand{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 3),
						Match:    Artifact(),
						Max:      1,
						Optional: true,
						Label:    "Faerie Mechanist — you may reveal an artifact card from among them and put it into your hand",
						Key:      takeRevealRestBottomAny,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
