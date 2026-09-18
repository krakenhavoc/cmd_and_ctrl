package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vindictive Vampire — 2/3 Creature — Vampire for {3}{B} (EDHREC
// rank 4407):
//
//	"Whenever another creature you control dies, this creature
//	 deals 1 damage to each opponent and you gain 1 life."
//
// Blood Artist's louder cousin: a point to EACH opponent rather than
// one, which at a four-player table is three damage a body instead of
// one. The trade is that it only sees YOUR creatures die, where the
// Artist sees the whole table's. Roadmap batch 42 (#449), "no new
// machinery".
//
// "Another creature YOU CONTROL" — so the Vampire does not see its
// own death (the "another" in CR 603.6c's terms), and does not see an
// opponent's creature die at all. AnotherCreatureDied plus the
// controller check is exactly that pair.
//
// Tokens count. The clause says "creature", not "nontoken creature",
// which is what makes this a sacrifice-deck card rather than a
// Midnight Reaper.
//
// DAMAGE, not life loss: it is preventable and it is redirectable,
// and a player with protection from black takes none. The life gain
// is a separate clause and happens whether or not any damage landed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11ad664e-36d3-4d5b-8a87-59ea715877e0",
		Name:         "Vindictive Vampire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverACreatureYouControlDies("Vindictive Vampire — 1 damage to each opponent, gain 1 life",
				func(g *game.Game, item *game.StackItem) error {
					if err := damageToEachOpponent(g, item, 1); err != nil {
						return err
					}
					return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
