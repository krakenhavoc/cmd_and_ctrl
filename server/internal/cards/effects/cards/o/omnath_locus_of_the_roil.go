package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ffd34151-457a-4a93-82db-24b18319a06b",
		Name:         "Omnath, Locus of the Roil",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.B06SelfETB,
				Targets:   effects.TargetAny(),
				Key:       "Omnath, Locus of the Roil — damage equal to the Elementals you control to any target",
				Effect:    effects.B34DamageChosenTargetPerElemental,
			},
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B33LandYouControlEntered(ev, source, g)
				},
				Targets: effects.TargetCreature("target Elemental you control", effects.YouControl(), effects.Subtype("Elemental")),
				Key:     "Omnath, Locus of the Roil — a +1/+1 counter on target Elemental; draw at eight lands",
				Effect:  effects.B34CounterOnChosenElementalThenDrawAtEightLands,
			},
		},
	})
}
