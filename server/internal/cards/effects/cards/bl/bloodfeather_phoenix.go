package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodfeather Phoenix — Creature — Phoenix {1}{R}, 2/2 (March of the
// Machine):
//
//	"Flying
//	 This creature can't block.
//	 Whenever an instant or sorcery spell you control deals damage to an
//	 opponent or battle, you may pay {R}. If you do, return this card
//	 from your graveyard to the battlefield. It gains haste until end of
//	 turn."
//
// A graveyard trigger (Flamewake Phoenix's) on damage from an instant or
// sorcery spell its controller controls, read while that spell is
// mid-resolution (Imodane's shape). The payment is offered as the trigger
// resolves, for the same graveyard object only (CR 400.7), and haste goes
// to the returned permanent only if the return happened.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f4490b09-dc55-47ed-a573-4ae0b95725b1",
		Name:            "Bloodfeather Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{bloodfeatherPhoenixCantBlock()},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventDealDamage, bloodfeatherPhoenixSpellDamage,
				"Bloodfeather Phoenix — you may pay {R} to return it from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPayForGraveyardReturnThen(g, item, "{R}",
						"Pay {R} to return Bloodfeather Phoenix from your graveyard to the battlefield?",
						bloodfeatherPhoenixPaid)
				})),
		},
	})
}

func bloodfeatherPhoenixCantBlock() game.StaticAbility {
	return game.StaticAbility{
		Label: "This creature can't block.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return effects.SelfOnly(target, g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= game.CantBlock
		},
	}
}

// bloodfeatherPhoenixSpellDamage is the trigger event: damage dealt by an
// instant or sorcery spell the Phoenix's controller controls to an
// opponent of theirs or to a battle.
func bloodfeatherPhoenixSpellDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source == uuid.Nil || ev.Target == uuid.Nil {
		return false
	}
	spell, item, ok := g.ResolvingSpellForEffect(ev.Source)
	if !ok || item.Kind != game.StackItemSpell || item.Controller != source.Controller {
		return false
	}
	if !(spell.IsInstant() || spell.IsSorcery()) {
		return false
	}
	if ev.Target != source.Controller {
		if p := g.PlayerByIDForEffect(ev.Target); p != nil {
			return true
		}
	}
	victim, ok := g.LookupCardForEffect(ev.Target)
	return ok && victim.IsBattle()
}

// bloodfeatherPhoenixPaid returns the card, then gives it haste until end
// of turn if it did come back.
var bloodfeatherPhoenixPaid = effects.PayUnlessBranches("pay-unless/bloodfeather-phoenix-return",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if !effects.CardObjectStillInGraveyard(ctx.Game, ctx.Item) {
			return nil
		}
		if err := effects.ReturnThisCardFromYourGraveyard(ctx.Game, ctx.Item); err != nil {
			return err
		}
		id := ctx.Item.SourceCardID
		if !effects.OnBattlefield(ctx.Game, id) {
			return nil
		}
		return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Bloodfeather Phoenix — haste"}.Apply(ctx)
	}, nil)
