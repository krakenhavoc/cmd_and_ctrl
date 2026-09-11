package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lord of Atlantis — "Other Merfolk creatures get +1/+1. Other
// Merfolk creatures have islandwalk."
//
// First S16 Layer-6 catalog card. Two static abilities on one
// permanent:
//
//   - Layer 7c: Merfolk creatures controlled by Lord's controller
//     (excluding Lord itself) get +1/+1.
//   - Layer 6: Same predicate; effective abilities gain "flying"
//     ... wait, the actual oracle is "islandwalk" only — flying is
//     a different lord. Sticking with the printed Lord of Atlantis
//     text: islandwalk only. (The plan listed "flying + islandwalk"
//     as a quick example; the actual card is islandwalk only.)
//
// Self-exclusion test: AppliesTo checks `target.InstanceID !=
// source.InstanceID` — the "other" predicate. Single Lord on the
// board with no other Merfolk shows printed 2/2, not 3/3.
//
// Behaviour of islandwalk lands with S18's combat keyword pipeline
// — S16 just exposes the keyword grant on the wire so the renderer
// + S18 SBA can consume it without a wire bump.
func init() {
	Register(Spec{
		OracleID: "cc7f290f-ca00-4285-9bdb-4b4402444f30",
		Name:     "Lord of Atlantis",
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: lordOfAtlantisOtherMerfolk,
				Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
			{
				Layer:     game.Layer6Ability,
				AppliesTo: lordOfAtlantisOtherMerfolk,
				Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
					for _, k := range c.Abilities {
						if k == "islandwalk" {
							return
						}
					}
					c.Abilities = append(c.Abilities, "islandwalk")
				},
			},
		},
	})
}

// lordOfAtlantisOtherMerfolk is the shared predicate for Lord's
// two static abilities: target is a creature, has the Merfolk
// subtype, controlled by Lord's controller, and is NOT the Lord
// itself ("other"). Reads from the post-Layer-2 controller and
// post-Layer-4 type set so type-add effects (Mycosynth Lattice +
// "all creatures are Merfolk" if such a card existed) compose
// naturally.
//
// Layered ordering note: Layer 6 (this predicate) runs after Layer
// 4 (type-add) so a hypothetical Conspiracy that turned a Bear
// into a Merfolk would correctly see the Lord's grant. Real
// Conspiracy isn't in S16 scope; documented for completeness.
func lordOfAtlantisOtherMerfolk(target *game.Card, g *game.Game, source *game.Card) bool {
	if !target.IsCreature() {
		return false
	}
	if target.Controller != source.Controller {
		return false
	}
	if target.InstanceID == source.InstanceID {
		return false
	}
	return target.HasSubtype("Merfolk")
}
