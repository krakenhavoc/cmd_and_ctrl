package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_creature_a_helpers.go — small shared pieces for the Reality
// Fracture (FRA / FRC) cards. Tracker #2795.

// rfCreatureAArtifactsYouControl counts the artifacts a player controls, after
// layers ("where X is the number of artifacts you control"). Caller
// holds g.mu.
func rfCreatureAArtifactsYouControl(g *game.Game, player uuid.UUID) int {
	g.RecomputeLayersIfStaleLocked()
	return b03ArtifactsControlled(g, player)
}

// rfCreatureACastSpellTargetsAny reports whether the spell a cast event names is on
// the stack with at least one target for which `match` is true. The
// cast path announces targets before it emits EventCast (CR 601.2c, i),
// so the stack item already carries them when a "whenever you cast a
// spell that targets …" trigger looks. Caller holds g.mu.
func rfCreatureACastSpellTargetsAny(g *game.Game, spellID uuid.UUID, match func(t game.TargetRef) bool) bool {
	item := g.StackItemForEffect(spellID)
	if item == nil {
		return false
	}
	for _, t := range item.Targets {
		if match(t) {
			return true
		}
	}
	return false
}

// rfCreatureATargetIsCreatureOnBattlefield reports whether a card target is a
// creature still on the battlefield, and returns it.
func rfCreatureATargetIsCreatureOnBattlefield(g *game.Game, t game.TargetRef) (game.Card, bool) {
	if t.Kind != game.TargetCard {
		return game.Card{}, false
	}
	if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneBattlefield {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(t.ID)
	if !ok || !c.IsCreature() {
		return game.Card{}, false
	}
	return c, true
}
