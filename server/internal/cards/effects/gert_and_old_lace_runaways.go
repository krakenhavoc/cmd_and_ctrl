package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gert and Old Lace, Runaways — Legendary Creature — Human Dinosaur Hero
// {2}{R/G}, 4/3 (Marvel Super Heroes Commander, #4455):
//
//	"Trample
//	 When Gert and Old Lace enter, you may discard a card. If you do,
//	 search your library for a basic land card, reveal it, put it into
//	 your hand, then shuffle."
//
// Pia, Aether Ascetic's shape (an up-to-one discard prompt and a keyed
// run-then that searches only if a card was discarded), searching for a
// basic land instead of an enchantment.
//
// No simplification.
func init() {
	const label = "Gert and Old Lace — you may discard a card; if you do, search for a basic land"
	Register(Spec{
		OracleID:        "aaded61e-32c2-4e02-8557-3f8cd927a64e",
		Name:            "Gert and Old Lace, Runaways",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player:   ctx.Controller(),
					Source:   ctx.Source(),
					N:        1,
					UpTo:     true,
					Question: label,
				}, gertSearch, nil)
			}),
		},
	})
}

var gertSearch = game.RegisterDiscardThen("choose-cards/gert-and-old-lace-search", func(g *game.Game, r game.DiscardsDone) error {
	if r.Discarded.Count() == 0 {
		return nil
	}
	return SearchLibrary{
		Player:    r.Player,
		Predicate: IsBasicLand,
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Gert and Old Lace — a basic land card, to your hand",
		Source:    r.Source,
	}.Apply(keyedContext(g, r.Player, r.Source))
})
