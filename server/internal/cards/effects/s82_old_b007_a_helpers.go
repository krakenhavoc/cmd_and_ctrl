package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB007AThisDealtDamageToAPlaneswalker is "whenever this [creature] deals
// damage to a planeswalker": the source dealt more than 0 damage, of any kind,
// to a planeswalker that is still on the battlefield.
func OldB007AThisDealtDamageToAPlaneswalker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source != source.InstanceID {
		return false
	}
	victim, ok := g.LookupCardForEffect(ev.Target)
	return ok && victim.IsPlaneswalker() && g.Battlefield.Contains(ev.Target)
}

// printedOldB007AAssassinToken is Vraska, Swarm's Eminence's 1/1 black Assassin
// with deathtouch and "Whenever this token deals damage to a planeswalker,
// destroy that planeswalker." The slug is an on-disk identity: never renamed,
// never reused.
func printedOldB007AAssassinToken() TokenTemplate {
	return TokenTemplate{
		Slug: "assassin-deathtouch-destroys-planeswalker",
		Card: game.Card{
			Name:     "Assassin",
			TypeLine: "Token Creature — Assassin",
			Power:    1, Toughness: 1,
			Colors:   []string{"B"},
			Keywords: []string{"deathtouch"},
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, OldB007AThisDealtDamageToAPlaneswalker,
				"Assassin — destroy that planeswalker",
				Do(Instruction(func(ctx *Context) error {
					victim := ctx.Trigger().Event.Target
					if !OnBattlefield(ctx.Game, victim) {
						return nil
					}
					return DestroyTarget{Target: victim}.Apply(ctx)
				}))),
		},
		Text: "Deathtouch\nWhenever this token deals damage to a planeswalker, destroy that planeswalker.",
	}
}

// OldB007AAssassinToken is the Assassin's card value for CreateToken.
func OldB007AAssassinToken() game.Card { return TokenFromCatalog(printedOldB007AAssassinToken) }
