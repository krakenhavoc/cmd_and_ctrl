package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Favorable Destiny — Enchantment — Aura {1}{W} (Mirage, S82 OLD batch 811,
// #5633):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+2 as long as it's white.
//	 Enchanted creature has shroud as long as its controller controls another
//	 creature. (It can't be the target of spells or abilities.)"
//
// Nyx Infusion's conditional pump and a conditional grant, each re-read on
// every recompute: the colour word is a word of the Aura's text (CR 612.2) and
// is judged on the host as earlier layers leave it; "another creature" is any
// creature its controller controls but the host (Markov Crusader's count).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee04afb0-184f-4a13-8d44-0fbe121b1eff",
		Name:         "Favorable Destiny",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(func(host *game.Card, g *game.Game, source *game.Card) bool {
				return effects.AttachedToSource(host, g, source) && host.HasColor(source.TextColor("W"))
			}, 1, 2),
			effects.GrantToAttachedWhile(effects.S82OldB811ControllerHasAnotherCreature, effects.KeywordShroud),
		},
	})
}
