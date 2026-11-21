package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Netherborn Altar — Artifact {1}{B} (Commander 2020, #5513):
//
//	"{T}, Put a soul counter on this artifact: Put your commander into your
//	 hand from the command zone. Then you lose 3 life for each soul counter
//	 on this artifact."
//
// The soul counter is a cost (Mazemind Tome's AddCounterToThis), paid as the
// ability is activated (CR 602.2b), so the first activation counts it. The
// commander move is Road of Return's: a commander you own in your command
// zone, asked for when there are two. The life loss counts the soul counters
// as the ability resolves, off the Altar as it last existed if it has left
// (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a127d62a-8709-46bb-9301-96f09e7924a1",
		Name:         "Netherborn Altar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Put a soul counter on this artifact: Put your commander into your hand from the command zone. Then you lose 3 life for each soul counter on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.AddCounterToThis(C20SoulCounter, 1)),
			Effect:  netherbornAltarResolve,
		}},
	})
}

const C20SoulCounter = "soul"

func netherbornAltarResolve(g *game.Game, item *game.StackItem) error {
	if err := effects.PutCommanderIntoHand(g, item, "Netherborn Altar", true); err != nil {
		return err
	}
	n := c20CountersOnSourceOrLastKnown(g, item.SourceCardID, C20SoulCounter)
	return effects.GainLife{Amount: -3 * n}.Apply(effects.NewContext(g, item))
}

// c20CountersOnSourceOrLastKnown is the number of `kind` counters on the
// permanent now, or as it last existed on the battlefield once it has left.
func c20CountersOnSourceOrLastKnown(g *game.Game, id uuid.UUID, kind string) int {
	if effects.OnBattlefield(g, id) {
		return effects.B39CountersOn(g, id, kind)
	}
	return g.LastKnownCountersForEffect(id)[kind]
}
