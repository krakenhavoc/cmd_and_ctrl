package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Iron Mastiff — Artifact Creature — Dog {4}, 4/4 (Commander Legends: Battle
// for Baldur's Gate, #5948):
//
//	"Whenever this creature attacks, roll a d20 for each player being attacked
//	 and ignore all but the highest roll.
//	 1—9 | This creature deals damage equal to its power to you.
//	 10—19 | This creature deals damage equal to its power to defending player.
//	 20 | This creature deals damage equal to its power to each opponent."
//
// One d20 for each player being attacked by any creature this combat, a
// planeswalker's or battle's controller included (CR 506.2a defending
// players), all rolled together and every roll but the highest ignored
// (RollDice{IgnoreLowest: n-1}, CR 706.6). The result is read from the kept roll
// (CR 706.3a). The Mastiff deals the damage, as many as its power as the ability
// resolves (or its last known power if it has left); "defending player" is the
// player it is attacking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d00e24f-c89f-49fa-869e-d788021aa084",
		Name:         "Iron Mastiff",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Iron Mastiff — roll a d20 for each player being attacked and keep the highest", func(g *game.Game, item *game.StackItem) error {
				attacked := map[uuid.UUID]bool{}
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.AttackingTarget == uuid.Nil {
						continue
					}
					if p := g.DefendingPlayerForAttackForEffect(c.AttackingTarget); p != uuid.Nil {
						attacked[p] = true
					}
				}
				n := max(len(attacked), 1)
				return effects.RollDice{Sides: 20, N: n, IgnoreLowest: n - 1, Then: ironMastiffRolled}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// ironMastiffRolled is the results table, read off the highest roll. An
// on-disk key: never renamed, never reused.
var ironMastiffRolled = effects.DiceTail("dice/iron-mastiff-rolled", func(ctx *effects.Context, rolls []int) error {
	g, mastiff := ctx.Game, ctx.Source()
	power := effects.B43PowerNowOrLastKnown(g, mastiff)
	if power <= 0 {
		return nil
	}
	var victims []uuid.UUID
	switch result := rolls[0]; {
	case result >= 20:
		victims = ctx.Opponents()
	case result >= 10:
		if p := g.DefendingPlayerForAttackerForEffect(mastiff); p != uuid.Nil {
			victims = []uuid.UUID{p}
		}
	default:
		victims = []uuid.UUID{ctx.Controller()}
	}
	for _, v := range victims {
		if err := (effects.DealDamage{Source: mastiff, Target: v, Amount: power}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
