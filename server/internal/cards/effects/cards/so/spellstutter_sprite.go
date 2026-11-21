package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spellstutter Sprite — Creature — Faerie Wizard {1}{U}, 1/1 (Jumpstart 2022, #5948):
//
//	"Flash
//	 Flying
//	 When this creature enters, counter target spell with mana value X or
//	 less, where X is the number of Faeries you control."
//
// X is read from the board, the Sprite included (it is on the battlefield
// as the trigger is put on the stack), both when the target is chosen and
// when the ability resolves, so a spell that is no longer small enough is
// an illegal target and the ability does nothing (CR 608.2b). Faeries are
// counted through the layers, so a changeling is one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "32e60fb4-e841-4f82-9c83-5e63766e8e6f",
		Name:            "Spellstutter Sprite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Spellstutter Sprite — counter target spell with mana value X or less, where X is the number of Faeries you control",
					effects.CounterTheChosenSpell),
				effects.TargetSpell("target spell with mana value X or less, where X is the number of Faeries you control",
					func(g *game.Game, caster uuid.UUID, c game.Card) bool {
						x := effects.CountControlled(g, caster, func(p game.Card) bool { return p.HasSubtypeWord("Faerie") })
						return effects.ManaValueLE(x)(g, caster, c)
					})),
		},
	})
}
