package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Troll Negotiations — Sorcery {2}{G}{G} (The Hobbit, HOB):
//
//	"Put two +1/+1 counters on target creature you control. Then it
//	 fights target creature an opponent controls."
//
// Savage Stomp's resolution with a fixed count: the counters go on the
// first clause's creature, then the fight between the two still-legal
// picks reads its power after them (CR 608.2c). If the opponent's
// creature is gone the counters still land (CR 608.2b); if mine is gone
// there is no fight (#4454).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6e275711-eb76-4924-ba38-c38a36d9d82a",
		Name:         "Troll Negotiations",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: ResolveSteps(Instruction(trollNegotiationsCounters), Instruction(fightTargetClauses)),
	})
}

// trollNegotiationsCounters puts two +1/+1 counters on the first
// clause's creature, when it is still a legal pick.
func trollNegotiationsCounters(ctx *Context) error {
	mine, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	return AddCounter{Target: mine.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
}
