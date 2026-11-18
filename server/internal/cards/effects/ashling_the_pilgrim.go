package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ashling the Pilgrim — Legendary Creature — Elemental Shaman {1}{R}, 1/1:
//
//	"{1}{R}: Put a +1/+1 counter on Ashling. If this is the third time
//	 this ability has resolved this turn, remove all +1/+1 counters from
//	 Ashling, and it deals that much damage to each creature and each
//	 player."
//
// The count is the per-turn tally of this Ashling's resolutions of this
// ability (Game.ResolvedThisTurn), which includes the one resolving, per
// object (CR 400.7). Only the third resolution blasts; the fourth and
// later only add a counter. "That much" is the number of +1/+1 counters
// actually removed, read after the new counter has been put on: with the
// counter this resolution adds, a fresh Ashling removes three. An Ashling
// that has left the battlefield (or come back as a new object) takes no
// counter, removes none and so deals no damage.
//
// Purpose: the third resolution is a damage sweep of at least 3 (the
// counters it adds are the least it removes).
//
// No simplification.
func init() {
	const label = "{1}{R}: Put a +1/+1 counter on Ashling. If this is the third time this ability has resolved this turn, remove all +1/+1 counters from Ashling, and it deals that much damage to each creature and each player."
	Register(Spec{
		OracleID:     "bc92f0ba-2cf0-453c-a0d3-98a8784fe55b",
		Name:         "Ashling the Pilgrim",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
			Cost:    ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id := item.SourceCardID
				if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield || sourceIsNewObject(g, item) {
					return nil
				}
				if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				if g.ResolvedThisTurn(id, label) != 3 {
					return nil
				}
				c, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				n := c.Counters[game.CounterPlusOne]
				if n <= 0 {
					return nil
				}
				if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: -n}).Apply(ctx); err != nil {
					return err
				}
				return b23DamageEachCreatureAndEachPlayer(ctx, n)
			},
		}},
	})
}
