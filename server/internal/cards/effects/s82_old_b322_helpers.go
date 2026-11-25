package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b322_helpers.go — helpers for Magic Origins, batch 322 (S82 older
// sets, #5549).

// OldB322AlhammarretAsEnters is "As Alhammarret enters, each opponent reveals
// their hand. You choose the name of a nonland card revealed this way." Every
// opponent's whole hand is revealed to the table (CR 701.20a), then the
// controller picks one of the distinct names among the revealed nonland cards
// (an option pick over the names, so a name that is not among them can't be
// chosen). The name is stored on the permanent as it enters (CR 614.12). With
// no nonland card revealed nothing is chosen and the restriction bars nothing.
func OldB322AlhammarretAsEnters(card *game.Card, ctx *Context) error {
	g := ctx.Game
	var options []game.ChoiceOption
	seen := map[string]bool{}
	for _, seat := range g.Seats {
		if seat == nil || seat.ID == card.Controller || seat.Eliminated || seat.Hand == nil {
			continue
		}
		g.RevealHandForEffect(seat.ID)
		for _, c := range seat.Hand.Cards {
			key := strings.ToLower(strings.TrimSpace(c.Name))
			if c.IsLand() || key == "" || seen[key] {
				continue
			}
			seen[key] = true
			options = append(options, game.ChoiceOption{Label: c.Name, Cards: []uuid.UUID{c.InstanceID}})
		}
	}
	if len(options) == 0 {
		return nil
	}
	g.QueueOptionPickForEffect(game.OptionPickPrompt{
		Chooser:  card.Controller,
		Source:   card.InstanceID,
		Question: "Alhammarret, High Arbiter — choose the name of a nonland card revealed this way",
		Options:  options,
		ThenKey:  oldB322AlhammarretNamed,
		Carry:    []uuid.UUID{card.InstanceID},
	})
	return nil
}

// oldB322AlhammarretNamed stores the chosen name. An on-disk key: never
// renamed, never reused.
var oldB322AlhammarretNamed = game.RegisterOptionPickThen("option-pick/s82-old-b322-alhammarret-name", func(g *game.Game, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 {
		return nil
	}
	g.SetChosenNameForEffect(r.Carry[0], r.Option.Label)
	return nil
})

// OldB322OpponentsCantCastTheChosenName is "Your opponents can't cast spells
// with the chosen name (as long as this creature is on the battlefield)": a
// cast restriction of the permanent, so it lifts the moment it leaves, matched
// on any face (CR 201.2b) against the name stored on it.
func OldB322OpponentsCantCastTheChosenName() game.CastRestriction {
	return game.CastRestriction{
		Label: "Your opponents can't cast spells with the chosen name.",
		Forbids: func(q game.CastQuery) bool {
			if q.Controller == q.Source.Controller {
				return false
			}
			named := strings.TrimSpace(q.Source.ChosenName)
			return named != "" && game.CardNameMatches(q.Card, named)
		},
	}
}

// OldB322RevealTopLandElseHand is "Reveal the top card of your library. If it's
// a land card, put it onto the battlefield. Otherwise, put it into your hand"
// (Coiling Oracle's reveal, Nissa, Sage Animist's +1). A revealed token is
// neither put nor moved (CR 111.8). An empty library does nothing.
func OldB322RevealTopLandElseHand(g *game.Game, item *game.StackItem) error {
	top, err := RevealTopThenPutIfMatch(g, item.SourceCardID, item.Controller, game.Card.IsLand, false,
		"Nissa, Sage Animist — revealed from the top of the library")
	if err != nil || top == uuid.Nil {
		return err
	}
	if c, ok := g.LookupCardForEffect(top); ok && (c.IsLand() || IsToken(c)) {
		return nil
	}
	return g.BounceToHandForEffect(top)
}

// OldB322JaceLoot is the {T} ability's first two sentences, "Draw a card, then
// discard a card", with the rest of the ability waiting on the discard. The
// draw hands the discard to a registered body (a draw can pause on a prompt),
// carrying the source object's epoch in N so the flip still knows whether Jace
// is the permanent that activated it (CR 400.7).
func OldB322JaceLoot(g *game.Game, item *game.StackItem) error {
	return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
		Ref: oldB322JaceDiscard, Player: item.Controller, Source: item.SourceCardID, N: item.SourceObject.Epoch,
	})
}

var oldB322JaceDiscard = game.RegisterDrawThen("s82-old-b322/jace-discard", func(g *game.Game, d game.DrawThen) error {
	asked := &game.StackItem{
		Kind: game.StackItemTriggered, Controller: d.Player, SourceCardID: d.Source,
		SourceObject: game.ObjectRef{ID: d.Source, Epoch: d.N},
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   d.Player,
		Source:   d.Source,
		N:        1,
		Question: "Jace, Vryn's Prodigy — discard a card",
		Asked:    game.NewAskedItem(asked),
	}, oldB322JaceFlip, nil)
})

