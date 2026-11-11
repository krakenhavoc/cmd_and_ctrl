package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b01_helpers.go — card-side helpers for the Commander-deck wave 2,
// batch 1 (simple lands) and batch 2 (mana rocks and mana triggers)
// (#5368). Every name starts with cmdr2B01.

// cmdr2B01TapForAnyColor is "{T}: Add one mana of any color."
func cmdr2B01TapForAnyColor() ManaAbility {
	return ManaAbility{
		Cost:     ManaAbilityCost{Tap: true},
		Produced: "{W|U|B|R|G}",
		Label:    "Add one mana of any color",
	}
}

// cmdr2B01TapForCreatureOnlyMana is "{T}: Add one mana of any color.
// Spend this mana only to cast a creature spell." (Ancient Ziggurat,
// Abundant Countryside; Beastcaller Savant's shape.)
func cmdr2B01TapForCreatureOnlyMana() ManaAbility {
	a := cmdr2B01TapForAnyColor()
	a.Restrictions = []string{ManaRestrictCast, ManaRestrictType("Creature")}
	a.Label = "Add one mana of any color. Spend this mana only to cast a creature spell"
	return a
}

// cmdr2B01TapForColor is a single-colour land's "{T}: Add {<c>}".
func cmdr2B01TapForColor(c string) ManaAbility {
	return ManaAbility{
		Cost:     ManaAbilityCost{Tap: true},
		Produced: "{" + c + "}",
		Label:    "Add {" + c + "}",
	}
}

// cmdr2B01MakeToken is "create a <token>" as an activated ability's whole
// effect, the token named by its tokens_table key.
func cmdr2B01MakeToken(tokenKey string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return CreateToken{
			Controller: item.Controller,
			Template:   TokenCard(tokenKey),
			N:          1,
		}.Apply(NewContext(g, item))
	}
}

// cmdr2B01SoulCountersOn reads the soul counters off the ability's own
// source (Séance Board's X).
func cmdr2B01SoulCountersOn(g *game.Game, _, source uuid.UUID) int {
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return 0
	}
	return c.Counters["soul"]
}

// cmdr2B01BrainstonePutBack is "then put two cards from your hand on top of
// your library in any order" with Brainstone's own prompt label. Brainstorm's
// registration (drawThenBrainstormPutBack) names Brainstorm in its prompt.
var cmdr2B01BrainstonePutBack = game.RegisterDrawThen("cmdr2-b01-brainstone-put-back", func(g *game.Game, d game.DrawThen) error {
	ctx := NewContext(g, &game.StackItem{Controller: d.Player, SourceCardID: d.Source})
	return PutFromHandOnTopInAnyOrder{
		Player: d.Player,
		N:      d.N,
		Label:  "Brainstone — put two cards from your hand on top of your library",
	}.Apply(ctx)
})

// cmdr2B01CondActorHandEmpty is "if that player has no cards in hand" for a
// trigger that fired off a player's step (Ghirapur Orrery's upkeep): the
// player is the event's actor, and the hand is read as the trigger goes on
// the stack and again as it resolves (CR 603.4, InterveningIf). It cannot
// declare a footprint, so the row keeps its trigger-order prompt.
type cmdr2B01CondActorHandEmpty struct{}

func (cmdr2B01CondActorHandEmpty) Holds(ctx *Context) bool {
	if ctx.Item == nil {
		return false
	}
	p := ctx.Game.PlayerByIDForEffect(ctx.Item.Trigger.Event.Actor)
	return p != nil && p.Hand != nil && p.Hand.Size() == 0
}

func (cmdr2B01CondActorHandEmpty) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// cmdr2B01EachOpponentOfDrawsAndGains is "each of <player>'s opponents draws
// a card and gains N life", in that order for each opponent: Mathas, Fiend
// Seeker's bounty (opponents of the creature's controller) and Bounty Board's
// trigger (opponents of the dead creature's controller).
func cmdr2B01EachOpponentOfDrawsAndGains(g *game.Game, item *game.StackItem, of uuid.UUID, life int) error {
	ctx := NewContext(g, item)
	var err error
	eachOpponent(g, of, func(opp uuid.UUID) bool {
		if err = (DrawCards{Player: opp, N: 1}).Apply(ctx); err != nil {
			return true
		}
		err = GainLife{Player: opp, Amount: life}.Apply(ctx)
		return err != nil
	})
	return err
}
