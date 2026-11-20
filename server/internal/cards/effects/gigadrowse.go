package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gigadrowse — Instant {U} (Guildpact, #5599):
//
//	"Replicate {U} (When you cast this spell, copy it for each time you
//	 paid its replicate cost. You may choose new targets for the copies.)
//	 Tap target permanent."
//
// Replicate is Shattering Spree's pair (replicateMana for the cost,
// Replicate() for the cast trigger, CR 702.56a); `max` caps one
// announcement at twenty payments, far beyond any mana a table holds.
// Each copy taps the permanent it targets (tapEachLegalTarget, nothing if
// that target left in response, CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "9cc64a91-82d2-4be1-812f-531863e99801",
		Name:          "Gigadrowse",
		Completeness:  CompletenessFull,
		Targets:       TargetPermanent("target permanent"),
		OptionalCosts: []game.AdditionalCost{replicateMana("{U}", 20)},
		Triggered:     []game.TriggeredAbility{Replicate()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tapEachLegalTarget(ctx.Game, item)
		},
	})
}
