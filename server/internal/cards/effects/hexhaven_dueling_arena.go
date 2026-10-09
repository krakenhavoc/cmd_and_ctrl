package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hexhaven Dueling Arena — Land:
//
//	"{T}: Add {C}.
//	 {2}, {T}: Target creature that attacked this turn becomes prepared.
//	 Activate only as a sorcery. (Only creatures with prepare spells can
//	 become prepared.)
//	 {4}, {T}: Target creature becomes prepared."
//
// "Attacked this turn" is the per-object attack tally (AttackedThisTurn),
// read when the target is chosen and again as the ability resolves
// (CR 608.2b). A target with no prepare spell, or one already prepared,
// is a legal target that simply does nothing (CR 722.3a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2cc075da-f79b-4050-84eb-2c1b9a30260c",
		Name:         "Hexhaven Dueling Arena",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{
			{
				Label:        "{2}, {T}: Target creature that attacked this turn becomes prepared. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{2}"), TapCost()),
				Targets:      TargetCreature("target creature that attacked this turn", AttackedThisTurn()),
				SorcerySpeed: true,
				Effect:       hexhavenPrepareTarget,
			},
			{
				Label:   "{4}, {T}: Target creature becomes prepared.",
				Cost:    Plus(ManaCost("{4}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  hexhavenPrepareTarget,
			},
		},
	})
}

func hexhavenPrepareTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return BecomePrepared{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
}
