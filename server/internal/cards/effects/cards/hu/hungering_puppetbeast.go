package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
		return effects.ModeSteps("Put a +1/+1 counter on it. It gains "+keyword+" until end of turn.", nil,
			effects.Instruction(func(ctx *effects.Context) error {
				self := ctx.Source()
				if !effects.OnBattlefield(ctx.Game, self) {
					return nil
				}
				return effects.AddCounter{Target: self, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				self := ctx.Source()
				if !effects.OnBattlefield(ctx.Game, self) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   self,
					Keywords: []string{keyword},
					Label:    "Hungering Puppetbeast — " + keyword + " until end of turn",
				}.Apply(ctx)
			}))
	}
	effects.Register(effects.Spec{
		OracleID:     "690ae865-87bd-46b2-8e65-c53bd80c1a18",
		Name:         "Hungering Puppetbeast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hungering Puppetbeast — create a Heartwood token",
				effects.Do(effects.CreateToken{Template: effects.HeartwoodToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice another artifact: Put a +1/+1 counter on this creature. It gains your choice of trample, hexproof, or haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerProtect | game.AnswerCombatGrant},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another artifact", effects.Artifact())),
			Modes:   effects.ChooseOne(grant("trample"), grant("hexproof"), grant("haste")),
		}},
	})
}
