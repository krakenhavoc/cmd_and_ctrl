package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra, Torch of Defiance — Legendary Planeswalker — Chandra
// {2}{R}{R}, starting loyalty 4:
//
//	"+1: Exile the top card of your library. You may cast that card. If
//	     you don't, Chandra deals 2 damage to each opponent.
//	 +1: Add {R}{R}.
//	 −3: Chandra deals 4 damage to target creature.
//	 −7: You get an emblem with "Whenever you cast a spell, this emblem
//	     deals 5 damage to any target.""
//
// The first +1 is game.ExileTopThenPaidCastOrElseForEffect (#3510,
// ADR 0099 amendment 2026-10-14): the cast happens as the ability
// resolves (CR 608.2g), so its timing is ignored but its mana cost is
// paid. The cast is made inside the ability's resolution through a
// resolution_cast prompt over the exiled card (#4250), paying for it then;
// "don't", a land or an empty library deals the 2 damage to each
// opponent, inside the resolution. The card stays in exile when it is not
// cast.
//
// The second +1 is a loyalty ability, so it uses the stack and is not a
// mana ability (CR 605.1a); the mana empties with the step. The emblem
// is the source of its own damage.
func init() {
	Register(Spec{
		OracleID:        "12cc800c-c7be-4811-8207-fd01a99cc892",
		Name:            "Chandra, Torch of Defiance",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Chandra, Torch of Defiance emblem",
			Text:  "Whenever you cast a spell, this emblem deals 5 damage to any target.",
			Triggered: []game.TriggeredAbility{
				Targeting(
					WheneverYouCast(nil, "Chandra, Torch of Defiance emblem — 5 damage to any target", chandraTorchEmblemDamage),
					TargetAny()),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Exile the top card of your library. You may cast that card. If you don't, Chandra deals 2 damage to each opponent.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return g.ExileTopThenPaidCastOrElseForEffect(game.PaidCastOrElse{
						Player:       ctx.Controller(),
						Source:       ctx.Source(),
						SourceName:   "Chandra, Torch of Defiance",
						IfNotCast:    chandraTorchTwoToEachOpponent,
						DeclineLabel: "Don't cast it (2 damage to each opponent)",
					})
				},
			},
			{
				Label: "+1: Add {R}{R}.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return AddMana{Produced: "{R}{R}"}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−3: Chandra deals 4 damage to target creature.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetCreature("target creature"),
				Purpose: ForTargets(DamageToTarget(0, 4)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := firstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 4}.Apply(ctx)
				},
			},
			{
				Label:  "−7: You get an emblem with \"Whenever you cast a spell, this emblem deals 5 damage to any target.\"",
				Cost:   LoyaltyCost(-7),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}

// chandraTorchTwoToEachOpponent is the first +1's "if you don't": run on
// the decline, with no question when the card can't be cast, and by the
// lapse of an accepted cast that was never made. Chandra is the source;
// a Chandra that has left the battlefield still deals it (CR 608.2h).
var chandraTorchTwoToEachOpponent = game.RegisterCastFollowUp("chandra-torch-of-defiance-2-to-each-opponent",
	func(g *game.Game, f game.CastFollowUp) error {
		var opponents []uuid.UUID
		for _, p := range g.Seats {
			if p == nil || p.Eliminated || p.ID == f.Player {
				continue
			}
			opponents = append(opponents, p.ID)
		}
		if len(opponents) == 0 {
			return nil
		}
		return g.DealDamageEachThenForEffect(f.Source, opponents, 2, nil)
	})

// chandraTorchEmblemDamage is the emblem trigger's resolution: 5 damage
// from the emblem to the target chosen as it triggered.
func chandraTorchEmblemDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := firstLegalTarget(ctx)
	if !ok {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}.Apply(ctx)
}
