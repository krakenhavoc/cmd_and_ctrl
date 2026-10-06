package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// block_requirements.go — the card-facing half of CR 509.1c block
// requirements (#1597, ADR 0045 amendment of 2026-09-28). The engine
// half — what a requirement is, how a declaration is judged against
// the most requirements it could obey, where it is refused — lives in
// game/block_requirements.go. A card file only says WHICH creatures a
// requirement reaches, WHAT it asks and for HOW LONG:
//
//	Static: []game.StaticAbility{BlocksEachCombat()},               // Razorgrass Screen
//	Static: []game.StaticAbility{BlocksEachCombatWhere(allCreatures)}, // Grand Melee
//	Static: []game.StaticAbility{AllAbleToBlockDoSo()},             // Prized Unicorn
//	Static: []game.StaticAbility{BlockRequirementWhere(game.BlockRequirementLure, AttachedToSource)}, // Lure
//	Static: []game.StaticAbility{MustBeBlocked()},                  // Gaea's Protector
//	Static: []game.StaticAbility{FilteredLure(game.BlockerFilterWall)}, // Marble Priest
//	BlockRequirementUntilEOT{Target: t, Kind: game.BlockRequirementMustBeBlocked}.Apply(ctx) // Irresistible Prey
//	BlocksAttackerUntilEOT{Blocker: t, Attacker: ref}.Apply(ctx)    // Provoke, Grappling Hook
//	Triggered: []game.TriggeredAbility{Provoke()},                  // Goblin Grappler
//
// Two sides, one slot. "Blocks each combat if able" sits on the
// creature that must block; Lure, "must be blocked" and "must be
// blocked by exactly one creature" sit on the ATTACKER, and the engine
// reads them off it when that attacker is declared. Either way the
// requirement is an entry in Characteristic.BlockRequirements, one per
// requirement, because CR 509.1c counts them.

// BlockRequirementWhere is a static that gives every creature
// `appliesTo` matches one block requirement of `kind`, attributed to
// the source so the refusal names the card. The requirement belongs to
// the SOURCE, so a creature that loses all its abilities under Grand
// Melee still has to block (Characteristic.BlockRequirements is never
// cleared by a layer-6 removal, exactly like Restrictions); a
// creature's OWN requirement goes with its abilities, because the
// catalog static that writes it is not applied once CatalogAbilityKey
// answers empty (CR 613.1f).
func BlockRequirementWhere(kind game.BlockRequirementKind, appliesTo func(target *game.Card, g *game.Game, source *game.Card) bool) game.StaticAbility {
	return blockRequirementStatic(kind, "", appliesTo)
}

// blockRequirementStatic is BlockRequirementWhere with a blocker filter
// key (#1684). It panics — at catalog load, so at boot — on a kind that
// can't be written by a static (blocksAttacker names an attacking
// object, which only a resolving effect knows) and on a filter this
// binary has not registered or that is set on anything but a Lure.
func blockRequirementStatic(kind game.BlockRequirementKind, filter string, appliesTo func(target *game.Card, g *game.Game, source *game.Card) bool) game.StaticAbility {
	if kind == game.BlockRequirementBlocksAttacker {
		panic("effects: a blocksAttacker requirement names an attacking object; use BlocksAttackerUntilEOT")
	}
	if filter != "" && (kind != game.BlockRequirementLure || !game.KnownBlockerFilter(filter)) {
		panic("effects: blocker filter " + filter + " is not a registered filter on a Lure")
	}
	return game.StaticAbility{
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && appliesTo(target, g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			r := game.BlockRequirement{Kind: kind, Filter: filter}
			if source != nil {
				r.Source, r.SourceName = source.InstanceID, source.Name
			}
			c.BlockRequirements = append(c.BlockRequirements, r)
		},
	}
}

// BlocksEachCombat is "This creature blocks each combat if able"
// (Razorgrass Screen, Watchdog).
func BlocksEachCombat() game.StaticAbility {
	return BlocksEachCombatWhere(selfOnly)
}

