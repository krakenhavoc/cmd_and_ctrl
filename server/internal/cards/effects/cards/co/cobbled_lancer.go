package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cobbled Lancer — Creature — Zombie Horse {U}, 3/3 (Innistrad: Crimson Vow):
//
//	"As an additional cost to cast this spell, exile a creature card from your
//	 graveyard.
//	 {3}{U}, Exile this card from your graveyard: Draw a card."
//
// The additional cost is Corpse Explosion's: the caster names the card at
// cast time and it is exiled with the spell on the stack (CR 601.2b, 601.2h),
// so with no creature card in the graveyard the spell cannot be cast. The
// graveyard ability exiles the card as a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "200bbf1d-835d-4d61-a6ab-a0b2abfbed79",
		Name:         "Cobbled Lancer",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.ExileFromGraveyardCost(1, "a creature card", func(c game.Card) bool {
			return c.IsCreature()
		}),
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}, Exile this card from your graveyard: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{U}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
