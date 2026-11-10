package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aplan Mortarium — Plane — Alfava Metraxis (Doctor Who, #4870):
//
//	"Byzantium Radiation — At the beginning of your upkeep, put an
//	 exposure counter on Aplan Mortarium. Then you lose life equal to the
//	 number of exposure counters on it.
//	 Whenever chaos ensues, create two 2/2 black Alien Angel artifact
//	 creature tokens with first strike, vigilance, and "Whenever an
//	 opponent casts a creature spell, this token isn't a creature until
//	 end of turn.""
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6),
// so "your upkeep" is the planar controller's and the counters build up
// over every player's upkeep while the plane stays face up. The counters
// sit on the face-up card in the command zone
// (game.PutCountersOnFaceUpPlanarCardForEffect); a planeswalk away makes it
// a new object with none (CR 901.7a). The chaos ability (CR 311.7) creates
// the catalog's Alien Angel token, whose own trigger removes the creature
// type until end of turn: it stays an artifact and, as the rulings say,
// stops being an Alien and an Angel too (CR 205.1a), and one attacking or
// blocking is removed from combat (CR 506.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c9d6193-1d50-4035-8c85-98c384f461be",
		Name:         "Aplan Mortarium",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Byzantium Radiation — Aplan Mortarium: put an exposure counter on it, then lose life equal to its exposure counters",
				func(g *game.Game, item *game.StackItem) error {
					n, ok := g.PutCountersOnFaceUpPlanarCardForEffect(item.SourceCardID, aplanMortariumExposure, 1)
					if !ok || n <= 0 {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -n)
				}),
			WheneverChaosEnsues("Aplan Mortarium — create two 2/2 black Alien Angel artifact creature tokens",
				Do(CreateToken{Template: AlienAngelToken(), N: 2})),
		},
	})
}

// aplanMortariumExposure is the exposure counter's kind.
const aplanMortariumExposure = "exposure"

// AlienAngelToken is Aplan Mortarium's token: a 2/2 black Alien Angel
// artifact creature with first strike, vigilance, and "Whenever an
// opponent casts a creature spell, this token isn't a creature until end
// of turn."
func AlienAngelToken() game.Card { return tokenFromCatalog(printedAlienAngelToken) }

// printedAlienAngelToken is the Alien Angel as PRINTED. "This token" is the
// trigger's source; "an opponent" is an opponent of its controller now.
func printedAlienAngelToken() tokenTemplate {
	return tokenTemplate{
		Slug: "alien-angel",
		Card: game.Card{
			Name:      "Alien Angel",
			TypeLine:  "Token Artifact Creature — Alien Angel",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"B"},
			Keywords:  []string{"first strike", "vigilance"},
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AnOpponentCast(Creature()), "Alien Angel — this token isn't a creature until end of turn",
				thisIsntACreatureUntilEndOfTurn("Alien Angel — isn't a creature until end of turn")),
		},
		Text: "First strike, vigilance\nWhenever an opponent casts a creature spell, this token isn't a creature until end of turn.",
	}
}
