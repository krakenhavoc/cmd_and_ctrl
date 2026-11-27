package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stamina — Enchantment — Aura {2}{G} (Mercadian Masques, S82 OLD batch 686, #5619):
//
//	"Enchant creature
//	Enchanted creature has vigilance.
//	Sacrifice this Aura: Regenerate enchanted creature."
//
// Vigilance as a grant to the host, and Soul Channeling's regenerate-the-host with the Aura's own sacrifice as the cost: the shield goes on the creature the Aura was attached to as it left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b77217b8-22d1-4dec-a5b4-18fccfcb70e9",
		Name:         "Stamina",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("vigilance")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this Aura: Regenerate enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.SacrificeThis(),
			Effect:  staminaRegenerate,
		}},
	})
}

// staminaRegenerate puts a regeneration shield on the creature the Aura was
// attached to, which is the Aura's last-known attachment since sacrificing it
// is the cost.
func staminaRegenerate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	host, ok := effects.EnchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return effects.Regenerate{Target: host}.Apply(ctx)
}
