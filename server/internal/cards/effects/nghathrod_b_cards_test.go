package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4092 slice b — the Captain N'ghathrod deck's Horrors: Aboleth Spawn,
// Brainstealer Dragon, Mind Flayer, Hunted Horror and Sewer Nemesis.

const (
	nghbAbolethSpawnOracle       = "d6ae77b3-59e7-475c-800c-930a5cb2e0ca"
	nghbBrainstealerDragonOracle = "13cb958f-1e66-4219-901a-ff58275c8475"
	nghbMindFlayerOracle         = "8dd473c7-afd9-40f1-bb8e-0a00f4687318"
	nghbHuntedHorrorOracle       = "3469463e-ed4d-44bb-81f3-42bbe09d34ad"
	nghbSewerNemesisOracle       = "c7f1a49a-cc22-432f-bf3e-50ba61982b87"
)

// nghbEnter puts a permanent onto the battlefield under `controller`
// (owned by `owner`) and announces its entry, then runs the boundary
// that puts what it triggered on the stack.
func nghbEnter(g *game.Game, owner, controller uuid.UUID, name, typeLine, oracle, manaCost string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Power: 2, Toughness: 2, Owner: owner, Controller: controller,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: controller})
	})
	g.RunStateChecksForTest()
	return id
}

// nghbItems lists every stack item whose label starts with prefix.
func nghbItems(g *game.Game, prefix string) []*game.StackItem {
	var out []*game.StackItem
	g.WithWriteLock(func() {
		for _, it := range g.StackMeta {
			if it != nil && strings.HasPrefix(it.Label, prefix) {
				out = append(out, it)
			}
		}
		for _, it := range g.PendingTriggers {
			if it != nil && strings.HasPrefix(it.Label, prefix) {
				out = append(out, it)
			}
		}
	})
	return out
}

const nghbProbingTelepathy = "Aboleth Spawn — Probing Telepathy"

// --- Aboleth Spawn ----------------------------------------------------

// The 2022-06-10 ruling: the Spawn's ability always goes on the stack
// on top of the entering creature's ability, even on the Spawn's
// controller's own turn, where APNAP alone would have put it below. The
// copy is the Spawn controller's and resolves first.
func TestAbolethSpawnCopiesAnOpponentsEntryTriggerFromAbove(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: seat 0 must be active, got %d", g.Turn.ActiveSeat)
	}
	pushCatalogPermanent(g, me.ID, "Aboleth Spawn", "Creature — Fish Horror", nghbAbolethSpawnOracle, false)
	nghbEnter(g, opp.ID, opp.ID, "Mulldrifter", "Creature — Elemental", mulldrifterOracle, "{4}{U}")

	drift := nghbItems(g, "Mulldrifter")
	spawn := nghbItems(g, nghbProbingTelepathy)
	if len(drift) != 1 || len(spawn) != 1 {
		t.Fatalf("want one Mulldrifter trigger and one Spawn trigger, got %d and %d", len(drift), len(spawn))
	}
	if g.StackMeta[spawn[0].ID] == nil || g.StackMeta[drift[0].ID] == nil {
		t.Fatal("both triggers should be on the stack")
	}
	if spawn[0].Seq <= drift[0].Seq {
		t.Fatalf("the Spawn's trigger (seq %d) is not above the Mulldrifter trigger (seq %d)", spawn[0].Seq, drift[0].Seq)
	}
	if spawn[0].Controller != me.ID {
		t.Fatalf("the Spawn's trigger is controlled by %v, want %v", spawn[0].Controller, me.ID)
	}

	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()
	answerLatestTriggerPrompt(t, g, me.ID, true)
	cp := acCopyOnStack(g)
	if cp == nil || cp.Controller != me.ID {
		t.Fatalf("the copy should be on the stack under the Spawn's controller: %+v", cp)
	}
	if cp.Seq <= drift[0].Seq {
		t.Error("the copy should resolve before the ability it copies")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - myHand; got != 2 {
		t.Errorf("the Spawn's controller drew %d from the copy, want 2", got)
	}
	if got := opp.Hand.Size() - theirHand; got != 2 {
		t.Errorf("the opponent drew %d from their own trigger, want 2", got)
	}
}

func TestAbolethSpawnDeclinedCopiesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Aboleth Spawn", "Creature — Fish Horror", nghbAbolethSpawnOracle, false)
	nghbEnter(g, opp.ID, opp.ID, "Mulldrifter", "Creature — Elemental", mulldrifterOracle, "{4}{U}")
	myHand := me.Hand.Size()
	answerLatestTriggerPrompt(t, g, me.ID, false)
	if acCopiesOnStack(g) != 0 {
		t.Fatal("declining made a copy")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != myHand {
		t.Errorf("declining still drew %d cards", me.Hand.Size()-myHand)
	}
}

// Only a triggered ability OF the entering creature, and only for a
// creature entering under an opponent's control: your own creature, and
// an opponent's other permanent that sees the creature enter, are not
// copied.
func TestAbolethSpawnIgnoresYourCreaturesAndOtherPermanentsTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Aboleth Spawn", "Creature — Fish Horror", nghbAbolethSpawnOracle, false)
	pushCatalogPermanent(g, opp.ID, "Impact Tremors", "Enchantment", impactTremorsOracle, false)

	nghbEnter(g, opp.ID, opp.ID, "Bear", "Creature — Bear", "", "{1}{G}")
	if n := len(nghbItems(g, "Impact Tremors")); n != 1 {
		t.Fatalf("setup: Impact Tremors should have triggered once, got %d", n)
	}
	if n := len(nghbItems(g, nghbProbingTelepathy)); n != 0 {
		t.Fatalf("another permanent's trigger was offered to the Spawn (%d)", n)
	}
	passPriorityAroundTable(t, g)

	nghbEnter(g, me.ID, me.ID, "Mulldrifter", "Creature — Elemental", mulldrifterOracle, "{4}{U}")
	if n := len(nghbItems(g, nghbProbingTelepathy)); n != 0 {
		t.Fatalf("your own creature's trigger was offered to the Spawn (%d)", n)
	}
}

// "You may choose new targets for the copy" (CR 707.10c): the copy of
// an opponent's Ravenous Chupacabra is the Spawn controller's, so it is
// aimed at a creature an opponent of THEIRS controls.
func TestAbolethSpawnCopyMayChooseNewTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Aboleth Spawn", "Creature — Fish Horror", nghbAbolethSpawnOracle, false)
	myBear := ctrlPushCreature(g, me.ID, "My Bear")
	theirBear := ctrlPushCreature(g, opp.ID, "Their Bear")

	nghbEnter(g, opp.ID, opp.ID, "Ravenous Chupacabra", "Creature — Horror", ravenousChupacabraOracle, "{2}{B}{B}")
	pickCard(t, g, opp.ID, myBear)
	g.RunStateChecksForTest()
	if n := len(nghbItems(g, nghbProbingTelepathy)); n != 1 {
		t.Fatalf("a targeted entry trigger should trigger the Spawn once, got %d", n)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, theirBear)
	passPriorityAroundTable(t, g)
	if e2Zone(g, myBear) == game.ZoneBattlefield {
		t.Error("the opponent's own trigger should still destroy its target")
	}
	if e2Zone(g, theirBear) == game.ZoneBattlefield {
		t.Error("the copy, retargeted, should destroy the opponent's creature")
	}
}

