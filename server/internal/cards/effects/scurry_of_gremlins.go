package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scurry of Gremlins — Enchantment {2}{R}{W}:
//
//	"When this enchantment enters, create two 1/1 red Gremlin creature
//	 tokens. Then you get an amount of {E} (energy counters) equal to
//	 the number of creatures you control.
//	 Pay {E}{E}{E}{E}: Creatures you control get +1/+0 and gain haste
//	 until end of turn."
//
// "Then": the creatures are counted after the two Gremlins are made, so
// they count. The energy is counted at resolution, so the Purpose
// declares only the tokens (ADR 0126 §6). The activated ability's set
// of creatures is locked as it resolves (CR 611.2c), Overrun's shape.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ff24d18d-9c5d-4ffe-846f-a8d568b32724",
		Name:         "Scurry of Gremlins",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Scurry of Gremlins — two Gremlins, then {E} for each creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (CreateToken{Template: TokenCard("1/1 red Gremlin"), N: 2}).Apply(ctx); err != nil {
						return err
					}
					return GetEnergy{N: b04CreaturesControlled(g, item.Controller)}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Pay {E}{E}{E}{E}: Creatures you control get +1/+0 and gain haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    PayEnergy(4),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				yours := And(Creature(), YouControl())
				if err := (BoostUntilEOT{Match: yours, Power: 1, Label: "Scurry of Gremlins — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"haste"}, Label: "Scurry of Gremlins — haste"}.Apply(ctx)
			},
		}},
	})
}
