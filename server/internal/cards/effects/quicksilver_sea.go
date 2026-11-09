package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quicksilver Sea — Plane — Mirrodin (Planechase Anthology, #4870):
//
//	"When you planeswalk to Quicksilver Sea and at the beginning of your
//	 upkeep, scry 4.
//	 Whenever chaos ensues, reveal the top card of your library. You may
//	 play it without paying its mana cost."
//
// A plane (ADR 0216 §6). One scry ability with two conditions. The chaos
// ability is the free-play-from-the-top offer (game.FreePlayFromTop) with
// LeaveOnTop, since "if you don't" sends the card nowhere: a land is played
// only with a land drop left, a spell is cast for {0} with X zero and any
// optional additional costs available, timing restrictions ignored, and a
// card not played stays on top of the library (the rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53517dbf-7b32-423a-af76-afd39b876a1a",
		Name:         "Quicksilver Sea",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Quicksilver Sea — scry 4", Do(Scry{N: 4})),
			WheneverChaosEnsues("Quicksilver Sea — reveal the top card of your library; you may play it without paying its mana cost",
				func(g *game.Game, item *game.StackItem) error {
					return g.OfferFreePlayFromTopOrHandForEffect(game.FreePlayFromTop{
						Player:       item.Controller,
						Source:       item.SourceCardID,
						SourceName:   "Quicksilver Sea",
						MaxManaValue: 1 << 20,
						Land:         true,
						Reveal:       true,
						LeaveOnTop:   true,
					})
				}),
		},
	})
}
