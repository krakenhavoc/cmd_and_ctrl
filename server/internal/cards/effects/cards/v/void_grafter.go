package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Void Grafter — {1}{G}{U} Creature — Eldrazi Drone 2/4 (Oath of the
// Gatewatch, S82 OLD b295):
//
//	"Devoid (This card has no color.)
//	 Flash
//	 When this creature enters, another target creature you control gains
//	 hexproof until end of turn."
//
// Ascendant Dustspeaker's "another target creature you control" enters
// trigger with a keyword granted until end of turn. The target is chosen as the
// trigger goes on the stack (CR 603.3d) and checked again as it resolves (CR
// 608.2b).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Void Grafter — another target creature you control gains hexproof until end of turn",
		effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"hexproof"}, Label: "Void Grafter — hexproof"}.Apply(ctx)
		})))
	enters.Targets = effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))
	effects.Register(effects.Spec{
		OracleID:        "01d202b9-1ac6-484c-a208-a8bc69bccdb2",
		Name:            "Void Grafter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flash"},
		Purpose:         game.Purpose{Answers: game.AnswerProtect},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
