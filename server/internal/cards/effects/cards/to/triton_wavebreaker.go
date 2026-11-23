package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triton Wavebreaker — Enchantment Creature — Merfolk Wizard {U}, 1/1
// (Modern Horizons 3):
//
//	"Bestow {1}{U} (If you cast this card for its bestow cost, it's an Aura
//	 spell with enchant creature. It becomes a creature again if it's not
//	 attached.)
//	 As long as this permanent is a creature, it has prowess. (Whenever you
//	 cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	 Enchanted creature gets +1/+1 and has prowess."
//
// Bestow is the shared alternative cost (ADR 0141). The prowess is the
// engine's keyword token (game/prowess.go) granted in layer 6 twice over:
// to the permanent itself only while it is a creature (read after layer 4,
// so a bestowed Aura has none) and, with the +1/+1, to the enchanted
// creature as Wizard's Staff grants it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "d2309fba-4273-4084-b206-642a43975c34",
		Name:             "Triton Wavebreaker",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Bestow("{1}{U}")},
		Static: []game.StaticAbility{
			effects.Labelled(effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfOnly(target, g, source) && target.IsCreature()
			}, game.KeywordProwess), "As long as this permanent is a creature, it has prowess."),
			effects.PumpAttached(1, 1),
			effects.GrantToAttached(game.KeywordProwess),
		},
	})
}
