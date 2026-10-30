package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Titania, Rugged Rumbler — Legendary Creature — Human Villain
// {2}{B/G}, 5/5:
//
//	"As an additional cost to cast this spell, discard a card or pay {2}.
//	 Ward—Discard a card or pay {2}."
//
// The cast cost is ADR 0100 §2's either/or additional cost, Lightning
// Axe's shape. The ward is #4566's either/or ward (WardEither): the
// targeting player is asked once, names the branch, and a decline — or a
// player who can pay neither — counters the spell or ability (CR 702.21a,
// CR 118.12a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e380e37d-926b-4a4b-a275-7844bf4956d5",
		Name:         "Titania, Rugged Rumbler",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			DiscardCost(1).Keyed("discard"),
			ManaAdditionalCost("{2}").Keyed("mana"),
		),
		Triggered: []game.TriggeredAbility{
			Ward(WardEither(WardDiscard(1), WardMana("{2}")), "Titania, Rugged Rumbler — ward, discard a card or pay {2}"),
		},
	})
}
