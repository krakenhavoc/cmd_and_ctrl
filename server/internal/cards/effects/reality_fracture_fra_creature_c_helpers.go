package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_creature_c_helpers.go — bodies and tokens the
// Reality Fracture creatures of slice fra-creature-c share or need. Every
// name carries the rfCreatureC prefix so a parallel slice cannot collide
// with it.

// printedLotusToken is Kwia Vigorbloom's Lotus: a colourless artifact
// token with "{T}, Sacrifice this token: Add three mana of any one
// color." The three mana are ONE colour pick, the Gilded Lotus shape.
func printedLotusToken() tokenTemplate {
	return tokenTemplate{
		Slug: "lotus",
		Card: game.Card{
			Name:     "Lotus",
			TypeLine: "Token Artifact",
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:       true,
			SacrificeCost: true,
			Produced:      OneColorOfAmount(3),
			Label:         "{T}, Sacrifice: Add three mana of any one color",
		}},
		Text: "{T}, Sacrifice this token: Add three mana of any one color.",
	}
}

// LotusToken is Kwia Vigorbloom's Lotus token.
func LotusToken() game.Card { return tokenFromCatalog(printedLotusToken) }

// rfCreatureCEnteredThisTurn is "creature that entered this turn".
func rfCreatureCEnteredThisTurn() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool { return g.EnteredThisTurn(c.InstanceID) }
}

// rfCreatureCGuidingHydraEffect is "you may remove a +1/+1 counter from
// this creature. If you do, put a +1/+1 counter on each other creature
// you control." The set is snapshotted first, so a creature a counter
// trigger makes mid-loop gets none.
func rfCreatureCGuidingHydraEffect(g *game.Game, item *game.StackItem) error {
	self, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !onBattlefield(g, self.InstanceID) || self.Counters[game.CounterPlusOne] < 1 {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: self.InstanceID, Kind: game.CounterPlusOne, N: -1}).Apply(ctx); err != nil {
		return err
	}
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.InstanceID != self.InstanceID {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if !onBattlefield(g, id) {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// rfCreatureCGreatestManaValueInGraveyard is "the greatest mana value
// among cards in your graveyard".
func rfCreatureCGreatestManaValueInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	best := 0
	for _, c := range p.Graveyard.Cards {
		if mv := c.ManaValue(); mv > best {
			best = mv
		}
	}
	return best
}

// rfCreatureCHapatraFangEffect puts X -1/-1 counters on each still-legal
// pick, X being the greatest mana value in the controller's graveyard,
// read as the ability resolves.
func rfCreatureCHapatraFangEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := rfCreatureCGreatestManaValueInGraveyard(g, item.Controller)
	if x <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: x}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// rfCreatureCHapatraFrostEffect taps each still-legal pick and puts a
// stun counter on it.
func rfCreatureCHapatraFrostEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		if err := (AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// rfCreatureCJhoiraEffect is Jhoira's reveal: the target opponent
// reveals cards from the top of their library until a historic permanent
// card; the controller puts it onto the battlefield under their control
// and loses life equal to its mana value; the opponent puts the rest on
// the bottom of their library in a random order.
func rfCreatureCJhoiraEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opp, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	run, hit := revealUntil(ctx, opp, func(c game.Card) bool {
		return isPermanentCard(c) && b09IsHistoric(c)
	}, "Jhoira, Weatherlight Corsair — revealed until a historic permanent card")
	bottom := func(g *game.Game) error {
		return g.PutOnBottomInRandomOrderForEffect(opp, game.ZoneLibrary, cardsStillInALibrary(g, run))
	}
	if hit == uuid.Nil {
		return bottom(g)
	}
	controller, source := item.Controller, ctx.Source()
	return g.PutCardsFromLibraryOntoBattlefieldThenForEffect([]uuid.UUID{hit},
		game.LibraryEntryOptions{Controller: controller},
		func(g *game.Game, entered []uuid.UUID) error {
			if len(entered) > 0 {
				if c, ok := g.LookupCardForEffect(entered[0]); ok {
					if mv := c.ManaValue(); mv > 0 {
						if err := g.ChangePlayerLifeForEffect(source, controller, -mv); err != nil {
							return err
						}
					}
				}
			}
			return bottom(g)
		})
}

