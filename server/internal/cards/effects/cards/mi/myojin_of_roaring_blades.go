package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Myojin of Roaring Blades — Legendary Creature — Spirit {5}{R}{R}{R},
// 7/4 (Kamigawa: Neon Dynasty Commander, #5493):
//
//	"Myojin of Roaring Blades enters with an indestructible counter on it
//	 if you cast it from your hand.
//	 Remove an indestructible counter from Myojin of Roaring Blades: It
//	 deals 7 damage to each of up to three targets."
//
// The counter is Myojin of Night's Reach's entry clause; the damage is
// Jeska's "each of up to three targets", the Myojin the source (its last
// known information if it has left).
//
// No simplification.
func init() {
	upToThree := effects.TargetAny()
	upToThree.Label = "up to three targets"
	upToThree.Min, upToThree.Max = 0, 3
	effects.Register(effects.Spec{
		OracleID:                   "20d20254-0f77-429a-a74e-0aec08d18f2d",
		Name:                       "Myojin of Roaring Blades",
		Completeness:               effects.CompletenessFull,
		EntersWithCountersFromCast: effects.NeoMyojinEntry(),
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove an indestructible counter from Myojin of Roaring Blades: It deals 7 damage to each of up to three targets.",
			Cost:    effects.NeoRemoveIndestructible(),
			Targets: upToThree,
			Purpose: roaringBladesPurpose(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				var ids []uuid.UUID
				for _, t := range ctx.LegalTargets() {
					ids = append(ids, t.ID)
				}
				return effects.DealDamageToEachThen(ctx, ids, 7, effects.RecipientThen{})
			},
		}},
	})
}

// roaringBladesPurpose declares the damage to its targets (ADR 0126) and,
// since "up to three" lets it be announced with none, what the row
// answers (ADR 0142): it removes other permanents.
func roaringBladesPurpose() game.Purpose {
	p := effects.ForTargets(effects.DamageToTarget(0, 7))
	p.Answers = game.AnswerRemove
	return p
}
