package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dr. Madison Li — Legendary Creature — Human Scientist {U}{R}{W}, 2/3:
//
//	"Whenever you cast an artifact spell, you get {E} (an energy
//	 counter).
//	 {T}, Pay {E}: Target creature gets +1/+0 and gains trample and haste
//	 until end of turn.
//	 {T}, Pay {E}{E}{E}: Draw a card.
//	 {T}, Pay {E}{E}{E}{E}{E}: Return target artifact card from your
//	 graveyard to the battlefield tapped."
//
// ADR 0129 §2 (#1995): each row's "Pay N {E}" is the energy cost
// component. The {T} rows need her to have been under her controller's
// control since the turn began (CR 302.6). The returned artifact comes
// back under its owner's control, which is hers: "your graveyard".
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c4ff8fe-7d69-42b9-bad1-9e2c8a3e29f1",
		Name:         "Dr. Madison Li",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WheneverYouCast(effects.Artifact(), "Dr. Madison Li — you get {E}", effects.Do(effects.GetEnergy{N: 1})),
				game.Purpose{Energy: 1}),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Pay {E}: Target creature gets +1/+0 and gains trample and haste until end of turn.",
				Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "trample", "haste")), game.AnswerPump|game.AnswerCombatGrant),
				Cost:    effects.Plus(effects.TapCost(), effects.PayEnergy(1)),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.PumpAndGrantFirstTarget(1, 0, "Dr. Madison Li — +1/+0, trample and haste", "trample", "haste"),
			},
			{
				Label:   "{T}, Pay {E}{E}{E}: Draw a card.",
				Cost:    effects.Plus(effects.TapCost(), effects.PayEnergy(3)),
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "{T}, Pay {E}{E}{E}{E}{E}: Return target artifact card from your graveyard to the battlefield tapped.",
				Cost:    effects.Plus(effects.TapCost(), effects.PayEnergy(5)),
				Targets: effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn()),
				Effect:  effects.ReturnFirstGraveyardTargetTapped(false),
			},
		},
	})
}
