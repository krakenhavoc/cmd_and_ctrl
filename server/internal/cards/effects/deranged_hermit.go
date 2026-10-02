package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deranged Hermit — Creature — Elf, {3}{G}{G}, 1/1:
//
//	"Echo {3}{G}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, create four 1/1 green Squirrel creature tokens.
//	 Squirrel creatures get +1/+1."
//
// "Squirrel creatures get +1/+1" names no controller and no "other", so it
// is TribalAnthem over every Squirrel on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9cd714b-2ad8-4fdb-a8aa-82b17730e071",
		Name:         "Deranged Hermit",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Squirrel"}}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			Echo("Deranged Hermit", "{3}{G}{G}"),
			WhenThisEnters("Deranged Hermit — create four 1/1 green Squirrels",
				Do(CreateToken{Template: TokenCard("1/1 green Squirrel"), N: 4})),
		},
	})
}
