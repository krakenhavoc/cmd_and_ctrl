package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Rani — Legendary Creature — Time Lord Scientist {1}{U}{B}{R}, 3/4:
//
//	"Whenever The Rani enters or attacks, create a red Aura enchantment
//	 token named Mark of the Rani attached to another target creature.
//	 That token has enchant creature and "Enchanted creature gets +2/+2
//	 and is goaded."
//	 Whenever a goaded creature deals combat damage to one of your
//	 opponents, investigate."
//
// The Mark is the printed token (MarkOfTheRaniToken), created attached to
// the target. If the target is gone or no longer a legal target as the
// ability resolves, the ability does nothing (CR 608.2b). The Mark's goad
// is static, so the creature stays goaded by the Mark's controller for as
// long as the Mark is attached to it (CR 701.15).
//
// The second ability counts any goaded creature, whoever controls it and
// whoever goaded it, dealing combat damage to a player who is an opponent
// of The Rani's controller: one trigger per creature (CR 510.2).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac9b0bdf-caf8-4dba-a900-e59de35e77d7",
		Name:         "The Rani",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEntersOrAttacks("The Rani — create Mark of the Rani attached to another target creature", theRaniMark),
				effects.Another(effects.TargetCreature("another target creature"))),
			effects.On(game.EventDealDamage, theRaniGoadedCreatureHitAnOpponent, "The Rani — investigate", effects.Do(effects.Investigate{N: 1})),
		},
	})
}

func theRaniMark(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	host, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	_, err := g.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(effects.MarkOfTheRaniToken()), 1,
		game.TokenEntryOptions{AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}})
	return err
}

// theRaniGoadedCreatureHitAnOpponent is "a goaded creature deals combat
// damage to one of your opponents", judged as the damage is dealt.
func theRaniGoadedCreatureHitAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	p := g.PlayerByIDForEffect(ev.Target)
	if p == nil || p.ID == source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && c.IsCreature() && c.Goaded()
}
