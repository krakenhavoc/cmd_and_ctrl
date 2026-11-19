package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ajani, Adversary of Tyrants — Legendary Planeswalker — Ajani {2}{W}{W},
// loyalty 4 (Foundations Jumpstart, #5929):
//
//	"+1: Put a +1/+1 counter on each of up to two target creatures.
//	 −2: Return target creature card with mana value 2 or less from your
//	 graveyard to the battlefield.
//	 −7: You get an emblem with "At the beginning of your end step, create
//	 three 1/1 white Cat creature tokens with lifelink.""
//
// The +1 names up to two targets as the ability is put on the stack and puts
// a counter on each one that is still legal as it resolves (CR 608.2b). The
// −2 targets a creature card in your graveyard and returns it under its
// owner's control (yours). The emblem is Kiora, the Crashing Wave's shape: one
// end-step trigger of yours (CR 114.4) making three tokens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b25bea73-7444-4ee2-8022-611461506117",
		Name:            "Ajani, Adversary of Tyrants",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Ajani, Adversary of Tyrants emblem",
			Text:  "At the beginning of your end step, create three 1/1 white Cat creature tokens with lifelink.",
			Triggered: []game.TriggeredAbility{
				AtYourEndStep("Ajani, Adversary of Tyrants emblem — create three 1/1 white Cats with lifelink",
					Do(CreateToken{Template: TokenCard("1/1 white Cat with lifelink"), N: 3})),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Put a +1/+1 counter on each of up to two target creatures.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to two target creatures").WithCount(0, 2),
				Effect:  putPlusOneCounterOnEachLegalTarget,
			},
			{
				Label: "−2: Return target creature card with mana value 2 or less from your graveyard to the battlefield.",
				Cost:  LoyaltyCost(-2),
				Targets: TargetCardInGraveyard("target creature card with mana value 2 or less from your graveyard",
					YouOwn(), Creature(), ManaValueLE(2)),
				Effect: b26ReturnFirstLegalGraveyardTargetToBattlefield,
			},
			{
				Label:  "−7: You get an emblem with \"At the beginning of your end step, create three 1/1 white Cat creature tokens with lifelink.\"",
				Cost:   LoyaltyCost(-7),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}
