package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Creepy Doll — {5} Artifact Creature — Construct 1/1 (Innistrad, S82 OLD b448,
// #5572):
//
//	"Indestructible
//	 Whenever this creature deals combat damage to a creature, flip a coin. If
//	 you win the flip, destroy that creature."
//
// Obelisk Spider's trigger names the damaged creature as the object it was
// (CR 400.7). The flip is the keyed coin flip its controller calls; on a win
// the creature is destroyed if it is still that object on the battlefield, and
// one that died to the damage or left gets nothing. The damaged creature is
// destroyed even by a Doll that has since left.
//
// No simplification.
func init() {
	const label = "Creepy Doll — flip a coin; if you win the flip, destroy that creature"
	effects.Register(effects.Spec{
		OracleID:        "f9c8f9e1-a3a8-47f7-8d29-f954d14ad37b",
		Name:            "Creepy Doll",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: effects.ThisDealtCombatDamageToACreature,
			Key:       label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Object = effects.DamagedCreatureRef(g, ev.Target)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				// The damaged creature rides the flip as data: its ID and the
				// epoch it had when the trigger went on the stack (CR 400.7).
				ref := item.Params.Object
				g.FlipCoinForEffect(game.CoinFlipSpec{
					Flipper:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Creepy Doll — call the coin flip",
					ThenKey:  creepyDollFlipped,
					Carry:    []uuid.UUID{ref.ID},
					Nums:     []int{ref.Epoch},
					Asked:    game.NewAskedItem(item),
				})
				return nil
			},
		}},
	})
}

// creepyDollFlipped destroys the damaged creature when the flip was won, and
// only while it is still the object that was damaged (CR 400.7). An on-disk
// key: never renamed, never reused.
var creepyDollFlipped = effects.CoinFlipThen("coin-flip/creepy-doll-destroy", func(ctx *effects.Context, r game.CoinFlipped) error {
	if !effects.WonFirstFlip(r) || len(r.Carry) == 0 || len(r.Nums) == 0 {
		return nil
	}
	id := r.Carry[0]
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != r.Nums[0] {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return effects.DestroyTarget{Target: id}.Apply(ctx)
})
