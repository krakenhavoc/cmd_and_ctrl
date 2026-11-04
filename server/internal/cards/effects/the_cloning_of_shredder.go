package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Cloning of Shredder — Enchantment — Saga {4}{B}{B} (Teenage
// Mutant Ninja Turtles, #4457):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Exile target creature card from your graveyard. Create a token
//	 that's a copy of it, except it isn't legendary and is a Mutant in
//	 addition to its other types.
//	 II, III — Create a token that's a copy of a card exiled with this
//	 Saga, except it isn't legendary and is a Mutant in addition to its
//	 other types."
//
// Chapter I is The Scarab God's exile-then-copy (ExileTargetThenTokenCopy)
// with the exile LINKED to this Saga (CR 607.2a), so chapters II and III
// can find the card: the copy is made only when the card really reached
// exile. The exception is one CR 707.9b modification of the copiable
// type line: "Legendary" comes off (Miirym's) and Mutant is added beside
// the card's own subtypes (CR 707.9a's "in addition", the disturb
// helper's).
//
// Chapters II and III copy a card exiled with this Saga object, as it is
// in exile when the chapter resolves. A card that has left exile is not
// one, and with none there is nothing to copy. Chapter I can trigger
// again if lore counters are removed and added (CR 714.2b), so more than
// one card can be linked; then the controller picks which to copy.
//
// No simplification.
func init() {
	const name = "The Cloning of Shredder"
	copyExiled := func(n int) game.TriggeredAbility {
		return ChapterTrigger(n, SagaChapterLabel(name, n, "create a token that's a copy of a card exiled with this Saga"),
			cloningOfShredderCopyExiled)
	}
	Register(Spec{
		OracleID:     "bd768e8f-8f04-487e-85e9-2828d6fd8745",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "exile target creature card from your graveyard and create a token that's a copy of it"),
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return ExileTarget{Target: id, Linked: true, ThenTail: ctx.ItemTail(cloningOfShredderCopyTail)}.Apply(ctx)
				}),
			copyExiled(2),
			copyExiled(3),
		},
	})
}

// cloningOfShredderExcept is the printed exception on every chapter's
// copy: not legendary, and a Mutant in addition to its other types. A
// transforming copy takes it on both faces.
func cloningOfShredderExcept(t *game.Card) {
	t.TypeLine = withSubtypeInAddition(removeLegendaryFromTypeLine(t.TypeLine), "Mutant")
	for i := range t.Faces {
		t.Faces[i].TypeLine = withSubtypeInAddition(removeLegendaryFromTypeLine(t.Faces[i].TypeLine), "Mutant")
	}
}

// cloningOfShredderCopyTail is chapter I's "create a token that's a copy
// of it", run once the card has landed in exile. An on-disk key.
var cloningOfShredderCopyTail = ExileThenTokenCopyTail("then/the-cloning-of-shredder-copy",
	func(*Context) func(t *game.Card) { return cloningOfShredderExcept })

// cloningOfShredderCopyExiled is chapters II and III: a copy of the one
// card exiled with this Saga, or of the one its controller picks when
// there are several.
func cloningOfShredderCopyExiled(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	exiled := g.ExiledWithForEffect(game.PermissionCardRef(src))
	ids := make([]uuid.UUID, 0, len(exiled))
	for _, c := range exiled {
		ids = append(ids, c.InstanceID)
	}
	switch len(ids) {
	case 0:
		return nil
	case 1:
		return cloningOfShredderCopy(ctx, ids[0])
	}
	options := make([]game.ChoiceOption, len(ids))
	for i, c := range exiled {
		options[i] = game.ChoiceOption{Label: "Copy " + c.Name, Cards: []uuid.UUID{c.InstanceID}}
	}
	return PickOption{
		Question: "The Cloning of Shredder — create a token that's a copy of which card exiled with this Saga?",
		Options:  options,
		ThenKey:  cloningOfShredderPick,
		Carry:    ids,
	}.Apply(ctx)
}

// cloningOfShredderPick is the controller's pick among several linked
// cards. An on-disk key.
var cloningOfShredderPick = OptionPickThen("option-pick/the-cloning-of-shredder-copy", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	i := r.Index
	if i < 0 || i >= len(r.Carry) {
		i = 0
	}
	return cloningOfShredderCopy(ctx, r.Carry[i])
})

// cloningOfShredderCopy creates the excepted token copy of a card still
// in exile.
func cloningOfShredderCopy(ctx *Context, id uuid.UUID) error {
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	return CreateTokenCopy{Controller: ctx.Controller(), Copy: id, N: 1, Except: cloningOfShredderExcept}.Apply(ctx)
}
