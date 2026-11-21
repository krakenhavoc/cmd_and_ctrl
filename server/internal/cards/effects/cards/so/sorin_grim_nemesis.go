package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sorin, Grim Nemesis — Legendary Planeswalker — Sorin {4}{W}{B},
// starting loyalty 6 (#4077, deck request Sorin's Birthday Party):
//
//	"+1: Reveal the top card of your library and put that card into
//	     your hand. Each opponent loses life equal to its mana value.
//	 −X: Sorin deals X damage to target creature or planeswalker and
//	     you gain X life.
//	 −9: Create a number of 1/1 black Vampire Knight creature tokens
//	     with lifelink equal to the highest life total among all
//	     players."
//
// The +1 is Dark Confidant's flip with the bill sent to the table: the
// reveal happens from the library, the card moves to the hand, and each
// opponent loses its mana value as one life loss apiece (life loss, not
// damage). Nothing is drawn. An empty library reveals nothing and costs
// nothing.
//
// The −X is a loyalty cost of X (LoyaltyMinusX, #1944; CR 606.6). The
// damage and the life gain are two sentences, so a prevented X still
// gains X; a target that is gone makes the whole ability do nothing.
//
// The −9 reads the highest life total among every player still in the
// game as it resolves, so it counts the activator's own total, and a
// total at or below zero makes no tokens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a01dd05-e289-489f-b3ad-88e15e157bd0",
		Name:         "Sorin, Grim Nemesis",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 6,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Reveal the top card of your library and put that card into your hand. Each opponent loses life equal to its mana value.",
				Cost:   effects.LoyaltyCost(1),
				Effect: sorinGrimNemesisReveal,
			},
			{
				Label:   "−X: Sorin deals X damage to target creature or planeswalker and you gain X life.",
				Cost:    effects.LoyaltyMinusX(),
				Targets: effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
				Purpose: effects.ForTargets(effects.DamageXToTarget(0)),
				// The life gain waits for a prompt the damage asks (Khârn the Betrayer's
				// question, a CR 616 order; CR 608.2c, #4155).
				Effect: effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						// CR 608.2b: a target that is gone makes the ability
						// do nothing at all, the life gain included.
						t, ok := effects.FirstLegalTarget(ctx)
						if !ok {
							return nil
						}
						x := ctx.X()
						if x <= 0 {
							return nil
						}
						ctx.RecordStep(sorinGrimNemesisXRecord, nil, x)
						return effects.DealDamage{Source: ctx.Item.SourceCardID, Target: t.ID, Amount: x}.Apply(ctx)
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						// Two sentences: a prevented X still gains X.
						if x := ctx.StepRecord(sorinGrimNemesisXRecord).N(); x > 0 {
							return effects.GainLife{Amount: x}.Apply(ctx)
						}
						return nil
					}),
				),
			},
			{
				Label:  "−9: Create a number of 1/1 black Vampire Knight creature tokens with lifelink equal to the highest life total among all players.",
				Cost:   effects.LoyaltyCost(-9),
				Effect: sorinGrimNemesisArmy,
			},
		},
	})
}

// sorinGrimNemesisReveal is the +1, in printed order: reveal, put into
// hand, then each opponent loses life equal to the card's mana value
// (read before the move, the last moment it is certainly findable).
func sorinGrimNemesisReveal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var revealed []uuid.UUID
	if err := (effects.RevealTopOfLibrary{
		Player:   item.Controller,
		N:        1,
		Reason:   "Sorin, Grim Nemesis — reveal the top card of your library",
		Revealed: &revealed,
	}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) == 0 {
		return nil
	}
	mv := 0
	if c, ok := g.LookupCardForEffect(revealed[0]); ok {
		mv = c.ManaValue()
	}
	// "Each opponent loses life equal to its mana value" is a separate
	// sentence, not gated on the card having reached the hand, so the
	// bounce's continuation ignores `bounced` and pays the value read
	// before the move. It runs after any CR 903.9 question about the
	// card has been answered, which keeps the drain from landing first.
	return effects.BounceToHand{Target: revealed[0], ThenTail: effects.WithN(ctx.ItemTail(effects.ThenEachOpponentLosesN), mv)}.Apply(ctx)
}

// sorinGrimNemesisArmy is the −9.
func sorinGrimNemesisArmy(g *game.Game, item *game.StackItem) error {
	best := 0
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if p.Life > best {
			best = p.Life
		}
	}
	if best <= 0 {
		return nil
	}
	return effects.CreateToken{
		Template: effects.TokenCard("1/1 black Vampire Knight with lifelink"),
		N:        best,
	}.Apply(effects.NewContext(g, item))
}

// sorinGrimNemesisXRecord is X, filed once the damage has a target.
const sorinGrimNemesisXRecord = "sorin-grim-nemesis/x"
