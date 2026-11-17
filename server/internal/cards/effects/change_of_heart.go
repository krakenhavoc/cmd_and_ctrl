package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Change of Heart — Instant {W} (Stronghold, #5627):
//
//	"Buyback {3} (You may pay an additional {3} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Target creature can't attack this turn."
//
// The can't-attack restriction until end of turn (RestrictUntilEOT,
// Oriss's game.CantAttack bit on one creature), checked as attackers are
// declared (CR 508.1c) and ending in the cleanup step (CR 514.2), with
// Capsize's buyback (CR 702.27a). A creature that has already attacked
// stays attacking: the restriction only stops a declaration. A target
// that became illegal in response means the spell doesn't resolve
// (CR 608.2b), so it goes to the graveyard, buyback paid or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "a8c0b41b-916c-4da9-9c39-a058c91a78f4",
		Name:          "Change of Heart",
		Completeness:  CompletenessFull,
		Purpose:       ForTargets(HarmUntilEOT(0, 0, 0)),
		Targets:       TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{Buyback("{3}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return RestrictUntilEOT{Target: item.Targets[0].ID, Restrictions: game.CantAttack,
				Label: "Change of Heart — can't attack this turn"}.Apply(ctx)
		},
	})
}
