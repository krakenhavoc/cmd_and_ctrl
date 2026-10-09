package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Entrust the Spark — Sorcery {3}{G}{U}:
//
//	"You may sacrifice a planeswalker. If you do, search your library for
//	 a planeswalker card, put it onto the battlefield, then shuffle."
//
// Daretti's shape: a yes/no question, then the sacrifice prompt, and the
// search is the sacrifice's continuation so it only happens once a
// planeswalker has really left ("if you do"). With no planeswalker to
// sacrifice the question is never asked. The fetched walker enters
// through the ordinary entry pipeline, so its starting loyalty is
// stamped as it arrives and its enters-tapped or ETB effects apply.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "b796ecc8-a2e8-4ab9-8736-6825b650434d",
		Name:         "Entrust the Spark",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			has := false
			for _, p := range ctx.Game.BattlefieldCardsForEffect() {
				if p.Controller == item.Controller && p.IsPlaneswalker() {
					has = true
					break
				}
			}
			if !has {
				return nil
			}
			return MayChoice{
				Question: "Entrust the Spark — sacrifice a planeswalker to search for a planeswalker?",
				YesLabel: "Sacrifice a planeswalker",
				NoLabel:  "Don't",
				OnYes:    entrustTheSparkSacrificeThenSearch,
			}.Apply(ctx)
		},
	})
}

func entrustTheSparkSacrificeThenSearch(ctx *Context) error {
	item := ctx.Item
	return ctx.Game.PlayerSacrificesThenForEffect(
		item.SourceCardID, item.Controller,
		sacrificeSpec("a planeswalker", Planeswalker()),
		"Entrust the Spark — sacrifice a planeswalker",
		1,
		func(g *game.Game, sacrificed game.PromptedSacrifices) error {
			if sacrificed.Count() == 0 {
				return nil
			}
			return SearchLibrary{
				Player:    item.Controller,
				Predicate: func(c game.Card) bool { return c.IsPlaneswalker() },
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Entrust the Spark — a planeswalker card",
				Source:    item.SourceCardID,
			}.Apply(NewContext(g, item))
		})
}
