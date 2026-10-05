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
// The +1 asks for a name at resolution (QueueCardNameChoiceThenForEffect)
// and carries the rest of the sentence in its continuation: nothing is
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

// tamiyoCollectorPlusOne asks for the name, then reveals and takes. A
// package-level func capturing only scalars, as every continuation must.
func tamiyoCollectorPlusOne(g *game.Game, item *game.StackItem) error {
	player, source := item.Controller, item.SourceCardID
	g.QueueCardNameChoiceThenForEffect(player, source,
		"Tamiyo, Collector of Tales — choose a nonland card name",
		tamiyoRevealAndTake(player, source))
	return nil
}

// tamiyoRevealAndTake is the rest of the +1 once the name is known.
func tamiyoRevealAndTake(player, source uuid.UUID) func(g *game.Game, name string) error {
	return func(g *game.Game, name string) error {
		const label = "Tamiyo, Collector of Tales — reveal the top four cards of your library"
		cards := g.RevealTopOfLibraryForEffect(player, source, 4, label)
		match := func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return !c.IsLand() && game.CardNameMatches(c, name)
		}
		return TakeFromLibraryToHand{
			Player: player,
			Cards:  cards,
			Match:  match,
			All:    true,
			Label:  label,
			Then:   TakeRestIntoGraveyard,
		}.Apply(NewContext(g, &game.StackItem{Controller: player, SourceCardID: source}))
	}
}
