package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tecutlan, the Searing Rift — Legendary Land — Cave, the back face of
// Brass's Tunnel-Grinder (brasss_tunnel_grinder.go), registered under
// "<oracle_id>#1":
//
//	"(Transforms from Brass's Tunnel-Grinder.)
//	 {T}: Add {R}.
//	 Whenever you cast a permanent spell using mana produced by
//	 Tecutlan, discover X, where X is that spell's mana value."
//
// The trigger is a spend rider (#1547) on the {R}: it fires when that
// mana pays for a permanent spell — an artifact, creature, enchantment,
// planeswalker or battle spell — and goes on the stack above it. The
// {R} itself is unrestricted. X is the spell's mana value as the
// trigger resolves (CR 608.2h), with the X it was cast for counted
// while it is on the stack (CR 202.3e); the spell is named on the
// item's payload, so one countered in response is still named.
//
// Discover is ADR 0099's (game/discover.go).
//
// One declared simplification, weaker than printed and shared with
// every spend rider: with strict mana off the pool is never spent
// (ADR 0068 §3), so the trigger never fires.
func init() {
	Register(Spec{
		OracleID:     brassTunnelGrinderOracleID + "#1",
		Name:         "Tecutlan, the Searing Rift",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"With strict mana off, the game doesn't see which mana you spent, so casting a permanent spell with Tecutlan's mana does nothing extra.",
		},
		Discovers: true,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "{T}: Add {R}",
			SpendRiders: []game.ManaSpendRider{WhenManaSpent("Tecutlan, the Searing Rift", game.ManaSpendTrigger{
				Label:  "Tecutlan, the Searing Rift — discover X",
				Effect: tecutlanDiscover,
			}, ManaRestrictCast, ManaRestrictAnyType("Artifact", "Creature", "Enchantment", "Planeswalker", "Battle"))},
		}},
	})
}

// tecutlanDiscover is "discover X, where X is that spell's mana value".
func tecutlanDiscover(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := 0
	if spells := ctx.PayloadCards(); len(spells) > 0 {
		if spell, ok := g.LookupCardForEffect(spells[0]); ok {
			x, _ = g.ManaValueForEffect(spell)
		}
	}
	return Discover{N: x}.Apply(ctx)
}
