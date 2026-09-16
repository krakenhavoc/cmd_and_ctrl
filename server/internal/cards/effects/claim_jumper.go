package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Claim Jumper — Creature — Rabbit Mercenary {2}{W}, 3/3 (EDHREC
// rank 1844):
//
//	"Vigilance
//	 When this creature enters, if an opponent controls more lands
//	 than you, you may search your library for a Plains card and put
//	 it onto the battlefield tapped. Then if an opponent controls
//	 more lands than you, repeat this process once. If you search
//	 your library this way, shuffle."
//
// White's catch-up ramp. The intervening if (CR 603.4) is checked in
// AppliesTo — behind on lands, or no trigger at all — and again as
// the trigger resolves. The search is optional and the searcher
// chooses (S22); "a Plains card" is any land with the Plains
// subtype, so a Tundra or a Plateau qualifies as printed. The
// repeat runs in the first search's continuation, after the fetched
// Plains is on the battlefield, so the second land count sees it —
// and a declined first search still offers the second while the
// condition holds, as "repeat this process" does. Each search that
// happens shuffles, which is what the last sentence permits.
//
// The fetched Plains enters tapped through the search's own tapped
// clause (#478: the search path cannot pause for the land's own
// entry replacements, so the flag is the reliable half).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "18f0cd0b-3e4f-4637-a62e-75dd1b2f3fce",
		Name:            "Claim Jumper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && b03OpponentControlsMoreLands(g, source.Controller)
			}, "Claim Jumper — search for a Plains card, onto the battlefield tapped", func(g *game.Game, item *game.StackItem) error {
				if !b03OpponentControlsMoreLands(g, item.Controller) {
					return nil
				}
				return b17ClaimJumperSearch(g, item, true)
			}),
		},
	})
}
