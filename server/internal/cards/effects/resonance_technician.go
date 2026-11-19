package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "ec8a4b26-633a-4e88-aa8a-82e9704b3439",
		Name:            "Resonance Technician",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player:   ctx.Controller(),
					Source:   ctx.Source(),
					N:        1,
					UpTo:     true,
					Question: label,
				}, resonanceTechnicianInvestigates, nil)
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Tap X untapped artifacts you control: Copy target instant or sorcery spell you control with mana value X. You may choose new targets for the copy.",
			Cost:    Plus(TapCost(), TapXUntapped("X untapped artifacts you control", Artifact())),
			Targets: instantOrSorcerySpell("target instant or sorcery spell you control with mana value X", YouControl()).WithManaValueEqualsX(),
			Effect:  copyTargetedSpell,
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
