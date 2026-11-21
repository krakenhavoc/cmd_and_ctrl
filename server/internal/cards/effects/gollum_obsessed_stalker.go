package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gollum, Obsessed Stalker — Legendary Creature — Halfling Horror {1}{B},
// 1/1 (Tales of Middle-earth Commander, #5948):
//
//	"Skulk (This creature can't be blocked by creatures with greater power.)
//	 At the beginning of your end step, each opponent dealt combat damage this
//	 game by a creature named Gollum, Obsessed Stalker loses life equal to the
//	 amount of life you gained this turn."
//
// Skulk is an engine keyword carried as printed data. "Dealt combat damage
// this game by a creature named Gollum" is read off the game's append-only
// event log as the trigger resolves: an opponent who was the target of a
// combat EventDealDamage whose dealing object was named Gollum, Obsessed
// Stalker when it dealt the damage (Event.SourceName, stamped at emit time:
// a copy named Gollum counts, and so does one that has since left the
// battlefield). The life you gained is the turn's tally (every gain path,
// lifelink included); with none gained nobody loses anything. The loss is
// simultaneous for every opponent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "95d59a06-5144-4da3-8a9c-846b087975b9",
		Name:            "Gollum, Obsessed Stalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"skulk"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Gollum, Obsessed Stalker — each opponent dealt combat damage this game by a creature named Gollum, Obsessed Stalker loses life equal to the life you gained this turn",
				gollumDrain),
		},
	})
}

const gollumName = "Gollum, Obsessed Stalker"

// gollumDrain is the trigger body.
func gollumDrain(g *game.Game, item *game.StackItem) error {
	n := b15LifeGainedThisTurn(g, item.Controller)
	if n <= 0 {
		return nil
	}
	for _, id := range gollumVictims(g, item.Controller) {
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, id, -n); err != nil {
			return err
		}
	}
	return nil
}

// gollumVictims are the controller's opponents who have been dealt combat
// damage this game by a creature named Gollum, Obsessed Stalker, in seat
// order.
func gollumVictims(g *game.Game, you uuid.UUID) []uuid.UUID {
	dealt := map[uuid.UUID]bool{}
	for _, ev := range g.Events {
		if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
			continue
		}
		name := ev.SourceName
		if name == gollumName {
			dealt[ev.Target] = true
		}
	}
	var out []uuid.UUID
	for _, p := range g.Seats {
		if p.ID != you && !p.Eliminated && dealt[p.ID] {
			out = append(out, p.ID)
		}
	}
	return out
}
