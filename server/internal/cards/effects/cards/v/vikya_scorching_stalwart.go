package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vikya, Scorching Stalwart — Legendary Creature — Human Warrior {2}{W}, 2/4
// (Secret Lair Drop):
//
//	"Training
//	 {4}{R}, {Q}, Discard a card: Vikya deals damage equal to its power to any
//	 target. If excess damage was dealt to a creature this way, draw a card."
//
// {Q} is AbilityCost.UntapSelf (CR 107.6). Excess damage is CR 120.4a's, read
// off the damage that landed (B27DealDamageWithExcess); a player or
// planeswalker target has none. Vikya's power is last-known when she has gone.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad43ccce-3052-4817-8c46-f5c23cb49300",
		Name:         "Vikya, Scorching Stalwart",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Training("Vikya, Scorching Stalwart — training: put a +1/+1 counter on it"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{R}, {Q}, Discard a card: Vikya deals damage equal to its power to any target. If excess damage was dealt to a creature this way, draw a card.",
			Cost:    vikyaCost(),
			Targets: effects.TargetAny(),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				amount := effects.B29PowerAsItLastStood(g, item.SourceCardID)
				t := ts[0]
				if t.Kind == game.TargetCard {
					if c, ok := g.LookupCardForEffect(t.ID); ok && c.IsCreature() {
						excess, err := effects.B27DealDamageWithExcess(ctx, t.ID, amount)
						if err != nil || excess <= 0 {
							return err
						}
						return effects.DrawCards{N: 1}.Apply(ctx)
					}
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: amount}.Apply(ctx)
			},
		}},
	})
}

// vikyaCost is "{4}{R}, {Q}, Discard a card". Plus does not carry UntapSelf,
// so the discard component is built first and the mana and {Q} set on it.
func vikyaCost() game.AbilityCost {
	cost := effects.DiscardACard()
	cost.Mana = "{4}{R}"
	cost.UntapSelf = true
	return cost
}
