package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB639Targets splits a resolving "target opponent … target creature"
// ability's legal targets into the player and the card, each zero when its
// clause is no longer legal (CR 608.2b): the rest of the ability still
// happens around a clause that went illegal.
func OldB639Targets(ctx *Context) (player uuid.UUID, card uuid.UUID) {
	for _, t := range ctx.LegalTargets() {
		switch t.Kind {
		case game.TargetPlayer:
			player = t.ID
		case game.TargetCard:
			card = t.ID
		}
	}
	return player, card
}

// OldB639RevealRandomFromHand is "<player> reveals a card at random from
// their hand": the card is picked from the game's keyed random stream (so an
// undo rewinds it) and revealed to the table. ok is false for an empty hand
// (nothing is revealed) or a player who is gone.
func OldB639RevealRandomFromHand(ctx *Context, player uuid.UUID, reason string) (id uuid.UUID, ok bool) {
	p := ctx.Game.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
		return uuid.Nil, false
	}
	pick := ctx.Game.ChooseAtRandomForEffect(RandomDraw(ctx), HandCardIDs(ctx.Game, player), 1)
	if len(pick) == 0 {
		return uuid.Nil, false
	}
	ctx.Game.RevealForEffect(game.RevealSpec{Player: player, Source: ctx.Source(), Cards: pick, Reason: reason})
	return pick[0], true
}

// OldB639RevealedManaValue is X for the Planeswalker's cycle: the revealed
// card's mana value, read as it is revealed (CR 608.2h), zero when no card
// was revealed.
func OldB639RevealedManaValue(ctx *Context, id uuid.UUID, ok bool) int {
	if !ok {
		return 0
	}
	c, found := ctx.Game.LookupCardForEffect(id)
	if !found {
		return 0
	}
	return c.ManaValue()
}

// OldB639RevealThenBoostTarget is Planeswalker's Favor and Scorn: the
// opponent reveals a card at random from their hand, then the creature gets
// sign*X/sign*X until end of turn, X the revealed card's mana value. Each
// clause is judged as the ability resolves (CR 608.2b): with the opponent
// gone nothing is revealed and X is 0, with the creature gone nothing gets
// the change.
func OldB639RevealThenBoostTarget(name string, sign int) Effect {
	return Do(Instruction(func(ctx *Context) error {
		opp, creature := OldB639Targets(ctx)
		revealed, ok := uuid.Nil, false
		if opp != uuid.Nil {
			revealed, ok = OldB639RevealRandomFromHand(ctx, opp, name)
		}
		x := OldB639RevealedManaValue(ctx, revealed, ok)
		if creature == uuid.Nil || x == 0 {
			return nil
		}
		return BoostUntilEOT{Target: creature, Power: sign * x, Toughness: sign * x, Label: name + " — X is the revealed card's mana value"}.Apply(ctx)
	}))
}

// OldB639MischiefReveal is Planeswalker's Mischief's resolution: the
// opponent reveals a card at random from their hand and, if it is an instant
// or sorcery card, it is exiled and its controller may cast it without paying
// its mana cost for as long as it remains exiled (CR 608.2g).
func OldB639MischiefReveal(ctx *Context) error {
	opp, _ := OldB639Targets(ctx)
	if opp == uuid.Nil {
		return nil
	}
	id, ok := OldB639RevealRandomFromHand(ctx, opp, "Planeswalker's Mischief")
	if !ok {
		return nil
	}
	c, found := ctx.Game.LookupCardForEffect(id)
	if !found || !(c.IsInstant() || c.IsSorcery()) {
		return nil
	}
	return ctx.Game.ExileCardKeyedForEffect(id, game.KeyedTail{Ref: oldB639MischiefExiled, Player: ctx.Controller(), Source: ctx.Source(), Asked: game.NewAskedItem(ctx.Item)})
}

// oldB639MischiefExiled grants the free cast over the card that reached exile
// and schedules the end-step return. An on-disk key: never renamed.
var oldB639MischiefExiled = CardTail("old-b639/planeswalkers-mischief-exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	g, id := ctx.Game, o.IDs[0]
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
		Player:     o.Tail.Player,
		Zone:       game.ZoneExile,
		Duration:   game.WhileInZoneDuration(),
		Cost:       "{0}",
		CastOnly:   true,
		Source:     o.Tail.Source,
		SourceName: "Planeswalker's Mischief",
		Label:      "Planeswalker's Mischief — cast " + c.Name + " without paying its mana cost",
	})
	return ScheduleDelayedTrigger{
		At:     game.StepEnd,
		Label:  "Planeswalker's Mischief — if you haven't cast " + c.Name + ", return it to its owner's hand",
		Cards:  []uuid.UUID{id},
		Body:   oldB639MischiefReturnBody,
		Params: game.EffectParams{Player: o.Tail.Player, Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})

// oldB639MischiefReturnBody returns the card to its owner's hand if it is
// still the object Mischief exiled: not cast, not moved. On-disk key.
var oldB639MischiefReturnBody = game.DelayedBody("old-b639/planeswalkers-mischief-return", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	z := g.FindCardZoneForEffect(id)
	if z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	return BounceToHand{Target: id}.Apply(NewContext(g, item))
})

// OldB639ProtectionFromThatPermanentsColors is Samite Elder: the creatures you
// control gain protection from each of the chosen permanent's colors until end
// of turn. The colors are read as the ability resolves (CR 608.2h); a
// colorless permanent grants nothing, and a permanent that has left grants
// what it had (last-known information).
func OldB639ProtectionFromThatPermanentsColors(name string) Effect {
	return Do(Instruction(func(ctx *Context) error {
		_, id := OldB639Targets(ctx)
		if id == uuid.Nil {
			return nil
		}
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			return nil
		}
		var kws []string
		for _, col := range game.AllColors {
			if c.HasColorWord(col) {
				kws = append(kws, "protection from "+game.ColorName(col))
			}
		}
		if len(kws) == 0 {
			return nil
		}
		return GrantKeywordUntilEOT{
			Match:    And(Creature(), YouControl()),
			Keywords: kws,
			Label:    name + " — protection from the permanent's colors",
		}.Apply(ctx)
	}))
}

// OldB639TargetGainsFlyingAndBecomesBlue is Disciple of Kangee's effect: the
// target creature gains flying and becomes blue (a layer-5 set, CR 613.1e)
// until end of turn, both on the target if still legal as the ability resolves.
func OldB639TargetGainsFlyingAndBecomesBlue(ctx *Context) error {
	_, id := OldB639Targets(ctx)
	if id == uuid.Nil {
		return nil
	}
	if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Disciple of Kangee — flying"}).Apply(ctx); err != nil {
		return err
	}
	return Gap2B05TargetBecomesColor(ctx, []string{"U"}, "Disciple of Kangee — blue until end of turn")
}