// BlocksEachCombatWhere is "<creatures> block each combat if able" from
// a permanent — Grand Melee's "All creatures".
func BlocksEachCombatWhere(appliesTo func(target *game.Card, g *game.Game, source *game.Card) bool) game.StaticAbility {
	return BlockRequirementWhere(game.BlockRequirementBlocks, appliesTo)
}

// AllAbleToBlockDoSo is "All creatures able to block this creature do
// so" (Prized Unicorn). For "… enchanted creature …" (Lure) pass
// AttachedToSource to BlockRequirementWhere instead.
func AllAbleToBlockDoSo() game.StaticAbility {
	s := BlockRequirementWhere(game.BlockRequirementLure, selfOnly)
	s.Label = "All creatures able to block it do so." // #2219: the tile's label
	return s
}

// FilteredLure is "All <filter> able to block this creature do so"
// (#1684): Marble Priest's Walls (game.BlockerFilterWall), Talruum
// Piper's creatures with flying (game.BlockerFilterFlying). The filter
// is a registered KEY on the requirement, never a closure, so the
// characteristic that carries it stays plain comparable data.
func FilteredLure(filter string) game.StaticAbility {
	if filter == "" {
		panic("effects: FilteredLure needs a filter; use AllAbleToBlockDoSo for an unfiltered Lure")
	}
	return blockRequirementStatic(game.BlockRequirementLure, filter, selfOnly)
}

// MustBeBlocked is "This creature must be blocked if able" (Gaea's
// Protector).
func MustBeBlocked() game.StaticAbility {
	s := BlockRequirementWhere(game.BlockRequirementMustBeBlocked, selfOnly)
	s.Label = "Must be blocked if able." // #2219: the tile's label
	return s
}

// BlockRequirementUntilEOT is a RESOLVING effect's block requirement on
// one creature for the turn: Irresistible Prey's "target creature must
// be blocked this turn if able", Taunting Challenge's "all creatures
// able to block target creature this turn do so". A data record pinned
// to that creature (ADR 0041 phase 3), so it survives undo and a
// restore point, and a creature that leaves and returns is a new
// object it no longer covers (CR 400.7).
type BlockRequirementUntilEOT struct {
	Target uuid.UUID
	Kind   game.BlockRequirementKind
	Label  string
}

// Apply registers the record. A zero Target registers nothing.
func (b BlockRequirementUntilEOT) Apply(ctx *Context) error {
	if b.Target == uuid.Nil {
		return nil
	}
	if b.Kind == game.BlockRequirementBlocksAttacker {
		panic("effects: a blocksAttacker requirement names an attacking object; use BlocksAttackerUntilEOT")
	}
	return ScopedEffectFor{
		Target:   b.Target,
		Mods:     []game.Mod{game.AddBlockRequirementMod(b.Kind)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    eotLabel(b.Label, "block requirement: "+string(b.Kind)),
	}.Apply(ctx)
}

// BlocksAttackerUntilEOT is "<Blocker> blocks <Attacker> this turn if
// able" (#1684) — Provoke's "untap and block it if able", Grappling
// Hook's and Turntimber Basilisk's "target creature blocks it this
// turn if able". A data record pinned to the would-be blocker, naming
// the attacking OBJECT (instance and epoch), so an attacker that left
// and came back is not it and the requirement asks nothing (CR 400.7).
// A requirement the blocker can't legally obey — a flyer it can't
// reach, a menace attacker it can't block alone — asks nothing either
// (CR 509.1c: a requirement never beats a restriction).
type BlocksAttackerUntilEOT struct {
	Blocker  uuid.UUID
	Attacker game.ObjectRef
	Label    string
}

// Apply registers the record. A zero Blocker or Attacker registers
// nothing.
func (b BlocksAttackerUntilEOT) Apply(ctx *Context) error {
	if b.Blocker == uuid.Nil || b.Attacker.ID == uuid.Nil {
		return nil
	}
	return ScopedEffectFor{
		Target:   b.Blocker,
		Mods:     []game.Mod{game.BlocksAttackerMod(b.Attacker)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    eotLabel(b.Label, "block requirement: blocks an attacker"),
	}.Apply(ctx)
}
