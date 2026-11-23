package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Amulet — Artifact {2} (Foundations, #4465):
//
//	"{1}, {T}, Sacrifice a creature: Draw a card and put a soul counter on
//	 this artifact. Activate only as a sorcery.
//	 {4}, {T}, Sacrifice this artifact: Each opponent loses life equal to
//	 the number of soul counters on this artifact."
//
// Reaper's Scythe's soul counters. The first ability is a sorcery-speed
// sacrifice outlet: the creature is part of the cost (CR 118.3), the card is
// drawn and the counter goes on as the ability resolves. The second reads
// the counters as the Amulet last existed on the battlefield (CR 608.2h),
// since sacrificing it was part of the cost; with none it drains nothing.
//
// No simplification.
func init() {
	const soul = "soul"
	effects.Register(effects.Spec{
		OracleID:     "87424e20-00ab-4f6d-a905-1955897b358e",
		Name:         "Ravenous Amulet",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:        "{1}, {T}, Sacrifice a creature: Draw a card and put a soul counter on this artifact. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeACreature()),
				SorcerySpeed: true,
				Purpose:      game.Purpose{Draws: 1},
				Effect:       effects.Do(effects.DrawCards{N: 1}, effects.CounterOnThis{Kind: soul, N: 1}),
			},
			{
				Label:   "{4}, {T}, Sacrifice this artifact: Each opponent loses life equal to the number of soul counters on this artifact.",
				Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeThis()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok {
						return nil
					}
					return effects.EachOpponentLosesLifeNow(g, item, info.Counters[soul])
				},
			},
		},
	})
}
