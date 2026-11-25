package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vraska, Swarm's Eminence — Legendary Planeswalker — Vraska {2}{B/G}{B/G},
// loyalty 5:
//
//	"Whenever a creature you control with deathtouch deals damage to a player
//	 or planeswalker, put a +1/+1 counter on that creature.
//	 −2: Create a 1/1 black Assassin creature token with deathtouch and
//	 "Whenever this token deals damage to a planeswalker, destroy that
//	 planeswalker.""
//
// The trigger fires once per damage event, for any damage (combat or not) the
// creature deals to a player or a planeswalker, and gives the counter to the
// creature that dealt it if it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cff8b4e9-c60c-42c1-ad2e-74ae9d7f3afb",
		Name:            "Vraska, Swarm's Eminence",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, vraskaSwarmsEminenceDeathtouchHit,
				"Vraska, Swarm's Eminence — put a +1/+1 counter on that creature",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					id := ctx.Trigger().Event.Source
					if !effects.OnBattlefield(ctx.Game, id) {
						return nil
					}
					return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:  "−2: Create a 1/1 black Assassin creature token with deathtouch and \"Whenever this token deals damage to a planeswalker, destroy that planeswalker.\"",
			Cost:   effects.LoyaltyCost(-2),
			Effect: effects.Do(effects.CreateToken{Template: effects.OldB007AAssassinToken(), N: 1}),
		}},
	})
}

func vraskaSwarmsEminenceDeathtouchHit(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	if !ok || !dealer.IsCreature() || dealer.Controller != source.Controller || !game.HasKeyword(&dealer, "deathtouch") || !g.Battlefield.Contains(ev.Source) {
		return false
	}
	if g.PlayerByIDForEffect(ev.Target) != nil {
		return true
	}
	victim, ok := g.LookupCardForEffect(ev.Target)
	return ok && victim.IsPlaneswalker()
}
