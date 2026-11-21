package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const lazotepSliverAfflictGrant = "lazotep-sliver/afflict-2"

// Lazotep Sliver — Creature — Zombie Sliver {3}{B}, 4/4 (Commander Masters):
//
//	"Sliver creatures you control have afflict 2. (Whenever a creature with
//	 afflict 2 becomes blocked, defending player loses 2 life.)
//	 Whenever a nontoken Sliver you control dies, amass Slivers 2. (Put two
//	 +1/+1 counters on an Army you control. It's also a Sliver. If you don't
//	 control an Army, create a 0/0 black Sliver Army creature token first.)"
//
// Lost Monarch of Ifnir's afflict grant over your Slivers, the Lazotep Sliver
// included (no Others): a creature with the ability triggers it when it
// becomes blocked, and Lazotep prints no afflict of its own beyond what it
// gives itself. The death trigger reads the Sliver as it last existed
// (CR 603.10a, leftAsSubtype), so a changeling counts and a token does not;
// the dying Sliver may be Lazotep itself. The Army is found or made by
// Amass, which makes it a Sliver too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2017d355-403a-4408-9779-94d1125e373f",
		Name:         "Lazotep Sliver",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 2},
		Grants: []effects.AbilityGrant{{
			Key:       lazotepSliverAfflictGrant,
			Triggered: []game.TriggeredAbility{effects.Afflict(2, "Lazotep Sliver")},
			Text:      "Afflict 2",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, lazotepSliverAfflictGrant),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, lazotepSliverNontokenSliverDied,
				"Lazotep Sliver — amass Slivers 2", effects.Do(effects.Amass{Subtype: "Sliver", N: 2})),
		},
	})
}

// lazotepSliverNontokenSliverDied is "a nontoken Sliver you control dies": a
// creature card that went to a graveyard, last controlled by the source's
// controller and a Sliver as it last existed.
func lazotepSliverNontokenSliverDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := effects.DiedCreature(ev, g)
	return ok && !effects.IsToken(dead) && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
		effects.LeftAsSubtype(ev, dead, source.TextSubtype("Sliver"))
}
