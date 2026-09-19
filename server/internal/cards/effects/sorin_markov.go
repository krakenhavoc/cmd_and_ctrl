package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sorin Markov — Legendary Planeswalker — Sorin {3}{B}{B}{B},
// starting loyalty 4 (EDHREC rank 4141):
//
//	"+2: Sorin Markov deals 2 damage to any target and you gain 2
//	     life.
//	 −3: Target opponent's life total becomes 10.
//	 −7: You control target player during that player's next turn."
//
// Six mana, and the −3 is what you pay it for. In a format that
// starts at 40, "becomes 10" is the largest single swing any card in
// this batch produces: it takes thirty life off the player who has
// been gaining it all game, and it can be used again two turns later
// because the +2 climbs faster than the −3 falls. The +2 is a real
// ability rather than filler — 2 damage and 2 life on a six-mana
// walker keeps him alive against a 2/2.
//
// "BECOMES 10" IS A SET, NOT A LOSS. CR 118.5 makes it a life change
// of whatever size closes the gap, so a player at 4 GAINS 6 and a
// player at 40 LOSES 30. That direction is printed and it matters:
// against a player who has been drained to single digits the −3 is a
// gift, which is why b31LifeBecomes computes the delta rather than
// assuming a loss. Because it is a life change and not damage, life
// loss triggers (a Vito, a Wound Reflection) see it and damage
// prevention does not.
//
// The +2 uses the widest target clause — "any target" is a player, a
// creature, a planeswalker or a battle — and the life gain happens
// whether or not the damage did (a fogged or prevented 2 still gains
// the 2).
//
// THE −7 IS NOT REGISTERED, and that is a deliberate omission rather
// than a stub, the same one Elspeth, Sun's Champion and Teferi, Hero
// of Dominaria declare. Controlling another player during their turn
// means playing from their hand, making their attacks and their
// blocks; the engine has one seat per player and no shape for a
// player acting as another. An ability whose label promised the
// takeover and delivered only a loyalty payment would be a worse lie
// than an ability that is not offered. The loyalty still accrues past
// 7; the day player control lands, this card gains one entry and
// nothing else changes.
func init() {
	Register(Spec{
		OracleID:     "c151e8b7-4b28-4f03-8a81-9bf623f893d7",
		Name:         "Sorin Markov",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The -7 ultimate isn't offered — controlling another player's turn doesn't exist yet."},
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Sorin Markov deals 2 damage to any target and you gain 2 life.",
				Cost:    LoyaltyCost(2),
				Targets: TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					// CR 608.2b: a target that left in response is
					// skipped. The life gain is NOT conditional on the
					// damage landing — it is a separate sentence.
					if legal := ctx.LegalTargets(); len(legal) > 0 {
						if err := (DealDamage{Source: item.SourceCardID, Target: legal[0].ID, Amount: 2}).Apply(ctx); err != nil {
							return err
						}
					}
					return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
				},
			},
			{
				Label:   "−3: Target opponent's life total becomes 10.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetPlayer("target opponent", Opponent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						return b31LifeBecomes(ctx, t.ID, 10)
					}
					return nil
				},
			},
		},
	})
}
