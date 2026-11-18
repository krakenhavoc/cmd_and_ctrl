package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "a4d67f2d-6b0f-4feb-bf7d-90ab4f2e46ab",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventAttack, neoB09SamuraiOrWarriorAttacksAlone,
					name+" — you may cast target artifact card from your graveyard this turn",
					neoB09MayCastFromGraveyardThisTurn(name)),
				TargetCardInGraveyard("target artifact card from your graveyard", YouOwn(), Artifact())),
		},
	})
}
