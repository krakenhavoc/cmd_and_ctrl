package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deep Analysis — Sorcery {3}{U}:
//
//	"Target player draws two cards.
//	 Flashback—{1}{U}, Pay 3 life."
//
// The flashback card whose flashback is cheaper in mana than its
// printed cost, which is exactly why Flashback binds the offer to the
// graveyard: claimable from hand, the {1}{U} would make this a
// two-mana Divination.
//
// The 3 life is the first non-mana component on a flashback cost,
// and it needs no new machinery. AlternativeCost.Life (S28, Force of
// Will and Snuff Out) is checked before anything is paid — a caster
// at 2 life cannot claim the offer at all — and paid with the spell
// already on the stack (CR 601.2h), so it behaves the same whether
// the cost is claimed from hand or from the graveyard. The
// constructor supplies the zone binding and the exile replacement;
// this file only adds the life and the printed label.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "579cbd92-797f-4cdf-91ed-fca7a523eae5",
		Name:             "Deep Analysis",
		Completeness:     CompletenessFull,
		Targets:          TargetPlayer("target player"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{deepAnalysisFlashback()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return DrawCards{Player: item.Targets[0].ID, N: 2}.Apply(ctx)
		},
	})
}

// deepAnalysisFlashback is Flashback("{1}{U}") plus the printed "Pay
// 3 life". Built from the constructor rather than by hand so the
// graveyard binding and the exile-on-leaving-the-stack replacement
// cannot be forgotten.
func deepAnalysisFlashback() game.AlternativeCost {
	fb := Flashback("{1}{U}")
	fb.Label = "Flashback—{1}{U}, Pay 3 life"
	fb.Life = 3
	return fb
}
