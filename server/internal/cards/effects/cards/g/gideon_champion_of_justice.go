package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gideon, Champion of Justice — Legendary Planeswalker — Gideon
// {2}{W}{W}, starting loyalty 4:
//
//	"+1: Put a loyalty counter on Gideon for each creature target
//	 opponent controls.
//	 0: Until end of turn, Gideon becomes a Human Soldier creature with
//	 power and toughness each equal to the number of loyalty counters
//	 on him and gains indestructible. He's still a planeswalker.
//	 Prevent all damage that would be dealt to him this turn.
//	 −15: Exile all other permanents."
//
// THE +1 counts the target opponent's creatures as it resolves and puts
// that many loyalty counters on Gideon, if he is still on the
// battlefield. Zero creatures is zero counters (the +1 cost was paid
// already).
//
// THE 0 is animateGideon (gideon_animate.go; #2046) with
// SizeFromLoyalty (#2569). The number of loyalty counters is read once,
// when the ability resolves, and becomes a fixed base power and
// toughness until end of turn (CR 608.2h, layer 7b; ADR 0032 amendment
// of 2026-10-09). Loyalty he gains or loses later in the turn, from
// damage that can't be prevented or from a proliferate, does not resize
// him. He gains no colour.
//
// THE −15 exiles every permanent on the battlefield except Gideon,
// everyone's, lands and tokens included, read as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5843bf12-27db-4e7d-81ee-98874bea72c7",
		Name:         "Gideon, Champion of Justice",
		Completeness: effects.CompletenessFull,
		// The fallback for tokens, fixtures and the dev spawner; an
		// imported deck reads printed loyalty (ADR 0032 §1).
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Put a loyalty counter on Gideon for each creature target opponent controls.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
				Effect:  gideonChampionCountersPerCreature,
			},
			{
				Label: "0: Until end of turn, Gideon becomes a Human Soldier creature with power and toughness each equal to the number of loyalty counters on him and gains indestructible. He's still a planeswalker. Prevent all damage that would be dealt to him this turn.",
				Cost:  effects.LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.AnimateGideon(g, item, effects.GideonAnimation{
						Label:           "Gideon, Champion of Justice — a Human Soldier creature with indestructible until end of turn",
						Subtypes:        []string{"Human", "Soldier"},
						SizeFromLoyalty: true,
						Indestructible:  true,
					})
				},
			},
			{
				Label: "−15: Exile all other permanents.",
				Cost:  effects.LoyaltyCost(-15),
				Purpose: game.Purpose{Sweep: game.Sweep{
					Matches: game.SweepAllPermanents, How: game.SweepExile, Partial: true}},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.ExileAllMatching{Match: effects.OtherThan(item.SourceCardID)}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

// gideonChampionCountersPerCreature is the +1: one loyalty counter on
// Gideon for each creature the target opponent controls, counted as the
// ability resolves.
func gideonChampionCountersPerCreature(g *game.Game, item *game.StackItem) error {
	self := item.SourceCardID
	if !effects.OnBattlefield(g, self) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	ids := effects.LegalTargetIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	n := effects.CreaturesControlledBy(g, ids[0])
	if n == 0 {
		return nil
	}
	return effects.AddCounter{Target: self, Kind: game.CounterLoyalty, N: n}.Apply(ctx)
}
