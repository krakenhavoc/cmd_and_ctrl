package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonds of Faith — {1}{W} Enchantment — Aura (Shadows over Innistrad via Duel
// Decks: Blessed vs. Cursed, S82 OLD b294):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 as long as it's a Human. Otherwise, it
//	 can't attack or block."
//
// Two halves of one condition read live every pass (Burden of Proof's
// shape): a Human host gets a layer 7c +2/+2; any other host gets the
// can't-attack and can't-block restriction bits. "Human" is a word of the
// Aura's text (CR 612.2).
//
// No simplification.
func init() {
	isHuman := func(host *game.Card, source *game.Card) bool { return host.HasSubtype(source.TextSubtype("Human")) }
	effects.Register(effects.Spec{
		OracleID:     "79af045e-2c73-4003-8f62-b8611a225f08",
		Name:         "Bonds of Faith",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			{
				Label:    "Enchanted creature gets +2/+2 as long as it's a Human.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return effects.AttachedToSource(target, g, source) && isHuman(target, source)
				},
				Apply: effects.PumpBy(2, 2),
			},
			effects.RestrictAttachedWhile(func(host *game.Card, _ *game.Game, source *game.Card) bool {
				return !isHuman(host, source)
			}, game.CantAttack|game.CantBlock),
		},
	})
}
