package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dai Li Agents — Creature — Human Soldier {3}{B}{G}, 3/4 (Avatar: The Last
// Airbender, #4459):
//
//	"When this creature enters, earthbend 1, then earthbend 1.
//	 Whenever this creature attacks, each opponent loses X life and you
//	 gain X life, where X is the number of creatures you control with
//	 +1/+1 counters on them."
//
// Two instances of earthbend, so two instances of the word "target": two
// clauses of "target land you control" (EarthbendTargets), and NOT
// Distinct, so the same land may fill both (CR 115.3, the ruling). Each
// clause is judged on its own as the trigger resolves (CR 608.2b): an
// illegal first land skips the first earthbend and the second still
// happens. The second earthbend waits for the first one's counters to
// land (Earthbend.Then), which is the "then".
//
// X is counted as the attack trigger resolves (the ruling), over the
// creatures you control then with a +1/+1 counter on them, an earthbent
// land creature included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c05e716-b6eb-4f95-866b-7e43a898ebe2",
		Name:         "Dai Li Agents",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Dai Li Agents — earthbend 1, then earthbend 1", daiLiAgentsEarthbend),
				Clauses(EarthbendTargets(), EarthbendTargets()),
			),
			WheneverThisAttacks("Dai Li Agents — each opponent loses X life and you gain X life (X = creatures you control with +1/+1 counters)",
				func(g *game.Game, item *game.StackItem) error {
					return b40DrainEachOpponent(b34CreaturesYouControlMatching(g, item.Controller, b34HasPlusCounter))(g, item)
				}),
		},
	})
}

// daiLiAgentsEarthbend is the enters trigger: the first earthbend on clause
// 0's land, then the second on clause 1's.
func daiLiAgentsEarthbend(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	first, ok := ctx.ClauseTarget(0)
	if !ok || first.Kind != game.TargetCard {
		return daiLiAgentsSecondEarthbend(ctx)
	}
	return Earthbend{Target: first.ID, N: 1, Then: ctx.ItemTail(daiLiAgentsThen)}.Apply(ctx)
}

// daiLiAgentsThen is "then earthbend 1", once the first earthbend's
// counter has landed. An on-disk key: never renamed, never reused.
var daiLiAgentsThen = CardTail("dai-li-agents/second-earthbend", func(ctx *Context, _ game.TailOutcome) error {
	return daiLiAgentsSecondEarthbend(ctx)
})

// daiLiAgentsSecondEarthbend earthbends clause 1's land, if it is still a
// legal target.
func daiLiAgentsSecondEarthbend(ctx *Context) error {
	second, ok := ctx.ClauseTarget(1)
	if !ok || second.Kind != game.TargetCard || second.ID == uuid.Nil {
		return nil
	}
	return Earthbend{Target: second.ID, N: 1}.Apply(ctx)
}
