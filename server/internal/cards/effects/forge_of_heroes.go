package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forge of Heroes — Land:
//
//	"{T}: Add {C}.
//	 {T}: Choose target commander that entered this turn. Put a +1/+1
//	 counter on it if it's a creature and a loyalty counter on it if
//	 it's a planeswalker."
//
// The second ability targets, so it uses the stack (it is not a mana
// ability, CR 605.1a). "A commander" is any player's (ACommander), and
// "entered this turn" is the per-object entry tally, so a commander that
// left and came back counts as its new object. The types are read at
// resolution: a creature planeswalker gets both counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77807103-bcd5-479f-bedd-f5d97aa6d3d2",
		Name:         "Forge of Heroes",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Choose target commander that entered this turn. Put a +1/+1 counter on it if it's a creature and a loyalty counter on it if it's a planeswalker.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target commander that entered this turn", ACommander(), rfCreatureCEnteredThisTurn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				c, ok := g.LookupCardForEffect(ts[0].ID)
				if !ok {
					return nil
				}
				if c.IsCreature() {
					if err := nothingIfGone(g.AddCounterByForEffect(item.Controller, c.InstanceID, game.CounterPlusOne, 1)); err != nil {
						return err
					}
				}
				if c.IsPlaneswalker() {
					return nothingIfGone(g.AddCounterByForEffect(item.Controller, c.InstanceID, game.CounterLoyalty, 1))
				}
				return nil
			},
		}},
	})
}
