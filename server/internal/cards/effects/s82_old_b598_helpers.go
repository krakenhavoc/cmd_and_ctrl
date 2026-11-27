package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b598_helpers.go — shared bodies for Betrayers of Kamigawa,
// batch 598 (#5602).

// TargetOpponentExilesTopYouMayPlay is the Effect body for "Target
// opponent exiles the top card of their library. You may play that card
// this turn" (Knacksaw Clique, Ornate Kanzashi): the opponent's top card is
// exiled and the ability's controller, not the owner, gets the play
// permission until end of turn. A target that has left the game does
// nothing (CR 608.2b).
func TargetOpponentExilesTopYouMayPlay(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opp := FirstLegalTargetPlayer(ctx)
	if opp == uuid.Nil {
		return nil
	}
	return ExileTopWithPermission{From: opp, GrantTo: item.Controller, N: 1}.Apply(ctx)
}

// SwapEachPlayersCreatures is Reins of Power's and Twist Allegiance's
// "You and target opponent each gain control of all creatures the other
// controls until end of turn. Untap those creatures. Those creatures gain
// haste until end of turn": both sets are snapshotted from who controls
// what before anything moves, so the untap and the swap read the original
// board. The swap is two runs of GainControl, because ExchangeControl is
// CR 701.12's all-or-nothing single pair and this clause exchanges two
// disjoint sets of any size. name is the card's, for the effect labels.
func SwapEachPlayersCreatures(ctx *Context, name string) error {
	victim, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	caster := ctx.Controller()
	mine := MatchingBattlefield(ctx, And(Creature(), YouControl()))
	theirs := MatchingBattlefield(ctx, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.Controller == victim
	})
	for _, set := range [][]game.Card{mine, theirs} {
		for _, c := range set {
			if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	for _, move := range []struct {
		creatures []game.Card
		to        uuid.UUID
	}{{mine, victim}, {theirs, caster}} {
		for _, c := range move.creatures {
			if err := (GainControl{
				Target:     c.InstanceID,
				Controller: move.to,
				Duration:   DurationUntilEndOfTurn(ctx),
				Label:      name + " — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		for _, c := range move.creatures {
			if err := (GrantKeywordUntilEOT{
				Target:   c.InstanceID,
				Keywords: []string{"haste"},
				Label:    name + " — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// RevealTopXTakeMatching is "Reveal the top X cards of your library. Put
// all <match> cards revealed this way into your hand and the rest on the
// bottom of your library in any order" with X read from the spell
// (Enshrined Memories; Lair Delve is the same with two cards).
func RevealTopXTakeMatching(name string, match CardPredicate) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		player := item.Controller
		revealed := ctx.Game.RevealTopOfLibraryForEffect(player, item.SourceCardID, ctx.X(), name+" — reveal the top X cards of your library")
		return TakeFromLibraryToHand{
			Player: player,
			Cards:  revealed,
			Match:  match,
			All:    true,
			Label:  name + " — put all the cards revealed this way into your hand, the rest on the bottom",
			Key:    TakeRestBottomAny,
		}.Apply(ctx)
	}
}
