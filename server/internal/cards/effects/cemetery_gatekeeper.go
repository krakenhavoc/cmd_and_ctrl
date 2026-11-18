package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "fa450433-85b0-4f95-bd4e-f141ba036bee",
		Name:            "Cemetery Gatekeeper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(cemeteryGatekeeperExileLabel, func(g *game.Game, item *game.StackItem) error {
				vowB16QueueExile(g, item, cemeteryGatekeeperExileLabel, uuid.Nil, cemeteryGatekeeperExiled)
				return nil
			}),
			InterveningIf(vowB16PlaySharesStillHolds(cemeteryGatekeeperExileLabel),
				OnAny(vowB16PlayEvents, vowB16PlaysSharingAType(cemeteryGatekeeperExileLabel, nil),
					"Cemetery Gatekeeper — this creature deals 2 damage to that player",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						return DealDamage{Source: ctx.Source(), Target: ctx.Trigger().Event.Actor, Amount: 2}.Apply(ctx)
					})),
		},
	})
}

const cemeteryGatekeeperExileLabel = "Cemetery Gatekeeper — exile a card from a graveyard"

var cemeteryGatekeeperExiled = vowB16LinkedExileThen("choose-cards/cemetery-gatekeeper-exile")
