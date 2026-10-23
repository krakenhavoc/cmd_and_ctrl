package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// world_enchantments_1997_test.go — #1997: the world enchantments ADR
// 0109 PR 7 left out, each with the one shape of its own it needed.

const (
	tombstoneStairwellOracle = "9aaf1149-eccb-434a-a1a7-14b93fa50141"
	fieldOfDreamsOracle      = "6dcc62f8-2fd9-473b-9e6a-2dd7e610f855"
	koskunFallsOracle        = "a9b9e798-debc-4f2c-9770-e16f670082f7"
	elkinLairOracle          = "b3488947-7142-4e6c-9c92-2728469b3ad9"
	wintersNightOracle       = "e4f73802-b639-47ab-a6c6-f761afa4e00e"
	hallOfGemstoneOracle     = "3da9fecc-064a-44e6-b88e-4b10194e367e"
)

// tombspawnOf lists the Tombspawn tokens `player` controls.
func tombspawnOf(g *game.Game, player uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Tombspawn" && c.Controller == player {
			out = append(out, c)
		}
	}
	return out
}

func tombspawnTotal(g *game.Game) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Tombspawn" {
			n++
		}
	}
	return n
}

// The upkeep trigger fires on another player's upkeep too, and each
// player makes one hasty 2/2 black Tombspawn for each CREATURE card in
// their own graveyard, under their own control, linked to the Stairwell.
func TestTombstoneStairwellMakesTombspawnForEachCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	owner, a, b := g.Seats[3], g.Seats[0], g.Seats[2]
	stair := pushPermanentForTest(g, owner.ID, "Tombstone Stairwell", tombstoneStairwellOracle, "World Enchantment")

	advanceToUpkeepOf(t, g, 2)
	// Seeded once the upkeep has begun: the earlier upkeeps fire too.
	pushGraveyardCardTyped(a, "Bear", "Creature — Bear")
	pushGraveyardCardTyped(a, "Elf", "Creature — Elf")
	pushGraveyardCardTyped(b, "Bear", "Creature — Bear")
	pushGraveyardCardTyped(b, "Shock", "Instant")
	if triggerOnStack(g, stair) == nil {
		t.Fatal("the upkeep trigger did not fire on another player's upkeep")
	}
	passPriorityAroundTable(t, g)

	if got := len(tombspawnOf(g, a.ID)); got != 2 {
		t.Errorf("seat 0 has %d Tombspawn, want 2 (two creature cards)", got)
	}
	if got := len(tombspawnOf(g, b.ID)); got != 1 {
		t.Errorf("seat 2 has %d Tombspawn, want 1 (the instant does not count)", got)
	}
	if got := len(tombspawnOf(g, owner.ID)); got != 0 {
		t.Errorf("the controller, with an empty graveyard, has %d Tombspawn", got)
	}
	var ref game.PermissionCardRef
	g.ReadSnapshot(func() {
		r, ok := g.PermanentRefForEffect(stair)
		if !ok {
			t.Fatal("the Stairwell left")
		}
		ref = game.PermissionCardRef(r)
	})
	for _, c := range tombspawnOf(g, a.ID) {
		if !c.IsCreature() || !c.HasSubtype("Zombie") || c.Power != 2 || c.Toughness != 2 ||
			len(c.Colors) != 1 || c.Colors[0] != "B" || !wpHas(g, c, "haste") {
			t.Errorf("a Tombspawn is not a 2/2 black Zombie with haste: %+v", c)
		}
		if c.CreatedWith != ref {
			t.Errorf("a Tombspawn is linked to %+v, want the Stairwell %+v", c.CreatedWith, ref)
		}
	}
}

