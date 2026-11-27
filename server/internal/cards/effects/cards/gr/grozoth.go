package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grozoth — Creature — Leviathan {6}{U}{U}{U}, 9/9 (Ravnica: City of Guilds):
//
//	"Defender
//	 When this creature enters, you may search your library for any number of
//	 cards that have mana value 9, reveal them, put them into your hand, then
//	 shuffle.
//	 {4}: This creature loses defender until end of turn.
//	 Transmute {1}{U}{U} ({1}{U}{U}, Discard this card: Search your library for a
//	 card with the same mana value as this card, reveal it, put it into your hand,
//	 then shuffle. Transmute only as a sorcery.)"
//
// Defender is the printed keyword. The enters trigger is Gathering Throng's
// optional unbounded search, over cards of mana value exactly 9 (a card with
// {X} counts X as 0 in a library, CR 202.3e), and it shuffles. The pump row is
// one layer-6 effect on this object with its own timestamp, ending at cleanup
// (CR 514.2), Gargoyle Sentinel's. Transmute is the keyword's constructor
// (CR 702.53a), finding a mana value 9 card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ecbc87e-a63e-4afc-b5eb-2fc3dcd685f5",
		Name:            "Grozoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Grozoth — search for any number of cards with mana value 9",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.ManaValue() == 9 },
						Dest:      game.ZoneHand,
						Unbounded: true,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Grozoth — you may search for any number of cards with mana value 9",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{4}: This creature loses defender until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{4}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.UntilEndOfTurn(effects.NewContext(g, item), item.SourceCardID, nil,
						"Grozoth — loses defender until end of turn", game.RemoveKeywordsMod("defender"))
				},
			},
			effects.Transmute("{1}{U}{U}"),
		},
	})
}
