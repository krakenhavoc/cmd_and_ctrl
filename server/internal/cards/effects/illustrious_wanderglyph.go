package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Illustrious Wanderglyph — Artifact Creature — Golem {4}{W}, 2/2
// (EDHREC rank 2100):
//
//	"Ascend (If you control ten or more permanents, you get the
//	 city's blessing for the rest of the game.)
//	 Other artifact creatures you control get +2/+2 as long as you
//	 have the city's blessing.
//	 At the beginning of each upkeep, create a 1/1 colorless Gnome
//	 artifact creature token."
//
// Tendershoot Dryad for artifact creatures: a Gnome every upkeep —
// every player's, as printed — and a +2/+2 lord once the board is
// wide, which the Gnomes themselves get it to. The token is
// Threefold Thunderhulk's Gnome (b17GnomeToken) on an upkeep trigger
// with no actor test; the anthem is a layer 7c static over OTHER
// artifact creatures the controller controls (post-layer types, so
// an animated artifact and every Gnome count).
//
// The city's blessing is the player designation ascend gives and keeps
// (CR 702.131, #2696): the anthem reads it, so a board that shrinks
// below ten permanents keeps the bonus, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "72adf091-4491-4afe-b893-b8d20ad38a86",
		Name:            "Illustrious Wanderglyph",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Illustrious Wanderglyph — create a 1/1 Gnome", Do(CreateToken{Template: TokenCard("1/1 colorless Gnome artifact"), N: 1})),
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: WhileCitysBlessing(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
					target.IsArtifact() && target.IsCreature()
			}),
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 2
			},
		}},
	})
}
