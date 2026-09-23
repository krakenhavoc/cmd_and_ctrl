package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lithoform Engine — Legendary Artifact {4} (#1223):
//
//	"{2}, {T}: Copy target activated or triggered ability you control.
//	 You may choose new targets for the copy.
//	 {3}, {T}: Copy target instant or sorcery spell you control. You
//	 may choose new targets for the copy.
//	 {4}, {T}: Copy target permanent spell you control. (The copy
//	 becomes a token.)"
//
// Three abilities, three target clauses, one sentence each. The card
// is here as the proof that the ability copy and the spell copy are
// the same rule reached two ways: the first ability is
// game.CopyAbilityForEffect, the second and third are
// game.CopySpellForEffect, and all three share CR 707.10c's
// "you may choose new targets" prompt, its gate and its frame
// (spell_copy.go's copyFrame).
//
// The third ability needs no clause of its own about tokens: CR
// 608.3f already makes a resolving copy of a permanent spell become
// one, through #923's single token-creation path (ADR 0043's
// 2026-09-18 amendment). The reminder text is describing the engine's
// existing behaviour, not asking for any.
//
// "You control" on all three, and the mana-ability exclusion Strionic
// Resonator's reminder text spells out applies here too — a mana
// ability never uses the stack (CR 605.3b), so the first ability's
// picker cannot offer one.
func init() {
	Register(Spec{
		OracleID:     "0bf299da-1854-4153-baea-3cee2eb01ee8",
		Name:         "Lithoform Engine",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, {T}: Copy target activated or triggered ability you control",
				Cost:    Plus(ManaCost("{2}"), TapCost()),
				Targets: AbilityOnStack("target activated or triggered ability you control", AnAbilityYouControl()),
				Effect:  copyTargetedAbility,
			},
			{
				Label:   "{3}, {T}: Copy target instant or sorcery spell you control",
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Targets: instantOrSorcerySpell("target instant or sorcery spell you control", YouControl()),
				Effect:  copyTargetedSpell,
			},
			{
				Label:   "{4}, {T}: Copy target permanent spell you control",
				Cost:    Plus(ManaCost("{4}"), TapCost()),
				Targets: TargetSpell("target permanent spell you control", Permanent(), YouControl()),
				Effect:  copyTargetedSpell,
			},
		},
	})
}

// copyTargetedSpell is the body Lithoform Engine's two spell modes
// share: copy the spell the activation named, with the CR 707.10c
// re-target offer.
//
// One body for both, because the difference between them is entirely
// in the TARGET CLAUSE — an instant or sorcery copy ceases to exist
// as it resolves and a permanent-spell copy becomes a token, and CR
// 608.3f settles that inside the engine off the copied spell's own
// types rather than off which ability made the copy.
func copyTargetedSpell(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return CopySpell{
		StackID:          id,
		Controller:       item.Controller,
		ChooseNewTargets: true,
	}.Apply(ctx)
}
