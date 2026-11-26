package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tetzimoc, Primal Death — Legendary Creature — Elder Dinosaur {4}{B}{B}, 6/6:
//
//	"Deathtouch
//	 {B}, Reveal this card from your hand: Put a prey counter on target
//	 creature. Activate only during your turn.
//	 When Tetzimoc enters, destroy each creature your opponents control with a
//	 prey counter on it."
//
// The activated ability works from the hand (CR 113.6b) with the reveal as
// part of its cost (Piercing Rays' forecast shape, game.AbilityCost.RevealThis)
// and is limited to its controller's turn. The enters trigger is one
// simultaneous destroy over every creature an opponent controls that carries
// a prey counter as it resolves (CR 700.4); it declares a Sweep for the audit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7565c101-a479-4812-9d18-63918084bfca",
		Name:            "Tetzimoc, Primal Death",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []effects.ActivatedAbility{{
			Label:     "{B}, Reveal this card from your hand: Put a prey counter on target creature. Activate only during your turn.",
			Zones:     []game.ZoneKind{game.ZoneHand},
			Cost:      effects.Plus(effects.ManaCost("{B}"), game.AbilityCost{RevealThis: true}),
			Targets:   effects.TargetCreature("target creature"),
			Condition: effects.DuringYourTurn(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, id := range effects.LegalTargetCardIDs(ctx) {
					return effects.AddCounter{Target: id, Kind: "prey", N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Tetzimoc — destroy each creature your opponents control with a prey counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.OpponentControls(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return effects.HasCounterOfKind(c, "prey") })}.Apply(effects.NewContext(g, item))
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, OpponentsOnly: true, Partial: true}}),
		},
	})
}
