package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sneaky Snacker — Creature — Faerie Rogue {U}{B}, 2/1 (Modern Horizons 3):
//
//	"Flying
//	 When you draw your third card in a turn, return this card from your
//	 graveyard to the battlefield tapped."
//
// A graveyard trigger (CR 603.6c-style): "you" is the card's owner there. Only
// the third card drawn that turn counts, so the fourth does nothing.
func init() {
	Register(Spec{
		OracleID:        "f255492e-4e12-48b4-9c19-8a2a2780a1b7",
		Name:            "Sneaky Snacker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && g.TurnTallyFor(source.Controller).CardsDrawn == 3
			}, "Sneaky Snacker — return it from your graveyard to the battlefield tapped", returnThisFromGraveyardTapped)),
		},
	})
}
