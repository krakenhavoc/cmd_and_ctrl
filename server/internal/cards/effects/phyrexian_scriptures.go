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
// Chapter I's artifact grant has no stated duration (CR 611.2a), so
// it is a ScopedEffectFor with IndefiniteDuration pinned to the creature
// (the Sealock Monster shape): it outlives the Saga and ends only when
// the creature leaves the battlefield and comes back a new object
// (CR 400.7). Chapter II reads the effective type line, so the grant is
// what spares the chosen creature. Until #2155 this was a declared gap
// (the only registry then expired at cleanup); IndefiniteDuration
// closed it.
//
// Chapter III exiles opponents' graveyards, not every graveyard —
// yours survives, which is the whole reason the card sits in
// graveyard-value decks.
func init() {
	Register(Spec{
		OracleID:     "11173ad3-c007-478f-bce0-d756eac07ccb",
		Name:         "Phyrexian Scriptures",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, "Phyrexian Scriptures — I: +1/+1 counter on up to one creature",
				TargetCreature("up to one target creature").WithCount(0, 1),
				scripturesChapterOne),
			TriggerWithPurpose(ChapterTrigger(2, "Phyrexian Scriptures — II: destroy all nonartifact creatures",
				scripturesWipeNonartifacts), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
			ChapterTrigger(3, "Phyrexian Scriptures — III: exile all opponents' graveyards",
				scripturesExileOpponentGraveyards),
		},
	})
}

// scripturesChapterOne puts the +1/+1 counter on the chosen creature
// and makes it an artifact in addition to its other types for as long
// as it stays on the battlefield.
func scripturesChapterOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		if err := (ScopedEffectFor{
			Target:   t.ID,
			Mods:     []game.Mod{game.AddTypesMod("Artifact")},
			Duration: g.PinnedTo(game.IndefiniteDuration(), t.ID),
			Label:    "Phyrexian Scriptures — it becomes an artifact in addition to its other types",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
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
