package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chitterspitter — Artifact {2}{G} (EDHREC rank 3132):
//
//	"At the beginning of your upkeep, you may sacrifice a token. If
//	 you do, put an acorn counter on this artifact.
//	 Squirrels you control get +1/+1 for each acorn counter on this
//	 artifact.
//	 {G}, {T}: Create a 1/1 green Squirrel creature token."
//
// The Squirrel deck's engine: make a Squirrel every turn, feed one
// to the Chitterspitter every upkeep, and the rest grow. Three
// abilities:
//
//   - The upkeep trigger asks "you may sacrifice a token" as it
//     RESOLVES (maySacrificeOne, CR 608.2c): any token you control
//     then, hexproof or not, and the acorn counter follows only a
//     token that really went (b29SacrificeChosenTokenThenAcorn).
//   - The lord is the scaling anthem over Squirrels the controller
//     controls, reading the acorn counters on every recompute
//     (b29SquirrelsYouControlPerAcornCounter).
//   - The activation is a CR 602 ability with a mana-and-tap cost,
//     no summoning sickness (an artifact), making one Squirrel.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4a338863-d599-46e7-9c30-e11b898ae1b0",
		Name:         "Chitterspitter",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b29SquirrelsYouControlPerAcornCounter(),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Key:    "Chitterspitter — sacrifice a token, put an acorn counter on it",
			Effect: b29SacrificeChosenTokenThenAcorn,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{G}, {T}: Create a 1/1 green Squirrel creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{G}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 green Squirrel"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
