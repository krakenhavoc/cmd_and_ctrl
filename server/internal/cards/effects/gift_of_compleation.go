package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gift of Compleation — Enchantment {1}{B} (March of the Machine):
//
//	"When this enchantment enters, incubate 3.
//	 Whenever a Phyrexian you control dies, surveil 1."
//
// The enters trigger makes the Incubator (CR 701.53a, ADR 0185). The
// surveil trigger reads the dying creature as it last existed on the
// battlefield (CR 603.10a), so a token Phyrexian counts, a transformed
// Incubator counts as the Phyrexian Token it was, and a changeling
// counts; an untransformed Incubator is an artifact, not a Phyrexian.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c00a13cc-3c2b-473f-801c-3cfcca642ac0",
		Name:         "Gift of Compleation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gift of Compleation — incubate 3", Do(Incubate{N: 3})),
			On(game.EventLTB, aPhyrexianYouControlDied,
				"Gift of Compleation — surveil 1", Do(Surveil{N: 1})),
		},
	})
}

// aPhyrexianYouControlDied is "whenever a Phyrexian you control dies".
func aPhyrexianYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && leftUnderControlOf(ev, dead) == source.Controller && leftAsSubtype(ev, dead, source.TextSubtype("Phyrexian"))
}
