package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b14_helpers.go — shared pieces for Secrets of Strixhaven Commander,
// batch 14 (#4834).

// socB14AuraSpell passes on an Aura spell (Transcendent Envoy's discount).
func socB14AuraSpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.HasSubtype("Aura") }
}

// socB14AttackersAtYouOrYourPlaneswalkers counts the creatures `attacker`
// controls that are attacking `defender` or a planeswalker `defender`
// controls (CR 508.1b). A creature attacking a battle is not counted, even
// when `defender` protects it: Tomik's text names players and planeswalkers
// only.
func socB14AttackersAtYouOrYourPlaneswalkers(g *game.Game, attacker, defender uuid.UUID) int {
	n := 0
	for _, id := range b13AttackingCreaturesYouControl(g, attacker) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		switch g.ClassifyAttackTargetForEffect(c.AttackingTarget) {
		case game.AttackTargetPlayer, game.AttackTargetPlaneswalker:
			if g.DefendingPlayerForAttackForEffect(c.AttackingTarget) == defender {
				n++
			}
		}
	}
	return n
}

// socB14OpponentAttackedYouTwice is the trigger-time half of Tomik's "whenever
// an opponent attacks with creatures, if two or more of those creatures are
// attacking you and/or planeswalkers you control": an attack declaration by an
// opponent of the source's controller.
func socB14OpponentAttackedYouTwice(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	return socB14AttackersAtYouOrYourPlaneswalkers(g, ev.Actor, source.Controller) >= 2
}

// socB14StillAttackedTwice is the CR 603.4 recheck as the trigger resolves:
// the opponent who attacked still has two or more attackers at you and/or your
// planeswalkers.
var socB14StillAttackedTwice = CondBoard{
	Name: "two or more attackers still attack you and/or your planeswalkers",
	Test: func(ctx *Context) bool {
		who := ctx.Trigger().Event.Actor
		return who != uuid.Nil && socB14AttackersAtYouOrYourPlaneswalkers(ctx.Game, who, ctx.Controller()) >= 2
	},
}

// socB14ThatOpponentLoses3 is "that opponent loses 3 life": the player who
// attacked, read off the triggering event. A player no longer seated loses
// nothing.
func socB14ThatOpponentLoses3(ctx *Context) error {
	who := ctx.Trigger().Event.Actor
	if p := ctx.Game.PlayerByIDForEffect(who); p == nil || p.Eliminated {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), who, -3)
}

// socB14HastyCopiesExiledAtEnd is "for each of them, create a token that's a
// copy of that creature, except it has haste. Exile those tokens at the
// beginning of the next end step" over `originals`, in order (Twinflame, and
// Kindred Charge's body). The exile is a CR 603.7 delayed trigger over exactly
// the tokens the copies made; no copies, no trigger.
func socB14HastyCopiesExiledAtEnd(ctx *Context, controller uuid.UUID, originals []uuid.UUID, label string) error {
	cursor := b25LastEventSeq(ctx.Game)
	for _, id := range originals {
		if err := (CreateTokenCopy{Controller: controller, Copy: id, N: 1, Except: TokenCopyGainsHaste}).Apply(ctx); err != nil {
			return err
		}
	}
	tokens := b27TokensCreatedByAfter(ctx.Game, controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: label,
		Cards: tokens,
		Body:  exileListedCardsBody,
	}.Apply(ctx)
}

// socB14CopyPerCommanderCastBody is Thunderclap Drake's delayed body: copy the
// spell the event named once for each time the trigger's controller has cast a
// commander from the command zone this game, new targets allowed. With none
// cast it does nothing (CopySpell would read a count of zero as one).
var socB14CopyPerCommanderCastBody = game.SimpleDelayedBody("copy/the-spell-per-commander-cast",
	func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		cast := ctx.PayloadCards()
		n := socB3CommanderCastsFromCommandZone(g, item.Controller)
		if len(cast) == 0 || n <= 0 {
			return nil
		}
		return CopySpell{
			StackID:          cast[0],
			Controller:       item.Controller,
			Count:            n,
			ChooseNewTargets: true,
			FromLastKnown:    true,
		}.Apply(ctx)
	})

// socB14YourCreatureDealtCombatDamageToAPlayerCondition is the event condition
// of "whenever a creature you control deals combat damage to a player this
// turn" (Surge to Victory): a combat damage event, to a player, from a creature
// controlled by the delayed trigger's controller (CR 603.7d) as it dealt it.
var socB14YourCreatureDealtCombatDamageToAPlayerCondition = game.DelayedCondition("damage/your-creature-combat-damage-to-a-player",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
			return false
		}
		src, ok := g.LookupCardForEffect(ev.Source)
		return ok && src.IsCreature() && b17DamageSourceController(ev, g) == dt.Controller
	})

