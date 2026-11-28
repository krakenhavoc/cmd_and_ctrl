package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b809_helpers.go — Mirage life gain and loss (#5633).

// S82OldB809OpponentsLandOfTypeBecameTapped is "a land of the chosen type an
// opponent controls becomes tapped" for one basic land type (Roots of Life):
// a tap event, or an attack that tapped the land, on a land with that subtype
// that a player other than the source's controller controls.
func S82OldB809OpponentsLandOfTypeBecameTapped(subtype string) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventTapCard && ev.Kind != game.EventAttack {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || c.Controller == source.Controller || !c.IsLand() || !c.HasSubtypeWord(subtype) {
			return false
		}
		return ev.Kind == game.EventTapCard || c.Tapped
	}
}

// S82OldB809ManaOrLifePayment is the either payment "pay <mana> or N life"
// for MayPay.Pay (Emberwilde Djinn): one prompt whose answer names the
// branch. effects.EitherPayment refuses a mana branch, so the two are
// assembled here from the engine's own kinds (see B24DiscardACardOrPay). A
// branch the chooser can't pay is not offered (CR 118.3).
func S82OldB809ManaOrLifePayment(mana string, life int) UpkeepPayment {
	pay := game.PayAction{Kind: game.PayActionMana, Mana: mana, Words: "pay " + mana}
	lose := game.PayAction{Kind: game.PayActionLife, Count: life, Words: game.PayActionWords(game.PayAction{Kind: game.PayActionLife, Count: life})}
	return UpkeepPayment{
		Action: &game.PayAction{Kind: game.PayActionEither, Either: []game.PayAction{pay, lose}},
		Words:  "Pay " + mana + " or " + lose.Words,
	}
}

// s82OldB809GainOneLife is the paid branch of "you may pay {1}. If you do,
// you gain 1 life" (Auspicious Ancestor). An on-disk key: never renamed,
// never reused.
var s82OldB809GainOneLife = PayUnlessBranches("pay-unless/old-b809-gain-1-life",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
	}, nil)

// S82OldB809MayPayOneToGainOneLife is the effect of Auspicious Ancestor's cast
// trigger: its controller may pay {1}; if they do, they gain 1 life.
func S82OldB809MayPayOneToGainOneLife(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return MayPay{
			Chooser:  item.Controller,
			Cost:     "{1}",
			Question: label + " — pay {1} to gain 1 life?",
			Then:     s82OldB809GainOneLife,
		}.Apply(NewContext(g, item))
	}
}

// S82OldB809AnyPlayerCastsAWhiteSpell is "whenever a player casts a white
// spell": any caster, the colour read as the spell is cast (CR 612.2: a word
// of the source's text, so a changed colour word is honoured).
func S82OldB809AnyPlayerCastsAWhiteSpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	return game.ThroughTextOf(source, spell).HasColorWord("W")
}

// S82OldB809Sirocco is Sirocco's resolution: the target player reveals their
// hand, then for each blue instant card revealed this way is asked whether to
// pay 4 life; one who doesn't (or can't, CR 119.4) discards that card.
func S82OldB809Sirocco(_ *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	p := ctx.Game.PlayerByIDForEffect(t.ID)
	if p == nil || p.Eliminated {
		return nil
	}
	ctx.Game.RevealHandForEffect(t.ID)
	// "Blue" is a word of Sirocco's text (CR 612.2).
	reader, _ := ctx.Game.LookupCardForEffect(ctx.Source())
	var cards []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.IsInstant() && game.ThroughTextOf(&reader, c).HasColorWord("U") {
			cards = append(cards, c.InstanceID)
		}
	}
	for _, id := range cards {
		name := "that card"
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			name = c.Name
		}
		if err := (B23PayLifeUnless{
			Chooser:  t.ID,
			Life:     4,
			Question: "Sirocco — pay 4 life? If you don't, you discard " + name + ".",
			Else:     s82OldB809SiroccoDiscard,
			Carry:    []uuid.UUID{id},
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// s82OldB809SiroccoDiscard is the decline: the revealed card is discarded if
// it is still in its owner's hand. An on-disk key: never renamed, never
// reused.
var s82OldB809SiroccoDiscard = PayUnlessBranches("pay-unless/old-b809-sirocco-discard", nil,
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		id := r.Carry[0]
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneHand {
			return nil
		}
		return ctx.Game.DiscardCardsKeyedForEffect(r.Chooser, []uuid.UUID{id}, ctx.Source(), game.KeyedTail{})
	})

// ExileTheDeadLinked is "exile that card", said of the card a death trigger
// names: it goes to exile linked to the source (CR 607.2a), only while it is
// still the object that died in its graveyard (CR 400.7). Koh, the Face
// Stealer and Purgatory share it.
func ExileTheDeadLinked(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	dead := item.Trigger.Event.CardID
	if !DiedCardStillInGraveyard(ctx, dead) {
		return nil
	}
	return ExileTarget{Target: dead, Linked: true}.Apply(ctx)
}
