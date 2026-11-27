package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reaper of the Wilds — Creature — Gorgon {2}{B}{G}, 4/5 (Theros):
//
//	"Whenever another creature dies, scry 1.
//	 {B}: This creature gains deathtouch until end of turn.
//	 {1}{G}: This creature gains hexproof until end of turn."
//
// Murder of Crows's death watcher (anyone's creature, the Reaper's own death
// excluded) scrying once per creature, and two self-granting rows that do
// nothing if the Reaper has left the battlefield (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10494ce1-7a38-44ae-9881-a83e44daaf7b",
		Name:         "Reaper of the Wilds",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B15AnotherCreatureDied(ev, source, g)
			},
			Key:     "Reaper of the Wilds — scry 1",
			Purpose: game.Purpose{DeathPayoff: true},
			Effect:  effects.Do(effects.Scry{N: 1}),
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}: This creature gains deathtouch until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{B}"),
				Effect:  effects.OldB395GainsUntilEOT("Reaper of the Wilds — deathtouch", "deathtouch"),
			},
			{
				Label:   "{1}{G}: This creature gains hexproof until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    effects.ManaCost("{1}{G}"),
				Effect:  effects.OldB395GainsUntilEOT("Reaper of the Wilds — hexproof", "hexproof"),
			},
		},
	})
}
