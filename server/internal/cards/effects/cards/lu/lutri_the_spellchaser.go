package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lutri, the Spellchaser — Legendary Creature — Elemental Otter
// {1}{U/R}{U/R}, 3/2:
//
//	"Companion — Each nonland card in your starting deck has a different
//	 name. (...)
//	 Flash
//	 When Lutri enters, if you cast it, copy target instant or sorcery
//	 spell you control. You may choose new targets for the copy."
//
// An intervening "if" (CR 603.4): the trigger exists only for a Lutri that
// entered from a spell, and is checked again on resolution. The target is
// chosen as the trigger goes on the stack; the copy is made under Lutri's
// controller with the new-target choice (CR 707.10c). Companion is the deck
// check's business (internal/deck companion registry, ADR 0172).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "158a6225-a246-4fd6-aa57-0df8067b4383",
		Name:            "Lutri, the Spellchaser",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.GeologicalAppraiserWasCast),
					"Lutri, the Spellchaser — copy target instant or sorcery spell you control",
					func(g *game.Game, item *game.StackItem) error {
						if effects.NewContext(g, item).CastProvenance().FromZone == "" {
							return nil
						}
						return effects.CopyTargetedSpell(g, item)
					}),
				effects.TargetInstantOrSorcerySpell("target instant or sorcery spell you control", effects.YouControl())),
		},
	})
}
