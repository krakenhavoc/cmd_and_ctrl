package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Court of Locthwain — Enchantment {2}{B}{B} (#1729):
//
//	"When this enchantment enters, you become the monarch.
//	 At the beginning of your upkeep, exile the top card of target
//	 opponent's library. You may play that card for as long as it
//	 remains exiled, and mana of any type can be spent to cast it. If
//	 you're the monarch, until end of turn, you may cast a spell from
//	 among cards exiled with this enchantment without paying its mana
//	 cost."
//
// The upkeep trigger makes TWO permissions, and they are different in
// kind:
//
//   - The card it exiles may be PLAYED, for as long as it stays in
//     exile, with mana of any type (ExileTopWithPermission, the Hostage
//     Taker grant without CastOnly — a land can be played).
//   - If you are the monarch as the trigger resolves (a clause of the
//     effect, CR 608.2, like every Court's), you may cast ONE spell from
//     among every card exiled with this Court so far, free, until end
//     of turn. One permission over the set, locked as it resolves (CR
//     611.2c), with CastsLeft 1 (#1729): the first spell cast with it
//     spends it, however many cards it names.
//
// Every card in that set also holds the first permission, so a card can
// be opened by both, and the caster chooses: the free cast is claimed as
// an alternative cost of its own ("court_of_locthwain",
// CastPermissionForClaimLocked), and a cast that does not claim it pays
// with the play permission and leaves the free one for another card.
// "Without paying its mana cost" is the {0} price: X is 0 (CR 107.3b),
// and additional costs are still paid, with mana of any type, because
// the card was exiled with that clause.
//
// "Cards exiled with this enchantment" is the event log's record of
// this Court's upkeep exile (b27ExiledWith, keyed on the trigger's
// label), so a Court that left and came back is a new object with
// nothing exiled with it yet (CR 400.7), and a card that has left exile
// is no longer among them.
//
// The ruling of 2023-09-01: "You may play the exiled cards (and spend
// mana of any type to do so) even if Court of Locthwain leaves the
// battlefield." Both permissions are stored on the player, not derived
// from the Court, so they outlive it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6106ba65-f07b-41ec-8b59-e052bbde5529",
		Name:         "Court of Locthwain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Court of Locthwain"),
			Targeting(AtYourUpkeep(courtOfLocthwainUpkeepLabel, courtOfLocthwainUpkeep),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

const courtOfLocthwainUpkeepLabel = "Court of Locthwain — exile the top card of target opponent's library; if you're the monarch, cast a spell exiled with it for free"

// courtOfLocthwainFreeCastKey is the alternative cost the free cast is
// claimed under. An on-the-wire identity: never renamed.
const courtOfLocthwainFreeCastKey = "court_of_locthwain"

func courtOfLocthwainUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim := uuid.Nil
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			victim = t.ID
			break
		}
	}
	if victim == uuid.Nil {
		return nil
	}
	if err := (ExileTopWithPermission{
		From: victim, GrantTo: item.Controller, N: 1, AnyType: true, WhileExiled: true,
	}).Apply(ctx); err != nil {
		return err
	}
	if !YoureTheMonarch(g, item.Controller) {
		return nil
	}
	var cards []game.Card
	for _, id := range b27ExiledWith(g, item.SourceCardID, courtOfLocthwainUpkeepLabel) {
		if c, ok := g.LookupCardForEffect(id); ok {
			cards = append(cards, c)
		}
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     item.Controller,
		Zone:       game.ZoneExile,
		CastOnly:   true,
		CastsLeft:  1,
		AltCostKey: courtOfLocthwainFreeCastKey,
		Cost:       "{0}",
		AnyColor:   true,
		AnyType:    true,
		Source:     item.SourceCardID,
		SourceName: "Court of Locthwain",
		Label:      "Court of Locthwain — cast it without paying its mana cost",
	}, cards)
	return nil
}
