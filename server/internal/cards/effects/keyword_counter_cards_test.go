package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyword_counter_cards_test.go — ADR 0101 (#1753): the catalog half of
// keyword counters. The engine rules are pinned in
// game/keyword_counters_test.go.

func TestPerennationReturnsAPermanentWithHexproofAndIndestructibleCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Glorious Anthem", TypeLine: "Enchantment",
		ManaCost: "{1}{W}{W}", Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Perennation", "Sorcery", perennationOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
	passPriorityAroundTable(t, g)

	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Perennation returns a permanent card of any type — the enchantment is not on the battlefield")
	}
	if card.Controller != me.ID {
		t.Errorf("controller %s, want the caster (its owner)", card.Controller)
	}
	if card.Counters[game.CounterHexproof] != 1 || card.Counters[game.CounterIndestructible] != 1 {
		t.Fatalf("counters = %v, want one hexproof and one indestructible", card.Counters)
	}
	abilities := effectiveAbilities(t, g, id)
	if !hasAbility(abilities, "hexproof") || !hasAbility(abilities, "indestructible") {
		t.Errorf("abilities = %v: the counters grant hexproof and indestructible (CR 122.1b) with nothing on the battlefield carrying a grant", abilities)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	if _, ok := battlefieldCard(g, id); !ok {
		t.Error("the indestructible counter keeps it from being destroyed")
	}
}

func TestPerennationTargetsOnlyYourOwnPermanentCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := seedGraveyardCreature(opp, "Their Zombie", "{4}{B}")
	sorcery := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: sorcery, Name: "Divination", TypeLine: "Sorcery",
		ManaCost: "{2}{U}", Owner: me.ID, Controller: me.ID,
	})
	spec, _ := Lookup(perennationOracle)
	for _, bad := range []uuid.UUID{theirs, sorcery} {
		var legal bool
		g.WithWriteLock(func() {
			for _, c := range g.LegalTargetsForEffect(game.SourceChooser(me.ID), spec.Targets).Cards {
				if c == bad {
					legal = true
				}
			}
		})
		if legal {
			t.Errorf("%s is not a permanent card in YOUR graveyard, but is a legal target", bad)
		}
	}
}

// The corner b24KeywordCounterGrant got wrong: a Tekuthal that lost
// all its abilities BEFORE its indestructible counter arrived is still
// indestructible, because the counter is newer than the removal
// (CR 613.3, 613.7c). The old static was an ability of Tekuthal, so
// the removal silenced it.
func TestTekuthalsCounterBeatsAnOlderLoseAllAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tekuthal := pushTekuthal(g, me.ID)
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(tekuthal),
			[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities") {
			t.Fatal("setup: the removal registered nothing")
		}
	})
	if hasAbility(effectiveAbilities(t, g, tekuthal), "flying") {
		t.Fatal("setup: the removal takes Tekuthal's printed flying")
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(tekuthal, game.CounterIndestructible, 1); err != nil {
			t.Fatal(err)
		}
	})
	if !hasAbility(effectiveAbilities(t, g, tekuthal), "indestructible") {
		t.Error("the indestructible counter is newer than the removal, so Tekuthal is indestructible")
	}
}

// No card declares the rule any more (ADR 0101 Decision 8): the five
// that used to carry b24KeywordCounterGrant have no static for it, and
// Vraska Joins Up and Sorin, Ravenous Neonate lost their caveats.
func TestTheKeywordCounterWorkaroundIsGone(t *testing.T) {
	for _, oracle := range []string{b24VraskaJoinsUpOracle, sorinOfHouseMarkovOracleID + "#1"} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v — want Full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	for _, oracle := range []string{b24VraskaJoinsUpOracle, sorinOfHouseMarkovOracleID + "#1", tekuthalOracle,
		"895f23a2-55b7-4cc0-8939-2efaaf097e6f", "51780f71-bf60-4208-94ea-76fa84790fb6"} {
		spec, _ := Lookup(oracle)
		if len(spec.Static) != 0 {
			t.Errorf("%s still declares %d static abilities; the engine owns keyword counters", spec.Name, len(spec.Static))
		}
	}
}
