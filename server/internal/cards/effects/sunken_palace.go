package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunken Palace — Land — Cave (Modern Horizons 3 Commander, #5948):
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 {1}{U}, {T}, Exile seven cards from your graveyard: Add {U}. When you
//	 spend this mana to cast a spell or activate an ability, copy that
//	 spell or ability. You may choose new targets for the copy. (Mana
//	 abilities can't be copied.)"
//
// The second mana ability is a mana ability whose cost exiles seven cards
// from your graveyard (CR 605.1a, 118.3: with fewer than seven it can't be
// activated). Its spend trigger is Study Hall's shape (WhenManaSpent): the
// trigger goes on the stack above the spell or ability the mana paid for,
// and "that spell or ability" is read off the trigger's payload rather than
// targeted (CR 115.10), so one countered in response is still named and
// copying it does nothing. It fires for a cast or for the activation of an
// ability that uses the stack; a mana ability never does (CR 605.3b), so the
// parenthesis needs no clause. The copy is controlled by whoever spent the
// mana (CR 707.10), who may choose new targets for it (CR 707.10c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "c098c507-5154-423a-a70b-f6dfd4959cf6",
		Name:           "Sunken Palace",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		Replacements:   []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{U}",
				Label:    "Add {U}",
			},
			{
				Cost: ManaAbilityCost{
					Tap:        true,
					Mana:       "{1}{U}",
					ExileCards: ExileCardsFromGraveyard(7, "seven cards from your graveyard", nil),
				},
				Produced: "{U}",
				Label:    "{1}{U}, {T}, Exile seven cards from your graveyard: Add {U}. When you spend this mana to cast a spell or activate an ability, copy that spell or ability.",
				SpendRiders: []game.ManaSpendRider{WhenManaSpent("Sunken Palace", game.ManaSpendTrigger{
					Label:  "Sunken Palace — copy that spell or ability",
					Effect: sunkenPalaceCopy,
				}, game.ManaRestrictAnyOf([]string{game.ManaRestrictCast}, []string{game.ManaRestrictActivate}))},
			},
		},
	})
}

// sunkenPalaceCopy is the rider's effect: copy the spell or ability the mana
// paid for, and choose new targets for the copy.
func sunkenPalaceCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cards := ctx.PayloadCards()
	if len(cards) == 0 {
		return nil
	}
	if _, _, onStack := g.AbilityItemOnStackForEffect(cards[0]); onStack {
		return CopyAbility{ItemID: cards[0], Controller: item.Controller, ChooseNewTargets: true}.Apply(ctx)
	}
	return CopySpell{StackID: cards[0], Controller: item.Controller, ChooseNewTargets: true, FromLastKnown: true}.Apply(ctx)
}
