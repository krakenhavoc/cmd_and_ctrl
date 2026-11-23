package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Norika Yamazaki, the Poet — Legendary Creature — Human Samurai {2}{W},
// 3/2 (Kamigawa: Neon Dynasty):
//
//	"Vigilance
//	 Whenever a Samurai or Warrior you control attacks alone, you may cast
//	 target enchantment card from your graveyard this turn."
//
// The trigger is Zul Ashur's permission on a targeted card: a cast-only
// grant on that card for this turn, under the controller (CR 601.2); the
// spell's costs are paid as usual and nothing exiles it afterwards. "Alone"
// is the only creature declared as an attacker (CR 506.5), and Samurai and
// Warrior are read through the source's text.
//
// No simplification.
func init() {
	const name = "Norika Yamazaki, the Poet"
	effects.Register(effects.Spec{
		OracleID:        "6f143c8f-cb5c-4abc-a01b-b6813f5a524c",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventAttack, effects.NeoB09SamuraiOrWarriorAttacksAlone,
					name+" — you may cast target enchantment card from your graveyard this turn",
					effects.NeoB09MayCastFromGraveyardThisTurn(name)),
				effects.TargetCardInGraveyard("target enchantment card from your graveyard", effects.YouOwn(), effects.Enchantment())),
		},
	})
}