// socB14CopyTheExiledCardBody is "copy the exiled card. You may cast the copy
// without paying its mana cost": the card the trigger carries, if it is still
// in exile (a card that left is a new object, CR 400.7). An on-disk key.
var socB14CopyTheExiledCardBody = game.SimpleDelayedBody("copy/surge-to-victory-exiled-card",
	func(g *game.Game, item *game.StackItem) error {
		var cards []uuid.UUID
		for _, t := range item.Targets {
			if z := g.FindCardZoneForEffect(t.ID); t.Kind == game.TargetCard && z != nil && z.Kind == game.ZoneExile {
				cards = append(cards, t.ID)
			}
		}
		if len(cards) == 0 {
			return nil
		}
		return CopyAndCast{Cards: cards, Free: true}.Apply(NewContext(g, item))
	})

// socB14YouCastAPermanentSpellWithX is the When of Unbound Flourishing's first
// ability: the source's controller cast a permanent spell (anything but an
// instant or a sorcery) with an {X} in its own printed mana cost.
func socB14YouCastAPermanentSpellWithX(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller || !socB9HasXInManaCost(g, ev.CardID) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsInstant() && !c.IsSorcery()
}

// socB14DoubleTheSpellsX is "double the value of X": the spell the cast event
// named, if it is still on the stack.
func socB14DoubleTheSpellsX(ctx *Context) error {
	ctx.Game.DoubleSpellXForEffect(ctx.Trigger().Event.CardID)
	return nil
}

// socB14YouCastOrActivateAnXThing is the When of Unbound Flourishing's second
// ability: the source's controller cast an instant or sorcery spell whose
// printed mana cost contains {X}, or activated an ability with an {X} mana
// symbol in its activation cost.
func socB14YouCastOrActivateAnXThing(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller {
		return false
	}
	switch ev.Kind {
	case game.EventCast:
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && (c.IsInstant() || c.IsSorcery()) && socB9HasXInManaCost(g, ev.CardID)
	case game.EventActivateAbility:
		return ev.XCost && ev.StackItemID != uuid.Nil
	}
	return false
}

// socB14TheCostContainsX is the CR 603.4 recheck of the same "if" as the
// trigger resolves. What it asks about is fixed when the spell is cast or the
// ability activated, so it reads the triggering event, not the stack.
var socB14TheCostContainsX = CondBoard{
	Name: "that spell's mana cost or that ability's activation cost contains {X}",
	Test: func(ctx *Context) bool {
		ev := ctx.Trigger().Event
		switch ev.Kind {
		case game.EventCast:
			return socB9HasXInManaCost(ctx.Game, ev.CardID)
		case game.EventActivateAbility:
			return ev.XCost
		}
		return false
	},
}

// socB14CopyThatSpellOrAbility is "copy that spell or ability. You may choose
// new targets for the copy": the spell or the ability the event named, under
// its controller's control as the trigger's controller (CR 707.10), from last
// known information if the original has left the stack.
func socB14CopyThatSpellOrAbility(ctx *Context) error {
	ev := ctx.Trigger().Event
	switch ev.Kind {
	case game.EventCast:
		return CopySpell{
			StackID:          ev.CardID,
			Controller:       ctx.Controller(),
			ChooseNewTargets: true,
			FromLastKnown:    true,
		}.Apply(ctx)
	case game.EventActivateAbility:
		return CopyAbility{
			ItemID:           ev.StackItemID,
			Controller:       ctx.Controller(),
			ChooseNewTargets: true,
		}.Apply(ctx)
	}
	return nil
}

// socB14OtherCreaturesBecomeTanazirsSize is Tanazir Quandrix's attack body:
// the creatures its controller has other than the source get base power and
// toughness equal to the source's power and toughness until end of turn
// (layer 7b). The source is read as the ability resolves, counters and pumps
// included, as it last existed there if it has gone (CR 608.2h).
func socB14OtherCreaturesBecomeTanazirsSize(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	return untilEndOfTurn(ctx, uuid.Nil, And(Creature(), YouControl(), OtherThan(ctx.Source())),
		"Tanazir Quandrix — base power and toughness equal to his",
		game.SetBasePowerMod(info.Power), game.SetBaseToughnessMod(info.Toughness))
}
