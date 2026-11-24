package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// River Song's Diary — {3} Artifact — Book
//
//	"Imprint — Whenever a player casts an instant or sorcery spell from
//	 their hand, exile it instead of putting it into a graveyard as it
//	 resolves.
//	 At the beginning of your upkeep, if there are four or more cards
//	 exiled with this artifact, choose one of them at random. You may cast
//	 it without paying its mana cost."
//
// The first ability is Rod of Absorption's shape with a narrower trigger:
// the cast event's origin zone must be the caster's hand (Event.OldZone),
// so a spell cast from a graveyard, from exile or from the command zone
// does not trigger it. Resolving, it marks the spell so that, as it
// resolves, it is exiled instead of going to a graveyard, linked to this
// Diary (CR 607.2a; Game.MarkSpellExilesLinkedForEffect). A spell countered
// or one that leaves the stack another way is not exiled ("as it
// resolves"). A Diary that has left the battlefield links nothing.
//
// The second is an intervening-if upkeep trigger (CR 603.4): it triggers
// only with four or more cards exiled with this object, and checks again
// as it resolves. One of those cards is chosen at random (the game's
// keyed random stream) and its controller may cast it without paying its
// mana cost during the resolution (CR 608.2g); declining leaves it in
// exile, still linked.
//
// No simplification.
func init() {
	upkeep := effects.AtYourUpkeep("River Song's Diary — choose a card exiled with it at random; you may cast it without paying its mana cost",
		riverSongsDiaryCastOne)
	applies := upkeep.AppliesTo
	upkeep.AppliesTo = func(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
		return applies(ev, source, c, g) && len(effects.Cmdr2S36CardsExiledWithThis(g, source)) >= 4
	}
	effects.Register(effects.Spec{
		OracleID:     "60a63eb4-b0f7-4930-9bf6-afd42261d4d0",
		Name:         "River Song's Diary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, riverSongsDiaryCastFromHand,
				"River Song's Diary — exile that spell instead of putting it into a graveyard as it resolves",
				effects.MarkTriggeringSpellExilesLinked),
			effects.InterveningIf(riverSongsDiaryFourOrMore{}, upkeep),
		},
	})
}

// riverSongsDiaryCastFromHand is "whenever a player casts an instant or
// sorcery spell from their hand".
func riverSongsDiaryCastFromHand(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == uuid.Nil || ev.OldZone != game.ZoneHand {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (spell.IsInstant() || spell.IsSorcery())
}

// riverSongsDiaryFourOrMore is "if there are four or more cards exiled
// with this artifact", read off the source object (FootprintCondSource).
type riverSongsDiaryFourOrMore struct{}

func (riverSongsDiaryFourOrMore) Holds(ctx *effects.Context) bool {
	ref, ok := ctx.SourceRef()
	return ok && len(ctx.Game.ExiledWithForEffect(game.PermissionCardRef(ref))) >= 4
}

func (riverSongsDiaryFourOrMore) Condition() (game.FootprintStep, bool) {
	return effects.CondStep(game.FootprintCondSource, 0), true
}

// riverSongsDiaryCastOne picks one of the linked cards at random and offers
// its controller a free cast of it.
func riverSongsDiaryCastOne(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	pick := g.ChooseAtRandomForEffect(effects.RandomDraw(ctx), ids, 1)
	if len(pick) == 0 {
		return nil
	}
	return g.CastExiledFreeKeepingForEffect(ctx.Controller(), ctx.Source(), pick[0])
}
