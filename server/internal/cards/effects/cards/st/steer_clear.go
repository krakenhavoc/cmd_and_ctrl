package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steer Clear — Instant {W} (Outlaws of Thunder Junction):
//
//	"Steer Clear deals 2 damage to target attacking or blocking creature.
//	 Steer Clear deals 4 damage to that creature instead if you controlled
//	 a Mount as you cast this spell."
//
// "As you cast this spell" fixes the Mount check when the spell is cast
// (CR 601.2i): Spec.CastTimeValues stamps it once every cost is paid, so a
// Mount that dies in response does not shrink the damage, and one that
// arrives in response does not grow it (the card's ruling).
//
// No simplification.
const steerClearMount = "controlled_a_mount"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c22af49-c6ef-4de6-bb84-6b3e99879148",
		Name:         "Steer Clear",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		CastTimeValues: []game.CastTimeValue{{
			Key: steerClearMount,
			Value: func(g *game.Game, item *game.StackItem) int {
				mount := "Mount"
				if src, ok := g.LookupCardForEffect(item.SourceCardID); ok {
					mount = src.TextSubtype("Mount")
				}
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.HasSubtype(mount) {
						return 1
					}
				}
				return 0
			},
		}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			amount := 2
			if v, _ := ctx.CastValue(steerClearMount); v > 0 {
				amount = 4
			}
			return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: amount}.Apply(ctx)
		},
	})
}
