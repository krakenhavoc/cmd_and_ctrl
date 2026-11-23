package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cayth, Famed Mechanist — Legendary Creature — Dwarf Artificer
// {1}{U}{R}{W}, 3/3 (Modern Horizons 3 Commander, #5948):
//
//	"Fabricate 1 (When this creature enters, put a +1/+1 counter on it or
//	 create a 1/1 colorless Servo artifact creature token.)
//	 Other nontoken creatures you control have fabricate 1.
//	 {2}, {T}: Choose one —
//	 • Populate.
//	 • Proliferate."
//
// Cayth's own fabricate is the shared keyword trigger (fabricate_keyword.go).
// The grant is Pontiff of Blight's shape (ADR 0093): a layer-6 grant of that
// same trigger to every OTHER NONTOKEN creature its controller controls, so
// each recipient carries its own fabricate, the counter goes on the
// recipient, and the trigger ends with the recipient's abilities or Cayth's.
// A token entering is not given it. The activated ability is modal, the
// mode chosen as it is announced (CR 602.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d18a014e-16a0-4641-9546-ff3729437eed",
		Name:         "Cayth, Famed Mechanist",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Fabricate("Cayth, Famed Mechanist", 1)},
		Grants: []effects.AbilityGrant{{
			Key:       caythFabricateGrant,
			Triggered: []game.TriggeredAbility{effects.Fabricate("Cayth, Famed Mechanist", 1)},
			Text:      "Fabricate 1 (When this creature enters, put a +1/+1 counter on it or create a 1/1 colorless Servo artifact creature token.)",
		}},
		Static: []game.StaticAbility{
			effects.GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.OtherCreaturesYouControl(target, g, source) && !effects.IsToken(*target)
			}, caythFabricateGrant),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Choose one — populate; or proliferate.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Modes: effects.ChooseOne(
				effects.ModeDoing("Populate.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.Populate{}.Apply(ctx)
				}),
				effects.ModeDoing("Proliferate.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.Proliferate{}.Apply(ctx)
				}),
			),
		}},
	})
}

const caythFabricateGrant = "cayth-famed-mechanist/fabricate"
