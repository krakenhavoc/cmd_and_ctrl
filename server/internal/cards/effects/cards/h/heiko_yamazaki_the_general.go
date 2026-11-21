package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heiko Yamazaki, the General — Legendary Creature — Human Samurai {3}{R},
// 3/3 (Kamigawa: Neon Dynasty):
//
//	"Trample
//	 Whenever a Samurai or Warrior you control attacks alone, you may cast
//	 target artifact card from your graveyard this turn."
//
// Norika Yamazaki's trigger over artifact cards: a cast-only grant on the
// targeted card for this turn (CR 601.2).
//
// No simplification.
func init() {
	const name = "Heiko Yamazaki, the General"
	effects.Register(effects.Spec{
		OracleID:        "a4d67f2d-6b0f-4feb-bf7d-90ab4f2e46ab",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventAttack, effects.NeoB09SamuraiOrWarriorAttacksAlone,
					name+" — you may cast target artifact card from your graveyard this turn",
					effects.NeoB09MayCastFromGraveyardThisTurn(name)),
				effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.YouOwn(), effects.Artifact())),
		},
	})
}
