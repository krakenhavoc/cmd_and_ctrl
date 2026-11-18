package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ragefire Hellkite — Creature — Dragon {4}{R}{R}, 5/3 (Dominaria United):
//
//	"Flying
//	 Whenever this creature attacks, you may sacrifice another creature. If
//	 you do, this creature gains double strike until end of turn."
//
// Ashnod's shape: the "you may sacrifice another creature" is asked as the
// trigger resolves over the other creatures its controller controls (the
// Hellkite is not a legal choice), and "if you do" follows only a creature
// that really went. The double strike is granted to the Hellkite itself, and
// does nothing if it has left the battlefield by then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0166996d-941f-49fe-bf4f-64b6fd4b69e9",
		Name:            "Ragefire Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Ragefire Hellkite — you may sacrifice another creature; if you do, it gains double strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Ragefire Hellkite — sacrifice another creature to give it double strike until end of turn?",
						Creature(), true, ragefireHellkiteSacrificed)
				}),
		},
	})
}

// ragefireHellkiteSacrificed is "if you do, this creature gains double
// strike until end of turn". An on-disk key: never renamed, never reused.
var ragefireHellkiteSacrificed = RegisterSacrificedOne("prompt-run/ragefire-hellkite-double-strike", func(ctx *Context, _ uuid.UUID) error {
	return GrantKeywordUntilEOT{
		Target:   ctx.Source(),
		Keywords: []string{"double strike"},
		Label:    "Ragefire Hellkite — double strike until end of turn",
	}.Apply(ctx)
})
