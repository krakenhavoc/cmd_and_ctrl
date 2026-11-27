package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sakura-Tribe Springcaller — Creature — Snake Shaman {3}{G}, 2/4
// (Betrayers of Kamigawa, S82 OLD, #5602):
//
//	"At the beginning of your upkeep, add {G}. Until end of turn, you don't
//	 lose this mana as steps and phases end."
//
// Kessig Naturalist's keep-until-end-of-turn mana (KeepManaUntilEndOfTurn)
// on an upkeep trigger, so the {G} is still there in the main phases. It
// is a triggered ability, not a mana ability: it uses the stack (CR 605.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "675a0eff-4da0-45cd-9371-fca96c313f4c",
		Name:         "Sakura-Tribe Springcaller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Sakura-Tribe Springcaller — add {G}; you don't lose it until end of turn", effects.Do(effects.AddMana{
				Produced: "{G}",
				Riders:   []game.ManaSpendRider{effects.KeepManaUntilEndOfTurn()},
			})),
		},
	})
}
