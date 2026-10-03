package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shield_families_source.go — ADR 0108 Delivery PR 7b (#1904): the
// object-source family of §7's preventFromSource shield. "Prevent all
// [combat] damage [that would be dealt by] target creature [would deal]
// this turn" (Fend Off, Kor Haven, Warning) is a preventFromSource record
// whose source is the TARGET, pinned as the spell or ability resolves
// with no prompt (CR 400.7), protecting anything (prevent_from_source.go,
// PreventDamageFromSource{From: id}). The property family (Luminesce,
// Ethereal Haze, Scarecrow) is the same shield with Queries and no source,
// written in each card file.
//
// ONE RECORD PER SOURCE. "Prevent all damage one or two target creatures
// would deal" (Soul Parry), "that creature and each creature blocking it"
// (Feint) and "X target creatures" (Serene Sunset) are one sentence about
// several sources. Each source gets its own record: a damage event has
// one source, so no event meets two of them, and nothing a player can see
// differs from one effect naming them all (none of these cards has a
// "prevented this way" follow-up, which is the one reader that counts
// applications, CR 615.13).
//
// A target that has become illegal is skipped and the rest still happens
// (CR 608.2b); with every target illegal the spell or ability does not
// resolve at all.
//
// Append-only, mechanic-named: the clone gate sees one body per shape.

// shieldAgainstEachTarget is "Prevent all [combat] damage <each target>
// would deal this turn": a preventFromSource record per still-legal card
// target, the target as its source, protecting anything.
func shieldAgainstEachTarget(ctx *Context, combatOnly bool) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (PreventDamageFromSource{From: t.ID, CombatOnly: combatOnly, Protect: ShieldAnything}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// shieldAgainstTargetsSpell is an instant's OnResolve whose whole effect
// is shieldAgainstEachTarget.
func shieldAgainstTargetsSpell(combatOnly bool) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return shieldAgainstEachTarget(ctx, combatOnly)
	}
}

// shieldAgainstTargetsRow is an activated ability whose whole effect is
// shieldAgainstEachTarget: "<cost>: Prevent all combat damage that would
// be dealt by target creature this turn" (Kor Haven, Horn of Deafening,
// Safeguard, Lady Evangela).
func shieldAgainstTargetsRow(label string, cost game.AbilityCost, targets *game.TargetSpec, combatOnly bool) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Cost:    cost,
		Targets: targets,
		Effect: func(g *game.Game, item *game.StackItem) error {
			return shieldAgainstEachTarget(NewContext(g, item), combatOnly)
		},
	}
}

// shieldAgainstTheTargetThen is "Prevent all [combat] damage target
// creature would deal this turn" followed by a sentence about that
// creature (Restrain's draw, Subdue's toughness, Inquisitor's Snare's
// destruction): the shield on the first still-legal card target, then
// `then` with that target. Nothing happens when the target has gone,
// which is CR 608.2b for a spell or ability with one target.
func shieldAgainstTheTargetThen(ctx *Context, combatOnly bool, then func(ctx *Context, target uuid.UUID) error) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (PreventDamageFromSource{From: t.ID, CombatOnly: combatOnly, Protect: ShieldAnything}).Apply(ctx); err != nil {
			return err
		}
		if then == nil {
			return nil
		}
		return then(ctx, t.ID)
	}
	return nil
}

// toughnessByManaValueUntilEOT is "That creature gets +0/+X until end of
// turn, where X is its mana value" (Subdue, Kry Shield): X read as the
// spell or ability resolves.
func toughnessByManaValueUntilEOT(ctx *Context, target uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	mv, ok := ctx.Game.ManaValueForEffect(c)
	if !ok || mv <= 0 {
		return nil
	}
	return BoostUntilEOT{Target: target, Toughness: mv, Label: shieldSourceName(ctx) + " — +0/+X until end of turn"}.Apply(ctx)
}

// blockersOf lists the creatures blocking `attacker` right now, in
// battlefield order (Feint's "each creature blocking it").
func blockersOf(g *game.Game, attacker uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsBlockingAttacker(attacker) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// blockingCreatures lists every creature blocking anything right now, in
// battlefield order (Fighting Chance's "each blocking creature").
func blockingCreatures(g *game.Game) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.BlockingTarget != uuid.Nil {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// shieldAgainstThisCombatDamage is "prevent all combat damage this
// creature would deal this turn" from the creature's own triggered
// ability (Ignoble Soldier, Loafing Giant, Zealot il-Vec, Mtenda Lion):
// the source is the permanent the ability came from, as long as it is
// still that object (CR 400.7).
func shieldAgainstThisCombatDamage(ctx *Context) error {
	return PreventDamageFromSource{FromThis: true, CombatOnly: true, Protect: ShieldAnything}.Apply(ctx)
}
