package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampirism — Enchantment — Aura {1}{B} (Multiverse Gift Box, #5632):
//
//	"Enchant creature
//	 When this Aura enters, draw a card at the beginning of the next
//	 turn's upkeep.
//	 Enchanted creature gets +1/+1 for each other creature you control.
//	 Other creatures you control get -1/-1."
//
// "You" is the Aura's controller throughout, whoever controls the
// creature, and "other" is other than the enchanted creature. So the
// bonus counts the Aura controller's creatures except the host, and every
// one of those creatures gets -1/-1; on an opponent's creature the host
// counts all of your creatures and shrinks none of its own side. Both are
// layer 7c changes (CR 613.4c) recounted on every layer pass, so the bonus
// tracks the board: a 1/1 the -1/-1 kills stops counting once the
// state-based action has put it in the graveyard.
//
// The enters trigger schedules Foxfire's delayed draw (CR 603.7) at the
// next upkeep, whoever's turn it is; it is scheduled even if the Aura has
// gone by the time the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "14df239a-c7cc-492a-94f5-2971f6efe17b",
		Name:         "Vampirism",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Vampirism — draw a card at the beginning of the next turn's upkeep", func(g *game.Game, item *game.StackItem) error {
				return ScheduleDelayedTrigger{
					At:    game.StepUpkeep,
					Label: "Vampirism — draw a card",
					Body:  drawOneBody,
				}.Apply(NewContext(g, item))
			}),
		},
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, vampirismOtherCreaturesYouControl),
			b16Anthem(vampirismOtherThanEnchanted, -1, -1),
		},
	})
}

// vampirismOtherCreaturesYouControl counts the creatures the Aura's
// controller controls, the enchanted creature left out.
func vampirismOtherCreaturesYouControl(g *game.Game, source *game.Card) int {
	return permanentsYouControl(g, source, func(c *game.Card) bool {
		return c.IsCreature() && !source.IsAttachedTo(c.InstanceID)
	})
}

// vampirismOtherThanEnchanted is "other creatures you control": the Aura
// controller's creatures but the one it enchants.
func vampirismOtherThanEnchanted(target *game.Card, g *game.Game, source *game.Card) bool {
	return b16CreaturesYouControl(target, g, source) && !source.IsAttachedTo(target.InstanceID)
}
