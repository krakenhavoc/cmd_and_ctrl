package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Storm-Kiln Artist — Creature — Dwarf Shaman {3}{R}, 2/2 (EDHREC
// rank 203):
//
//	"This creature gets +1/+0 for each artifact you control.
//	 Magecraft — Whenever you cast or copy an instant or sorcery
//	 spell, create a Treasure token."
//
// The spellslinger deck's mana engine: every cantrip refunds itself
// with a Treasure, and every Treasure makes the Artist bigger.
//
// Two abilities, two mechanisms:
//
//   - The pump is a Layer 7c static that applies to the Artist
//     itself and reads the board on every recompute. It counts
//     post-layer types (an animated or Lattice-made artifact
//     counts), and it counts Treasures — including the ones the
//     Artist's own trigger makes, which is the loop the card is
//     played for.
//   - Magecraft is a cast trigger of Beast Whisperer's shape, gated
//     on the spell's type read off the stack.
//
// The "or copy" half of magecraft rides EventSpellCopied
// (createSpellCopyLocked), so a copy of an instant or sorcery you
// control makes a Treasure too (CR 207.2c). No simplification.
func init() {
	Register(Spec{
		OracleID:     "a145ff8c-5812-4bcb-bd16-9839dc25121d",
		Name:         "Storm-Kiln Artist",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			Label:    "Gets +1/+0 for each artifact you control.",
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += artifactsControlledBy(g, source)
			},
		}},
		Triggered: []game.TriggeredAbility{
			OnAny(magecraftEvents, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return instantOrSorceryCastOrCopiedByYou(ev, source, g)
			}, "Storm-Kiln Artist — create a Treasure", Do(CreateToken{
				Template: TreasureToken(),
				N:        1,
			})),
		},
	})
}

// artifactsControlledBy counts the artifacts source's controller
// controls, reading each permanent's post-layer types so a layer 4
// type-add composes (Layer 7 runs after Layer 4 in the recompute
// pass, so the effective type set is settled by the time this
// reads it). Runs inside a layer recompute, so it walks the live
// battlefield slice rather than taking a copy.
func artifactsControlledBy(g *game.Game, source *game.Card) int {
	n := 0
	for i := range g.Battlefield.Cards {
		other := &g.Battlefield.Cards[i]
		if other.Controller != source.Controller {
			continue
		}
		for _, t := range other.Effective().Types {
			if t == "Artifact" {
				n++
				break
			}
		}
	}
	return n
}
