package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nicol Bolas, Planeswalker — Legendary Planeswalker — Bolas {4}{U}{B}{B}{R},
// loyalty 5:
//
//	"+3: Destroy target noncreature permanent.
//	 −2: Gain control of target creature.
//	 −9: Nicol Bolas deals 7 damage to target player or planeswalker. That
//	 player or that planeswalker's controller discards seven cards, then
//	 sacrifices seven permanents of their choice."
//
// −2 has no stated duration, so the control lasts indefinitely (CR 611.2a).
// −9 reads the planeswalker's controller before the damage, since seven
// damage may kill it; the discard is asked first and the sacrifice of seven
// permanents (as many as they have) is its continuation, so it comes after
// the cards are gone (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "05e9b55e-6329-48ec-b7d7-24c6b9692244",
		Name:            "Nicol Bolas, Planeswalker",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+3: Destroy target noncreature permanent.",
				Cost:    effects.LoyaltyCost(3),
				Targets: effects.TargetPermanent("target noncreature permanent", effects.Not(effects.Creature())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
				},
			},
			{
				Label:   "−2: Gain control of target creature.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GainControl{Target: id, Controller: item.Controller, Duration: game.IndefiniteDuration(),
						Label: "Nicol Bolas, Planeswalker — gain control (no stated duration)"}.Apply(ctx)
				},
			},
			{
				Label: "−9: Nicol Bolas deals 7 damage to target player or planeswalker. That player or that " +
					"planeswalker's controller discards seven cards, then sacrifices seven permanents of their choice.",
				Cost:    effects.LoyaltyCost(-9),
				Targets: effects.TargetPlayerOrPlaneswalker(),
				Effect:  nicolBolasPlaneswalkerUltimate,
			},
		},
	})
}

func nicolBolasPlaneswalkerUltimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var target, victim uuid.UUID
	for _, t := range ctx.LegalTargets() {
		target = t.ID
		break
	}
	if target == uuid.Nil {
		return nil
	}
	victim = target
	if g.PlayerByIDForEffect(target) == nil {
		c, ok := g.LookupCardForEffect(target)
		if !ok {
			return nil
		}
		victim = c.Controller
	}
	if err := (effects.DealDamage{Source: item.SourceCardID, Target: target, Amount: 7}).Apply(ctx); err != nil {
		return err
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   victim,
		Source:   item.SourceCardID,
		N:        7,
		Question: "Nicol Bolas, Planeswalker — discard seven cards",
	}, nicolBolasPlaneswalkerSacrifice, nil)
}

// nicolBolasPlaneswalkerSacrifice is "then sacrifices seven permanents of their
// choice". An on-disk key: never renamed, never reused.
var nicolBolasPlaneswalkerSacrifice = game.RegisterDiscardThen("choose-cards/nicol-bolas-planeswalker-sacrifice", func(g *game.Game, r game.DiscardsDone) error {
	g.PlayerSacrificesNForEffect(r.Source, r.Player, nil, "Nicol Bolas, Planeswalker — sacrifice seven permanents", 7)
	return nil
})
