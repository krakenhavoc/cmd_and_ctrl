package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pure Reflection — Enchantment {2}{W} (Invasion):
//
//	"Whenever a player casts a creature spell, destroy all Reflections. Then
//	 that player creates an X/X white Reflection creature token, where X is
//	 the mana value of that spell."
//
// Tangleroot's any-player creature-cast trigger. It destroys every Reflection
// on the battlefield (a Reflection is a permanent with that creature type,
// whoever controls it, Reflections of other cards included), and only then,
// from the destruction's keyed tail, the caster makes the token: so a pause for
// a replacement on a destroyed Reflection can't let the new one die. X is the
// mana value of the spell (counting {X} as announced while it is on the stack,
// or as it last was there if it left before the trigger resolved, CR 608.2h).
// The token is the 0/0 row, set to X as it is created, like every X/X token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fee53820-a17c-4f34-9971-581bf8e12fd6",
		Name:         "Pure Reflection",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.OldB607CastACreatureSpell,
				"Pure Reflection — destroy all Reflections; then that player creates an X/X white Reflection creature token, where X is the mana value of that spell",
				pureReflectionResolve),
		},
	})
}

func pureReflectionResolve(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Actor == uuid.Nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	tail := ctx.TailN(pureReflectionToken, effects.SocB4TriggeringSpellManaValue(g, item))
	tail.Carry = []uuid.UUID{item.Trigger.Event.Actor}
	return effects.DestroyAllMatching{Match: effects.HasSubtype("Reflection"), ThenTail: tail}.Apply(ctx)
}

// pureReflectionToken is "Then that player creates an X/X white Reflection
// creature token": the caster rides the tail's carry and X its number. Keyed
// (#4253): an on-disk key, never renamed, never reused.
var pureReflectionToken = effects.MassTailOf("then/pure-reflection-create-token", func(ctx *effects.Context, _ []game.Card, _ int, t game.KeyedTail) error {
	if len(t.Carry) == 0 {
		return nil
	}
	tok := ctx.TextToken(effects.TokenCard("0/0 white Reflection"))
	tok.Power, tok.Toughness = t.N, t.N
	return ctx.Game.CreateTokenForEffect(t.Carry[0], tok, 1)
})
