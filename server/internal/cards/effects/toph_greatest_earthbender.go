package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Toph, Greatest Earthbender — Legendary Creature — Human Warrior Ally
// {2}{R}{G}, 3/3 (Avatar: The Last Airbender Eternal, #5075):
//
//	"When Toph enters, earthbend X, where X is the amount of mana spent
//	 to cast her.
//	 Land creatures you control have double strike."
//
// The enter trigger is the set's earthbend with its keyword target
// ("target land you control", chosen as the trigger goes on the stack)
// and X read off the permanent's cast record as it resolves: the total
// mana that paid for the spell that became Toph. A Toph that was not
// cast (put onto the battlefield) had nothing spent, so X is 0 — the
// land still becomes a 0/0 and returns when it dies (the keyword's
// zero case). The static is a layer-6 grant over every permanent you
// control that is both a land and a creature, re-read each layer pass,
// so a land earthbent later has double strike from the moment it is a
// creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "3741206f-8ef5-4b8f-ae4c-df3ab0a17683",
		Name:           "Toph, Greatest Earthbender",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Toph, Greatest Earthbender — earthbend X, where X is the amount of mana spent to cast her",
				avB14Earthbend(func(ctx *Context) int { return ctx.ManaSpentToCastThis().Total() })),
				EarthbendTargets()),
		},
		Static: []game.StaticAbility{
			avB27LandCreaturesYouControlHave("double strike"),
		},
	})
}
