package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anhelo, the Painter — Legendary Creature — Vampire Assassin {U}{B}{R},
// 1/3 (Streets of New Capenna Commander, #5491):
//
//	"Deathtouch
//	 The first instant or sorcery spell you cast each turn has casualty 2.
//	 (As you cast that spell, you may sacrifice a creature with power 2 or
//	 greater. When you do, copy the spell and you may choose new targets
//	 for the copy.)"
//
// Silverquill, the Disputant's granted casualty with one more condition on
// the spell: it is the first instant or sorcery its caster has cast this
// turn. The grant is read as the spell is announced (CR 601.2b), before the
// cast is tallied, so "first" is an instant-and-sorcery tally of zero
// (Arcane Bombardment's per-turn tally, read one step earlier). An
// Adventure's instant or sorcery half counts, a creature spell never does.
// The copy is the engine's casualty trigger, read off what was paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d3c3f293-6790-4dce-900a-8af0464c2572",
		Name:            "Anhelo, the Painter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		GrantedOptionalCosts: []game.GrantedOptionalCost{
			SpellsYouCastHave("The first instant or sorcery spell you cast each turn has casualty 2.",
				s82SncFirstInstantOrSorceryThisTurn, Casualty(2)),
		},
	})
}
