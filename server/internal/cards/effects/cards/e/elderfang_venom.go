package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elderfang Venom — Enchantment {2}{B}{G} (Kaldheim Commander, #5506):
//
//	"Attacking Elves you control have deathtouch.
//	 Whenever an Elf you control dies, each opponent loses 1 life and you
//	 gain 1 life."
//
// The grant is Ohran Frostfang's layer-6 shape over attacking creatures that
// are Elves (read through the layered view, "Elf" a word of this card's
// text). The drain is one trigger per Elf that dies, read as it last
// existed (CR 603.10a); each opponent loses 1 and you gain 1
// (Bastion of Remembrance's drain: you gain 1, not one per opponent).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4741d842-12e6-4a8c-96b1-9e68c7074874",
		Name:         "Elderfang Venom",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:                    "Attacking Elves you control have deathtouch.",
			Layer:                    game.Layer6Ability,
			DependsOnAttackingStatus: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.AttackingTarget != uuid.Nil && target.HasSubtype(source.TextSubtype("Elf"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				effects.AppendKeywordsTo(c, []string{"deathtouch"})
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.S82OldB074ElfYouControlDied(ev, source, g)
			}, "Elderfang Venom — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
