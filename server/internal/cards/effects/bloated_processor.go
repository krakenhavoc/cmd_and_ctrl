package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloated Processor — Creature — Phyrexian {2}{B}, 3/2 (March of the
// Machine):
//
//	"Sacrifice another Phyrexian: Put a +1/+1 counter on this creature.
//	 When this creature dies, incubate X, where X is its power."
//
// The dies trigger reads the creature's last-known power
// (incubateItsPower), so counters it grew with count. The sacrifice is a
// cost, paid at announce.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1a26310-8c14-424b-8a7f-36793769c2a7",
		Name:         "Bloated Processor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice another Phyrexian: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    SacrificeAnotherN(1, "another Phyrexian", HasSubtype("Phyrexian")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return AddCounter{Target: ctx.Source(), Kind: "+1/+1", N: 1}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Bloated Processor — incubate X, where X is its power", incubateItsPower),
		},
	})
}
