package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soulblade Djinn — Creature — Djinn {3}{U}{U}, 4/3 (Game Night):
//
//	"Flying
//	 Whenever you cast a noncreature spell, creatures you control get +1/+1
//	 until end of turn."
//
// A cast trigger on the stack above the spell: the creatures are those you
// control as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a598e6ce-21ee-49c1-989e-ee864208262f",
		Name:            "Soulblade Djinn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Not(effects.Creature()), "Soulblade Djinn — creatures you control get +1/+1 until end of turn",
				effects.Do(effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()),
					Power: 1, Toughness: 1,
					Label: "Soulblade Djinn — +1/+1",
				})),
		},
	})
}
