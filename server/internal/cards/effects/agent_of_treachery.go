package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Agent of Treachery — 2/3 Human Rogue for {5}{U}{U}:
//
//	"When this creature enters, gain control of target permanent.
//	 At the beginning of your end step, if you control three or more
//	 permanents you don't own, draw three cards."
//
// The CR 611.2a card: a continuous effect with NO stated duration,
// which lasts until the game ends. Killing the Agent does not give
// the permanent back — the effect was created by the trigger's
// resolution and has nothing to do with the Agent afterwards, which
// is exactly the difference between this and Sower of Temptation.
//
// The only thing that ends it is the stolen permanent ceasing to be
// the object it was: dying, being exiled, being flickered (CR 400.7).
// The registry entry is pinned to it and goes when it does, so a game
// full of resolved Agents does not accumulate dead entries (ADR 0063
// Decision 6).
//
// "Target permanent", not "target creature": a land, an artifact, an
// enchantment and a planeswalker are all fair game, and the layer-2
// bucket does not care which — it changes a controller.
//
// The end-step trigger is the card's own payoff and reads the board
// rather than remembering what it took: "permanents you don't own" is
// every battlefield permanent you control whose owner is someone
// else, however you came by it. An intervening-if (CR 603.4), so it
// checks on the way onto the stack and again on resolution.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "4cd82c7a-86df-4d29-b578-c009208c0d4b",
		Name:         "Agent of Treachery",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Targets: TargetPermanent("target permanent"),
				Key:     "Agent of Treachery — gain control of target permanent",
				Effect: func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					return GainControl{
						Target:     item.Targets[0].ID,
						Controller: item.Controller,
						Duration:   game.IndefiniteDuration(),
						Label:      "Agent of Treachery — gain control (no stated duration)",
					}.Apply(NewContext(g, item))
				},
			},
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventBeginEndStep && ev.Actor == source.Controller &&
					permanentsYouControlButDontOwn(g, source.Controller) >= 3
			}, "Agent of Treachery — draw three cards",
				func(g *game.Game, item *game.StackItem) error {
					// CR 603.4: the condition is checked again on
					// resolution, so an Agent whose stolen permanents
					// went home in response draws nothing.
					if permanentsYouControlButDontOwn(g, item.Controller) < 3 {
						return nil
					}
					return DrawCards{Player: item.Controller, N: 3}.Apply(NewContext(g, item))
				}),
		},
	})
}

// permanentsYouControlButDontOwn counts the battlefield permanents
// `player` controls whose owner is somebody else — Agent of
// Treachery's "permanents you don't own". It reads the board rather
// than any record of what an effect took, so a permanent gained any
// other way (Mind Control, an exchange, a Homeward Path that has not
// been activated) counts exactly as the card says.
func permanentsYouControlButDontOwn(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.Owner != player {
			n++
		}
	}
	return n
}
