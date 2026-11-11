package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Evercoat Ursine — Creature — Elemental Bear {4}{G}, 6/5 (Bloomburrow
// Commander, #5368):
//
//	"Trample
//	 Hideaway 3, hideaway 3 (When this creature enters, look at the top three
//	 cards of your library, exile one face down, then put the rest on the
//	 bottom in a random order. Then do it again.)
//	 Whenever this creature deals combat damage to a player, if there are
//	 cards exiled with it, you may play one of them without paying its mana
//	 cost."
//
// Two Hideaway 3 abilities, each the shared Hideaway: both link their card to
// the Ursine OBJECT (CR 607.2a), so the permanent has up to two hidden cards
// at once and the second ability's exile does not displace the first's. The
// combat-damage trigger has an intervening "if" (CR 603.4): it is read as the
// trigger fires and again as it resolves, against the cards still hidden by
// THIS object. "One of them" is a choice among the hidden cards (skipped when
// there is one), floor zero for "you may"; the pick is played the way a
// single hideaway card is — a land now with the turn's land drop, a spell as
// a {0} grant (ADR 0091). The other stays hidden for a later hit.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "65c32889-3d70-449f-af76-05385c78c5a9",
		Name:            "Evercoat Ursine",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Hideaway("Evercoat Ursine", 3),
			Hideaway("Evercoat Ursine", 3),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: AllOf(ThisDealtCombatDamageToAPlayer,
					func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return len(g.HiddenCardsForEffect(game.ObjectRefOf(*source))) > 0
					}),
				Key: evercoatUrsineLabel,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, evercoatUrsineLabel)
					item.Params.Object = game.ObjectRef(game.ObjectRefOf(*source))
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ref := game.PermissionCardRef(item.Params.Object)
					hidden := g.HiddenCardsForEffect(ref) // CR 603.4: the intervening if again
					switch len(hidden) {
					case 0:
						return nil
					case 1:
						cmdr2S50PlayHiddenCard(g, item.Controller, ref, hidden[0], "Evercoat Ursine — play the exiled card without paying its mana cost")
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:   item.Controller,
						Source:    item.SourceCardID,
						Question:  "Evercoat Ursine — you may play one of the exiled cards without paying its mana cost",
						Cards:     hidden,
						Min:       0,
						Max:       1,
						LiveZone:  game.ZoneExile,
						ThenKey:   evercoatUrsinePicked,
						Carry:     []uuid.UUID{ref.ID},
						CarryNums: []int{ref.Epoch},
					})
					return nil
				},
			},
		},
	})
}

const evercoatUrsineLabel = "Evercoat Ursine — you may play one of the exiled cards without paying its mana cost"

// evercoatUrsinePicked is the answered "one of them": the carry is the
// Ursine object, and the pick is played. An on-disk key: never renamed,
// never reused.
var evercoatUrsinePicked = game.RegisterChooseCardsThen("choose-cards/evercoat-ursine-play-one", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 || len(r.Nums) == 0 || len(r.Picked) == 0 {
		return nil
	}
	ref := game.PermissionCardRef{ID: r.Carry[0], Epoch: r.Nums[0]}
	cmdr2S50PlayHiddenCard(g, r.Chooser, ref, r.Picked[0], "Evercoat Ursine — play the exiled card without paying its mana cost")
	return nil
})
