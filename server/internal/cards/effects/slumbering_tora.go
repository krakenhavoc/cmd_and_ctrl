package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slumbering Tora — Artifact {3}:
//
//	"{2}, Discard a Spirit or Arcane card: This artifact becomes an X/X
//	 Cat artifact creature until end of turn, where X is the discarded
//	 card's mana value."
//
// ADR 0109 §8 (#1862): X is the discarded card's mana value, read off
// the payment record (Context.DiscardedManaValue, CR 400.7j) as the
// ability resolves and fixed then (CR 608.2h). The animation is one
// until-end-of-turn record: layer 4 adds Creature and Cat, layer 7b
// sets the base power and toughness to X/X. A 0-mana-value card makes
// a 0/0, which the toughness state-based action puts into the
// graveyard (CR 704.5f), as printed. "This artifact" pinned to the
// object it was: a Tora that left and came back is a new object and is
// not animated (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9317e70d-3d01-4652-b9ea-b4015a0f7b80",
		Name:         "Slumbering Tora",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Discard a Spirit or Arcane card: This artifact becomes an X/X Cat artifact creature until end of turn, where X is the discarded card's mana value.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost: Plus(ManaCost("{2}"), DiscardCardsMatching(1, "a Spirit or Arcane card", func(c game.Card) bool {
				return c.HasSubtype("Spirit") || c.HasSubtype("Arcane")
			})),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.DiscardedManaValue()
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Artifact", "Creature"),
						game.AddSubtypesMod("Cat"),
					}, game.SetBasePTMods(x, x)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Slumbering Tora — becomes an X/X Cat artifact creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
