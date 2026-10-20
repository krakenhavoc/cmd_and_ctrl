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
// "That player controls" is ControlledByThatPlayer (#3946): the clause
// is judged against the player THIS trigger's damage event named, in
// the one shared target walk, so the picker, the announce check and the
// resolution re-check agree even when two Predators hit two players in
// one combat.
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
			Targets: ControlledByThatPlayer(TargetPermanent("target artifact or enchantment that player controls",
				Or(Artifact(), Enchantment()))),
			Effect: destroyFirstLegalTarget,
		}},
	})
}
