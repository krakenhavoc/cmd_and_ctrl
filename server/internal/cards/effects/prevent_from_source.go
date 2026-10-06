package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prevent_from_source.go — the card-facing half of ADR 0108 §7 (#1904):
// "Prevent all damage a source of your choice would deal [to you] this
// turn" and "Prevent the next N damage that would be dealt to <target>
// this turn by a source of your choice". The engine half is
// game/prevent_from_source.go (the shield) and game/divide_shield.go
// (CR 615.7's division of a charge among simultaneous damage).
//
// Append-only, mechanic-named: every member of the family is one call.
//
// Writing a card:
//
//	PreventDamageFromChosenSource(ShieldAnything)
//	  — Pay No Heed's "a source of your choice would deal this turn"
//	PreventDamageFromChosenSource(ShieldYou, QueryColors("R"))
//	  — "a red source of your choice would deal to you this turn"
//	PreventDamageFromChosenSource(ShieldTheTarget).Charged(3)
//	  — Healing Grace's "the next 3 damage … to any target … by a
//	    source of your choice"
//	PreventDamageFromSource{From: id, CombatOnly: true, Protect: ShieldAnything}
//	  — "all combat damage that would be dealt by target creature"
//
// and set Then to a follow-up body for "prevented this way" (CR 615.5).

// PreventDamageFromSource is the shield. Choose, FromThis and From name
// the source as on PreventNextDamageFromSource; with none of them any
// source matching Queries is the one, and with no Queries either, every
// source ("prevent all damage that would be dealt to X this turn").
type PreventDamageFromSource struct {
	Choose   bool
	FromThis bool
	From     uuid.UUID

	// Queries is the property the source must have, offered by the
	// prompt and rechecked as the damage would be dealt (CR 615.9).
	Queries []game.PermanentQuery

	// Filter is the rest of an unnamed source's description (#2026):
	// "non-Spider", "with power 3 or less", "attacking", "without
	// flying", "colorless", "with no +1/+1 counters", "your opponents
	// control" — read as the damage would be dealt, every time
	// (game.DamageSourceFilter). Only with no Choose, FromThis or From.
	Filter game.DamageSourceFilter
	// exceptThis and exceptClause add objects pinned as the shield
	// resolves to Filter.ExceptObjects: this object (Haze Frog's "other
	// creatures") or the target answering clause exceptClause
	// (Terrifying Presence's "creatures other than target creature").
	// Set with OtherThanThis and OtherThanTarget.
	exceptThis      bool
	exceptClause    int
	hasExceptClause bool

	// Protect is what the shield protects.
	Protect ShieldTarget

	// CombatOnly is "all combat damage".
	CombatOnly bool

	// AndDealtBy is "dealt to and dealt by" (Maze of Ith): the protected
	// permanents are also the sources the shield stops, in one record.
	// It names no other source (ADR 0108 Delivery PR 7).
	AndDealtBy bool

	// Amount is "the next N damage" (CR 615.7); zero is all damage this
	// turn.
	Amount int

	// Then is the CR 615.5 follow-up. Zero is none.
	Then game.BodyRef
	// ThenPerSource runs Then once per damage source within one instance
	// of damage rather than once for the instance (#2026): Comeuppance's
	// "If damage from a creature source is prevented this way,
	// Comeuppance deals that much damage to that creature".
	ThenPerSource bool
	// thenTo is the target clause whose answer the follow-up deals its
	// damage to (Refraction Trap's "any target", ADR 0108 §9), when
	// hasThenTo is set. Set with DealingTo.
	thenTo    int
	hasThenTo bool

	// UntilYourNextTurn is "until your next turn" (Gideon of the
	// Trials); false is "this turn".
	UntilYourNextTurn bool

	// Question is the source prompt's header; Label the shield's. Both
	// default to the card's name.
	Question string
	Label    string
}

// PreventDamageFromChosenSource is "Prevent all damage a [<queries>]
// source of your choice would deal [to <protect>] this turn" — Pay No
// Heed, Auriok Replica, Burrenton Forge-Tender.
func PreventDamageFromChosenSource(protect ShieldTarget, queries ...game.PermanentQuery) PreventDamageFromSource {
	return PreventDamageFromSource{Choose: true, Protect: protect, Queries: queries}
}

// Charged returns the shield as "the next n damage" (CR 615.7).
func (p PreventDamageFromSource) Charged(n int) PreventDamageFromSource {
	p.Amount = n
	return p
}

// OtherThanThis returns the shield with this object's damage left alone:
// Haze Frog's "other creatures". The object is the one the ability came
// from (CR 400.7): if it left and came back before the ability resolved,
// the new object is "another" creature, and its damage is prevented.
func (p PreventDamageFromSource) OtherThanThis() PreventDamageFromSource {
	p.exceptThis = true
	return p
}

// OtherThanTarget returns the shield with the damage of the target that
// answered clause i left alone, pinned as the object it is as the shield
// resolves: Terrifying Presence's "creatures other than target
// creature".
func (p PreventDamageFromSource) OtherThanTarget(clause int) PreventDamageFromSource {
	p.exceptClause, p.hasExceptClause = clause, true
	return p
}

// WithThen returns the shield with a CR 615.5 follow-up.
func (p PreventDamageFromSource) WithThen(then game.BodyRef) PreventDamageFromSource {
	p.Then = then
	return p
}

// DealingTo returns the shield with its follow-up dealing its damage to
// the target that answered clause i, chosen as the spell was cast
// (Refraction Trap, ADR 0108 §9). A target that is illegal as the shield
// resolves leaves the follow-up nothing to deal its damage to.
func (p PreventDamageFromSource) DealingTo(clause int) PreventDamageFromSource {
	p.thenTo, p.hasThenTo = clause, true
	return p
}

func (p PreventDamageFromSource) Apply(ctx *Context) error {
	label := p.Label
	if label == "" {
		label = shieldSourceName(ctx) + " — prevent damage from a source this turn"
	}
	// What the shield protects is read exactly as the next-damage
	// shield reads it, or, for several permanents in one record, by
	// protectMany (ADR 0108 Delivery PR 7).
	protected := game.NextDamageShield{Controller: ctx.Controller()}
	var many []uuid.UUID
	switch p.Protect.kind {
	case shieldTheTargetPermanents, shieldObjects:
		if many = p.Protect.protectMany(ctx); len(many) == 0 {
			return nil
		}
	case shieldRecipientSet:
		// #2045: the set is read as the damage would be dealt; nothing
		// is named now.
		if p.AndDealtBy {
			return nil
		}
	default:
		if !(PreventNextDamageFromSource{Protect: p.Protect}).protect(ctx, &protected) {
			return nil
		}
	}
	if p.AndDealtBy {
		// "Dealt to and dealt by" names the protected permanents as the
		// sources too, and nothing else.
		if protected.ProtectPermanent == uuid.Nil && len(many) == 0 {
			return nil
		}
		ctx.Game.PreventDamageFromSourceThisTurnForEffect(game.DamageShield{
			EffectSource:      ctx.Source(),
			Controller:        ctx.Controller(),
			ProtectPermanent:  protected.ProtectPermanent,
			ProtectPermanents: many,
			AndDealtBy:        true,
			CombatOnly:        p.CombatOnly,
			Then:              p.Then,
			UntilYourNextTurn: p.UntilYourNextTurn,
			Label:             label,
		})
		return nil
	}
	filter, ok := p.filter(ctx)
	if !ok {
		return nil
	}
	shield := game.DamageShield{
		EffectSource:      ctx.Source(),
		Controller:        ctx.Controller(),
		Queries:           p.Queries,
		Filter:            filter,
		ProtectPlayer:     protected.ProtectPlayer,
		ProtectTypes:      protected.ProtectTypes,
		ProtectPermanent:  protected.ProtectPermanent,
		ProtectPermanents: many,
		Recipients:        p.Protect.recipients,
		CombatOnly:        p.CombatOnly,
		Amount:            p.Amount,
		Then:              p.Then,
		ThenPerSource:     p.ThenPerSource,
		UntilYourNextTurn: p.UntilYourNextTurn,
		Label:             label,
	}
	if p.hasThenTo {
		if t, ok := ctx.ClauseTarget(p.thenTo); ok && (t.Kind == game.TargetCard || t.Kind == game.TargetPlayer) {
			shield.To = t.ID
		}
	}
	pick := shieldSourcePick{Choose: p.Choose, FromThis: p.FromThis, From: p.From, Queries: p.Queries, Question: p.Question}
	return pick.resolve(ctx, registerSourceShield(shield))
}

// filter is the shield's Filter with the objects pinned as it resolves
// added. False when one of those objects can't be named: the target is
// gone (and with it, for a spell whose only target it was, the spell) or
// the ability has no source object.
func (p PreventDamageFromSource) filter(ctx *Context) (game.DamageSourceFilter, bool) {
	f := p.Filter
	f.ExceptObjects = append([]game.ObjectRef(nil), f.ExceptObjects...)
	if p.exceptThis {
		ref, ok := ctx.SourceRef()
		if !ok {
			return f, false
		}
		f.ExceptObjects = append(f.ExceptObjects, ref)
	}
	if p.hasExceptClause {
		t, ok := ctx.ClauseTarget(p.exceptClause)
		if !ok || t.Kind != game.TargetCard {
			return f, false
		}
		ref, ok := ctx.Game.PermanentRefForEffect(t.ID)
		if !ok {
			return f, false
		}
		f.ExceptObjects = append(f.ExceptObjects, ref)
	}
	if len(f.ExceptObjects) == 0 {
		f.ExceptObjects = nil
	}
	return f, true
}

// registerSourceShield finishes a source shield once its source is
// pinned. It captures the shield's plain description and nothing else.
func registerSourceShield(shield game.DamageShield) func(*game.Game, game.ObjectRef, game.ZoneKind) error {
	return func(g *game.Game, ref game.ObjectRef, zone game.ZoneKind) error {
		s := shield
		s.Source, s.SourceZone = ref, zone
		g.PreventDamageFromSourceThisTurnForEffect(s)
		return nil
	}
}

// sourceShieldSpell is an instant's OnResolve whose whole effect is a
// source shield.
func sourceShieldSpell(shield PreventDamageFromSource) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return shield.Apply(ctx)
	}
}