// oldB322JaceFlip is the last sentence: "If there are five or more cards in
// your graveyard, exile Jace, then return him to the battlefield transformed
// under his owner's control." The count is taken once the discard is done, and
// a Jace that is no longer the permanent that activated it is not exiled.
// An on-disk key: never renamed, never reused.
var oldB322JaceFlip = game.RegisterDiscardThen("choose-cards/s82-old-b322-jace-flip", func(g *game.Game, r game.DiscardsDone) error {
	p := g.PlayerByIDForEffect(r.Player)
	if p == nil || p.Graveyard == nil || len(p.Graveyard.Cards) < 5 {
		return nil
	}
	if r.Item != nil && g.AbilitySourceIsNewObjectForEffect(r.Item) {
		return nil
	}
	if z := g.FindCardZoneForEffect(r.Source); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return g.ExileAndReturnTransformedForEffect(r.Source, uuid.Nil)
})

// OldB322ChandraTap is the front face's {T} ability: 1 damage to the target
// player or planeswalker, then "If Chandra has dealt 3 or more damage this
// turn, exile her, then return her to the battlefield transformed under her
// owner's control." The count is every damage the object has dealt this turn, to
// players and permanents, read after this damage (CR 120.3, 400.7).
func OldB322ChandraTap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if t, ok := FirstLegalTarget(ctx); ok {
		if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	ref, ok := oldB322LiveRef(g, item.SourceCardID)
	if !ok || g.DamageDealtThisTurnByObjectForEffect(ref) < 3 {
		return nil
	}
	return ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(ctx)
}

// OldB322ChandraUltimate is the −7: 6 damage to each opponent, then each player
// dealt damage this way gets an emblem. Chandra may already be gone (a walker
// with seven loyalty has none left once the cost is paid), so who was dealt
// damage is read off each player's damage taken this turn before and after,
// never off Chandra; a player the damage was prevented to gets no emblem.
func OldB322ChandraUltimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opponents := ctx.Opponents()
	before := make(map[uuid.UUID]int, len(opponents))
	for _, opp := range opponents {
		before[opp] = g.DamageTakenThisTurn(opp)
	}
	if err := DamageToEachOpponent(g, item, 6); err != nil {
		return err
	}
	for _, opp := range opponents {
		if g.DamageTakenThisTurn(opp) <= before[opp] {
			continue
		}
		if err := (CreateEmblem{Player: opp}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// OldB322MayCastFromGraveyard is "You may cast target instant or sorcery card
// from your graveyard this turn. If that spell would be put into your graveyard,
// exile it instead" (Vohar, Vodalian Desecrator's grant): a one-card permission
// for the rest of the turn, at the card's mana cost and normal timing. A target
// that left the graveyard in response is granted nothing (CR 608.2b).
func OldB322MayCastFromGraveyard(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		z := g.FindCardZoneForEffect(id)
		c, found := g.LookupCardForEffect(id)
		if !found || z == nil || z.Kind != game.ZoneGraveyard || c.Owner != item.Controller {
			return nil
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:                  item.Controller,
			Zone:                    game.ZoneGraveyard,
			Duration:                g.UntilEndOfTurnDuration(),
			AltCostKey:              game.AltCostKeyGrantedCost,
			CastOnly:                true,
			CastsLeft:               1,
			ExileInsteadOfGraveyard: true,
			Source:                  item.SourceCardID,
			SourceName:              name,
			Label:                   name + " — cast " + c.Name + " from your graveyard",
		}, []game.Card{c})
		return nil
	}
}

// oldB322LiveRef is the object a permanent is now, as the damage tally names a
// source: its instance ID and the ObjectEpoch it has on the battlefield.
func oldB322LiveRef(g *game.Game, id uuid.UUID) (game.ObjectRef, bool) {
	c, ok := g.LookupCardForEffect(id)
	if !ok || !OnBattlefield(g, id) {
		return game.ObjectRef{}, false
	}
	return game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}, true
}

// OldB322CreatureAnOpponentControlsBecameYourTarget is "whenever a creature an
// opponent controls becomes the target of a spell or ability you control" (Shay
// Cormac, Willbreaker): one event for each target slot (CR 115.3), naming a
// creature on the battlefield that an opponent of the source's controller
// controls, targeted by a spell or ability the source's controller controls.
func OldB322CreatureAnOpponentControlsBecameYourTarget(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || ev.CardID == uuid.Nil || ev.Actor != source.Controller || !OnBattlefield(g, ev.CardID) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller != source.Controller
}
