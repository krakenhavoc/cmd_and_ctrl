package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// harold_and_bob_test.go — #2809: Harold and Bob, First Numens, the first
// card that returns to the battlefield as an Aura (ADR 0154's 2026-10-23
// amendment). The dies trigger and its intervening-if, CR 303.4f's host
// choice and CR 303.4g's no-host branch, the Aura form's characteristics
// (CR 205.1a, CR 613.1d, CR 613.1f), the grant to the Forest, the
// CR 704.5m re-check of "Forest you control", and a restore point with
// the Aura live.

const haroldAndBobOracle = "d0ecea04-9768-4b09-8718-dd59135251e3"

func hbHarold(g *game.Game, controller uuid.UUID) uuid.UUID {
	return apaPush(g, controller, controller, game.Card{
		Name: "Harold and Bob, First Numens", OracleID: haroldAndBobOracle,
		TypeLine: "Legendary Creature — Treefolk Mutant", ManaCost: "{2}{G}",
		Colors: []string{"G"}, Power: 3, Toughness: 3,
	})
}

func hbForest(g *game.Game, owner, controller uuid.UUID) uuid.UUID {
	return apaPush(g, owner, controller, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
}

// hbDie destroys Harold and Bob and resolves the dies trigger, stopping
// at whatever the entry asks.
func hbDie(t *testing.T, g *game.Game, harold uuid.UUID) {
	t.Helper()
	radDestroy(t, g, harold)
	passPriorityAroundTable(t, g)
}

// The headline: with one Forest to enchant the choice is forced, and the
// card comes back attached to it as a legendary Aura that is not a
// creature and has lost its other abilities.
func TestHaroldAndBobReturnsAsAnAuraOnTheOnlyForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := hbForest(g, me.ID, me.ID)
	harold := hbHarold(g, me.ID)
	if live := gaLayered(t, g, harold); !live.IsCreature() || live.IsAura() || !game.HasKeyword(&live, "reach") {
		t.Fatalf("setup: Harold and Bob should be the printed 3/3 with reach: %+v", live.Effective())
	}

	hbDie(t, g, harold)

	if adZone(g, harold) != game.ZoneBattlefield {
		t.Fatalf("Harold and Bob is in %v, want the battlefield", adZone(g, harold))
	}
	c := gaLayered(t, g, harold)
	if !c.IsAttachedTo(forest) || !c.ReturnedAura {
		t.Fatalf("should enchant the Forest as a returned Aura: attached=%+v returned=%v", c.AttachedTo, c.ReturnedAura)
	}
	if !c.IsAura() || !c.HasCardType("enchantment") || c.IsCreature() {
		t.Errorf("not an Aura enchantment that is no creature: %+v", c.Effective())
	}
	if !c.HasSupertype("legendary") {
		t.Error("lost the legendary supertype; an effect that sets card types does not touch supertypes (CR 205.4b)")
	}
	if c.HasSubtype("Treefolk") || c.HasSubtype("Mutant") {
		t.Error("kept its creature types; setting the card type to Aura enchantment replaces the subtypes (CR 205.1a)")
	}
	if !c.HasLostAllAbilities() || game.HasKeyword(&c, "reach") || game.HasKeyword(&c, "vigilance") {
		t.Errorf("did not lose its other abilities: lost=%v abilities=%+v", c.HasLostAllAbilities(), c.Effective().Abilities)
	}
	if n := countGrantedMana(t, g, forest); n != 1 {
		t.Errorf("the Forest has %d granted mana abilities, want 1", n)
	}
}

func countGrantedMana(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := 0
	_, origins := game.ManaAbilitiesWithOrigins(gaLayered(t, g, id))
	for _, o := range origins {
		if o.Granted() {
			n++
		}
	}
	return n
}

// The granted ability: three mana of ONE colour (one pick), and its
// controller gets two rad counters.
func TestHaroldAndBobsForestTapsForThreeManaAndTwoRadCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := hbForest(g, me.ID, me.ID)
	harold := hbHarold(g, me.ID)
	hbDie(t, g, harold)

	if err := tapGrantedMana(t, g, me.ID, forest, "R"); err != nil {
		t.Fatalf("tap the enchanted Forest: %v", err)
	}
	got := batch01PoolColors(me)
	if len(got) != 3 || got[0] != "R" || got[1] != "R" || got[2] != "R" {
		t.Errorf("pool %v, want three {R}", got)
	}
	if radOf(me) != 2 || radOf(opp) != 0 {
		t.Errorf("rad counters: me %d (want 2), opponent %d (want 0)", radOf(me), radOf(opp))
	}
	if !gaLayered(t, g, forest).Tapped {
		t.Error("the granted {T} did not tap the Forest")
	}
}

