package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b569_haunt_test.go — haunt on an instant or sorcery (CR 702.55a's
// second sentence, #6083), proved on three Guildpact cards of OLD batch 569
// (#5599). Each test plays the real catalog entry.

const (
	oldB569BenedictionOracle = "3ff2fc3c-0871-4ba7-9b7c-8634ec770d6b"
	oldB569CryOracle         = "2b455d6a-dfee-4828-8873-094caf85c584"
	oldB569SeizeOracle       = "91557f08-15b1-4e3e-90fd-cec35ed0868a"
)

func TestOldB569HauntSpellsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB569BenedictionOracle: "Benediction of Moons",
		oldB569CryOracle:         "Cry of Contrition",
		oldB569SeizeOracle:       "Seize the Soul",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// oldB569Settle resolves the stack, answering each target pick with pick,
// and stops at any other prompt.
func oldB569Settle(t *testing.T, g *game.Game, pick func(c *game.PendingChoice) (game.TargetRef, bool)) {
	t.Helper()
	for i := 0; i < 40; i++ {
		if c := effectstest.FirstPendingTarget(g); c != nil {
			ref, ok := pick(c)
			if !ok {
				t.Fatalf("unexpected target pick: cards %v players %v", c.PickTargetCards, c.PickTargetPlayers)
			}
			if err := g.ResolvePickTarget(c.ID, c.Chooser, ref); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if len(g.PendingChoices) > 0 || effectstest.StackFullyEmpty(g) {
			return
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	t.Fatal("the stack never settled")
}

func oldB569Card(id uuid.UUID) func(*game.PendingChoice) (game.TargetRef, bool) {
	return func(c *game.PendingChoice) (game.TargetRef, bool) {
		if !effectstest.HasID(c.PickTargetCards, id) {
			return game.TargetRef{}, false
		}
		return game.TargetRef{Kind: game.TargetCard, ID: id}, true
	}
}

func oldB569Player(id uuid.UUID) func(*game.PendingChoice) (game.TargetRef, bool) {
	return func(c *game.PendingChoice) (game.TargetRef, bool) {
		if !effectstest.HasID(c.PickTargetPlayers, id) {
			return game.TargetRef{}, false
		}
		return game.TargetRef{Kind: game.TargetPlayer, ID: id}, true
	}
}

func oldB569NoPick(*game.PendingChoice) (game.TargetRef, bool) { return game.TargetRef{}, false }

func oldB569Kill(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	effectstest.KillCreature(t, g, owner, id)
	g.RunStateChecksForTest()
}

// oldB569Haunting is the card `id` in exile and what it haunts.
func oldB569Haunting(g *game.Game, id uuid.UUID) (uuid.UUID, bool) {
	var haunted uuid.UUID
	var ok bool
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.InstanceID == id {
				haunted, ok = c.HauntedID, true
			}
		}
	})
	return haunted, ok
}

func oldB569Spirits(g *game.Game, controller uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == "Spirit" && c.IsToken() {
				n++
			}
		}
	})
	return n
}

func oldB569Players(g *game.Game) int {
	n := 0
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			n++
		}
	}
	return n
}

// --- Benediction of Moons ------------------------------------------------

func TestOldB569BenedictionOfMoonsHauntsAsItResolvesAndGainsAgainWhenTheCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	bear := effectstest.DftABPush(g, opp.ID, "Opp Bear", "Creature — Bear", "", 2, 2)
	other := effectstest.DftABPush(g, opp.ID, "Opp Elk", "Creature — Elk", "", 2, 2)
	players := oldB569Players(g)

	life := me.Life
	id := effectstest.P1g2Cast(t, g, "Benediction of Moons", "Sorcery", oldB569BenedictionOracle, "{W}", game.CastSpellParams{})
	oldB569Settle(t, g, oldB569Card(bear))
	if me.Life != life+players {
		t.Fatalf("the spell: life %d→%d, want +%d (one for each player)", life, me.Life, players)
	}
	if haunted, ok := oldB569Haunting(g, id); !ok || haunted != bear {
		t.Fatalf("after resolving: in exile %v, haunting %v; want haunting the bear", ok, haunted)
	}

	// Another creature dying is not the creature it haunts.
	life = me.Life
	oldB569Kill(t, g, opp.ID, other)
	oldB569Settle(t, g, oldB569NoPick)
	if me.Life != life {
		t.Fatalf("another creature's death gained life: %d→%d", life, me.Life)
	}

	oldB569Kill(t, g, opp.ID, bear)
	oldB569Settle(t, g, oldB569NoPick)
	if me.Life != life+players {
		t.Errorf("the haunted creature died: life %d→%d, want +%d", life, me.Life, players)
	}
	if _, ok := oldB569Haunting(g, id); !ok {
		t.Error("the haunting card left exile")
	}
}

