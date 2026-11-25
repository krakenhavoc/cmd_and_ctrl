package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grotag Bug-Catcher — Creature — Goblin Warrior {1}{R}, 1/2 (Zendikar
// Rising, #5508):
//
//	"Trample
//	 Whenever this creature attacks, it gets +1/+0 until end of turn for
//	 each creature in your party. (Your party consists of up to one each of
//	 Cleric, Rogue, Warrior, and Wizard.)"
//
// The party (CR 700.8) is counted as the trigger resolves, the Bug-Catcher
// itself being the Warrior if nothing else is. An ability whose source has
// left the battlefield does nothing to it (CR 400.7).
//
// No simplification.
func init() {
	const label = "Grotag Bug-Catcher — it gets +1/+0 until end of turn for each creature in your party"
	effects.Register(effects.Spec{
		OracleID:        "1b6bfd22-2964-4274-a566-81964348b003",
		Name:            "Grotag Bug-Catcher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if !effects.OnBattlefield(g, item.SourceCardID) || g.AbilitySourceIsNewObjectForEffect(item) {
					return nil
				}
				x := effects.FinB9PartySize(g, item.Controller)
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Label: label}.Apply(ctx)
			}),
		},
	})
}
