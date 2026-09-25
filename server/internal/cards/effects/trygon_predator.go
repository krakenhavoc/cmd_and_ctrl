package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trygon Predator — Creature — Beast {1}{G}{U}, 2/3 (EDHREC rank
// 3434):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, you may
//	 destroy target artifact or enchantment that player controls."
//
// The Simic flier that eats a rock every turn. Flying rides
// PrintedKeywords; the trigger is the ordinary "deals combat damage
// to a player" condition on the Predator itself, optional ("you
// may"), and targeted.
//
// "That player controls" is the interesting clause: a target
// predicate is not handed the trigger's event, so it cannot be told
// who was hit. It reads the per-turn tally instead — the players a
// Trygon Predator under the caster's control dealt combat damage to
// this turn, recorded at the damage while the Predator is still on
// the battlefield to be named
// (b32ArtifactOrEnchantmentOfPlayerHitByYourTrygonPredator) —
// and the trigger's body re-checks the pick against the specific
// player its own event named. With one Predator the picker offers
// exactly the victim's artifacts and enchantments, as printed; with
// two Predators connecting with two players in the same combat the
// picker offers both players' permanents and a pick under the wrong
// player's control does nothing at resolution. Never stronger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c744b5f4-fbcf-48b8-9d60-5e9c6ac297e0",
		Name:            "Trygon Predator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			Key:     b32TrygonPredatorLabel,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Trygon Predator — destroy an artifact or enchantment that player controls?",
			},
			Targets: TargetPermanent("target artifact or enchantment that player controls",
				b32ArtifactOrEnchantmentOfPlayerHitByYourTrygonPredator),
			Effect: b32DestroyChosenIfControlledByTriggerVictim,
		}},
	})
}
