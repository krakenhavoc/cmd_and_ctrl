package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deepwood Elder — Creature — Dryad Spellshaper {G}{G}, 2/2:
//
//	"{X}{G}{G}, {T}, Discard a card: X target lands become Forests until
//	 end of turn."
//
// The X announced with the activation is the number of targets (CR
// 601.2c, 602.2b; TargetSpec.CountFromX). Each target still legal as it
// resolves (CR 608.2b) becomes a Forest until end of turn: ADR 0109 §1's
// (#1881) CR 305.7 type set, one effect over the set the targets name. The
// lands lose their other land types and their rules-text abilities, keep
// their other subtypes (CR 205.1a) and tap for {G} (CR 305.6).
//
// No simplification.
func init() {
	targets := TargetPermanent("X target lands", Land())
	targets.CountFromX = true
	Register(Spec{
		OracleID:     "ffb62e4a-802d-4f1e-b92e-ec0919baafd5",
		Name:         "Deepwood Elder",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{X}{G}{G}, {T}, Discard a card: X target lands become Forests until end of turn.",
			Cost:    Plus(ManaCost("{X}{G}{G}"), TapCost(), DiscardACard()),
			Targets: targets,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				lands := map[uuid.UUID]bool{}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						lands[t.ID] = true
					}
				}
				if len(lands) == 0 {
					return nil
				}
				return LandBecomes{
					Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return lands[c.InstanceID] },
					Types:    []string{"Forest"},
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Deepwood Elder — those lands are Forests",
				}.Apply(ctx)
			},
		}},
	})
}
