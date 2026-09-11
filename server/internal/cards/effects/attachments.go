package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachments.go is the shared catalog surface for S24's Equipment
// and Auras (ADR 0036). Everything an equipment or aura card file
// needs that is not the card itself lives here:
//
//	EquipAbility("{2}")           the CR 702.6 activated ability
//	AttachedToSource              "equipped creature" / "enchanted creature"
//	PumpAttached(2, 2)            a layer 7c static on the host
//	GrantToAttached("haste")      a layer 6 keyword grant on the host
//	EnchantCreature(...)          the aura's enchant clause
//
// The deliberate absence here is a `Spec.Equip` or `Spec.Enchant`
// field. Equip is an ordinary `Spec.Activated` entry, because it IS
// an ordinary activated ability — a mana cost, a target clause, a
// sorcery-speed gate — and every one of those already exists and is
// already enumerated by the bot, rendered by the right-click menu
// and validated at announce. An aura's enchant clause is an ordinary
// `Spec.Targets`. Neither needed a new engine verb.

// EquipAbility builds the CR 702.6 equip ability: "{cost}: Attach to
// target creature you control. Equip only as a sorcery."
//
// Three clauses, all of them load-bearing and none of them new:
//
//   - SorcerySpeed is CR 702.6b's "any time you could cast a
//     sorcery", gated by the engine's existing main-phase /
//     empty-stack / active-player check.
//   - The target clause is "creature you control", validated at
//     announce (CR 601.2c) and re-checked at resolution (CR 608.2b),
//     so an equip whose target dies in response fizzles with no card
//     code. Note the restriction is checked only WHILE ACTIVATING:
//     once attached, control of the creature may change and the
//     Equipment stays put (see attachmentLegalLocked in the game
//     package for the other half of that rule).
//   - Re-activating moves the Equipment (CR 702.6d), which the
//     attach primitive gets right by overwriting.
func EquipAbility(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label:        "Equip " + cost,
		Cost:         ManaCost(cost),
		Targets:      TargetCreature("target creature you control", YouControl()),
		SorcerySpeed: true,
		Effect:       AttachSourceToTarget,
	}
}

// AttachSourceToTarget is the equip ability's resolution: attach the
// source permanent to the chosen creature.
//
// Reads the source off the StackItem by ID rather than capturing a
// *Card, which is the standing rule for anything that resolves off
// the stack — the battlefield slice may have been reallocated since
// the ability was announced.
//
// A target that became illegal in response is skipped rather than
// errored: CR 608.2b's re-check says the ability does as much as it
// can, and for equip with one target that is nothing at all.
func AttachSourceToTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return g.AttachForEffect(item.SourceCardID, t)
	}
	return nil
}

// AttachedToSource is the AppliesTo predicate for every "equipped
// creature" and "enchanted creature" static. One card scoped by one
// relation — the same shape as selfOnly, which is the precedent for
// a static that applies to exactly one permanent.
//
// Evaluated per battlefield card per recompute, so it re-reads the
// current attachment every pass: moving a sword with a second equip
// moves its bonus in the same beat, with no invalidation bookkeeping
// beyond the EventAttach the primitive already emits.
func AttachedToSource(target *game.Card, _ *game.Game, source *game.Card) bool {
	return source.IsAttachedTo(target.InstanceID)
}

// PumpAttached is "Equipped creature gets +P/+T" / "Enchanted
// creature gets +P/+T" — a layer 7c modify, which is where every
// non-setting P/T change belongs (CR 613.4c).
func PumpAttached(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// GrantToAttached is "Equipped creature has X" / "Enchanted creature
// has X" for one or more keywords — a layer 6 ability grant.
//
// Keywords MUST be canonical lowercase tokens the engine actually
// honours (game.CanonicalKeyword is the table). A grant of something
// outside it would render as a badge the rules layer never backs,
// which is the one failure mode ADR 0038 exists to prevent.
//
// Deduped on apply, so an Equipment granting haste to a creature
// that already prints haste does not produce two badges.
func GrantToAttached(keywords ...string) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, kw := range keywords {
				if !keywordSliceContains(c.Abilities, kw) {
					c.Abilities = append(c.Abilities, kw)
				}
			}
		},
	}
}

// EnchantCreature is an Aura's "Enchant creature" clause. It is an
// ordinary target spec, because at cast time that is exactly what it
// is (CR 303.4a) — and because the state-based action that keeps the
// Aura legal afterwards re-runs this same spec, one definition
// serves both.
func EnchantCreature(preds ...CardPredicate) *game.TargetSpec {
	return TargetCreature("enchant creature", preds...)
}

// EnchantPlayer is a Curse's "Enchant player" clause. The reason
// Card.AttachedTo is a TargetRef and not a card ID.
func EnchantPlayer(preds ...PlayerPredicate) *game.TargetSpec {
	return TargetPlayer("enchant player", preds...)
}

// equippedCreatureDied reports whether ev is the death of the
// creature this Equipment is attached to — Skullclamp's trigger
// condition.
//
// The ordering this depends on was checked end to end and is not
// obvious: the trigger harvester runs synchronously inside EmitEvent
// under the write lock, and the CR 704.5m unattach is a state-based
// action that has not run yet. So at the instant EventLTB fires, the
// Equipment is still on the battlefield still pointing at the card
// that just died, and a direct read of AttachedTo is correct. Had it
// gone the other way this would have needed the LKI snapshot to grow
// an attachment field, which it has not got.
func equippedCreatureDied(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventLTB &&
		ev.NewZone == game.ZoneGraveyard &&
		source.IsAttachedTo(ev.CardID)
}

// attachedCreatureDealtCombatDamageToPlayer is the Swords' trigger
// condition: "Whenever equipped creature deals combat damage to a
// player". The same EventDealDamage + ev.Combat shape Bident of
// Thassa uses, with the controller check swapped for the attachment.
func attachedCreatureDealtCombatDamageToPlayer(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if !source.IsAttachedTo(ev.Source) {
		return false
	}
	return g.PlayerByIDForEffect(ev.Target) != nil
}

// untapAllLandsControlledBy is Sword of Feast and Famine's back half
// ("untap all lands you control"), which is a plain loop rather than
// a primitive because nothing else in the catalog untaps a set.
//
// Reads the battlefield through the effect API under the resolution
// write lock; collects IDs first so the untap does not mutate the
// slice being walked.
func untapAllLandsControlledBy(g *game.Game, controller uuid.UUID, ctx *Context) error {
	var ids []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && c.IsLand() && c.Tapped {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (UntapTarget{Target: id}.Apply(ctx)); err != nil {
			return err
		}
	}
	return nil
}
