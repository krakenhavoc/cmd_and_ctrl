package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aang and La, Ocean's Fury — Legendary Creature — Avatar Spirit Ally,
// 5/5 (colorless):
//
//	"Reach, trample
//	 Whenever Aang and La attack, put a +1/+1 counter on each tapped
//	 creature you control."
//
// The BACK face of Aang, Swift Savior ("<oracle_id>#1", ADR 0034),
// reached by transforming Aang IN PLACE through the Waterbend {8}
// ability declared in aang_swift_savior.go. That is a genuine first
// for the catalog: every other back face landing in this batch is
// reached through a Saga's exile-and-return (a NEW object, CR 400.7);
// this one is CR 712.18's same-object verb, so Aang's counters,
// damage, attachments and CR 613.7 timestamp all ride along, and
// nothing about the flip is an entry — no ETB fires, and a creature
// that was or wasn't summoning sick stays that way.
//
// The counter body is b13PutCounterOnEach's shape (one +1/+1 per
// listed permanent, re-checking each is still on the battlefield
// before placing it — a creature that leaves in response to the
// attack trigger gets nothing), with the candidate set narrowed to
// "tapped creature you control" rather than carried in as a fixed
// list, since the printed clause names the condition rather than a
// prior instruction's targets.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        aangSwiftSaviorOracle + "#1",
		Name:            "Aang and La, Ocean's Fury",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(
				"Aang and La, Ocean's Fury — a +1/+1 counter on each tapped creature you control",
				aangAndLaCounterEachTappedCreature),
		},
	})
}

// aangAndLaCounterEachTappedCreature puts one +1/+1 counter on each
// tapped creature the item's controller controls. Snapshots the set
// first so a creature tapped or created by an earlier placement's
// trigger mid-loop is not double-counted, mirroring
// b11PutCountersOnEachCreatureYouControl's own caution.
func aangAndLaCounterEachTappedCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.Tapped {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (AddCounter{Target: id, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
