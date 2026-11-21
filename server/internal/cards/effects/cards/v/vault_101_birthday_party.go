package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vault 101: Birthday Party — Enchantment — Saga {3}{W} (Fallout):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Create a 1/1 white Human Soldier creature token and a Food token.
//	 II, III — You may put an Aura or Equipment card from your hand or
//	 graveyard onto the battlefield. If an Equipment is put onto the
//	 battlefield this way, you may attach it to a creature you control."
//
// I makes the two tokens. II and III are Kastral's hand-or-graveyard pick (a
// keyed choose_cards prompt, none a legal answer) with Armored Skyhunter's
// entry: an Aura put onto the battlefield chooses what it enchants as it
// enters (CR 303.4f) and stays where it was if nothing can be enchanted
// (CR 303.4g), a card that has left its zone since the prompt is skipped,
// and a token is never a card. After an Equipment has entered, the controller
// is offered the creatures they control (none is a legal answer) and it is
// attached to the one picked, both re-checked (CR 701.3). An Aura gets no
// such offer, as printed.
//
// No simplification.
func init() {
	const name = "Vault 101: Birthday Party"
	offer := func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		p := ctx.PlayerByID(ctx.Controller())
		if p == nil {
			return nil
		}
		var cards []uuid.UUID
		for _, pile := range []*game.Zone{p.Hand, p.Graveyard} {
			if pile == nil {
				continue
			}
			for _, c := range pile.Cards {
				if c.HasSubtype("Aura") || c.HasSubtype("Equipment") {
					cards = append(cards, c.InstanceID)
				}
			}
		}
		if len(cards) == 0 {
			return nil
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: name + " — you may put an Aura or Equipment card from your hand or graveyard onto the battlefield",
			Cards:    cards,
			Min:      0,
			Max:      1,
			ThenKey:  vault101Put,
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	effects.Register(effects.Spec{
		OracleID:     "77322d69-330e-4b00-b984-62c6d45c9540",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create a 1/1 white Human Soldier creature token and a Food token"),
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1}, effects.CreateToken{Template: effects.FoodToken(), N: 1})),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "you may put an Aura or Equipment card from your hand or graveyard onto the battlefield; you may attach an Equipment to a creature you control"), offer),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "you may put an Aura or Equipment card from your hand or graveyard onto the battlefield; you may attach an Equipment to a creature you control"), offer),
		},
	})
}

// vault101Put puts the chosen card onto the battlefield from the zone it is in
// and then offers the attach. An on-disk key: never renamed, never reused.
var vault101Put = effects.AskedPick("choose-cards/vault-101-put", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	z := ctx.Game.FindCardZoneForEffect(r.Picked[0])
	if z == nil || (z.Kind != game.ZoneHand && z.Kind != game.ZoneGraveyard) {
		return nil
	}
	return ctx.Game.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: r.Picked[0], From: z.Kind}},
		game.ZoneEntryOptions{Controller: ctx.Controller()},
		ctx.Tail(vault101Entered))
})

// vault101Entered offers an Equipment that entered to a creature the
// controller controls. An on-disk key.
var vault101Entered = effects.TogetherTail("vault-101/entered", func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
	for _, id := range entered {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !effects.OnBattlefield(ctx.Game, id) || !c.HasSubtype("Equipment") {
			continue
		}
		var creatures []uuid.UUID
		for _, b := range ctx.Game.BattlefieldCardsForEffect() {
			if b.Controller == ctx.Controller() && b.IsCreature() {
				creatures = append(creatures, b.InstanceID)
			}
		}
		if len(creatures) == 0 {
			continue
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: "Vault 101: Birthday Party — you may attach " + c.Name + " to a creature you control",
			Cards:    creatures,
			Min:      0,
			Max:      1,
			ThenKey:  vault101Attach,
			Carry:    []uuid.UUID{id},
			Asked:    game.NewAskedItem(ctx.Item),
		})
	}
	return nil
})

// vault101Attach attaches the Equipment to the creature picked, if both are
// still on the battlefield and the attach is legal. An on-disk key.
var vault101Attach = effects.AskedPick("choose-cards/vault-101-attach", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	if !effects.OnBattlefield(ctx.Game, r.Carry[0]) || !effects.OnBattlefield(ctx.Game, r.Picked[0]) {
		return nil
	}
	return ctx.Game.AttachForEffect(r.Carry[0], game.TargetRef{Kind: game.TargetCard, ID: r.Picked[0]})
})
