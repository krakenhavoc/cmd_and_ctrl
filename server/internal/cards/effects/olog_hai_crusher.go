package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Olog-hai Crusher — Creature — Troll Soldier {3}{R} (The Hobbit Commander), 4/4:
//
//	"Trample
//	 This creature can't block unless you control a Goblin or Orc."
//
// #4454, S76. A CR 509.1b block rule on the creature itself
// (CantBlockUnless, as Goblin Goon does): the condition is asked of its
// controller as blockers are declared, with effective subtypes, so a
// changeling or a creature made a Goblin by an effect counts. The Crusher
// is a Troll, so it never satisfies its own clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fb6f04bc-3514-49b2-9d17-f80980d996a8",
		Name:            "Olog-hai Crusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		BlockRules: []game.BlockRule{CantBlockUnless(hobB3YouControlAGoblinOrOrc,
			"it can't block unless its controller controls a Goblin or Orc")},
	})
}

// hobB3YouControlAGoblinOrOrc is "you control a Goblin or Orc".
func hobB3YouControlAGoblinOrOrc(g *game.Game, you uuid.UUID) bool {
	return g.CountMatchingForEffect(you, []game.PermanentQuery{{Subtypes: []string{"Goblin", "Orc"}}}) > 0
}
