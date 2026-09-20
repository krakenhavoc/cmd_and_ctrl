package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reckless Fireweaver — 1/3 Creature — Human Artificer for {1}{R}:
//
//	"Whenever an artifact you control enters, this creature deals 1
//	damage to each opponent."
//
// The reason an Izzet artifact deck can win without attacking: every
// Treasure, Clue and Blood token that hits the battlefield pings the
// whole table. See artifactEnteredUnderYourControl for the
// one-trigger-per-artifact batching note — it doesn't apply here.
// Unlike Ingenious Artillerist, this card is printed per-artifact
// ("Whenever AN artifact... enters", not "one or more"), so firing
// once per artifact for a flat 1 damage each is exactly the card, not
// a simplification of it. No simplification.
func init() {
	Register(Spec{
		OracleID:     "180e1a7e-890d-477c-80a5-da8a5f2857b3",
		Name:         "Reckless Fireweaver",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return artifactEnteredUnderYourControl(ev, source, g)
			}, "Reckless Fireweaver — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
		},
	})
}
