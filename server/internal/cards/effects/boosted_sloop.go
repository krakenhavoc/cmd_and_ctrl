package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boosted Sloop — Artifact — Vehicle {1}{U}{R}, 3/3 (Aetherdrift, #4464):
//
//	"Menace
//	 Whenever you attack, draw a card, then discard a card.
//	 Crew 1"
//
// "Whenever you attack" is Bard, Heir of Girion's: one trigger per attack
// declaration however many creatures attack (OncePerBatch over
// attackDeclaredByYou), and it does not ask that the Sloop be among them,
// so an uncrewed Sloop loots too. The loot is lootOne: the discard opens
// once the draw is done, so the drawn card may be the one discarded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cec10abd-2a2d-4d89-9ecc-ea50c2ac1538",
		Name:            "Boosted Sloop",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Boosted Sloop — draw a card, then discard a card", func(g *game.Game, item *game.StackItem) error {
				return lootOne(g, item, 1)
			})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Boosted Sloop"),
		}},
	})
}
