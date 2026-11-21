package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lucius the Eternal — Legendary Creature — Astartes Warrior {3}{B}{R}, 5/3:
//
//	"Haste
//	 Armour of Shrieking Souls — When Lucius the Eternal dies, exile it and
//	 choose target creature an opponent controls. When that creature leaves the
//	 battlefield, return this card from exile to the battlefield under its
//	 owner's control."
//
// The second sentence is a delayed trigger, created as the dies trigger
// resolves and waiting on the chosen creature (GapB190LuciusDies).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fa726aac-f4ae-4e34-ac08-36c1f50215f3",
		Name:            "Lucius the Eternal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Lucius the Eternal — exile it and choose target creature an opponent controls", effects.GapB190LuciusDies),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
