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
// Sandbox simplification: the "or COPY" half of magecraft is not
// implemented. The engine has no spell-copy event (nothing in the
// catalog copies a spell), so a copied instant makes no Treasure.
// Weaker than printed until spell copying exists.
func init() {
	Register(Spec{
		OracleID:     "a145ff8c-5812-4bcb-bd16-9839dc25121d",
		Name:         "Storm-Kiln Artist",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Copied instants and sorceries make no Treasure; only spells you actually cast trigger it."},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += artifactsControlledBy(g, source)
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (spell.IsInstant() || spell.IsSorcery())
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Storm-Kiln Artist — create a Treasure",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   TreasureToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
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
