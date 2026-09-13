package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Triplicate Titan — Artifact Creature — Golem {9}, 9/9 (EDHREC rank
// 2208):
//
//	"Flying, vigilance, trample
//	 When this creature dies, create a 3/3 colorless Golem artifact
//	 creature token with flying, a 3/3 colorless Golem artifact
//	 creature token with vigilance, and a 3/3 colorless Golem artifact
//	 creature token with trample."
//
// Wurmcoil Engine's shape, one keyword wider: the printed keywords
// ride PrintedKeywords, and the dies trigger (cardDied — graveyard
// only, so an exile or a bounce makes no Golems) creates three
// tokens that each carry exactly one of the three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "56e05b17-aa1d-483d-857c-4cab4f12aa8a",
		Name:            "Triplicate Titan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Triplicate Titan — create three 3/3 Golems",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, kw := range []string{"flying", "vigilance", "trample"} {
							if err := (CreateToken{
								Controller: item.Controller,
								Template:   b20GolemToken(kw),
								N:          1,
							}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
			},
		}},
	})
}
