package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lin Sivvi, Defiant Hero — Legendary Creature — Human Rebel {1}{W}{W}, 1/3
// (Nemesis, S82 OLD batch 680, #5618):
//
//	"{X}, {T}: Search your library for a Rebel permanent card with mana
//	 value X or less, put it onto the battlefield, then shuffle.
//	 {3}: Put target Rebel card from your graveyard on the bottom of your
//	 library."
//
// The first ability is the Rebel recruiter with the X announced on
// activation (CR 107.3k); the search may find nothing (CR 701.23b) and the
// library is shuffled either way. The second targets a Rebel card in your
// graveyard (re-checked on resolution) and tucks it under your library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c3847b6-9bfb-49b5-8f1a-bdb40e6900f1",
		Name:         "Lin Sivvi, Defiant Hero",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{X}, {T}: Search your library for a Rebel permanent card with mana value X or less, put it onto the battlefield, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.SearchForASubtypePermanentWithManaValueAtMost("Rebel", ctx.X(),
						"Lin Sivvi, Defiant Hero — a Rebel permanent card with mana value X or less, onto the battlefield")(g, item)
				},
			},
			{
				Label:   "{3}: Put target Rebel card from your graveyard on the bottom of your library.",
				Cost:    effects.ManaCost("{3}"),
				Targets: effects.TargetCardInGraveyard("target Rebel card from your graveyard", effects.YouOwn(), effects.Subtype("Rebel")),
				Effect:  effects.S82One13TuckTargetToLibraryBottom,
			},
		},
	})
}
