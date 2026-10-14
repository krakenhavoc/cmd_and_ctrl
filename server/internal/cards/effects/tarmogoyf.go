package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tarmogoyf — "*/1+* — Tarmogoyf's power is equal to the number of
// card types among cards in all graveyards and its toughness is
// equal to that number plus 1."
//
// First S16 Layer-7a (CDA — characteristic-defining ability)
// catalog card. The classic CDA shape: AppliesTo == self only,
// Apply SETS power and toughness (not modifies). 7a runs before
// 7c (anthems) and 7d (counters), so Tarmogoyf with a +1/+1
// counter and a Glorious Anthem is `(N+2) / (N+3)` where N is the
// distinct-types count: 7a sets N/(N+1), 7c +1/+1 → (N+1)/(N+2),
// counter at 7d via CurrentPower delegation → (N+2)/(N+3) (counter
// integration ships when sub-PR 5+ wires layer 7d's delegation).
//
// "Card types among cards in all graveyards" reads the union of
// every graveyard's cards' printed types — a recompute pass walks
// every player's graveyard zone and unions Card.printedCharacteristic().
// Types into a set, then sets Tarmogoyf's power to len(set).
//
// Tarmogoyf's PRINTED P/T on the wire is "*"/"1+*" but the engine
// only carries numeric stats; the card file ships printed 0/1 and
// the CDA overrides at recompute time. The wire effective P/T is
// what shows.
//
// Reviewed against the oracle text (rf-a, 2026-10-14): everything
// printed happens, so it is CompletenessFull. The count is the union of
// card types over every player's graveyard (DistinctCardTypesInAllGraveyards),
// recomputed whenever a card enters or leaves a graveyard (the layer
// listener bumps on either direction). A split card there has both
// halves' types (CR 709.4c); a modal or adventure card only its front
// face; Kindred is a card type; Legendary, Basic and Snow are
// supertypes and do not count. Pinned by the Tarmogoyf tests in
// reality_fracture_rf_a_test.go. The one deviation the engine has
// here is not this card's: the type-line parser reads the old word
// "Tribal" as a supertype, so a graveyard holding only a pre-Kindred
// "Tribal" card (two playtest cards, never Commander-legal) would count
// one type short.
func init() {
	Register(Spec{
		OracleID:     "45900b2f-f6a9-4c42-9642-008f3c1cf6dd",
		Name:         "Tarmogoyf",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:    game.Layer7PT,
				Label:    "Power is the number of card types among cards in all graveyards; toughness is that plus 1.",
				SubLayer: game.SubLayer7A_CDA,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
					n := game.DistinctCardTypesInAllGraveyards(g)
					c.Power = n
					c.Toughness = n + 1
				},
			},
		},
	})
}
