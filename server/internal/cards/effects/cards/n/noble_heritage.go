package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Noble Heritage — Legendary Enchantment — Background {1}{W} (Commander
// Legends: Battle for Baldur's Gate, S82 GAP2 b11, #6098):
//
//	"Commander creatures you own have "When this creature enters and at
//	 the beginning of your upkeep, each player may put two +1/+1 counters
//	 on a creature they control. For each opponent who does, you gain
//	 protection from that player until your next turn.""
//
// Haunted One's layer-6 grant (ADR 0093) to each commander creature you
// own, under anyone's control. The granted ability is one ability on two
// events (Wildfire Devils' ThisEntersOrYourUpkeep), "your" being the
// creature's controller. The fan-out is Orzhov Advokist's: each player in
// turn order may choose a creature they control (a choice, not a target),
// and the counters go on together. Each opponent whose creature really got
// them gives the trigger's controller protection from that player (CR
// 702.16k) until that player's next turn (CR 611.2b), Eon Frolicker's
// player keyword.
//
// No simplification.
const nobleHeritageGrant = "noble-heritage/counters-for-protection"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e342aa3-2590-4b59-852a-9adb31bafe5c",
		Name:         "Noble Heritage",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: nobleHeritageGrant,
			Triggered: []game.TriggeredAbility{
				effects.OnAny([]game.EventKind{game.EventETB, game.EventBeginUpkeep}, effects.ThisEntersOrYourUpkeep,
					"Noble Heritage — each player may put two +1/+1 counters on a creature they control; you gain protection from each opponent who does",
					func(g *game.Game, item *game.StackItem) error {
						return effects.Cmdr2S34EachPlayerMayPutCounters(effects.NewContext(g, item), effects.Cmdr2S34RiderProtection)
					}),
			},
			Text: "When this creature enters and at the beginning of your upkeep, each player may put two +1/+1 counters on a creature they control. For each opponent who does, you gain protection from that player until your next turn.",
		}},
		Static: []game.StaticAbility{
			effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.IsCommander && target.Owner == source.Controller
			}, nobleHeritageGrant),
		},
	})
}
