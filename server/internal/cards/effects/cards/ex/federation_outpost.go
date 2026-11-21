package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Federation Outpost — Land (Star Trek, #4973):
//
//	"This land enters tapped unless you control two or more other lands.
//	 When this land enters untapped, choose one —
//	 • You gain 2 life.
//	 • Surveil 1.
//	 {T}: Add {W}."
//
// The slowland condition (otherLandsAtLeast(2)) as the enters-tapped
// self-replacement, then a modal enters trigger gated on the land being
// untapped once it is on the battlefield (selfEnteredUntapped, the
// Mystic Sanctuary reading): entered tapped, no trigger at all. The
// mode is chosen as the trigger goes on the stack (CR 700.2b).
//
// No simplification.
func init() {
	etb := effects.On(game.EventETB, effects.SelfEnteredUntapped, "Federation Outpost — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = effects.ChooseOne(
		effects.ModeDoing("You gain 2 life.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.GainLife{Amount: 2}.Apply(ctx)
		}),
		effects.ModeDoing("Surveil 1.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.Surveil{Player: item.Controller, N: 1}.Apply(ctx)
		}),
	)
	effects.Register(effects.Spec{
		OracleID:     "a8bcc930-3efe-4595-b164-f5985e1dc289",
		Name:         "Federation Outpost",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.EntersTappedUnless(effects.OtherLandsAtLeast(2))},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Triggered: []game.TriggeredAbility{etb},
	})
}
