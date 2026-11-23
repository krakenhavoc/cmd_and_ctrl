package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// clause, which is what the card prints; the land's own entry
// replacements are OR-ed on top of it by the CR 614 pipeline and may
// now pause for their own question (#478).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "18f0cd0b-3e4f-4637-a62e-75dd1b2f3fce",
		Name:            "Claim Jumper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && effects.B03OpponentControlsMoreLands(g, source.Controller)
			}, "Claim Jumper — search for a Plains card, onto the battlefield tapped", func(g *game.Game, item *game.StackItem) error {
				if !effects.B03OpponentControlsMoreLands(g, item.Controller) {
					return nil
				}
				return effects.B17ClaimJumperSearch(g, item, true)
			}),
		},
	})
}
