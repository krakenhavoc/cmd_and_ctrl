package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Canopy Gargantuan — Creature — Dragon {5}{G}{G}, 7/7 (Tarkir: Dragonstorm
// Commander):
//
//	"Flying, ward {2}
//	 At the beginning of your upkeep, put a number of +1/+1 counters on each
//	 other creature you control equal to that creature's toughness."
//
// Ward is Aegis Sculptor's trigger. The upkeep trigger reads every other
// creature's toughness as it resolves, all of them before any counter is
// placed (CR 608.2: one effect, no creature's new toughness feeds another's
// count), then places each creature's counters as one placement
// (b11PutCountersOnEachCreatureYouControl's grouping, so a counter replacement
// applies once per creature). A creature that left in between gets none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ec1e69b1-f664-494b-8da2-c7199872b08b",
		Name:            "Canopy Gargantuan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Canopy Gargantuan — ward {2}"),
			AtYourUpkeep("Canopy Gargantuan — put +1/+1 counters on each other creature you control equal to its toughness",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					type pick struct {
						id uuid.UUID
						n  int
					}
					var picks []pick
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller != item.Controller || !c.IsCreature() || c.InstanceID == item.SourceCardID {
							continue
						}
						if t := c.CurrentToughness(); t > 0 {
							picks = append(picks, pick{c.InstanceID, t})
						}
					}
					for _, p := range picks {
						if !onBattlefield(g, p.id) {
							continue
						}
						if err := (AddCounter{Target: p.id, Kind: game.CounterPlusOne, N: p.n}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
