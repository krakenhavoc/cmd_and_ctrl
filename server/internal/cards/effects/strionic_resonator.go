package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strionic Resonator — Artifact {2} (#298, #1223):
//
//	"{2}, {T}: Copy target triggered ability you control. You may
//	 choose new targets for the copy. (A triggered ability uses the
//	 words 'when,' 'whenever,' or 'at.')"
//
// The card the ability-copy seam was filed on. Everything it needs is
// the two halves of #1223 and nothing of its own: a target clause
// that can name an ability ITEM rather than a card
// (AbilityOnStack / game.TargetSpec.Abilities), and a copy primitive
// that builds a second StackItem beside the first
// (game.CopyAbilityForEffect).
//
// "Triggered ability YOU CONTROL" is two predicates and they are both
// load-bearing: TriggeredAbilityOnly keeps the Resonator off an
// activated ability, which is the whole difference between this card
// and Lithoform Engine, and AnAbilityYouControl keeps it off an
// opponent's triggers. A mana ability is unreachable without either
// of them — it never uses the stack (CR 605.3b), so there is no item
// to offer.
//
// The copy remembers the original's triggering event (CR 707.10,
// StackItem.Trigger), which is what makes copying "whenever this
// creature deals combat damage to a player, draw that many cards"
// draw the same number rather than zero.
func init() {
	Register(Spec{
		OracleID:     "cf751552-156f-4f81-ac94-9814dce099f9",
		Name:         "Strionic Resonator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Copy target triggered ability you control",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: AbilityOnStack("target triggered ability you control", AnAbilityYouControl(), TriggeredAbilityOnly()),
			Effect:  copyTargetedAbility,
		}},
	})
}

// copyTargetedAbility is the body Strionic Resonator and both of
// Lithoform Engine's ability modes share: copy the ability the
// activation named, with the CR 707.10c re-target offer.
//
// A target that left the stack between announce and resolution —
// countered, or resolved in response — has already been caught by the
// CR 608.2b re-check, which counters this activation by game rules
// before the effect runs. The guard is for the ordinary "no target
// slot" defensive read every effect in the catalog makes.
func copyTargetedAbility(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return CopyAbility{
		ItemID:           id,
		Controller:       item.Controller,
		ChooseNewTargets: true,
	}.Apply(ctx)
}
