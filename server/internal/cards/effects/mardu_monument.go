package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mardu Monument — Artifact {2} (Tarkir: Dragonstorm, #4463):
//
//	"When this artifact enters, search your library for a basic
//	 Mountain, Plains, or Swamp card, reveal it, put it into your hand,
//	 then shuffle.
//	 {2}{R}{W}{B}, {T}, Sacrifice this artifact: Create three 1/1 red
//	 Warrior creature tokens. They gain menace and haste until end of
//	 turn. Activate only as a sorcery. (A creature with menace can't be
//	 blocked except by two or more creatures.)"
//
// Abzan Monument's shape. The entry search is the basic-of-three
// predicate (IsBasicLandOfAnySubtype) sent to the hand and revealed. The
// Warriors are made by one creation and the grant rides its keyed tail
// (thenCreatedGainKeywordUntilEOT, with haste as its second word), so a
// token doubler's extra Warriors gain menace and haste too, and the grant
// is an effect on those tokens only: a Warrior made later does not have
// them. Menace and haste are gained, not printed, so they end at the
// cleanup step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f662383-1684-4bd3-be05-e696d301e26c",
		Name:         "Mardu Monument",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mardu Monument — search for a basic Mountain, Plains, or Swamp card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLandOfAnySubtype("Mountain", "Plains", "Swamp"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Mardu Monument — a basic Mountain, Plains, or Swamp card",
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{R}{W}{B}, {T}, Sacrifice this artifact: Create three 1/1 red Warrior creature tokens. " +
				"They gain menace and haste until end of turn. Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 3},
			Cost:         Plus(ManaCost("{2}{R}{W}{B}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Effect:       marduMonumentWarriors,
		}},
	})
}

// marduMonumentWarriors creates the three Warriors; the tail gives each of
// them menace and haste until end of turn.
func marduMonumentWarriors(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	gain := withWord(ctx.ItemTail(thenCreatedGainKeywordUntilEOT), "menace")
	gain.Words = []string{"haste"}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("1/1 red Warrior")), Count: 3}},
	}, gain)
}
