package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anim Pakal, Thousandth Moon — Legendary Creature — Human Soldier
// {1}{R}{W}, 1/2 (The Lost Caverns of Ixalan, #4470):
//
//	"Whenever you attack with one or more non-Gnome creatures, put a +1/+1
//	 counter on Anim Pakal, then create X 1/1 colorless Gnome artifact
//	 creature tokens that are tapped and attacking, where X is the number of
//	 +1/+1 counters on Anim Pakal."
//
// "One or more" is one trigger per attack declaration (OncePerBatch), and
// it needs a creature in the declaration that is not a Gnome, read through
// the layered subtypes. The counter goes on first, X is counted after it,
// as the trigger resolves (so a Pakal that has left makes X from its last
// known counters, CR 608.2h, with no counter placed), and the Gnomes enter
// tapped and attacking the same player or planeswalker the triggering
// attacker is attacking (CR 508.4): they were never declared as
// attackers, so they do not trigger this again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05551d91-50c6-46d1-86a4-cd3d177d0923",
		Name:         "Anim Pakal, Thousandth Moon",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !c.HasSubtypeWord("Gnome")
			}, "Anim Pakal — put a +1/+1 counter on it, then create X tapped and attacking Gnome artifact creature tokens",
				animPakalAttack)),
		},
	})
}

func animPakalAttack(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	x := effects.CounterCountOn(g, ctx.Source(), game.CounterPlusOne)
	if x <= 0 || item.Trigger == nil {
		return nil
	}
	tmpl := effects.TokenCard("1/1 colorless Gnome artifact")
	tmpl.Tapped = true
	if attacker, ok := g.LookupCardForEffect(item.Trigger.Event.CardID); ok {
		tmpl.AttackingTarget = attacker.AttackingTarget
	}
	return effects.CreateToken{Template: tmpl, N: x}.Apply(ctx)
}
