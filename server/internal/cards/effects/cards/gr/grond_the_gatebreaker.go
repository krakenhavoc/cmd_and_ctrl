package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grond, the Gatebreaker — Legendary Artifact — Vehicle {3}{B}, 5/5 (The
// Lord of the Rings: Tales of Middle-earth):
//
//	"Trample
//	 As long as it's your turn and you control an Army, Grond is an
//	 artifact creature.
//	 Crew 3"
//
// Phoenix Fleet Airship's conditional animation (selfIsArtifactCreatureWhile,
// a layer-4 static on the Vehicle read on every layer pass): it is a
// creature while its controller is the active player and controls an Army
// (a permanent with the Army subtype, an amassed token or a card). Crew is
// the ordinary CR 702.122 ability, so it can still be animated on any
// turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ea52d9e8-5c25-429a-8de4-feb4558bdb99",
		Name:            "Grond, the Gatebreaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{effects.SelfIsArtifactCreatureWhile(
			"As long as it's your turn and you control an Army, Grond is an artifact creature.",
			func(g *game.Game, source *game.Card) bool {
				if effects.ActivePlayerIDOf(g) != source.Controller {
					return false
				}
				army := source.TextSubtype("Army")
				for i := range g.Battlefield.Cards {
					if c := &g.Battlefield.Cards[i]; c.Controller == source.Controller && effects.CardHasSubtype(*c, army) {
						return true
					}
				}
				return false
			})},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Grond, the Gatebreaker"),
		}},
	})
}
