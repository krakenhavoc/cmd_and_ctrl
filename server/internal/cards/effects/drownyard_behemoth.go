package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drownyard Behemoth — Creature — Eldrazi Crab {9}, 5/7:
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Emerge {7}{U} (You may cast this spell by sacrificing a creature and
//	 paying the emerge cost reduced by that creature's mana value.)
//	 This creature has hexproof as long as it entered this turn."
//
// Emerge is the shared alternative cost (ADR 0135 §4). The hexproof is a
// layer 6 grant to itself whose condition reads the turn's entry tally
// (Game.EnteredThisTurn), not summoning sickness: a Behemoth flashed in on
// an opponent's end step has hexproof for the rest of that turn and loses
// it as the next turn begins, while it is still summoning sick. Both ends
// invalidate the layer cache: the entry is a battlefield zone move, and
// the turn boundary bumps it (rotation.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "bfe1a52a-f2f8-4c62-8502-037f2d1395b2",
		Name:             "Drownyard Behemoth",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flash"},
		AlternativeCosts: []game.AlternativeCost{Emerge("{7}{U}")},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && g.EnteredThisTurn(source.InstanceID)
			}, "hexproof"),
		},
	})
}
