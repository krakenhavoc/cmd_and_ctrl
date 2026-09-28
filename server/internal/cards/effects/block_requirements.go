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
//	BlockRequirementUntilEOT{Target: t, Kind: game.BlockRequirementMustBeBlocked}.Apply(ctx) // Irresistible Prey
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
	return game.StaticAbility{
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && appliesTo(target, g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			r := game.BlockRequirement{Kind: kind}
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
	return BlockRequirementWhere(game.BlockRequirementLure, selfOnly)
}

// MustBeBlocked is "This creature must be blocked if able" (Gaea's
// Protector).
func MustBeBlocked() game.StaticAbility {
	return BlockRequirementWhere(game.BlockRequirementMustBeBlocked, selfOnly)
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
	return ScopedEffectFor{
		Target:   b.Target,
		Mods:     []game.Mod{game.AddBlockRequirementMod(b.Kind)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    eotLabel(b.Label, "block requirement: "+string(b.Kind)),
	}.Apply(ctx)
}
