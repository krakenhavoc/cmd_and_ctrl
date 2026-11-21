package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cass, Hand of Vengeance — Legendary Creature — Human Ranger {2}{R}{W},
// 4/3 (Fallout, #5948):
//
//	"Vigilance
//	 Whenever Cass or another creature you control dies, if it was enchanted
//	 or equipped, return any number of Aura cards that were attached to it from
//	 your graveyard to the battlefield attached to target creature, then attach
//	 any number of Equipment that were attached to it to that creature."
//
// One trigger over any creature its controller controlled as it died (Cass
// included, a leaves-the-battlefield look-back), with an intervening "if"
// (CR 603.4) read from the departure record of what was attached as it died.
// The target creature is chosen as the trigger goes on the stack. As it
// resolves the controller picks any number of the Aura cards that were attached
// and are in their graveyard now (a choice, not a target), each coming back
// attached to the target if it could enchant it (CR 303.4j) in one entry; then
// any number of the Equipment that was attached and is still on the battlefield
// is attached to it. A target that has left makes no attachment, and an Aura
// that left the graveyard since is skipped.
//
// No simplification.
func init() {
	cond := effects.CondBoard{
		Name: "it was enchanted or equipped",
		Test: func(ctx *effects.Context) bool {
			if ctx.Item == nil || ctx.Item.Trigger == nil {
				return false
			}
			id := ctx.Item.Trigger.Event.CardID
			return len(effects.GapB64AttachmentsOnTheDead(ctx.Game, id, true))+len(effects.GapB64AttachmentsOnTheDead(ctx.Game, id, false)) > 0
		},
	}
	trigger := effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		dead, ok := effects.DiedCreature(ev, g)
		if !ok || effects.LeftUnderControlOf(ev, dead) != source.Controller {
			return false
		}
		return len(effects.GapB64AttachmentsOnTheDead(g, ev.CardID, true))+len(effects.GapB64AttachmentsOnTheDead(g, ev.CardID, false)) > 0
	}, "Cass, Hand of Vengeance — return the Auras that were attached to it to the battlefield attached to target creature, then attach the Equipment", cassReturn)
	effects.Register(effects.Spec{
		OracleID:        "f9b6b1db-37c6-43d4-b762-6c9770da5964",
		Name:            "Cass, Hand of Vengeance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.InterveningIf(cond, trigger), effects.TargetCreature("target creature")),
		},
	})
}

// cassReturn offers the Aura cards that were attached to the dead creature and
// are in their controller's graveyard now.
func cassReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if item.Trigger == nil {
		return nil
	}
	dead := item.Trigger.Event.CardID
	var candidates []uuid.UUID
	for _, id := range effects.GapB64AttachmentsOnTheDead(g, dead, true) {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneGraveyard {
			if c, ok := g.LookupCardForEffect(id); ok && c.Owner == item.Controller {
				candidates = append(candidates, id)
			}
		}
	}
	if len(candidates) == 0 {
		return cassOfferEquipment(g, item, dead)
	}
	_ = ctx
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Cass, Hand of Vengeance — return any number of Aura cards that were attached to it to the battlefield attached to the target creature",
		Cards:      candidates,
		Min:        0,
		Max:        len(candidates),
		ThenKey:    cassAurasChosen,
		LiveZone:   game.ZoneGraveyard,
		Carry:      []uuid.UUID{dead},
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// cassAurasChosen returns the picked Auras attached to the target, then offers
// the Equipment. An on-disk key: never renamed, never reused.
var cassAurasChosen = effects.AskedPick("choose-cards/cass-hand-of-vengeance-auras", func(ctx *effects.Context, r game.CardsChosen) error {
	g := ctx.Game
	host, ok := effects.B16FirstLegalTargetCard(ctx)
	if ok && effects.OnBattlefield(g, host) && len(r.Picked) > 0 {
		var ids []uuid.UUID
		for _, id := range r.Picked {
			if card, found := g.LookupCardForEffect(id); found && effects.CanBeAttachedTo(g, ctx.Controller(), card, host) {
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			if err := (effects.ReturnFromGraveyardTogether{
				Targets:    ids,
				AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
			}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	if len(r.Carry) == 0 {
		return nil
	}
	return cassOfferEquipment(g, ctx.Item, r.Carry[0])
})

// cassOfferEquipment asks which of the Equipment that was attached to the dead
// creature, and is still on the battlefield, to attach to the target creature.
func cassOfferEquipment(g *game.Game, item *game.StackItem, dead uuid.UUID) error {
	ctx := effects.NewContext(g, item)
	host, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok || !effects.OnBattlefield(g, host) {
		return nil
	}
	var candidates []uuid.UUID
	for _, id := range effects.GapB64AttachmentsOnTheDead(g, dead, false) {
		if effects.OnBattlefield(g, id) {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Cass, Hand of Vengeance — attach any number of the Equipment that was attached to it to the target creature",
		Cards:      candidates,
		Min:        0,
		Max:        len(candidates),
		ThenKey:    cassEquipmentChosen,
		LiveZone:   game.ZoneBattlefield,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// cassEquipmentChosen attaches each picked Equipment to the target creature. An
// on-disk key: never renamed, never reused.
var cassEquipmentChosen = effects.AskedPick("choose-cards/cass-hand-of-vengeance-equipment", func(ctx *effects.Context, r game.CardsChosen) error {
	host, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok || !effects.OnBattlefield(ctx.Game, host) {
		return nil
	}
	for _, id := range r.Picked {
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := ctx.Game.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			return err
		}
	}
	return nil
})
