package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necromancer's Covenant — Enchantment {3}{W}{B}{B} (Commander 2015, S82 OLD
// #5546):
//
//	"When this enchantment enters, exile all creature cards from target
//	 player's graveyard, then create a 2/2 black Zombie creature token for each
//	 card exiled this way.
//	 Zombies you control have lifelink."
//
// Crypt Incursion's exile of the creature cards in one graveyard: the cards
// that actually reach exile (a commander may take the command zone instead,
// CR 903.9) are what the tokens count, one 2/2 black Zombie for each, made
// under your control. The lifelink is a static grant to every Zombie you
// control, tokens and the Covenant's own Zombies included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9103f01f-e60b-4cda-9102-edc4aef74cc3",
		Name:         "Necromancer's Covenant",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Zombie"}, YoursOnly: true}, "lifelink"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Necromancer's Covenant — exile all creature cards from target player's graveyard, then create a 2/2 black Zombie for each card exiled this way",
					necromancersCovenantExile),
				effects.TargetPlayer("target player")),
		},
	})
}

func necromancersCovenantExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player, ok := effects.FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	ids := effects.B42CardsInGraveyardOf(g, player, func(c game.Card) bool { return c.IsCreature() })
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, ctx.Tail(necromancersCovenantZombies))
}

// necromancersCovenantZombies is "create a 2/2 black Zombie creature token
// for each card exiled this way". An on-disk key: never renamed, never reused.
var necromancersCovenantZombies = effects.CardTail("then/necromancers-covenant-zombies", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: len(o.IDs)}.Apply(ctx)
})
