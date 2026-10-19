package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hazoret's Undying Fury — Sorcery {4}{R}{R}:
//
//	"Shuffle your library, then exile the top four cards. You may cast any
//	 number of spells with mana value 5 or less from among them without
//	 paying their mana costs. Lands you control don't untap during your
//	 next untap step."
//
// The shuffle is the controller's, then the top four are exiled with ONE
// pass-closed permission over the cards that landed (#4021,
// ExileTopWithPermission.EachManaValueAtMost): each cast is free, capped at
// mana value 5 on its own and, unlike Primeval Spawn's running total, spends
// nothing but its own card, so any number of them may be cast (CR 608.2g).
// The casts are taken after the spell resolves and the window closes when
// the caster next passes priority, as for every resolution-time cast (ADR
// 0066); a card not cast, a land among the four and a spell over mana value
// 5 stay in exile. The last clause is #3995's player-wide skip over the
// caster's lands, read at their next untap step, so a land played before
// then stays tapped too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af4d3489-6bba-4538-928c-50c9a7073813",
		Name:         "Hazoret's Undying Fury",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (ShuffleLibrary{Player: item.Controller}).Apply(ctx); err != nil {
				return err
			}
			limit := 5
			if err := (ExileTopWithPermission{
				From:                item.Controller,
				GrantTo:             item.Controller,
				N:                   4,
				EachManaValueAtMost: &limit,
			}).Apply(ctx); err != nil {
				return err
			}
			return LandsYouControlDontUntapNextStep(ctx)
		},
	})
}
