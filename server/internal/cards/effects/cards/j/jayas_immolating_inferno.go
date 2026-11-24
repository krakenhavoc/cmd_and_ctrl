package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jaya's Immolating Inferno — Legendary Sorcery {X}{R}{R} (Secret Lair Drop,
// #5462):
//
//	"(You may cast a legendary sorcery only if you control a legendary
//	 creature or planeswalker.)
//	 Jaya's Immolating Inferno deals X damage to each of up to three
//	 targets."
//
// The cast restriction is CR 205.4e's, checked at announce only. The targets
// are up to three of any kind (Jeska, Thrice Reborn's -X over a spell): each
// takes X damage as one instruction, a target that became illegal takes
// nothing and the others still do (CR 608.2b), and with none chosen the spell
// does nothing.
//
// No simplification.
func init() {
	upToThree := effects.TargetAny()
	upToThree.Label = "up to three targets"
	upToThree.Min, upToThree.Max = 0, 3
	effects.Register(effects.Spec{
		OracleID:           "f5f0deb0-070a-45b6-9b81-0bc8143f3040",
		Name:               "Jaya's Immolating Inferno",
		Completeness:       effects.CompletenessFull,
		XMatters:           true,
		CastCondition:      effects.LegendarySorcery(),
		CastConditionLabel: effects.LegendarySorceryLabel,
		Targets:            upToThree,
		Purpose:            effects.ForTargets(effects.DamageXToTarget(0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				ids = append(ids, t.ID)
			}
			return effects.DealDamageToEachThen(ctx, ids, ctx.X(), effects.RecipientThen{})
		},
	})
}