// At the beginning of each end step, every token created with it is
// destroyed and can't be regenerated, whoever controls it now; a token
// it did not create stays. The link survives a restore.
func TestTombstoneStairwellDestroysItsTokensAtEachEndStepAcrossARestore(t *testing.T) {
	g := newCatalogGame(t)
	owner, a := g.Seats[3], g.Seats[0]
	stair := pushPermanentForTest(g, owner.ID, "Tombstone Stairwell", tombstoneStairwellOracle, "World Enchantment")
	other := b12Creature(g, a.ID, "Tombspawn", "Token Creature — Zombie", 2, 2)

	advanceToUpkeepOf(t, g, 2)
	// Seeded once the upkeep has begun: the earlier upkeeps fire too.
	pushGraveyardCardTyped(a, "Bear", "Creature — Bear")
	pushGraveyardCardTyped(a, "Elf", "Creature — Elf")
	passPriorityAroundTable(t, g)
	spawn := tombspawnOf(g, a.ID)
	if len(spawn) != 3 { // two made, plus the unlinked one
		t.Fatalf("seat 0 has %d Tombspawn, want 3", len(spawn))
	}
	// One of them changes control and grows a regeneration shield.
	var moved uuid.UUID
	for _, c := range spawn {
		if c.InstanceID != other {
			moved = c.InstanceID
			break
		}
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if c := &g.Battlefield.Cards[i]; c.InstanceID == moved {
				c.Controller = owner.ID
				c.RegenerationShields = 1
			}
		}
	})

	restored := restoreRoundTrip(t, g, true)
	advanceToEndStepOf(t, restored, 2)
	rstair := findBattlefieldByName(restored, "Tombstone Stairwell")
	if rstair != stair || triggerOnStack(restored, stair) == nil {
		t.Fatal("the end-step trigger did not fire after the restore")
	}
	passPriorityAroundTable(t, restored)
	if n := tombspawnTotal(restored); n != 1 {
		t.Fatalf("%d Tombspawn left, want only the one it did not create", n)
	}
	if _, ok := battlefieldCard(restored, other); !ok {
		t.Error("a token the Stairwell did not create was destroyed")
	}
	if _, ok := battlefieldCard(restored, moved); ok {
		t.Error("a linked token under another player's control, with a regeneration shield, survived")
	}
}

// When it leaves the battlefield, its leaves trigger destroys its
// tokens, read by the object that left.
func TestTombstoneStairwellDestroysItsTokensWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	owner, a := g.Seats[3], g.Seats[0]
	stair := pushPermanentForTest(g, owner.ID, "Tombstone Stairwell", tombstoneStairwellOracle, "World Enchantment")

	advanceToUpkeepOf(t, g, 2)
	// Seeded once the upkeep has begun: the earlier upkeeps fire too.
	pushGraveyardCardTyped(a, "Bear", "Creature — Bear")
	passPriorityAroundTable(t, g)
	if tombspawnTotal(g) != 1 {
		t.Fatalf("%d Tombspawn, want 1", tombspawnTotal(g))
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(stair) })
	passPriorityAroundTable(t, g)
	if n := tombspawnTotal(g); n != 0 {
		t.Errorf("%d Tombspawn survived the Stairwell leaving", n)
	}
}

// The "if this enchantment is on the battlefield" is checked as the
// trigger resolves (CR 603.4): gone by then, nobody gets a Tombspawn.
func TestTombstoneStairwellMakesNothingOnceItHasLeft(t *testing.T) {
	g := newCatalogGame(t)
	owner, a := g.Seats[3], g.Seats[0]
	stair := pushPermanentForTest(g, owner.ID, "Tombstone Stairwell", tombstoneStairwellOracle, "World Enchantment")

	advanceToUpkeepOf(t, g, 2)
	// Seeded once the upkeep has begun: the earlier upkeeps fire too.
	pushGraveyardCardTyped(a, "Bear", "Creature — Bear")
	if triggerOnStack(g, stair) == nil {
		t.Fatal("the upkeep trigger did not fire")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(stair) })
	passPriorityAroundTable(t, g)
	if n := tombspawnTotal(g); n != 0 {
		t.Errorf("%d Tombspawn made by a Stairwell that had left", n)
	}
}

// --- Koskun Falls ---------------------------------------------------

