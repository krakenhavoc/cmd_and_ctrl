package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resonance Technician — Creature — Weird Detective {3}{U/R}{U/R}, 4/4
// (Ravnica: Clue Edition, #5929):
//
//	"Flying
//	 When this creature enters, you may discard a card. If you do,
//	 investigate twice.
//	 {T}, Tap X untapped artifacts you control: Copy target instant or
//	 sorcery spell you control with mana value X. You may choose new
//	 targets for the copy."
//
// The enter trigger is the shared "you may discard a card. If you do …" with
// two Clues as its payoff: the discard is the controller's own pick and a
// player with an empty hand discards nothing and investigates nothing. The
// activation is Secluded Starforge's "Tap X untapped artifacts" cost, where X
// is the announced X, and the target clause compares the spell's mana value
// with that X (Lazav's WithManaValueEqualsX). The tap symbol is part of the
// cost, so the Technician itself cannot pay it as one of the artifacts unless
// it is an artifact. The copy is Stella Lee's, with the CR 707.10c re-target
// offer.
//
// No simplification.
func init() {
	const label = "Resonance Technician — you may discard a card. If you do, investigate twice"
	effects.Register(effects.Spec{
		OracleID:        "ec8a4b26-633a-4e88-aa8a-82e9704b3439",
		Name:            "Resonance Technician",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player:   ctx.Controller(),
					Source:   ctx.Source(),
					N:        1,
					UpTo:     true,
					Question: label,
				}, resonanceTechnicianInvestigates, nil)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Tap X untapped artifacts you control: Copy target instant or sorcery spell you control with mana value X. You may choose new targets for the copy.",
			Cost:    effects.Plus(effects.TapCost(), effects.TapXUntapped("X untapped artifacts you control", effects.Artifact())),
			Targets: effects.TargetInstantOrSorcerySpell("target instant or sorcery spell you control with mana value X", effects.YouControl()).WithManaValueEqualsX(),
			Effect:  effects.CopyTargetedSpell,
		}},
	})
}

// resonanceTechnicianInvestigates is the "If you do, investigate twice", keyed.
var resonanceTechnicianInvestigates = game.RegisterDiscardThenRule("choose-cards/resonance-technician-investigate-twice", nil,
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Discarded.Count() == 0 {
			return nil
		}
		return g.InvestigateForEffect(r.Player, r.Source, 2)
	})
