package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tamiyo, Collector of Tales — Legendary Planeswalker — Tamiyo
// {2}{G}{U}, starting loyalty 5:
//
//	"Spells and abilities your opponents control can't cause you to
//	 discard cards or sacrifice permanents.
//	 +1: Choose a nonland card name, then reveal the top four cards of
//	     your library. Put all cards with the chosen name from among
//	     them into your hand and the rest into your graveyard.
//	 −3: Return target card from your graveyard to your hand."
//
// The static is `OpponentEffectProtections` (#2178): read off the
// battlefield whenever an opponent's resolving spell or ability asks
// this player to discard or sacrifice, so an edict skips its controller
// and an opponent's "each player discards" passes over them. A cost
// they chose to pay is never asked, and their own Fleshbag Marauder
// still bites.
//
// The +1 asks for a name at resolution (QueueCardNameChoiceKeyedForEffect)
// and carries the rest of the sentence in its continuation, a registered
// tail that is told the name as its Word: nothing is
// revealed until the name is in. The name is free text (CR 201.2 admits
// any card name), so "nonland" is enforced where the engine can: a land
// card is never a match, which is what naming a land would be if it were
// legal, and so nothing is gained by naming one. The cards that match go
// to hand as one move and the rest are PUT into the graveyard (not
// milled, so a mill-doubler does not apply) through TakeRestIntoGraveyard.
func init() {
	Register(Spec{
		OracleID:                  "75d56a0a-2f64-4e80-b83e-85942d3e6dd7",
		Name:                      "Tamiyo, Collector of Tales",
		Completeness:              CompletenessFull,
		StartingLoyalty:           5,
		OpponentEffectProtections: CantBeMadeToDiscardOrSacrifice(),
		Activated: []ActivatedAbility{
			{
				Label:  "+1: Choose a nonland card name, then reveal the top four cards of your library. Put all cards with the chosen name from among them into your hand and the rest into your graveyard.",
				Cost:   LoyaltyCost(1),
				Effect: tamiyoCollectorPlusOne,
			},
			{
				Label:   "−3: Return target card from your graveyard to your hand.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetCardInGraveyard("target card in your graveyard", YouOwn()),
				Effect:  returnTargetedGraveyardCardsToHand,
			},
		},
	})
}

// tamiyoCollectorPlusOne asks for the name; the rest of the +1 is a
// registered tail told the name.
func tamiyoCollectorPlusOne(g *game.Game, item *game.StackItem) error {
	player, source := item.Controller, item.SourceCardID
	g.QueueCardNameChoiceKeyedForEffect(player, source,
		"Tamiyo, Collector of Tales — choose a nonland card name",
		game.KeyedTail{Ref: tamiyoNamed, Player: player, Source: source})
	return nil
}

// tamiyoNamed is the rest of the +1 once the name (the tail's Word) is
// known. An on-disk key.
var tamiyoNamed = CardTail("then/tamiyo-collector-named", func(ctx *Context, o game.TailOutcome) error {
	player, source, name := o.Tail.Player, o.Tail.Source, o.Tail.Word
	const label = "Tamiyo, Collector of Tales — reveal the top four cards of your library"
	cards := ctx.Game.RevealTopOfLibraryForEffect(player, source, 4, label)
	match := func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return !c.IsLand() && game.CardNameMatches(c, name)
	}
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  cards,
		Match:  match,
		All:    true,
		Label:  label,
		Key:    takeRestGraveyard,
	}.Apply(ctx)
})
