package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Owen Grady, Raptor Trainer — Legendary Creature — Human Soldier Scientist
// {1}{R}{G}, 3/2 (Jurassic World Collection, #5948):
//
//	"Partner with Blue, Loyal Raptor
//	 {T}: Put your choice of a reach, menace, trample, or haste counter on
//	 target Dinosaur. Activate only as a sorcery."
//
// Partner with is the shared entry trigger (PartnerWith). The counter kind is
// chosen as the ability resolves (Aragorn's keyed pick), and the target is
// re-checked then: a Dinosaur that left gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4eda40a2-39e5-4b0e-ba6d-8a6f89d6f55f",
		Name:         "Owen Grady, Raptor Trainer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Owen Grady, Raptor Trainer", "Blue, Loyal Raptor"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Put your choice of a reach, menace, trample, or haste counter on target Dinosaur. Activate only as a sorcery.",
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target Dinosaur", effects.Subtype("Dinosaur")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if len(ctx.LegalTargets()) == 0 {
					return nil
				}
				options := make([]game.ChoiceOption, len(effects.GapB82OwenKinds))
				for i, kind := range effects.GapB82OwenKinds {
					options[i] = game.ChoiceOption{Label: "A " + kind + " counter"}
				}
				return effects.PickOption{
					Question: "Owen Grady — put your choice of a counter on target Dinosaur",
					Options:  options,
					ThenKey:  effects.GapB82OwenCounterChosen,
					WithItem: true,
				}.Apply(ctx)
			},
		}},
	})
}
