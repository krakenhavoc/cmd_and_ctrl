package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Field Marshal — Creature — Human Soldier {1}{W}{W}, 2/2 (Secret Lair Drop,
// #5462):
//
//	"Other Soldier creatures get +1/+1 and have first strike."
//
// Captain of the Watch's lord over every OTHER Soldier on the battlefield,
// an opponent's as well as the controller's (the text does not say "you
// control"): a layer 7c anthem and a layer 6 keyword grant.
//
// No simplification.
func init() {
	soldiers := effects.TribeFilter{Tribes: []string{"Soldier"}, Others: true}
	effects.Register(effects.Spec{
		OracleID:     "1c0a53fc-8037-46e3-90ea-cb8b73631a83",
		Name:         "Field Marshal",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(soldiers, 1, 1),
			effects.TribalKeywordGrant(soldiers, "first strike"),
		},
	})
}
