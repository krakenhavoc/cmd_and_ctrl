package cardspr

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prime Speaker Vannifar — Legendary Creature — Elf Ooze Wizard {2}{G}{U}, 2/4
// (Ravnica Allegiance):
//
//	"{T}, Sacrifice another creature: Search your library for a creature card with
//	 mana value equal to 1 plus the sacrificed creature's mana value, put that card
//	 onto the battlefield, then shuffle. Activate only as a sorcery."
//
// Birthing Pod's search at sorcery speed with a tap: the mana value is the
// sacrificed creature's as it last existed (CR 608.2h), a token's being 0. The
// sacrifice must be another creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b543bfa2-0174-4dd8-9ee6-09b92fb44934",
		Name:         "Prime Speaker Vannifar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Sacrifice another creature: Search your library for a creature card with mana value equal to 1 plus the sacrificed creature's mana value, put that card onto the battlefield, then shuffle. Activate only as a sorcery.",
			Cost: effects.Plus(effects.TapCost(),
				effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			SorcerySpeed: true,
			Effect:       vannifarSearch,
		}},
	})
}

func vannifarSearch(g *game.Game, item *game.StackItem) error {
	sacrificed, ok := effects.B17PermanentSacrificedToPay(g, item)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(sacrificed)
	if !ok {
		return nil
	}
	want := c.ManaValue() + 1
	return effects.SearchLibrary{
		Player:    item.Controller,
		Predicate: func(c game.Card) bool { return c.IsCreature() && c.ManaValue() == want },
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Source:    item.SourceCardID,
		Reason:    "Prime Speaker Vannifar — a creature card with mana value " + strconv.Itoa(want),
	}.Apply(effects.NewContext(g, item))
}
