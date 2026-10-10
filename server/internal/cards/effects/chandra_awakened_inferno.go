package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "7b2a600b-d6c8-45ff-a7fc-06105d27111f",
		Name:            "Chandra, Awakened Inferno",
		Completeness:    CompletenessFull,
		XMatters:        true,
		CantBeCountered: true,
		StartingLoyalty: 6,
		Emblem: &EmblemSpec{
			Label: "Chandra, Awakened Inferno emblem",
			Text:  "At the beginning of your upkeep, this emblem deals 1 damage to you.",
			Triggered: []game.TriggeredAbility{
				AtYourUpkeep("Chandra, Awakened Inferno emblem — 1 damage to you", func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: 1}.Apply(ctx)
				}),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+2: Each opponent gets an emblem with \"At the beginning of your upkeep, this emblem deals 1 damage to you.\"",
				Cost:  LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, opp := range ctx.Opponents() {
						if err := (CreateEmblem{Player: opp}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:   "−3: Chandra deals 3 damage to each non-Elemental creature.",
				Cost:    LoyaltyCost(-3),
				Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return damageEachMatching(NewContext(g, item), And(Creature(), Not(HasSubtype("Elemental"))), 3)
				},
			},
			{
				Label:   "−X: Chandra deals X damage to target creature or planeswalker. If a permanent dealt damage this way would die this turn, exile it instead.",
				Cost:    LoyaltyMinusX(),
				Targets: TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
				Purpose: ForTargets(DamageXToTarget(0)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := firstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return DealDamageToEachThen(ctx, []uuid.UUID{t.ID}, ctx.X(), ExileIfDealtDamageWouldDie(item, true))
				},
			},
		},
	})
}
