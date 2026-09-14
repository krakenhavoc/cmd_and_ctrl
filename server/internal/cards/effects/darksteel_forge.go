package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darksteel Forge — Artifact {9}:
//
//	"Artifacts you control have indestructible."
//
// The artifact half of Avacyn, and the one grant-shaped
// indestructible card S25 left behind: #380 shipped the keyword and
// Avacyn's "other permanents you control have indestructible", and
// this is the same static narrowed to artifacts.
//
// Two details that are the card rather than the implementation:
//
//   - The Forge is itself an artifact you control, so the static
//     covers it. That is printed — the text says "artifacts you
//     control", not "other artifacts" — and it is the whole reason a
//     nine-mana do-nothing rock is a lock piece rather than a
//     liability: the sweeper that would turn the lock off cannot.
//   - `target.IsArtifact()` reads the PRINTED type line, so a
//     creature animated into an artifact by a Layer 4 effect does
//     not pick the grant up. That is the narrower reading and the
//     one this repo errs toward; every other keyword-granting static
//     in the catalog (Garruk's Uprising, Lord of Atlantis, The
//     Wandering Rescuer) reads printed types the same way, and one
//     card is the wrong place to change the convention.
//
// No simplification: indestructible is enforced in the destruction
// path (server/internal/game/indestructible.go) — the single-target
// verb and the damage SBAs as of S25, and the mass-destroy path that
// every board wipe reaches as of S30 (#470 / #446). The second half
// is the one this card is bought for: a nine-mana lock piece that
// the sweeper could still turn off was not a lock.
func init() {
	Register(Spec{
		OracleID:     "9b3bec05-441f-4fdf-8b51-69fa8613fcd4",
		Name:         "Darksteel Forge",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsArtifact() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !keywordSliceContains(c.Abilities, "indestructible") {
					c.Abilities = append(c.Abilities, "indestructible")
				}
			},
		}},
	})
}
