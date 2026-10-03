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

	// Protect is what the shield protects.
	Protect ShieldTarget

	// CombatOnly is "all combat damage".
	CombatOnly bool

	// Amount is "the next N damage" (CR 615.7); zero is all damage this
	// turn.
	Amount int

	// Then is the CR 615.5 follow-up. Zero is none.
	Then game.BodyRef

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

// WithThen returns the shield with a CR 615.5 follow-up.
func (p PreventDamageFromSource) WithThen(then game.BodyRef) PreventDamageFromSource {
	p.Then = then
	return p
}

func (p PreventDamageFromSource) Apply(ctx *Context) error {
	label := p.Label
	if label == "" {
		label = shieldSourceName(ctx) + " — prevent damage from a source this turn"
	}
	// What the shield protects is read exactly as the next-damage
	// shield reads it.
	protected := game.NextDamageShield{Controller: ctx.Controller()}
	if !(PreventNextDamageFromSource{Protect: p.Protect}).protect(ctx, &protected) {
		return nil
	}
	shield := game.DamageShield{
		EffectSource:     ctx.Source(),
		Controller:       ctx.Controller(),
		Queries:          p.Queries,
		ProtectPlayer:    protected.ProtectPlayer,
		ProtectTypes:     protected.ProtectTypes,
		ProtectPermanent: protected.ProtectPermanent,
		CombatOnly:       p.CombatOnly,
		Amount:           p.Amount,
		Then:             p.Then,
		Label:            label,
	}
	pick := shieldSourcePick{Choose: p.Choose, FromThis: p.FromThis, From: p.From, Queries: p.Queries, Question: p.Question}
	return pick.resolve(ctx, registerSourceShield(shield))
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
