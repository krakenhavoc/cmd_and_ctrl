package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jacob Frye — Legendary Creature — Human Assassin {2}{B}, 3/2 (Assassin's
// Creed, #5948):
//
//	"Partner with Evie Frye (When this creature enters, target player may put
//	 Evie into their hand from their library, then shuffle.)
//	 Whenever one or more Assassins you control deal combat damage to a player,
//	 exile up to one target Assassin card or card with freerunning from your
//	 graveyard. If you do, copy it. You may cast the copy."
//
// Partner with is the shared entry trigger (PartnerWith). The damage trigger
// is once per player dealt combat damage by Assassins in a step
// (CR 603.2c; the creature type is read through the Jacob's text, CR 612.2),
// and its target is chosen as it goes on the stack: up to one card in the
// controller's graveyard that is an Assassin card or prints freerunning (read
// from the catalog row of the card, so a card with no catalog entry, which
// can't be cast for the cost, never counts). As it resolves the card is
// exiled, and only if it reached exile ("if you do") is it copied
// (CR 707.12), and the controller may cast that copy, paying its mana cost.
// A card that has left the graveyard in response is not exiled and nothing is
// copied (CR 608.2b).
//
// No simplification.
func init() {
	isAssassin := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasSubtypeWord("Assassin") }
	assassinOrFreerunning := Or(isAssassin, gapB36HasFreerunning)
	trigger := WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(isAssassin,
		"Jacob Frye — exile up to one target Assassin card or card with freerunning from your graveyard; copy it and you may cast the copy",
		jacobFryeExile)
	Register(Spec{
		OracleID:     "31af65c3-602b-4ac9-a25f-dccb88f8f751",
		Name:         "Jacob Frye",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Jacob Frye", "Evie Frye"),
			Targeting(trigger,
				TargetCardInGraveyard("up to one target Assassin card or card with freerunning from your graveyard",
					YouOwn(), assassinOrFreerunning).WithCount(0, 1)),
		},
	})
}

func jacobFryeExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(jacobFryeExiled)}.Apply(ctx)
	}
	return nil
}

// jacobFryeExiled is "if you do, copy it. You may cast the copy", run once
// the card has landed in exile. On-disk key: never renamed, never reused.
var jacobFryeExiled = CardTail("jacob-frye/copy-and-cast", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return CopyAndCast{Cards: o.IDs[:1], Player: ctx.Controller()}.Apply(ctx)
})
