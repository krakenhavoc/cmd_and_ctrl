package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ecologist's Terrarium — Artifact {2} (Kamigawa: Neon Dynasty, S82):
//
//	"When this artifact enters, you may search your library for a basic land
//	 card, reveal it, put it into your hand, then shuffle.
//	 {2}, {T}, Sacrifice this artifact: Put a +1/+1 counter on target
//	 creature. Activate only as a sorcery."
//
// The enters trigger is Archaeomancer's Map's search with "you may" (the
// chooser can decline, and the shuffle with it). The sacrifice is a cost, so
// the counter lands on the target even though the Terrarium has gone; only an
// illegal target fizzles it (CR 608.2b). {T} is the artifact's own, so the
// activation is not offered to a summoning-sick creature (it is not one).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04b33a43-57f1-46e9-b2c2-1dd3f8a4935c",
		Name:         "Ecologist's Terrarium",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Ecologist's Terrarium — you may search for a basic land card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Ecologist's Terrarium — a basic land card",
				}.Apply(effects.NewContext(g, item))
			}), game.Purpose{Tutors: 1}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}, {T}, Sacrifice this artifact: Put a +1/+1 counter on target creature. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Targets:      effects.TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect:       effects.Cmdr2S46PlusOneCounterOnClauseTarget,
		}},
	})
}