// The seam itself: one EventTriggerPlaced per triggered ability the
// drain puts on the stack, naming the item, while something watches the
// kind; and none at all on a table where nothing does, so the event log
// (and every frozen snapshot of one) is what it was.
func TestTriggerPlacedNamesEveryPlacedTriggerOnlyWhileWatched(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nghbEnter(g, opp.ID, opp.ID, "Mulldrifter", "Creature — Elemental", mulldrifterOracle, "{4}{U}")
	if n := acCountEvents(g, game.EventTriggerPlaced); n != 0 {
		t.Fatalf("%d trigger_placed events with nothing watching, want 0", n)
	}
	passPriorityAroundTable(t, g)

	pushCatalogPermanent(g, me.ID, "Aboleth Spawn", "Creature — Fish Horror", nghbAbolethSpawnOracle, false)
	nghbEnter(g, opp.ID, opp.ID, "Mulldrifter", "Creature — Elemental", mulldrifterOracle, "{4}{U}")
	drift := nghbItems(g, "Mulldrifter")
	spawn := nghbItems(g, nghbProbingTelepathy)
	if len(drift) != 1 || len(spawn) != 1 {
		t.Fatalf("want one Mulldrifter and one Spawn trigger, got %d and %d", len(drift), len(spawn))
	}
	per := map[uuid.UUID]int{}
	for _, ev := range g.Events {
		if ev.Kind == game.EventTriggerPlaced {
			per[ev.StackItemID]++
			if ev.StackItemID == drift[0].ID && ev.Actor != opp.ID {
				t.Errorf("event %+v names the wrong controller", ev)
			}
		}
	}
	if per[drift[0].ID] != 1 || per[spawn[0].ID] != 1 || len(per) != 2 {
		t.Errorf("trigger_placed per item %v, want exactly one for each of the two placed triggers", per)
	}
}

// --- Mind Flayer -------------------------------------------------------

func TestMindFlayerStealsForAsLongAsYouControlIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := ctrlPushCreature(g, opp.ID, "Bear")
	flayer := castAndResolveCreature(t, g, "Mind Flayer", "Creature — Horror", nghbMindFlayerOracle)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %v after Mind Flayer entered, want %v", got, me.ID)
	}
	// Losing control of the Flayer, not only losing the Flayer, ends it.
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), flayer, opp.ID, game.IndefiniteDuration(), "test — take the Flayer")
	})
	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Errorf("the Flayer changed hands and the creature stayed: controller %v, want %v", got, opp.ID)
	}
}

func TestMindFlayerGivesItBackWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := ctrlPushCreature(g, opp.ID, "Bear")
	flayer := castAndResolveCreature(t, g, "Mind Flayer", "Creature — Horror", nghbMindFlayerOracle)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(flayer) })
	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Errorf("the Flayer died and the creature stayed: controller %v, want %v", got, opp.ID)
	}
}

// --- Hunted Horror -----------------------------------------------------

func TestHuntedHorrorGivesTheOpponentTwoProtectedCentaurs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castAndResolveCreature(t, g, "Hunted Horror", "Creature — Horror", nghbHuntedHorrorOracle)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	centaurs := battlefieldIDsNamed(g, "Centaur")
	if len(centaurs) != 2 {
		t.Fatalf("%d Centaurs, want 2", len(centaurs))
	}
	for _, id := range centaurs {
		if got := controllerOf(t, g, id); got != opp.ID {
			t.Errorf("Centaur controlled by %v, want the target opponent %v", got, opp.ID)
		}
		if p, tough := effectivePTOf(t, g, id); p != 3 || tough != 3 {
			t.Errorf("Centaur is %d/%d, want 3/3", p, tough)
		}
		if !effectiveAbilitiesContain(t, g, id, "protection from black") {
			t.Errorf("Centaur abilities %v lack protection from black", effectiveAbilities(t, g, id))
		}
	}
}

func TestHuntedHorrorCannotTargetItsController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Hunted Horror", "Creature — Horror", nghbHuntedHorrorOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Error("the Horror's controller is not a target opponent")
	}
}

// --- Sewer Nemesis -----------------------------------------------------

