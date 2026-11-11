package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necropolis Fiend — Creature — Demon {7}{B}{B}, 4/5 (Tarkir: Dragonstorm
// Commander):
//
//	"Delve
//	 Flying
//	 {X}, {T}, Exile X cards from your graveyard: Target creature gets
//	 -X/-X until end of turn."
//
// Delve is `Delve: true` (CR 702.66, ADR 0100). The ability's X is announced
// by naming the cards: Winter, Cursed Rider's "exile X cards" clause
// (ExileXFromGraveyardAsCost, game.ExileCost.CountIsX) with the {X} of the
// mana cost sized by it. When a cost has both, the number of cards the
// activator names decides X (activateCatalogAbilityLocked), so the mana
// paid and the -X/-X always equal the cards exiled. The pump is paid in
// full before it can be responded to (CR 602.2b); a target gone by
// resolution makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bb714bc6-9f32-42f6-9064-f28a383675b5",
		Name:            "Necropolis Fiend",
		Completeness:    CompletenessFull,
		Delve:           true,
		XMatters:        true,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{X}, {T}, Exile X cards from your graveyard: Target creature gets -X/-X until end of turn.",
			Purpose: ForTargets(HarmUntilEOT(0, -1, -1)),
			Cost: Plus(ManaCost("{X}"), TapCost(),
				ExileXFromGraveyardAsCost("X cards", func(game.Card) bool { return true })),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard && x > 0 {
						return BoostUntilEOT{Target: t.ID, Power: -x, Toughness: -x, Label: "Necropolis Fiend — -X/-X"}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
