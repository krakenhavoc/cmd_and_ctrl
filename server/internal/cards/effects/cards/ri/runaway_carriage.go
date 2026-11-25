package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runaway Carriage — {4} Artifact Creature — Construct 5/6 (Shadows over
// Innistrad, S82 OLD b293):
//
//	"Trample
//	 When this creature attacks or blocks, sacrifice it at end of combat."
//
// Dorothea's attack-or-block trigger over the delayed end-of-combat
// sacrifice (CR 511.2): one printed ability, one registered trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "67c16435-54ef-4786-85cd-cacca4f06670",
		Name:            "Runaway Carriage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
			}, "Runaway Carriage — sacrifice it at end of combat",
				effects.SacrificeThisAtEndOfCombat("Runaway Carriage — sacrifice it")),
		},
	})
}
