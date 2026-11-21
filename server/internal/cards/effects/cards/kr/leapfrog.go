package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leapfrog — Creature — Frog {2}{U}, 3/1 (Ravnica: Clue Edition,
// Bloomburrow):
//
//	"This creature has flying as long as you've cast an instant or sorcery
//	 spell this turn."
//
// A layer 6 grant on this permanent gated on its controller's cast tally
// for the turn (a spell counts once cast, whether or not it resolves). The
// static declares DependsOnSpellsCast so the layers are recomputed the
// moment the spell is cast, and the turn change that clears the tally ends
// it. Only the controller's spells count.
//
// No simplification.
func init() {
	flying := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && g.CastTallyFor(source.Controller).InstantSorcery > 0
	}, "flying")
	flying.Label = "This creature has flying as long as you've cast an instant or sorcery spell this turn."
	flying.DependsOnSpellsCast = true
	effects.Register(effects.Spec{
		OracleID:     "6614dbe5-d7c7-4a51-91db-912764050269",
		Name:         "Leapfrog",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{flying},
	})
}
