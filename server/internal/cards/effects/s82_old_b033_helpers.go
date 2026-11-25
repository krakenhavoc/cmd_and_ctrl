package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b033_helpers.go — shared pieces of S82 older-sets batch 33
// (Modern Horizons 2: token makers).

// OldB033LoyaltyAbilityOnStack is "target loyalty ability of a planeswalker"
// (Verdant Command): an activated ability on the stack whose label is one of
// its source card's abilities that has a loyalty cost (CR 606.2). A triggered
// ability is never one, and a source that no card can name fails the clause.
func OldB033LoyaltyAbilityOnStack() AbilityPredicate {
	return func(g *game.Game, _ uuid.UUID, item *game.StackItem) bool {
		if item.Kind != game.StackItemActivated {
			return false
		}
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok {
			return false
		}
		abs, _ := game.ActivatedAbilitiesWithOrigins(c)
		for _, ab := range abs {
			if ab.Label == item.Label && ab.Cost.Loyalty != nil {
				return true
			}
		}
		return false
	}
}

// OldB033CastCreatureAndNoncreatureThisTurn is Fae Offering's intervening
// "if": the player has cast at least one creature spell and at least one
// noncreature spell this turn.
func OldB033CastCreatureAndNoncreatureThisTurn(g *game.Game, player uuid.UUID) bool {
	t := g.CastTallyFor(player)
	return t.Noncreature > 0 && t.Total-t.Noncreature > 0
}

// OldB033SquirrelSanctuaryPaid is Squirrel Sanctuary's paid branch: "If you
// do, return this enchantment to its owner's hand." It reads the Sanctuary as
// the object the ability came from (CR 400.7), so one that left and came back
// in response stays where it is.
var OldB033SquirrelSanctuaryPaid = PayUnlessBranches("pay-unless/old-b033-squirrel-sanctuary-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReturnThisPermanentToOwnersHand(ctx.Game, ctx.Item)
	}, nil)

// OldB033SquirrelSanctuaryAsk is the resolving trigger's body: offer {1}.
func OldB033SquirrelSanctuaryAsk(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Pay {1} to return Squirrel Sanctuary to its owner's hand?",
		Then:     OldB033SquirrelSanctuaryPaid,
	}.Apply(NewContext(g, item))
}

// OldB033OwnYourGraveyard is Magus of the Bridge's first trigger condition:
// a nontoken creature was put into the source controller's graveyard from the
// battlefield (the graveyard's owner is the card's owner, CR 404.2).
func OldB033NontokenCreatureToYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	return ok && !IsToken(dead) && dead.Owner == source.Controller
}

// OldB033CreatureToAnOpponentsGraveyard is Magus of the Bridge's second
// trigger condition: a creature was put into an opponent's graveyard from the
// battlefield. A token that goes there is put there for a moment (CR 111.7),
// so it counts as the printed text says "a creature".
func OldB033CreatureToAnOpponentsGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	return ok && dead.Owner != source.Controller
}

// OldB033YouSacrificedAToken is Burdened Aerialist's condition: you sacrificed
// a token.
func OldB033YouSacrificedAToken(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && IsToken(c)
}

// OldB033UnderworldHermitSquirrels is "create a number of 1/1 green Squirrel
// creature tokens equal to your devotion to black", counted as the trigger
// resolves.
func OldB033UnderworldHermitSquirrels(g *game.Game, item *game.StackItem) error {
	n := DevotionTo(g, item.Controller, NewContext(g, item).TextColor("B"))
	if n <= 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 green Squirrel"), N: n}.Apply(NewContext(g, item))
}

// OldB033TraprunnerFlipped is Goblin Traprunner's continuation: for each flip
// you win, a 1/1 red Goblin enters tapped and attacking.
var OldB033TraprunnerFlipped = CoinFlipThen("coin-flip/old-b033-goblin-traprunner", func(ctx *Context, r game.CoinFlipped) error {
	wins := 0
	for _, w := range r.Result.Won {
		if w {
			wins++
		}
	}
	return CreateTokensAttackingYourChoice{
		Template: TokenCard("1/1 red Goblin"),
		N:        wins,
		Tapped:   true,
		Label:    "Goblin Traprunner",
	}.Apply(ctx)
})

// OldB033TraprunnerFlip is "flip three coins" on the attack trigger.
func OldB033TraprunnerFlip(g *game.Game, item *game.StackItem) error {
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Goblin Traprunner — call the three coin flips",
		Coins:    3,
		ThenKey:  OldB033TraprunnerFlipped,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}
