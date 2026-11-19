package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        oracle,
		Name:            "Ondu Knotmaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && dead.InstanceID != source.InstanceID &&
					leftUnderControlOf(ev, dead) == source.Controller && msc17WasModified(g, dead.InstanceID)
			}, "Ondu Knotmaster — put two +1/+1 counters on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Throw a Line",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("one or two target creatures").WithCount(1, 2).Dividing(Divide(2)),
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: ResolveSteps(Instruction(func(ctx *Context) error {
			return PutDividedCounters(ctx, game.CounterPlusOne)
		})),
	})
}