// koskunPrompt advances to the controller's upkeep, resolves the trigger
// and returns its pay-unless prompt.
func koskunPrompt(t *testing.T, g *game.Game, seat int) *game.PendingChoice {
	t.Helper()
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	return remoraUpkeepPrompt(t, g, g.Seats[seat].ID)
}

// The payment is a tap: the prompt offers only the untapped creatures the
// controller controls, and tapping one keeps Koskun Falls.
func TestKoskunFallsIsKeptByTappingAnUntappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	falls := pushPermanentForTest(g, me.ID, "Koskun Falls", koskunFallsOracle, "World Enchantment")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	tapped := b12Creature(g, me.ID, "Tired Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")

	advanceToUpkeepOf(t, g, 1)
	// Tapped once the untap step is over.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tapped {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	passPriorityAroundTable(t, g)
	choice := remoraUpkeepPrompt(t, g, me.ID)
	var opts []uuid.UUID
	g.ReadSnapshot(func() { opts = g.PayActionOptionsForEffect(me.ID, choice.PayAction()) })
	if len(opts) != 1 || opts[0] != bear {
		t.Fatalf("the payment offers %v, want only the untapped Bear %s", opts, bear)
	}
	if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, true, []uuid.UUID{tapped}); err == nil {
		t.Fatal("a tapped creature was accepted as the payment")
	}
	if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, true, []uuid.UUID{bear}); err != nil {
		t.Fatalf("tap the Bear: %v", err)
	}
	if c, ok := battlefieldCard(g, bear); !ok || !c.Tapped {
		t.Error("the Bear was not tapped as the payment")
	}
	if _, ok := battlefieldCard(g, falls); !ok {
		t.Error("Koskun Falls was sacrificed although the payment was made")
	}
}

// Declining sacrifices it, and so does having no untapped creature: there
// is nothing to pay with (CR 118.3).
func TestKoskunFallsIsSacrificedWithoutATap(t *testing.T) {
	for _, tc := range []struct {
		name     string
		creature bool
	}{
		{"declined", true},
		{"no untapped creature", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[1]
			falls := pushPermanentForTest(g, me.ID, "Koskun Falls", koskunFallsOracle, "World Enchantment")
			if tc.creature {
				b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			}
			choice := koskunPrompt(t, g, 1)
			if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, false, nil); err != nil {
				t.Fatalf("decline: %v", err)
			}
			if _, ok := battlefieldCard(g, falls); ok {
				t.Error("Koskun Falls survived an unpaid upkeep")
			}
		})
	}
}

// The upkeep trigger waiting on the stack is a restore point, and the
// restored table asks for the same tap.
func TestKoskunFallsUpkeepSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	falls := pushPermanentForTest(g, me.ID, "Koskun Falls", koskunFallsOracle, "World Enchantment")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, falls) == nil {
		t.Fatal("the upkeep trigger did not fire")
	}

	restored := restoreRoundTrip(t, g, true)
	passPriorityAroundTable(t, restored)
	choice := remoraUpkeepPrompt(t, restored, me.ID)
	if a := choice.PayAction(); a == nil || a.Kind != game.PayActionTap || a.Count != 1 {
		t.Fatalf("the restored prompt's payment is %+v, want a tap of one", a)
	}
	if err := restored.ResolvePayUnlessWithCards(choice.ID, me.ID, true, []uuid.UUID{bear}); err != nil {
		t.Fatalf("tap the Bear after the restore: %v", err)
	}
	if c, ok := battlefieldCard(restored, bear); !ok || !c.Tapped {
		t.Error("the Bear was not tapped")
	}
	if _, ok := battlefieldCard(restored, falls); !ok {
		t.Error("Koskun Falls was sacrificed although the payment was made")
	}
}

// --- Elkin Lair -----------------------------------------------------

