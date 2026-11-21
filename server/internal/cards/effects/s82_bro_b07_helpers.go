package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82BroB07SecondDrawCounter is "Whenever you draw your second card each
// turn, put a +1/+1 counter on this creature." (Lat-Nam Adept, Thopter
// Mechanic, Gurgling Anointer).
func S82BroB07SecondDrawCounter(name string) game.TriggeredAbility {
	return On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return B41DrewYourSecondCardThisTurn(ev, source, g)
	}, name+" — put a +1/+1 counter on this creature",
		Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))
}

// s82BroB07PumpSelfOnArtifact is "Whenever an artifact you control enters,
// this creature gets +1/+0 until end of turn." (Phalanx Vanguard,
// Perimeter Patrol). The artifact may be the creature itself, and the pump
// is only for the creature that is still on the battlefield.
func S82BroB07PumpSelfOnArtifact(name string) game.TriggeredAbility {
	return On(game.EventETB, ArtifactEnteredUnderYourControl, name+" — this creature gets +1/+0 until end of turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if !B15OnBattlefield(g, ctx.Source()) {
				return nil
			}
			return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: name + " — +1/+0"}.Apply(ctx)
		})
}

// s82BroB07DrewTwoOrMore is "as long as you've drawn two or more cards
// this turn" read for the static's source. The tally is not on the
// battlefield, so a static using it declares DependsOnHandSize.
func S82BroB07DrewTwoOrMore(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID && g.TurnTallyFor(source.Controller).CardsDrawn >= 2
}

// s82BroB07SelfKeywordWhileDrawnTwo grants one keyword to the source while
// its controller has drawn two or more cards this turn.
func S82BroB07SelfKeywordWhileDrawnTwo(keyword string) game.StaticAbility {
	ab := KeywordGrant(S82BroB07DrewTwoOrMore, keyword)
	ab.Label = "Has " + keyword + " as long as you've drawn two or more cards this turn."
	ab.DependsOnHandSize = true
	return ab
}
