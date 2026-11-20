package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dal R'El, Augment Orphan — Legendary Creature — Human Mutant Rogue
// {1}{R}{G}, 3/3 (Star Trek, #4973):
//
//	"Landfall — Whenever a land you control enters, add {R} or {G}.
//	 {3}, Sacrifice a land: Draw a card."
//
// The landfall line is Lotus Cobra's shape with a two-colour choice. It
// is not a mana ability: it triggers on a land entering, not on a mana
// ability (CR 605.1b), so it uses the stack and the mana empties at the
// end of the step (CR 106.4). The activated ability sacrifices a land you
// control as part of its cost (CR 602.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7cecf129-3df9-4ebf-b194-28d1b34faaab",
		Name:         "Dal R'El, Augment Orphan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Dal R'El, Augment Orphan — add {R} or {G} (landfall)", game.LandfallPayoff{Mana: 1},
				effects.Do(effects.AddMana{Produced: "{R|G}"})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, Sacrifice a land: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.SacrificeN(1, "a land", effects.Land())),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
