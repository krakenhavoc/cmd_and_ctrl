package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Run for Your Life — Instant {U}{R} (Doctor Who):
//
//	"One or two target creatures each gain haste until end of turn. They
//	 can't be blocked this turn except by creatures with haste.
//	 Escape—{2}{U}{R}, Exile four other cards from your graveyard. (You may
//	 cast this card from your graveyard for its escape cost.)"
//
// The haste is a keyword grant until end of turn over the targets that are
// still legal; the evasion is the shared "can't be blocked this turn except
// by creatures with haste" block rule on each of them, read when blockers are
// declared (so a creature that gains haste later still counts). Escape is the
// engine's (CR 702.138).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c95bedcc-911f-4c0f-a148-1cf5787e82bd",
		Name:             "Run for Your Life",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("one or two target creatures").WithCount(1, 2),
		Purpose:          effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Escape("{2}{U}{R}", 4)},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"},
					Label: "Run for Your Life — haste until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.CantBeBlockedThisTurnExceptBy{
					Target:   t.ID,
					Keywords: []string{"haste"},
					Text:     "creatures with haste",
					Label:    "Run for Your Life — can't be blocked except by creatures with haste",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
