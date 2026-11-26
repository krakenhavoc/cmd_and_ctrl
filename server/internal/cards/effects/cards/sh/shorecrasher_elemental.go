package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shorecrasher Elemental — Creature — Elemental {U}{U}{U}, 3/3 (Dragons of
// Tarkir, S82 old batch 331):
//
//	"{U}: Exile this creature, then return it to the battlefield face down
//	 under its owner's control.
//	 {1}: This creature gets +1/-1 or -1/+1 until end of turn.
//	 Megamorph {4}{U} (You may cast this card face down as a 2/2 creature for
//	 {3}. Turn it face up any time for its megamorph cost and put a +1/+1
//	 counter on it.)"
//
// The first ability is Aethergeode Miner's blink of its own source with a
// face-down return (Death in Heaven's and Ashcloud Phoenix's entry door,
// CR 708.2a). The Elemental prints megamorph, so the face-down permanent can be
// turned up again for {4}{U} (CR 702.37e). The second is Brightling's choice,
// asked as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "5eb0d8dc-be4b-4a20-b480-f70514ffb299",
		Name:             "Shorecrasher Elemental",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{4}{U}")},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}: Exile this creature, then return it to the battlefield face down under its owner's control.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{U}"),
				Effect:  effects.ExileThisThenReturnItFaceDown,
			},
			{
				Label:   "{1}: This creature gets +1/-1 or -1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Shorecrasher Elemental — +1/-1 or -1/+1 until end of turn?",
						YesLabel: "+1/-1",
						NoLabel:  "-1/+1",
						Then:     shorecrasherElementalPump,
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

// shorecrasherElementalPump is the Elemental's choice: +1/-1 on a yes, -1/+1 on a
// no. An on-disk key: never renamed, never reused.
var shorecrasherElementalPump = effects.MayYesNo("confirm/shorecrasher-elemental-pump",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return effects.ThisGetsUntilEndOfTurn(1, -1, "Shorecrasher Elemental — +1/-1 until end of turn")(ctx.Game, ctx.Item)
	},
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return effects.ThisGetsUntilEndOfTurn(-1, 1, "Shorecrasher Elemental — -1/+1 until end of turn")(ctx.Game, ctx.Item)
	})