// On another player's upkeep, that player's one card in hand is exiled
// with a play permission for them this turn; unplayed, it goes to their
// graveyard at the end step. The delayed trigger waits across a restore.
func TestElkinLairExilesACardItsPlayerMayPlayThenBinsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, p := g.Seats[0], g.Seats[2]
	pushPermanentForTest(g, me.ID, "Elkin Lair", elkinLairOracle, "World Enchantment")
	advanceToUpkeepOf(t, g, 2)
	emptyHandToLibrary(g, p)
	bolt := handCardFull(p, "Bolt", "Instant", "{R}", "", nil)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bolt) {
		t.Fatal("the upkeep player's card was not exiled")
	}
	perm := exiledPermission(g, bolt)
	if perm.Player != p.ID || perm.CastOnly || perm.Duration.Kind != game.UntilEndOfTurn {
		t.Errorf("grant %+v: the upkeep player may PLAY it this turn", perm)
	}
	if len(g.DelayedTriggers) == 0 {
		t.Fatal("no delayed trigger was set up for the end step")
	}

	restored := restoreRoundTrip(t, g, true)
	advanceToEndStepOf(t, restored, 2)
	passPriorityAroundTable(t, restored)
	rp := restored.Seats[2]
	if restored.Exile.Contains(bolt) || !rp.Graveyard.Contains(bolt) {
		t.Error("the unplayed card is not in its player's graveyard after the end step")
	}
}

// A card the player played is not put into the graveyard: the land played
// from exile stays on the battlefield through the end step.
func TestElkinLairLeavesAPlayedCardAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, p := g.Seats[1], g.Seats[0]
	pushPermanentForTest(g, me.ID, "Elkin Lair", elkinLairOracle, "World Enchantment")
	advanceToUpkeepOf(t, g, 0)
	emptyHandToLibrary(g, p)
	land := handCardFull(p, "Forest", "Basic Land — Forest", "", "", nil)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(land) {
		t.Fatal("the upkeep player's card was not exiled")
	}
	advanceToMain(t, g)
	if err := g.CastSpell(p.ID, land, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("playing the exiled land: %v", err)
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) || p.Graveyard.Contains(land) {
		t.Error("the played land left the battlefield at the end step")
	}
}

// --- Winter's Night -------------------------------------------------

// A snow land any player taps gives that player one more mana of the type
// it made and then skips its controller's next untap step; a land that is
// not snow is left alone. The marker survives a restore.
func TestWintersNightDoublesASnowLandThatThenStaysTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	b31Push(g, me.ID, "Winter's Night", "World Enchantment", wintersNightOracle, "", 0, 0)
	snow := b31Push(g, them.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "", "", 0, 0)
	plain := b31Push(g, them.ID, "Mountain", "Basic Land — Mountain", "", "", 0, 0)

	tapForMana(t, g, them.ID, snow)
	if got := sortedPool(them); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Fatalf("their pool = %v, want two green from the snow Forest", got)
	}
	tapForMana(t, g, them.ID, plain)
	if got := sortedPool(them); len(got) != 3 {
		t.Errorf("their pool = %v, want one more mana only, from a land that is not snow", got)
	}
	if c, _ := battlefieldCard(g, plain); len(c.NextUntapSkips) != 0 {
		t.Error("a land that is not snow got the untap rider")
	}

	restored := restoreRoundTrip(t, g, true)
	advanceToUpkeepOf(t, restored, 1)
	if c, ok := battlefieldCard(restored, snow); !ok || !c.Tapped {
		t.Error("the snow land untapped during its controller's next untap step")
	}
	if c, ok := battlefieldCard(restored, plain); !ok || c.Tapped {
		t.Error("the other land did not untap")
	}
	if _, err := restored.AdvanceStep(); err != nil {
		t.Fatalf("leave the upkeep: %v", err)
	}
	advanceToUpkeepOf(t, restored, 1)
	if c, ok := battlefieldCard(restored, snow); !ok || c.Tapped {
		t.Error("the snow land did not untap in the untap step after that")
	}
}

// --- Hall of Gemstone -----------------------------------------------

