package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB429LifeBelowSeven is "your life total is less than 7" (Elderscale
// Wurm's enters trigger).
func S82OldB429LifeBelowSeven(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	return p != nil && p.Life < 7
}

// S82OldB429SelfAlsoSubtype is "this creature is a <subtype> in addition to its
// other types" as a layer 4 static on the source alone.
func S82OldB429SelfAlsoSubtype(label, subtype string) game.StaticAbility {
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target != nil && source != nil && target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if !slices.Contains(c.Subtypes, subtype) {
				c.Subtypes = append(c.Subtypes, subtype)
			}
		},
	}
}

// S82OldB429SevenLifeFloor is Elderscale Wurm's "As long as you have 7 or more
// life, damage that would reduce your life total to less than 7 reduces it to 7
// instead": a replacement on the life window (CR 614, CR 120.3a) for its
// controller, applying only to the life damage costs and only while the total
// stands at 7 or more as the damage is dealt. The damage itself stays whole.
func S82OldB429SevenLifeFloor(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventChangeLife},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || src.Controller == uuid.Nil || ev.LifePlayer != src.Controller || !ev.LifeFromDamage || ev.LifeDelta >= 0 {
				return false
			}
			p := g.PlayerByIDForEffect(ev.LifePlayer)
			return p != nil && p.Life >= 7 && p.Life+ev.LifeDelta < 7
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
			g.LifeTotalBecomesInsteadForEffect(ev, 7)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// S82OldB429PermanentsYouControl counts every permanent the controller controls,
// read as the ability resolves (Touch of the Eternal).
func S82OldB429PermanentsYouControl(ctx *Context) int {
	return CountYouControl(ctx, nil)
}

// S82OldB429TuckThisAndTarget is Void Stalker's ability: "Put this creature and
// target creature on top of their owners' libraries, then those players shuffle
// their libraries." Each creature still on the battlefield goes on top of its
// owner's library as one simultaneous move; the tail then shuffles the owner of
// each card that reached a library (a player who owns both shuffles once).
func S82OldB429TuckThisAndTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	if OnBattlefield(g, item.SourceCardID) {
		ids = append(ids, item.SourceCardID)
	}
	if t, ok := FirstTargetCard(item); ok && ctx.IsTargetLegal(t) && OnBattlefield(g, t.ID) {
		ids = append(ids, t.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{}, ctx.Tail(s82OldB429VoidStalkerShuffle))
}

// s82OldB429VoidStalkerShuffle is "then those players shuffle their libraries":
// the owner of each card that was put into a library shuffles it, once. An
// on-disk key: never renamed, never reused.
var s82OldB429VoidStalkerShuffle = CardTail("then/void-stalker-shuffle", func(ctx *Context, o game.TailOutcome) error {
	var owners []uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || slices.Contains(owners, c.Owner) {
			continue
		}
		owners = append(owners, c.Owner)
	}
	for _, owner := range owners {
		if err := ctx.Game.ShuffleLibraryForEffect(owner); err != nil {
			return err
		}
	}
	return nil
})

// S82OldB429RevealAndFreeCast is Mindclaw Shaman's resolution: the target
// opponent reveals their hand, and the controller may cast an instant or
// sorcery spell from among those cards without paying its mana cost (the
// resolution cast of Extract Brain, out of the opponent's hand).
func S82OldB429RevealAndFreeCast(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opp := TargetedPlayerOf(ctx)
	if opp == uuid.Nil {
		return nil
	}
	g.RevealHandForEffect(opp)
	perm := FreeInstantOrSorcery().On(g.FreeCastFromHandPermission(ctx.Controller(), ctx.Source(), "Mindclaw Shaman"))
	candidates := FreeCastCandidates(g, &perm, AllHandCardIDs(g, opp))
	if len(candidates) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      ctx.Controller(),
		Source:       ctx.Source(),
		Reason:       "Mindclaw Shaman",
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it in their hand",
	})
	return nil
}
