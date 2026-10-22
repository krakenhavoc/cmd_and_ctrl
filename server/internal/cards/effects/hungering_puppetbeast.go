package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hungering Puppetbeast — Artifact Creature — Beast Construct
// {3}{G}{G}, 5/5:
//
//	"When this creature enters, create a Heartwood token. (It's a red
//	 and green artifact with "{T}: Add {R} or {G}.")
//	 {1}, Sacrifice another artifact: Put a +1/+1 counter on this
//	 creature. It gains your choice of trample, hexproof, or haste until
//	 end of turn."
//
// "Your choice of" is a modal activated ability (ADR 0065, Cankerbloom's
// shape): the keyword is picked as the ability is activated, and every
// mode also puts the counter. The {1} and the sacrifice are paid up
// front.
//
// No simplification.
func init() {
	// The keyword waits for a prompt the counter asks (a CR 616 order; CR 608.2c, #4155).
	grant := func(keyword string) game.ModeOption {
		return ModeSteps("Put a +1/+1 counter on it. It gains "+keyword+" until end of turn.", nil,
			Instruction(func(ctx *Context) error {
				self := ctx.Source()
				if !onBattlefield(ctx.Game, self) {
					return nil
				}
				return AddCounter{Target: self, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				self := ctx.Source()
				if !onBattlefield(ctx.Game, self) {
					return nil
				}
				return GrantKeywordUntilEOT{
					Target:   self,
					Keywords: []string{keyword},
					Label:    "Hungering Puppetbeast — " + keyword + " until end of turn",
				}.Apply(ctx)
			}))
	}
	Register(Spec{
		OracleID:     "690ae865-87bd-46b2-8e65-c53bd80c1a18",
		Name:         "Hungering Puppetbeast",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hungering Puppetbeast — create a Heartwood token",
				Do(CreateToken{Template: HeartwoodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice another artifact: Put a +1/+1 counter on this creature. It gains your choice of trample, hexproof, or haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerProtect | game.AnswerCombatGrant},
			Cost:    Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another artifact", Artifact())),
			Modes:   ChooseOne(grant("trample"), grant("hexproof"), grant("haste")),
		}},
	})
}
