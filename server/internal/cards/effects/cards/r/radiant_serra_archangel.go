package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Radiant, Serra Archangel — Legendary Creature — Angel {6}{W}, 6/4
// (Commander Legends, #5507):
//
//	"Flying
//	 Tap another untapped creature you control with flying: Radiant gains
//	 protection from the color of your choice until end of turn.
//	 Partner"
//
// The cost is tap-others (#758): another untapped creature with flying (read
// through the layered view), not the {T} symbol, so one that arrived this turn
// may pay it; Radiant never pays it herself. The colour is chosen as the
// ability resolves (CR 700.2a) and the protection is Bathe in Light's
// chosen-colour grant, on Radiant only and until end of turn; it does nothing
// if she has left. Partner is a deck-construction rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4a8227ff-0ae4-4405-8e44-1f5187a7f9f5",
		Name:            "Radiant, Serra Archangel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap another untapped creature you control with flying: Radiant gains protection from the color of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.TapAnotherUntapped("another untapped creature you control with flying", effects.Creature(), effects.HasKeyword("flying")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				tail := ctx.ItemTail(effects.RadianceProtectionFromChosenColor)
				tail.Carry = []uuid.UUID{item.SourceCardID}
				tail.Words = []string{"Radiant, Serra Archangel"}
				effects.ChooseColorTail(game.ColorForProtection, g, item.Controller, item.SourceCardID, "Radiant, Serra Archangel — choose a color", tail)
				return nil
			},
		}},
	})
}
