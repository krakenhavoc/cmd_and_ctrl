package effects

// Azorius Guildmage — Creature — Vedalken Wizard {W/U}{W/U}, 2/2 (Ravnica
// Remastered, #5948):
//
//	"{2}{W}: Tap target creature.
//	 {2}{U}: Counter target activated ability. (Mana abilities can't be
//	 targeted.)"
//
// The tap is the ordinary targeted tap. The counter targets an activated
// ability on the stack (Reroute's AnActivatedAbilityItem): a spell, a triggered
// ability and a mana ability, which never uses the stack, are not legal targets.
// Countering an ability removes it from the stack without resolving it
// (CR 701.5a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16f61f69-cdde-4b6c-b855-c74db9038ceb",
		Name:         "Azorius Guildmage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{W}: Tap target creature.",
				Cost:    ManaCost("{2}{W}"),
				Targets: TargetCreature("target creature"),
				Effect:  tapChosenPermanent,
			},
			{
				Label:   "{2}{U}: Counter target activated ability.",
				Cost:    ManaCost("{2}{U}"),
				Targets: TargetSpellOrAbility("target activated ability", AnActivatedAbilityItem()),
				Effect:  counterTheChosenSpell,
			},
		},
	})
}
