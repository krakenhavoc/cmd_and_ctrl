package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhystic Deluge — Enchantment {2}{U} (Prophecy, #5617):
//
//	"{U}: Tap target creature unless its controller pays {1}."
//
// Erratic Portal's shape with a tap in place of the bounce: the creature's
// controller is asked as the ability resolves, and the table waits for the
// answer (#5793, CR 118.12). A decline taps it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d59cfff9-2bce-439a-9456-ae68eabbf67d",
		Name:         "Rhystic Deluge",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Tap target creature unless its controller pays {1}.",
			Cost:    effects.ManaCost("{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  rhysticDelugeAsk,
		}},
	})
}

func rhysticDelugeAsk(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		return effects.PayUnless{
			Chooser:  c.Controller,
			Cost:     "{1}",
			Question: "Rhystic Deluge — pay {1}, or " + c.Name + " becomes tapped?",
			Else:     rhysticDelugeTap,
		}.Apply(ctx)
	}
	return nil
}

// rhysticDelugeTap is the decline. Keyed, an on-disk identity.
var rhysticDelugeTap = effects.PayUnlessElse("pay-unless/rhystic-deluge-tap",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.Targets() {
			if t.Kind == game.TargetCard {
				return effects.TapTarget{Target: t.ID}.Apply(ctx)
			}
		}
		return nil
	})
