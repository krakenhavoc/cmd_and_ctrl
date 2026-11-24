package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loathsome Curator — Creature — Gorgon Wizard {4}{B}, 5/4 (Modern Horizons 2):
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 Menace
//	 When this creature exploits a creature, destroy target creature you don't
//	 control with mana value 3 or less."
//
// Exploit is the shared constructor (CR 702.110); the payoff is a reflexive
// trigger raised only when a creature really was sacrificed, and its target is
// chosen then, a creature an opponent controls with mana value 3 or less.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8a20707f-9764-4236-9aa9-858a2419f012",
		Name:            "Loathsome Curator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Loathsome Curator", loathsomeCuratorDestroyBody),
		},
	})
}

// loathsomeCuratorDestroyBody is the exploit payoff. An on-disk key: never
// renamed, never reused.
var loathsomeCuratorDestroyBody = game.ReflexiveBody("loathsome-curator/destroy", effects.SimpleBody(effects.DestroyFirstLegalTarget),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature you don't control with mana value 3 or less", effects.OpponentControls(), effects.ManaValueLE(3))
	}))
