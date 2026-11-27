package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b700_test.go — Mercadian Masques, "Counterspells and bounce,
// Artifacts" (S82 OLD batch 700, #5619): each card driven through a real
// activation, cast, trigger or damage event.

const (
	oldB700Beads     = "ab050d55-a47d-4880-b65a-29f562773d7a"
	oldB700Anemone   = "52305f9d-2bfd-4b35-92b3-c08cdd758cde"
	oldB700Escort    = "55cc14d0-11ed-444d-884a-cfb692ba632b"
	oldB700Sanctuary = "f9269fda-e1e6-4e10-9a64-26e8c0f38b0c"
)

func TestS82OldB700AllAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{oldB700Beads, oldB700Anemone, oldB700Escort, oldB700Sanctuary} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82OldB700WorryBeadsMillsTheUpkeepPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Worry Beads", oldB700Beads, "Artifact")
	for _, p := range g.Seats {
		effectstest.SeedLibrary(p, "A", "B", "C", "D", "E", "F")
	}
	for seat := range g.Seats {
		p := g.Seats[seat]
		lib, gy := p.Library.Size(), p.Graveyard.Size()
		effectstest.AdvanceToUpkeepOf(t, g, seat)
		effectstest.PassPriorityAroundTable(t, g)
		if p.Library.Size() != lib-1 || p.Graveyard.Size() != gy+1 {
			t.Errorf("seat %d: library %d→%d graveyard %d→%d, want one card milled at their own upkeep",
				seat, lib, p.Library.Size(), gy, p.Graveyard.Size())
		}
	}
}

func TestS82OldB700GlowingAnemoneMayBounceALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreatureNoResolve(t, g, "Glowing Anemone", "Creature — Jellyfish Beast", oldB700Anemone, 1, 3)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.PickTargetPrompt(t, g)
	for _, id := range p.PickTargetCards {
		if id == bear {
			t.Error("a nonland permanent was offered")
		}
	}
	effectstest.AnswerPickTarget(t, g, land)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !opp.Hand.Contains(land) {
		t.Error("the land did not return to its owner's hand")
	}
}

func TestS82OldB700GlowingAnemoneCanDeclineTheBounce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastHoldCreatureNoResolve(t, g, "Glowing Anemone", "Creature — Jellyfish Beast", oldB700Anemone, 1, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, land)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("the land was bounced after declining")
	}
}

func TestS82OldB700DiplomaticEscortCountersASpellThatTargetsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	escort := effectstest.PushCatalogPermanent(g, me.ID, "Diplomatic Escort", "Creature — Human Spellshaper", oldB700Escort, false)
	bear := effectstest.SeedCreature(g, "Their Bear", other.ID)
	for me.Hand.Size() > 0 {
		me.Hand.PopTop()
	}
	pitch := effectstest.HandSpell(me, "Pitched", "Sorcery", "")

	// A spell at a player is not a legal target.
	stray := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: other.ID})
	effectstest.FloatForTest(g, me, "U")
	if err := g.ActivateCatalogAbility(me.ID, escort, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: stray}}, DiscardIDs: []uuid.UUID{pitch}}); err == nil {
		t.Fatal("a spell that targets no creature was accepted")
	}
	effectstest.PassPriorityAroundTable(t, g)

	bolt := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: bear})
	effectstest.FloatForTest(g, me, "U")
	if err := g.ActivateCatalogAbility(me.ID, escort, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err == nil {
		t.Fatal("activated without discarding a card")
	}
	if err := g.ActivateCatalogAbility(me.ID, escort, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discard was not paid as a cost")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, escort); !tapped {
		t.Error("the Escort was not tapped")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || effectstest.DamageMarkedOn(g, bear) != 0 {
		t.Error("the Bolt was not countered")
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the countered Bolt is not in its owner's graveyard")
	}
}

func TestS82OldB700DiplomaticEscortIsSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	escort := effectstest.PushCatalogPermanent(g, me.ID, "Diplomatic Escort", "Creature — Human Spellshaper", oldB700Escort, true)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	pitch := effectstest.HandSpell(me, "Pitched", "Sorcery", "")
	bolt := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: bear})
	effectstest.FloatForTest(g, me, "U")
	if err := g.ActivateCatalogAbility(me.ID, escort, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}, DiscardIDs: []uuid.UUID{pitch}}); err == nil {
		t.Error("a creature that just arrived paid its {T} cost")
	}
}

func TestS82OldB700CrumblingSanctuaryExilesLibraryCardsInsteadOfDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Crumbling Sanctuary", oldB700Sanctuary, "Artifact")
	for _, p := range g.Seats {
		effectstest.SeedLibrary(p, "A", "B", "C", "D", "E", "F")
	}
	for _, victim := range []*game.Player{me, opp} {
		life, lib, exile := victim.Life, victim.Library.Size(), g.Exile.Size()
		effectstest.GptB1DamagePlayer(g, victim.ID, 3)
		effectstest.PassPriorityAroundTable(t, g)
		if victim.Life != life {
			t.Errorf("%s lost life: %d → %d", victim.Name, life, victim.Life)
		}
		if victim.Library.Size() != lib-3 || g.Exile.Size() != exile+3 {
			t.Errorf("%s: library %d→%d exile %d→%d, want 3 cards exiled", victim.Name, lib, victim.Library.Size(), exile, g.Exile.Size())
		}
	}
	// Damage to a creature is not replaced.
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	lib := opp.Library.Size()
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, bear, 1) })
	if opp.Library.Size() != lib {
		t.Error("damage to a creature exiled library cards")
	}
	if effectstest.DamageMarkedOn(g, bear) != 1 {
		t.Error("damage to a creature was replaced")
	}
}

func TestS82OldB700CrumblingSanctuaryWithAShortLibraryExilesWhatThereIs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Crumbling Sanctuary", oldB700Sanctuary, "Artifact")
	for me.Library.Size() > 0 {
		me.Library.PopTop()
	}
	effectstest.SeedLibrary(me, "A", "B")
	life := me.Life
	effectstest.GptB1DamagePlayer(g, me.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life || me.Library.Size() != 0 {
		t.Errorf("life %d→%d library %d, want no life lost and an empty library", life, me.Life, me.Library.Size())
	}
}
