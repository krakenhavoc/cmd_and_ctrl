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
//   - The upkeep trigger is The Goose Mother's shape: the "you may"
//     is the trigger's optional prompt, the token is chosen as the
//     trigger's target (any token you control — the picker the
//     engine has for a choice among your own permanents), and on
//     resolution it is sacrificed and the acorn counter follows
//     only from the sacrifice (b29SacrificeChosenTokenThenAcorn).
//   - The lord is the scaling anthem over Squirrels the controller
//     controls, reading the acorn counters on every recompute
//     (b29SquirrelsYouControlPerAcornCounter).
//   - The activation is a CR 602 ability with a mana-and-tap cost,
//     no summoning sickness (an artifact), making one Squirrel.
//
// Sandbox simplification, declared (the Goose Mother's caveat):
// the token is chosen when the trigger goes on the stack, not on
// resolution, so an opponent who removes it in response fizzles
// the trigger and no counter is placed, where printed you would
// pick another. Weaker, never stronger.
func init() {
	Register(Spec{
		OracleID:     "4a338863-d599-46e7-9c30-e11b898ae1b0",
		Name:         "Chitterspitter",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You pick the token to sacrifice when the upkeep trigger goes on the stack rather than on resolution, so opponents can respond to the choice."},
		Static: []game.StaticAbility{
			b29SquirrelsYouControlPerAcornCounter(),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Chitterspitter — sacrifice a token to put an acorn counter on it?"},
			Targets:        TargetPermanent("a token you control", IsTokenPredicate(), YouControl()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Chitterspitter — sacrifice a token, put an acorn counter on it", b29SacrificeChosenTokenThenAcorn)
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{G}, {T}: Create a 1/1 green Squirrel creature token.",
			Cost:  Plus(ManaCost("{G}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: b29GreenSquirrelToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
