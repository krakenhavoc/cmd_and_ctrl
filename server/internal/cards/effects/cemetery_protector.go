package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cemetery Protector — Creature — Human Soldier {2}{W}{W}, 3/4 (Innistrad:
// Crimson Vow):
//
//	"Flash
//	 When this creature enters, exile a card from a graveyard.
//	 Whenever you play a land or cast a spell, if it shares a card type
//	 with the exiled card, create a 1/1 white Human creature token."
//
// The exile is not targeted: the controller picks any card from any
// graveyard as the trigger resolves, and it is exiled linked to that trigger
// (CR 607.2a), so it is "the exiled card" of this object only. The second
// ability is a CR 603.4 intervening "if": checked as the land is played or
// the spell cast, and again as the trigger resolves. Playing a land and
// casting a spell are both read (a land played is not cast; a copy of a
// spell is not cast either).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "34c8fffd-1b23-4f40-ad29-1f0acb519e7e",
		Name:            "Cemetery Protector",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(cemeteryProtectorExileLabel, func(g *game.Game, item *game.StackItem) error {
				vowB16QueueExile(g, item, cemeteryProtectorExileLabel, uuid.Nil, cemeteryProtectorExiled)
				return nil
			}),
			InterveningIf(vowB16PlaySharesStillHolds(cemeteryProtectorExileLabel),
				OnAny(vowB16PlayEvents, vowB16PlaysSharingAType(cemeteryProtectorExileLabel, vowB16YouPlay),
					"Cemetery Protector — create a 1/1 white Human creature token",
					Do(CreateToken{Template: TokenCard("1/1 white Human"), N: 1}))),
		},
	})
}

const cemeteryProtectorExileLabel = "Cemetery Protector — exile a card from a graveyard"

var cemeteryProtectorExiled = vowB16LinkedExileThen("choose-cards/cemetery-protector-exile")