// Two Forests of yours: the controller chooses (CR 303.4f). Only Forests
// they control are offered, so an opponent's Forest and a land that is no
// Forest are not, and a pick outside the list is refused.
func TestHaroldAndBobChoosesAmongForestsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	one := hbForest(g, me.ID, me.ID)
	two := hbForest(g, me.ID, me.ID)
	theirs := hbForest(g, opp.ID, opp.ID)
	island := apaPush(g, me.ID, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	harold := hbHarold(g, me.ID)

	hbDie(t, g, harold)

	choice := auraHostChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatalf("no aura_host prompt for the controller: %+v", g.PendingChoices)
	}
	if adZone(g, harold) != game.ZoneGraveyard {
		t.Errorf("Harold and Bob is in %v while the choice is open, want the graveyard (nothing has entered)", adZone(g, harold))
	}
	if len(choice.PickTargetCards) != 2 || !hbHas(choice.PickTargetCards, one) || !hbHas(choice.PickTargetCards, two) {
		t.Fatalf("candidates %v, want exactly my two Forests", choice.PickTargetCards)
	}
	for _, bad := range []uuid.UUID{theirs, island} {
		if err := g.ResolveAuraHost(choice.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); !errors.Is(err, game.ErrIllegalTarget) {
			t.Errorf("an illegal host was accepted: %v", err)
		}
	}
	answerAuraHost(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: two})
	passPriorityAroundTable(t, g)

	c := gaLayered(t, g, harold)
	if adZone(g, harold) != game.ZoneBattlefield || !c.IsAttachedTo(two) {
		t.Fatalf("should enchant the chosen Forest: zone=%v attached=%+v", adZone(g, harold), c.AttachedTo)
	}
	if countGrantedMana(t, g, two) != 1 || countGrantedMana(t, g, one) != 0 {
		t.Error("the grant went to a Forest that is not enchanted")
	}
}

func hbHas(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// CR 303.4g: no Forest to enchant, so it stays in the graveyard.
func TestHaroldAndBobStaysInTheGraveyardWithNoForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	apaPush(g, me.ID, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	harold := hbHarold(g, me.ID)

	hbDie(t, g, harold)

	if adZone(g, harold) != game.ZoneGraveyard {
		t.Errorf("Harold and Bob is in %v, want the graveyard", adZone(g, harold))
	}
	if c := adCard(t, g, harold); c.ReturnedAura || c.IsAttached() {
		t.Errorf("a card that stayed in the graveyard carries the Aura form: %+v", c)
	}
}

// "Forest you control": an opponent's Forest is no host, so with only
// theirs the card stays in the graveyard.
func TestHaroldAndBobCannotEnchantAnOpponentsForest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hbForest(g, opp.ID, opp.ID)
	harold := hbHarold(g, me.ID)

	hbDie(t, g, harold)

	if adZone(g, harold) != game.ZoneGraveyard {
		t.Errorf("Harold and Bob is in %v, want the graveyard", adZone(g, harold))
	}
}

// The Aura is no creature, so "if it was a creature" is false when it
// dies: its Forest is destroyed, CR 704.5m buries it, and it stays in the
// graveyard instead of coming back a second time.
func TestHaroldAndBobAuraDoesNotReturnWhenItsForestIsDestroyed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := hbForest(g, me.ID, me.ID)
	harold := hbHarold(g, me.ID)
	hbDie(t, g, harold)
	if adZone(g, harold) != game.ZoneBattlefield {
		t.Fatalf("setup: Harold and Bob is in %v", adZone(g, harold))
	}

	// A second Forest it could return onto.
	hbForest(g, me.ID, me.ID)
	radDestroy(t, g, forest)
	settleStateBasedActions(t, g)

	if adZone(g, harold) != game.ZoneGraveyard {
		t.Errorf("Harold and Bob is in %v, want the graveyard (an Aura attached to nothing, and it does not return again)", adZone(g, harold))
	}
	if c := adCard(t, g, harold); c.ReturnedAura || c.IsAura() {
		t.Errorf("the card in the graveyard is still the Aura: %+v", c.Effective())
	}
}

