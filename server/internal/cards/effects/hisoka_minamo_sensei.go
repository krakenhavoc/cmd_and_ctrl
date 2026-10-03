package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hisoka, Minamo Sensei — Legendary Creature — Human Wizard {2}{U}{U}, 1/3:
//
//	"{2}{U}, Discard a card: Counter target spell if it has the same
//	 mana value as the discarded card."
//
// ADR 0109 §8 (#1862): the discarded card comes off the payment record
// (Context.DiscardedCard, CR 400.7j). The target is any spell, chosen
// as the ability is activated; the comparison is made as it resolves,
// with the spell's mana value on the stack — an {X} counts as the X
// chosen for it (CR 202.3e) — against the discarded card's in the
// graveyard, where an {X} is zero. A card whose cost can't be read
// matches nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b69701d6-8e46-4184-991b-9f5e95a5e16d",
		Name:         "Hisoka, Minamo Sensei",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}, Discard a card: Counter target spell if it has the same mana value as the discarded card.",
			Cost:    Plus(ManaCost("{2}{U}"), DiscardACard()),
			Targets: TargetSpell("target spell"),
			Effect:  hisokaCounter,
		}},
	})
}

// hisokaCounter counters the target when its mana value is the
// discarded card's.
func hisokaCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	discarded, ok := ctx.DiscardedCard()
	if !ok {
		return nil
	}
	want, ok := g.ManaValueForEffect(discarded)
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		spell, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		if mv, ok := g.ManaValueForEffect(spell); ok && mv == want {
			return CounterTarget{StackID: t.ID}.Apply(ctx)
		}
	}
	return nil
}