// sourceShieldRow is an activated ability whose whole effect is a source
// shield (the shared row, next_damage_shield_cards.go).
func sourceShieldRow(label string, cost game.AbilityCost, targets *game.TargetSpec, shield PreventDamageFromSource) ActivatedAbility {
	return nextDamageShieldRow(label, cost, targets, shield)
}

// --- the follow-ups (CR 615.5) ----------------------------------------

var (
	// "Whenever damage from a black or red source is prevented this way
	// this turn, you gain that much life" (Samite Ministration). A
	// triggered ability (CR 615.13: once each time the shield is applied
	// to simultaneous damage and prevents some of it), so it goes on the
	// stack, New Way Forward's reflexive shape.
	preventedBlackOrRedTriggerBody = game.DelayedBody("prevention/black-or-red-gain-life-trigger", preventedBlackOrRedTrigger)
	preventedGainLifeTriggerBody   = game.DelayedBody("prevention/gain-life-trigger", preventedGainLife)
)

func preventedBlackOrRedTrigger(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 || item == nil || item.Trigger == nil {
		return nil
	}
	colors := item.Trigger.Event.Colors
	if !slices.Contains(colors, "B") && !slices.Contains(colors, "R") {
		return nil
	}
	t := WhenYouDo(shieldSourceName(NewContext(g, item))+" — you gain the prevented damage as life", preventedGainLifeTriggerBody)
	t.Params = game.EffectParams{Amount: p.Amount}
	return t.Apply(NewContext(g, item))
}
