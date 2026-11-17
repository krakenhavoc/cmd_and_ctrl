package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Power Plant Worker — Artifact Creature — Assembly-Worker {5}, 4/4 (The
// Brothers' War, #5482):
//
//	"{3}: This creature gets +2/+2 until end of turn. If you control
//	 creatures named Mine Worker and Tower Worker, put two +1/+1 counters
//	 on this creature instead. Activate only once each turn."
//
// The trio is counted as the ability resolves (CR 608.2c). A Worker that
// left in response gets nothing (CR 400.7).
//
// No simplification.
func init() {
	const label = "{3}: This creature gets +2/+2 until end of turn. If you control creatures named Mine Worker and Tower Worker, put two +1/+1 counters on this creature instead. Activate only once each turn."
	Register(Spec{
		OracleID:     "054c6250-ac96-42f6-931d-925761ab7764",
		Name:         "Power Plant Worker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      ManaCost("{3}"),
			Condition: OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				if s82BroB11AssemblyTrio(g, item.Controller, "Mine Worker", "Tower Worker") {
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
				}
				return thisGetsUntilEndOfTurn(2, 2, "Power Plant Worker — +2/+2 until end of turn")(g, item)
			},
		}},
	})
}
