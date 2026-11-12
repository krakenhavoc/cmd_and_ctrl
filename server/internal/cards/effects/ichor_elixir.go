package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ichor Elixir — Artifact {4} (March of the Machine Commander, #5648):
//
//	"If you would roll one or more planar dice, instead roll that many
//	 planar dice plus one and ignore one.
//	 {T}: Add {C}{C}."
//
// The first ability is a replacement on the dice roll's window (CR
// 614.1a, game.RepEventRollDice with RollPlanar): one more planar die is
// rolled, and once they are rolled its controller, who is the roller,
// chooses one to ignore. An ignored roll never happened (CR 706.6): no
// chaos ensues and no planeswalking ability triggers because of it. It
// applies to the planar die special action (CR 901.9) and to a roll an
// effect asks for alike, and never to a numbered die. Two Elixirs add two
// dice and ignore two. Without a planar die in the game, at a table not
// playing Planechase, it replaces nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81b06e8d-ed30-4f90-b729-0aa93b0d870f",
		Name:         "Ichor Elixir",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventRollDie},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventRollDice && ev.RollPlanar && ev.RollCount > 0 && ev.RollPlayer == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
				ev.RollOneMoreAndIgnoreOne(src.Controller, src.InstanceID, "Ichor Elixir")
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label:    "Ichor Elixir: roll one more planar die and ignore one",
			RowLabel: "If you would roll one or more planar dice, instead roll that many planar dice plus one and ignore one.",
		}},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
	})
}
