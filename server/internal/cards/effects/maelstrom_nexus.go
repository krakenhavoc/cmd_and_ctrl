package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maelstrom Nexus — Enchantment {W}{U}{B}{R}{G}:
//
//	"The first spell you cast each turn has cascade."
//
// The other shape of the keyword: a permanent that GIVES cascade
// rather than a spell that has it, so the trigger lives on the
// battlefield and needs no FromStack.
//
// "The first spell you cast each turn" reads the per-turn tally,
// which the cast path bumps BEFORE it emits EventCast — so a count of
// exactly one means "this is the first", and the second spell of the
// turn sees two and is skipped. The comment on that bump in
// mutations.go says the same thing about "first noncreature spell
// each turn" predicates; this is the same clock.
func init() {
	Register(Spec{
		OracleID:     "b1c346a4-fc9b-474a-9d51-cac168e363fa",
		Name:         "Maelstrom Nexus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			GrantsCascade("Maelstrom Nexus", func(_ game.Card, source *game.Card, g *game.Game) bool {
				return g.SpellsCastThisTurn[source.Controller].Total == 1
			}),
		},
	})
}
