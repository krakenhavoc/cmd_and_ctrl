package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Unluckiest Planeswalker — Legendary Planeswalker {2}{R}, loyalty 4
// (Mystery Booster Commander, #5948):
//
//	"Rulebreaker — A deck with this commander can have Aura cards of any color
//	 identity and any basic land cards.
//	 +1: Add one mana of any color. When you spend this mana to cast an Aura
//	 spell this turn, create a 1/1 red Gremlin creature token.
//	 −3: Discard your hand, then draw cards equal to twice the number of Auras
//	 you control. For each Aura attached to you, attach it to an opponent of
//	 your choice.
//	 The Unluckiest Planeswalker can be your commander."
//
// The +1 adds one mana of any colour carrying a spend trigger (Study Hall's
// WhenManaSpent): when it pays for an Aura spell, its caster makes a 1/1 red
// Gremlin. The mana empties with the phase, which is the printed "this turn" at
// its longest. The −3 discards the whole hand, draws twice the Auras its
// controller controls as they stand after the discard, then walks every Aura
// attached to its controller (whoever controls it) and asks which opponent each
// goes to, one prompt per Aura (a ChoosePlayer chain). The move is the engine's
// attach (CR 303.4), which leaves an Aura where it is when its enchant clause
// can't take the new host (CR 701.3b). Rulebreaker is deck construction, read from
// the oracle text by the deck validator (internal/deck); "can be your commander"
// likewise.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "c7ea2ce7-f684-44a2-94b2-d4aa7353f6fc",
		Name:           "The Unluckiest Planeswalker",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Activated: []ActivatedAbility{
			{
				Label:  "+1: Add one mana of any color. When you spend this mana to cast an Aura spell this turn, create a 1/1 red Gremlin creature token.",
				Cost:   LoyaltyCost(1),
				Effect: unluckiestPlusOne,
			},
			{
				Label:   "−3: Discard your hand, then draw cards equal to twice the number of Auras you control. For each Aura attached to you, attach it to an opponent of your choice.",
				Cost:    LoyaltyCost(-3),
				Purpose: game.Purpose{Draws: 2},
				Effect:  unluckiestMinusThree,
			},
		},
	})
}

var unluckiestGremlinRider = WhenManaSpent("The Unluckiest Planeswalker", game.ManaSpendTrigger{
	Label: "The Unluckiest Planeswalker — create a 1/1 red Gremlin creature token",
	Condition: func(_ *game.Game, _ uuid.UUID, paidFor game.Card) bool {
		return paidFor.IsAura()
	},
	Effect: func(g *game.Game, item *game.StackItem) error {
		return CreateToken{Template: TokenCard("1/1 red Gremlin"), N: 1}.Apply(NewContext(g, item))
	},
}, game.ManaRestrictCast)

func unluckiestPlusOne(g *game.Game, item *game.StackItem) error {
	return AddMana{
		Player:   item.Controller,
		Produced: "{W|U|B|R|G}",
		Riders:   []game.ManaSpendRider{unluckiestGremlinRider},
	}.Apply(NewContext(g, item))
}

func unluckiestMinusThree(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if _, err := discardWholeHand(g, item.Controller); err != nil {
		return err
	}
	auras, onMe := 0, []uuid.UUID(nil)
	for _, c := range g.BattlefieldCardsForEffect() {
		if !c.IsAura() {
			continue
		}
		if c.Controller == item.Controller {
			auras++
		}
		if c.AttachedTo.Kind == game.TargetPlayer && c.AttachedTo.ID == item.Controller {
			onMe = append(onMe, c.InstanceID)
		}
	}
	if err := (DrawCards{N: 2 * auras}).Apply(ctx); err != nil {
		return err
	}
	return unluckiestAttachNext(ctx, onMe)
}

// unluckiestAttachNext asks which opponent the next Aura goes to, skipping an
// Aura that is no longer attached to the controller; with none left it is done.
func unluckiestAttachNext(ctx *Context, auras []uuid.UUID) error {
	for len(auras) > 0 {
		c, ok := ctx.Game.LookupCardForEffect(auras[0])
		if ok && c.IsAura() && c.AttachedTo.Kind == game.TargetPlayer && c.AttachedTo.ID == ctx.Controller() {
			return ChoosePlayer{
				Among:    OpponentsOf(ctx.Controller()),
				Question: "The Unluckiest Planeswalker — attach " + c.Name + " to an opponent of your choice",
				ThenKey:  unluckiestAttachPicked,
				Carry:    auras,
			}.Apply(ctx)
		}
		auras = auras[1:]
	}
	return nil
}

// unluckiestAttachPicked attaches the Aura the question named to the chosen
// opponent, then asks about the next. An on-disk key.
var unluckiestAttachPicked game.OptionPickThen

func init() {
	unluckiestAttachPicked = ChoosePlayerThen("option-pick/the-unluckiest-planeswalker-attach", unluckiestAttachBody)
}

func unluckiestAttachBody(ctx *Context, chosen uuid.UUID, carry []uuid.UUID) error {
	if len(carry) == 0 {
		return nil
	}
	aura, rest := carry[0], carry[1:]
	if chosen != uuid.Nil {
		if c, ok := ctx.Game.LookupCardForEffect(aura); ok && c.IsAura() &&
			c.AttachedTo.Kind == game.TargetPlayer && c.AttachedTo.ID == ctx.Controller() {
			if err := ctx.Game.AttachForEffect(aura, game.TargetRef{Kind: game.TargetPlayer, ID: chosen}); err != nil {
				return err
			}
		}
	}
	return unluckiestAttachNext(ctx, rest)
}
