package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Chief Warg — Legendary Creature — Wolf {2}{B}{G}, 3/3 (The Hobbit):
//
//	"Menace
//	 Ferocious — Whenever you attack while you control a creature with
//	 power 4 or greater, you draw a card and lose 1 life."
//
// "Whenever you attack" is once per declaration however many creatures
// attack (OncePerBatch, the Sparring Regimen shape). The power check is
// a condition on the event, read as attackers are declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5ebe8de1-aa3d-410d-b43d-1685259c7a97",
		Name:            "The Chief Warg",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && b24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
			}, "The Chief Warg — ferocious: you draw a card and lose 1 life", b36DrawAndLoseOne)),
		},
	})
}