// rfCreatureCAttacksAPlayerAlone is "whenever a creature you control
// attacks a player alone": one declared attacker, aimed at a player
// rather than a planeswalker or battle.
func rfCreatureCAttacksAPlayerAlone(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return attackDeclaredByYou(ev, source.Controller) &&
		game.AttackedAlone(g) &&
		g.PlayerByIDForEffect(ev.Target) != nil
}

// rfCreatureCJiangAloneEffect is "discard a card, then draw a card. Then
// put a +1/+1 counter on that creature for each card you've discarded
// this turn." The discard is the controller's choice; everything after
// it waits for the answer.
func rfCreatureCJiangAloneEffect(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	player, source, attacker := item.Controller, item.SourceCardID, item.Trigger.Event.CardID
	return g.PlayerDiscardsThenForEffect(game.DiscardPrompt{
		Player:   player,
		Source:   source,
		N:        1,
		Question: "Jiang Yanggu, Alone — discard a card, then draw a card",
	}, func(g *game.Game, _ game.PromptedDiscards) error {
		if err := g.DrawNForEffect(player, 1); err != nil {
			return err
		}
		n := g.TurnTallyFor(player).CardsDiscarded
		if n <= 0 || !onBattlefield(g, attacker) {
			return nil
		}
		return g.AddCounterForEffect(attacker, game.CounterPlusOne, n)
	})
}

// rfCreatureCKarnDrawEffect is "draw a card for each color among other
// artifacts you control".
func rfCreatureCKarnDrawEffect(g *game.Game, item *game.StackItem) error {
	seen := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.IsArtifact() || c.InstanceID == item.SourceCardID {
			continue
		}
		for _, col := range c.EffectiveColors() {
			seen[col] = true
		}
	}
	if len(seen) == 0 {
		return nil
	}
	return DrawCards{Player: item.Controller, N: len(seen)}.Apply(NewContext(g, item))
}

// rfCreatureCEnteredLandHasSubtype reads the land that triggered a
// landfall ability, which may have left the battlefield by resolution.
func rfCreatureCEnteredLandHasSubtype(g *game.Game, item *game.StackItem, subtype string) bool {
	if item.Trigger == nil {
		return false
	}
	c, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
	return ok && c.HasSubtype(subtype)
}

// rfCreatureCDifferentPowers is "each different power among creatures
// you control".
func rfCreatureCDifferentPowers(g *game.Game, controller, _ uuid.UUID) int {
	seen := map[int]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() {
			seen[c.CurrentPower()] = true
		}
	}
	return len(seen)
}

// rfCreatureCPutCounterOnEachAngel is Lyra, Archangel of Dawn's "put a
// +1/+1 counter on each Angel you control", the set snapshotted first.
func rfCreatureCPutCounterOnEachAngel(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.HasSubtype("Angel") {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if !onBattlefield(g, id) {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// rfCreatureCDrewThreeOrMore is Lyra, Tolarian Archangel's intervening
// "if you've drawn three or more cards this turn".
func rfCreatureCDrewThreeOrMore(g *game.Game, player uuid.UUID) bool {
	return g.TurnTallyFor(player).CardsDrawn >= 3
}

// Lyra, Tolarian Archangel's {3}{U}{U}: "Until end of turn, whenever
// Lyra deals combat damage to a player, draw two cards." CR 603.7b
// repeating delayed trigger, keyed on the ability's source.
var (
	rfCreatureCLyraCombatDamageCondition = game.DelayedCondition("fra-creature-c/lyra-combat-damage-to-a-player",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
			return ev.Source == dt.SourceCardID && combatDamageToPlayerBy(ev, dt.Controller, g)
		})
	rfCreatureCLyraDrawTwoBody = game.SimpleDelayedBody("fra-creature-c/lyra-draw-two",
		func(g *game.Game, item *game.StackItem) error {
			return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
		})
)
