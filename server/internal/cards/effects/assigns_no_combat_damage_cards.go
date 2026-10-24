package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// assigns_no_combat_damage_cards.go — the shared shapes of the "whenever
// this creature attacks and isn't blocked, you may <do something>. If you
// do, this creature assigns no combat damage this turn" cards (Ophidian,
// Farrel's Zealot, Floral Spuzzem and the rest of #3531). The rule itself
// is game.AssignsNoCombatDamageThisTurnForEffect (CR 510.1a).

// assignsNoCombatDamageThisTurn is "this creature assigns no combat damage
// this turn": the resolving ability's source, pinned as the object it is
// now, so a creature that left and came back is not marked (CR 400.7).
func assignsNoCombatDamageThisTurn(ctx *Context, label string) {
	if ctx.isNewSourceObjectAsThis(ctx.Source()) {
		return
	}
	ctx.Game.AssignsNoCombatDamageThisTurnForEffect(ctx.Source(), ctx.Source(), label)
}

// unblockedAttackerMay builds "whenever this creature attacks and isn't
// blocked, you may <do>. If you do, this creature assigns no combat damage
// this turn". `do` gets the ability's first legal card target (uuid.Nil
// for an ability with none) and runs only on "yes". The caller adds the
// target clause (Targeting, or TargetsFrom for "defending player controls").
func unblockedAttackerMay(card, question, key string, do func(ctx *Context, target uuid.UUID) error) game.TriggeredAbility {
	return unblockedAttackerMayDraw(card, question, 0, key, do)
}

// unblockedAttackerMayDraw is unblockedAttackerMay for a "yes" that is a
// choice to draw `draws` cards (Ophidian's "you may draw a card"), which a
// player a draw limit stops drawing can't take (MayChoice.Draws, #4016).
//
// `key` registers the "yes" as a keyed answer, "confirm/<key>" (#4291): an
// on-disk identity, never renamed, never reused. This is called once per
// card, from its init.
func unblockedAttackerMayDraw(card, question string, draws int, key string, do func(ctx *Context, target uuid.UUID) error) game.TriggeredAbility {
	label := card + " — " + question
	then := MayYes("confirm/"+key, func(ctx *Context, r game.ConfirmAnswer) error {
		if err := do(ctx, carryAt(r, 0)); err != nil {
			return err
		}
		assignsNoCombatDamageThisTurn(ctx, card+" — assigns no combat damage this turn")
		return nil
	})
	return WhenAttacksAndIsNotBlockedEffect(label, func(g *game.Game, item *game.StackItem, _ uuid.UUID) error {
		ctx := NewContext(g, item)
		target := firstLegalCardTarget(ctx)
		return MayChoice{
			Question: label + "?",
			Draws:    draws,
			Then:     then,
			Carry:    []uuid.UUID{target},
		}.Apply(ctx)
	})
}

// laccolithBlocked builds "whenever this creature becomes blocked, you may
// have it deal damage equal to its power to target creature. If you do, this
// creature assigns no combat damage this turn" (the Laccolith creatures).
// The target is chosen as the trigger goes on the stack and "you may" is
// asked as it resolves.
func laccolithBlocked(card string) game.TriggeredAbility {
	label := card + " — you may have it deal damage equal to its power to target creature"
	return Targeting(On(game.EventBecomesBlocked, Self, label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := firstLegalCardTarget(ctx)
		return MayChoice{
			Question: label + "? (it then assigns no combat damage this turn)",
			Then:     mayLaccolithDamage,
			Word:     card,
			Carry:    []uuid.UUID{target},
		}.Apply(ctx)
	}), TargetCreature("target creature"))
}

// enchantedCreatureAttacksAndIsNotBlocked is the AppliesTo half of "whenever
// enchanted creature attacks and isn't blocked" (CR 509.3): `ev` is the
// completed block declaration of the player the enchanted creature attacks,
// and nothing blocks it. See attacksAndIsNotBlocked.
func enchantedCreatureAttacksAndIsNotBlocked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBlockersDeclared || source == nil || source.AttachedTo.Kind != game.TargetCard {
		return false
	}
	host := source.AttachedTo.ID
	hc, ok := g.LookupCardForEffect(host)
	if !ok || hc.AttackingTarget == uuid.Nil {
		return false
	}
	if g.DefendingPlayerForAttackerForEffect(host) != ev.Actor {
		return false
	}
	return g.UnblockedAttackerForEffect(host)
}

// enchantedCreatureOf is the creature the resolving Aura's source is
// attached to, or uuid.Nil if the Aura has left or is unattached.
func enchantedCreatureOf(ctx *Context) uuid.UUID {
	aura, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return uuid.Nil
	}
	return aura.AttachedTo.ID
}

// enchantedAssignsNoCombatDamage is "<the enchanted creature> assigns no
// combat damage this turn" for an Aura's ability.
func enchantedAssignsNoCombatDamage(ctx *Context, host uuid.UUID, label string) {
	ctx.Game.AssignsNoCombatDamageThisTurnForEffect(ctx.Source(), host, label)
}

// firstLegalCardTarget is the first still-legal card target of the
// resolving ability, in any zone, or uuid.Nil.
func firstLegalCardTarget(ctx *Context) uuid.UUID {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			return t.ID
		}
	}
	return uuid.Nil
}

// sourcePowerNow is "its power" for the resolving ability's source: its
// current power on the battlefield, else its last known power (CR 608.2h).
func sourcePowerNow(ctx *Context) int {
	id := ctx.Source()
	if onBattlefield(ctx.Game, id) {
		return b31CurrentPowerOf(ctx.Game, id)
	}
	return b17LastKnownPowerOffBattlefield(ctx.Game, id)
}

// withTargetsFrom gives a trigger a TargetsFrom clause (chosen as the
// ability goes on the stack, from the board as it is then).
func withTargetsFrom(t game.TriggeredAbility, f func(game.TriggerContext, *game.Card, *game.Game) *game.TargetSpec) game.TriggeredAbility {
	t.TargetsFrom = f
	return t
}

// TargetOfDefendingPlayer is the TargetsFrom clause for "target <thing>
// defending player controls" on a trigger that fires while the source is
// attacking (Floral Spuzzem, Orcish Squatters): the defending player is
// read off the source's own attack, as TargetCreatureDefendingPlayerControls
// does. `pred` narrows the permanent.
func TargetOfDefendingPlayer(mode, label string, pred CardPredicate) func(game.TriggerContext, *game.Card, *game.Game) *game.TargetSpec {
	return func(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
		defender := g.DefendingPlayerForAttackForEffect(source.AttackingTarget)
		if defender == uuid.Nil {
			return nil
		}
		return &game.TargetSpec{
			Mode: mode, Label: label,
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.Controller == defender && pred(g, caster, c)
			},
			Min: 1, Max: 1,
		}
	}
}

// mayLaccolithDamage is the Laccolith creatures' "yes": it deals damage
// equal to its power to the target, then assigns no combat damage this turn
// (Word is the card's name). Keyed (#4291).
var mayLaccolithDamage = MayYes("confirm/laccolith-blocked-damage", func(ctx *Context, r game.ConfirmAnswer) error {
	if err := (DealDamage{Source: ctx.Source(), Target: carryAt(r, 0), Amount: sourcePowerNow(ctx)}).Apply(ctx); err != nil {
		return err
	}
	assignsNoCombatDamageThisTurn(ctx, r.Word+" — assigns no combat damage this turn")
	return nil
})