// The Aura is held to "Forest you control" (CR 704.5m): a Forest another
// player takes is illegal and the Aura goes to the graveyard.
func TestHaroldAndBobAuraFallsOffAForestAnOpponentTakes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := hbForest(g, me.ID, me.ID)
	harold := hbHarold(g, me.ID)
	hbDie(t, g, harold)

	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.Nil, forest, opp.ID, game.IndefiniteDuration(), "theft") {
			t.Fatal("setup: control of the Forest did not change")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	settleStateBasedActions(t, g)

	if adZone(g, harold) != game.ZoneGraveyard {
		t.Errorf("Harold and Bob is in %v, want the graveyard (the Forest is no longer one I control)", adZone(g, harold))
	}
}

// A restore point with the Aura live keeps the Aura, its attachment, the
// lost abilities and the Forest's grant, and the granted ability still
// works afterwards.
func TestHaroldAndBobSurvivesARestoreAsAnAura(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := hbForest(g, me.ID, me.ID)
	harold := hbHarold(g, me.ID)
	hbDie(t, g, harold)

	restored := restoreRoundTrip(t, g, true)
	restored.WithWriteLock(func() { restored.RecomputeLayersIfStaleLocked() })
	rme := restored.Seats[0]

	c := gaLayered(t, restored, harold)
	if !c.ReturnedAura || !c.IsAura() || c.IsCreature() || !c.IsAttachedTo(forest) || !c.HasLostAllAbilities() {
		t.Fatalf("restored Aura lost its form: returned=%v aura=%v creature=%v attached=%+v lost=%v",
			c.ReturnedAura, c.IsAura(), c.IsCreature(), c.AttachedTo, c.HasLostAllAbilities())
	}
	if !c.HasSupertype("legendary") {
		t.Error("restored Aura is no longer legendary")
	}
	if err := tapGrantedMana(t, restored, rme.ID, forest, "G"); err != nil {
		t.Fatalf("tap the restored Forest: %v", err)
	}
	if got := batch01PoolColors(rme); len(got) != 3 {
		t.Errorf("restored pool %v, want three mana", got)
	}
	if radOf(rme) != 2 {
		t.Errorf("restored rad counters %d, want 2", radOf(rme))
	}

	// It stays legal after the restore: a second pass of state-based
	// actions leaves it on the battlefield.
	passPriorityAroundTable(t, restored)
	if adZone(restored, harold) != game.ZoneBattlefield {
		t.Errorf("restored Aura is in %v, want the battlefield", adZone(restored, harold))
	}
}

// A card with no Aura form can't be returned as one: the primitive
// refuses it and moves nothing.
func TestReturnAsAuraRefusesACardWithNoAuraForm(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hbForest(g, me.ID, me.ID)
	id := adGraveyardCreature(me, "Gravecrawler")
	var err error
	g.WithWriteLock(func() {
		_, err = g.ReturnFromGraveyardAsAuraForEffect(id, uuid.Nil, false)
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam", err)
	}
	if adZone(g, id) != game.ZoneGraveyard {
		t.Errorf("the card is in %v, want the graveyard", adZone(g, id))
	}
}

// CR 303.4f: it is an Aura as it enters, so what enters is no creature.
// A Soul Warden watching "another creature enters" gains nothing, and the
// Aura is never on the battlefield unattached for CR 704.5m to bury.
func TestHaroldAndBobEntersAsAnAuraNotACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hbForest(g, me.ID, me.ID)
	apaPush(g, me.ID, me.ID, game.Card{
		Name: "Soul Warden", OracleID: "f3fad295-1af2-4ecc-8546-b121ad6be27b",
		TypeLine: "Creature — Human Cleric", Power: 1, Toughness: 1,
	})
	harold := hbHarold(g, me.ID)
	before := me.Life

	hbDie(t, g, harold)

	if adZone(g, harold) != game.ZoneBattlefield || !gaLayered(t, g, harold).IsAura() {
		t.Fatalf("setup: Harold and Bob is in %v", adZone(g, harold))
	}
	if me.Life != before {
		t.Errorf("life %d -> %d: Soul Warden saw a creature enter, but Harold and Bob entered as an Aura", before, me.Life)
	}
}
