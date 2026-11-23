package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cemetery Gatekeeper — Creature — Vampire {1}{R}, 2/1 (Innistrad: Crimson
// Vow):
//
//	"First strike
//	 When this creature enters, exile a card from a graveyard.
//	 Whenever a player plays a land or casts a spell, if it shares a card
//	 type with the exiled card, this creature deals 2 damage to that
//	 player."
//
// The exile is linked to the enters trigger (CR 607.2a). The second ability
// is a CR 603.4 intervening "if", checked as the land is played or the spell
// cast and again as the trigger resolves, and it hits whichever player did
// it, its own controller included: the player is read off the triggering
// event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fa450433-85b0-4f95-bd4e-f141ba036bee",
		Name:            "Cemetery Gatekeeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(cemeteryGatekeeperExileLabel, func(g *game.Game, item *game.StackItem) error {
				effects.VowB16QueueExile(g, item, cemeteryGatekeeperExileLabel, uuid.Nil, cemeteryGatekeeperExiled)
				return nil
			}),
			effects.InterveningIf(effects.VowB16PlaySharesStillHolds(cemeteryGatekeeperExileLabel),
				effects.OnAny(effects.VowB16PlayEvents, effects.VowB16PlaysSharingAType(cemeteryGatekeeperExileLabel, nil),
					"Cemetery Gatekeeper — this creature deals 2 damage to that player",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.DealDamage{Source: ctx.Source(), Target: ctx.Trigger().Event.Actor, Amount: 2}.Apply(ctx)
					})),
		},
	})
}

const cemeteryGatekeeperExileLabel = "Cemetery Gatekeeper — exile a card from a graveyard"

var cemeteryGatekeeperExiled = effects.VowB16LinkedExileThen("choose-cards/cemetery-gatekeeper-exile")
