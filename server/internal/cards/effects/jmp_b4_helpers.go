package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b4_helpers.go — shared pieces for the fourth Jumpstart deck-request
// batch (Selesnya, #4893).

// OpponentsWhoCastASpellThisTurnCantAttack is Angelic Arbiter's first
// static: "Each opponent who cast a spell this turn can't attack with
// creatures." A layer-6 CantAttack on every creature controlled by an
// opponent of the source's controller whose cast tally this turn is not
// zero. The tally is not on the battlefield, so the static declares
// DependsOnSpellsCast (Stoic Sphinx's seam, #1325): the cast that bumps
// the tally drops the layer cache, and the turn change that resets it
// does too. Restrictions are checked at declaration (CR 508.1c), so a
// spell cast after attackers are declared removes nothing from combat.
func OpponentsWhoCastASpellThisTurnCantAttack() game.StaticAbility {
	return game.StaticAbility{
		Label:               "Each opponent who cast a spell this turn can't attack with creatures.",
		Layer:               game.Layer6Ability,
		DependsOnSpellsCast: true,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			if !target.IsCreature() || target.Controller == source.Controller {
				return false
			}
			return g.CastTallyFor(target.Controller).Total > 0
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= game.CantAttack
		},
	}
}

// OpponentsWhoAttackedWithACreatureThisTurnCantCast is Angelic Arbiter's
// second static: "Each opponent who attacked with a creature this turn
// can't cast spells." Read off this turn's attack record
// (Game.DistinctAttackersThisTurn), which holds declared attackers only, so
// a creature put onto the battlefield attacking is not an attack (CR
// 508.4). A static on the permanent, so its presence is the duration: once
// the Arbiter leaves, that opponent may cast spells again.
func OpponentsWhoAttackedWithACreatureThisTurnCantCast(label string) game.CastRestriction {
	return game.CastRestriction{
		Label: label,
		Forbids: func(q game.CastQuery) bool {
			if q.Source.Controller == q.Controller {
				return false
			}
			return q.Game.DistinctAttackersThisTurn(q.Controller) > 0
		},
	}
}

// nontokenAttackersInTheSameBatch is "that many" for "whenever one or more
// nontoken creatures attack": the nontoken creatures declared as attackers
// in the event batch ev belongs to, whoever declared them. Read off the
// log, so one that died or left combat before the trigger resolved still
// counts; a token that has left has ceased to exist (CR 111.7) and is not
// found, so it is never counted.
//
// Caller must hold g.mu.
func nontokenAttackersInTheSameBatch(g *game.Game, ev game.Event) int {
	n := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		e := g.Events[i]
		if e.Batch < ev.Batch {
			break
		}
		if e.Batch == ev.Batch && e.Kind == game.EventAttack && attackerIsNontoken(g, e.CardID) {
			n++
		}
	}
	return n
}

// attackerIsNontoken reports whether the attacking object is a card rather
// than a token.
func attackerIsNontoken(g *game.Game, id uuid.UUID) bool {
	if id == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(id)
	return ok && !IsToken(c)
}

var (
	// nontokenCreaturesAttackCondition is "whenever one or more nontoken
	// creatures attack": an EventAttack for a nontoken creature. With
	// OncePerBatch it fires once per declaration (CR 603.2c).
	nontokenCreaturesAttackCondition = game.DelayedCondition("attack/nontoken-creatures-attack",
		func(ev game.Event, _ *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
			return ev.Kind == game.EventAttack && attackerIsNontoken(g, ev.CardID)
		})

	// basriKetSoldiersBody is Basri Ket's −2 delayed trigger: that many
	// 1/1 white Soldiers, tapped and attacking.
	basriKetSoldiersBody = game.SimpleDelayedBody("basri-ket/soldiers-tapped-and-attacking", basriKetSoldiers)
)

