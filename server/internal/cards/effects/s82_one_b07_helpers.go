package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set ONE, batch 7 (#5479): Phyrexia: All Will Be One's corrupted
// commons and uncommons.
//
// Corrupted is an ability word with no rules meaning of its own (CR 207.2c):
// each card spells its own condition, and every one in this batch reads the
// same thing, "an opponent has three or more poison counters". Poison is
// public, and the count is read live: on every layer recompute for a static,
// as the ability is announced for an activation condition, as the event
// happens and again as the trigger resolves for an intervening "if" (CR 603.4).

// playerIsCorrupted is "<player> has three or more poison counters".
func PlayerIsCorrupted(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && p.Poison >= 3
}

// CondCorrupted is "if an opponent has three or more poison counters" as an
// intervening-if Condition, read for the resolving ability's controller. It
// declares no footprint, so a row that carries it is simply not reordered
// (CondDescend's reason).
type CondCorrupted struct{}

func (CondCorrupted) Holds(ctx *Context) bool {
	return SeedcoreCorrupted(ctx.Game, ctx.Controller(), uuid.Nil)
}

func (CondCorrupted) Condition() (game.FootprintStep, bool) { return game.FootprintStep{}, false }

// YouAreCorrupted is the When for "if an opponent has three or more poison
// counters", read for the source's controller as the event happens, for
// composing with AllOf.
func YouAreCorrupted(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return SeedcoreCorrupted(g, source.Controller, uuid.Nil)
}

// SelfKeywordWhileCorrupted is "Corrupted — As long as an opponent has three
// or more poison counters, this creature has <keyword>": a layer 6 grant on
// the source alone (CR 613.1f), re-answered on every recompute.
func SelfKeywordWhileCorrupted(keyword string) game.StaticAbility {
	ab := KeywordGrant(WhileYouHave(func(g *game.Game, you uuid.UUID) bool {
		return SeedcoreCorrupted(g, you, uuid.Nil)
	}, SelfCreature), keyword)
	ab.Label = "Corrupted — has " + keyword + " as long as an opponent has three or more poison counters."
	return ab
}

// corruptedAnd is an activation condition: corrupted, and `also` when set.
func CorruptedAnd(also ActivationCondition) ActivationCondition {
	return func(g *game.Game, controller, source uuid.UUID) bool {
		return SeedcoreCorrupted(g, controller, source) && (also == nil || also(g, controller, source))
	}
}
