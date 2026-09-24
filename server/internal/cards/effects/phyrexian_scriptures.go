package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Scriptures — Enchantment — Saga for {2}{B}{B}:
//
//	"I — Put a +1/+1 counter on up to one target creature. That
//	     creature becomes an artifact in addition to its other types.
//	 II — Destroy all nonartifact creatures.
//	 III — Exile all opponents' graveyards."
//
// SANDBOX SIMPLIFICATION, and it makes the card WEAKER than printed:
// chapter I places the counter but does NOT grant the artifact type.
// The grant is a continuous effect with no duration — it lasts for
// as long as the creature is on the battlefield, outliving the Saga
// that made it — and the engine's only floating-continuous-effect
// registry (game.TurnScopedStatics, ADR 0035) expires at cleanup.
// Registering it there would end the protection one turn later and
// silently, which is worse than not having it; a permanent-duration
// registry is its own piece of work.
//
// The consequence is exactly the one the card is played for: the
// creature you meant to save from chapter II is not saved. Weaker,
// never stronger — the #259 direction. Chapter II is unchanged and
// still spares creatures that are artifacts for any OTHER reason,
// because it reads the effective type line.
//
// Chapter III exiles opponents' graveyards, not every graveyard —
// yours survives, which is the whole reason the card sits in
// graveyard-value decks.
func init() {
	Register(Spec{
		OracleID:     "11173ad3-c007-478f-bce0-d756eac07ccb",
		Name:         "Phyrexian Scriptures",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Chapter I doesn't make the chosen creature an artifact, so chapter II doesn't spare it."},
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, "Phyrexian Scriptures — I: +1/+1 counter on up to one creature",
				TargetCreature("up to one target creature").WithCount(0, 1),
				putPlusOneCounterOnEachLegalTarget),
			ChapterTrigger(2, "Phyrexian Scriptures — II: destroy all nonartifact creatures",
				scripturesWipeNonartifacts),
			ChapterTrigger(3, "Phyrexian Scriptures — III: exile all opponents' graveyards",
				scripturesExileOpponentGraveyards),
		},
	})
}

func scripturesWipeNonartifacts(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var doomed []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && !c.IsArtifact() {
			doomed = append(doomed, c)
		}
	}
	for _, c := range doomed {
		if err := (DestroyTarget{Target: c.InstanceID}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

func scripturesExileOpponentGraveyards(g *game.Game, item *game.StackItem) error {
	for _, p := range g.Seats {
		if p.ID == item.Controller || p.Graveyard == nil {
			continue
		}
		// Snapshot the IDs: exiling moves cards out of the slice the
		// loop would otherwise be walking.
		ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
		for _, c := range p.Graveyard.Cards {
			ids = append(ids, c.InstanceID)
		}
		for _, id := range ids {
			if err := g.ExileCardForEffect(id); err != nil {
				return err
			}
		}
	}
	return nil
}
