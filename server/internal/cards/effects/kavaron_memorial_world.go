package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kavaron, Memorial World — Land — Planet (Edge of Eternities, #4461):
//
//	"This land enters tapped.
//	 {T}: Add {R}.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Planet. Station only as a sorcery.)
//	 12+ | {1}{R}, {T}, Sacrifice a land: Create a 2/2 colorless Robot
//	 artifact creature token, then creatures you control get +1/+0 and
//	 gain haste until end of turn."
//
// Susur Secundi's shape: a station Planet land with a charge-gated
// activated ability (ActiveWhen, absent below twelve). The land
// sacrificed is Hammer of Purphoros's cost and may be Kavaron itself. The
// Robot is made first, so it is pumped and hasted with the rest; the pump
// waits for any prompt the token's creation asks (CR 608.2c) and covers
// the creatures you control as it runs. It is an instant-speed ability
// with no target, so it declares what it answers with: a blocker, a pump
// and a combat grant. It declares no TeamPump, whose text is a team pump
// alone (#4320); this row makes a Robot first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4fa826ca-d361-4391-ad0d-989ebcfa4a91",
		Name:         "Kavaron, Memorial World",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []ActivatedAbility{
			Station(),
			{
				Label:      "{1}{R}, {T}, Sacrifice a land: Create a 2/2 colorless Robot artifact creature token, then creatures you control get +1/+0 and gain haste until end of turn.",
				Purpose:    game.Purpose{Answers: game.AnswerMakesBlocker | game.AnswerPump | game.AnswerCombatGrant},
				Cost:       Plus(ManaCost("{1}{R}"), TapCost(), game.AbilityCost{SacrificeOther: b29SacrificeALand()}),
				ActiveWhen: AtChargeCounters(12),
				Effect: Do(
					CreateToken{Template: TokenCard("2/2 colorless Robot artifact"), N: 1},
					Instruction(func(ctx *Context) error {
						return untilEndOfTurn(ctx, uuid.Nil, And(Creature(), YouControl()),
							"Kavaron, Memorial World — +1/+0 and haste", game.ModifyPTMod(1, 0), game.AddKeywordsMod("haste"))
					}),
				),
			},
		},
	})
}
