package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lokhust Heavy Destroyer — Artifact Creature — Necron {1}{B}{B}{B}, 3/2
// (Warhammer 40,000 Commander, S82 GAP batch 188, #5948):
//
//	"Flying
//	 Enmitic Exterminator — When this creature enters, each player
//	 sacrifices a creature of their choice.
//	 Unearth {5}{B}{B}{B}"
//
// Wasteland Raider's edict: every player, the controller included,
// sacrifices a creature of their own choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fd03f20a-d365-48fb-8b2a-ad4b502a0045",
		Name:            "Lokhust Heavy Destroyer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.Unearth("{5}{B}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WhenThisEnters("Lokhust Heavy Destroyer — each player sacrifices a creature of their choice",
				effects.Do(effects.EachPlayerSacrifices{Match: effects.Creature(), Label: "a creature"}))),
		},
	})
}
