package legal

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// CombatKindForTest is #2871's combat read as it stood before ADR 0142
// routed it through fallbackAnswers: 0 none, 1 any combat window, 2
// defender only. TestCombatFlagsKeepEveryCatalogVerdict holds the new
// flags to it.
func CombatKindForTest(ab game.ActivatedAbilityShape) int { return int(abilityCombatKind(ab)) }
