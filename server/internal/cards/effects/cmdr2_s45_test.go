package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2S45ThiefOfSanityOracle = "68b51d58-7566-4a26-8c31-8da32799bea7"
)

// stackLibrary replaces `p`'s library with n named cards, the first returned
// the top.
func cmdr2S45StackLibrary(p *game.Player, n int) []uuid.UUID {
	p.Library.Cards = nil
	var lib []uuid.UUID
	for i := 1; i <= n; i++ {
		c := game.NewCard("Library Card "+string(rune('A'+i-1)), p.ID)
		c.TypeLine = "Creature — Test"
		c.ManaCost = "{1}"
		p.Library.PushTop(c)
		lib = append([]uuid.UUID{c.InstanceID}, lib...)
	}
	return lib
}

func TestCmdr2S45ThiefOfSanityExilesOneOfTopThreeAndBinsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := cmdr2S45StackLibrary(opp, 6)
	thief := b43Catalog(g, me.ID, "Thief of Sanity", "Creature — Specter", cmdr2S45ThiefOfSanityOracle, 2, 2)
	attackWith(t, g, opp.ID, thief)
	passPriorityAroundTable(t, g)

	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("want a choose-exactly-one prompt, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lib[3]}); err == nil {
		t.Fatal("a card below the top three was accepted")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lib[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	perm := e2Permission(g, lib[1])
	if perm == nil || perm.Player != me.ID || !perm.AnyType || !perm.CastOnly {
		t.Errorf("permission %+v, want an any-type cast grant to the controller", perm)
	}
	e2AssertFaceDownFor(t, g, lib[1], me.ID)
	if z := e2Zone(g, lib[1]); z != game.ZoneExile {
		t.Errorf("picked card in %q, want exile", z)
	}
	for _, i := range []int{0, 2} {
		if z := e2Zone(g, lib[i]); z != game.ZoneGraveyard {
			t.Errorf("card %d in %q, want the graveyard", i, z)
		}
	}
	if len(opp.Library.Cards) != 3 {
		t.Errorf("library has %d cards, want 3", len(opp.Library.Cards))
	}
}

const (
	cmdr2S45BandanaOracle = "3c958b5f-680b-45fc-8c3d-27b23677436e"
	cmdr2S45GontiOracle   = "7923ef4a-b1be-4f6a-b17b-4eab101390de"
)

func TestCmdr2S45BandanaIgnoresAnUnequippedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	lib := cmdr2S45StackLibrary(opp, 4)
	b43Catalog(g, me.ID, "Dream-Thief's Bandana", "Artifact — Equipment", cmdr2S45BandanaOracle, 0, 0)
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, other)
	passPriorityAroundTable(t, g)
	if z := e2Zone(g, lib[0]); z != game.ZoneLibrary {
		t.Fatalf("an unequipped creature's damage exiled the top card (%q)", z)
	}
}

func TestCmdr2S45BandanaEquippedAttackerGrantsAnyTypePlay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	lib := cmdr2S45StackLibrary(opp, 4)
	bandana := b43Catalog(g, me.ID, "Dream-Thief's Bandana", "Artifact — Equipment", cmdr2S45BandanaOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	equipTo(t, g, me.ID, bandana, bear)

	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if z := e2Zone(g, lib[0]); z != game.ZoneExile {
		t.Fatalf("top card in %q, want exile", z)
	}
	perm := e2Permission(g, lib[0])
	if perm == nil || perm.Player != me.ID || perm.CastOnly || !perm.AnyType {
		t.Errorf("permission %+v, want a play (not cast-only) any-type grant to the Bandana's controller", perm)
	}
	e2AssertFaceDownFor(t, g, lib[0], me.ID)
	if z := e2Zone(g, lib[1]); z != game.ZoneLibrary {
		t.Errorf("second card in %q, want it left in the library", z)
	}
}

