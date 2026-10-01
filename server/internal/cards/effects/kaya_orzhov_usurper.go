package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaya, Orzhov Usurper — Legendary Planeswalker — Kaya {1}{W}{B},
// loyalty 3:
//
//	"+1: Exile up to two target cards from a single graveyard. You gain
//	     2 life if at least one creature card was exiled this way.
//	 −1: Exile target nonland permanent with mana value 1 or less.
//	 −5: Kaya deals damage to target player equal to the number of
//	     cards that player owns in exile and you gain that much life."
//
// #1807, ADR 0106 §5. The +1 is the family's clause; its life gain is
// gated on a creature card that actually reached exile (#870). The −1
// reads mana value as the engine does everywhere: X counts as 0 and a
// token that copies nothing has mana value 0 (the 2019-01-25 rulings).
//
// The −5 counts every card in exile that player owns, face up or face
// down, at resolution. "That much life" is that COUNT, not the damage
// dealt: the 2019-01-25 ruling says the life gained is the number of
// cards even if an effect makes Kaya deal more or less damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7dd4a1a1-d5f4-4ac7-a9f6-34af411f070b",
		Name:            "Kaya, Orzhov Usurper",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Exile up to two target cards from a single graveyard. You gain 2 life if at least one creature card was exiled this way.",
				Cost:    LoyaltyCost(1),
				Targets: upToNCardsFromASingleGraveyard(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return exileTargetCardsThen(NewContext(g, item), func(ctx *Context, exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) error {
						if !anyWasCreature(exiled, wasCreature) {
							return nil
						}
						return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
					})
				},
			},
			{
				Label:   "−1: Exile target nonland permanent with mana value 1 or less.",
				Cost:    LoyaltyCost(-1),
				Targets: TargetPermanent("target nonland permanent with mana value 1 or less", Nonland(), ManaValueLE(1)),
				Effect:  ExileFirstTarget,
			},
			{
				Label:   "−5: Kaya deals damage to target player equal to the number of cards that player owns in exile and you gain that much life.",
				Cost:    LoyaltyCost(-5),
				Targets: TargetPlayer("target player"),
				Effect:  kayaOrzhovUsurperUltimate,
			},
		},
	})
}

// kayaOrzhovUsurperUltimate is the −5: damage equal to the cards the
// target player owns in exile, and that many life.
func kayaOrzhovUsurperUltimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	n := 0
	if g.Exile != nil {
		for _, c := range g.Exile.Cards {
			if c.Owner == t.ID {
				n++
			}
		}
	}
	if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: item.Controller, Amount: n}.Apply(ctx)
}
