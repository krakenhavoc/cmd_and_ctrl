package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lumbering Battlement — Creature — Beast {4}{W}, 4/5 (Ravnica Allegiance):
//
//	"Vigilance
//	 When this creature enters, exile any number of other nontoken creatures you
//	 control until it leaves the battlefield.
//	 This creature gets +2/+2 for each card exiled with it."
//
// Abdel Adrian's choose-then-exile-until-this-leaves chain, each exile linked to
// the Battlement (CR 607.2a) so the static counts the cards still exiled with
// this object; once it leaves they return and the link is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dcf7d099-a726-4388-9744-70d37c5ea860",
		Name:            "Lumbering Battlement",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{{
			Label:     "Gets +2/+2 for each card exiled with it.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := len(effects.Cmdr2S36CardsExiledWithThis(g, source))
				c.Power += 2 * n
				c.Toughness += 2 * n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Lumbering Battlement — exile any number of other nontoken creatures you control until it leaves the battlefield", lumberingBattlementEnters),
		},
	})
}

func lumberingBattlementEnters(g *game.Game, item *game.StackItem) error {
	var mine []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && !effects.IsToken(c) && c.InstanceID != item.SourceCardID {
			mine = append(mine, c.InstanceID)
		}
	}
	if len(mine) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Lumbering Battlement — exile any number of other nontoken creatures you control until it leaves the battlefield",
		Cards:    mine,
		Min:      0,
		Max:      0, // "any number"
		ThenKey:  lumberingBattlementPick,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// lumberingBattlementPick is the chosen set. An on-disk key: never renamed.
var lumberingBattlementPick = effects.AskedPick("choose-cards/lumbering-battlement-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	return lumberingBattlementExileNext(ctx, r.Picked)
})

func lumberingBattlementExileNext(ctx *effects.Context, picked []uuid.UUID) error {
	if len(picked) == 0 {
		return nil
	}
	id, rest := picked[0], picked[1:]
	t := ctx.ItemTail(lumberingBattlementExiledOne)
	t.Carry = append([]uuid.UUID(nil), rest...)
	return effects.ExileUntil{
		Target:     id,
		ThisLeaves: true,
		Linked:     true,
		Label:      "Lumbering Battlement — returns when Lumbering Battlement leaves the battlefield",
		ThenTail:   t,
	}.Apply(ctx)
}

var lumberingBattlementExiledOne game.TailRef

func init() {
	lumberingBattlementExiledOne = effects.CardTail("then/lumbering-battlement-exiled-one", func(ctx *effects.Context, o game.TailOutcome) error {
		return lumberingBattlementExileNext(ctx, o.Tail.Carry)
	})
}
