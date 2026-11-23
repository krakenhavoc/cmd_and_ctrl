package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhys the Exiled — Legendary Creature — Elf Warrior {2}{G}, 3/2
// (Morningtide, #5592):
//
//	"Whenever Rhys attacks, you gain 1 life for each Elf you control.
//	 {B}, Sacrifice an Elf: Regenerate Rhys."
//
// The attack trigger counts the Elves you control as it resolves
// (CR 608.2h), Rhys among them while he is still on the battlefield, and
// any Elf permanent counts, not only creatures. The regeneration is Wake
// of Vultures' sacrifice-fed shield on the source (regenerateThisSource);
// "an Elf" may be Rhys himself, who is then gone and gets no shield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7613c50b-4568-4340-b64a-ab427aaffd51",
		Name:         "Rhys the Exiled",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Rhys the Exiled — gain 1 life for each Elf you control", rhysTheExiledGainLife),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice an Elf: Regenerate Rhys.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeN(1, "an Elf", effects.HasSubtype("Elf"))),
			Effect:  effects.RegenerateThisSource,
		}},
	})
}

// rhysTheExiledGainLife is "you gain 1 life for each Elf you control".
func rhysTheExiledGainLife(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.GainLife{Amount: effects.CountYouControl(ctx, effects.HasSubtype("Elf"))}.Apply(ctx)
}
