package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b136_helpers.go — Theros Beyond Death batch 136 (S82 OLD, #5514).

// CantAttackUnlessYouControlAnother is "This creature can't attack unless you
// control another <permanent matching any of the queries>" (Warden of the Chained,
// CR 508.1c): the clause is read live at declaration over the controller's OTHER
// permanents, so the attacker itself never satisfies it (CR 109.1).
//
//	Static: []game.StaticAbility{CantAttackUnlessYouControlAnother(game.PermanentQuery{Types: []string{"creature"}}.WithPowerAtLeast(4))},
func CantAttackUnlessYouControlAnother(queries ...game.PermanentQuery) game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{ControllerMustControlAnother: queries})
}

// LoyaltyAbilityOfAnOpponentsPlaneswalker is the cost predicate of "Loyalty
// abilities of planeswalkers your opponents control cost {1} more to activate"
// (Eidolon of Obstruction): the ability being priced is a CR 606 loyalty ability
// (the cost query's Loyalty bit), and its source is controlled by an opponent of
// the modifier's controller.
func LoyaltyAbilityOfAnOpponentsPlaneswalker() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Ability != nil && q.Ability.Loyalty && q.Card.Controller != uuid.Nil && q.Card.Controller != q.Source.Controller
	}
}

// AtYourEndStepMayExileAnEnchantmentFromYourGraveyardForAZombie is Aphemia's
// ability: "At the beginning of your end step, you may exile an enchantment card
// from your graveyard. If you do, create a 2/2 black Zombie creature token." One
// keyed choose_cards over the graveyard (at most one card), then the Zombie from
// the exile's tail, made only if a card was exiled (Masked Vandal's shape).
func AphemiaEndStep(label string) game.TriggeredAbility {
	return AtYourEndStep(label, aphemiaOffer)
}

func aphemiaOffer(g *game.Game, item *game.StackItem) error {
	ids := B42CardsInGraveyardOf(g, item.Controller, func(c game.Card) bool { return c.HasCardType("enchantment") })
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Aphemia, the Cacophony — you may exile an enchantment card from your graveyard (if you do, create a 2/2 black Zombie)",
		Cards:    ids,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  aphemiaExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

var aphemiaExile = AskedPick("choose-cards/aphemia-the-cacophony-exile", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(r.Picked, ctx.ItemTail(aphemiaZombie))
})

var aphemiaZombie = CardTail("exile-then/aphemia-the-cacophony-zombie", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	return CreateToken{Controller: ctx.Controller(), Template: BlackZombieToken(), N: 1}.Apply(ctx)
})
