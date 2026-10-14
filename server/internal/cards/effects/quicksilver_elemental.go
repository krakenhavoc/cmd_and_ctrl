package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quicksilver Elemental — Creature — Elemental {3}{U}{U}, 3/4:
//
//	"{U}: This creature gains all activated abilities of target creature
//	 until end of turn. (If any of the abilities use that creature's
//	 name, use this creature's name instead.)
//	 You may spend blue mana as though it were mana of any color to pay
//	 the activation costs of this creature's abilities."
//
// The gain is a resolved borrow (ADR 0155 Decision 9): the rows the
// target has as the ability resolves — its own, a grant on it, a row it
// borrowed itself — are captured as data and lent to this Elemental
// until end of turn, so it keeps them after the target leaves (CR
// 608.2h). The reminder text is CR 201.5b, which the borrowed row's label
// already applies. A target that has gone, or one the engine does not
// automate, lends nothing.
//
// The rider is a one-colour any-colour spend over this permanent's own
// abilities, lent ones included: blue mana pays any coloured symbol of
// their costs, and a {C} still needs colourless mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c88e3aff-c539-4e8b-8ee1-a43b3d0962b4",
		Name:         "Quicksilver Elemental",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}: This creature gains all activated abilities of target creature until end of turn.",
			Cost:    ManaCost("{U}"),
			Targets: TargetCreature("target creature"),
			Effect:  quicksilverElementalBorrow,
		}},
		AnyColorSpend: []game.AnyColorSpendStatic{{
			Label:        "You may spend blue mana as though it were mana of any color to pay the activation costs of this creature's abilities.",
			Whose:        game.AnyColorSpendYou,
			From:         "U",
			OwnAbilities: true,
		}},
	})
}

// quicksilverElementalBorrow lends the target's activated abilities, as
// they are now, to the source until end of turn.
func quicksilverElementalBorrow(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		lender, ok := g.LookupCardPtrForEffect(t.ID)
		if !ok {
			return nil
		}
		return BorrowAbilitiesFor{
			Target: item.SourceCardID,
			Lender: lender,
			Label:  "Quicksilver Elemental — gains all activated abilities of " + lender.Name + " until end of turn",
		}.Apply(ctx)
	}
	return nil
}
