package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bb86849c-fdec-4ca5-a8c0-ed682d921759",
		Name:         "Genesis Device",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Genesis Device — search for a basic land card or Planet card", func(g *game.Game, item *game.StackItem) error {
				return trekB20SearchBasicOrPlanetTapped(NewContext(g, item), "Genesis Device — a basic land card or Planet card to put onto the battlefield tapped")
			}),
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}{G}{G}, {T}, Sacrifice Genesis Device: Sacrifice any number of lands you control, then search your library for up to that many land cards, put them onto the battlefield tapped, then shuffle. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}{G}{G}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ChoosePermanents{
					Question:   "Genesis Device — sacrifice any number of lands you control",
					Candidates: scapeshiftLands,
					Keyed:      genesisDeviceChosen,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// genesisDeviceChosen is the chosen lands sacrificed and the search for
// that many, keyed (#4290). An on-disk key: never renamed, never reused.
var genesisDeviceChosen = RegisterPermanentsPick("prompt-run/genesis-device-sacrifice", PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		return trekB20FetchThatManyLandsTapped(ctx, picked.Count(), "Genesis Device — land cards onto the battlefield tapped")
	},
})
