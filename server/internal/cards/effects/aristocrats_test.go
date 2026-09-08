package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aristocrats_test.go — S21 sub-PR 3: the payoffs, and the sprint's
// exit criteria end to end.

const (
	bloodArtistOracle       = "310f141c-7f37-4729-aed6-dd9c09db448d"
	zulaportCutthroatOracle = "76b003e0-15af-4f22-bdf2-1ade5430964a"
	mayhemDevilOracle       = "4709f11c-aef8-45ac-b2bf-e640c568dfac"
	midnightReaperOracle    = "e8c7566d-7cc0-48af-a986-83223ec7e06c"
)

// pickPlayer answers the latest pick_target prompt with a player.
func pickPlayer(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: target}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// --- Blood Artist ------------------------------------------------

// Blood Artist triggers on ANY creature's death, including its own
// and including creatures it doesn't control.
func TestBloodArtistDrainsOnAnyCreatureDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	theirs := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	meBefore, oppBefore := me.Life, opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	// The trigger needs a target before it can go on the stack.
	for i := 0; i < 6 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	if opp.Life != oppBefore-1 {
		t.Errorf("target lost %d life, want 1", oppBefore-opp.Life)
	}
	if me.Life != meBefore+1 {
		t.Errorf("controller gained %d life, want 1", me.Life-meBefore)
	}
}

func TestBloodArtistSeesItsOwnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	artist := pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	oppBefore := opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(artist) })
	for i := 0; i < 6 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppBefore-1 {
		t.Errorf("Blood Artist should drain for its own death: %d -> %d", oppBefore, opp.Life)
	}
}

// --- Zulaport Cutthroat ------------------------------------------

func TestZulaportHitsEveryOpponentButOnlyForYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
	mine := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	theirs := pushCatalogPermanent(g, g.Seats[1].ID, "Theirs", "Creature — Bear", "", false)
	meBefore := me.Life
	oppBefore := []int{g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}

	// An opponent's creature dying does nothing.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if me.Life != meBefore {
		t.Errorf("an opponent's creature must not trigger Zulaport")
	}

	// Mine does, and drains the whole table.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	passPriorityAroundTable(t, g)
	for i, before := range oppBefore {
		seat := g.Seats[i+1]
		if seat.Life != before-1 {
			t.Errorf("seat %d: %d -> %d, want -1", i+1, before, seat.Life)
		}
	}
	if me.Life != meBefore+1 {
		t.Errorf("controller gains exactly 1 regardless of table size: %d -> %d", meBefore, me.Life)
	}
}

// --- Mayhem Devil ------------------------------------------------

// Sacrifice, not death: a destroyed creature doesn't trigger it, a
// sacrificed one does — and so does a non-creature permanent.
func TestMayhemDevilTriggersOnSacrificeNotDestruction(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Mayhem Devil", "Creature — Devil", mayhemDevilOracle, false)
	destroyed := pushCatalogPermanent(g, me.ID, "A", "Creature — Goblin", "", false)
	sacrificed := pushCatalogPermanent(g, me.ID, "B", "Creature — Goblin", "", false)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	oppBefore := opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(destroyed) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatalf("destruction must not trigger Mayhem Devil")
	}

	if err := g.SacrificePermanent(me.ID, sacrificed); err != nil {
		t.Fatal(err)
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppBefore-1 {
		t.Errorf("sacrificing a creature: %d -> %d, want -1", oppBefore, opp.Life)
	}

	// A non-creature permanent counts too.
	if err := g.SacrificePermanent(me.ID, rock); err != nil {
		t.Fatal(err)
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppBefore-2 {
		t.Errorf("sacrificing an artifact should ping too: want -2 total, got %d", oppBefore-opp.Life)
	}
}

// An opponent's sacrifice triggers it as well ("whenever A PLAYER
// sacrifices").
func TestMayhemDevilTriggersOnOpponentSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Mayhem Devil", "Creature — Devil", mayhemDevilOracle, false)
	theirs := pushCatalogPermanent(g, opp.ID, "Theirs", "Creature — Bear", "", false)
	oppBefore := opp.Life

	if err := g.SacrificePermanent(opp.ID, theirs); err != nil {
		t.Fatal(err)
	}
	// The Devil's controller chooses the target, not the sacrificer.
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppBefore-1 {
		t.Errorf("an opponent's sacrifice should still ping: %d -> %d", oppBefore, opp.Life)
	}
}

// --- Midnight Reaper ---------------------------------------------

func TestMidnightReaperIgnoresTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Midnight Reaper", "Creature — Zombie Knight", midnightReaperOracle, false)
	token := pushCatalogPermanent(g, me.ID, "Goblin", "Token Creature — Goblin", "", false)
	real := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	lifeBefore, handBefore := me.Life, me.Hand.Size()

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	passPriorityAroundTable(t, g)
	if me.Life != lifeBefore || me.Hand.Size() != handBefore {
		t.Errorf("a token death must not trigger the Reaper: life %d, hand %d",
			me.Life, me.Hand.Size())
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(real) })
	passPriorityAroundTable(t, g)
	if me.Life != lifeBefore-1 {
		t.Errorf("nontoken death: life %d -> %d, want -1", lifeBefore, me.Life)
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("nontoken death: hand %d -> %d, want +1", handBefore, me.Hand.Size())
	}
}

// --- S21 exit criteria -------------------------------------------

// "Cast Goblin Bombardment + Blood Artist + Krenko, Mob Boss;
// sacrifice tokens to Bombardment one at a time → opponent's life
// ticks down (Blood Artist + Bombardment damage); your life ticks
// up (Blood Artist gain)."
//
// This is the whole S21 arc in one test: activated abilities with a
// sacrifice cost (sub-PR 2), sacrifice as a distinct game action
// (sub-PR 1), token creation, and the payoffs. It also pins the
// ordering that makes the archetype work — the cost's dies-trigger
// resolves before the ability that caused it.
func TestS21ExitCriteriaAristocratsCombo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bomb := pushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, false)
	pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	krenko := pushCatalogPermanent(g, me.ID, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", krenkoOracle, false)

	// Krenko taps for Goblins: himself + Blood Artist isn't a Goblin,
	// so X = 1 (Krenko alone).
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Krenko: %v", err)
	}
	passPriorityAroundTable(t, g)
	goblins := goblinTokenIDs(g, me.ID)
	if len(goblins) != 1 {
		t.Fatalf("Krenko made %d Goblins, want 1", len(goblins))
	}

	meBefore, oppBefore := me.Life, opp.Life

	// Feed the Goblin to the Bombardment, aimed at the opponent.
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{goblins[0]},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("Bombardment: %v", err)
	}
	// Paying the cost killed the Goblin, so Blood Artist's trigger is
	// waiting for its target — above the Bombardment ability.
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	// Blood Artist drained 1 and the Bombardment dealt 1.
	if opp.Life != oppBefore-2 {
		t.Errorf("opponent %d -> %d, want -2 (Artist drain + Bombardment ping)", oppBefore, opp.Life)
	}
	if me.Life != meBefore+1 {
		t.Errorf("caster %d -> %d, want +1 (Artist gain)", meBefore, me.Life)
	}
	if g.Battlefield.Contains(goblins[0]) {
		t.Errorf("the sacrificed Goblin should be gone")
	}
	if !g.Battlefield.Contains(bomb) {
		t.Errorf("the Bombardment itself is not consumed")
	}
}

func goblinTokenIDs(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Goblin" && c.Controller == controller {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
