package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elspeth Tirel — Legendary Planeswalker — Elspeth {3}{W}{W}, loyalty 4:
//
//	"+2: You gain 1 life for each creature you control.
//	 −2: Create three 1/1 white Soldier creature tokens.
//	 −5: Destroy all other permanents except for lands and tokens."
//
// The +2 counts creatures as it resolves (CR 608.2h), through the layered view.
// The −5 destroys every permanent but lands, tokens and Elspeth herself,
// planeswalkers, artifacts and enchantments included. All are destroyed
// together; indestructible and regeneration are honoured by the destroy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5f815e69-fdd6-4186-9317-25f5c7e0b08e",
		Name:            "Elspeth Tirel",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label: "+2: You gain 1 life for each creature you control.",
				Cost:  LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return GainLife{Player: item.Controller, Amount: n}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−2: Create three 1/1 white Soldier creature tokens.",
				Purpose: game.Purpose{Tokens: 3},
				Cost:    LoyaltyCost(-2),
				Effect:  createThreeSoldiers,
			},
			{
				Label:   "−5: Destroy all other permanents except for lands and tokens.",
				Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
				Cost:    LoyaltyCost(-5),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var doomed []game.Card
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.InstanceID == item.SourceCardID || c.IsLand() || IsToken(c) {
							continue
						}
						doomed = append(doomed, c)
					}
					for _, c := range doomed {
						if err := (DestroyTarget{Target: c.InstanceID}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}
