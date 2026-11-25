package effects

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// restrictions.go — the card-facing primitives for S24's restriction
// vocabulary: "enchanted creature can't attack or block" (Pacifism,
// Arrest, Faith's Fetters), "equipped creature can't be blocked"
// (Whispersilk Cloak), "this creature can't block" (Carrion Feeder),
// "target creature can't be blocked this turn" (Rogue's Passage).
//
// The bits and the engine gates that read them live in
// server/internal/game/restrictions.go; that file's doc comment is
// the taxonomy. This one is three one-liners over the same
// `game.StaticAbility` every other static uses, in the three
// durations a card can want:
//
//	RestrictAttached(…)   for as long as the Aura / Equipment is on
//	                      its host        (Pacifism, the Cloak)
//	RestrictSelf(…)       for as long as the permanent is on the
//	                      battlefield     (Carrion Feeder)
//	RestrictUntilEOT{…}   until end of turn (Rogue's Passage)
//
// WHICH LAYER. Layer 6, and the choice does not matter. A
// restriction is not applied in any CR 613 layer — it is not a
// characteristic — but the layer pass is the only machinery that
// knows which permanents a continuous effect applies to, so the
// restriction rides through it. Because `Characteristic.Restrictions`
// is only ever OR'd into and nothing clears it, timestamp order
// between two restriction effects is unobservable, and a layer-6
// "loses all abilities" cannot strip a Pacifism. That last part is
// the rules-correct outcome, not a happy accident: the "can't
// attack" belongs to the Aura, not to the creature. (The one exception
// is RestrictUntilEOT's mass form, whose live set is folded in after
// the last layer so it reads finished characteristics — #1650.)

// RestrictAttached is "Enchanted creature can't attack or block" /
// "Equipped creature can't be blocked" — a restriction scoped to
// whatever the source is attached to, by the same AttachedToSource
// predicate every other attachment static uses.
//
// The restriction ends the moment the attachment does, with no
// bookkeeping: the AppliesTo re-reads the relation on every
// recompute, so destroying the Pacifism un-pacifies the creature in
// the same pass.
func RestrictAttached(r game.Restriction) game.StaticAbility {
	return game.StaticAbility{
		Label:     restrictedLabel("Attached creature ", r, 0),
		Layer:     game.Layer6Ability,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= r
		},
	}
}

// RestrictAttachedExtra is RestrictAttached with bits from the second
// restriction set as well (game.ExtraRestriction: "can't crew
// Vehicles", Intercessor's Arrest, #5673), for a card whose one printed
// sentence spans both sets.
func RestrictAttachedExtra(r game.Restriction, x game.ExtraRestriction) game.StaticAbility {
	return game.StaticAbility{
		Label:     restrictedLabel("Attached creature ", r, x),
		Layer:     game.Layer6Ability,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= r
			c.ExtraRestrictions |= x
		},
	}
}

// RestrictSelf is a restriction a permanent prints on itself —
// Carrion Feeder's "This creature can't block". Scoped by selfOnly,
// the same predicate a self-referential pump uses.
func RestrictSelf(r game.Restriction) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= r
		},
		Label: selfRestrictionLabel(r),
	}
}

// RestrictSelfExtra is RestrictSelf with bits from the second
// restriction set as well (game.ExtraRestriction).
func RestrictSelfExtra(r game.Restriction, x game.ExtraRestriction) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= r
			c.ExtraRestrictions |= x
		},
		Label: restrictedLabel("", r, x),
	}
}

// CantAttackOrBlockUnlessPowerAtLeast is "This creature can't attack or
// block unless its power is N or greater" (Technodrome, #4678). Not a
// RestrictSelf with a condition: the condition is the creature's own
// power, which a layer-6 static cannot read (layer 7 has not run yet).
// The static records the floor and the engine folds it into
// CantAttack | CantBlock once every layer has run (CR 613.11,
// game/power_floor.go), so an anthem or counters that lift it to N make
// it eligible and a -N/-N takes it back out. A self static, so it goes
// with the creature's abilities (CR 613.1f).
func CantAttackOrBlockUnlessPowerAtLeast(n int) game.StaticAbility {
	return game.StaticAbility{
		Label:     fmt.Sprintf("This creature can't attack or block unless its power is %d or greater.", n),
		Layer:     game.Layer6Ability,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if n > c.AttackBlockPowerFloor {
				c.AttackBlockPowerFloor = n
			}
		},
	}
}

// selfRestrictionLabel is RestrictSelf's tile label (#2219): the
// printed clause for the three single restrictions a permanent prints
// on itself, and nothing for any other set, which the tile then
// describes by its layer.
func selfRestrictionLabel(r game.Restriction) string {
	return restrictedLabel("", r, 0)
}

// restrictedLabel is "<prefix>can't attack or block." for a restriction
// set, or nothing for a set restrictionClause does not word. With no
// prefix the sentence opens on the clause ("Can't attack.").
func restrictedLabel(prefix string, r game.Restriction, x game.ExtraRestriction) string {
	clause := restrictionClause(r, x)
	if clause == "" {
		return ""
	}
	l := prefix + clause + "."
	return strings.ToUpper(l[:1]) + l[1:]
}

