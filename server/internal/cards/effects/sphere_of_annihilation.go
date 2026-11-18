package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sphere of Annihilation — Artifact {X}{B} (Adventures in the Forgotten
// Realms):
//
//	"This artifact enters with X void counters on it.
//	 At the beginning of your upkeep, exile this artifact, all creatures
//	 and planeswalkers with mana value less than or equal to the number
//	 of void counters on it, and all creature and planeswalker cards in
//	 graveyards with mana value less than or equal to the number of void
//	 counters on it."
//
// The counters are read as the ability resolves, from last-known
// information if the Sphere has left (CR 608.2h). The sets are fixed
// before anything is exiled, and the Sphere itself goes with them
// whatever its own mana value. A permanent or card whose {X} is in its
// cost counts X as 0 off the stack (CR 202.3e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "b96399f9-c0ba-4507-be13-65cf716b7757",
		Name:                       "Sphere of Annihilation",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		Purpose:                    game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters("void")},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Sphere of Annihilation — exile it, all creatures and planeswalkers with mana value ≤ its void counters, and all such cards in graveyards",
				sphereOfAnnihilationExile),
		},
	})
}

func sphereOfAnnihilationExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := 0
	if s, ok := ctx.SourcePermanent(); ok {
		n = s.Counters["void"]
	}
	match := And(Or(Creature(), Planeswalker()), ManaValueLE(n))
	var ids []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, match) {
		ids = append(ids, c.InstanceID)
	}
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if match(g, item.Controller, c) {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	if src := ctx.Source(); src != uuid.Nil {
		ids = append(ids, src)
	}
	seen := map[uuid.UUID]bool{}
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
