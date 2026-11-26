package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glade Watcher — Creature — Elemental {1}{G}, 3/3 (Dragons of Tarkir, S82 older
// sets):
//
//	"Defender
//	 Formidable — {G}: This creature can attack this turn as though it didn't have
//	 defender. Activate only if creatures you control have total power 8 or
//	 greater."
//
// Assault Formation's waiver (CR 702.3b) pinned to the Watcher until end of turn:
// the keyword stays and the waiver is read at declaration. The formidable total is
// read as the ability is announced (CR 602.5b), and the waiver does nothing to a
// Watcher that is no longer the permanent that activated it (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f696b727-083e-4825-80d9-4bac35289958",
		Name:            "Glade Watcher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{effects.OldB324FormidableRow(effects.ActivatedAbility{
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.WaiveDefenderUntilEOT(effects.NewContext(g, item), item.SourceCardID,
					"Glade Watcher — can attack this turn as though it didn't have defender")
			},
		}, "{G}: This creature can attack this turn as though it didn't have defender.")},
	})
}