// hallChooses advances to `seat`'s upkeep, resolves the Hall's trigger
// and answers its colour prompt for that seat.
func hallChooses(t *testing.T, g *game.Game, seat int, color string) {
	t.Helper()
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	answerColor(t, g, g.Seats[seat].ID, color)
}

func landManaColorRecords(g *game.Game) int {
	n := 0
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModLandManaIsColor {
				n++
			}
		}
	}
	return n
}

// The upkeep player chooses; for the rest of the turn every land any
// player taps makes that colour instead (colourless and the amount are
// game/land_mana_color_test.go's).
func TestHallOfGemstoneRecoloursEveryLandsMana(t *testing.T) {
	g := newCatalogGame(t)
	owner, me, other := g.Seats[3], g.Seats[1], g.Seats[0]
	pushPermanentForTest(g, owner.ID, "Hall of Gemstone", hallOfGemstoneOracle, "World Enchantment")
	hallChooses(t, g, 1, "R")
	forest := restrictedBasic(t, g, me, "Forest")
	island := restrictedBasic(t, g, other, "Island")

	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("the Forest made %v, want one red", got)
	}
	tapForMana(t, g, other.ID, island)
	if got := sortedPool(other); len(got) != 1 || got[0] != "R" {
		t.Errorf("another player's Island made %v, want one red", got)
	}
}

// Auto-tap books each land as the colour it really makes: under red, a
// Forest and a {W}/{U} dual pay {R}{R}, and they can no longer pay {G}.
// The record is data, so it survives a restore, and it ends with the
// turn.
func TestHallOfGemstoneAutoTapPaysWithTheChosenColour(t *testing.T) {
	g := newCatalogGame(t)
	owner, me := g.Seats[3], g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Hall of Gemstone", hallOfGemstoneOracle, "World Enchantment")
	hallChooses(t, g, 1, "R")
	advanceToMain(t, g)
	forest := restrictedBasic(t, g, me, "Forest")
	dual := seedManaLand(g, me.ID, "Tundra", "Land — Plains Island", "W", "U")
	if landManaColorRecords(g) != 1 {
		t.Fatalf("%d recolouring records, want 1", landManaColorRecords(g))
	}

	restored := restoreRoundTrip(t, g, true)
	rme := restored.Seats[1]
	green, err := game.ParseCost("{G}")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := restored.AutoTapForCost(rme.ID, green, 0); ok {
		t.Error("auto-tap planned {G} from lands that make red")
	}
	castFromHandForTest(t, restored, rme, "Two Goblins", "Creature — Goblin", "{R}{R}", "", autoPay)
	if !permanentTapped(restored, forest) || !permanentTapped(restored, dual) {
		t.Error("auto-tap did not pay {R}{R} with the Forest and the dual")
	}

	advanceToUpkeepOf(t, restored, 2)
	if n := landManaColorRecords(restored); n != 0 {
		t.Errorf("%d recolouring records outlived the turn", n)
	}
}

// --- Field of Dreams ----------------------------------------------

// Every player's top library card is revealed to every player, the
// controller's own and each opponent's, and only while Field of Dreams
// is on the battlefield.
func TestFieldOfDreamsRevealsEveryLibrarysTopCard(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[1], g.Seats[0], g.Seats[2]
	if topVisibleTo(g, a.ID, b.ID) {
		t.Fatal("a library's top card is visible with nothing on the battlefield")
	}
	field := pushPermanentForTest(g, me.ID, "Field of Dreams", fieldOfDreamsOracle, "World Enchantment")
	for _, owner := range g.Seats {
		for _, viewer := range g.Seats {
			if !topVisibleTo(g, owner.ID, viewer.ID) {
				t.Errorf("seat %s cannot see the top of %s's library", viewer.ID, owner.ID)
			}
		}
	}
	removeFromBattlefield(g, field)
	if topVisibleTo(g, a.ID, b.ID) || topVisibleTo(g, me.ID, a.ID) {
		t.Error("the top cards stay revealed once Field of Dreams has left")
	}
}
