package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ExileThisTokenAtEndStepGrant is the quoted ability "At the beginning of
// the end step, exile this token." as a named bundle (ability_grant.go), for
// a copy effect that grants it as part of the copying process (Heat
// Shimmer). SacrificeThisTokenAtEndStepGrant's sibling: it is the token's
// own triggered ability, firing at the beginning of every end step it sees,
// and it exiles the object it is on.
func ExileThisTokenAtEndStepGrant(key, label string) AbilityGrant {
	return AbilityGrant{
		Key:  key,
		Text: "At the beginning of the end step, exile this token.",
		Triggered: []game.TriggeredAbility{
			AtEachStep(game.StepEnd, label, func(g *game.Game, item *game.StackItem) error {
				return ExileTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}),
		},
	}
}

// copyTheTriggeringSpellWithNewTargets is "copy that spell. You may choose new
// targets for the copy" as a cast trigger's whole body (Double Vision,
// Kalamax, the Stormsire): the copy is made under the trigger's controller
// from last-known information if the original has already left the stack
// (CR 608.2h), with the new-targets choice its controller is owed (CR
// 707.10c). It is created, not cast, so it triggers no cast trigger.
func copyTheTriggeringSpellWithNewTargets(g *game.Game, item *game.StackItem) error {
	return CopySpell{
		StackID:          item.Trigger.Event.CardID,
		Controller:       item.Controller,
		ChooseNewTargets: true,
		FromLastKnown:    true,
	}.Apply(NewContext(g, item))
}
