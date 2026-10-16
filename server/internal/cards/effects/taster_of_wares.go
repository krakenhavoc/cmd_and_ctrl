package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Taster of Wares — Creature — Goblin Warlock {2}{B}, 3/2:
//
//	"When this creature enters, target opponent reveals X cards from
//	 their hand, where X is the number of Goblins you control. You choose
//	 one of those cards. That player exiles it. If an instant or sorcery
//	 card is exiled this way, you may cast it for as long as you control
//	 this creature, and mana of any type can be spent to cast that
//	 spell."
//
// X is counted as the trigger resolves (the Goblins its controller
// controls then, the Taster itself among them if it is still there).
// The OPPONENT chooses which X cards to reveal (a choose_cards prompt
// over their own hand; with X at least their hand size they reveal it
// all and nobody is asked), and only those are revealed, to every
// player (CR 701.20a). The Taster's controller then picks one of the
// revealed cards (the reveal_pick prompt, which shows the chooser only
// the revealed cards of a pool that is not theirs), and its owner exiles
// it.
//
// If it landed in exile and is an instant or a sorcery, its exiler's
// opponent gets a cast permission whose window is "for as long as you
// control this creature" (#3192, ADR 0066's second 2026-10-13 amendment:
// game.ForAsLongAs + WhileYouControlSource, read by the one liveness
// test), cast only, spending mana of any type. A Taster that left, or
// changed hands, before the card was exiled gives a window that never
// started, and the card simply stays exiled (CR 611.2b). Any other card
// type is exiled with nothing granted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "020aae46-cdba-4c37-9c54-283727c09994",
		Name:         "Taster of Wares",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Taster of Wares — target opponent reveals X cards from their hand; you choose one and they exile it", tasterOfWaresEnters),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

func tasterOfWaresEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opp := TargetedPlayer(ctx)
	p := g.PlayerByIDForEffect(opp)
	if p == nil || p.Hand == nil {
		return nil
	}
	controller := item.Controller
	// "Goblin" is a word of the source's text (CR 612.2, #3984).
	goblin := ctx.TextSubtype("Goblin")
	x := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype(goblin) {
			x++
		}
	}
	hand := make([]uuid.UUID, 0, len(p.Hand.Cards))
	for _, c := range p.Hand.Cards {
		hand = append(hand, c.InstanceID)
	}
	if x <= 0 || len(hand) == 0 {
		return nil
	}
	revealAndPick := func(g *game.Game, revealed []uuid.UUID) error {
		if len(revealed) == 0 {
			return nil
		}
		g.RevealForEffect(game.RevealSpec{
			Player: opp,
			Source: item.SourceCardID,
			Reason: "Taster of Wares — reveal cards from your hand",
			Cards:  revealed,
		})
		_, err := g.RevealPickThenForEffect(game.RevealPickPrompt{
			Chooser:  controller,
			Owner:    opp,
			Source:   item.SourceCardID,
			Question: "Taster of Wares — choose one of those cards; that player exiles it",
			Cards:    revealed,
			Min:      1,
			Max:      1,
		}, func(g *game.Game, picked, _ []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return tasterOfWaresExile(g, item, picked[0])
		})
		return err
	}
	if x >= len(hand) {
		return revealAndPick(g, hand)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    opp,
		FromPlayer: opp,
		Source:     item.SourceCardID,
		Question:   "Taster of Wares — reveal cards from your hand",
		Cards:      hand,
		Min:        x,
		Max:        x,
		Zone:       game.ZoneHand,
		Then:       revealAndPick,
	})
	return nil
}

// tasterOfWaresExile is "that player exiles it", and the cast permission
// when an instant or sorcery card lands in exile.
func tasterOfWaresExile(g *game.Game, item *game.StackItem, id uuid.UUID) error {
	card, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	spell := card.IsInstant() || card.IsSorcery()
	return g.ExileCardsThenForEffect([]uuid.UUID{id}, func(g *game.Game, exiled []uuid.UUID) error {
		if !spell {
			return nil
		}
		for _, eid := range exiled {
			c, ok := g.LookupCardForEffect(eid)
			if !ok {
				continue
			}
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{
				Player:     item.Controller,
				Zone:       game.ZoneExile,
				CastOnly:   true,
				AnyColor:   true,
				AnyType:    true,
				Duration:   forAsLongAsYouControlThisWindow(g, item),
				Source:     item.SourceCardID,
				SourceName: "Taster of Wares",
				Label:      "You may cast it for as long as you control Taster of Wares",
			}, []game.Card{c})
		}
		return nil
	})
}
