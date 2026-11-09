package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chaos Shrine's Black Crystal — Legendary Artifact {3}{B} (Final Fantasy,
// #4462):
//
//	"Whenever a nontoken creature you control dies, exile it.
//	 At the beginning of your upkeep, you may put a creature card exiled
//	 with Chaos Shrine's Black Crystal onto the battlefield under your
//	 control with a finality counter on it."
//
// The death trigger reads the creature's types and controller as it last
// existed (CR 603.10a) and, as it resolves, exiles THAT card — the object
// it became in the graveyard — linked to this Crystal (ExileTarget.Linked,
// CR 607.2a), so a card that has since left the graveyard is left alone.
// The upkeep ability offers the creature cards exiled with this very
// Crystal object (Negative Zone Portal's read) and puts the one chosen
// onto the battlefield under your control, whoever owns it, with a
// finality counter as it enters (CR 614.1c, 122.1h: it is exiled instead
// of dying). A Crystal that left and returned is a new object with no
// cards exiled with it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1ffba56-1c1b-44a5-ad8c-097c417ab276",
		Name:         "Chaos Shrine's Black Crystal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && !IsToken(c) && leftAsType(ev, c, "creature") && leftUnderControlOf(ev, c) == source.Controller
				},
				Key: blackCrystalExileLabel,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, blackCrystalExileLabel)
					epoch := -1
					if c, ok := g.LookupCardForEffect(ev.CardID); ok {
						epoch = c.ObjectEpoch
					}
					item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: epoch}
					return item
				},
				Effect: blackCrystalExile,
			},
			AtYourUpkeep(blackCrystalPutLabel, blackCrystalOffer),
		},
	})
}

const (
	blackCrystalExileLabel = "Chaos Shrine's Black Crystal — exile the creature that died"
	blackCrystalPutLabel   = "Chaos Shrine's Black Crystal — you may put a creature card exiled with it onto the battlefield with a finality counter"
)

// blackCrystalExile exiles the dead creature if it is still the same card
// in a graveyard.
func blackCrystalExile(g *game.Game, item *game.StackItem) error {
	ref := item.Params.Object
	c, ok := g.LookupCardForEffect(ref.ID)
	if !ok || c.ObjectEpoch != ref.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(ref.ID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ExileTarget{Target: ref.ID, Linked: true}.Apply(NewContext(g, item))
}

// blackCrystalLinkedCreatures lists the creature cards exiled with this
// Crystal object.
func blackCrystalLinkedCreatures(g *game.Game, crystal uuid.UUID) []uuid.UUID {
	c, ok := g.LookupCardForEffect(crystal)
	if !ok {
		return nil
	}
	var out []uuid.UUID
	for _, e := range g.ExiledWithForEffect(game.PermissionCardRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}) {
		if e.IsCreature() {
			out = append(out, e.InstanceID)
		}
	}
	return out
}

func blackCrystalOffer(g *game.Game, item *game.StackItem) error {
	cands := blackCrystalLinkedCreatures(g, item.SourceCardID)
	if len(cands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Chaos Shrine's Black Crystal — you may put a creature card exiled with it onto the battlefield under your control with a finality counter on it",
		Cards:    cands,
		Min:      0,
		Max:      1,
		ThenKey:  blackCrystalPick,
		LiveZone: game.ZoneExile,
	})
	return nil
}

// blackCrystalPick returns the chosen card with its finality counter. An
// on-disk key.
var blackCrystalPick = game.RegisterChooseCardsThen("choose-cards/chaos-shrines-black-crystal-return", func(g *game.Game, r game.CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	return g.ReturnFromExileWithCountersKeyedForEffect(id, r.Chooser, false,
		map[string]int{game.CounterFinality: 1}, game.KeyedTail{})
})
