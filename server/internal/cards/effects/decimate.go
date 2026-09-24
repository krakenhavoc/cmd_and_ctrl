package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Decimate — Sorcery {2}{R}{G}:
//
//	"Destroy target artifact, target creature, target enchantment, and
//	 target land. (You can't cast this spell unless you have legal
//	 choices for all its targets.)"
//
// Four "target" words, four different predicates, one statement:
// TargetPermanent(...).Then(...) chains the artifact / creature /
// enchantment / land clauses (#764, ADR 0065 §1). Each clause is
// independently validated at announce, which is what the reminder
// text is describing — a caster with no land in play simply cannot
// cast Decimate, because the fourth clause has nothing to offer.
//
// CR 601.2c leaves it legal for one object to fill two of these
// clauses when it qualifies for both (an artifact creature could
// answer both the first and second target) — the card prints no
// "different" restriction, so neither clause sets Distinct.
//
// The four destructions land as ONE simultaneous event
// (DestroyPermanentsForEffect) rather than four sequential ones, so a
// Blood Artist watching the board sees every permanent Decimate
// killed at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4e5693f-12a0-451e-818d-d6efc7b4ed25",
		Name:         "Decimate",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target artifact", Artifact()).Then(
			TargetPermanent("target creature", Creature()),
			TargetPermanent("target enchantment", Enchantment()),
			TargetPermanent("target land", Land()),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, t := range item.Targets {
				if ctx.IsTargetLegal(t) {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			ctx.Game.DestroyPermanentsForEffect(ids)
			return nil
		},
	})
}
