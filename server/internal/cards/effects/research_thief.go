package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Research Thief — Artifact Creature — Moonfolk Wizard {4}{U}, 3/3
// (Foundations Commander):
//
//	"Flash
//	 Flying
//	 Whenever an artifact creature you control deals combat damage to a
//	 player, draw a card."
//
// One trigger per artifact creature that connects (CR 603.2c; Ohran
// Frostfang's per-creature shape), the Thief itself included. "Artifact
// creature" is read off the dealer as it is, so a creature that is both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ba5055ce-6809-47a4-bee6-fb97a5b60f8a",
		Name:            "Research Thief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !combatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				return ok && dealer.IsArtifact()
			}, "Research Thief — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
