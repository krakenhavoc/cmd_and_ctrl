package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Living Lore — Creature — Avatar {3}{U}, */* (Commander 2021, #5504):
//
//	"As this creature enters, exile an instant or sorcery card from your
//	 graveyard.
//	 Living Lore's power and toughness are each equal to the exiled card's
//	 mana value.
//	 Whenever this creature deals combat damage, you may sacrifice it. If
//	 you do, you may cast the exiled card without paying its mana cost."
//
// The exile is Mimeoplasm's entry slot (Spec.AsEnters, CR 614.12): a
// mandatory pick among the instant and sorcery cards in your graveyard,
// exiled LINKED to this object (ExileLinkedForEffect, CR 607.2a). With none
// there is nothing to exile, the creature is 0/0 and dies. The power and
// toughness are a characteristic-defining ability (CR 604.3, layer 7a)
// reading the linked card's mana value (an X is 0, CR 107.3m).
//
// The trigger fires on combat damage to any recipient, once per damage
// event. The "may sacrifice" is asked as it resolves; the exiled card is
// named BEFORE the sacrifice, since the creature that leaves is a new
// object with no claim on it (CR 400.7), and "if you do" casts it as part
// of the resolution (CR 608.2g) for free, leaving it in exile when the
// cast is declined.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c48c341-1d56-4210-9d86-181ebb206098",
		Name:         "Living Lore",
		Completeness: effects.CompletenessFull,
		AsEnters:     livingLoreExileAsEnters,
		Static: []game.StaticAbility{{
			Label:          "Power and toughness are each equal to the exiled card's mana value.",
			Layer:          game.Layer7PT,
			SubLayer:       game.SubLayer7A_CDA,
			DependsOnExile: true,
			AppliesTo:      effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				mv := 0
				if linked := effects.Cmdr2S36CardsExiledWithThis(g, source); len(linked) > 0 {
					mv = linked[len(linked)-1].ManaValue()
				}
				c.Power = mv
				c.Toughness = mv
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.S82Vow03ThisDealtCombatDamage,
				"Living Lore — you may sacrifice it, then you may cast the exiled card without paying its mana cost",
				effects.Do(effects.MayChoice{
					Question: "Living Lore — sacrifice it to cast the exiled card without paying its mana cost?",
					YesLabel: "Sacrifice",
					NoLabel:  "Decline",
					Then:     livingLoreAnswer,
				})),
		},
	})
}

// livingLoreExileAsEnters queues "exile an instant or sorcery card from
// your graveyard": nothing to ask when the graveyard holds none.
func livingLoreExileAsEnters(card *game.Card, ctx *effects.Context) error {
	p := ctx.Game.PlayerByIDForEffect(card.Controller)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if (c.IsInstant() || c.IsSorcery()) && !c.IsToken() {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  card.Controller,
		Source:   card.InstanceID,
		Question: "Living Lore — exile an instant or sorcery card from your graveyard",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		ThenKey:  livingLoreExiled,
		LiveZone: game.ZoneGraveyard,
	})
	return nil
}

// livingLoreExiled is the keyed answer: the card named is exiled linked to
// the creature. An on-disk key: never renamed, never reused.
var livingLoreExiled = game.RegisterChooseCardsThen("choose-cards/living-lore-exile", func(g *game.Game, r game.CardsChosen) error {
	ref, ok := g.PermanentRefForEffect(r.Source)
	if !ok {
		return nil
	}
	for _, id := range r.Picked {
		if err := g.ExileLinkedForEffect(ref, id); err != nil {
			return err
		}
	}
	return nil
})

// livingLoreAnswer is the "yes": name the exiled card, sacrifice the
// creature, and if it went, offer the free cast. Keyed.
var livingLoreAnswer = effects.MayYes("confirm/living-lore-sacrifice", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	linked := effects.Cmdr2S36CardsExiledWithThis(ctx.Game, &src)
	exiled := uuid.Nil
	if len(linked) > 0 {
		exiled = linked[len(linked)-1].InstanceID
	}
	return effects.SacrificePermanent{Target: ctx.Source(), ThenTail: effects.WithCarry(ctx.ItemTail(livingLoreCast), exiled)}.Apply(ctx)
})

// livingLoreCast is "you may cast the exiled card without paying its mana
// cost", run once the creature is sacrificed.
var livingLoreCast = effects.CardTail("then/living-lore-cast", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	id := effects.TailCarry0(o)
	if id == uuid.Nil {
		return nil
	}
	return ctx.Game.CastExiledFreeKeepingForEffect(ctx.Controller(), ctx.Source(), id)
})
