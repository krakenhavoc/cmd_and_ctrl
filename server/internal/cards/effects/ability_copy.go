package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_copy.go — the catalog side of CR 707.10 for ABILITIES
// (#1223), next to spell_copy.go's for spells.
//
// Three cards, one sentence with a different adjective on the target
// clause, exactly as the spell family is: Strionic Resonator copies a
// triggered ability, Lithoform Engine either kind, and Rings of
// Brighthearth copies the one you just activated without targeting at
// all.

// CopyAbility copies an activated or triggered ability that is
// currently on the stack, per CR 707.10.
//
// ItemID is a STACK ITEM id, not a card id. An ability's source
// permanent stays on the battlefield and can have several of its
// abilities on the stack at once, so the card is not an answer —
// a target clause built with AbilityOnStack yields the item id in
// an ordinary TargetRef, and a watcher on an activation reads it off
// game.Event.StackItemID.
//
// An ability that is no longer on the stack — countered, or already
// resolved, in response — is skipped silently, which is the same
// outcome CopySpell gives for a spell that left. So is a MANA
// ability: it never used the stack (CR 605.3b), so there is nothing
// to find, and Rings of Brighthearth's "if it isn't a mana ability"
// needs no clause of its own.
type CopyAbility struct {
	// ItemID is the ability to copy — a stack item id.
	ItemID uuid.UUID

	// Controller is who controls the copy. Usually ctx.Controller():
	// CR 707.10 gives the copy to the player who created it, not to
	// the controller of the copied ability.
	Controller uuid.UUID

	// Count is how many copies to create. Zero is treated as one.
	Count int

	// ChooseNewTargets carries the "You may choose new targets for
	// the copy" clause (CR 707.10c). All three cards in the family
	// print it; the engine skips the prompt anyway when the copied
	// ability has no target clause.
	ChooseNewTargets bool
}

func (c CopyAbility) Apply(ctx *Context) error {
	n := c.Count
	if n <= 0 {
		n = 1
	}
	controller := c.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	for i := 0; i < n; i++ {
		err := ctx.Game.CopyAbilityForEffect(c.ItemID, controller, c.ChooseNewTargets)
		if err == game.ErrCardNotFound || err == game.ErrInvalidParam {
			// The ability left the stack, was a mana ability that was
			// never on it, or the id named a spell. The copies
			// already made stand; there is nothing left to copy.
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// AbilityOnStack — "target activated or triggered ability", CR 115.6.
//
// The clause Strionic Resonator and Lithoform Engine print. Narrow it
// the way the cards do: AnAbilityYouControl() for "you control",
// TriggeredAbilityOnly() for Strionic's "triggered ability".
//
// The picked ref is a TargetRef{Kind: TargetCard} carrying the STACK
// ITEM's id — see game.TargetSpec.Abilities for why an ability rides
// the card kind rather than a fifth one — so the resolution reads it
// with ctx.Targets()[0].ID and hands it straight to CopyAbility.
func AbilityOnStack(label string, preds ...AbilityPredicate) *game.TargetSpec {
	pred := AllAbilities(preds...)
	return &game.TargetSpec{
		// The client picker has no ability mode of its own yet, so
		// the hint names the surface the item is drawn on: an ability
		// row sits in the stack overlay beside the spells.
		Mode:      "stack_spell",
		Label:     label,
		Abilities: true,
		AbilityOK: func(g *game.Game, chooser uuid.UUID, item *game.StackItem) bool {
			return pred(g, chooser, item)
		},
		Min: 1, Max: 1,
	}
}

// AbilityPredicate narrows an AbilityOnStack clause. The ability
// twin of CardPredicate, over the stack item rather than a card.
type AbilityPredicate = func(g *game.Game, chooser uuid.UUID, item *game.StackItem) bool

// AllAbilities composes predicates with AND. A clause with none
// admits every activated and triggered ability on the stack.
func AllAbilities(preds ...AbilityPredicate) AbilityPredicate {
	return func(g *game.Game, chooser uuid.UUID, item *game.StackItem) bool {
		for _, p := range preds {
			if !p(g, chooser, item) {
				return false
			}
		}
		return true
	}
}

// AnAbilityYouControl — "…ability you control". Both cards in the
// family print it, and it is the clause that keeps Strionic Resonator
// off an opponent's triggers.
func AnAbilityYouControl() AbilityPredicate {
	return func(_ *game.Game, chooser uuid.UUID, item *game.StackItem) bool {
		return item.Controller == chooser
	}
}

// TriggeredAbilityOnly — "target TRIGGERED ability", Strionic
// Resonator's whole restriction against Lithoform Engine's "activated
// or triggered".
func TriggeredAbilityOnly() AbilityPredicate {
	return func(_ *game.Game, _ uuid.UUID, item *game.StackItem) bool {
		return item.Kind == game.StackItemTriggered
	}
}

// AbilityFromSource — "…ability you control FROM AN ENCHANTMENT
// SOURCE" (Weaver of Harmony): the ability's source card passes
// `pred`.
//
// Read off the source as it is NOW rather than as it was when the
// ability was announced, which is the rule: CR 608.2 lets an ability
// resolve after its source has left, and an ability's source is an
// object whose characteristics can change while the ability sits on
// the stack. A source that is no longer anywhere fails the clause,
// which is the weaker-than-printed answer and the honest one — the
// engine cannot say what type an object it cannot find had.
func AbilityFromSource(pred CardPredicate) AbilityPredicate {
	return func(g *game.Game, chooser uuid.UUID, item *game.StackItem) bool {
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		return ok && pred(g, chooser, c)
	}
}
