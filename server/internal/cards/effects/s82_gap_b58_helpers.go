package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82GapB58Landscape is a row of the Modern Horizons 3 "Landscape" cycle
// (Twisted, Seething, Sheltering): "{T}: Add {C}. {T}, Sacrifice this land:
// Search your library for a basic A, B, or C card, put it onto the
// battlefield tapped, then shuffle. Cycling {x}{y}{z}".
func S82GapB58Landscape(oracleID, name, cycling string, subtypes ...string) Spec {
	match := IsBasicLandOfAnySubtype(subtypes...)
	list := "a basic " + subtypes[0] + ", " + subtypes[1] + ", or " + subtypes[2]
	return Spec{
		OracleID:      oracleID,
		Name:          name,
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{TapForColorless()},
		Activated: []ActivatedAbility{
			{
				Label:   "{T}, Sacrifice this land: Search your library for " + list + " card, put it onto the battlefield tapped, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
				Cost:    Plus(TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     match,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Reveal:        true,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Choose " + list + " card to put onto the battlefield tapped",
					}.Apply(NewContext(g, item))
				},
			},
			Cycling(cycling),
		},
	}
}

// s82GapB58FlareCultivationHand is the hand half of Flare of Cultivation's
// two-land search (cultivate.go's twoLandSearch). An on-disk key.
var S82GapB58FlareCultivationHand = LandToHandTail("flare-of-cultivation/basic-land-to-hand", "Flare of Cultivation", "basic land", IsBasicLand)
