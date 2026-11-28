package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sea Troll — Creature — Troll {2}{U}, 2/1 (Homelands, S82 OLD b821, #5635):
//
//	"{U}: Regenerate this creature. Activate only if this creature blocked or
//	was blocked by a blue creature this turn."
//
// The regeneration shield is RegenerateThisSource. The window reads the turn's
// per-pair block record (TurnTally.CombatBlocks): a pair that names the Troll
// as the blocker or as the attacker, whose other creature is blue now (CR
// 105.2, read through the layers). A record is history, so a blue creature
// that has since left still counts, as it is read from last known information
// where it can be; one that cannot be read does not count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "feabe30a-3c28-43a1-b538-b0207758520c",
		Name:         "Sea Troll",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{U}: Regenerate this creature. Activate only if this creature blocked or was blocked by a blue creature this turn.",
			Purpose:   game.Purpose{Answers: game.AnswerSurvive},
			Cost:      effects.ManaCost("{U}"),
			Condition: seaTrollBlockedOrWasBlockedByBlue,
			Effect:    effects.RegenerateThisSource,
		}},
	})
}

// seaTrollBlockedOrWasBlockedByBlue is Sea Troll's activation window.
func seaTrollBlockedOrWasBlockedByBlue(g *game.Game, _ uuid.UUID, source uuid.UUID) bool {
	me, ok := g.LookupCardForEffect(source)
	if !ok {
		return false
	}
	blue := func(id uuid.UUID) bool {
		c, ok := g.LookupCardForEffect(id)
		return ok && c.HasColorWord("U")
	}
	for _, b := range g.TurnTally.CombatBlocks {
		switch {
		case b.Blocker.ID == source && b.Blocker.Epoch == me.ObjectEpoch && blue(b.Attacker.ID):
			return true
		case b.Attacker.ID == source && b.Attacker.Epoch == me.ObjectEpoch && blue(b.Blocker.ID):
			return true
		}
	}
	return false
}
