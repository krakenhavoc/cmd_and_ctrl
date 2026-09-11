package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Wandering Emperor — Legendary Planeswalker — Wanderer with
// starting loyalty 3 and three activated loyalty abilities:
//
//	+1: Create a 2/2 white Samurai creature token with vigilance.
//	−1: Exile target tapped creature.
//	−2: Up to one target creature gets +2/+1 and gains lifelink
//	    until end of turn.
//
// Also: "Flash" and "As long as The Wandering Emperor entered this
// turn, you may activate her loyalty abilities any time you could
// cast an instant."
//
// S14 registered her starting loyalty and left the abilities as a
// note for a future sprint ("S19's ability auto-fire sprint will
// wire OnResolve-style handlers to each loyalty ability via a new
// Spec.Abilities field; this registration is the attach point").
// This is that wiring, eight sprints late, arriving with #329 /
// #334: `AbilityCost.Loyalty` plus the CR 606 gates makes a loyalty
// ability an ordinary CR 602 activation, and all three of hers
// happened to already have primitives waiting.
//
// WHAT IS WIRED
//
//   - +1 in full. WhiteSamuraiToken has existed since S21 sub-PR 1
//     precisely so this ability could use it without a new template.
//   - −1 in full. The "tapped" restriction is a target predicate,
//     so an untapped creature never reaches the picker.
//   - −2 in full, as two primitives: the P/T change is layer 7c and
//     the lifelink grant is layer 6, and a single turn-scoped entry
//     cannot sort into both (see until_end_of_turn.go). Lifelink is
//     one of the twelve keywords the combat code honours, so the
//     life gain is real. "Up to one target" is Min 0, matching
//     Teferi's −3.
//
// WHAT IS NOT WIRED
//
//   - Flash. Printed keywords reach the cast gate through
//     CatalogPrintedKeywords, and hers come off Scryfall for an
//     imported deck — no catalog work needed, so nothing is claimed
//     here.
//   - "You may activate her loyalty abilities any time you could
//     cast an instant" while she entered this turn. That is a
//     per-permanent override of CR 606.5's sorcery-speed half, and
//     the gate that enforces it (activated.go) has no hook for one.
//     She is strictly slower than printed on the turn she lands.
func init() {
	Register(Spec{
		OracleID: "0c7f18d5-36cb-4bc6-a358-443b97666215",
		Name:     "The Wandering Emperor",
		// The fallback for tokens, fixtures and the dev spawner;
		// an imported deck reads printed loyalty (ADR 0032 §1).
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Create a 2/2 white Samurai creature token with vigilance.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return CreateToken{
						Controller: ctx.Controller(),
						Template:   WhiteSamuraiToken(),
						N:          1,
					}.Apply(ctx)
				},
			},
			{
				Label:   "−1: Exile target tapped creature.",
				Cost:    LoyaltyCost(-1),
				Targets: TargetCreature("target tapped creature", tappedPermanent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if len(item.Targets) == 0 {
						return nil
					}
					return ExileTarget{Target: item.Targets[0].ID}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Up to one target creature gets +2/+1 and gains lifelink until end of turn.",
				Cost:    LoyaltyCost(-2),
				Targets: upToOneCreature(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if len(item.Targets) == 0 {
						return nil
					}
					target := item.Targets[0].ID
					if err := (BoostUntilEOT{
						Target:    target,
						Power:     2,
						Toughness: 1,
						Label:     "The Wandering Emperor — +2/+1",
					}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{
						Target:   target,
						Keywords: []string{"lifelink"},
						Label:    "The Wandering Emperor — lifelink",
					}.Apply(ctx)
				},
			},
		},
	})
}

// tappedPermanent is "…that is tapped", the restriction on the
// Emperor's −1. A target predicate rather than a resolution-time
// check so an untapped creature never appears in the picker, which
// is what CR 115.4 wants: an illegal target can't be chosen.
func tappedPermanent() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Tapped }
}

// upToOneCreature is the Emperor's −2 clause. Min 0 is the "up to
// one": the picker confirms with nothing selected.
func upToOneCreature() *game.TargetSpec {
	spec := TargetCreature("up to one target creature")
	spec.Min = 0
	return spec
}
