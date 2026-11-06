package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dsk_hl2_cards_test.go — Duskmourn: House of Horror, the H–L cards
// #5220 left (#4466). Each test plays the real catalog entry and asserts
// what its text says.

const (
	dskHL2HandOracle         = "6b348051-3ba8-4aa0-b337-b42d8ef8cbba"
	dskHL2CartographerOracle = "76d93d50-4fc1-4ed4-a6dc-5fb193286b54"
	dskHL2InfernoOracle      = "fa7c8092-94da-4147-aa90-0a951d2a8be8"
	dskHL2JumpScareOracle    = "783aab37-3ec3-478f-a4fa-cccdfffbae18"
	dskHL2KonaOracle         = "76372327-41ec-4516-abff-3b612e073629"
	dskHL2CartographerType   = "Creature — Human Scout Survivor"
)

func TestDskHL2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dskHL2HandOracle:         "Hand That Feeds",
		dskHL2CartographerOracle: "House Cartographer",
		dskHL2InfernoOracle:      "Impossible Inferno",
		dskHL2JumpScareOracle:    "Jump Scare",
		dskHL2KonaOracle:         "Kona, Rescue Beastie",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// dskHL2SeedTypes puts the first n of four card types into `p`'s
// graveyard: four is delirium, three is not.
func dskHL2SeedTypes(p *game.Player, n int) {
	cards := []struct{ name, typeLine, cost string }{
		{"Old Bear", "Creature — Bear", "{1}{G}"},
		{"Old Bolt", "Instant", "{R}"},
		{"Old Ritual", "Sorcery", "{B}"},
		{"Old Forest", "Basic Land — Forest", ""},
	}
	for _, c := range cards[:n] {
		gyCard(p, c.name, c.typeLine, c.cost, "")
	}
}

// --- Hand That Feeds -----------------------------------------------------------------

func TestDskHL2HandThatFeedsGetsTwoAndMenaceWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := b12Push(g, me.ID, "Hand That Feeds", "Creature — Mutant", dskHL2HandOracle, 2, 2)
	dskHL2SeedTypes(me, 4)
	declareAttack(t, g, opp.ID, hand)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, hand); p != 4 || tough != 2 {
		t.Errorf("Hand That Feeds is %d/%d, want 4/2", p, tough)
	}
	if !hasKeywordOf(g, hand, "menace") {
		t.Error("Hand That Feeds did not gain menace")
	}
}

func TestDskHL2HandThatFeedsDoesNotTriggerWithoutDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := b12Push(g, me.ID, "Hand That Feeds", "Creature — Mutant", dskHL2HandOracle, 2, 2)
	dskHL2SeedTypes(me, 3)
	declareAttack(t, g, opp.ID, hand)
	if n := g.Stack.Size(); n != 0 {
		t.Errorf("three card types: %d items on the stack, want no trigger", n)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, hand); p != 2 || tough != 2 {
		t.Errorf("Hand That Feeds is %d/%d, want 2/2", p, tough)
	}
	if hasKeywordOf(g, hand, "menace") {
		t.Error("Hand That Feeds gained menace without delirium")
	}
}

// --- House Cartographer --------------------------------------------------------------

func TestDskHL2HouseCartographerRevealsUntilALandWhenTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := samiLibraryTopFirst(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	scout := pushCatalogPermanent(g, me.ID, "House Cartographer", dskHL2CartographerType, dskHL2CartographerOracle, false)
	dskB1Tap(g, scout)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[2]) {
		t.Fatal("the land revealed did not reach the hand")
	}
	if me.Hand.Contains(ids[0]) || me.Hand.Contains(ids[1]) {
		t.Error("a nonland card revealed went to the hand")
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[3] {
		t.Error("the card below the land must now be on top; the revealed nonlands go to the bottom")
	}
	bottom := []uuid.UUID{me.Library.Cards[0].InstanceID, me.Library.Cards[1].InstanceID}
	if !slices.Contains(bottom, ids[0]) || !slices.Contains(bottom, ids[1]) {
		t.Error("the revealed nonland cards are not on the bottom of the library")
	}
}

func TestDskHL2HouseCartographerDoesNothingUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := samiLibraryTopFirst(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	pushCatalogPermanent(g, me.ID, "House Cartographer", dskHL2CartographerType, dskHL2CartographerOracle, false)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(ids[1]) {
		t.Error("an untapped Cartographer revealed for a land")
	}
}

// --- Impossible Inferno --------------------------------------------------------------

func TestDskHL2ImpossibleInfernoDealsSixAndImpulsesWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	dskHL2SeedTypes(me, 4)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Top Land", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	castCatalogSpell(t, g, "Impossible Inferno", "Instant", dskHL2InfernoOracle, dskDGTarget(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("6 damage did not kill the 6/6")
	}
	if !g.Exile.Contains(ids[0]) || !me.Library.Contains(ids[1]) {
		t.Fatal("with delirium the top card of the library should be exiled and the next one left alone")
	}
	perm := exiledPermission(g, ids[0])
	if perm == nil || perm.Player != me.ID || perm.CastOnly || !permissionLive(g, perm, me.ID) {
		t.Errorf("%+v: the caster may PLAY the exiled card, a land included", perm)
	}
}

func TestDskHL2ImpossibleInfernoOnlyDamagesWithoutDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 7)
	dskHL2SeedTypes(me, 3)
	ids := seedSearchLibrary(me, game.Card{Name: "Top Card", TypeLine: "Sorcery"})
	castCatalogSpell(t, g, "Impossible Inferno", "Instant", dskHL2InfernoOracle, dskDGTarget(big))
	passPriorityAroundTable(t, g)
	if c := b12Card(t, g, big); c.DamageMarked != 6 {
		t.Errorf("the 6/7 has %d damage marked, want 6", c.DamageMarked)
	}
	if !me.Library.Contains(ids[0]) || g.Exile.Contains(ids[0]) {
		t.Error("without delirium the top card was exiled")
	}
}

// --- Jump Scare ----------------------------------------------------------------------

func TestDskHL2JumpScarePumpsGrantsFlyingAndMakesAHorrorEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	// The table starts in the draw step: an instant may be cast there.
	if err := dskDGCastAtDrawStep(t, g, "Jump Scare", "Instant", dskHL2JumpScareOracle, bear); err != nil {
		t.Fatalf("cast in the draw step: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	if !hasKeywordOf(g, bear, "flying") {
		t.Error("the Bear did not gain flying")
	}
	if !hasEffectiveType(t, g, bear, "Enchantment") || !hasEffectiveType(t, g, bear, "Creature") {
		t.Errorf("the Bear's types are %v, want an enchantment creature", effectiveTypes(t, g, bear))
	}
	if subs := effectiveSubtypes(t, g, bear); !slices.Contains(subs, "Horror") || !slices.Contains(subs, "Bear") {
		t.Errorf("the Bear's subtypes are %v, want Horror in addition to Bear", subs)
	}
}

// --- Kona, Rescue Beastie ------------------------------------------------------------

func TestDskHL2KonaPutsAPermanentCardFromHandWhenTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kona := pushCatalogPermanent(g, me.ID, "Kona, Rescue Beastie", "Legendary Creature — Beast Survivor", dskHL2KonaOracle, false)
	rock := handCardForTest(me, "Big Rock", "Artifact", "")
	bolt := handCardForTest(me, "Bolt", "Instant", "")
	dskB1Tap(g, kona)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("a tapped Kona did not offer a permanent card from hand")
	}
	if pick.ChooseMin != 0 {
		t.Errorf("ChooseMin = %d, want 0: the put is a \"you may\"", pick.ChooseMin)
	}
	if slices.Contains(pick.ChooseCards, bolt) {
		t.Error("an instant was offered: it is not a permanent card")
	}
	answerChooseCards(t, g, me.ID, rock)
	g.SettleResolution()
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact did not enter the battlefield")
	}
}

func TestDskHL2KonaDoesNothingUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Kona, Rescue Beastie", "Legendary Creature — Beast Survivor", dskHL2KonaOracle, false)
	handCardForTest(me, "Big Rock", "Artifact", "")
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("an untapped Kona offered a card from hand")
	}
}
