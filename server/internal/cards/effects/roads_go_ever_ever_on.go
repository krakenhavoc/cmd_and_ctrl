package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roads Go Ever, Ever On — Enchantment — Saga {1}{W} (The Hobbit, HOB):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I — Search your library for up to two basic Plains cards, exile them,
//	     then shuffle. You gain 2 life.
//	 II, III — Put a card exiled with this Saga into its owner's hand.
//	 IV — Whenever you attack this turn, target creature you control gets
//	      +1/+1 until end of turn for each Plains you control."
//
// "A card exiled with this Saga" is CR 607.2a's link, the one The Creation
// of Avacyn uses: chapter I links what it finds to this Saga OBJECT
// (SearchLibrary.LinkTo), and chapters II and III read it back with
// ExiledWithForEffect. A Saga that was bounced and replayed is a new object
// with no claim on the cards. With two cards left the controller chooses
// one; with one it is taken, and with none the chapter does nothing. The
// life is gained after the search has finished and the library is shuffled.
//
// Chapter IV creates a delayed triggered ability (CR 603.7): "whenever you
// attack this turn" is Dalkovan Encampment's DelayedOnEvent{Repeats,
// OncePerBatch}, once per attack declaration however many creatures attack
// (CR 603.2c). The Saga is gone by then (it is sacrificed after IV), and the
// ability does not care. The creature is targeted as the trigger goes on the
// stack (CR 603.3d) and the Plains are counted as it resolves.
//
// No simplification.
func init() {
	const name = "Roads Go Ever, Ever On"
	toHand := "put a card exiled with this Saga into its owner's hand"
	Register(Spec{
		OracleID:     "3cd3cdc9-2ed2-4cdf-b58a-12c3a2ef218b",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "search for up to two basic Plains cards, exile them, then shuffle; you gain 2 life"),
				Do(Instruction(roadsSearch), GainLife{Amount: 2})),
			ChapterTrigger(2, SagaChapterLabel(name, 2, toHand), roadsToHand),
			ChapterTrigger(3, SagaChapterLabel(name, 3, toHand), roadsToHand),
			ChapterTrigger(4, SagaChapterLabel(name, 4, "whenever you attack this turn, target creature you control gets +1/+1 until end of turn for each Plains you control"),
				func(g *game.Game, item *game.StackItem) error {
					return DelayedOnEvent{
						Label:        "Roads Go Ever, Ever On — target creature you control gets +1/+1 until end of turn for each Plains you control",
						On:           []game.EventKind{game.EventAttack},
						Condition:    youAttackCondition,
						Body:         roadsPumpBody,
						Repeats:      true,
						OncePerBatch: true,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// roadsSearch is chapter I's search: up to two basic Plains cards, exiled
// face up and linked to this Saga object.
func roadsSearch(ctx *Context) error {
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	return SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: IsBasicLandWithSubtype("plains"),
		Dest:      game.ZoneExile,
		Limit:     2,
		Shuffle:   true,
		LinkTo:    game.PermissionCardRef(ref),
		Reason:    "Roads Go Ever, Ever On — up to two basic Plains cards",
	}.Apply(ctx)
}

// roadsToHand is chapters II and III: one card exiled with this Saga goes
// to its owner's hand.
func roadsToHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	switch len(ids) {
	case 0:
		return nil
	case 1:
		return BounceToHand{Target: ids[0]}.Apply(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Roads Go Ever, Ever On — put a card exiled with this Saga into its owner's hand",
		Cards:    ids,
		Min:      1,
		Max:      1,
		ThenKey:  roadsPutInHand,
		LiveZone: game.ZoneExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// roadsPutInHand is the keyed answer to "which exiled card" (#4253). An
// on-disk key: never renamed, never reused.
var roadsPutInHand = game.RegisterChooseCardsThen("choose-cards/roads-go-ever-ever-on-put-in-hand", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || r.Asked == nil {
		return nil
	}
	return BounceToHand{Target: r.Picked[0]}.Apply(NewContext(g, r.Asked))
})

// roadsPumpBody is chapter IV's delayed trigger: its target clause is
// announced as each instance goes on the stack (ADR 0187 amendment).
var roadsPumpBody = game.TargetedDelayedBody("roads-go-ever-ever-on/pump-per-plains", simpleBody(roadsPump),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature you control", YouControl()) }))

// roadsPump gives the target +1/+1 for each Plains its controller controls
// as the trigger resolves; a target that has left is skipped (CR 608.2b).
func roadsPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n := len(permanentsControlledByMatching(g, item.Controller, Subtype("Plains")))
	if n == 0 {
		return nil
	}
	return BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: "Roads Go Ever, Ever On — +1/+1 for each Plains"}.Apply(ctx)
}
