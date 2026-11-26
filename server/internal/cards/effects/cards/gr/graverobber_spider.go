package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Graverobber Spider — Creature — Spider {3}{G}, 2/4 (Born of the Gods, S82 OLD
// b378, #5559):
//
//	"Reach
//	 {3}{B}: This creature gets +X/+X until end of turn, where X is the number
//	 of creature cards in your graveyard. Activate only once each turn."
//
// Barbarian Bully's once-each-turn limit, which counts activations (CR
// 602.1b); X is read as the ability resolves.
//
// No simplification.
func init() {
	const label = "{3}{B}: This creature gets +X/+X until end of turn, where X is the number of creature cards in your graveyard. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "f0ad4a60-ce7f-4117-b8b0-266ffd0447c2",
		Name:            "Graverobber Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Cost:      effects.ManaCost("{3}{B}"),
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Condition: effects.OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.B11CreatureCardsInGraveyard(g, item.Controller)
				if x <= 0 || !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: x, Toughness: x, Label: "Graverobber Spider — +X/+X"}.Apply(ctx)
			},
		}},
	})
}
