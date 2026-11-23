package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thunderwave — Sorcery {2}{R}{R} (Commander Legends: Battle for Baldur's Gate,
// S82 GAP #5948):
//
//	"Roll a d20.
//	 1—9 | Thunderwave deals 3 damage to each creature.
//	 10—19 | You may choose a creature. Thunderwave deals 3 damage to each
//	 creature not chosen this way.
//	 20 | Thunderwave deals 6 damage to each creature your opponents control."
//
// The d20 goes through the one dice door and its tail reads the table. Each branch
// is one damage event from Thunderwave to the creatures on the battlefield at
// that moment (the set fixed before the first point lands, CR 615.8). 10—19 asks
// the caster for one creature to spare (a choice, not a target, so it may be an
// opponent's or their own, or none), then damages every other creature.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43de944b-e630-45f0-bce4-c66803e8b5c1",
		Name:         "Thunderwave",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.RollDice{Sides: 20, N: 1, Then: thunderwaveRolled}.Apply(ctx)
		},
	})
}

// thunderwaveRolled reads the table. An on-disk key.
var thunderwaveRolled = effects.DiceTail("dice/thunderwave-rolled", func(ctx *effects.Context, rolls []int) error {
	switch n := rolls[0]; {
	case n >= 20:
		return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.OpponentControls()), 6)
	case n >= 10:
		var creatures []uuid.UUID
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if c.IsCreature() {
				creatures = append(creatures, c.InstanceID)
			}
		}
		if len(creatures) == 0 || ctx.Item == nil {
			return nil
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    ctx.Controller(),
			FromPlayer: ctx.Controller(),
			Source:     ctx.Source(),
			Question:   "Thunderwave — you may choose a creature; Thunderwave deals 3 damage to each creature not chosen",
			Cards:      creatures,
			Min:        0,
			Max:        1,
			ThenKey:    thunderwaveSpared,
			LiveZone:   game.ZoneBattlefield,
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	return effects.DamageEachMatching(ctx, effects.Creature(), 3)
})

// thunderwaveSpared deals the 3 damage to every creature but the chosen one. An
// on-disk key.
var thunderwaveSpared = effects.AskedPick("choose-cards/thunderwave-spared", func(ctx *effects.Context, r game.CardsChosen) error {
	var spared uuid.UUID
	if len(r.Picked) > 0 {
		spared = r.Picked[0]
	}
	return effects.DamageEachMatching(ctx, func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.InstanceID != spared
	}, 3)
})
