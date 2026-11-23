package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// She-Hulk, Wallbreaker — {5}{R} Legendary Creature — Gamma Hero, 5/5
// (Marvel Super Heroes, #4455):
//
//	"Trample
//	 Other Heroes you control have trample.
//	 Whenever a Hero you control becomes blocked, put a +1/+1 counter on that
//	 Hero for each creature blocking it."
//
// The grant is Captain America's Hero filter with trample. A creature
// becomes blocked once however many block it (CR 506.4), and the number of
// blockers is read as the trigger resolves, so a blocker removed in response
// is not counted. A Hero that has left the battlefield by then gets nothing.
//
// No simplification.
func init() {
	heroes := effects.TribeFilter{Tribes: []string{"Hero"}, Others: true, YoursOnly: true}
	grant := effects.KeywordGrant(heroes.Matches, "trample")
	grant.Label = "Other Heroes you control have trample."

	effects.Register(effects.Spec{
		OracleID:        "c8d74a25-78a9-46ad-ab24-97f6221fd8dd",
		Name:            "She-Hulk, Wallbreaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{grant},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, sheHulkHeroBecameBlocked,
				"She-Hulk — put a +1/+1 counter on that Hero for each creature blocking it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					hero := ctx.Trigger().Event.CardID
					return effects.AddCounter{Target: hero, Kind: game.CounterPlusOne, N: len(effects.BlockersOf(g, hero))}.Apply(ctx)
				}),
		},
	})
}

// sheHulkHeroBecameBlocked is "a Hero you control becomes blocked".
func sheHulkHeroBecameBlocked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == ev.CardID {
			return c.Controller == source.Controller && c.HasSubtype(source.TextSubtype("Hero"))
		}
	}
	return false
}
