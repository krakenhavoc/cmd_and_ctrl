package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "0b2481d6-97a0-4f92-b7e5-3451863fd7fb",
		Name:            "Archangel Elspeth",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Create a 1/1 white Soldier creature token with lifelink.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("1/1 white Soldier with lifelink"),
						N:          1,
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−2: Put two +1/+1 counters on target creature. It becomes an Angel in addition to its other types and gains flying.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
							return err
						}
						return ScopedEffectFor{
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
				Cost:  LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					match := And(NonlandPermanentCard(), ManaValueLE(3))
					return ReturnFromGraveyardTogether{
						Targets: graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
							return match(ctx.Game, ctx.Controller(), c)
						}),
						Controller: ctx.Controller(),
					}.Apply(ctx)
				},
			},
		},
	})
}
