package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jaya, Fiery Negotiator — Legendary Planeswalker — Jaya {2}{R}{R}, starting
// loyalty 4 (Dominaria United, #5487):
//
//	"+1: Create a 1/1 red Monk creature token with prowess.
//	 −1: Exile the top two cards of your library. Choose one of them. You may
//	 play that card this turn.
//	 −2: Choose target creature an opponent controls. Whenever you attack this
//	 turn, Jaya deals damage equal to the number of attacking creatures to that
//	 creature.
//	 −8: You get an emblem with "Whenever you cast a red instant or sorcery
//	 spell, copy it twice. You may choose new targets for the copies.""
//
// The −1 is Chandra, Flameshaper's exile-and-choose: both cards go to exile
// face up, the chosen one alone may be played (a land uses the turn's land
// drop) until end of turn. The −2 targets as it is activated and creates a
// delayed trigger that fires on every attack declaration of its controller
// this turn (CR 603.7b); the damage is read as the delayed trigger resolves,
// the number of creatures attacking then, and a target that has left the
// battlefield is simply not dealt damage. Jaya is the damage's source, by
// last-known information if she has left (CR 608.2h). The emblem copies a red
// instant or sorcery spell its owner casts, twice, as it is cast (CR 707.10),
// each copy with the option of new targets.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2458aa66-5b20-4811-a4a4-8375ad0a6498",
		Name:            "Jaya, Fiery Negotiator",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &effects.EmblemSpec{
			Label: "Jaya, Fiery Negotiator emblem",
			Text:  "Whenever you cast a red instant or sorcery spell, copy it twice. You may choose new targets for the copies.",
			Triggered: []game.TriggeredAbility{
				effects.WheneverYouCast(effects.And(effects.OfColor("R"), effects.Or(effects.Instant(), effects.Sorcery())),
					"Jaya, Fiery Negotiator emblem — copy that spell twice",
					jayaFieryNegotiatorCopyTwice),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Create a 1/1 red Monk creature token with prowess.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Monk with prowess"), N: 1}),
			},
			{
				Label: "−1: Exile the top two cards of your library. Choose one of them. You may play that card this turn.",
				Cost:  effects.LoyaltyCost(-1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.ExileTopKeyedForEffect(item.Controller, item.Controller, 2, ctx.ItemTail(jayaFieryNegotiatorExiled))
				},
			},
			{
				Label:   "−2: Choose target creature an opponent controls. Whenever you attack this turn, Jaya deals damage equal to the number of attacking creatures to that creature.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return effects.DelayedOnEvent{
						Label:        "Jaya, Fiery Negotiator — damage to the chosen creature equal to the number of attacking creatures",
						On:           []game.EventKind{game.EventAttack},
						Condition:    effects.YouAttackCondition,
						Cards:        []uuid.UUID{t.ID},
						Body:         jayaFieryNegotiatorDamageBody,
						Repeats:      true,
						OncePerBatch: true,
					}.Apply(ctx)
				},
			},
			{
				Label:  "−8: You get an emblem with \"Whenever you cast a red instant or sorcery spell, copy it twice. You may choose new targets for the copies.\"",
				Cost:   effects.LoyaltyCost(-8),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

var jayaFieryNegotiatorDamageBody = game.SimpleDelayedBody("jaya-fiery-negotiator/damage", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	n := len(effects.AttackingCreatures(g))
	if n == 0 {
		return nil
	}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || !effects.OnBattlefield(g, t.ID) {
			continue
		}
		d := effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}
		if ref, ok := ctx.SourceRef(); ok {
			d.SourceObject = &ref
		}
		if err := d.Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

func jayaFieryNegotiatorCopyTwice(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return effects.CopySpell{
		StackID:          item.Trigger.Event.CardID,
		Controller:       item.Controller,
		Count:            2,
		ChooseNewTargets: true,
		FromLastKnown:    true,
	}.Apply(effects.NewContext(g, item))
}

// jayaFieryNegotiatorExiled is the exile's keyed rest: "Choose one of them"
// among the cards that landed in exile. An on-disk key: never renamed, never reused.
var jayaFieryNegotiatorExiled = effects.CardTail("jaya-fiery-negotiator/choose-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Jaya, Fiery Negotiator — choose one of the exiled cards; you may play it this turn",
		Cards:    o.IDs,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  jayaFieryNegotiatorChosen,
	})
	return nil
})

var jayaFieryNegotiatorChosen = game.RegisterChooseCardsThen("choose-cards/jaya-fiery-negotiator-play", func(g *game.Game, r game.CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(r.Picked[0])
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(c.InstanceID); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     r.Chooser,
		Zone:       game.ZoneExile,
		Source:     r.Source,
		SourceName: "Jaya, Fiery Negotiator",
		Label:      "Play the card Jaya, Fiery Negotiator let you choose",
	}, []game.Card{c})
	return nil
})
