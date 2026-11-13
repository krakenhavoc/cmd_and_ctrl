package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battlefield Medic — Creature — Human Cleric {1}{W}, 1/1 (Onslaught):
//
//	"{T}: Prevent the next X damage that would be dealt to target creature
//	 this turn, where X is the number of Clerics on the battlefield."
//
// Kitsune Healer's charged shield (PreventNextDamage, CR 615.7) with a
// counted size: X is every player's Clerics, the Medic included while it
// is there, counted as the ability resolves (Airborne Aid's table-wide
// count). "Cleric" is a word of the card's text (CR 612.2). With no
// Cleric left by then, X is 0 and nothing is prevented. A target gone by
// resolution gets no shield (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a600af0b-ea04-4813-b700-f51820868c86",
		Name:         "Battlefield Medic",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Prevent the next X damage that would be dealt to target creature this turn, where X is the number of Clerics on the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				x := b42CountCreaturesWithSubtype(g, g.TextReaderForEffect(item).TextSubtype("Cleric"))
				return PreventNextDamage{Target: ts[0].ID, Amount: x, Label: "Battlefield Medic: prevent the next X damage"}.Apply(ctx)
			},
		}},
	})
}
