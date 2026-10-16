package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cleaver Skaab — Creature — Zombie Horror {3}{U}, 2/4:
//
//	"{3}, {T}, Sacrifice another Zombie: Create two tokens that are
//	 copies of the sacrificed creature."
//
// The copies are of the sacrificed OBJECT, by its last-known copiable
// values (CR 707.2, CR 608.2h): TokenCopyTemplateForObjectForEffect
// reads PermanentInfo.Copiable, so a sacrificed Clone that was copying
// something makes two of what it was copying, and counters, damage and
// anthems are not copied. The cost names "another Zombie"; the effect
// says "creature", so a Zombie that was not a creature as it was
// sacrificed is still copied.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9fa66e03-7b20-43f6-8c53-1ce18169c382",
		Name:         "Cleaver Skaab",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice another Zombie: Create two tokens that are copies of the sacrificed creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{3}"), TapCost(), SacrificeAnotherN(1, "another Zombie", HasSubtype("Zombie"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				refs := ctx.Paid().SacrificedObjects
				if len(refs) == 0 {
					return nil
				}
				tmpl, ok := g.TokenCopyTemplateForObjectForEffect(refs[0])
				if !ok {
					return nil
				}
				return createTokenCopies(ctx, ctx.Controller(), tmpl, 2)
			},
		}},
	})
}
