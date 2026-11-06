package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_up_to_count.go — "choose up to X —" where X is a count of the
// board, read as the triggered ability is put on the stack (CR 603.3c):
// Bumi, King of Three Trials' "the number of Lesson cards in your
// graveyard". The engine half is game.ModeSpec.UpToCount
// (game/mode_up_to_count.go); ChooseUpToX is the sibling that reads the
// X its spell was cast for instead.

// ChooseUpToCount — "choose up to X —, where X is <count>". The engine
// counts as the trigger goes on the stack and asks for up to that many
// bullets, each at most once, so it never asks for more than the card
// prints (CR 700.2d). At zero no bullet can be chosen and the trigger is
// removed (CR 700.2b). For a triggered ability only: effects.Register
// refuses it anywhere else.
func ChooseUpToCount(count game.ModeUpToCount, options ...game.ModeOption) *game.ModeSpec {
	return &game.ModeSpec{Prompt: "Choose up to X", Options: options, Min: 0, Max: len(options), UpToCount: count}
}

// checkUpToCount validates a counted "choose up to X" at boot. `t` is
// the trigger that owns the spec, or nil for a spell's or an activated
// ability's modes, which are refused: a spell announces its modes as it
// is cast and no printed spell or activated ability counts them this
// way. On a trigger it must start at zero, cap at its bullets and carry
// no other count — no repeat, raise, random pick, cast X, pawprints or
// "hasn't been chosen" memory, none of which the count reads.
func checkUpToCount(name, owner string, ms *game.ModeSpec, t *game.TriggeredAbility) {
	if ms == nil || ms.UpToCount.IsZero() {
		return
	}
	if t == nil {
		panic(fmt.Sprintf("effects.Register: %q declares a counted \"choose up to X\" on %s — only a triggered ability counts its modes as it is put on the stack (CR 603.3c)", name, owner))
	}
	if ms.Min != 0 || ms.Max != len(ms.Options) || ms.Repeatable || ms.RaisedMax != 0 || !ms.RaiseMaxIf.IsZero() ||
		ms.AtRandom || !ms.NotRandomIf.IsZero() || ms.UpToCastX || ms.Pawprints != 0 || ms.NotChosen != game.ModeMemoryNone {
		panic(fmt.Sprintf("effects.Register: %q declares a counted \"choose up to X\" that is not a plain up-to count — build it with ChooseUpToCount", name))
	}
}
