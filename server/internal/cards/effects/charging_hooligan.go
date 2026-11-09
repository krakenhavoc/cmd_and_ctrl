package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charging Hooligan — Creature — Human Peasant {3}{R}, 3/3 (Wilds of
// Eldraine, #4471):
//
//	"Whenever this creature attacks, it gets +1/+0 until end of turn for
//	 each attacking creature. If a Rat is attacking, this creature gains
//	 trample until end of turn."
//
// One trigger with two sentences, both read as it resolves: the count is
// every attacking creature, any player's, the Hooligan included (Ancestor
// Dragon's AttackingCreatures), and the trample rider needs an attacking
// Rat of any controller, a changeling counting. The Hooligan getting
// the bonus is the permanent that attacked: one that left does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c808892a-83d5-4b09-b0da-118b0d6fe58e",
		Name:         "Charging Hooligan",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked, "Charging Hooligan — +1/+0 for each attacking creature; trample if a Rat is attacking",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					if n := len(AttackingCreatures(g)); n > 0 {
						if err := thisGetsUntilEndOfTurn(n, 0, "Charging Hooligan — +1/+0 for each attacking creature")(g, item); err != nil {
							return err
						}
					}
					for _, c := range b26AttackingCreatures(g) {
						if c.HasSubtype("Rat") {
							return thisGainsKeywordUntilEndOfTurn("trample", "Charging Hooligan — trample until end of turn")(g, item)
						}
					}
					return nil
				}),
		},
	})
}
