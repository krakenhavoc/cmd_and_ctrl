package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Consuming Ashes — Instant {2}{B}{B} (Outlaws of Thunder Junction,
// #4468):
//
//	"Exile target creature. If it had mana value 3 or less, surveil 2.
//	 (Look at the top two cards of your library, then put any number of
//	 them into your graveyard and the rest on top of your library in any
//	 order.)"
//
// Deluge of the Dead's shape (ExileThenIfItWas): the creature's mana value
// is read as it was on the battlefield, before the exile (CR 608.2h), and
// the surveil rides the exile as a keyed tail, so a table waiting on the
// exile is a restore point. A creature that is an illegal target as the
// spell resolves leaves the spell with nothing to do, and there is no
// surveil. A face-down creature has mana value 0 (CR 708.2a) and counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ef676f0-53c5-4efe-9341-e960d00dbc69",
		Name:         "Consuming Ashes",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return ExileThenIfItWas{
					Target: t.ID,
					Was: func(c game.Card) bool {
						mv, ok := c.ParsedManaValue()
						return ok && mv <= 3
					},
					Keyed: consumingAshesSurveil,
				}.Apply(ctx)
			}
			return nil
		},
	})
}

// consumingAshesSurveil is the exile's keyed "if it had mana value 3 or
// less, surveil 2", for the spell's controller. An on-disk key: never
// renamed, never reused.
var consumingAshesSurveil = IfItWasTail("if-it-was/consuming-ashes-surveil", func(ctx *Context) error {
	return Surveil{N: 2}.Apply(ctx)
}, nil)
