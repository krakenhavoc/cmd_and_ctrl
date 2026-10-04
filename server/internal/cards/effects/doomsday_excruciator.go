package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doomsday Excruciator — Creature — Demon {B}{B}{B}{B}{B}{B}, 6/6:
//
//	"Flying
//	 When this creature enters, if it was cast, each player exiles all
//	 but the bottom six cards of their library face down.
//	 At the beginning of your upkeep, draw a card."
//
// "If it was cast" is an intervening if (CR 603.4) read off the
// permanent's cast provenance (CR 400.7d), the way Geological Appraiser
// reads it: a reanimated or flickered Excruciator exiles nothing. The
// condition is asked as it enters and again as the ability resolves.
//
// Each player exiles every card of their library but the bottom six,
// face down (CR 406.3), so nobody, the owner included, may look at
// them. A library of six or fewer loses nothing. The engine exiles from
// the top, so "all but the bottom six" is the top (size - 6) cards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ad0db433-a406-4ef0-8ffa-416af610c4e7",
		Name:            "Doomsday Excruciator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, doomsdayExcruciatorWasCast),
				"Doomsday Excruciator — each player exiles all but the bottom six cards of their library face down",
				doomsdayExcruciatorExile),
			AtYourUpkeep("Doomsday Excruciator — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

// doomsdayExcruciatorWasCast is the enter trigger's intervening if.
func doomsdayExcruciatorWasCast(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && source.Provenance.FromZone != ""
}

// doomsdayExcruciatorExile is the enters trigger's body.
func doomsdayExcruciatorExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if ctx.CastProvenance().FromZone == "" {
		return nil
	}
	for _, id := range tablePlayers(ctx) {
		p := g.PlayerByIDForEffect(id)
		if p == nil || p.Library == nil {
			continue
		}
		if n := len(p.Library.Cards) - 6; n > 0 {
			if _, err := g.ExileTopFaceDownForEffect(id, n); err != nil {
				return err
			}
		}
	}
	return nil
}
