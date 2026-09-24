package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spell_copy.go — the catalog side of CR 707.10.
//
// One primitive, because every card in the family is the same
// sentence with a different adjective on the target clause:
// Reverberate copies any instant or sorcery, Increasing Vengeance
// copies one you control, Twincast is Reverberate in blue, and
// Doublecast copies the next one you cast rather than one on the
// stack. What differs between them is the TARGETING and the number
// of copies, never the copy itself.

// CopySpell copies a spell that is currently on the stack, per CR
// 707.10.
//
// Count > 1 produces that many copies, each created separately and
// each offered its own target choice — which is what Increasing
// Vengeance's "copy that spell twice instead" means (two
// independent copies, not one copy resolving twice).
//
// ChooseNewTargets carries the "You may choose new targets for the
// copy" clause. Leave it false only for a card that does not print
// it; the engine skips the prompt anyway when the copied spell has
// no target clause, so setting it on an untargeted copy costs
// nothing.
//
// A target that is no longer on the stack — countered, or already
// resolved, in response to this spell — is skipped silently, unless
// FromLastKnown says the effect does not target it. That
// is the CR 608.2b outcome for the whole spell in the single-target
// case, and the engine's re-check has already handled it by the
// time OnResolve runs; the guard here is for the multi-copy loop,
// where the first copy can change the board under the second.
type CopySpell struct {
	// StackID is the spell to copy — an instance ID on the stack.
	StackID uuid.UUID

	// Controller is who controls the copies. Usually
	// ctx.Controller(): CR 707.10 gives the copy to the player who
	// created it, NOT to the controller of the copied spell, and
	// that asymmetry is the whole point of Reverberate on an
	// opponent's Time Warp.
	Controller uuid.UUID

	// Count is how many copies to create. Zero is treated as one.
	Count int

	// ChooseNewTargets enables the CR 707.10c re-target prompt.
	ChooseNewTargets bool

	// Except is the card's "except …" clause (CR 707.10a) — Double
	// Major's "except it isn't legendary if the spell is legendary".
	// It edits the copiable values the copy is created with, the same
	// game.PrintedValues an entering permanent's except clause edits,
	// so "if the spell is legendary" needs no condition of its own:
	// RemoveSupertype is a no-op when the supertype is not there.
	//
	// It is applied to every copy independently, which is what a
	// multi-copy card with an except clause would mean. Nil for the
	// plain copies, which is every card in the family but one.
	Except func(v *game.PrintedValues)

	// FromLastKnown copies the spell from last-known information when
	// it has already left the stack (CR 608.2h, #1255) — the storm
	// spell countered in response to its own trigger still gets its
	// copies.
	//
	// Set it ONLY for an effect that names the spell without TARGETING
	// it: storm, Thousand-Year Storm, Doublecast's "copy that spell".
	// A copy effect that targets the spell (Reverberate, Twincast,
	// Dualcaster Mage) is governed by CR 608.2b instead — a target that
	// has left the stack is illegal and the copy must not happen — so
	// it leaves this false and keeps the strict lookup.
	FromLastKnown bool
}

func (c CopySpell) Apply(ctx *Context) error {
	n := c.Count
	if n <= 0 {
		n = 1
	}
	controller := c.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	for i := 0; i < n; i++ {
		copyFn := ctx.Game.CopySpellForEffect
		if c.FromLastKnown {
			copyFn = ctx.Game.CopyLastKnownSpellForEffect
		}
		err := copyFn(c.StackID, controller, c.ChooseNewTargets, c.Except)
		if err == game.ErrCardNotFound {
			// The copied spell left the stack between copies. The
			// copies already made stand; there is nothing left to
			// copy for the rest.
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// instantOrSorcerySpell is the target clause every card in this
// family prints. Kept here rather than in targets.go so a
// concurrent batch touching that file doesn't collide; move it if a
// third family wants it.
func instantOrSorcerySpell(label string, extra ...CardPredicate) *game.TargetSpec {
	preds := append([]CardPredicate{Or(Instant(), Sorcery())}, extra...)
	return TargetSpell(label, preds...)
}
