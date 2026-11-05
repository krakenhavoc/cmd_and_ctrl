package game

import "testing"

// TriggerDoublingQuery.SourceIsPermanent (#5011, Krang, the All-Powerful):
// true for a battlefield permanent's ability, false for a spell's own
// "when you cast" ability.
func TestTriggerDoublerSourceIsPermanent(t *testing.T) {
	onlyPermanents := func(t *testing.T, doublerOracle string) {
		withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
			if key == doublerOracle {
				return []TriggerDoubler{{Applies: func(_ *Game, q TriggerDoublingQuery) bool { return q.SourceIsPermanent }}}
			}
			return nil
		})
	}
	t.Run("permanent", func(t *testing.T) {
		g := newActiveGame(t)
		owner := g.Seats[0]
		const sourceOracle, doublerOracle = "permanent-source", "permanent-doubler"
		source := seedTriggerSource(g, owner, sourceOracle)
		pushBattlefieldForTest(g, owner.ID, "Doubler", "Artifact", doublerOracle)
		withCatalogTriggers(t, func(key string) []TriggeredAbility {
			if key != sourceOracle {
				return nil
			}
			return []TriggeredAbility{{Watches: []EventKind{EventDrawCard}, AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, "draw")
			}}}
		})
		onlyPermanents(t, doublerOracle)
		g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventDrawCard, CardID: source, Actor: owner.ID}) })
		if got := len(g.PendingTriggers); got != 2 {
			t.Fatalf("permanent instances = %d, want 2", got)
		}
	})
	t.Run("stack spell", func(t *testing.T) {
		g := newActiveGame(t)
		owner := g.Seats[0]
		const spellOracle, doublerOracle = "spell-source", "spell-doubler"
		spell := NewCard("Spell Source", owner.ID)
		spell.OracleID, spell.TypeLine = spellOracle, "Creature"
		g.Stack.PushTop(spell)
		pushBattlefieldForTest(g, owner.ID, "Doubler", "Artifact", doublerOracle)
		withCatalogTriggers(t, func(key string) []TriggeredAbility {
			if key != spellOracle {
				return nil
			}
			return []TriggeredAbility{{FromStack: true, Watches: []EventKind{EventCast}, AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true }, Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, "cast")
			}}}
		})
		onlyPermanents(t, doublerOracle)
		g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventCast, CardID: spell.InstanceID, Actor: owner.ID}) })
		if got := len(g.PendingTriggers); got != 1 {
			t.Fatalf("stack spell instances = %d, want 1", got)
		}
	})
}
