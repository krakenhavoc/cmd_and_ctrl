package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zameck Guildmage — Creature — Elf Wizard {G}{U}, 2/2 (Gatecrash, S82 OLD
// #5565):
//
//	"{G}{U}: This turn, each creature you control enters with an additional
//	 +1/+1 counter on it.
//	 {G}{U}, Remove a +1/+1 counter from a creature you control: Draw a card."
//
// The first ability is Combine Guildmage's: a replacement effect (CR 614.1c)
// held until cleanup, read live as each creature enters, one more per
// resolution. The second is Shapers of Nature's: the counter is taken off a
// creature chosen as the cost is paid (the Guildmage or another), and the card
// is drawn even if that creature then leaves. Neither needs {T}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98250fad-7fd0-431e-a87e-551185d26ec1",
		Name:         "Zameck Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{G}{U}: This turn, each creature you control enters with an additional +1/+1 counter on it.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{G}{U}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					g.CreaturesEnterWithAnAdditionalPlusOneCounterThisTurnForEffect(item.SourceCardID, item.Controller,
						"Zameck Guildmage — each creature you control enters with an additional +1/+1 counter this turn")
					return nil
				},
			},
			{
				Label:   "{G}{U}, Remove a +1/+1 counter from a creature you control: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{G}{U}"), effects.RemoveCountersFrom(game.CounterPlusOne, 1, "a creature you control", effects.Creature())),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
		},
	})
}
