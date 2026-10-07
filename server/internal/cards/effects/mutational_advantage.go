package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mutational Advantage — Instant {1}{G}{U}:
//
//	"Permanents you control with counters on them gain hexproof and
//	 indestructible until end of turn. Prevent all damage that would be
//	 dealt to those permanents this turn. Proliferate."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the permanents are fixed as the
// spell resolves (CR 611.2c for the grant; "those permanents" names the
// same set for the shield), so a permanent that gets its first counter
// from the proliferate that follows gains nothing. The shield is one
// not-one-use record protecting all of them (ShieldObjects).
//
// The proliferate asks the player what to proliferate (#2525).
func init() {
	Register(Spec{
		OracleID:     "2daa5b89-e772-4a2b-ad52-bbbf148c7b2f",
		Name:         "Mutational Advantage",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var those []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller != item.Controller {
					continue
				}
				for _, n := range c.Counters {
					if n > 0 {
						those = append(those, c.InstanceID)
						break
					}
				}
			}
			for _, id := range those {
				if err := (GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"hexproof", "indestructible"},
					Label:    "Mutational Advantage — hexproof and indestructible",
				}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			if len(those) > 0 {
				if err := (PreventDamageFromSource{Protect: ShieldObjects(those...)}).Apply(ctx); err != nil {
					return err
				}
			}
			return Proliferate{}.Apply(ctx)
		},
	})
}
