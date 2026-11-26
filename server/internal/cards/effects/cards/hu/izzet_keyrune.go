package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Izzet Keyrune — Artifact {3} (Return to Ravnica, S82 OLD):
//
//	"{T}: Add {U} or {R}.
//	 {U}{R}: Until end of turn, this artifact becomes a 2/1 blue and red
//	 Elemental artifact creature.
//	 Whenever this artifact deals combat damage to a player, you may draw a
//	 card. If you do, discard a card."
//
// Boros Keyrune's mana ability and animation (without a keyword), plus
// Salvage Drone's optional loot on combat damage: drawing is the "may", and
// the discard follows only if you chose to draw. It is still an artifact; a
// Keyrune that came under your control this turn can animate but can't
// attack or tap for mana while it is a creature (CR 302.6).
//
// No simplification.
func init() {
	const name = "Izzet Keyrune"
	effects.Register(effects.Spec{
		OracleID:     "8b8c0285-424f-4a82-bfb6-17752cfa71dd",
		Name:         name,
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U|R}",
			Label:    "Add {U} or {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{R}: Until end of turn, this artifact becomes a 2/1 blue and red Elemental artifact creature.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{U}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.Cmdr2B13Animate(effects.NewContext(g, item), item.SourceCardID,
					name+" — a 2/1 blue and red Elemental artifact creature until end of turn",
					2, 1, "Elemental", []string{"U", "R"})
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(name+" — you may draw a card. If you do, discard a card",
				effects.S82OldB420MayLoot(name)),
		},
	})
}
