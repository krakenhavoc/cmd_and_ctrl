package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ajani's Pridemate — Creature — Cat Soldier {1}{W}, 2/2 (EDHREC rank
// 2587):
//
//	"Whenever you gain life, put a +1/+1 counter on this creature."
//
// The lifegain deck's two-drop. b10YouGainedLife is the condition —
// every lifegain path emits a positive EventChangeLife, lifelink
// included — and the trigger fires once per gain event, not once per
// point, so a 3-life gain is one counter, as printed. A Pridemate
// that left the battlefield before the trigger resolves gets
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "95e94dea-5ac0-4d6f-adec-ca147aee861f",
		Name:         "Ajani's Pridemate",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Ajani's Pridemate — put a +1/+1 counter on it", func(g *game.Game, item *game.StackItem) error {
				if !b09SourceStillOnBattlefield(g, item) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
