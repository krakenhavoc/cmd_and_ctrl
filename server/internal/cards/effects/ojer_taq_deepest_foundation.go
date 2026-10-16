package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ojer Taq, Deepest Foundation — Legendary Creature — God {4}{W}{W},
// 6/6, the front face of a transforming card whose back is Temple of
// Civilization (EDHREC rank 921):
//
//	"Vigilance
//	 If one or more creature tokens would be created under your
//	 control, three times that many of those tokens are created
//	 instead.
//	 When Ojer Taq dies, return it to the battlefield tapped and
//	 transformed under its owner's control."
//
// The tripler is Parallel Lives' CR 614 replacement on token creation
// narrowed to CREATURE tokens: only the groups whose template is a
// creature are multiplied, so an instruction that makes a Soldier and
// a Treasure together makes three Soldiers and one Treasure. Two Ojer
// Taqs, or Ojer Taq and a doubler, multiply in turn (CR 616.1).
//
// The dies trigger is the Ojer cycle's (ojerDiesReturnTransformed).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        ojerTaqOracleID,
		Name:            "Ojer Taq, Deepest Foundation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{ojerDiesReturnTransformed("Ojer Taq", nil)},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventTokenCreated},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCreateTokens || ev.TokenCount() <= 0 || ev.TokenController != src.Controller {
					return false
				}
				return ev.TokenTemplatesMatch(func(c game.Card) bool { return c.IsCreature() })
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				for i := range ev.TokenGroups {
					if ev.TokenGroups[i].Count > 0 && ev.TokenGroups[i].Template.IsCreature() {
						ev.TokenGroups[i].Count *= 3
					}
				}
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Ojer Taq, Deepest Foundation: three times that many creature tokens",
		}},
	})
}

// Temple of Civilization — Land, the back face of Ojer Taq, Deepest
// Foundation:
//
//	"(Transforms from Ojer Taq, Deepest Foundation.)
//	 {T}: Add {W}.
//	 {2}{W}, {T}: Transform this land. Activate only if you attacked
//	 with three or more creatures this turn and only as a sorcery."
//
// The condition is Windbrisk Heights' (attackedWithAtLeastThisTurn):
// three different creatures you declared as attackers this turn, in any
// number of combats, whether or not they are still on the battlefield.
//
// No simplification.
func init() {
	attackedWithThree := attackedWithAtLeastThisTurn(3)
	Register(Spec{
		OracleID:     ojerTaqOracleID + "#1",
		Name:         "Temple of Civilization",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "{T}: Add {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{2}{W}, {T}: Transform this land. Activate only if you attacked with three or more creatures this turn and only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{W}"), TapCost()),
			SorcerySpeed: true,
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return attackedWithThree(g, controller)
			},
			Effect: Do(TransformThis{}),
		}},
	})
}

const ojerTaqOracleID = "486bb9a5-73f1-4cec-b097-fb07ac80b72e"