func TestOldB569BenedictionOfMoonsCountsOnlyPlayersStillInTheGame(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var gone *game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			gone = p
			break
		}
	}
	if err := g.Concede(gone.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	life := me.Life
	effectstest.P1g2Cast(t, g, "Benediction of Moons", "Sorcery", oldB569BenedictionOracle, "{W}", game.CastSpellParams{})
	oldB569Settle(t, g, oldB569NoPick)
	if want := len(g.Seats) - 1; me.Life != life+want {
		t.Errorf("life %d→%d, want +%d with one player gone", life, me.Life, want)
	}
}

func TestOldB569HauntSpellWithNoCreatureStaysInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.P1g2Cast(t, g, "Benediction of Moons", "Sorcery", oldB569BenedictionOracle, "{W}", game.CastSpellParams{})
	oldB569Settle(t, g, oldB569NoPick)
	if !me.Graveyard.Contains(id) {
		t.Error("with no creature to haunt, the card left its owner's graveyard")
	}
}

// --- Seize the Soul ------------------------------------------------------

func TestOldB569SeizeTheSoulDestroysMakesASpiritHauntsAndDoesItAgain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	victim := effectstest.DftABPush(g, opp.ID, "Opp Bear", "Creature — Bear", "", 2, 2)
	host := effectstest.DftABPush(g, opp.ID, "Opp Elk", "Creature — Elk", "", 2, 2)
	second := effectstest.DftABPush(g, opp.ID, "Opp Wolf", "Creature — Wolf", "", 2, 2)

	id := effectstest.P1g2Cast(t, g, "Seize the Soul", "Instant", oldB569SeizeOracle, "{2}{B}{B}",
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
	oldB569Settle(t, g, oldB569Card(host))
	if !opp.Graveyard.Contains(victim) {
		t.Fatal("the spell did not destroy its target")
	}
	if n := oldB569Spirits(g, me.ID); n != 1 {
		t.Fatalf("the spell made %d Spirits, want 1", n)
	}
	if haunted, ok := oldB569Haunting(g, id); !ok || haunted != host {
		t.Fatalf("after resolving: in exile %v, haunting %v; want haunting the elk", ok, haunted)
	}

	oldB569Kill(t, g, opp.ID, host)
	oldB569Settle(t, g, oldB569Card(second))
	if !opp.Graveyard.Contains(second) {
		t.Error("the haunt trigger did not destroy its target")
	}
	if n := oldB569Spirits(g, me.ID); n != 2 {
		t.Errorf("after the haunt trigger: %d Spirits, want 2", n)
	}
}

// A spell removed for want of legal targets never resolved, so it neither
// makes the Spirit nor haunts (CR 608.2b, CR 702.55a).
func TestOldB569SeizeTheSoulWithItsTargetGoneDoesNotHaunt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	victim := effectstest.DftABPush(g, opp.ID, "Opp Bear", "Creature — Bear", "", 2, 2)
	effectstest.DftABPush(g, opp.ID, "Opp Elk", "Creature — Elk", "", 2, 2)

	id := effectstest.P1g2Cast(t, g, "Seize the Soul", "Instant", oldB569SeizeOracle, "{2}{B}{B}",
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
	oldB569Kill(t, g, opp.ID, victim)
	oldB569Settle(t, g, oldB569NoPick)
	if n := oldB569Spirits(g, me.ID); n != 0 {
		t.Errorf("a fizzled Seize the Soul made %d Spirits", n)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("a fizzled Seize the Soul is not in its owner's graveyard")
	}
}

// --- Cry of Contrition ---------------------------------------------------

func TestOldB569CryOfContritionDiscardsHauntsAndDiscardsAgain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.P1g2Seats(g)
	host := effectstest.DftABPush(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	hand := len(opp.Hand.Cards)

	id := effectstest.P1g2Cast(t, g, "Cry of Contrition", "Sorcery", oldB569CryOracle, "{B}",
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	oldB569Settle(t, g, oldB569NoPick)
	effectstest.DiscardFromHand(t, g, opp.ID)
	oldB569Settle(t, g, oldB569Card(host))
	if len(opp.Hand.Cards) != hand-1 {
		t.Fatalf("the spell: hand %d→%d, want one discarded", hand, len(opp.Hand.Cards))
	}
	if haunted, ok := oldB569Haunting(g, id); !ok || haunted != host {
		t.Fatalf("after resolving: in exile %v, haunting %v; want haunting my bear", ok, haunted)
	}

	oldB569Kill(t, g, me.ID, host)
	oldB569Settle(t, g, oldB569Player(opp.ID))
	effectstest.DiscardFromHand(t, g, opp.ID)
	oldB569Settle(t, g, oldB569NoPick)
	if len(opp.Hand.Cards) != hand-2 {
		t.Errorf("after the haunt trigger: hand %d→%d, want two discarded", hand, len(opp.Hand.Cards))
	}
}
