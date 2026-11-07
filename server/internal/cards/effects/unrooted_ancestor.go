package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unrooted Ancestor — Creature — Spirit Cleric {2}{B}, 3/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flash
//	 {1}, Sacrifice another creature: This creature gains indestructible
//	 until end of turn. Tap it."
//
// Flash is a printed keyword. The cost is Yahenni's "sacrifice another
// creature" with a {1}; the Ancestor itself is not a legal payment. The
// effect is Underrealm Lich's: indestructible until end of turn, then the
// Ancestor taps. An already tapped Ancestor still gains indestructible.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "de5c6c9d-04a2-4287-8c6e-fdf56c411507",
		Name:            "Unrooted Ancestor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: This creature gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another creature", Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Unrooted Ancestor — indestructible until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				return TapTarget{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}
