package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Evolutionary Escalation — Enchantment {1}{G} (Commander 2016, #5537):
//
//	"At the beginning of your upkeep, put three +1/+1 counters on target
//	 creature you control and three +1/+1 counters on target creature an
//	 opponent controls."
//
// Two target clauses in printed order, chosen as the trigger goes on the
// stack (CR 603.3d), each re-checked as it resolves (CR 608.2b): a target
// that became illegal in response is skipped and the other still gets
// its three counters, as the ruling says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f9837988-a097-4f00-a94f-4856d5d9894e",
		Name:         "Evolutionary Escalation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourUpkeep("Evolutionary Escalation — three +1/+1 counters on target creature you control and on target creature an opponent controls",
				evolutionaryEscalationCounters),
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature an opponent controls", OpponentControls()),
				)),
		},
	})
}

func evolutionaryEscalationCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 3}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
