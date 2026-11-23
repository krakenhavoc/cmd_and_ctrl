package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// You Cannot Pass! — Instant {W} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Destroy target creature that blocked or was blocked by a legendary
//	 creature this turn."
//
// The target clause reads this turn's block declarations (one event per
// blocker and attacker pair): the creature must be a blocker whose attacker is
// legendary, or an attacker blocked by a legendary creature. A legendary
// creature is read from the card as it is now, wherever it is (a legendary
// creature that has died since still counts, as last-known information would
// have it). The clause is checked again as the spell resolves (CR 608.2b), and
// the destroy is the ordinary one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e94ccdb-5978-440b-b7cd-11c7ca6b2c98",
		Name:         "You Cannot Pass!",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature that blocked or was blocked by a legendary creature this turn", blockedWithALegendaryThisTurn),
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.KineticOozeDestroy)),
	})
}

// blockedWithALegendaryThisTurn is the target filter: the card was a blocker
// or an attacker in a block pair of this turn whose other half is a legendary
// creature.
func blockedWithALegendaryThisTurn(g *game.Game, _ uuid.UUID, c game.Card) bool {
	legendaryCreature := func(id uuid.UUID) bool {
		other, ok := g.LookupCardForEffect(id)
		return ok && other.IsLegendary() && other.IsCreature()
	}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventBlock {
			continue
		}
		if ev.CardID == c.InstanceID && legendaryCreature(ev.Target) {
			return true
		}
		if ev.Target == c.InstanceID && legendaryCreature(ev.CardID) {
			return true
		}
	}
	return false
}
