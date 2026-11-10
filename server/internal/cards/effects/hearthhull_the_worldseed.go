package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hearthhull, the Worldseed — Legendary Artifact — Spacecraft
// {1}{B}{R}{G}, 6/7 (Edge of Eternities Commander):
//
//	"Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 8+.)
//	 2+ | {1}, {T}, Sacrifice a land: Draw two cards. You may play an
//	 additional land this turn.
//	 8+ | Flying, vigilance, haste
//	 Whenever you sacrifice a land, each opponent loses 2 life."
//
// Station and the 8+ creature line are the shared constructors
// (SpacecraftAt, ThresholdKeywords). The 2+ ability is an activated
// ability behind the charge-counter gate (ActivatedAbility.ActiveWhen),
// with a land sacrifice as a cost (itself a sacrifice, so it also feeds
// the 8+ trigger) and Explore's one-shot extra land play, which lasts the
// turn and is the controller's own (the rulings). The 8+ trigger is a
// per-land sacrifice watcher behind the same gate, with the footprinted
// drain primitive.
//
// No simplification.
func init() {
	drain := On(game.EventSacrifice, cmdrB9YouSacrificedLand,
		"Hearthhull — each opponent loses 2 life", Do(EachOpponentLosesLife{N: 2}))
	drain.ActiveWhen = AtChargeCounters(8)
	Register(Spec{
		OracleID:     "8d6f82d9-5c5b-4c41-8857-c0936f2d894c",
		Name:         "Hearthhull, the Worldseed",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{drain},
		Activated: []ActivatedAbility{
			Station(),
			{
				Label:      "{1}, {T}, Sacrifice a land: Draw two cards. You may play an additional land this turn.",
				Purpose:    game.Purpose{Draws: 2, Answers: game.AnswerValue},
				Cost:       Plus(ManaCost("{1}"), TapCost(), b08SacrificeALand()),
				ActiveWhen: AtChargeCounters(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := (DrawCards{N: 2}).Apply(NewContext(g, item)); err != nil {
						return err
					}
					g.GrantAdditionalLandPlayForEffect(item.Controller, 1)
					return nil
				},
			},
		},
		Static: append(
			SpacecraftAt(8, 6, 7),
			ThresholdKeywords(8, "flying", "vigilance", "haste"),
		),
	})
}
