package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ondu Knotmaster // Throw a Line — Creature — Kor Rogue {2}{W}{B}, 2/2 //
// Sorcery — Adventure {W}{B} (Modern Horizons 3):
//
//	"Lifelink
//	 Whenever another modified creature you control dies, put two +1/+1 counters
//	 on this creature."
//
//	Throw a Line: "Distribute two +1/+1 counters among one or two target
//	 creatures."
//
// "Modified" is read as the creature last existed on the battlefield (CR 603.10a).
func init() {
	const oracle = "dc69a130-967a-4d9f-992f-db7fad2fcd6d"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Ondu Knotmaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && dead.InstanceID != source.InstanceID &&
					effects.LeftUnderControlOf(ev, dead) == source.Controller && effects.Msc17WasModified(g, dead.InstanceID)
			}, "Ondu Knotmaster — put two +1/+1 counters on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Throw a Line",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("one or two target creatures").WithCount(1, 2).Dividing(effects.Divide(2)),
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			return effects.PutDividedCounters(ctx, game.CounterPlusOne)
		})),
	})
}
