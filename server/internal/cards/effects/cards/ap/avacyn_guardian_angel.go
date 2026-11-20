package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avacyn, Guardian Angel — Legendary Creature — Angel {2}{W}{W}{W}, 5/4:
//
//	"Flying, vigilance
//	 {1}{W}: Prevent all damage that would be dealt to another target
//	 creature this turn by sources of the color of your choice.
//	 {5}{W}{W}: Prevent all damage that would be dealt to target player or
//	 planeswalker this turn by sources of the color of your choice."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the colour is chosen as the
// ability resolves (ColorForProtection), and the shield is the
// not-one-use one pinned to the target with that colour as its property:
// every source of that colour, rechecked as it would deal the damage
// (CR 615.9), so a source that stops being that colour gets through.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6da0b6d2-3c5e-49ef-9264-076269c04744",
		Name:            "Avacyn, Guardian Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{W}: Prevent all damage that would be dealt to another target creature this turn by sources of the color of your choice.",
				Cost:    effects.ManaCost("{1}{W}"),
				Targets: effects.Another(effects.TargetCreature("another target creature")),
				Effect:  shieldTargetFromChosenColor,
			},
			{
				Label:   "{5}{W}{W}: Prevent all damage that would be dealt to target player or planeswalker this turn by sources of the color of your choice.",
				Cost:    effects.ManaCost("{5}{W}{W}"),
				Targets: effects.TargetPlayerOrPlaneswalker(),
				Effect:  shieldTargetFromChosenColor,
			},
		},
	})
}

// shieldTargetFromChosenColor is "Prevent all damage that would be dealt
// to <target> this turn by sources of the color of your choice": the
// target, then the colour, then the shield.
func shieldTargetFromChosenColor(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	target := ts[0].ID
	effects.ChooseColorTail(game.ColorForProtection, g, item.Controller, item.SourceCardID,
		effects.ShieldSourceName(ctx)+" — choose a color of sources to prevent damage from",
		effects.WithCarry(ctx.ItemTail(effects.PreventDamageFromChosenColor), target))
	return nil
}
