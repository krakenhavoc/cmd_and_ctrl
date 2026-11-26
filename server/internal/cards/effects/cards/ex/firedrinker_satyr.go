package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firedrinker Satyr — Creature — Satyr Shaman {R}, 2/1 (Theros):
//
//	"Whenever this creature is dealt damage, it deals that much damage to you.
//	 {1}{R}: This creature gets +1/+0 until end of turn and deals 1 damage to
//	 you."
//
// The trigger waits for the damage to finish and reflects the TOTAL the Satyr
// was dealt (CR 120.3, CR 510.2), from the Satyr as it last existed if it died
// to it (CR 608.2h). The activated ability pumps the Satyr and has it deal 1
// damage to its controller; nothing happens if it has left the battlefield by
// resolution (a new object is not "this creature", CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb107b33-ae2e-4d45-b537-1e6e25c66ec8",
		Name:         "Firedrinker Satyr",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Firedrinker Satyr — it deals that much damage to you", effects.FiredrinkerSatyrReflect),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn and deals 1 damage to you.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Firedrinker Satyr — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.DamageController{N: 1}.Apply(ctx)
			},
		}},
	})
}
