package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Mightstone and Weakstone — Legendary Artifact — Powerstone {5}:
//
//	"When The Mightstone and Weakstone enters, choose one —
//	 • Draw two cards.
//	 • Target creature gets -5/-5 until end of turn.
//	 {T}: Add {C}{C}. This mana can't be spent to cast nonartifact
//	 spells.
//	 (Melds with Urza, Lord Protector.)"
//
// The enters trigger is a modal triggered ability (#764): the mode is
// picked as it goes on the stack and the -5/-5 mode's target is chosen
// then and re-checked on resolution. The mana ability carries the
// nonartifact-spell restriction Karn, Legacy Reforged uses.
//
// The reminder line is the other half of Urza, Lord Protector's meld
// ability (CR 701.42, 712.5e; ADR 0145, #2699): the card carries its
// printed meld data from the deck import, so Urza's {7} finds it, exiles
// both and returns them as Urza, Planeswalker. Nothing on this card does
// anything for that; the ability is Urza's.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c396db03-bf11-4e20-b630-4f9aa8fd78da",
		Name:         "The Mightstone and Weakstone",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{C}{C}",
			Label:        "Add {C}{C}. This mana can't be spent to cast nonartifact spells.",
			Restrictions: []string{effects.ManaRestrictNotNonartifactSpell},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.B06SelfETB,
			Key:       "The Mightstone and Weakstone — choose one",
			Effect:    func(*game.Game, *game.StackItem) error { return nil },
			Modes: effects.ChooseOne(
				effects.ModeDoing("Draw two cards.", nil,
					func(item *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
					}),
				effects.ModeDoing("Target creature gets -5/-5 until end of turn.",
					effects.TargetCreature("target creature"),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						t, ok := effects.ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return effects.BoostUntilEOT{Target: t.ID, Power: -5, Toughness: -5,
							Label: "The Mightstone and Weakstone — -5/-5"}.Apply(ctx)
					}),
			),
		}},
	})
}
