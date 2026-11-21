package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Screamer-Killer — Creature — Tyranid {4}{R}, 5/5 (Warhammer 40,000
// Commander, S82 GAP batch 194, #5948):
//
//	"Trample
//	 Bio-Plasmic Scream — Whenever you cast a creature spell with mana value 5
//	 or greater, this creature deals 5 damage to any target."
//
// Kalemne's cast watch on a creature spell of mana value 5 or more (an X
// counted as announced), with the target chosen as the trigger goes on the
// stack. The Screamer-Killer's own cast does not trigger it, since it is not on
// the battlefield yet.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb839b8c-a2b3-42fb-a7a2-8552b7656004",
		Name:            "Screamer-Killer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouCast(effects.And(effects.Creature(), effects.ManaValueGE(5)),
				"Screamer-Killer — 5 damage to any target", effects.DamageToFirstTarget(5)),
				effects.TargetAny()),
		},
	})
}
