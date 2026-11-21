package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Peregrine Dynamo — Legendary Artifact Creature — Construct {3}, 1/5
// (Dominaria United Commander):
//
//	"Haste
//	 {1}, {T}: Copy target activated or triggered ability you control from
//	 another legendary source that's not a commander. You may choose new
//	 targets for the copy. (Mana abilities can't be targeted.)"
//
// Lithoform Engine's first ability (Strionic Resonator's copy, CR 707.10) with a
// narrower clause. "From another legendary source" is the ability's source as
// it is now: legendary, and not the Dynamo itself (Another, which drops an
// ability whose source is the clause's own source, as it does for a card).
// "That's not a commander" reads the source's commander designation
// (CR 903.3): a commander card is excluded whoever controls it, a token copy of
// one is not a commander and is allowed. An ability whose source has left the
// battlefield is read as the card it is in its new zone (AbilityFromSource); a
// token source that has ceased to exist is no longer found. Mana abilities
// never use the stack (CR 605.3b), so they are never offered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "da21f1a7-65d0-46d0-8ad1-f84a387ba753",
		Name:            "The Peregrine Dynamo",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, {T}: Copy target activated or triggered ability you control from another legendary source that's not a commander. You may choose new targets for the copy.",
			Cost:  effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.Another(effects.AbilityOnStack("target activated or triggered ability you control from another legendary source that's not a commander",
				effects.AnAbilityYouControl(),
				effects.AbilityFromSource(effects.And(effects.Legendary(), peregrineDynamoNotACommander)))),
			Effect: effects.CopyTargetedAbility,
		}},
	})
}

// peregrineDynamoNotACommander is "that's not a commander" for the source of
// the ability.
func peregrineDynamoNotACommander(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return !c.IsCommander
}