// basriKetSoldiers counts the nontoken attackers of the triggering
// declaration and creates that many tapped and attacking Soldiers. The
// controller chooses what each attacks (CR 508.4); the first triggering
// attacker's target is offered first. The tokens were never declared as
// attackers, so they trigger nothing that watches attacks.
func basriKetSoldiers(g *game.Game, item *game.StackItem) error {
	if item == nil || item.Trigger == nil {
		return nil
	}
	n := nontokenAttackersInTheSameBatch(g, item.Trigger.Event)
	if n <= 0 {
		return nil
	}
	return CreateTokensAttackingYourChoice{
		Template: TokenCard("1/1 white Soldier"),
		N:        n,
		Tapped:   true,
		Prefer:   triggerAttackTarget(item),
		Label:    "Basri Ket",
	}.Apply(NewContext(g, item))
}

// putAPlusOneCounterOnEachCreatureYouControlStep is "put a +1/+1 counter
// on each creature you control" as a Do step, read as the step runs, so a
// creature an earlier step of the same effect made is among them (Basri
// Ket's emblem).
func putAPlusOneCounterOnEachCreatureYouControlStep(ctx *Context) error {
	return b11PutCounterOnEachCreatureYouControl(ctx.Game, ctx.Item)
}

// enchantedCreaturesControllerNow is "its controller" for an Aura's
// "whenever enchanted creature deals combat damage" trigger: the damaging
// creature's controller as the trigger resolves, or as it last existed on
// the battlefield once it has left (CR 608.2h). The event's actor, the
// creature's controller as it dealt the damage, is the fallback.
func enchantedCreaturesControllerNow(g *game.Game, ev game.Event) uuid.UUID {
	if c, ok := g.LookupCardForEffect(ev.Source); ok {
		if z := g.FindCardZoneForEffect(ev.Source); z != nil && z.Kind == game.ZoneBattlefield {
			return c.Controller
		}
	}
	if info, ok := g.LastKnownPermanentForEffect(ev.Source); ok && info.Controller != uuid.Nil {
		return info.Controller
	}
	return ev.Actor
}

// doubleLifeTotal is "double <player>'s life total": the player gains (or,
// below zero, loses) life equal to their life total (CR 701.10d, as Beacon
// of Immortality applies it), through the ordinary life-change path, so a
// "can't gain life" stops it. A total of zero changes nothing.
func doubleLifeTotal(g *game.Game, source, player uuid.UUID) error {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Life == 0 {
		return nil
	}
	return g.ChangePlayerLifeForEffect(source, player, p.Life)
}

// plusOneCounterThenKeywordOnTheTarget is "Put a +1/+1 counter on target
// creature. It gains <keyword> until end of turn." as two Do steps over
// the first target, each skipped when that target is no longer legal (CR
// 608.2b): Basri Ket's +1 (indestructible, up to one target, so no target
// does nothing) and Battlefield Promotion (first strike). The counter
// settles before the grant, so a prompt the counter asks holds the grant.
func plusOneCounterThenKeywordOnTheTarget(keyword, label string) []Applier {
	return plusOneCounterThenKeywordsOnTheTarget(label, keyword)
}

// plusOneCounterThenKeywordsOnTheTarget is plusOneCounterThenKeywordOnTheTarget
// for a grant of several keywords at once: Biosynthic Burst's "It gains
// reach, trample, and indestructible until end of turn."
func plusOneCounterThenKeywordsOnTheTarget(label string, keywords ...string) []Applier {
	return []Applier{
		Instruction(func(ctx *Context) error {
			if id := firstLegalCardTarget(ctx); id != uuid.Nil {
				return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}
			return nil
		}),
		Instruction(func(ctx *Context) error {
			if id := firstLegalCardTarget(ctx); id != uuid.Nil {
				return GrantKeywordUntilEOT{Target: id, Keywords: keywords, Label: label}.Apply(ctx)
			}
			return nil
		}),
	}
}
