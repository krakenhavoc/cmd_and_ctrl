package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urza, Planeswalker — Legendary Planeswalker — Urza, loyalty 7. The
// combined back face of The Mightstone and Weakstone and Urza, Lord
// Protector (CR 712.5e); it is never a card in a deck, only the melded
// permanent the two become (ADR 0145, #2699):
//
//	"You may activate the loyalty abilities of Urza twice each turn
//	 rather than only once.
//	 +2: Artifact, instant, and sorcery spells you cast this turn cost
//	 {2} less to cast. You gain 2 life.
//	 +1: Draw two cards, then discard a card.
//	 0: Create two 1/1 colorless Soldier artifact creature tokens.
//	 −3: Exile target nonland permanent.
//	 −10: Artifacts and planeswalkers you control gain indestructible
//	 until end of turn. Destroy all nonland permanents."
//
// THE STATIC is Spec.LoyaltyTwiceEachTurn: CR 606.3's limit is two for
// this permanent, and the same ability may be activated both times.
// THE +2 is a promise about every artifact, instant and sorcery spell
// its controller casts for the rest of the turn (NextSpellPromise with
// EveryMatchingSpell), priced as an ordinary CR 601.2f reduction; it
// outlives Urza, as an effect of a resolved ability does. THE −10
// grants indestructible first and destroys second, in printed order, so
// its controller's artifacts and planeswalkers — Urza included —
// survive.
//
// The melded permanent enters with seven loyalty counters (CR 306.5b)
// off the back face's printed loyalty, and its mana value is 8, the sum
// of its cards' front faces (CR 712.8g).
func init() {
	Register(Spec{
		OracleID:             "759406d7-44ae-4260-9ef5-3bb2c92f751a",
		Name:                 "Urza, Planeswalker",
		Completeness:         CompletenessFull,
		LoyaltyTwiceEachTurn: true,
		Activated: []ActivatedAbility{
			{
				Label: "+2: Artifact, instant, and sorcery spells you cast this turn cost {2} less to cast. You gain 2 life.",
				Cost:  LoyaltyCost(2),
				Effect: Do(
					GrantNextSpellPromise{From: "Urza, Planeswalker", Promise: game.NextSpellPromise{
						Filter:             game.PermissionFilter{ArtifactInstantOrSorceryOnly: true},
						Reduce:             2,
						EveryMatchingSpell: true,
						Text:               "Artifact, instant, and sorcery spells you cast this turn cost {2} less to cast.",
					}},
					GainLife{Amount: 2},
				),
			},
			{
				Label:   "+1: Draw two cards, then discard a card.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 2, Discards: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return drawThenDiscard(g, item.Controller, item.SourceCardID, 2, 1, "")
				},
			},
			{
				Label:   "0: Create two 1/1 colorless Soldier artifact creature tokens.",
				Cost:    LoyaltyCost(0),
				Purpose: game.Purpose{Tokens: 2},
				Effect:  Do(CreateToken{Template: TokenCard("1/1 colorless Soldier artifact"), N: 2}),
			},
			{
				Label:   "−3: Exile target nonland permanent.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetPermanent("target nonland permanent", Nonland()),
				Effect:  ExileFirstTarget,
			},
			{
				Label: "−10: Artifacts and planeswalkers you control gain indestructible until end of turn. Destroy all nonland permanents.",
				Cost:  LoyaltyCost(-10),
				Purpose: game.Purpose{Sweep: game.Sweep{
					Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
				Effect: urzaPlaneswalkerMinusTen,
			},
		},
	})
}

// urzaPlaneswalkerMinusTen is the −10, in printed order: the
// indestructible grant lands first, so the destruction that follows
// spares the controller's artifacts and planeswalkers (CR 702.12b).
func urzaPlaneswalkerMinusTen(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GrantKeywordUntilEOT{
		Match:    And(Or(Artifact(), Planeswalker()), YouControl()),
		Keywords: []string{"indestructible"},
		Label:    "Urza, Planeswalker — indestructible",
	}).Apply(ctx); err != nil {
		return err
	}
	return DestroyAllMatching{Match: Nonland()}.Apply(ctx)
}
