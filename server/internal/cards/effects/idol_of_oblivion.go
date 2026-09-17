package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Idol of Oblivion — Artifact {2} (EDHREC rank 200):
//
//	"{T}: Draw a card. Activate only if you created a token this turn.
//	 {8}, {T}, Sacrifice this artifact: Create a 10/10 colorless
//	 Eldrazi creature token."
//
// A token deck's draw engine with a late-game mana sink. "Activate
// only if you created a token this turn" is the draw's activation
// condition (CR 602.1b, #743), read off the turn tally's
// TokensCreated for you (#586) — tokens created under your control,
// in any step of the turn, including the untap step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c232d974-9c65-4bce-b045-e5a7cc0c62e0",
		Name:         "Idol of Oblivion",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:     "{T}: Draw a card. Activate only if you created a token this turn.",
				Cost:      TapCost(),
				Condition: idolOfOblivionCreatedATokenThisTurn,
				Effect:    b36DrawOne,
			},
			{
				Label: "{8}, {T}, Sacrifice this artifact: Create a 10/10 colorless Eldrazi creature token.",
				Cost:  Plus(ManaCost("{8}"), TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: TokenCard("10/10 colorless Eldrazi"), N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// idolOfOblivionCreatedATokenThisTurn is the Idol's draw condition.
func idolOfOblivionCreatedATokenThisTurn(g *game.Game, controller, _ uuid.UUID) bool {
	return g.TurnTallyFor(controller).TokensCreated > 0
}
