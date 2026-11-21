package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crack in Time — Enchantment {3}{W} (Doctor Who):
//
//	"Vanishing 3 (This enchantment enters with three time counters on it.
//	 At the beginning of your upkeep, remove a time counter from it. When
//	 the last is removed, sacrifice it.)
//	 When this enchantment enters and at the beginning of your first main
//	 phase, exile target creature an opponent controls until this
//	 enchantment leaves the battlefield."
//
// One printed ability with two trigger conditions (OnAny over ETB and the
// beginning of your first main phase, with Võx's condition). Each exile is its own
// CR 610.3 "until this leaves" record, so every creature it holds returns
// when the enchantment leaves.
//
// No simplification.
func init() {
	exile := effects.OnAny([]game.EventKind{game.EventETB, game.EventBeginPrecombatMain}, effects.ThisEntersOrYourFirstMain,
		"Crack in Time — exile target creature an opponent controls until this enchantment leaves the battlefield",
		effects.ExileChosenTargetUntilThisLeaves("Crack in Time — the exiled creature returns when Crack in Time leaves the battlefield"))
	exile.Targets = effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())
	effects.Register(effects.Spec{
		OracleID:     "a3099143-0146-435f-8d29-cd364aefbc2b",
		Name:         "Crack in Time",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Cmdr2B17VanishingEntry(3)},
		Triggered:    append(effects.Cmdr2B17VanishingTriggers(), exile),
	})
}
