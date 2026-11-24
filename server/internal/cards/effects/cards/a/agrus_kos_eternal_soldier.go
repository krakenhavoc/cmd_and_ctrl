package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Agrus Kos, Eternal Soldier — Legendary Creature — Spirit Soldier {3}{W},
// 3/4 (Jumpstart 2022, S82 GAP2 b05 #5948, #6094):
//
//	"Vigilance
//	 Whenever Agrus Kos becomes the target of an ability that targets only
//	 it, you may pay {1}{R/W}. If you do, copy that ability for each other
//	 creature you control that ability could target. Each copy targets a
//	 different one of those creatures."
//
// The trigger watches EventBecomesTarget for Agrus Kos itself, from an
// activated or triggered ability (not a spell), whose targets are Agrus Kos
// and nothing else (CR 115.9c, StackItemSoleTargetForEffect). On resolution
// its controller may pay {1}{R/W} (MayPay, a registered branch). The copies
// are one each for every other creature its controller controls that the
// ability could target instead of Agrus Kos (CR 707.10; the engine's
// CopyAbilityRetargetedEachForEffect judges each new target with the
// announce gate, so hexproof and protection are honoured). The ability must
// still be on the stack: one countered in response is not copied.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3d19823a-dbb8-4384-81f4-5c6818b7f46d",
		Name:            "Agrus Kos, Eternal Soldier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventBecomesTarget},
			AppliesTo: agrusKosTargetedByAnAbilityAlone,
			Key:       "Agrus Kos, Eternal Soldier — you may pay {1}{R/W} to copy that ability for each other creature you control",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if ctx.Trigger().Event.StackItemID == uuid.Nil {
					return nil
				}
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}{R/W}",
					Question: "Agrus Kos — pay {1}{R/W} to copy that ability for each other creature you control it could target?",
					Then:     agrusKosPaid,
				}.Apply(ctx)
			},
		}},
	})
}

func agrusKosTargetedByAnAbilityAlone(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if g == nil || ev.CardID != source.InstanceID || ev.StackItemID == uuid.Nil {
		return false
	}
	if ev.TargeterKind != game.StackItemActivated && ev.TargeterKind != game.StackItemTriggered {
		return false
	}
	sole, ok := g.StackItemSoleTargetForEffect(ev.StackItemID)
	return ok && sole.Kind == game.TargetCard && sole.ID == source.InstanceID
}

// agrusKosPaid makes the copies once {1}{R/W} is paid. An on-disk key.
var agrusKosPaid = effects.PayUnlessBranches("pay-unless/agrus-kos-copy-ability",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		ability := ctx.Trigger().Event.StackItemID
		agrus := ctx.Source()
		you := ctx.Controller()
		var tos []uuid.UUID
		for _, c := range ctx.Game.Battlefield.Cards {
			if c.InstanceID == agrus || c.Controller != you || !c.IsCreature() {
				continue
			}
			tos = append(tos, c.InstanceID)
		}
		ctx.Game.CopyAbilityRetargetedEachForEffect(ability, you, agrus, tos)
		return nil
	}, nil)
