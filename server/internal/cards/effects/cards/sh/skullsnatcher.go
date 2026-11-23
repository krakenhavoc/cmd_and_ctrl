package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skullsnatcher — Creature — Rat Ninja {1}{B}, 2/1 (Betrayers of
// Kamigawa, #5602):
//
//	"Ninjutsu {B} ({B}, Return an unblocked attacker you control to hand:
//	 Put this card onto the battlefield from your hand tapped and
//	 attacking.)
//	 Whenever this creature deals combat damage to a player, exile up to
//	 two target cards from that player's graveyard."
//
// The keyword is Ninjutsu("{B}") (ninjutsu.go). "That player's graveyard"
// is Tinybones, the Pickpocket's clause: built from the trigger's event as
// it goes on the stack, the cards owned by the player the damage named.
// "Up to two" lets the controller pick none, one or two; the exile is
// ExileTargetCards over the picks still there as it resolves.
//
// No simplification.
func init() {
	damage := effects.WheneverThisDealsCombatDamageToAPlayer(
		"Skullsnatcher — exile up to two target cards from that player's graveyard",
		effects.ExileTargetCards)
	damage.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		return effects.TargetCardInGraveyard("up to two target cards from that player's graveyard",
			effects.OwnedBy(tc.Event.Target)).WithCount(0, 2)
	}
	effects.Register(effects.Spec{
		OracleID:     "5326381e-e5e7-4c3e-9565-4837b3ffaec0",
		Name:         "Skullsnatcher",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Ninjutsu("{B}")},
		Triggered:    []game.TriggeredAbility{damage},
	})
}
