package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Entropic Battlecruiser — Artifact — Spacecraft {3}{B}, 3/10 (Edge of
// Eternities, #4461):
//
//	"Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 8+.)
//	 1+ | Whenever an opponent discards a card, they lose 3 life.
//	 8+ | Flying, deathtouch
//	 Whenever this Spacecraft attacks, each opponent discards a card.
//	 Each opponent who can't loses 3 life."
//
// The 1+ line is Liliana's Caress's trigger (the discarding player is the
// event's actor) behind D-7 Battle Cruiser's charge-counter gate, so it
// exists only from the first charge counter on. The attack trigger asks
// each opponent with a card in hand to discard one (Burglar Rat's
// prompt), and an opponent whose hand is empty as it resolves can't, so
// loses 3 life instead. A discard it forces is still a discard, so with
// a charge counter on the Spacecraft that opponent loses 3 life from the
// 1+ line. Station, the 8+ creature threshold and its keywords are the
// shared constructors (Galvanizing Sawship).
//
// No simplification.
func init() {
	discards := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDiscardCard},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return b18OpponentDiscarded(ev, source)
		},
		Key: "Entropic Battlecruiser — that player loses 3 life",
		Effect: func(g *game.Game, item *game.StackItem) error {
			return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Trigger.Event.Actor, -3)
		},
		ActiveWhen: AtChargeCounters(1),
	}
	Register(Spec{
		OracleID:     "90c5ea2b-d004-4219-b60c-d51bc1a5ee6b",
		Name:         "Entropic Battlecruiser",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(8, 3, 10),
			ThresholdKeywords(8, "flying", "deathtouch"),
		),
		Triggered: []game.TriggeredAbility{
			discards,
			WheneverThisAttacks("Entropic Battlecruiser — each opponent discards a card; each who can't loses 3 life",
				func(g *game.Game, item *game.StackItem) error {
					for _, opp := range NewContext(g, item).Opponents() {
						if b14HandSize(g, opp) == 0 {
							if err := g.ChangePlayerLifeForEffect(item.SourceCardID, opp, -3); err != nil {
								return err
							}
							continue
						}
						g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: opp, Source: item.SourceCardID, N: 1})
					}
					return nil
				}),
		},
	})
}
