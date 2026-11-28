package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meteor Shower — Sorcery {X}{X}{R} (Ice Age, S82 OLD, #5636):
//
//	"Meteor Shower deals X plus 1 damage divided as you choose among any number of targets."
//
// Rolling Thunder's any-number division with the amount X + 1 (DivideXPlus(1), a new additive
// field on the divide spec, carried through the protocol and the client's divide picker). X is
// paid twice in mana; each target takes at least 1 (CR 601.2d), and with zero targets nothing
// is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2110083-b30f-43f1-9352-7536f5e19d3c",
		Name:         "Meteor Shower",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetAny().WithCount(0, 0).Dividing(effects.DivideXPlus(1)),
		OnResolve:    effects.S82OldB841DividedXPlusOne,
	})
}
