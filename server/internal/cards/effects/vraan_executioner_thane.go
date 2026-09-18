package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vraan, Executioner Thane — Legendary Creature — Phyrexian Vampire
// {1}{B}, 2/2 (EDHREC rank 4219):
//
//	"Whenever one or more other creatures you control die, each
//	 opponent loses 2 life and you gain 2 life. This ability triggers
//	 only once each turn."
//
// A two-mana aristocrats payoff whose rate is enormous and whose
// throttle is the point: at a four-player table one death a turn is a
// six-point swing, and a hundred deaths a turn is still a six-point
// swing. That throttle is what lets it cost two rather than four.
//
// Three clauses, three mechanisms, and getting any of them wrong
// makes a different card:
//
//   - "ONE OR MORE … die" is a batch (CR 603.2d): a board wipe that
//     kills five of your creatures is ONE trigger, not five. That is
//     OncePerBatch, which the engine dedupes per event batch.
//   - "OTHER creatures YOU CONTROL" excludes Vraan himself — a wipe
//     that takes Vraan along with the rest still triggers on the
//     others, because they died in the same batch; Vraan alone dying
//     does not.
//   - "only once each turn" is a separate and stricter limit than the
//     batch, and it is the one that matters in a sacrifice deck:
//     three separate Viscera Seer activations across a turn are three
//     batches, and only the first pays. The tally is the engine's own
//     trigger log, keyed by source and label
//     (b11TriggeredThisTurn) — a trigger that was countered still
//     counts, as printed.
//
// The drain is life LOSS, not damage, so nothing prevents it and no
// damage replacement doubles it.
//
// No simplification.
const b40VraanLabel = "Vraan, Executioner Thane — each opponent loses 2, you gain 2"

func init() {
	Register(Spec{
		OracleID:     "b2f2645f-5f74-456a-bd02-83169d8b8a7e",
		Name:         "Vraan, Executioner Thane",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				dead, ok := diedCreature(ev, g)
				if !ok || dead.Controller != source.Controller {
					return false
				}
				return !b11TriggeredThisTurn(g, source.InstanceID, b40VraanLabel)
			}, b40VraanLabel, b40DrainEachOpponent(2))),
		},
	})
}
