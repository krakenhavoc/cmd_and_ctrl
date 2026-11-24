package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b034_helpers.go — shared pieces of S82 older-sets batch 34
// (Modern Horizons 2: life gain and loss).

// OldB034DiscerningTasteTake is Discerning Taste's look: the top four, one of
// them into your hand (the answer is mandatory), then the rest into your
// graveyard and the life gain read off what really went there.
var OldB034DiscerningTasteTake = RegisterTake("choose-cards/old-b034-discerning-taste-take",
	TakeDef{Then: oldB034DiscerningTasteRest})

// oldB034DiscerningTasteRest is "put the rest into your graveyard. You gain
// life equal to the greatest power among creature cards put into your
// graveyard this way": the creature cards that reached the graveyard are
// counted, by the power they print.
func oldB034DiscerningTasteRest(g *game.Game, res TakeFromLibraryResult) error {
	if err := RestIntoGraveyard(g, res.Rest); err != nil {
		return err
	}
	greatest := 0
	for _, id := range res.Rest {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok || !c.IsCreature() || IsToken(c) {
			continue
		}
		if c.Power > greatest {
			greatest = c.Power
		}
	}
	if greatest <= 0 {
		return nil
	}
	ctx := NewContext(g, &game.StackItem{Controller: res.Player, SourceCardID: res.Source})
	return GainLife{Player: res.Player, Amount: greatest}.Apply(ctx)
}

// OldB034DiscerningTaste is the spell's body: look at the top four cards of
// your library and take one.
func OldB034DiscerningTaste(item *game.StackItem, ctx *Context) error {
	player := ctx.Controller()
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
		Max:    1,
		Label:  "Discerning Taste — put one of them into your hand (the rest go into your graveyard)",
		Key:    OldB034DiscerningTasteTake,
	}.Apply(ctx)
}

// OldB034YouHaveMoreLifeThanAnOpponent is Glorious Enforcer's intervening
// "if": your life total is greater than at least one opponent still in the
// game.
func OldB034YouHaveMoreLifeThanAnOpponent(g *game.Game, you uuid.UUID) bool {
	me := g.PlayerByIDForEffect(you)
	if me == nil {
		return false
	}
	for _, p := range g.Seats {
		if p == nil || p.ID == you || p.Eliminated {
			continue
		}
		if me.Life > p.Life {
			return true
		}
	}
	return false
}

// OldB034HellbentKeyword is "Hellbent — This creature has <keyword> as long as
// you have no cards in hand": the grant, held to an empty hand and read live on
// every recompute.
func OldB034HellbentKeyword(keyword string) game.StaticAbility {
	s := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && R35YouHaveNoCardsInHand(g, source.Controller, uuid.Nil)
	}, keyword)
	s.Label = "Hellbent — Has " + keyword + " as long as you have no cards in hand."
	s.DependsOnHandSize = true
	return s
}

// OldB034TapThisAfter is "Tap it": the ability's source is tapped after the
// grant, as the object the ability came from.
func OldB034GrantIndestructibleThenTap(g *game.Game, item *game.StackItem) error {
	if err := ThisGainsKeywordUntilEndOfTurn("indestructible", "Necromancer's Familiar — indestructible until end of turn")(g, item); err != nil {
		return err
	}
	return TapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
}
