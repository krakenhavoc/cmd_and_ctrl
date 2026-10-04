package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_stack_item_id_test.go — ADR 0119 §3, PR 4. A resolve, fizzle or
// counter entry names the stack item that left the stack, for an
// ability as well as a spell, so the client can match the entry to the
// item that departed between two frames.

func TestAnAbilitysResolveNamesItsStackItem(t *testing.T) {
	g := buildMainPhaseGame(t)
	me := g.Seats[0]
	rock := battlefieldCard(t, g, "Mind Stone", "Artifact", me.ID)
	if err := g.AnnounceTrigger(me.ID, rock, game.AbilityParams{Label: "Mind Stone — draw a card"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	item := onlyStackItem(t, g)
	resolveAbilities(t, g)

	entry := findLog(t, ViewOfGame(g).Log, LogResolve)
	if entry.StackItemID != item.String() {
		t.Errorf("stack_item_id = %q, want the item's %s", entry.StackItemID, item)
	}
	if entry.StackItemID == entry.CardID {
		t.Errorf("an ability's item id must not be its source's id %q", entry.CardID)
	}
}

func TestAnAbilitysFizzleNamesItsStackItem(t *testing.T) {
	g := buildMainPhaseGame(t)
	me := g.Seats[0]
	seer := battlefieldCard(t, g, "Prodigal Pyromancer", "Creature — Human Wizard", me.ID)
	victim := battlefieldCard(t, g, "Grizzly Bears", "Creature — Bear", g.Seats[1].ID)
	if err := g.ActivateAbility(me.ID, seer, game.AbilityParams{
		Label:   "{T}: deal 1 damage to any target",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	item := onlyStackItem(t, g)
	g.WithWriteLock(func() {
		if _, err := g.Battlefield.Remove(victim); err != nil {
			t.Fatalf("remove the target: %v", err)
		}
	})
	resolveAbilities(t, g)

	entry := findLog(t, ViewOfGame(g).Log, LogFizzle)
	if entry.StackItemID != item.String() {
		t.Errorf("stack_item_id = %q, want the item's %s", entry.StackItemID, item)
	}
}

func TestACounteredAbilityNamesItsStackItem(t *testing.T) {
	g := buildActiveGame(t)
	opp := g.Seats[1]
	rock := battlefieldCard(t, g, "Their Rock", "Artifact", opp.ID)
	if err := g.AnnounceTrigger(opp.ID, rock, game.AbilityParams{Label: "Their Rock — do a thing"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	item := onlyStackItem(t, g)
	if err := g.CounterAbility(item); err != nil {
		t.Fatalf("CounterAbility: %v", err)
	}
	entry := findLog(t, ViewOfGame(g).Log, LogCounter)
	if entry.StackItemID != item.String() {
		t.Errorf("stack_item_id = %q, want the item's %s", entry.StackItemID, item)
	}
	if entry.Target != "" {
		t.Errorf("target = %q; the ability's id rides stack_item_id only", entry.Target)
	}
}

// A spell's item ID is its card's ID, so its entries carry that.
func TestASpellsEntriesNameItsStackItem(t *testing.T) {
	g := buildMainPhaseGame(t)
	bolt := battlefieldCard(t, g, "Lightning Bolt", "Instant", g.Seats[0].ID)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Source: bolt, CardID: bolt})
		g.EmitEvent(game.Event{Kind: game.EventFizzle, Source: bolt, CardID: bolt})
		g.EmitEvent(game.Event{Kind: game.EventCounterSpell, Target: bolt, CardID: bolt})
	})
	log := ViewOfGame(g).Log
	for _, k := range []LogKind{LogResolve, LogFizzle, LogCounter} {
		if e := findLog(t, log, k); e.StackItemID != bolt.String() {
			t.Errorf("%s stack_item_id = %q, want the spell's %s", k, e.StackItemID, bolt)
		}
	}
}

// The item id survives redaction: the stack is public.
func TestTheStackItemIDSurvivesRedaction(t *testing.T) {
	g := buildMainPhaseGame(t)
	me, other := g.Seats[0], g.Seats[1]
	inHand := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: inHand, Name: "Hidden Card", TypeLine: "Instant",
			Owner: me.ID, Controller: me.ID,
			KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	if err := g.AnnounceTrigger(me.ID, inHand, game.AbilityParams{Label: "Hidden Card — reveal"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	item := onlyStackItem(t, g)
	resolveAbilities(t, g)
	theirs := findLog(t, FilterViewFor(ViewOfGame(g), other.ID.String()).Log, LogResolve)
	if theirs.CardID != "" || theirs.StackItemID != item.String() {
		t.Errorf("redacted entry = card_id %q stack_item_id %q; want no card and the item %s", theirs.CardID, theirs.StackItemID, item)
	}
}
