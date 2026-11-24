package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shai-Hulud — Creature — Wurm {6}{G}{G}{G}, 12/12 (Secret Lair Drop):
//
//	"Trample, indestructible
//	 When this creature enters, destroy all artifacts and enchantments. For
//	 each player who controlled a permanent destroyed this way, search your
//	 library for a basic land card or a Desert card. Put those cards onto the
//	 battlefield tapped, then shuffle."
//
// The count is the number of DIFFERENT players who controlled a permanent that
// was actually destroyed (an indestructible permanent, or one regenerated, is
// not), read off the sweep's own list of what was destroyed. The cards found
// are one search for up to that many cards, each a basic land or a Desert,
// put onto the battlefield together tapped; the library is shuffled either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "155ee70b-898c-472f-a5e7-71f14da94894",
		Name:            "Shai-Hulud",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "indestructible"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepDestroy}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shai-Hulud — destroy all artifacts and enchantments; search for a basic land or Desert for each player who lost one",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.DestroyAllMatching{
						Match:    effects.Or(effects.Artifact(), effects.Enchantment()),
						ThenTail: ctx.ItemTail(shaiHuludSearch),
					}.Apply(ctx)
				}),
		},
	})
}

// shaiHuludSearch is the clause after the sweep: one search for a land per
// player who controlled a destroyed permanent. An on-disk key: never renamed,
// never reused.
var shaiHuludSearch = effects.MassTail("mass/shai-hulud-search", func(ctx *effects.Context, swept []game.Card, _ int) error {
	players := map[any]bool{}
	for _, c := range swept {
		players[c.Controller] = true
	}
	n := len(players)
	if n == 0 {
		return nil
	}
	return effects.SearchLibrary{
		Player:        ctx.Controller(),
		Predicate:     func(c game.Card) bool { return effects.IsBasicLand(c) || c.HasSubtype("Desert") },
		Dest:          game.ZoneBattlefield,
		Limit:         n,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        "Shai-Hulud — search for a basic land card or a Desert card for each player who controlled a destroyed permanent",
	}.Apply(ctx)
})
