package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b15_helpers.go — shared bodies for the S82 reprint pass, batch 15
// (#5908). Keys are on-disk identities: never renamed, never reused.

// rprB15SpellbombPaid is Nihil Spellbomb's paid branch: draw a card.
var RprB15SpellbombPaid = PayUnlessBranches("pay-unless/nihil-spellbomb-draw",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}, nil)

// rprB15RestOnBottomAnyOrderTail is "put the rest on the bottom of your
// library in any order" as a PutFromLibraryOntoBattlefield rider.
var RprB15RestOnBottomAnyOrderTail = PutFromLibraryTail("put-from-library/rest-on-bottom-in-any-order-s82-rpr-b15",
	func(ctx *Context, res PutFromLibraryResult) error {
		return TakeRestOnBottomInAnyOrder(ctx.Game, TakeFromLibraryResult{Player: res.Player, Source: res.Source, Rest: res.Rest})
	})

// rprB15TargetExilesOwnGraveyardCard is Relic of Progenitus's "target
// player exiles a card from their graveyard": the TARGET player chooses,
// from their own graveyard, as the ability resolves.
func RprB15TargetExilesOwnGraveyardCard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var player uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			player = t.ID
			break
		}
	}
	if player == uuid.Nil {
		return nil
	}
	cards := GraveyardCardsOfPlayer(g, player)
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     ctx.Source(),
		Question:   "Relic of Progenitus — exile a card from your graveyard",
		Cards:      cards,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    rprB15RelicExiled,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

var rprB15RelicExiled = AskedPick("choose-cards/relic-of-progenitus-exile", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// rprB15ExileAllGraveyardsThenDraw is Relic of Progenitus's second
// ability: every seat's graveyard, yours too, then you draw a card.
func RprB15ExileAllGraveyardsThenDraw(g *game.Game, item *game.StackItem) error {
	if err := B02ExileAllGraveyards(g, item); err != nil {
		return err
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
}

// rprB15FateCounter is the kind Oblivion Stone stamps. The game has no
// constant for it: it is only a name on Card.Counters.
const RprB15FateCounter = "fate"

// rprB15DestroyUnfatedThenClear is Oblivion Stone's second ability: destroy
// each nonland permanent without a fate counter, then remove all fate
// counters from all permanents. The permanents to clear are read AFTER the
// destruction, so a survivor (indestructible, regenerated) loses its
// counter as well.
func RprB15DestroyUnfatedThenClear(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DestroyAllMatching{
		Match: And(Nonland(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.Counters[RprB15FateCounter] <= 0
		}),
	}).Apply(ctx); err != nil {
		return err
	}
	type fated struct {
		id uuid.UUID
		n  int
	}
	var all []fated
	for _, c := range g.BattlefieldCardsForEffect() {
		if n := c.Counters[RprB15FateCounter]; n > 0 {
			all = append(all, fated{c.InstanceID, n})
		}
	}
	for _, f := range all {
		if err := (AddCounter{Target: f.id, Kind: RprB15FateCounter, N: -f.n}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
