package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gimli of the Glittering Caves — Legendary Creature — Dwarf Warrior
// {2}{R}, 1/1 (EDHREC rank 3141):
//
//	"Double strike
//	 Whenever another legendary creature you control enters, put a
//	 +1/+1 counter on Gimli.
//	 Whenever Gimli deals combat damage to a player, create a
//	 Treasure token."
//
// The legends deck's double-striker. The keyword rides
// PrintedKeywords; the first trigger is Champion of the Perished's
// shape for legendary creatures (b29AnotherLegendaryCreatureYouControlEntered
// — effective supertypes, so a token copy of a legend counts), the
// counter landing on Gimli if he is still on the battlefield; the
// second is the combat-damage-to-a-player condition on Gimli
// himself, once per damage step — first-strike and regular damage
// are two Treasures, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f6747295-d21d-40d7-bc70-baa4a37ae668",
		Name:            "Gimli of the Glittering Caves",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B29AnotherLegendaryCreatureYouControlEntered(ev, source, g)
			}, "Gimli of the Glittering Caves — put a +1/+1 counter on Gimli", func(g *game.Game, item *game.StackItem) error {
				if !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(effects.NewContext(g, item))
			}),
			effects.WheneverThisDealsCombatDamageToAPlayer("Gimli of the Glittering Caves — create a Treasure", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
