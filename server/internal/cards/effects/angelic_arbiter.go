package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angelic Arbiter — Creature — Angel {5}{W}{W}, 5/6 (Jumpstart Selesnya
// deck request, #4893):
//
//	"Flying
//	 Each opponent who cast a spell this turn can't attack with creatures.
//	 Each opponent who attacked with a creature this turn can't cast
//	 spells."
//
// Two statics, each one opponent at a time. The first is a layer-6
// CantAttack on the creatures of every opponent whose cast tally this turn
// is not zero, declared DependsOnSpellsCast so the cast that bumps the
// tally refreshes the layers (Stoic Sphinx's seam). The second is a cast
// restriction over this turn's declared attackers (the Sandswirl
// Wanderglyph shape, attacking anything, not only you). Both bind
// opponents only: the Arbiter's controller may cast and attack freely,
// and an opponent under one half can still do the other until they have
// done both. Restrictions are judged as attackers are declared and as a
// spell is announced, so neither undoes an attack or a spell already made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "34dd789c-4d80-4ccc-a4f7-2e5388a9689d",
		Name:            "Angelic Arbiter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{OpponentsWhoCastASpellThisTurnCantAttack()},
		CastRestrictions: []game.CastRestriction{
			OpponentsWhoAttackedWithACreatureThisTurnCantCast("Each opponent who attacked with a creature this turn can't cast spells."),
		},
	})
}
