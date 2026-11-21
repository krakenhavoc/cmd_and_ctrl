package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Izzet Guildmage — Creature — Human Wizard {U/R}{U/R}, 2/2 (Ravnica
// Remastered, #5948):
//
//	"{2}{U}: Copy target instant spell you control with mana value 2 or less.
//	 You may choose new targets for the copy.
//	 {2}{R}: Copy target sorcery spell you control with mana value 2 or less.
//	 You may choose new targets for the copy."
//
// Reverberate's CopySpell with the new-targets offer (CR 707.10c), narrowed to
// "you control" and a mana value of 2 or less, read off the spell as it
// sits on the stack (CR 202.3). The copy is put onto the stack, not cast
// (CR 707.10), so it triggers nothing that watches a spell being cast. An
// activated ability on the stack cannot be targeted: only the spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34778e85-cd02-40c8-855c-4f4529aa458b",
		Name:         "Izzet Guildmage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{U}: Copy target instant spell you control with mana value 2 or less. You may choose new targets for the copy.",
				Cost:    ManaCost("{2}{U}"),
				Targets: TargetSpell("target instant spell you control with mana value 2 or less", Instant(), YouControl(), ManaValueLE(2)),
				Effect:  gapB74CopyTheTargetSpell,
			},
			{
				Label:   "{2}{R}: Copy target sorcery spell you control with mana value 2 or less. You may choose new targets for the copy.",
				Cost:    ManaCost("{2}{R}"),
				Targets: TargetSpell("target sorcery spell you control with mana value 2 or less", Sorcery(), YouControl(), ManaValueLE(2)),
				Effect:  gapB74CopyTheTargetSpell,
			},
		},
	})
}

// gapB74CopyTheTargetSpell is "Copy target spell. You may choose new targets
// for the copy." for the first target still legal as the ability resolves.
func gapB74CopyTheTargetSpell(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return CopySpell{StackID: t.ID, Controller: item.Controller, ChooseNewTargets: true}.Apply(ctx)
	}
	return nil
}
