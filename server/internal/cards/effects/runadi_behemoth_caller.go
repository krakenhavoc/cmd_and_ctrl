package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Runadi, Behemoth Caller — Legendary Creature — Cat Shaman {2}{G}, 1/3
// (Jumpstart 2022, #5948):
//
//	"Whenever you cast a creature spell with mana value 5 or greater, that
//	 creature enters with X additional +1/+1 counters on it, where X is its
//	 mana value minus 4.
//	 Creatures you control with three or more +1/+1 counters on them have
//	 haste.
//	 {T}: Add {G}."
//
// The cast trigger is Wildgrowth Archaic's, with X read off the spell's mana
// value as it is cast. The haste is a layer-6 grant over the controller's
// creatures with three or more +1/+1 counters, Runadi included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "95dc1f55-9dad-4a35-ac5e-abc53d442ffb",
		Name:         "Runadi, Behemoth Caller",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			gapB153RunadiCastTrigger("Runadi, Behemoth Caller — that creature enters with X additional +1/+1 counters on it, where X is its mana value minus 4"),
		},
		Static: []game.StaticAbility{b16GrantKeywords(gapB153ThreeOrMorePlusOneCounters, "haste")},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
