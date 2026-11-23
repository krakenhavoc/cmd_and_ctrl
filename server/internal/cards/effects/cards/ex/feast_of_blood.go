package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Feast of Blood — Sorcery {1}{B} (Jumpstart 2022, #5948):
//
//	"Cast this spell only if you control two or more Vampires.
//	 Destroy target creature. You gain 4 life."
//
// The restriction is a CastCondition, checked once at announce and never
// again (the spell is not countered if the Vampires leave). Creature types
// are read through the layered view, so a changeling counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fed05740-01b2-4c7a-8b97-55e64837c07f",
		Name:         "Feast of Blood",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		CastCondition: func(g *game.Game, controller uuid.UUID, spell game.Card) bool {
			n := 0
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == controller && c.IsCreature() && c.HasSubtype(spell.TextSubtype("Vampire")) {
					n++
				}
			}
			return n >= 2
		},
		CastConditionLabel: "Cast this spell only if you control two or more Vampires.",
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: 4}.Apply(ctx)
		},
	})
}
