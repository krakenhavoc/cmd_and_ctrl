package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shieldmage Advocate — Creature — Human Cleric {2}{W}:
//
//	"{T}: Return target card from an opponent's graveyard to their hand. Prevent all damage that would be dealt to any target this turn by a source of your choice."
//
// ADR 0108 §7 (#1904): two target clauses, then a shield against a
// source chosen as the ability resolves (CR 609.7a) that protects the
// second target from every instance of that source's damage this turn.
// A target that has become illegal is skipped and the rest still happens
// (CR 608.2b): with the graveyard card gone the shield is still made, and
// with the shielded target gone the card still goes back.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "57188b3d-567c-4dff-8b93-7cc7a47894be",
		Name:         "Shieldmage Advocate",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}: Return target card from an opponent's graveyard to their hand. Prevent all damage that would be dealt to any target this turn by a source of your choice.",
			Cost:  TapCost(),
			Targets: Clauses(
				TargetCardInGraveyard("target card from an opponent's graveyard", notOwnedByCaster()),
				TargetAny(),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
					if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
						return err
					}
				}
				t, ok := ctx.ClauseTarget(1)
				if !ok {
					return nil
				}
				return PreventDamageFromChosenSource(ShieldObject(t.ID)).Apply(ctx)
			},
		}},
	})
}
