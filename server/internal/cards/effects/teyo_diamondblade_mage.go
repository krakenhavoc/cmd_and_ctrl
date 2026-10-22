package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teyo, Diamondblade Mage — Legendary Creature — Human Warlock {3}{B}, 3/1
// (Reality Fracture, tracker #2795):
//
//	"Flash
//	 When Teyo enters, target permanent you control gains deathtouch until
//	 end of turn. Put a +1/+1 counter on it if it's a creature. Put a
//	 loyalty counter on it if it's a planeswalker."
//
// Teyo may target himself. The two "if" clauses read the target's types as
// the trigger resolves, so an animated planeswalker gets both. The loyalty
// counter names its placer, so "whenever you put loyalty counters on a
// planeswalker" payoffs see it. Teyo's sibling, Lightshield Expert, shares
// the body (teyoEntersTrigger).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3bfbe1a-7831-45bd-9e42-790c28f6054c",
		Name:            "Teyo, Diamondblade Mage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{teyoEntersTrigger("Teyo, Diamondblade Mage — target permanent you control gains deathtouch until end of turn", "deathtouch")},
	})
}

// teyoEntersTrigger is the shared Teyo ETB: target permanent you control
// gains `keyword` until end of turn, then a +1/+1 counter if it is a
// creature and a loyalty counter if it is a planeswalker.
func teyoEntersTrigger(label, keyword string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventETB},
		AppliesTo: Self,
		Key:       label,
		Targets:   TargetPermanent("target permanent you control", YouControl()),
		// The counters wait for a prompt the keyword's grant asks, and the loyalty counter for
		// one the +1/+1 counter asks (a CR 616 order; CR 608.2c, #4155).
		Effect: Do(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{keyword}, Label: label}.Apply(ctx)
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok && c.IsCreature() {
						return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
					}
					return nil
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok && c.IsPlaneswalker() {
						return nothingIfGone(ctx.Game.AddCounterByForEffect(ctx.Controller(), t.ID, game.CounterLoyalty, 1))
					}
					return nil
				}
				return nil
			}),
		),
	}
}
