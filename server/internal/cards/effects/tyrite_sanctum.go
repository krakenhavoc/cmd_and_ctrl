package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tyrite Sanctum — Land:
//
//	"{T}: Add {C}.
//	 {2}, {T}: Target legendary creature becomes a God in addition to its
//	 other types. Put a +1/+1 counter on it.
//	 {4}, {T}, Sacrifice this land: Put an indestructible counter on
//	 target God."
//
// The God type has no stated duration, so it lasts while the creature
// stays on the battlefield (CR 611.2a): a pinned layer 4 subtype add, as
// Antarctic Research Base's Plant. Both clauses read the target as the
// ability resolves, and a target that has become illegal fizzles the
// whole ability (CR 608.2b). "Target God" is any permanent with the God
// subtype, a creature or not, read through the layered view; the land is
// sacrificed as a cost, so the counter lands even after the Sanctum has
// gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38c83332-2d21-450c-ae97-34109b565f59",
		Name:         "Tyrite Sanctum",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, {T}: Target legendary creature becomes a God in addition to its other types. Put a +1/+1 counter on it.",
				Cost:    Plus(ManaCost("{2}"), TapCost()),
				Targets: TargetCreature("target legendary creature", Legendary()),
				Effect: Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
					if err := (ScopedEffectFor{
						Target:   id,
						Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtype("God"))},
						Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
						Label:    "Tyrite Sanctum — a God in addition to its other types",
					}).Apply(ctx); err != nil {
						return err
					}
					return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				})),
			},
			{
				Label:   "{4}, {T}, Sacrifice this land: Put an indestructible counter on target God.",
				Cost:    Plus(ManaCost("{4}"), TapCost(), SacrificeThis()),
				Targets: TargetPermanent("target God", HasSubtype("God")),
				Effect: Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
					return AddCounter{Target: id, Kind: game.CounterIndestructible, N: 1}.Apply(ctx)
				})),
			},
		},
	})
}
