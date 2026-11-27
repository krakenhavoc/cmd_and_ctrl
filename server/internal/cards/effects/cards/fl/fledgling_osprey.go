package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fledgling Osprey — Creature — Bird {U}, 1/1 (Urza's Destiny, S82 #5621):
//
//	"This creature has flying as long as it's enchanted."
//
// Gate Hound's live read: a layer-6 flying on the Osprey alone while an Aura is
// attached to it (CR 303.4), so it follows the Aura on and off.
//
// No simplification.
func init() {
	flying := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfCreature(target, g, source) && effects.IsEnchantedByAura(g, source.InstanceID)
	}, "flying")
	flying.Label = "This creature has flying as long as it's enchanted."
	effects.Register(effects.Spec{
		OracleID:     "d5482ccf-7c3f-45b8-aaa8-8f6686fe584d",
		Name:         "Fledgling Osprey",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{flying},
	})
}
