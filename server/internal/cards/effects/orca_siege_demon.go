package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orca, Siege Demon — Legendary Creature — Demon {5}{B}{R}, 5/5
// (#1657):
//
//	"Trample
//	 Whenever another creature dies, put a +1/+1 counter on Orca.
//	 When Orca dies, it deals damage equal to its power divided as you
//	 choose among any number of targets."
//
// The death trigger's amount is DivideBy(DivideSourcePower): Orca's
// power as it last existed on the battlefield, counters included
// (CR 603.10, b13LastKnownPower), read as the trigger is put on the
// stack — the target walk, where the division is announced (the
// ruling: "You announce how the damage will be divided as part of
// putting Orca's last triggered ability on the stack"). A target that
// leaves in response takes nothing and its share is lost (CR 608.2b).
//
// "Another creature dies" is b15AnotherCreatureDied (the type the dying
// permanent LAST had, #1675) with Orca itself excluded; the counter is
// put on Orca only while it is still on the battlefield, so a creature
// that died alongside Orca adds nothing to the power the death trigger
// reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b8600894-3229-420e-9464-d92cf0a7d0fc",
		Name:            "Orca, Siege Demon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b15AnotherCreatureDied(ev, source, g)
				},
				Key:    "Orca, Siege Demon — put a +1/+1 counter on Orca",
				Effect: putCounterOnSelf,
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return cardDied(ev, source)
				},
				Targets: TargetAny().WithCount(0, 0).Dividing(DivideBy(DivideSourcePower)),
				Key:     "Orca, Siege Demon — damage equal to its power divided among any number of targets",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DealDividedDamage(NewContext(g, item))
				},
			},
		},
	})
}
