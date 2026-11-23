package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opponent_mode.go — #5771, ADR 0065's 2026-11-19 amendment: "An
// opponent chooses one —" (Fatal Lore, Library of Lat-Nam, Misfortune),
// and "that player" in its bullets. The engine half is
// game/opponent_mode.go.

// OpponentChoosesOne — "An opponent chooses one —" (CR 700.2e). The
// opponent chooses the bullet as the spell is cast, when the caster
// otherwise would (CR 601.2b), and the caster chooses any targets after
// (Fatal Lore's ruling). With more than one opponent the caster says which
// one chooses. The engine parks the cast and asks, so the card file
// declares the bullets and nothing else: the caster's payload names no
// mode, the client skips the mode picker, and the bot names none. That
// opponent is "that player" in the bullets: read it at resolution with
// ctx.ModeChooser(), and write a target clause "… that player controls"
// as ThatPlayerControls(…).
//
// A spell's Modes only: effects.Register refuses it on a trigger or an
// activated ability, and beside a cost, a raised count or a repeat.
func OpponentChoosesOne(options ...game.ModeOption) *game.ModeSpec {
	return &game.ModeSpec{Prompt: "An opponent chooses one", Options: options, Min: 1, Max: 1, ChosenByOpponent: true}
}

// ThatPlayerControls is "… that player controls" on a bullet of an
// OpponentChoosesOne spell, where "that player" is the opponent who chose
// the bullet (game.TargetSpec.ControlledByModeChooser): Fatal Lore's "up
// to two target creatures that player controls". Judged at announce, at
// the CR 608.2b re-check and at a retarget against the opponent the cast
// recorded. Only on such a bullet; effects.Register refuses it anywhere
// else.
func ThatPlayerControls(spec *game.TargetSpec) *game.TargetSpec {
	spec.ControlledByModeChooser = true
	return spec
}

// ModeChooser is the opponent who chose this spell's mode, "that player"
// of an OpponentChoosesOne spell's bullets (game.StackItem.ModeChooser),
// or uuid.Nil for every other item.
func (c *Context) ModeChooser() uuid.UUID {
	if c.Item == nil {
		return uuid.Nil
	}
	return c.Item.ModeChooser
}

// checkOpponentChoosesMode validates "An opponent chooses one" on a
// spell's Modes: a plain one-of, every bullet without a cost, and no
// count that rises or repeats. No printed card has more, and the parked
// mode_pick asks exactly that question.
func checkOpponentChoosesMode(name string, ms *game.ModeSpec) {
	if ms == nil || !ms.ChosenByOpponent {
		return
	}
	if ms.Min != 1 || ms.Max != 1 || ms.Repeatable || ms.RaisedMax != 0 || ms.RaisedMin != 0 || !ms.RaiseMaxIf.IsZero() ||
		ms.Escalate != nil || ms.Pawprints != 0 || ms.UpToCastX || !ms.UpToCount.IsZero() ||
		ms.NotChosen != game.ModeMemoryNone || ms.AtRandom {
		panic(fmt.Sprintf("effects.Register: %q declares \"an opponent chooses\" beside another mode count, a repeat or escalate — build it with OpponentChoosesOne", name))
	}
	for i, o := range ms.Options {
		if o.Cost != "" {
			panic(fmt.Sprintf("effects.Register: %q mode %d has a cost on a mode an opponent chooses — no printed card does", name, i))
		}
	}
}

// refuseOpponentChoosesMode refuses "an opponent chooses" on a trigger's
// or an activated ability's Modes: CR 700.2e reaches both, but no printed
// card has one, and only a spell's cast is parked for it.
func refuseOpponentChoosesMode(name, owner string, ms *game.ModeSpec) {
	if ms != nil && ms.ChosenByOpponent {
		panic(fmt.Sprintf("effects.Register: %q declares \"an opponent chooses\" on %s — only a spell's Modes can be (OpponentChoosesOne)", name, owner))
	}
}

// refuseThatPlayerControls refuses ThatPlayerControls on a clause that is
// not a bullet of an OpponentChoosesOne spell: with nobody having chosen
// a mode there is no "that player".
func refuseThatPlayerControls(name, owner string, spec *game.TargetSpec) {
	for i := 0; i < spec.ClauseCount(); i++ {
		if spec.Clause(i).ControlledByModeChooser {
			panic(fmt.Sprintf("effects.Register: %q says \"that player controls\" on %s clause %d, where no opponent chose a mode — only a bullet of OpponentChoosesOne can", name, owner, i))
		}
	}
}
