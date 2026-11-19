package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temperamental Oozewagg — Creature — Ooze Brushwagg {3}{G}, 4/4 (Modern
// Horizons 3):
//
//	"{2}{G}: Adapt 2.
//	 Modified creatures you control have trample. (Equipment, Auras you control,
//	 and counters are modifications.)"
func init() {
	Register(Spec{
		OracleID:     "07b0b2ad-4364-4a8d-94d5-8bf984b23d05",
		Name:         "Temperamental Oozewagg",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}: Adapt 2.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}{G}"),
			Effect:  adaptTwo,
		}},
		Static: []game.StaticAbility{Labelled(b16GrantKeywords(func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller && b07IsModified(g, *target)
		}, "trample"), "Modified creatures you control have trample.")},
	})
}
