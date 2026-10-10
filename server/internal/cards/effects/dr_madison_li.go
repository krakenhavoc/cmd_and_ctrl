package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9c4ff8fe-7d69-42b9-bad1-9e2c8a3e29f1",
		Name:         "Dr. Madison Li",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WheneverYouCast(Artifact(), "Dr. Madison Li — you get {E}", Do(GetEnergy{N: 1})),
				game.Purpose{Energy: 1}),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{T}, Pay {E}: Target creature gets +1/+0 and gains trample and haste until end of turn.",
				Cost:    Plus(TapCost(), PayEnergy(1)),
				Targets: TargetCreature("target creature"),
				Effect:  pumpAndGrantFirstTarget(1, 0, "Dr. Madison Li — +1/+0, trample and haste", "trample", "haste"),
			},
			{
				Label:   "{T}, Pay {E}{E}{E}: Draw a card.",
				Cost:    Plus(TapCost(), PayEnergy(3)),
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Effect:  Do(DrawCards{N: 1}),
			},
			{
				Label:   "{T}, Pay {E}{E}{E}{E}{E}: Return target artifact card from your graveyard to the battlefield tapped.",
				Cost:    Plus(TapCost(), PayEnergy(5)),
				Targets: TargetCardInGraveyard("target artifact card from your graveyard", Artifact(), YouOwn()),
				Effect:  returnFirstGraveyardTargetTapped(false),
			},
		},
	})
}
