package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b9cd714b-2ad8-4fdb-a8aa-82b17730e071",
		Name:         "Deranged Hermit",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Squirrel"}}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.Echo("Deranged Hermit", "{3}{G}{G}"),
			effects.WhenThisEnters("Deranged Hermit — create four 1/1 green Squirrels",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Squirrel"), N: 4})),
		},
	})
}
