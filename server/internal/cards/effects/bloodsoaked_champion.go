package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodsoaked Champion — Creature — Human Warrior {B}, 2/1:
//
//	"This creature can't block.
//	 Raid — {1}{B}: Return this card from your graveyard to the
//	 battlefield. Activate only if you attacked this turn."
//
// The can't-block restriction is a static on the creature itself. The
// return is an activated ability of the card in the graveyard (CR 602.2),
// gated on having declared an attacker this turn (the turn tally's
// attacks, ability word "raid"); it is read again as the ability
// resolves that the card is still in the graveyard (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cad09970-14c8-4d80-82fe-6c855efb0191",
		Name:         "Bloodsoaked Champion",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Activated: []ActivatedAbility{{
			Label:   "Raid — {1}{B}: Return this card from your graveyard to the battlefield. Activate only if you attacked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{1}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return b18AttackedThisTurn(g, controller)
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return returnThisFromGraveyardToBattlefield(g, item)
			},
		}},
	})
}
