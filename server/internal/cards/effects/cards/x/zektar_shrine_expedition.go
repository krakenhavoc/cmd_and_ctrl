package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zektar Shrine Expedition — Enchantment {1}{R} (Zendikar, S82 OLD b496):
//
//	"Landfall — Whenever a land you control enters, you may put a quest counter on
//	 this enchantment.
//	 Remove three quest counters from this enchantment and sacrifice it: Create a
//	 7/1 red Elemental creature token with trample and haste. Exile it at the
//	 beginning of the next end step."
//
// Khalni Heart Expedition's landfall counter and cost; the token and its delayed
// exile are Rakdos Guildmage's goblin (Old496ZektarElemental): only the token the
// ability made is exiled, at the next end step, wherever it then is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45bd5e97-f34e-47bc-ad21-da42d3444431",
		Name:         "Zektar Shrine Expedition",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Landfall("Zektar Shrine Expedition — put a quest counter on it",
				effects.Do(effects.CounterOnThis{Kind: "quest", N: 1})), "Put a quest counter on Zektar Shrine Expedition?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove three quest counters from this enchantment and sacrifice it: Create a 7/1 red Elemental creature token with trample and haste. Exile it at the beginning of the next end step.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.RemoveCountersFromThis("quest", 3), effects.SacrificeThis()),
			Effect:  effects.Old496ZektarElemental,
		}},
	})
}
