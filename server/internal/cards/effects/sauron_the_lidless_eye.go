package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sauron, the Lidless Eye — Legendary Creature — Avatar Horror {3}{B}{R},
// 4/4 (The Hobbit):
//
//	"When Sauron enters, gain control of target creature an opponent
//	 controls until end of turn. Untap it. It gains haste until end of
//	 turn.
//	 {1}{B}{R}: Creatures you control get +2/+0 until end of turn. Each
//	 opponent loses 2 life."
//
// The enter trigger is the shared threaten body (threatenAndGrant: control
// until end of turn, untap, haste), the Act of Treason and Captivating
// Crew shape, on a trigger with a target (CR 603.3d: chosen when it goes on
// the stack; an opponent's creature that is gone by resolution fizzles it).
// The pump is a BoostUntilEOT over the creatures you control as it
// resolves (CR 611.2c: the set is fixed on resolution), so it counts the
// stolen creature that Sauron's first ability just gave you, and a
// creature that arrives later this turn is not boosted. The life loss is
// the second sentence of the same ability, each opponent, once per resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00b9d3a3-fd64-4757-9159-b3af06b5f5b1",
		Name:         "Sauron, the Lidless Eye",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Sauron — gain control of target creature an opponent controls until end of turn, untap it, it gains haste",
				sauronSteal),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}{R}: Creatures you control get +2/+0 until end of turn. Each opponent loses 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{B}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (BoostUntilEOT{
					Match: And(Creature(), YouControl()),
					Power: 2,
					Label: "Sauron — +2/+0",
				}).Apply(ctx); err != nil {
					return err
				}
				return EachOpponentLosesLife{N: 2}.Apply(ctx)
			},
		}},
	})
}

func sauronSteal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return threatenAndGrant(ctx, id, "Sauron", "haste")
}