// RestrictUntilEOT is "target creature can't be blocked this turn"
// (Rogue's Passage) or a mass form ("creatures without flying can't
// block this turn", Falter). The turn-scoped sibling of
// RestrictAttached, registered as an ADR 0041 data record and swept at
// cleanup (CR 514.2).
//
// THE TWO FORMS ARE DIFFERENT EFFECTS, and only one of them locks a
// set.
//
//   - Target is "target creature can't block this turn". It follows
//     that one object: a creature that leaves and returns is a new
//     object (CR 400.7) and is no longer covered, which the entry
//     stamp in the pinned record enforces.
//   - Scope is the mass form, and its set is a RULE read live for the
//     rest of the turn (#1650). CR 611.2c locks the affected set only
//     for an effect that changes characteristics or control. "Can't
//     block", "can't be blocked" and "can't attack" change neither;
//     they change the rules of the game. So Falter also stops a
//     creature flashed in after it resolved, and Glaring Spotlight's
//     "creatures you control … can't be blocked this turn" also
//     covers the creature you cast next. BoostUntilEOT and
//     GrantKeywordUntilEOT are the opposite case (they DO change
//     characteristics) and keep their snapshot.
//
// A mass form takes a closed game.AffectedScope rather than a
// CardPredicate because a closure cannot be carried by undo or a
// restore point, and this one has to be re-read on every pass until
// cleanup. A printed set that no scope names needs a new scope in
// game/scoped_effects.go.
type RestrictUntilEOT struct {
	// Target pins the effect to one permanent. Ignored when Scope is
	// set.
	Target uuid.UUID

	// Scope is the mass form's affected set, read live until cleanup
	// against the controller of the resolving spell or ability (the
	// "you" of game.ScopeYourCreatures / game.ScopeOpponentsCreatures).
	Scope game.AffectedScope

	// Restrictions is the bit set to add. game.CantAttackOrBlock is
	// the common pair.
	Restrictions game.Restriction

	// ExtraRestrictions is the second set's bits to add
	// (game.CantCrew).
	ExtraRestrictions game.ExtraRestriction

	// Queries is the FILTER of the two query-filtered scopes (#6237,
	// game.ScopeCreaturesMatching and ScopeOpponentsCreaturesMatching):
	// Goblin Locksmith's "creatures with defender", Rumbling Ruin's
	// "creatures your opponents control with power less than or equal to
	// that number". Any of them matching is enough, and a number in a
	// query is locked as the ability resolves while the SET stays live
	// (CR 611.2c). Required with those scopes and refused with any
	// other, so a filter can never be silently dropped.
	Queries []game.PermanentQuery

	// Label is attribution for logs and tests; defaults to a
	// generic string when empty.
	Label string
}

func (r RestrictUntilEOT) Apply(ctx *Context) error {
	if r.Restrictions == 0 && r.ExtraRestrictions == 0 {
		return nil
	}
	label := eotLabel(r.Label, "restriction until end of turn")
	mod := game.AddRestrictionsMod(r.Restrictions)
	mod.ExtraRestrictions = r.ExtraRestrictions
	if len(r.Queries) > 0 {
		mod = game.MatchingRestrictionMod(r.Restrictions, r.Queries)
		mod.ExtraRestrictions = r.ExtraRestrictions
	}
	if r.Scope != game.ScopeNone {
		if !game.KnownAffectedScope(r.Scope) {
			return fmt.Errorf("effects: RestrictUntilEOT %q: unknown scope %q", label, r.Scope)
		}
		if game.ScopeFiltersByQueries(r.Scope) != (len(r.Queries) > 0) {
			return fmt.Errorf("effects: RestrictUntilEOT %q: scope %q and %d queries do not go together", label, r.Scope, len(r.Queries))
		}
		ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), r.Scope, ctx.Controller(),
			[]game.Mod{mod}, DurationUntilEndOfTurn(ctx), label)
		return nil
	}
	if len(r.Queries) > 0 {
		return fmt.Errorf("effects: RestrictUntilEOT %q: queries need a query-filtered scope", label)
	}
	return UntilEndOfTurn(ctx, r.Target, nil, label, mod)
}

// CantBlockThisTurnMatching is "creatures <matching> can't block this
// turn" (#6237): Goblin Locksmith's "creatures with defender can't block
// this turn", and — with the opponents' form —  Rumbling Ruin's
// "creatures your opponents control with power less than or equal to
// that number can't block this turn".
//
// The filter is data (game.PermanentQuery), so a number in it is locked
// as the ability resolves while the SET stays live until cleanup: a
// creature that arrives this turn and matches cannot block either
// (CR 611.2c, 509.1b).
func CantBlockThisTurnMatching(opponentsOnly bool, label string, queries ...game.PermanentQuery) RestrictUntilEOT {
	scope := game.ScopeCreaturesMatching
	if opponentsOnly {
		scope = game.ScopeOpponentsCreaturesMatching
	}
	return RestrictUntilEOT{
		Scope:        scope,
		Restrictions: game.CantBlock,
		Queries:      queries,
		Label:        label,
	}
}

// RestrictAttachedWhile is "During your turn, equipped creature can't
// be blocked" (Bilbo's Ring) — the restriction-bit sibling of
// GrantToAttachedWhile. The Aura/Equipment relation still gates
// through AttachedToSource, `cond` adds the printed condition on top
// (read fresh on every recompute, same as GrantToAttachedWhile's),
// and both must hold for the bits to apply.
func RestrictAttachedWhile(cond func(host *game.Card, g *game.Game, source *game.Card) bool, r game.Restriction) game.StaticAbility {
	return game.StaticAbility{
		Label: strings.TrimSuffix(restrictedLabel("Attached creature ", r, 0), ".") + " while a condition holds.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return AttachedToSource(target, g, source) && cond(target, g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= r
		},
	}
}
