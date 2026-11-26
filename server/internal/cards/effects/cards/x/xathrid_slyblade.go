package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Xathrid Slyblade — Creature — Human Assassin {2}{B}, 2/1 (Magic 2015,
// #5556):
//
//	"Hexproof
//	 {3}{B}: Until end of turn, this creature loses hexproof and gains first
//	 strike and deathtouch."
//
// Hexproof is the printed keyword. The ability is two layer-6 effects on the
// source until end of turn: the removal of hexproof (Shadowspear's shape, with
// its own timestamp) and the grant of first strike and deathtouch. A Slyblade
// that left and came back is a new object and is not changed (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "931c0cd3-06fb-42da-86f2-476d08a238f9",
		Name:            "Xathrid Slyblade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}: Until end of turn, this creature loses hexproof and gains first strike and deathtouch.",
			Cost:    effects.ManaCost("{3}{B}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.ThisLosesKeywordUntilEndOfTurn("hexproof", "Xathrid Slyblade — loses hexproof")(g, item); err != nil {
					return err
				}
				ctx := effects.NewContext(g, item)
				return effects.UntilEndOfTurn(ctx, item.SourceCardID, nil, "Xathrid Slyblade — first strike and deathtouch",
					game.AddKeywordsMod("first strike", "deathtouch"))
			},
		}},
	})
}
