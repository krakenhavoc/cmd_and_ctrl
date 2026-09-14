package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dazzling Angel — Creature — Angel {2}{W}, 2/3 (EDHREC rank 3071):
//
//	"Flying
//	 Whenever another creature you control enters, you gain 1 life."
//
// Soul Warden on a flier, for your own creatures only. Flying rides
// PrintedKeywords; the trigger is Impact Tremors' condition
// (b13AnotherCreatureYouControlEntered — tokens included, the Angel
// itself excluded) and the life is the source's, so a lifegain
// payoff sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7de464e3-fae3-44cc-8233-776fc727c00a",
		Name:            "Dazzling Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b13AnotherCreatureYouControlEntered(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Dazzling Angel — you gain 1 life",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
