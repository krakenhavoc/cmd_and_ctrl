package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82BnsB51WarriorAttackAPlayer is the When of "whenever one or more
// Warriors you control attack a player": a Warrior the source's controller
// controls was declared attacking a player.
func S82BnsB51WarriorAttackAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !B16YouAttackedAPlayer(ev, source, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Warrior"))
}

// s82BnsB51BecomeCoward makes the chosen creature a Coward, with no end
// (CR 611.2a), in addition to its other types.
func S82BnsB51BecomeCoward(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok || !OnBattlefield(g, id) {
		return nil
	}
	return ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Coward")...)},
		Duration: g.PinnedTo(game.IndefiniteDuration(), id),
		Label:    "Gornog, the Red Reaper — becomes a Coward",
	}.Apply(ctx)
}

// s82BnsB51CowardsOpponentsControl counts the Coward creatures `you`'s
// opponents control.
func S82BnsB51CowardsOpponentsControl(g *game.Game, source *game.Card) int {
	you, coward := source.Controller, source.TextSubtype("Coward")
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you && c.IsCreature() && c.HasSubtype(coward) {
			n++
		}
	}
	return n
}

// s82BnsB51IsarethReturnBody is Isareth's reflexive trigger: the chosen
// creature card enters under your control with a corpse counter, and is
// exiled instead of leaving the battlefield. X rides Params.Amount, fixed
// when the trigger was made (CR 603.12).
var S82BnsB51IsarethReturnBody = game.ReflexiveBody("isareth/return-with-corpse-counter",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := g.ReturnFromGraveyardWithCountersKeyedForEffect(t.ID, ctx.Controller(), false,
				map[string]int{"corpse": 1}, ctx.Tail(s82BnsB51IsarethExileInstead)); err != nil {
				return NothingIfGone(err)
			}
			return nil
		}
		return nil
	},
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetCardInGraveyard("target creature card with mana value "+strconv.Itoa(p.Amount)+" from your graveyard",
			YouOwn(), Creature(), ManaValueLE(p.Amount), ManaValueGE(p.Amount))
	})

// s82BnsB51IsarethExileInstead is the entry's continuation, a keyed tail:
// the creature that entered is exiled instead of leaving the battlefield.
var s82BnsB51IsarethExileInstead = EnteredTail("isareth/exile-instead", func(ctx *Context, entered uuid.UUID, o game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	ExileInsteadOfLeavingBattlefield(ctx.Game, entered, o.Tail.Player,
		"Isareth the Awakener: if that creature would leave the battlefield, exile it instead")
	return nil
})

// s82BnsB51HurskaPaid is the paid half of Hurska's life trigger (keyed): the
// reflexive trigger carries the life gained, read from the triggering event
// as the payment is made.
var S82BnsB51HurskaPaid = PayUnlessBranches("pay-unless/hurska-sweet-tooth-pump",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		amount := 0
		if ctx.Item != nil && ctx.Item.Trigger != nil {
			amount = ctx.Item.Trigger.Event.Amount
		}
		return ReflexiveTrigger{
			Label:  "Hurska Sweet-Tooth — target creature gets +" + strconv.Itoa(amount) + "/+" + strconv.Itoa(amount) + " until end of turn",
			Body:   s82BnsB51HurskaPumpBody,
			Params: game.EffectParams{Amount: amount},
		}.Apply(ctx)
	}, nil)

// s82BnsB51HurskaPumpBody is the reflexive trigger: the chosen creature gets
// +X/+X until end of turn, X the life gained (Params.Amount).
var s82BnsB51HurskaPumpBody = game.ReflexiveBody("hurska-sweet-tooth/pump",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok || p.Amount <= 0 {
			return nil
		}
		return BoostUntilEOT{Target: id, Power: p.Amount, Toughness: p.Amount,
			Label: "Hurska Sweet-Tooth — +X/+X until end of turn"}.Apply(ctx)
	},
	ConstTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))

// s82BnsB51NazarFeast is "remove three feeding counters from it. If you do,
// you draw three cards and you lose 3 life": all or nothing, on whether the
// three counters really came off.
func S82BnsB51NazarFeast(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
		return nil
	}
	if c, ok := g.LookupCardForEffect(item.SourceCardID); !ok || c.Counters["feeding"] < 3 {
		return nil
	}
	removed, err := SocB9RemoveCounters(ctx, item.SourceCardID, "feeding", 3)
	if err != nil || removed < 3 {
		return err
	}
	if err := (DrawCards{N: 3}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Amount: -3}.Apply(ctx)
}

// s82BnsB51YouUseYourGraveyard is "you cast a spell from your graveyard or
// activate an ability of a card in your graveyard".
func S82BnsB51YouUseYourGraveyard(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventCast:
		return YouCastASpellFromYourGraveyard(ev, source, lki, g)
	case game.EventActivateAbility:
		return ev.Actor == source.Controller && ev.ActivatedFrom == game.ZoneGraveyard && s82BnsB51YouOwnTheActivatedCard(ev, source, g)
	}
	return false
}

// s82BnsB51YouOwnTheActivatedCard reads "a card in YOUR graveyard": the
// activated card's owner is the source's controller. The card may have left
// the graveyard as a cost (embalm), and is then found in its new zone.
func s82BnsB51YouOwnTheActivatedCard(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner == source.Controller
}

// s82BnsB51NeerdivMill is "target player mills cards equal to its power":
// Neerdiv's power now, or the power it last had on the battlefield.
func S82BnsB51NeerdivMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	p := FirstLegalTargetPlayer(ctx)
	if p == uuid.Nil {
		return nil
	}
	power := B13LastKnownPower(g, item.SourceCardID, game.Characteristic{})
	if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && OnBattlefield(g, item.SourceCardID) {
		power = max(c.CurrentPower(), 0)
	}
	return MillCards{Player: p, N: power}.Apply(ctx)
}
