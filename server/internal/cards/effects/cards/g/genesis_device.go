package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Genesis Device — Legendary Artifact {2}{G} (Star Trek, #4973):
//
//	"When Genesis Device enters, search your library for a basic land
//	 card or Planet card, put it onto the battlefield tapped, then
//	 shuffle.
//	 {3}{G}{G}, {T}, Sacrifice Genesis Device: Sacrifice any number of
//	 lands you control, then search your library for up to that many land
//	 cards, put them onto the battlefield tapped, then shuffle. Activate
//	 only as a sorcery."
//
// The enters trigger is Final Frontier's search. The activated ability
// is Scapeshift's shape: one "any number" question over the lands you
// control (zero allowed), and "that many" is what actually left the
// battlefield. Nothing targets, so a hexproof land can be sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb86849c-fdec-4ca5-a8c0-ed682d921759",
		Name:         "Genesis Device",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Genesis Device — search for a basic land card or Planet card", func(g *game.Game, item *game.StackItem) error {
				return effects.TrekB20SearchBasicOrPlanetTapped(effects.NewContext(g, item), "Genesis Device — a basic land card or Planet card to put onto the battlefield tapped")
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{G}{G}, {T}, Sacrifice Genesis Device: Sacrifice any number of lands you control, then search your library for up to that many land cards, put them onto the battlefield tapped, then shuffle. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{3}{G}{G}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ChoosePermanents{
					Question:   "Genesis Device — sacrifice any number of lands you control",
					Candidates: effects.ScapeshiftLands,
					Keyed:      genesisDeviceChosen,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// genesisDeviceChosen is the chosen lands sacrificed and the search for
// that many, keyed (#4290). An on-disk key: never renamed, never reused.
var genesisDeviceChosen = effects.RegisterPermanentsPick("prompt-run/genesis-device-sacrifice", effects.PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		return effects.TrekB20FetchThatManyLandsTapped(ctx, picked.Count(), "Genesis Device — land cards onto the battlefield tapped")
	},
})
