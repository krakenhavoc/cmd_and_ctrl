package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Judith, Carnage Connoisseur — Legendary Creature — Human Shaman
// {3}{B}{R}, 3/4 (Murders at Karlov Manor, #4469):
//
//	"Whenever you cast an instant or sorcery spell, choose one —
//	 • That spell gains deathtouch and lifelink.
//	 • Create a 2/2 red Imp creature token with 'When this creature dies,
//	 it deals 2 damage to each opponent.'"
//
// A modal cast trigger (Kykar's shape). The first mode gives the spell
// the triggering EventCast named deathtouch and lifelink as long as it
// is that object on the stack (ThatSpellGains, CR 613.1f); the trigger
// resolves before the spell, and still does if the spell is countered.
// A spell that deals damage itself then deals it with deathtouch and
// lifelink (CR 702.2b, 702.15b); one that has another object deal the
// damage doesn't. The Imp is ImpDamageToken.
//
// No simplification.
func init() {
	t := effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()), "Judith, Carnage Connoisseur — you cast an instant or sorcery spell: choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("That spell gains deathtouch and lifelink.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.ThatSpellGains{Keywords: []string{"deathtouch", "lifelink"}}.Apply(ctx)
			}),
		effects.ModeDoing("Create a 2/2 red Imp creature token with \"When this creature dies, it deals 2 damage to each opponent.\"", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.ImpDamageToken(), N: 1}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "e86c965b-ac10-4fff-b682-dddd6d9747c6",
		Name:         "Judith, Carnage Connoisseur",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
