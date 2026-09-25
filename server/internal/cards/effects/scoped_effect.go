package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scoped_effect.go is the card-facing half of ADR 0041 phase 3's data
// records (#1497, game/scoped_effects.go): a continuous effect from a
// resolving spell or ability, written as operations from a closed
// vocabulary rather than as a raw `game.StaticAbility`.
//
// It replaces `StaticForDuration` for every card that has moved. The
// difference a player never sees and an operator does: a
// `ScopedEffectFor` is data, so a table holding one is still a restore
// point, where a `StaticForDuration` froze the restore point for as
// long as the effect lived — for The Legend of Kyoshi's Island, the
// rest of the game.
//
//	ScopedEffectFor{
//	    Match:    And(Creature(), ControlledBy(victim)), // locked once (CR 611.2c)
//	    Mods:     game.SetBasePTMods(1, 1),
//	    Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
//	    Label:    "Mass Diminish — base power and toughness 1/1",
//	}.Apply(ctx)
//
// The mod constructors live in internal/game beside the kinds they
// build: SetControllerMod, AddTypesMod, RemoveTypesMod, AddSubtypesMod,
// AllCreatureTypesMod, SetColorsMod, AddKeywordsMod, RemoveKeywordsMod,
// LoseAllAbilitiesMod, AddRestrictionsMod, SetBasePowerMod,
// SetBaseToughnessMod, SetBasePTMods and ModifyPTMod.
type ScopedEffectFor struct {
	// Target pins the effect to one permanent. Ignored when Match is
	// set.
	Target uuid.UUID

	// Match selects the affected permanents, evaluated ONCE, now
	// (CR 611.2c). The set is locked on {instance, entry stamp}, so a
	// permanent that leaves and returns is a new object the effect no
	// longer follows (CR 400.7).
	Match CardPredicate

	// Mods are the operations, each applied in the layer its kind
	// names, all at one timestamp.
	Mods []game.Mod

	// Duration is how long the effect lasts (CR 611.2). Build it with
	// one of the Duration* builders in durations.go.
	Duration game.Duration

	// Label is attribution for logs and tests.
	Label string
}

// Apply registers the record. Nothing matched, or no mods, registers
// nothing.
//
// Caller must be inside the resolution frame (holds g.mu write).
func (s ScopedEffectFor) Apply(ctx *Context) error {
	if len(s.Mods) == 0 {
		return nil
	}
	// ADR 0093 PR 4: a grantAbilities mod must name a bundle the
	// catalog registers (duration_grants.go).
	if err := checkGrantMods(s.Mods); err != nil {
		return err
	}
	set := eotSnapshot(ctx, s.Target, s.Match)
	if set == nil {
		return nil
	}
	ctx.Game.RegisterScopedEffectForEffect(ctx.Source(), set.affectedObjects(), s.Mods,
		s.Duration, eotLabel(s.Label, "continuous effect"))
	return nil
}

// affectedObjects is the snapshot as a record's affected set, sorted by
// instance ID so the record — and so a restore point — is the same
// bytes whatever order the map iterates in.
func (s eotAffected) affectedObjects() []game.AffectedObject {
	out := make([]game.AffectedObject, 0, len(s))
	for id, stamp := range s {
		// PinObject, not a bare stamp: an unstamped permanent is pinned
		// exactly, the way appliesTo's `==` pins it, and not as the 0
		// wildcard that would follow it through a flicker (#1558).
		out = append(out, game.PinObject(id, stamp))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID.String() < out[j].ID.String() })
	return out
}
