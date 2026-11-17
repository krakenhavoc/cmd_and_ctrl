package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Kenriths' Royal Funeral — Legendary Enchantment {2}{W}{B} (March of
// the Machine: The Aftermath):
//
//	"When The Kenriths' Royal Funeral enters, exile up to two target
//	 legendary creature cards from your graveyard. You draw X cards and
//	 you lose X life, where X is the greatest mana value among cards
//	 exiled this way.
//	 Legendary spells you cast cost {1} less to cast for each card exiled
//	 with The Kenriths' Royal Funeral."
//
// The enters trigger targets up to two legendary creature cards in your
// graveyard (re-checked as it resolves, CR 608.2b) and exiles them as one
// LINKED exile (Card.ExiledWith naming this object, CR 607.2a, as Angel of
// Serenity and Pinnacle Starcage do). Its tail reads the cards that really
// reached exile: X is the greatest mana value among them (an {X} in a cost
// is 0, CR 202.3e), and you draw X then lose X life; if none was exiled, X
// is 0 and nothing happens. The static is a cost reduction, "{1} less"
// for each card the Funeral's own object has exiled and which is still in
// exile (a card that left is no longer exiled with it, CR 400.7): it
// applies to legendary spells its controller casts from any zone, reducing
// generic mana only, never below {0} (CR 601.2f). A second Funeral counts
// only its own cards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4b0a4b7-b36f-4c96-b07b-a86c1efd333e",
		Name:         "The Kenriths' Royal Funeral",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("The Kenriths' Royal Funeral — exile up to two target legendary creature cards from your graveyard; draw X cards and lose X life, X the greatest mana value among them",
					kenrithsFuneralExile),
				TargetCardInGraveyard("up to two target legendary creature cards from your graveyard",
					Creature(), Legendary(), YouOwn()).WithCount(0, 2)),
		},
		CostModifiers: []game.CostModifier{
			CostsLessEach(kenrithsFuneralDiscount,
				"Legendary spells you cast cost {1} less to cast for each card exiled with The Kenriths' Royal Funeral.",
				YourSpell(), legendarySpell()),
		},
	})
}

// legendarySpell passes a spell that is legendary as it is cast.
func legendarySpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.IsLegendary() }
}

// kenrithsFuneralDiscount is the number of cards still exiled with the
// modifier's source object.
func kenrithsFuneralDiscount(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	return len(q.Game.ExiledWithForEffect(game.PermissionCardRef{ID: q.Source.InstanceID, Epoch: q.Source.ObjectEpoch}))
}

func kenrithsFuneralExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	then := ctx.ItemTail(kenrithsFuneralTail)
	if len(ids) == 0 {
		return then.Run(g, nil)
	}
	return g.ExileLinkedKeyedForEffect(ref, ids, then)
}

// kenrithsFuneralTail draws and loses X once the exile has landed. An
// on-disk key: never renamed, never reused.
var kenrithsFuneralTail = CardTail("the-kenriths-royal-funeral/draw-and-lose", func(ctx *Context, o game.TailOutcome) error {
	x := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			if mv := c.ManaValueWithX(0); mv > x {
				x = mv
			}
		}
	}
	if x <= 0 {
		return nil
	}
	if err := (DrawCards{N: x}).Apply(ctx); err != nil {
		return err
	}
	return ctx.Game.ChangePlayerLifeKeyedForEffect(ctx.Source(), ctx.Controller(), -x, game.KeyedTail{})
})
