package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karn, Silver Golem — Legendary Artifact Creature — Golem {5} 4/4 (Urza's
// Saga, #5624):
//
//	"Whenever Karn blocks or becomes blocked, it gets -4/+4 until end of turn.
//	 {1}: Target noncreature artifact becomes an artifact creature with power
//	 and toughness each equal to its mana value until end of turn."
//
// The trigger is Bushido's shape: once when Karn blocks, once when it becomes
// blocked however many creatures block it. The target keeps its other types
// and abilities; the base power and toughness are its mana value, read as the
// ability resolves (CR 608.2h).
//
// No simplification.
func init() {
	const label = "{1}: Target noncreature artifact becomes an artifact creature with power and toughness each equal to its mana value until end of turn."
	effects.Register(effects.Spec{
		OracleID:     "8a7c6391-020a-4d7a-8abb-702631bbee56",
		Name:         "Karn, Silver Golem",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, effects.ThisBlocksOrBecomesBlocked,
				"Karn, Silver Golem — it gets -4/+4 until end of turn",
				effects.ThisGetsUntilEndOfTurn(-4, 4, "Karn, Silver Golem — -4/+4 until end of turn")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetPermanent("target noncreature artifact", effects.And(effects.Artifact(), effects.Noncreature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.ArtifactBecomesCreatureOfItsManaValue(ctx, effects.DurationUntilEndOfTurn(ctx),
					"Karn, Silver Golem — an artifact creature until end of turn")
			},
		}},
	})
}
