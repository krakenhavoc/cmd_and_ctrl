package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jace, Mirror Mage — Legendary Planeswalker — Jace {1}{U}{U}, starting
// loyalty 4 (S82 GAP2, #5948, #6089):
//
//	"Kicker {2}
//	 When Jace enters, if Jace was kicked, create a token that's a copy of
//	 Jace, except it's not legendary and its starting loyalty is 1.
//	 +1: Scry 2.
//	 0: Draw a card and reveal it. Remove a number of loyalty counters
//	    equal to that card's mana value from Jace."
//
// The kicked enters trigger is an intervening "if" read off the
// permanent's kicker record (CR 603.4); the copy takes Jace's copiable
// values with the two printed exceptions (CR 707.9b): Legendary comes off
// the type line and the starting loyalty is 1, stamped as known so the
// token enters with one loyalty counter (CR 306.5b). The token was not
// cast, so it was not kicked and makes no copy of its own.
//
// The 0's draw is one draw instruction, which a replacement may change
// (CR 121.6, 614.11), so "that card" is what the draw actually put into
// the hand: the count of cards the player had drawn this turn rides the
// continuation (DrawThen.N) and the drawn cards are Game.DrawnThisTurn
// past it once the draw has finished (Mad Wizard's Lair's reading). Each is
// revealed, and the loyalty counters removed are their total mana value,
// read in the hand (an {X} is 0, CR 202.3e). A draw replaced by something
// else (Abundance, dredge) draws no card, so nothing is revealed and no
// counter is removed. A Jace that left the battlefield before the ability
// resolved still draws and reveals, and loses nothing (CR 608.2h: there is
// no Jace to remove counters from).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7ca99e2f-faf2-4301-a0ae-7cb880ba965b",
		Name:            "Jace, Mirror Mage",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{2}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Jace, Mirror Mage — kicked, create a token that's a copy of Jace, except it's not legendary and its starting loyalty is 1",
				jaceMirrorMageCopy),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Scry 2.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Scry: 2},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Scry{Player: item.Controller, N: 2}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "0: Draw a card and reveal it. Remove a number of loyalty counters equal to that card's mana value from Jace.",
				Cost:    effects.LoyaltyCost(0),
				Purpose: game.Purpose{Draws: 1},
				Effect:  jaceMirrorMageDraw,
			},
		},
	})
}

const jaceMirrorMageName = "Jace, Mirror Mage"

// jaceMirrorMageCopy creates the nonlegendary copy with starting loyalty 1.
func jaceMirrorMageCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.CreateTokenCopy{
		Controller: item.Controller,
		Copy:       ctx.Source(),
		N:          1,
		Except: func(t *game.Card) {
			t.TypeLine = effects.RemoveLegendaryFromTypeLine(t.TypeLine)
			t.StartingLoyalty = 1
			t.StartingLoyaltyKnown = true
		},
	}.Apply(ctx)
}

// jaceMirrorMageDraw is the 0: draw one card, then reveal what was drawn
// and remove its mana value in loyalty from this Jace.
func jaceMirrorMageDraw(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	you := item.Controller
	jace := ctx.Source()
	if ctx.SourceLeftBattlefield(jace) || ctx.IsNewSourceObject(jace) {
		jace = uuid.Nil
	}
	return g.DrawNThenForEffect(you, 1, game.DrawThen{
		Ref:    jaceMirrorMageAfterDraw,
		Player: you,
		Source: jace,
		N:      len(g.DrawnThisTurn[you]),
	})
}

// jaceMirrorMageAfterDraw reveals the drawn card and removes loyalty. d.N is
// how many cards the player had drawn this turn before the ability's draw;
// d.Source is Jace, or zero when Jace had already left. An on-disk key.
var jaceMirrorMageAfterDraw = game.RegisterDrawThen("jace-mirror-mage-reveal", func(g *game.Game, d game.DrawThen) error {
	all := g.DrawnThisTurn[d.Player]
	if d.N < 0 || d.N > len(all) {
		return nil
	}
	var drawn []uuid.UUID
	mv := 0
	for _, id := range all[d.N:] {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneHand {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok {
			drawn = append(drawn, id)
			mv += c.ManaValue()
		}
	}
	if len(drawn) == 0 {
		return nil
	}
	g.RevealForEffect(game.RevealSpec{
		Player: d.Player,
		Source: d.Source,
		Reason: jaceMirrorMageName + " — reveal the drawn card",
		Cards:  drawn,
	})
	if d.Source == uuid.Nil || mv == 0 {
		return nil
	}
	if z := g.FindCardZoneForEffect(d.Source); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	have := 0
	if c, ok := g.LookupCardForEffect(d.Source); ok {
		have = c.Counters[game.CounterLoyalty]
	}
	if mv > have {
		mv = have
	}
	if mv == 0 {
		return nil
	}
	return g.AddCounterForEffect(d.Source, game.CounterLoyalty, -mv)
})
