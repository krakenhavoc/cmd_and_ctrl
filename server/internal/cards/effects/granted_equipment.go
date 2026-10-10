package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// granted_equipment.go — permanents that are made Equipment, or given
// equip, by another card (#2562, ADR 0093 amendment 2026-10-10):
// Gemcutter Buccaneer's "Treasures you control are Equipment in addition
// to their other types and have 'Equipped creature gets +2/+0' … and
// equip {3}", Puresteel Paladin's "Equipment you control have equip
// {0}". The Irencrag's own "becomes a … Equipment artifact" is a
// resolved effect, game.SetTypesMod.
//
// The type is a layer-4 static; the equip rows and the "Equipped
// creature gets …" static are bundles a layer-6 GrantAbilities names.
// An equip row granted this way is the recipient's own activated
// ability (CR 702.6a), so it attaches the recipient, and a granted
// layer-7c static is applied after layer 6 (game/granted_statics.go).
//
// Append-only, per the shared-vocabulary rule.

// AreAlsoEquipment is "<applies> are Equipment in addition to their
// other types" — a layer-4 subtype add (CR 205.1b), shown as `label`.
// It takes nothing away, and does nothing to a permanent that is
// already an Equipment. An object that is not an artifact cannot have
// an artifact type (CR 205.3d), so a non-artifact the predicate matches
// is left alone.
func AreAlsoEquipment(label string, applies func(target *game.Card, g *game.Game, source *game.Card) bool) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer4Type,
		AppliesTo: applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if !eotHasType(c.Types, "Artifact") || eotHasType(c.Subtypes, "Equipment") {
				return
			}
			c.Subtypes = append(append([]string(nil), c.Subtypes...), "Equipment")
		},
		Label: label,
	}
}

// equipmentYouControl is "Equipment you control": the source's
// controller's permanents with the Equipment subtype, read post-layer.
func equipmentYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.HasSubtype("Equipment")
}
