package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pipsqueak, Rebel Strongarm — Legendary Creature — Human Rebel Ally
// {2}{W}, 4/4 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Pipsqueak can't attack alone unless he has a +1/+1 counter on him."
//
// Pet Sehlat's restriction (CantAttackAlone, CR 506.5, 508.1c) gated by
// AsLongAs on the counter: the limit exists only while Pipsqueak has no
// +1/+1 counter, read live at every check of the declaration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08a5445b-e9e6-44c3-9bc4-aff1b91c18a7",
		Name:         "Pipsqueak, Rebel Strongarm",
		Completeness: CompletenessFull,
		AttackLimits: []game.AttackLimit{AsLongAs(avB28HasNoPlusOneCounter, CantAttackAlone())},
	})
}

// avB28HasNoPlusOneCounter is "unless he has a +1/+1 counter on him",
// turned round into the condition the limit holds under.
func avB28HasNoPlusOneCounter(_ *game.Game, source *game.Card) bool {
	return source != nil && source.Counters[game.CounterPlusOne] <= 0
}
