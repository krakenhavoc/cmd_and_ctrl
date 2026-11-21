package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cynical Loner — Creature — Human Survivor {1}{B}, 3/1 (Duskmourn:
// House of Horror, #4466):
//
//	"This creature can't be blocked by Glimmers.
//	 Survival — At the beginning of your second main phase, if this
//	 creature is tapped, you may search your library for a card, put it
//	 into your graveyard, then shuffle."
//
// The block rule is April O'Neil, Kunoichi Trainee's CantBeBlockedBy
// with a creature-type test (Subtype "Glimmer", read through the layers).
// The Survival trigger (survival.go) is Vile Entomber's search made
// optional: "you may search" lets the player decline the card and the
// shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d5ff568-1c55-4757-94d0-5743ca674bab",
		Name:         "Cynical Loner",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.Subtype("Glimmer"), "Glimmers"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Survival("Cynical Loner — you may search for a card and put it into your graveyard", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:   item.Controller,
					Dest:     game.ZoneGraveyard,
					Limit:    1,
					Shuffle:  true,
					Optional: true,
					Reason:   "Cynical Loner — a card, into your graveyard",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
