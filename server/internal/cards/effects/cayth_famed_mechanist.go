package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d18a014e-16a0-4641-9546-ff3729437eed",
		Name:         "Cayth, Famed Mechanist",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Fabricate("Cayth, Famed Mechanist", 1)},
		Grants: []AbilityGrant{{
			Key:       caythFabricateGrant,
			Triggered: []game.TriggeredAbility{Fabricate("Cayth, Famed Mechanist", 1)},
			Text:      "Fabricate 1 (When this creature enters, put a +1/+1 counter on it or create a 1/1 colorless Servo artifact creature token.)",
		}},
		Static: []game.StaticAbility{
			GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return otherCreaturesYouControl(target, g, source) && !IsToken(*target)
			}, caythFabricateGrant),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Choose one — populate; or proliferate.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Modes: ChooseOne(
				ModeDoing("Populate.", nil, func(_ *game.StackItem, ctx *Context, _ int) error {
					return Populate{}.Apply(ctx)
				}),
				ModeDoing("Proliferate.", nil, func(_ *game.StackItem, ctx *Context, _ int) error {
					return Proliferate{}.Apply(ctx)
				}),
			),
		}},
	})
}

const caythFabricateGrant = "cayth-famed-mechanist/fabricate"
