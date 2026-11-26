package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heretic's Punishment — {4}{R} Enchantment (Innistrad, S82 OLD b448, #5572):
//
//	"{3}{R}: Choose any target, then mill three cards. This enchantment deals
//	 damage to that permanent or player equal to the greatest mana value among
//	 the milled cards."
//
// The target is chosen as the ability is activated (CR 602.2b), before the mill.
// The mill is the keyed mill, so the damage waits for a replacement's prompt and
// counts only cards that actually reached the graveyard (CR 400.7); the greatest
// mana value is read as the cards were in the library (CR 608.2h). A target that
// is no longer legal as the ability resolves takes no damage, but the mill still
// happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a484120f-1027-4fc0-b4ea-cba94c262203",
		Name:         "Heretic's Punishment",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}: Choose any target, then mill three cards. This enchantment deals damage to that permanent or player equal to the greatest mana value among the milled cards.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{3}{R}"),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(hereticsPunishmentMilled)}.Apply(ctx)
			},
		}},
	})
}

// hereticsPunishmentMilled is the damage, by the greatest mana value among the
// cards milled. An on-disk key: never renamed, never reused.
var hereticsPunishmentMilled = effects.CardTail("heretics-punishment/damage-by-greatest-mana-value", func(ctx *effects.Context, o game.TailOutcome) error {
	greatest := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.ManaValue() > greatest {
			greatest = c.ManaValue()
		}
	}
	if greatest <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		return effects.DealDamage{Source: ctx.Item.SourceCardID, Target: t.ID, Amount: greatest}.Apply(ctx)
	}
	return nil
})
