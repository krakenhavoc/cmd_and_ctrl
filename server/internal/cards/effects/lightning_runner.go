package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightning Runner — Creature — Human Warrior {3}{R}{R}, 2/2:
//
//	"Double strike, haste
//	 Whenever this creature attacks, you get {E}{E} (two energy
//	 counters), then you may pay eight {E}. If you pay, untap all
//	 creatures you control, and after this phase, there is an additional
//	 combat phase."
//
// ADR 0129 §3 (#1995): the energy is gotten, then the eight are asked
// for through the energy prompt (CR 118.12), held in the declare
// attackers step, so "after this phase" is the combat in progress
// (Hellkite Charger's anchor). The added combat has no main phase
// before it, and the Runner attacks again and gets two more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2f82b232-21ff-44cc-bc35-0999a1d1c52f",
		Name:            "Lightning Runner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "haste"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Lightning Runner — you get {E}{E}, then may pay eight {E}",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (GetEnergy{N: 2}).Apply(ctx); err != nil {
						return err
					}
					return MayPayEnergy{
						N:        8,
						Question: "Lightning Runner — pay eight {E} to untap all creatures you control and get an additional combat phase?",
						OnPay: func(ctx *Context) error {
							if err := (UntapAllCreaturesYouControl{}).Apply(ctx); err != nil {
								return err
							}
							return ExtraCombatAfterThisPhase().Apply(ctx)
						},
					}.Apply(ctx)
				}),
		},
	})
}
