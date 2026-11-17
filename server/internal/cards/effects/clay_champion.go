package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clay Champion — Artifact Creature — Construct {X}{4}, 2/2 (The
// Brothers' War):
//
//	"This creature enters with three +1/+1 counters on it for each {G}{G}
//	 spent to cast it.
//	 When this creature enters, choose up to two other target creatures
//	 you control. For each {W}{W} spent to cast this creature, put a +1/+1
//	 counter on each of them."
//
// The entry counters are a CR 614.1c self-replacement read off the mana
// that paid for the spell (SelfEntersWithCountersFromSpentMana): three per
// pair of green, so four green is six. The enters trigger counts the white
// pairs as it is built (WhenThisEntersForEachPairSpent) and puts that many
// counters on each target still legal as it resolves (CR 608.2b).
func init() {
	const label = "Clay Champion — for each {W}{W} spent, a +1/+1 counter on each of up to two other target creatures you control"
	Register(Spec{
		OracleID:       "80920ac0-2bbe-4439-a937-282f63e23141",
		Name:           "Clay Champion",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		Replacements: []game.ReplacementEffect{
			SelfEntersWithCountersFromSpentMana(game.CounterPlusOne,
				"Enters with three +1/+1 counters on it for each {G}{G} spent to cast it.",
				func(m game.ManaSpent) int { return 3 * pairsSpent(m, "G") }),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEntersForEachPairSpent("W", label, func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: item.Params.Amount}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
				Another(TargetCreature("up to two other target creatures you control", YouControl()).WithCount(0, 2))),
		},
	})
}