func TestCmdr2S45GontiExilesOneCardPerPlayerHit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b43Catalog(g, me.ID, "Gonti, Canny Acquisitor", "Legendary Creature — Aetherborn Rogue", cmdr2S45GontiOracle, 5, 5)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	lib := cmdr2S45StackLibrary(opp, 4)

	attackWith(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if z := e2Zone(g, lib[0]); z != game.ZoneExile {
		t.Fatalf("top card in %q, want exile", z)
	}
	if z := e2Zone(g, lib[1]); z != game.ZoneLibrary {
		t.Errorf("two creatures connected but %q holds the second card: want one trigger, one card", z)
	}
	perm := e2Permission(g, lib[0])
	if perm == nil || perm.Player != me.ID || perm.CastOnly || !perm.AnyType {
		t.Errorf("permission %+v, want a play any-type grant to Gonti's controller", perm)
	}
	e2AssertFaceDownFor(t, g, lib[0], me.ID)
}

func TestCmdr2S45GontiMakesSpellsYouDontOwnCostOneLess(t *testing.T) {
	cast := func(withGonti bool, owner func(me, opp *game.Player) uuid.UUID) error {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if withGonti {
			b43Catalog(g, me.ID, "Gonti, Canny Acquisitor", "Legendary Creature — Aetherborn Rogue", cmdr2S45GontiOracle, 5, 5)
		}
		id := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Stolen Bolt", TypeLine: "Instant", ManaCost: "{1}",
			Owner: owner(me, opp), Controller: me.ID,
		})
		advanceToMain(t, g)
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	theirs := func(_, opp *game.Player) uuid.UUID { return opp.ID }
	mine := func(me, _ *game.Player) uuid.UUID { return me.ID }
	if err := cast(false, theirs); err == nil {
		t.Fatal("setup: a {1} spell was cast with no mana and no Gonti")
	}
	if err := cast(true, theirs); err != nil {
		t.Errorf("a spell you don't own should cost {0} with Gonti: %v", err)
	}
	if err := cast(true, mine); err == nil {
		t.Error("a spell you own was reduced by Gonti")
	}
}

func TestCmdr2S45PredatorsHourGrantsMenaceAndTheExileTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	lib := cmdr2S45StackLibrary(opp, 4)
	b12PlayFromHand(t, g, "Predators' Hour", "Sorcery", "4d504d5f-02a4-4104-89b7-88884483f53a", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, a, "menace") {
		t.Fatal("creature did not gain menace")
	}
	// A creature that enters after resolution gets nothing (CR 611.2c).
	late := b12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if hasEffectiveKeyword(t, g, late, "menace") {
		t.Error("a creature that entered after resolution gained menace")
	}
	attackWith(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if z := e2Zone(g, lib[0]); z != game.ZoneExile {
		t.Fatalf("top card in %q, want exile", z)
	}
	perm := e2Permission(g, lib[0])
	if perm == nil || perm.Player != me.ID || perm.CastOnly || perm.AnyType || !perm.AnyColor {
		t.Errorf("permission %+v, want a play grant with any-colour (not any-type) mana to the controller", perm)
	}
	e2AssertFaceDownFor(t, g, lib[0], me.ID)
}

const (
	cmdr2S45NashiOracle   = "d04b6c2c-50fc-464d-9991-230efec70d2b"
	cmdr2S45InkEyesOracle = "5d520476-740a-4005-801e-472b24fa6497"
	cmdr2S45ShinobiOracle = "4fc5f871-a809-4a4a-adfb-eb4cd52837c8"
)

func TestCmdr2S45NinjutsuCardsDeclareTheirKeywordCosts(t *testing.T) {
	for _, tc := range []struct{ name, oracle, mana string }{
		{"Nashi", cmdr2S45NashiOracle, "{3}{B}"},
		{"Ink-Eyes", cmdr2S45InkEyesOracle, "{3}{B}{B}"},
		{"Fallen Shinobi", cmdr2S45ShinobiOracle, "{2}{U}{B}"},
	} {
		var found bool
		for _, ab := range game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle}) {
			if ab.Cost.Mana == tc.mana && !ab.Cost.ReturnToHand.Empty() && game.AbilityFunctionsFromZone(ab, game.ZoneHand) {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: no ninjutsu %s ability that works from the hand", tc.name, tc.mana)
		}
	}
}

