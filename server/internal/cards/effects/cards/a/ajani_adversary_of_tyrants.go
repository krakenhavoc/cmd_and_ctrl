package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "b25bea73-7444-4ee2-8022-611461506117",
		Name:            "Ajani, Adversary of Tyrants",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &effects.EmblemSpec{
			Label: "Ajani, Adversary of Tyrants emblem",
			Text:  "At the beginning of your end step, create three 1/1 white Cat creature tokens with lifelink.",
			Triggered: []game.TriggeredAbility{
				effects.AtYourEndStep("Ajani, Adversary of Tyrants emblem — create three 1/1 white Cats with lifelink",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Cat with lifelink"), N: 3})),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Put a +1/+1 counter on each of up to two target creatures.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				Effect:  effects.PutPlusOneCounterOnEachLegalTarget,
			},
			{
				Label: "−2: Return target creature card with mana value 2 or less from your graveyard to the battlefield.",
				Cost:  effects.LoyaltyCost(-2),
				Targets: effects.TargetCardInGraveyard("target creature card with mana value 2 or less from your graveyard",
					effects.YouOwn(), effects.Creature(), effects.ManaValueLE(2)),
				Effect: effects.B26ReturnFirstLegalGraveyardTargetToBattlefield,
			},
			{
				Label:  "−7: You get an emblem with \"At the beginning of your end step, create three 1/1 white Cat creature tokens with lifelink.\"",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}
