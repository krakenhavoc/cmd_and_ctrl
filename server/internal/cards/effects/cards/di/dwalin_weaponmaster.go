package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwalin, Weaponmaster — Legendary Creature — Dwarf Warrior {1}{R/W},
// 2/1 (#4500, The Hobbit):
//
//	"First strike
//	 Whenever Dwalin enters or attacks, put a hone counter on each
//	 Equipment you control. (Each hone counter on an Equipment grants
//	 +1/+0 to equipped creature.)"
//
// One ability with two trigger conditions (WhenThisEntersOrAttacks).
// The Equipment are found as the trigger resolves, so one that arrived
// while it waited gets a counter and one that left does not, and an
// Equipment that is a creature right now (a reconfigured or animated one)
// is still an Equipment. The counters' +1/+0 for the equipped creature is
// the engine's (game/hone_counters.go, CR 122.1j).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cee583b7-7cc3-40ea-a227-b760839ec291",
		Name:            "Dwalin, Weaponmaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Dwalin, Weaponmaster — put a hone counter on each Equipment you control",
				dwalinHoneEachEquipment),
		},
	})
}

// dwalinHoneEachEquipment puts one hone counter on each Equipment the
// trigger's controller controls.
func dwalinHoneEachEquipment(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != ctx.Controller() || !c.HasSubtype("Equipment") {
			continue
		}
		if err := (effects.AddCounter{Target: c.InstanceID, Kind: game.CounterHone, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
