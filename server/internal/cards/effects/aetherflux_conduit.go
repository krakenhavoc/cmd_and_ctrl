package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aetherflux Conduit — Artifact {6}:
//
//	"Whenever you cast a spell, you get an amount of {E} equal to the
//	 amount of mana spent to cast that spell.
//	 {T}, Pay fifty {E}: Draw seven cards. You may cast any number of
//	 spells from your hand without paying their mana costs."
//
// The trigger is a cast trigger. The amount is read as the spell is cast,
// off the payment record (#761, as Abby, Merciless Soldier reads it) and
// carried on the item, so a spell that is countered in response still
// pays out. A spell cast for no mana (a free cast, a copy is not cast at
// all) triggers nothing worth a stack object, so it is not given one.
//
// The activated ability pays fifty energy as a cost (ADR 0129 §2,
// AbilityCost.Energy). It draws seven through DrawNThenForEffect and
// grants the free casts from the draw's continuation, so a draw that
// stops on a prompt (a dredge offer) grants over the hand the player
// really ends up with, the seven new cards included. "Any number" is
// grantFreeHandCastsFrom: one pass-closed, any-timing {0} grant per
// nonland card in hand (CR 608.2g), each cast an ordinary cast_spell
// taken in the same priority round, the whole window closing on the
// holder's next pass (free_hand_cast.go). Additional costs are still
// paid and X is 0, as for every free cast.
func init() {
	t := On(game.EventCast, conduitYouCastASpell,
		"Aetherflux Conduit — you get {E} equal to the mana spent to cast that spell",
		func(g *game.Game, item *game.StackItem) error {
			return GetEnergy{N: item.Params.Amount}.Apply(NewContext(g, item))
		})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, "Aetherflux Conduit — you get {E} equal to the mana spent to cast that spell")
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:     "775617cd-7264-4365-94c4-ffae4eb9ec8b",
		Name:         "Aetherflux Conduit",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"When “Charge mana costs” is off, or you choose Cast anyway, the game doesn't track what you spent, so casting a spell gives you no energy — turn on “Charge mana costs” for it to count.",
		},
		Triggered: []game.TriggeredAbility{t},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Pay fifty {E}: Draw seven cards. You may cast any number of spells from your hand without paying their mana costs.",
			Cost:    Plus(TapCost(), PayEnergy(50)),
			Purpose: game.Purpose{Draws: 7, Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return g.DrawNThenForEffect(ctx.Controller(), 7, game.DrawThen{
					Ref:    conduitAfterDraw,
					Player: ctx.Controller(),
					Source: ctx.Source(),
				})
			},
		}},
	})
}

// conduitYouCastASpell is "whenever you cast a spell" that paid some
// mana: an unrecorded or free payment is worth zero energy.
func conduitYouCastASpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller &&
		g.StackItemPaidForEffect(ev.CardID).Spent().Total() > 0
}

// conduitAfterDraw is the rest of the ability, once the seven are in
// hand: the free casts over everything castable there.
var conduitAfterDraw = game.RegisterDrawThen("aetherflux-conduit-after-draw", func(g *game.Game, d game.DrawThen) error {
	grantFreeHandCastsFrom(g, d.Player, d.Source, "Aetherflux Conduit", nil)
	return nil
})
