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
// "Sacrifice three Foods" is a sacrifice cost with a count of three
// (#747, SacrificeN), paid at announce. The target is chosen with the
// activation and checked again on resolution; "historic" is the
// artifact / legendary / Saga predicate the Jhoira cards share.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ce37c26-91ea-493f-bfc3-a890d4538bc1",
		Name:         "Samwise Gamgee",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15AnotherNontokenCreatureYouControlEntered(ev, source, g)
			}, "Samwise Gamgee — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice three Foods: Return target historic card from your graveyard to your hand.",
			Cost:    SacrificeN(3, "three Foods", HasSubtype("Food")),
			Targets: TargetCardInGraveyard("target historic card from your graveyard", YouOwn(), b14Historic()),
			Effect:  returnFirstLegalGraveyardTargetToHand,
		}},
	})
}
