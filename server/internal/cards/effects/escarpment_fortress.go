package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Escarpment Fortress — Creature — Wall {4}{W}, 3/5 (Assassin's Creed,
// #5948):
//
//	"Defender
//	 Reach
//	 Other creatures you control get +1/+0.
//	 Whenever you attack with two or more creatures, draw a card."
//
// The anthem is Lord of the Accursed's TribalAnthem with no tribe: every
// other creature you control, read through the layers. The trigger is one per
// attack declaration (Adeline's OncePerBatch over attackDeclaredByYou), with
// two or more of your creatures attacking at the declaration, so an extra combat
// is a second declaration and triggers again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5c3f45cb-d2f0-4dac-b124-18e8f6d7e2dc",
		Name:            "Escarpment Fortress",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Others: true, YoursOnly: true}, 1, 0),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && len(b13AttackingCreaturesYouControl(g, source.Controller)) >= 2
			}, "Escarpment Fortress — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
