package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bond of Discipline — Sorcery {4}{W} (War of the Spark, #5520):
//
//	"Tap all creatures your opponents control. Creatures you control gain
//	 lifelink until end of turn."
//
// Summon: Alexander's mass tap over the creatures your opponents control as
// the spell resolves, then a lifelink grant over the creatures you control at
// that moment (CR 611.2c); one that enters later does not get it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00e9957a-5684-4a07-954e-8161925fefe4",
		Name:         "Bond of Discipline",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.S82OldB172TapAllOpponentCreaturesAndGrantLifelink,
	})
}
