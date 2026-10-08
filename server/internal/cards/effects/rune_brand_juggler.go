package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rune-Brand Juggler — Creature — Human Shaman {B}{R}:
//
//	"When this creature enters, suspect up to one target creature you
//	 control. (A suspected creature has menace and can't block.)
//	 {3}{B}{R}, Sacrifice a suspected creature: Target creature gets
//	 -5/-5 until end of turn."
//
// The enters trigger is the shared "suspect up to one target" clause
// (SuspectEachLegalTarget over UpToOneTargetCreature). The activation's
// cost clause is the Suspected() predicate on a sacrifice spec: the
// Juggler can feed itself when it is the suspected creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c105948-4b8f-4615-b66a-2bbdb3f3d6dd",
		Name:         "Rune-Brand Juggler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Rune-Brand Juggler — suspect up to one target creature you control", SuspectEachLegalTarget),
			UpToOneTargetCreature("up to one target creature you control", YouControl()),
		)},
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}{R}, Sacrifice a suspected creature: Target creature gets -5/-5 until end of turn.",
			Cost:    Plus(ManaCost("{3}{B}{R}"), SacrificeASuspectedCreature()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (BoostUntilEOT{
						Target: t.ID, Power: -5, Toughness: -5,
						Label: "Rune-Brand Juggler — -5/-5 until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
