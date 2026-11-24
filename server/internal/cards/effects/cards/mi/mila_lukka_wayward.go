package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mila, Crafty Companion // Lukka, Wayward Bonder — modal double-faced card
// (Strixhaven: School of Mages, #5502):
//
//	Mila, Crafty Companion — Legendary Creature — Fox {1}{W}{W}, 2/3
//	 "Whenever an opponent attacks one or more planeswalkers you control, put
//	  a loyalty counter on each planeswalker you control.
//	  Whenever a permanent you control becomes the target of a spell or ability
//	  an opponent controls, you may draw a card."
//	Lukka, Wayward Bonder — Legendary Planeswalker — Lukka {4}{R}{R}, loyalty 5
//	 "+1: You may discard a card. If you do, draw a card. If a creature card
//	  was discarded this way, draw two cards instead.
//	  −2: Return target creature card from your graveyard to the battlefield.
//	  It gains haste. Exile it at the beginning of your next upkeep.
//	  −7: You get an emblem with "Whenever a creature you control enters, it
//	  deals damage equal to its power to any target.""
//
// # Two keys, one card
//
// Mila keeps the bare oracle ID and Lukka takes "<oracle_id>#1".
//
// # Mila
//
// The first trigger is one ability for the whole attack declaration (CR 508.3a
// and the "one or more" of CR 603.2c: OncePerBatch): it asks that an opponent
// of its controller attacks, and that what is attacked is a planeswalker its
// controller controls. The counters go on each planeswalker its controller
// controls as it resolves, the attacked ones and any other. The second trigger
// is Surrak, Elusive Hunter's targeting watch widened to any permanent on the
// battlefield and made optional: a spell or ability controlled by an opponent
// of the permanent's controller.
//
// # Lukka
//
// The +1 is Rummage's prompt: a discard of up to one card, then a draw whose
// size is read off the card that was discarded (two for a creature card). The
// −2 is Whip of Erebos' return, with haste that lasts as long as it stays and a
// delayed exile at the beginning of its controller's next upkeep (CR 603.7).
// The −7 emblem fires for each creature entering under its owner's control;
// the creature deals the damage (last-known power if it has left, CR 608.2h) to
// the target chosen as the trigger goes on the stack.
//
// No simplification on either face.
func init() {
	const id = "ee524e82-35c4-4cae-b65e-3e147eac2927"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Mila, Crafty Companion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, oldB51OpponentAttacksYourPlaneswalker,
				"Mila, Crafty Companion — a loyalty counter on each planeswalker you control",
				effects.FrPutLoyaltyCounterOnEachPlaneswalkerYouControl)),
			effects.OptionalDraw(effects.On(game.EventBecomesTarget, oldB51PermanentYouControlTargetedByOpponent,
				"Mila, Crafty Companion — draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Mila, Crafty Companion — draw a card?", 1),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        id + "#1",
		Name:            "Lukka, Wayward Bonder",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &effects.EmblemSpec{
			Label: "Lukka, Wayward Bonder emblem",
			Text:  "Whenever a creature you control enters, it deals damage equal to its power to any target.",
			Triggered: []game.TriggeredAbility{
				effects.Targeting(
					effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
						return ok && c.IsCreature()
					}, "Lukka, Wayward Bonder emblem — it deals damage equal to its power to any target", oldB51EnteredCreatureDealsPower),
					effects.TargetAny()),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: You may discard a card. If you do, draw a card. If a creature card was discarded this way, draw two cards instead.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1, Discards: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        1,
						UpTo:     true,
						Question: "Lukka, Wayward Bonder — you may discard a card",
					}, oldB51LukkaDraw, nil)
				},
			},
			{
				Label:   "−2: Return target creature card from your graveyard to the battlefield. It gains haste. Exile it at the beginning of your next upkeep.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetCreatureInYourGraveyard(),
				Effect:  oldB51LukkaReanimate,
			},
			{
				Label:  "−7: You get an emblem with \"Whenever a creature you control enters, it deals damage equal to its power to any target.\"",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// oldB51OpponentAttacksYourPlaneswalker is the first Mila trigger's condition:
// an attacker declared by an opponent against a planeswalker the source's
// controller controls.
func oldB51OpponentAttacksYourPlaneswalker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack || source == nil || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	if g.PlayerByIDForEffect(ev.Actor) == nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && effects.OnBattlefield(g, ev.Target) && c.IsPlaneswalker() && c.Controller == source.Controller
}

// oldB51PermanentYouControlTargetedByOpponent is the second Mila trigger's
// condition: a permanent on the battlefield under the source's controller's
// control became the target of a spell or ability an opponent controls.
func oldB51PermanentYouControlTargetedByOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || source == nil || ev.CardID == uuid.Nil {
		return false
	}
	you := source.Controller
	if you == uuid.Nil || ev.Actor == uuid.Nil || ev.Actor == you || g.PlayerByIDForEffect(ev.Actor) == nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && effects.OnBattlefield(g, ev.CardID) && c.Controller == you
}

// oldB51LukkaDraw is the draw after Lukka's optional discard: one card, or two
// if a creature card was discarded this way. An on-disk key: never renamed,
// never reused.
var oldB51LukkaDraw = game.RegisterDiscardThen("choose-cards/old-b051-lukka-discard-draw", func(g *game.Game, r game.DiscardsDone) error {
	cards := r.Discarded.Cards()
	if len(cards) == 0 {
		return nil
	}
	n := 1
	for _, id := range cards {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
			n = 2
			break
		}
	}
	return g.DrawNForEffect(r.Player, n)
})

func oldB51LukkaReanimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
		return err
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	if err := (effects.ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: game.IndefiniteDuration(),
		Label:    "Lukka, Wayward Bonder — haste",
	}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              "Lukka, Wayward Bonder — exile the returned creature",
		Cards:              []uuid.UUID{id},
		Body:               effects.ExileListedCardsBody,
	}.Apply(ctx)
}

// oldB51EnteredCreatureDealsPower is the emblem trigger's body: the creature
// that entered (the event names it) deals damage equal to its power to the
// chosen target.
func oldB51EnteredCreatureDealsPower(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if len(item.Targets) == 0 || item.Trigger == nil {
		return nil
	}
	creature, ok := ctx.TriggeringPermanent()
	if !ok {
		return nil
	}
	ref := ctx.Trigger().Object.Ref()
	return effects.DealDamage{SourceObject: &ref, Target: item.Targets[0].ID, Amount: creature.Power}.Apply(ctx)
}
