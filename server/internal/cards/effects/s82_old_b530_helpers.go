package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b530_helpers.go — shared pieces for the S82 OLD batch 530 cards
// (Shadowmoor damage, #5591).

// OldB530DivideMountains is Jaws of Stone's amount: the Mountains the caster
// controls, read as the spell is cast (the division is announced with the
// targets, CR 601.2d, before costs are paid), so a Mountain that goes away in
// response changes nothing. "Mountains" is the subtype, so a land another
// effect made a Mountain counts.
var OldB530DivideMountains = game.RegisterDivideAmount("mountains-you-control",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		return game.DivideSpec{Total: B08LandsWithSubtypeControlled(g, a.Controller, "Mountain")}
	})

// OldB530TargetBlocksThisAttacker is "target creature blocks it this turn if
// able" for the trigger's own target: the requirement names the attacker as the
// object it is now, and asks nothing of a creature that can't legally block it
// (CR 509.1c). A target that is no longer legal does nothing (CR 608.2b).
func OldB530TargetBlocksThisAttacker(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		blocker, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return BlocksAttackerUntilEOT{Blocker: blocker, Attacker: AttackingObjectOf(ctx), Label: label}.Apply(ctx)
	}
}

// OldB530CantBlockAndDamageWhiteAndBlue is Ember Gale: creatures the target
// player controls can't block this turn (a rules effect read live until
// cleanup, Headliner Scarlett's scope), and 1 damage to each white and/or blue
// creature that player controls, all at once from the spell.
func OldB530CantBlockAndDamageWhiteAndBlue(ctx *Context, player uuid.UUID, label string) error {
	ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeYourCreatures, player,
		[]game.Mod{game.AddRestrictionsMod(game.CantBlock)}, DurationUntilEndOfTurn(ctx), label)
	return DamageEachMatching(ctx, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.Controller == player && (c.HasColorWord("W") || c.HasColorWord("U"))
	}, 1)
}

// OldB530RemoveAllMinusOneCounters removes every -1/-1 counter from a
// permanent through the effect-side adder, so the layer pass sees them go.
func OldB530RemoveAllMinusOneCounters(ctx *Context, id uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || !OnBattlefield(ctx.Game, id) || c.Counters[game.CounterMinusOne] <= 0 {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterMinusOne, N: -c.Counters[game.CounterMinusOne]}.Apply(ctx)
}
