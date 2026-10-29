package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Black Widow, Super Spy — Legendary Creature — Human Spy Hero {1}{B},
// 2/1 (Marvel Super Heroes, #4455):
//
//	"Menace
//	 Whenever Black Widow deals combat damage to a player, that player
//	 exiles cards from the top of their library until they exile a
//	 nonland card. You may put a +1/+1 counter on Black Widow. If you
//	 don't, you may cast the exiled nonland card until end of turn and
//	 mana of any type can be spent to cast that spell."
//
// Nicol Bolas, God-Pharaoh's +2 (the damaged player's library, exiled until
// a nonland card, a keyed tail once the run has settled) with a choice in
// place of the free cast. The choice is a keyed yes/no (MayYesNo): yes puts
// a +1/+1 counter on Black Widow if she is still the same permanent; no
// grants the exiled card a cast-only permission for the rest of the turn at
// its printed cost, with AnyType so mana of any type pays any part of it
// (CR 106.1b). The permission is for casting, so a card that is a land (a
// library of only lands) offers nothing. The card keeps its own timing, as
// "cast it this turn" does not make a sorcery an instant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "579f9a86-a1af-4c2a-a7e5-8fea0bd0eb61",
		Name:         "Black Widow, Super Spy",
		Completeness: CompletenessFull,

		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Black Widow — that player exiles cards from the top of their library until they exile a nonland card", mshA3BlackWidowDamage),
		},
	})
}

// mshA3BlackWidowDamage exiles from the damaged player's library until a
// nonland card, then asks the counter question once the run has settled.
func mshA3BlackWidowDamage(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	return MillToZone{
		Player:   item.Trigger.Event.Target,
		To:       game.ZoneExile,
		Until:    UntilNonland,
		ThenTail: ctx.ItemTail(mshA3BlackWidowExiled),
	}.Apply(ctx)
}

// mshA3BlackWidowExiled is the clause after the exile: "you may put a +1/+1
// counter on Black Widow. If you don't, you may cast the exiled nonland
// card".
var mshA3BlackWidowExiled = CardTail("msh-a3/black-widow-after-exile", func(ctx *Context, o game.TailOutcome) error {
	var hits []uuid.UUID
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsLand() {
			hits = append(hits, id)
		}
	}
	return MayChoice{
		Question: "Black Widow — put a +1/+1 counter on her? If you don't, you may cast the exiled card this turn.",
		YesLabel: "Put a +1/+1 counter on Black Widow",
		NoLabel:  "Don't",
		Then:     mshA3BlackWidowAnswer,
		Carry:    hits,
	}.Apply(ctx)
})

// mshA3BlackWidowAnswer is the choice: yes is the counter, no is the cast.
var mshA3BlackWidowAnswer = MayYesNo("confirm/black-widow-counter-or-cast",
	func(ctx *Context, _ game.ConfirmAnswer) error {
		return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	},
	func(ctx *Context, r game.ConfirmAnswer) error {
		var cards []game.Card
		for _, id := range r.Carry {
			z := ctx.Game.FindCardZoneForEffect(id)
			c, ok := ctx.Game.LookupCardForEffect(id)
			if ok && z != nil && z.Kind == game.ZoneExile {
				cards = append(cards, c)
			}
		}
		if len(cards) == 0 {
			return nil
		}
		ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:     r.Chooser,
			Zone:       game.ZoneExile,
			Duration:   ctx.Game.UntilEndOfTurnDuration(),
			CastOnly:   true,
			AnyType:    true,
			Source:     ctx.Source(),
			SourceName: "Black Widow, Super Spy",
			Label:      "Black Widow, Super Spy — cast it, spending mana of any type",
		}, cards)
		return nil
	})
