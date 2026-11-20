package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archangel Elspeth — Legendary Planeswalker — Elspeth {2}{W}{W}, starting
// loyalty 4 (March of the Machine):
//
//	"+1: Create a 1/1 white Soldier creature token with lifelink.
//	 −2: Put two +1/+1 counters on target creature. It becomes an Angel in
//	     addition to its other types and gains flying.
//	 −6: Return all nonland permanent cards with mana value 3 or less from
//	     your graveyard to the battlefield."
//
// The −2's type and flying last indefinitely (no duration printed): a
// scoped effect pinned to the creature as that object (Ascent of the
// Worthy's Angel), so a creature that leaves and returns is a new object
// and loses both. The −6 puts every card in at once (Return to the Light
// Realms' return together), the mana value read as each card sits in the
// graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b2481d6-97a0-4f92-b7e5-3451863fd7fb",
		Name:            "Archangel Elspeth",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Create a 1/1 white Soldier creature token with lifelink.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{
						Controller: item.Controller,
						Template:   effects.TokenCard("1/1 white Soldier with lifelink"),
						N:          1,
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−2: Put two +1/+1 counters on target creature. It becomes an Angel in addition to its other types and gains flying.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
							return err
						}
						return effects.ScopedEffectFor{
							Target:   t.ID,
							Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Angel")...), game.AddKeywordsMod("flying")},
							Duration: g.PinnedTo(game.IndefiniteDuration(), t.ID),
							Label:    "Archangel Elspeth — it becomes an Angel in addition to its other types and gains flying",
						}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label: "−6: Return all nonland permanent cards with mana value 3 or less from your graveyard to the battlefield.",
				Cost:  effects.LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					match := effects.And(effects.NonlandPermanentCard(), effects.ManaValueLE(3))
					return effects.ReturnFromGraveyardTogether{
						Targets: effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
							return match(ctx.Game, ctx.Controller(), c)
						}),
						Controller: ctx.Controller(),
					}.Apply(ctx)
				},
			},
		},
	})
}
