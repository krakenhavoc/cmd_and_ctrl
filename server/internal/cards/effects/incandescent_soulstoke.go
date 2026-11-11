package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Incandescent Soulstoke — Creature — Elemental Shaman {2}{R}, 2/2 (Lorwyn
// Eclipsed Commander, #5368):
//
//	"Other Elemental creatures you control get +1/+1.
//	 {1}{R}, {T}: You may put an Elemental creature card from your hand onto
//	 the battlefield. That creature gains haste until end of turn. Sacrifice
//	 it at the beginning of the next end step."
//
// The anthem is Cemetery Reaper's tribal lord over Elemental (a changeling
// counts); the ability is Meek Attack's put from hand with a tail that grants
// haste and schedules the sacrifice for that object. "May" is a real decline,
// and declining makes no sacrifice. The {T} makes it subject to summoning
// sickness.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "934bd9f9-937b-4c6b-a849-012ae6bd3321",
		Name:         "Incandescent Soulstoke",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Elemental"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, {T}: You may put an Elemental creature card from your hand onto the battlefield. That creature gains haste until end of turn. Sacrifice it at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{1}{R}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutFromHandOntoBattlefield{
					Match:    And(Creature(), AnySubtype("Elemental")),
					Optional: true,
					Label:    "Incandescent Soulstoke — you may put an Elemental creature card from your hand onto the battlefield",
					ThenTail: cmdr2S12SoulstokeEntered,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
