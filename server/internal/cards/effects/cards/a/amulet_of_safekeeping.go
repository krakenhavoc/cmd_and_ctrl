package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Amulet of Safekeeping — Artifact {2} (Core Set 2019):
//
//	"Whenever you become the target of a spell or ability an opponent controls,
//	 counter that spell or ability unless its controller pays {1}.
//	 Creature tokens get -1/-0."
//
// The trigger is Frost Titan's ward-shaped body (WardEffect) watching for a
// PLAYER target: the targeting event names the Amulet's controller and no card,
// from a spell or ability an opponent controls. The static is a layer 7c -1/-0
// on every creature token, whoever controls it, re-read on every recompute so a
// token made later is shrunk as it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98709d61-4da4-4d3a-a654-bb2a7a0edb6b",
		Name:         "Amulet of Safekeeping",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Creature tokens get -1/-0.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsToken() && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power--
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBecomesTarget},
			Key:     "Amulet of Safekeeping — counter that spell or ability unless its controller pays {1}",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == uuid.Nil && ev.Target == source.Controller &&
					ev.StackItemID != uuid.Nil && ev.Actor != uuid.Nil && ev.Actor != source.Controller
			},
			Effect: effects.WardEffect(effects.WardMana("{1}")),
		}},
	})
}
