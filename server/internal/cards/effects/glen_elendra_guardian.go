package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glen Elendra Guardian — Creature — Faerie Wizard {2}{U}, 3/4 (Lorwyn
// Eclipsed, #4458):
//
//	"Flash
//	 Flying
//	 This creature enters with a -1/-1 counter on it.
//	 {1}{U}, Remove a counter from this creature: Counter target
//	 noncreature spell. Its controller draws a card."
//
// "A counter" is one counter of a kind the activator names (Loch Mare's
// RemoveCountersFromThis("", 1)), so after the -1/-1 counter is gone a
// +1/+1 counter it picked up pays too. The spell's controller is read
// before the counter, and draws whether or not the spell is actually
// countered (a spell that can't be countered still names a legal
// target); with the target gone by resolution the ability does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "97febf6b-302c-4c41-8838-eef7c7353bdc",
		Name:            "Glen Elendra Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 1, "Glen Elendra Guardian: enters with a -1/-1 counter"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, Remove a counter from this creature: Counter target noncreature spell. Its controller draws a card.",
			Cost:    Plus(ManaCost("{1}{U}"), RemoveCountersFromThis("", 1)),
			Targets: TargetSpell("target noncreature spell", Noncreature()),
			Effect:  glenElendraGuardianCounter,
		}},
	})
}

// glenElendraGuardianCounter counters the target spell and has its
// controller draw a card.
func glenElendraGuardianCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		spell := g.StackItemForEffect(t.ID)
		if spell == nil {
			return nil
		}
		victim := spell.Controller
		if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return DrawCards{Player: victim, N: 1}.Apply(ctx)
	}
	return nil
}
