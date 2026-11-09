package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Takenuma — Plane — Kamigawa (Planechase Anthology, #4870):
//
//	"Whenever a creature leaves the battlefield, its controller draws a
//	 card.
//	 Whenever chaos ensues, return target creature you control to its
//	 owner's hand."
//
// A plane (ADR 0216 §6). The first ability reads the creature's last
// known controller off the leaves-the-battlefield event (CR 603.10a), so a
// token and a stolen creature draw for the right player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de53c948-b371-4f79-a965-33c44d0b44de",
		Name:         "Takenuma",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, aCreatureLeftTheBattlefield, "Takenuma — its controller draws a card",
				func(g *game.Game, item *game.StackItem) error {
					who := item.Trigger.Event.LastKnownController
					if who == uuid.Nil {
						return nil
					}
					return DrawCards{Player: who, N: 1}.Apply(NewContext(g, item))
				}),
			Targeting(WheneverChaosEnsues("Takenuma — return target creature you control to its owner's hand",
				bounceFirstLegalTargetCard),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

func aCreatureLeftTheBattlefield(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB {
		return false
	}
	if was, known := ev.WasType("creature"); known {
		return was
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && leftAsType(ev, c, "creature")
}
