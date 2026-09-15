package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Samwise Gamgee — Legendary Creature — Halfling Peasant {G}{W}, 2/2
// (EDHREC rank 2361):
//
//	"Whenever another nontoken creature you control enters, create a
//	 Food token. (It's an artifact with "{2}, {T}, Sacrifice this
//	 token: You gain 3 life.")
//	 Sacrifice three Foods: Return target historic card from your
//	 graveyard to your hand. (Artifacts, legendaries, and Sagas are
//	 historic.)"
//
// The Food engine. The trigger is Soul of the Harvest's condition —
// another creature the controller controls entered, tokens excluded
// — and the shared Food template.
//
// Declared simplification (weaker than printed): the second ability
// is not offered. "Sacrifice three Foods" is a cost that sacrifices
// three permanents, and the engine's sacrifice-cost component pays
// exactly one (validateSacrificeCostLocked); a cost with no shape is
// left out rather than priced at one Food (#259). The seam is a
// counted SacrificeOther.
func init() {
	Register(Spec{
		OracleID:     "7ce37c26-91ea-493f-bfc3-a890d4538bc1",
		Name:         "Samwise Gamgee",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The second ability isn't available — a cost can't sacrifice three Foods, so historic cards can't be returned from your graveyard."},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15AnotherNontokenCreatureYouControlEntered(ev, source, g)
			}, "Samwise Gamgee — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
	})
}
