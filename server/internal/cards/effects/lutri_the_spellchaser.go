package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "158a6225-a246-4fd6-aa57-0df8067b4383",
		Name:            "Lutri, the Spellchaser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, geologicalAppraiserWasCast),
					"Lutri, the Spellchaser — copy target instant or sorcery spell you control",
					func(g *game.Game, item *game.StackItem) error {
						if NewContext(g, item).CastProvenance().FromZone == "" {
							return nil
						}
						return copyTargetedSpell(g, item)
					}),
				instantOrSorcerySpell("target instant or sorcery spell you control", YouControl())),
		},
	})
}
