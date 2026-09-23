package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harsh Mentor — Creature — Human Cleric {1}{R}, 2/2:
//
//	"Whenever an opponent activates an ability of an artifact,
//	 creature, or land on the battlefield, if it isn't a mana ability,
//	 this creature deals 2 damage to that player."
//
// The first card to WATCH somebody ELSE's activation (#1210), and the
// proof that #1184 built its event wide enough: Rangers' Refueler
// reads `ByYou` plus the exhaust bit, this one reads `ByAnOpponent`
// and no exhaust test, off the same `EventActivateAbility`. The row
// closed on that event rather than on a second one, which is what its
// doc comment promised.
//
// THREE CLAUSES, THREE ARGUMENTS, no card logic:
//
//   - "an opponent" is the event's ACTOR — the player who activated —
//     not the source's controller. They are the same player for a
//     permanent's own ability and are not for one activated from
//     somebody else's permanent, and the card asks about the
//     activator.
//   - "if it isn't a mana ability" is `includeMana: false`, which
//     means the trigger does not WATCH EventManaAbilityActivated at
//     all. CR 605.1a gives mana abilities their own event kind
//     precisely so a watcher can tell, and expressing the clause as
//     the absence of a kind rather than as a predicate is what makes
//     it impossible for a card file to forget.
//   - "of an artifact, creature, or land on the battlefield" runs
//     against the ability's SOURCE object, looked up live. "On the
//     battlefield" is printed, so a source that has already left —
//     a Lotus Petal that sacrificed itself to its own cost — does not
//     trigger this, which is both the rule and the safe direction.
//
// "THAT PLAYER" is the activator, read off the event, not "target
// opponent": the ability does not target, so hexproof-style
// protections and "can't be the target" effects do nothing about it
// and nothing is re-checked at resolution (CR 608.2b).
//
// It triggers once PER ACTIVATION, not per ability: an opponent
// cracking three fetchlands takes six.
//
// The intervening-if ("if it isn't a mana ability") is part of the
// trigger condition here rather than a CR 603.4 re-check, because it
// is a fact about the event and not about the board — there is
// nothing that could stop being true between the trigger and its
// resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "168336e2-d795-4b75-bf21-ce128b0dd7b0",
		Name:         "Harsh Mentor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnOpponentActivates(
				"Harsh Mentor — 2 damage to that player",
				Or(Artifact(), Creature(), Land()),
				false,
				func(activator uuid.UUID) Effect {
					return func(g *game.Game, item *game.StackItem) error {
						return DealDamage{
							Source: item.SourceCardID,
							Target: activator,
							Amount: 2,
						}.Apply(NewContext(g, item))
					}
				}),
		},
	})
}
