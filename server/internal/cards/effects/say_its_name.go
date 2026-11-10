package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Say Its Name — Sorcery {1}{G} (Duskmourn):
//
//	"Mill three cards. Then you may return a creature or land card from
//	 your graveyard to your hand.
//	 Exile this card and two other cards named Say Its Name from your
//	 graveyard: Search your graveyard, hand, and/or library for a card
//	 named Altanak, the Thrice-Called and put it onto the battlefield. If
//	 you search your library this way, shuffle. Activate only as a
//	 sorcery."
//
// The spell is Grapple with the Past's: the mill, then a choice made as it
// resolves (ReturnChosenFromGraveyard), so a card it just milled can be
// returned. The second ability works from the graveyard (CR 113.6): the
// card exiles itself (ExileThis) and two OTHER cards named Say Its Name
// (ExileFromGraveyard; the source is never a legal pick of that cost).
//
// The search is SearchLibrary{AlsoGraveyard, AlsoHand}, the three-zone
// and/or search (#5408): the searcher first chooses any of their graveyard,
// hand and library, and only the chosen zones' cards are offered. The
// library is shuffled, and the search announced, if and only if it was
// chosen (CR 701.23a, 701.24a). A name is a stated quality, so the searcher
// may find nothing (CR 701.23b). Altanak enters through the ordinary entry
// pipeline.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "474f0675-4088-4f77-8081-f9ed7d492837",
		Name:         "Say Its Name",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			// The pick waits for a prompt the mill asks (CR 608.2c, #4155).
			MillCards{N: 3},
			Instruction(func(ctx *Context) error {
				return ReturnChosenFromGraveyard{
					Player:   ctx.Controller(),
					Question: "Say Its Name — you may return a creature or land card from your graveyard to your hand",
					Match:    func(c game.Card) bool { return c.IsCreature() || c.IsLand() },
					Max:      1,
					ToHand:   true,
				}.Apply(ctx)
			}),
		),
		Activated: []ActivatedAbility{{
			Label: "Exile this card and two other cards named Say Its Name from your graveyard: Search your graveyard, hand, and/or library for a card named Altanak, the Thrice-Called and put it onto the battlefield",
			Cost: Plus(ExileThis(), ExileFromGraveyard(2, "two other cards named Say Its Name",
				func(c game.Card) bool { return c.Name == "Say Its Name" })),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     func(c game.Card) bool { return c.Name == "Altanak, the Thrice-Called" },
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					AlsoGraveyard: true,
					AlsoHand:      true,
					Reason:        "Say Its Name — a card named Altanak, the Thrice-Called from your graveyard, hand or library",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
