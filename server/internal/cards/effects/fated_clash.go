package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fated Clash — Sorcery {3}{W}{W} (the Squall deck request, #3918):
//
//	"You may cast this spell as though it had flash if a creature is
//	 attacking and a creature is blocking.
//	 Target creature you control and target creature an opponent
//	 controls each gain indestructible until end of turn. Then destroy
//	 all creatures."
//
// The flash clause is Timely Ward's and Silver Scrutiny's shape: an
// offer at the printed cost that opens the instant-speed window for its
// own claim (AlternativeCost.AsThoughFlash), gated by Condition. The
// condition is read as the cast begins (CR 601.3): some creature on the
// battlefield is attacking and some creature is blocking, whoever
// controls them. Outside that, the card is a sorcery.
//
// The two targets are separate clauses (CR 115.1d). Each still-legal one
// gains indestructible (CR 608.2b), so one that left in response does
// not stop the other's; then every creature is destroyed, and the
// indestructible ones survive (CR 702.12b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "610d9f58-4182-49cb-855c-9b6030e7f714",
		Name:         "Fated Clash",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		AlternativeCosts: []game.AlternativeCost{{
			Key:           "flash-if-attacking-and-blocking",
			Label:         "Cast with flash — a creature is attacking and a creature is blocking ({3}{W}{W})",
			ManaCost:      "{3}{W}{W}",
			AsThoughFlash: true,
			Condition:     fatedClashCreaturesInCombat,
		}},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"indestructible"},
					Label:    "Fated Clash — indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return DestroyAllMatching{Match: Creature()}.Apply(ctx)
		},
	})
}

// fatedClashCreaturesInCombat is "if a creature is attacking and a
// creature is blocking": both somewhere on the battlefield right now.
func fatedClashCreaturesInCombat(g *game.Game, _ uuid.UUID) bool {
	attacking, blocking := false, false
	for _, c := range g.BattlefieldCardsForEffect() {
		if !c.IsCreature() {
			continue
		}
		if c.AttackingTarget != uuid.Nil {
			attacking = true
		}
		if c.BlockingTarget != uuid.Nil {
			blocking = true
		}
	}
	return attacking && blocking
}
