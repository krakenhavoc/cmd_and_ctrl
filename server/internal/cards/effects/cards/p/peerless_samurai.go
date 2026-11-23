package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peerless Samurai — Creature — Human Samurai {2}{R}, 2/3 (Kamigawa: Neon
// Dynasty):
//
//	"Menace
//	 Whenever a Samurai or Warrior you control attacks alone, the next
//	 spell you cast this turn costs {1} less to cast."
//
// Hardened Berserker's next-spell promise on the lone attack (CR 506.5).
//
// No simplification.
func init() {
	const name = "Peerless Samurai"
	effects.Register(effects.Spec{
		OracleID:        "fe3b0393-dc50-4fa6-9b3f-1175a6a825e4",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.NeoB09SamuraiOrWarriorAttacksAlone,
				name+" — the next spell you cast this turn costs {1} less", peerlessSamuraiPromise),
		},
	})
}

func peerlessSamuraiPromise(g *game.Game, item *game.StackItem) error {
	return effects.GrantNextSpellPromise{From: "Peerless Samurai", Promise: game.NextSpellPromise{
		Reduce: 1,
		Text:   "The next spell you cast this turn costs {1} less to cast.",
	}}.Apply(effects.NewContext(g, item))
}
