package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Those Old Scientists — Enchantment — Saga {2}{U} (Star Trek):
//
//	(As this Saga enters and after your draw step, add a lore counter.
//	Sacrifice after III.)
//	I — Draw two cards, then discard two cards.
//	II — Return up to one target artifact, instant, or sorcery card from
//	     your graveyard to your hand.
//	III — Choose up to two target cards in graveyards. The owners of
//	      those cards shuffle them into their libraries.
//
// Three chapter abilities (CR 714.2). Chapter I is Careful Study's body:
// the discard is asked after both draws, so the drawn cards are legal
// discards. Chapter II takes zero or one target. Chapter III's two targets
// may be in different graveyards; each card goes to its owner's library and
// each of those owners shuffles (CR 701.24c).
//
// No simplification.
func init() {
	const name = "Those Old Scientists"
	Register(Spec{
		OracleID:     "5ca4e8c4-8358-49be-bafb-85849bf9bd01",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Discards: 2},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "draw two cards, then discard two cards"),
				func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 2) }),
			ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "return up to one artifact, instant, or sorcery card to your hand"),
				TargetCardInGraveyard("up to one target artifact, instant, or sorcery card in your graveyard",
					YouOwn(), Or(Artifact(), Instant(), Sorcery())).WithCount(0, 1),
				returnTargetCardToHand),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "the owners of up to two target cards in graveyards shuffle them into their libraries"),
				TargetCardInGraveyard("up to two target cards in graveyards").WithCount(0, 2),
				trekB9ShuffleTargetsIntoOwnersLibraries),
		},
	})
}
