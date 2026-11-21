package effects

// Adric, Mathematical Genius — Legendary Creature — Human Artificer {1}{U},
// 1/1 (Doctor Who, #5948):
//
//	"{2}{U}, {T}: Copy target activated or triggered ability you control.
//	 You may choose new targets for the copy.
//	 Ultimate Sacrifice — {1}{U}, Sacrifice Adric: Counter target activated
//	 or triggered ability.
//	 Doctor's companion"
//
// The first ability is Lithoform Engine's copy of an ability you control
// (copyTargetedAbility, new targets may be chosen); the second is Stifle's
// counter, paid by sacrificing Adric as a cost, so it still resolves when he
// is gone. Doctor's companion is a deck-construction rule.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af566158-23b2-4f91-8d59-a58342a9f576",
		Name:         "Adric, Mathematical Genius",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{U}, {T}: Copy target activated or triggered ability you control. You may choose new targets for the copy.",
				Cost:    Plus(ManaCost("{2}{U}"), TapCost()),
				Targets: AbilityOnStack("target activated or triggered ability you control", AnAbilityYouControl()),
				Effect:  copyTargetedAbility,
			},
			{
				Label:   "Ultimate Sacrifice — {1}{U}, Sacrifice Adric: Counter target activated or triggered ability.",
				Cost:    Plus(ManaCost("{1}{U}"), SacrificeThis()),
				Targets: AbilityOnStack("target activated or triggered ability"),
				Effect:  counterTheChosenSpell,
			},
		},
	})
}
