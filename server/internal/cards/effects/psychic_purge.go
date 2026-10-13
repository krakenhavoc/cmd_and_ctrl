package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Psychic Purge — Sorcery {U}:
//
//	"Psychic Purge deals 1 damage to any target.
//	 When a spell or ability an opponent controls causes you to discard
//	 this card, that player loses 5 life."
//
// The spell is a one-damage Lightning Bolt. The second ability is the
// family's "this card" trigger (ADR 0153 §4), watching from the
// GRAVEYARD on a discard an opponent's spell or ability caused. "That
// player" is the controller of that spell or ability (CR 113.8), which
// the discard event records as its CauseController: a stolen permanent's
// ability makes its thief lose the life, not the permanent's owner.
//
// Caveat, declared: if madness or Rest in Peace sends the discarded card
// to exile instead of the graveyard, its ability does not trigger from
// exile (owner decision 6).
func init() {
	Register(Spec{
		OracleID:          "6af596d2-f075-4f55-b088-a5237fcdaa51",
		Name:              "Psychic Purge",
		Completeness:      CompletenessCaveats,
		Caveats:           []string{discardTriggerFromExileCaveat},
		DiscardByOpponent: game.DiscardByOpponentTrigger,
		Targets:           TargetAny(),
		Purpose:           ForTargets(DamageToTarget(0, 1)),
		OnResolve:         damageToFirstTarget(1),
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventDiscardCard, DiscardedByAnOpponent(true),
				"Psychic Purge — discarded: that player loses 5 life",
				func(g *game.Game, item *game.StackItem) error {
					cause := item.Trigger.Event.CauseController
					p := g.PlayerByIDForEffect(cause)
					if p == nil || p.Eliminated {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, cause, -5)
				})),
		},
	})
}
