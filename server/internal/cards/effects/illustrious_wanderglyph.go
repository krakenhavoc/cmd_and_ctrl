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
// SANDBOX GAP, weaker than printed (Tendershoot Dryad's, the same
// words): the city's blessing is a player designation that, once
// earned, lasts the rest of the game, and the engine has no
// per-player designation to keep it in. The anthem therefore reads
// the condition live — "as long as you control ten or more
// permanents" — so a board that shrinks below ten loses the bonus
// where the printed card would keep it. Never stronger: the
// condition that earns the blessing is the same one.
func init() {
	Register(Spec{
		OracleID:     "72adf091-4491-4afe-b893-b8d20ad38a86",
		Name:         "Illustrious Wanderglyph",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The city's blessing isn't kept once earned — your other artifact creatures get +2/+2 only while you control ten or more permanents."},
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Illustrious Wanderglyph — create a 1/1 Gnome", Do(CreateToken{Template: TokenCard("1/1 colorless Gnome artifact"), N: 1})),
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
					target.IsArtifact() && target.IsCreature() &&
					b11PermanentsControlled(g, source.Controller) >= 10
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 2
			},
		}},
	})
}
