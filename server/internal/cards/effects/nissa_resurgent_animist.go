package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nissa, Resurgent Animist — Legendary Creature — Elf Scout {2}{G}, 3/3
// (March of the Machine: The Aftermath):
//
//	"Landfall — Whenever a land you control enters, add one mana of any
//	 color. Then if this is the second time this ability has resolved
//	 this turn, reveal cards from the top of your library until you
//	 reveal an Elf or Elemental card. Put that card into your hand and
//	 the rest on the bottom of your library in a random order."
//
// Landfall adds the mana as the ability resolves (a pipe pick, the same
// pick a Lotus Cobra landfall queues). "The second time this ability has
// resolved this turn" is the per-object resolution tally Ms. Bumbleflower
// and Harvestrite Host read (Game.ResolvedThisTurn, CR 400.7), counting the
// resolution in progress, so a third resolution only adds mana. The reveal
// is Page, Loose Leaf's Grandeur's run (socB30RevealUntilToHandRestBottom):
// every card revealed is shown, the Elf or Elemental goes to hand and the
// rest under the library in a random order; with no hit the whole library
// goes under. "Elf" and "Elemental" are words of Nissa's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c1fc5923-c3cd-448a-98d1-c154661c2812",
		Name:         "Nissa, Resurgent Animist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall(nissaResurgentAnimistLabel, nissaResurgentAnimistResolve),
		},
	})
}

const nissaResurgentAnimistLabel = "Nissa, Resurgent Animist — landfall: add one mana of any color; the second time this turn, reveal until an Elf or Elemental card and put it into your hand"

func nissaResurgentAnimistResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (AddMana{Produced: "{W|U|B|R|G}"}).Apply(ctx); err != nil {
		return err
	}
	if g.ResolvedThisTurn(item.SourceCardID, nissaResurgentAnimistLabel) != 2 {
		return nil
	}
	elf, elemental := ctx.TextSubtype("Elf"), ctx.TextSubtype("Elemental")
	return socB30RevealUntilToHandRestBottom("Nissa, Resurgent Animist", func(c game.Card) bool {
		return c.HasSubtype(elf) || c.HasSubtype(elemental)
	})(g, item)
}