func TestSewerNemesisIsAsBigAsTheChosenGraveyardAndMillsItsCaster(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	for i := 0; i < 3; i++ {
		topOfLibrary(opp, "Their Card", "Instant", "{U}")
		topOfLibrary(other, "Other Card", "Instant", "{U}")
	}
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(opp.ID, 2); err != nil {
			t.Fatalf("mill: %v", err)
		}
		if err := g.MillNForEffect(other.ID, 3); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	id := castCatalogSpell(t, g, "Sewer Nemesis", "Creature — Horror", nghbSewerNemesisOracle, nil)
	passPriorityAroundTable(t, g)
	// A state check with the choice still open must not find a 0/0:
	// an unchosen player is not an empty graveyard.
	g.RunStateChecksForTest()
	if e2Zone(g, id) != game.ZoneBattlefield {
		t.Fatal("the Nemesis died before its controller chose a player")
	}
	answerChosenPlayer(t, g, me.ID, opp.ID)
	if p, tough := effectivePTOf(t, g, id); p != 2 || tough != 2 {
		t.Fatalf("P/T %d/%d, want 2/2 (the chosen player's graveyard)", p, tough)
	}

	// The chosen player casts: they mill one, and the Nemesis grows.
	topOfLibrary(opp, "Milled", "Instant", "{U}")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventCast, Actor: opp.ID, CardID: uuid.New()})
	})
	passPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size(); got != 3 {
		t.Errorf("chosen player's graveyard %d, want 3 after the mill", got)
	}
	if p, _ := effectivePTOf(t, g, id); p != 3 {
		t.Errorf("power %d after the mill, want 3", p)
	}

	// Anyone else casting does nothing.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventCast, Actor: other.ID, CardID: uuid.New()})
	})
	passPriorityAroundTable(t, g)
	if got := other.Graveyard.Size(); got != 3 {
		t.Errorf("an unchosen caster's graveyard is %d, want 3 (no mill)", got)
	}
}

// --- Brainstealer Dragon -----------------------------------------------

func TestBrainstealerDragonExilesEachOpponentsTopCardToPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Brainstealer Dragon", "Creature — Dragon Horror", nghbBrainstealerDragonOracle, false)
	tops := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats[1:] {
		tops[p.ID] = topOfLibrary(p, "Their Land", "Land", "")
	}
	mine := topOfLibrary(me, "My Card", "Instant", "{U}")
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for seat, top := range tops {
		if z := e2Zone(g, top); z != game.ZoneExile {
			t.Errorf("seat %v's top card is in %q, want exile", seat, z)
			continue
		}
		perm := e2Permission(g, top)
		if perm == nil || perm.Player != me.ID || perm.CastOnly || !perm.AnyColor || perm.AnyType {
			t.Errorf("permission %+v, want a play grant with any-colour mana to the Dragon's controller", perm)
			continue
		}
		if perm.Duration.Kind != game.WhileInZone {
			t.Errorf("duration %+v, want for as long as it remains exiled", perm.Duration)
		}
	}
	if z := e2Zone(g, mine); z == game.ZoneExile {
		t.Error("the Dragon exiled its controller's own card")
	}
}

func TestBrainstealerDragonDrainsTheOwnerOfAStolenEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Brainstealer Dragon", "Creature — Dragon Horror", nghbBrainstealerDragonOracle, false)
	life := opp.Life
	nghbEnter(g, opp.ID, me.ID, "Their Ogre", "Creature — Ogre", "", "{3}{B}{B}")
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 5 {
		t.Fatalf("the owner lost %d life, want 5 (the mana value)", got)
	}

	// A land they own, and a permanent you own, do nothing.
	life, myLife := opp.Life, me.Life
	nghbEnter(g, opp.ID, me.ID, "Their Swamp", "Basic Land — Swamp", "", "")
	if n := len(nghbItems(g, "Brainstealer Dragon — its owner")); n != 0 {
		t.Errorf("a land entering triggered the Dragon (%d)", n)
	}
	nghbEnter(g, me.ID, me.ID, "My Ogre", "Creature — Ogre", "", "{3}{B}{B}")
	passPriorityAroundTable(t, g)
	if opp.Life != life || me.Life != myLife {
		t.Errorf("life moved on a land or an own permanent: opp %d→%d, me %d→%d", life, opp.Life, myLife, me.Life)
	}
	// A permanent the opponent owns entering under THEIR control does
	// nothing either.
	nghbEnter(g, opp.ID, opp.ID, "Their Other Ogre", "Creature — Ogre", "", "{3}{B}{B}")
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("an opponent's own entry drained them: %d→%d", life, opp.Life)
	}
}
