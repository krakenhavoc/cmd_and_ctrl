package game

import (
	"testing"

	"github.com/google/uuid"
)

type exchangeObserver func(*Game, Event)

func (f exchangeObserver) OnEvent(g *Game, ev Event) { f(g, ev) }

func TestArtifactExchangeWaitsForEntryAndSurvivesUndo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	out, in := ebLand(me.ID, "out"), ebLand(me.ID, "in")
	out.TypeLine, in.TypeLine = "Artifact", "Artifact"
	g.Battlefield.PushTop(out)
	me.Graveyard.PushTop(in)
	g.WithWriteLock(func() {
		ebPayLifeToUntap(g, in.InstanceID, me.ID, 2)
		if err := g.ExchangeArtifactWithGraveyardForEffect(uuid.Nil, out.InstanceID, in.InstanceID); err != nil {
			t.Fatal(err)
		}
	})
	choice := ebPayLifePrompt(g)
	if choice == nil || !g.Battlefield.Contains(out.InstanceID) || !me.Graveyard.Contains(in.InstanceID) {
		t.Fatal("the exchange moved before its entry question was answered")
	}
	snapshot := g.Clone()
	for i := 0; i < 2; i++ {
		if i > 0 {
			g.WithWriteLock(func() { g.RestoreFrom(snapshot) })
			me = g.Seats[0]
		}
		if err := g.ResolveEntryPayLife(choice.ID, me.ID, true); err != nil {
			t.Fatal(err)
		}
		if !me.Graveyard.Contains(out.InstanceID) || !g.Battlefield.Contains(in.InstanceID) || ebTapped(g, in.InstanceID) {
			t.Fatal("exchange did not finish identically after the entry answer")
		}
	}
}

func TestArtifactExchangeAnnouncesTheFinishedBoard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	out, in := ebLand(me.ID, "out"), ebLand(me.ID, "in")
	out.TypeLine, in.TypeLine, in.OracleID = "Artifact Creature — Test", "Artifact Creature — Test", "exchange-newcomer"
	out.Power, out.Toughness, in.Power, in.Toughness = 2, 2, 2, 2
	g.Battlefield.PushTop(out)
	me.Graveyard.PushTop(in)
	old := CatalogTriggers
	t.Cleanup(func() { CatalogTriggers = old })
	newcomerSawExit := false
	CatalogTriggers = func(key string) []TriggeredAbility {
		if key != in.OracleID {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventLTB}, AppliesTo: func(Event, *Card, Characteristic, *Game) bool { newcomerSawExit = true; return false }}}
	}
	var exits, entries int
	g.RegisterListener(exchangeObserver(func(g *Game, ev Event) {
		if ev.Kind == EventLTB && ev.CardID == out.InstanceID {
			exits++
			if !g.Battlefield.Contains(in.InstanceID) {
				t.Error("exit observed before the incoming artifact landed")
			}
		}
		if ev.Kind == EventETB && ev.CardID == in.InstanceID {
			entries++
			if g.Battlefield.Contains(out.InstanceID) {
				t.Error("entry observed before the outgoing artifact left")
			}
		}
	}))
	g.WithWriteLock(func() {
		if err := g.ExchangeArtifactWithGraveyardForEffect(uuid.Nil, out.InstanceID, in.InstanceID); err != nil {
			t.Fatal(err)
		}
	})
	if exits != 1 || entries != 1 || newcomerSawExit {
		t.Fatalf("exits=%d entries=%d newcomer saw exit=%v", exits, entries, newcomerSawExit)
	}
}

func TestArtifactExchangeExitReplacementPausesAndSurvivesUndo(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		g := newActiveGame(t)
		me := g.Seats[0]
		out, in := ebLand(me.ID, "out"), ebLand(me.ID, "in")
		out.TypeLine, in.TypeLine = "Artifact", "Artifact"
		g.Battlefield.PushTop(out)
		me.Graveyard.PushTop(in)
		g.WithWriteLock(func() {
			g.testReplacements = append(g.testReplacements, ReplacementEffect{
				Watches: []EventKind{EventZoneMove}, Optional: true,
				AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
					return ev.Kind == RepEventMove && ev.CardID == out.InstanceID && ev.OldZone == ZoneBattlefield
				},
				Replace:    func(ev *ReplacementEvent, _ *Game, _ *Card) error { ev.Cancel(); return nil },
				Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return me.ID },
				Label:      "Test cancel sacrifice", PromptQuestion: "Cancel this move?",
			})
			if err := g.ExchangeArtifactWithGraveyardForEffect(uuid.Nil, out.InstanceID, in.InstanceID); err != nil {
				t.Fatal(err)
			}
		})
		choice := onlyPrompt(t, g)
		if !g.Battlefield.Contains(out.InstanceID) || !me.Graveyard.Contains(in.InstanceID) {
			t.Fatal("exchange moved before its exit replacement was answered")
		}
		snapshot := g.Clone()
		for run := 0; run < 2; run++ {
			if run > 0 {
				g.WithWriteLock(func() { g.RestoreFrom(snapshot) })
			}
			if err := g.ResolveOptionalReplacement(choice.ID, choice.Chooser, cancel); err != nil {
				t.Fatal(err)
			}
			if !g.Battlefield.Contains(in.InstanceID) || g.Battlefield.Contains(out.InstanceID) != cancel {
				t.Fatalf("exit replacement result differs after undo, cancel=%v", cancel)
			}
		}
	}
}

func TestArtifactExchangeNewcomerDoesNotModifyDepartureTriggers(t *testing.T) {
	for _, modifier := range []string{"doubler", "suppressor"} {
		t.Run(modifier, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			out, in := ebLand(me.ID, "out"), ebLand(me.ID, "in")
			out.TypeLine, in.TypeLine = "Artifact", "Artifact"
			out.OracleID, in.OracleID = "exchange-departing", "exchange-modifier"
			g.Battlefield.PushTop(out)
			me.Graveyard.PushTop(in)
			withCatalogTriggers(t, func(key string) []TriggeredAbility {
				if key != out.OracleID {
					return nil
				}
				return []TriggeredAbility{{Watches: []EventKind{EventLTB}, AppliesTo: func(ev Event, _ *Card, _ Characteristic, _ *Game) bool { return ev.CardID == out.InstanceID },
					Build: func(_ Event, c *Card, _ Characteristic, _ *Game) *StackItem { return NewTriggeredItem(c, "Departure") }}}
			})
			withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
				if modifier == "doubler" && key == in.OracleID {
					return []TriggerDoubler{{Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
				}
				return nil
			})
			withCatalogTriggerSuppressors(t, func(key string) []TriggerSuppressor {
				if modifier == "suppressor" && key == in.OracleID {
					return []TriggerSuppressor{{Suppresses: func(*Game, TriggerSuppressionQuery) bool { return true }}}
				}
				return nil
			})
			g.WithWriteLock(func() {
				if err := g.ExchangeArtifactWithGraveyardForEffect(uuid.Nil, out.InstanceID, in.InstanceID); err != nil {
					t.Fatal(err)
				}
			})
			if len(g.PendingTriggers) != 1 {
				t.Fatalf("newcomer %s changed departing trigger count: %d", modifier, len(g.PendingTriggers))
			}
		})
	}
}
