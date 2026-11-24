package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cleaving Reaper — Creature — Angel Berserker {3}{B}{B}, 5/3 (Kaldheim):
//
//	"Flying, trample
//	 Pay 3 life: Return this card from your graveyard to your hand. Activate
//	 only if you had an Angel or Berserker enter the battlefield under your
//	 control this turn."
//
// An activated ability of the card in the graveyard (CR 602.2) whose life is a
// cost; the gate reads the turn's entries (Game.EnteredWithSubtypeThisTurn:
// each permanent's types as it entered, a changeling counts) and so counts a
// creature that entered before the Reaper died. The card is re-checked as the
// ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "70c268af-3fb4-42b8-a87a-d06d33c1c641",
		Name:            "Cleaving Reaper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 3 life: Return this card from your graveyard to your hand. Activate only if you had an Angel or Berserker enter the battlefield under your control this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.PayLife(3),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Condition: func(g *game.Game, controller, source uuid.UUID) bool {
				src, ok := g.LookupCardForEffect(source)
				if !ok {
					return false
				}
				return g.EnteredWithSubtypeThisTurn(controller, src.TextSubtype("Angel")) > 0 ||
					g.EnteredWithSubtypeThisTurn(controller, src.TextSubtype("Berserker")) > 0
			},
			Effect: effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
