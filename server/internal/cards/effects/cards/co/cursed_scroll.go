package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cursed Scroll — Artifact {1} (Tempest, S82 OLD batch 757, #5628):
//
//	"{3}, {T}: Choose a card name, then reveal a card at random from your
//	 hand. If that card has the chosen name, this artifact deals 2 damage
//	 to any target."
//
// The target is chosen as the ability goes on the stack (CR 601.2c), before
// the name is. Naming is free text (CR 201.2), asked as it resolves; the
// random card comes from the game's keyed random stream and is revealed to
// the table (Planeswalker's Mischief's reveal). An empty hand reveals
// nothing and deals no damage. A target that has become illegal takes none
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "514752ea-d453-41d4-97cf-b1dcdb4b4a08",
		Name:         "Cursed Scroll",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Choose a card name, then reveal a card at random from your hand. If that card has the chosen name, this artifact deals 2 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				g.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
					"Cursed Scroll — choose a card name", ctx.ItemTail(cursedScrollNamed))
				return nil
			},
		}},
	})
}

var cursedScrollNamed = effects.CardTail("card-name/s82-old-b757-cursed-scroll", func(ctx *effects.Context, o game.TailOutcome) error {
	id, ok := effects.OldB639RevealRandomFromHand(ctx, ctx.Controller(), "Cursed Scroll — a card at random from your hand")
	if !ok {
		return nil
	}
	c, found := ctx.Game.LookupCardForEffect(id)
	if !found || !game.CardNameMatches(c, o.Tail.Word) {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
	}
	return nil
})
