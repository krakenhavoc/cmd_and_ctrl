package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// annihilator.go — the catalog half of #2073 (ADR 0113 §2): the
// Eldrazi titans' shared shapes. Append-only, like every mechanic-named
// helper file.
//
// A printed "Annihilator N" needs nothing here: it is a canonical
// keyword token ("annihilator N", game/annihilator.go) and the engine
// derives one trigger per instance from the ability list, so declare
// it in PrintedKeywords and grant it with KeywordGrant like any other
// keyword. What lives here is what a token cannot say:
//
//   - AnnihilatorCounted: "annihilator X" whose X is read as it
//     resolves (Ulamog, the Defiler).
//   - WhenThisIsPutIntoAGraveyardFromAnywhere and
//     ShuffleYourGraveyardIntoYourLibrary: the titans' "When ~ is put
//     into a graveyard from anywhere, its owner shuffles their
//     graveyard into their library."
//   - WhenYouCastThisSpell: the titans' cast triggers.

// AnnihilatorCounted is "this creature has annihilator X, where X is
// …": an annihilator trigger (CR 702.86a) whose number is not printed
// and is read as the trigger RESOLVES — Ulamog, the Defiler's "the
// number of +1/+1 counters on it", which the 2024-06-07 ruling reads
// at resolution.
//
// The trigger is the keyword's own: it watches the source being
// declared as an attacker (CR 508.3a), records the defending player as
// it triggers (CR 508.5), and resolves through the same engine
// sacrifice (game.AnnihilatorSacrificeForEffect), which re-reads the
// defending player exactly as the keyword's body does. Only N differs:
// `count` is asked at resolution, and nothing happens when it is zero.
//
// The row carries Keyword "annihilator", so the coverage guard sees it
// as one.
func AnnihilatorCounted(label string, count func(ctx *Context) int) game.TriggeredAbility {
	return game.TriggeredAbility{
		Keyword:   game.KeywordAnnihilator,
		Watches:   []game.EventKind{game.EventAttack},
		AppliesTo: ThisAttacked,
		Key:       label,
		// A fill-in Build (ADR 0041 P9): it records the defending
		// player and leaves the Effect to the row.
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Player = g.AnnihilatorDefenderForEffect(ev, source.InstanceID)
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			return g.AnnihilatorSacrificeForEffect(item, count(NewContext(g, item)))
		},
	}
}

// CountersOnThisAsItResolves is the number of `kind` counters on the
// ability's source as the ability resolves: live while it is the same
// object on the battlefield, and as it last existed once it has left
// (CR 608.2h). Zero when there is no record of it.
func CountersOnThisAsItResolves(kind string) func(ctx *Context) int {
	return func(ctx *Context) int {
		info, ok := ctx.SourcePermanent()
		if !ok {
			return 0
		}
		return info.Counters[kind]
	}
}

// ThisWasPutIntoAGraveyardFromAnywhere — "when this is put into a
// graveyard from anywhere": the card arrived in a graveyard from any
// other zone. Each route emits exactly one event for the move, so the
// condition is one per route and an arrival never triggers twice:
//
//   - EventZoneMove: off the battlefield (a death, a sacrifice), out
//     of exile or the command zone, a spell that finished resolving, a
//     card put there from a library or hand by an effect;
//   - EventMill: milled (CR 701.17a);
//   - EventDiscardCard: discarded (CR 701.9a);
//   - EventCounterSpell: countered (CR 701.6a). The event names the
//     spell in both CardID and Target; a countered ABILITY names its
//     source in CardID and the ability in Target, and is not this.
//
// The ability is declared in the graveyard (WhenThisIsPutIntoAGraveyard
// FromAnywhere), so the harvest finds the card where it landed: a
// countered spell exiled instead, or a dying permanent a replacement
// sent elsewhere, is not in a graveyard and does not trigger.
func ThisWasPutIntoAGraveyardFromAnywhere(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	if ev.CardID != source.InstanceID {
		return false
	}
	switch ev.Kind {
	case game.EventCounterSpell:
		return ev.Target == ev.CardID
	case game.EventZoneMove, game.EventMill, game.EventDiscardCard:
		return ev.NewZone == game.ZoneGraveyard && ev.OldZone != game.ZoneGraveyard
	}
	return false
}

// WhenThisIsPutIntoAGraveyardFromAnywhere is the printed shape over
// that condition, scoped to the graveyard: the
// WhenThisIsPutIntoYourGraveyardFromYourLibrary shape widened to every
// zone the card can come from (ADR 0113 §2 decision 7). The ability's
// controller is the card's owner (CR 108.4: a card in a graveyard has
// no controller, and the harvest hands the owner the trigger).
func WhenThisIsPutIntoAGraveyardFromAnywhere(label string, effect Effect) game.TriggeredAbility {
	return InGraveyard(OnAny(
		[]game.EventKind{game.EventZoneMove, game.EventMill, game.EventDiscardCard, game.EventCounterSpell},
		ThisWasPutIntoAGraveyardFromAnywhere, label, effect))
}

// ShuffleYourGraveyardIntoYourLibrary is "its owner shuffles their
// graveyard into their library" as a graveyard trigger's effect: the
// trigger's controller is the card's owner, and every card in that
// player's graveyard — the titan included, if it is still there — goes
// into their library, which is then shuffled. An empty graveyard still
// shuffles the library.
func ShuffleYourGraveyardIntoYourLibrary(g *game.Game, item *game.StackItem) error {
	return finaleShuffleGraveyardIntoLibrary(NewContext(g, item), item.Controller)
}

// WhenYouCastThisSpell is "When you cast this spell, …" (CR 601.2i,
// 603.2): a trigger that fires from the stack as the spell is cast and
// resolves above it, so it resolves even if the spell is countered.
// The trigger is controlled by the caster. A target clause, if the
// effect has one, goes on the returned row's Targets.
func WhenYouCastThisSpell(label string, effect Effect) game.TriggeredAbility {
	return game.TriggeredAbility{
		FromStack: true,
		Watches:   []game.EventKind{game.EventCast},
		AppliesTo: Self,
		Key:       label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Controller, item.Owner = ev.Actor, ev.Actor
			return item
		},
		Effect: effect,
	}
}

// targetOpponentExilesTopHalfOfTheirLibrary is "target opponent exiles
// the top half of their library, rounded up" (Ulamog, the Defiler):
// the library is counted as the ability resolves, and exiling the top
// of a library is not milling (CR 701.17a), so nothing that watches a
// mill sees it.
func targetOpponentExilesTopHalfOfTheirLibrary(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var opp uuid.UUID
	for _, ref := range ctx.LegalTargets() {
		if ref.Kind == game.TargetPlayer {
			opp = ref.ID
			break
		}
	}
	p := ctx.PlayerByID(opp)
	if p == nil || p.Eliminated || p.Library == nil {
		return nil
	}
	n := (p.Library.Size() + 1) / 2
	if n <= 0 {
		return nil
	}
	return MillToZone{Player: p.ID, N: n, To: game.ZoneExile}.Apply(ctx)
}
