package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stormbreath Dragon — 4/4 Dragon for {3}{R}{R}:
//
//	"Flying, haste, protection from white
//	 {5}{R}{R}: Monstrosity 3. (If this creature isn't monstrous, put
//	 three +1/+1 counters on it and it becomes monstrous.)
//	 When this creature becomes monstrous, it deals damage to each
//	 opponent equal to the number of cards in that player's hand."
//
// #1700's headline card. The hand sizes are read as the trigger
// RESOLVES, per opponent, so a player who casts or discards in
// response takes less — the count is taken when the damage is dealt
// (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7a770865-9383-45aa-883f-21ee649d1ea3",
		Name:            "Stormbreath Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste", "protection from white"},
		Activated:       []ActivatedAbility{Monstrosity(ManaCost("{5}{R}{R}"), 3)},
		Triggered: []game.TriggeredAbility{
			WhenBecomesMonstrous("Stormbreath Dragon — damage each opponent equal to their hand size", stormbreathDamagesByHandSize),
		},
	})
}

func stormbreathDamagesByHandSize(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		p := g.PlayerByIDForEffect(opp)
		if p == nil || p.Hand == nil {
			continue
		}
		if err := (DealDamage{Source: ctx.Source(), Target: opp, Amount: len(p.Hand.Cards)}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
