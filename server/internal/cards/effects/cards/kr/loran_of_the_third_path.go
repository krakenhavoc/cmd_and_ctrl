package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loran of the Third Path — Legendary Creature — Human Artificer
// {2}{W}, 2/1 (EDHREC rank 344):
//
//	"Vigilance
//	 When Loran enters, destroy up to one target artifact or
//	 enchantment.
//	 {T}: You and target opponent each draw a card."
//
// White's Reclamation Sage with a political tap ability stapled on.
// Three abilities, all real:
//
//   - Vigilance is printed and enforced.
//   - The ETB's "up to one target" is modelled as a "you may" trigger
//     with a single target clause — the same shape Sun Titan and
//     Reclamation Sage use. Observably identical: with no legal
//     target the trigger is removed without a prompt (CR 603.3d),
//     with one the controller may decline. The destroy goes through
//     the S20 pick_target prompt and the CR 608.2b re-check.
//   - The tap ability targets an opponent and draws one card for the
//     controller and one for the target. CR 302.6 summoning sickness
//     applies (a tap cost on a creature), as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b3d81980-76f2-44e2-b1c9-01e30c726312",
		Name:            "Loran of the Third Path",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Loran of the Third Path — destroy up to one target artifact or enchantment?",
			},
			Key:    "Loran of the Third Path — destroy target artifact or enchantment",
			Effect: effects.DestroyChosenPermanent,
		}},
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: You and target opponent each draw a card.",
			// ADR 0126 §6: the opponent draws too, which no field says.
			// You draw one, and so does the opponent you aim it at.
			Purpose: game.Purpose{Draws: 1, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1})},
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: effects.Do(
				// The second draw waits for a prompt the first asks (CR 608.2c, #4155).
				effects.DrawCards{N: 1},
				effects.Instruction(func(ctx *effects.Context) error {
					if len(ctx.Item.Targets) == 0 || ctx.Item.Targets[0].Kind != game.TargetPlayer {
						return nil
					}
					return effects.DrawCards{Player: ctx.Item.Targets[0].ID, N: 1}.Apply(ctx)
				}),
			),
		}},
	})
}
