package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clone — "You may have this creature enter as a copy of any
// creature on the battlefield."
//
// The baseline of the copy-effect class: no "except" clause, no
// controller restriction, no type restriction beyond "creature". If
// the player declines — or there is no creature to copy — Clone
// enters as its printed 0/0.
//
// Declared deviation, pre-existing and engine-wide: that 0/0 then
// survives. The CR 704.5f state-based action skips creatures with
// printed toughness 0 and no counters on purpose, because the same
// shape is the engine's placeholder convention for cards whose stats
// would not parse (see stateBasedActionsLocked). A player who
// declines a Clone keeps a 0/0 Shapeshifter that should have died.
// Not fixed here: the convention protects every fixture and demo
// card in the project, and "you declined a Clone" is the rarest
// possible way to notice it.
//
// "Any creature on the battlefield" is not targeting: a copy choice
// is made as the permanent enters, so hexproof, shroud and
// protection do not stop it (CR 706.2 — the ruling that makes Clone
// the answer to an opposing Blightsteel Colossus).
func init() {
	Register(Spec{
		OracleID:     "42226b87-0746-4ebf-9fd0-108d508462af",
		Name:         "Clone",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"If you decline the copy, the 0/0 Clone stays on the battlefield instead of dying."},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Clone",
				anyCreatureOnBattlefield,
				nil,
			),
		},
	})
}
