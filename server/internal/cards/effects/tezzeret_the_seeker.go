package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tezzeret the Seeker — Legendary Planeswalker — Tezzeret {3}{U}{U},
// loyalty 4:
//
//	"+1: Untap up to two target artifacts.
//	 −X: Search your library for an artifact card with mana value X or
//	     less, put it onto the battlefield, then shuffle.
//	 −5: Artifacts you control become artifact creatures with base power
//	     and toughness 5/5 until end of turn."
//
// +1 untaps each still-legal target; picking none is legal (Min 0).
//
// −X is LoyaltyMinusX: X is announced with the activation, no more than
// Tezzeret's loyalty (CR 606.6), and read at resolution. X = 0 still
// finds a mana value 0 artifact, so the guard allowlists it rather than
// declaring XMatters. The search is SearchLibrary's, which shuffles
// afterwards; finding nothing is legal (CR 701.23b).
//
// −5 locks the set of artifacts you control as it resolves (CR 611.2c):
// one that arrives later this turn is not animated. Each becomes a
// creature (it was already an artifact) with base 5/5 in layer 7b, so
// counters and pumps still apply on top.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "4aeafdf3-7828-4114-8c74-3bd46357caec",
		Name:         "Tezzeret the Seeker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Untap up to two target artifacts.",
				Cost:    LoyaltyCost(1),
				Targets: TargetPermanent("up to two target artifacts", Artifact()).WithCount(0, 2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard || t.ID == uuid.Nil {
							continue
						}
						if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label: "−X: Search your library for an artifact card with mana value X or less, put it onto the battlefield, then shuffle.",
				Cost:  LoyaltyMinusX(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x := ctx.X()
					return SearchLibrary{
						Player: ctx.Controller(),
						Predicate: func(c game.Card) bool {
							return c.IsArtifact() && c.ManaValue() <= x
						},
						Dest:   game.ZoneBattlefield,
						Limit:  1,
						Reason: "Tezzeret the Seeker — an artifact card with mana value X or less",
						Source: ctx.Source(),
					}.Apply(ctx)
				},
			},
			{
				Label: "−5: Artifacts you control become artifact creatures with base power and toughness 5/5 until end of turn.",
				Cost:  LoyaltyCost(-5),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ScopedEffectFor{
						Match: And(Artifact(), YouControl()),
						Mods: []game.Mod{
							game.AddTypesMod("Artifact", "Creature"),
							game.SetBasePowerMod(5),
							game.SetBaseToughnessMod(5),
						},
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Tezzeret the Seeker — artifact creatures with base power and toughness 5/5",
					}.Apply(ctx)
				},
			},
		},
	})
}
