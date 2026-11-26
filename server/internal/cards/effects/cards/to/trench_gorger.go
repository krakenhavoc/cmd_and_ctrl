package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trench Gorger — Creature — Leviathan {6}{U}{U}, 6/6 (Commander 2011, #5574):
//
//	"Trample
//	 When this creature enters, you may search your library for any number of
//	 land cards, exile them, then shuffle. If you do, this creature has base
//	 power and toughness each equal to the number of cards exiled this way."
//
// The "you may" is the controller's CR 603.5 choice; saying yes is "you do", so
// a search that finds no land (or exiles none) still makes the Gorger 0/0 and
// it dies. The base power and toughness are a layer 7b set for as long as this
// Gorger stays on the battlefield (a different object it later becomes is not
// affected, CR 400.7), under counters and pumps.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "df23a4a5-eddd-4b0e-9f7e-a3733ce49c0d",
		Name:            "Trench Gorger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisEnters("Trench Gorger — search for any number of land cards, exile them, then shuffle; this creature's base power and toughness become the number exiled",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsLandCard,
						Dest:      game.ZoneExile,
						Unbounded: true,
						Shuffle:   true,
						Reason:    "Trench Gorger — search for any number of land cards to exile",
						Source:    item.SourceCardID,
						ThenTail:  ctx.ItemTail(trenchGorgerSearched),
					}.Apply(ctx)
				}),
				"Trench Gorger — search your library for any number of land cards?"),
		},
	})
}

// trenchGorgerSearched is "If you do, this creature has base power and
// toughness each equal to the number of cards exiled this way". An on-disk key.
var trenchGorgerSearched = effects.CardTail("then/trench-gorger-searched", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) || effects.SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	n := len(o.IDs)
	return effects.ScopedEffectFor{
		Target:   ctx.Source(),
		Mods:     game.SetBasePTMods(n, n),
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), ctx.Source()),
		Label:    "Trench Gorger — base power and toughness equal to the number of cards exiled",
	}.Apply(ctx)
})
