package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wakestone Gargoyle — Creature — Gargoyle {3}{W}, 3/4 (Foundations
// Commander):
//
//	"Defender, flying
//	 {1}{W}: Creatures you control with defender can attack this turn as
//	 though they didn't have defender."
//
// A defender waiver (ADR 0106 amendment 2026-10-16, CR 702.3b, 609.4) on
// a LIVE "creatures you control" scope until end of turn: the effect
// changes no characteristic of the creatures, it lifts a rule, so CR
// 611.2c does not lock the set and a defender that arrives after the
// ability resolves is covered too (the printed ruling). The waiver is
// only ever asked of a creature that has defender, so "you control" is
// the whole filter. The keyword stays, and the Gargoyle itself is among
// the creatures.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "11284767-7095-4e0e-bc60-c69d58098348",
		Name:            "Wakestone Gargoyle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Creatures you control with defender can attack this turn as though they didn't have defender.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				g.RegisterScopedRuleEffectForEffect(item.SourceCardID, game.ScopeYourCreatures, item.Controller,
					[]game.Mod{game.WaiveDefenderMod()}, g.UntilEndOfTurnDuration(),
					"Wakestone Gargoyle — creatures you control with defender can attack this turn as though they didn't have defender")
				return nil
			},
		}},
	})
}
