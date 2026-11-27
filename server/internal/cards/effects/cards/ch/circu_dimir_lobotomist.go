package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Circu, Dimir Lobotomist — Legendary Creature — Human Wizard {2}{U}{B}, 2/3
// (Ravnica: City of Guilds):
//
//	"Whenever you cast a blue spell, exile the top card of target player's library.
//	 Whenever you cast a black spell, exile the top card of target player's library.
//	 Your opponents can't cast spells with the same name as a card exiled with
//	 Circu."
//
// Two triggers, one per printed ability: a blue spell and a black spell each
// target a player as they go on the stack, and a spell that is both triggers both
// (CR 603.2c). The exile is linked to this object (CR 607.2a, Mindreaver's);
// a library with no cards exiles nothing. The restriction is a static read live off
// the cards exiled with this Circu, so it lifts when Circu leaves and a Circu that
// returns is a new object with nothing exiled (CR 400.7). It bars Circu's
// controller's opponents only.
//
// No simplification.
func init() {
	exile := func(color, label string) game.TriggeredAbility {
		return effects.Targeting(effects.On(game.EventCast, effects.YouCast(effects.OfColor(color)), label, effects.ExileTopCardsOfTargetPlayerLinked(1)),
			effects.TargetPlayer("target player"))
	}
	effects.Register(effects.Spec{
		OracleID:     "c87d547b-00a4-4fdd-ade4-f3f032e5ea3b",
		Name:         "Circu, Dimir Lobotomist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			exile("U", "Circu, Dimir Lobotomist — exile the top card of target player's library (blue spell)"),
			exile("B", "Circu, Dimir Lobotomist — exile the top card of target player's library (black spell)"),
		},
		CastRestrictions: []game.CastRestriction{effects.OpponentsCantCastSpellsNamedLikeTheCardsExiledWithThis()},
	})
}
