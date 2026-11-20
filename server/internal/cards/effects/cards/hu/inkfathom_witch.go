package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inkfathom Witch — Creature — Merfolk Wizard {1}{U/B}, 1/1:
//
//	"Fear
//	 {2}{U}{B}: Each unblocked creature has base power and toughness 4/1
//	 until end of turn."
//
// The set of creatures is fixed as the ability resolves (CR 611.2c): every
// attacking creature that is unblocked at that moment, whoever controls
// it. It sets BASE power and toughness (layer 7b), so counters and other
// boosts still apply on top, and it lasts until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "60b6e96a-63e6-4520-b2dd-b20b7dfbcd57",
		Name:            "Inkfathom Witch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}{B}: Each unblocked creature has base power and toughness 4/1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{U}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ScopedEffectFor{
					Match: func(g *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.IsCreature() && g.UnblockedAttackerForEffect(c.InstanceID)
					},
					Mods:     game.SetBasePTMods(4, 1),
					Duration: g.UntilEndOfTurnDuration(),
					Label:    "Inkfathom Witch — base power and toughness 4/1",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