func TestCmdr2S45NashiExilesATopCardPerPlayerAndSpendsOneBudgetOnLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nashi := b43Catalog(g, me.ID, "Nashi, Moon Sage's Scion", "Legendary Creature — Rat Ninja", cmdr2S45NashiOracle, 3, 2)
	_ = nashi
	var tops []uuid.UUID
	for _, p := range g.Seats {
		lib := cmdr2S45StackLibrary(p, 3)
		tops = append(tops, lib[0])
	}
	attackWith(t, g, opp.ID, nashi)
	passPriorityAroundTable(t, g)

	for i, id := range tops {
		if z := e2Zone(g, id); z != game.ZoneExile {
			t.Fatalf("seat %d's top card is in %q, want exile", i, z)
		}
		perm := e2Permission(g, id)
		if perm == nil || perm.Player != me.ID || !perm.LifeEqualToManaValue || perm.CastsLeft != 1 || perm.BudgetGroup == uuid.Nil {
			t.Fatalf("seat %d's card permission %+v, want a one-use life-for-mana-value grant to you", i, perm)
		}
	}
	advanceToMain(t, g)
	// Cast one of the opponent's cards: life equal to its mana value ({1} = 1).
	before := me.Life
	if err := g.CastSpell(me.ID, tops[1], game.CastSpellParams{FromZone: "exile", AlternativeCost: "nashi_moon_sages_scion"}); err != nil {
		t.Fatalf("cast an exiled card for life: %v", err)
	}
	if got := before - me.Life; got != 1 {
		t.Errorf("paid %d life, want its mana value 1", got)
	}
	if p := e2Permission(g, tops[2]); p != nil && p.CastsLeft > 0 {
		t.Errorf("the other cards still carry a live permission %+v: it is one play in all", p)
	}
	if err := g.CastSpell(me.ID, tops[0], game.CastSpellParams{FromZone: "exile", AlternativeCost: "nashi_moon_sages_scion", Strict: true}); err == nil {
		t.Error("a second card was played: Nashi allows one of those cards")
	}
}

func TestCmdr2S45InkEyesMayReanimateFromTheDamagedPlayersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ink := b43Catalog(g, me.ID, "Ink-Eyes, Servant of Oni", "Legendary Creature — Rat Ninja", cmdr2S45InkEyesOracle, 5, 4)
	victim := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: victim, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Dead Spell", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})

	attackWith(t, g, opp.ID, ink)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("the creature card did not come back")
	}
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Errorf("it came back under %s, want your control", got)
	}
	if g.Battlefield.Contains(spell) || g.Battlefield.Contains(mine) {
		t.Error("a non-creature card or your own graveyard card came back")
	}
}

func TestCmdr2S45InkEyesDeclinedLeavesTheGraveyardAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ink := b43Catalog(g, me.ID, "Ink-Eyes, Servant of Oni", "Legendary Creature — Rat Ninja", cmdr2S45InkEyesOracle, 5, 4)
	victim := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: victim, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	attackWith(t, g, opp.ID, ink)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("a declined reanimation happened")
	}
}

func TestCmdr2S45FallenShinobiExilesTwoForAFreePlay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shinobi := b43Catalog(g, me.ID, "Fallen Shinobi", "Creature — Zombie Ninja", cmdr2S45ShinobiOracle, 5, 4)
	lib := cmdr2S45StackLibrary(opp, 5)
	attackWith(t, g, opp.ID, shinobi)
	passPriorityAroundTable(t, g)
	for _, i := range []int{0, 1} {
		if z := e2Zone(g, lib[i]); z != game.ZoneExile {
			t.Fatalf("card %d in %q, want exile", i, z)
		}
		perm := e2Permission(g, lib[i])
		if perm == nil || perm.Player != me.ID || perm.Cost != "{0}" || perm.CastOnly {
			t.Errorf("card %d permission %+v, want a free play grant to the controller", i, perm)
		}
	}
	if z := e2Zone(g, lib[2]); z != game.ZoneLibrary {
		t.Errorf("third card in %q, want it left in the library", z)
	}
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, lib[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Logf("cast outside the window: %v", err)
	}
}
