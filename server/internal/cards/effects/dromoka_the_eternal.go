package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dromoka, the Eternal — Legendary Creature — Dragon {3}{G}{W}, 5/5 (Fate Reforged,
// #5552):
//
//	"Flying
//	 Whenever a Dragon you control attacks, bolster 2. (Choose a creature
//	 with the least toughness among creatures you control and put two
//	 +1/+1 counters on it.)"
//
// Utvara Hellkite's watch: a Dragon you control being declared as an
// attacker, Dromoka included (its creature type is read through the layers,
// so a changeling counts). Each such Dragon triggers once. Bolster (CR
// 701.39a) chooses the creature you control with the least toughness as the
// trigger resolves, with no target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "73520b5a-3d55-40c9-af60-90fdd811934f",
		Name:            "Dromoka, the Eternal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b14DragonYouControlAttacked(ev, source, g)
			}, "Dromoka, the Eternal — bolster 2", Do(Bolster{N: 2})),
		},
	})
}
