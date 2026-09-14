package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omnath, Locus of the Roil — Legendary Creature — Elemental
// {1}{G}{U}{R}, 3/3 (EDHREC rank 3568):
//
//	"When Omnath enters, it deals damage to any target equal to the
//	 number of Elementals you control.
//	 Landfall — Whenever a land you control enters, put a +1/+1
//	 counter on target Elemental you control. If you control eight or
//	 more lands, draw a card."
//
// The Temur Elementals commander. Both abilities are targeted
// triggers: the entry damage is "any target", counted as it
// resolves (Omnath counts himself, as printed); the landfall
// counter targets an Elemental the controller controls — Omnath is
// usually one — and the draw is read at resolution off the lands
// the controller controls then. A landfall target that left in
// response counters the whole ability, draw included (CR 608.2b);
// with no Elemental to target the trigger is removed (CR 603.3d),
// draw included, as the rules have it. Effective subtypes
// throughout, so a changeling is an Elemental.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ffd34151-457a-4a93-82db-24b18319a06b",
		Name:         "Omnath, Locus of the Roil",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Targets:   TargetAny(),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Omnath, Locus of the Roil — damage equal to the Elementals you control to any target",
						b34DamageChosenTargetPerElemental)
				},
			},
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b33LandYouControlEntered(ev, source, g)
				},
				Targets: TargetCreature("target Elemental you control", YouControl(), Subtype("Elemental")),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Omnath, Locus of the Roil — a +1/+1 counter on target Elemental; draw at eight lands",
						b34CounterOnChosenElementalThenDrawAtEightLands)
				},
			},
		},
	})
}
