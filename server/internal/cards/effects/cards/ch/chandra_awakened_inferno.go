package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra, Awakened Inferno — Legendary Planeswalker — Chandra
// {4}{R}{R}, starting loyalty 6 (EDHREC rank 3348):
//
//	"This spell can't be countered.
//	 +2: Each opponent gets an emblem with "At the beginning of your
//	     upkeep, this emblem deals 1 damage to you."
//	 −3: Chandra deals 3 damage to each non-Elemental creature.
//	 −X: Chandra deals X damage to target creature or planeswalker. If a
//	     permanent dealt damage this way would die this turn, exile it
//	     instead."
//
// The +2 gives each opponent their own emblem: CR 114.2's owner is the
// player who gets it, so "your upkeep" and "you" are that opponent's,
// and the emblem is the source of the damage. The −3 is the ordinary
// mass-damage helper over a non-Elemental filter. The −X is a loyalty
// cost of X (LoyaltyMinusX, #1944): X is announced with the activation
// and no more than her loyalty (CR 606.6), and its rider is ADR 0108's
// exile-if-it-would-die shield on each permanent the damage reached
// (a planeswalker too: "a permanent dealt damage this way").
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7b2a600b-d6c8-45ff-a7fc-06105d27111f",
		Name:            "Chandra, Awakened Inferno",
		Completeness:    effects.CompletenessFull,
		XMatters:        true,
		CantBeCountered: true,
		StartingLoyalty: 6,
		Emblem: &effects.EmblemSpec{
			Label: "Chandra, Awakened Inferno emblem",
			Text:  "At the beginning of your upkeep, this emblem deals 1 damage to you.",
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Chandra, Awakened Inferno emblem — 1 damage to you", func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: 1}.Apply(ctx)
				}),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: Each opponent gets an emblem with \"At the beginning of your upkeep, this emblem deals 1 damage to you.\"",
				Cost:  effects.LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, opp := range ctx.Opponents() {
						if err := (effects.CreateEmblem{Player: opp}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:   "−3: Chandra deals 3 damage to each non-Elemental creature.",
				Cost:    effects.LoyaltyCost(-3),
				Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.Not(effects.HasSubtype("Elemental"))), 3)
				},
			},
			{
				Label:   "−X: Chandra deals X damage to target creature or planeswalker. If a permanent dealt damage this way would die this turn, exile it instead.",
				Cost:    effects.LoyaltyMinusX(),
				Targets: effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
				Purpose: effects.ForTargets(effects.DamageXToTarget(0)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return effects.DealDamageToEachThen(ctx, []uuid.UUID{t.ID}, ctx.X(), effects.ExileIfDealtDamageWouldDie(item, true))
				},
			},
		},
	})
}
