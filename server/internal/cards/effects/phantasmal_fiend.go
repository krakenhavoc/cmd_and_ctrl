package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phantasmal Fiend — Creature — Illusion {3}{B}, 1/5:
//
//	"{B}: This creature gets +1/-1 until end of turn.
//	 {1}{U}: Switch this creature's power and toughness until end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
// The ability's source must still be the same object (CR 400.7). The +1/-1 is layer 7c, so a pump
// activated after the switch still applies before it: a switched 1/5
// pumped once is 4/2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "821ce349-5379-4415-89e2-224e21dda24b",
		Name:         "Phantasmal Fiend",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{B}: This creature gets +1/-1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: -1, Label: "Phantasmal Fiend — +1/-1"}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{1}{U}: Switch this creature's power and toughness until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{1}{U}"),
				Effect:  switchThisPT("Phantasmal Fiend — switch its power and toughness"),
			},
		},
	})
}
