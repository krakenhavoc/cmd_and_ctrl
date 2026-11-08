package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Larval Scoutlander — Artifact — Spacecraft {2}{G}, 3/3 (Edge of
// Eternities, #4461):
//
//	"When this Spacecraft enters, you may sacrifice a land or Lander. If
//	 you do, search your library for up to two basic land cards, put
//	 them onto the battlefield tapped, then shuffle.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 7+.)
//	 7+ | Flying"
//
// The sacrifice is asked as the trigger resolves (maySacrificeOne, Cornered
// Crook's shape) over the lands and Landers you control then, and the
// search runs only once one has really gone. The search is Burnished
// Hart's (up to two basic land cards, onto the battlefield tapped, then
// shuffle), without the reveal the text doesn't print. Station, the 7+ creature threshold and its keyword are the
// shared constructors (Galvanizing Sawship).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3f3d55c-a3f6-441b-a380-60c0734dc514",
		Name:         "Larval Scoutlander",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Larval Scoutlander — you may sacrifice a land or Lander to search for up to two basic lands",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Larval Scoutlander — sacrifice a land or Lander to search for up to two basic lands?",
						Or(Land(), OfSubtype("Lander")), false, larvalScoutlanderSearch)
				}),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(7, 3, 3),
			ThresholdKeywords(7, "flying"),
		),
	})
}

// larvalScoutlanderSearch is "if you do": the search, keyed. An on-disk
// key.
var larvalScoutlanderSearch = RegisterSacrificedOne("prompt-run/larval-scoutlander-search", func(ctx *Context, _ uuid.UUID) error {
	return SearchLibrary{
		Player:        ctx.Controller(),
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         2,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        "Larval Scoutlander — up to two basic lands",
	}.Apply(ctx)
})
