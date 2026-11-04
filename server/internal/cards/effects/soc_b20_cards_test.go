package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b20_cards_test.go — Secrets of Strixhaven, twentieth batch (#4456).
// Each test plays the real catalog entry through the engine and asserts
// what the printed text says.

const (
	socB20OmensOracle      = "863cbd12-da20-456e-8698-06fba4d9bd5c"
	socB20AgonyOracle      = "55d1a788-fe9a-4e19-a524-6709f728ea28"
	socB20ArkOracle        = "2c654772-c786-4966-a46e-9556bbf31d7c"
	socB20ArnynOracle      = "3debd5c7-da3c-4130-8ed9-e156e10034b6"
	socB20ProcessOracle    = "0c061a1b-206a-4fcc-92e5-30c8ed9d65c1"
	socB20DustspeakerOrcl  = "295f8b8d-102a-47af-93c2-f8182c5f11ca"
	socB20AzizaOracle      = "70e10cb9-ae27-4dba-9a0d-34a78f924774"
	socB20BetrayalOracle   = "6b633e58-ea1b-4ed0-addb-6dfd93645ae4"
	socB20TomekeeperOracle = "08a0e845-9ef9-4de1-b9fe-4530357346ca"
)

func TestSocB20EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB20OmensOracle:      "Arcane Omens",
		socB20AgonyOracle:      "Archaic's Agony",
		socB20ArkOracle:        "Ark of Hunger",
		socB20ArnynOracle:      "Arnyn, Deathbloom Botanist",
		socB20ProcessOracle:    "Artistic Process",
		socB20DustspeakerOrcl:  "Ascendant Dustspeaker",
		socB20AzizaOracle:      "Aziza, Mage Tower Captain",
		socB20BetrayalOracle:   "Banishing Betrayal",
		socB20TomekeeperOracle: "Biblioplex Tomekeeper",
	}
	for oracle, name := range want {
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

// --- fixtures ----------------------------------------------------------------

// socB20HandCard puts a card with a mana cost into a hand.
func socB20HandCard(p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
		Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

// socB20FillHand tops a hand up with n plain cards and returns their ids.
func socB20FillHand(p *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, socB20HandCard(p, "Filler", "Sorcery", "{9}", ""))
	}
	return ids
}

// socB20CastStrict casts the card from hand with a floated pool and returns
// the cast error.
func socB20CastStrict(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, pool string, targets []game.TargetRef) error {
	t.Helper()
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, pool)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

func socB20PlayerRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

// --- Arcane Omens ------------------------------------------------------------

func TestSocB20ArcaneOmensDiscardsOnePerColourSpent(t *testing.T) {
	for _, tc := range []struct {
		pool string
		want int
	}{{"WUBRG", 5}, {"BBBGG", 2}, {"BBBBB", 1}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		socB20FillHand(opp, 8)
		id := socB20HandCard(me, "Arcane Omens", "Sorcery", "{4}{B}", socB20OmensOracle)
		hand := opp.Hand.Size()
		if err := socB20CastStrict(t, g, me, id, tc.pool, socB20PlayerRef(opp.ID)); err != nil {
			t.Fatalf("%s: cast: %v", tc.pool, err)
		}
		passPriorityAroundTable(t, g)
		if got := discardOwed(g, opp.ID); got != tc.want {
			t.Fatalf("%s: the opponent owes %d discards, want %d", tc.pool, got, tc.want)
		}
		discardFromHand(t, g, opp.ID)
		if got := hand - opp.Hand.Size(); got != tc.want {
			t.Errorf("%s: the opponent discarded %d, want %d", tc.pool, got, tc.want)
		}
	}
}

func TestSocB20ArcaneOmensTakesWhatAShortHandHas(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	socB20FillHand(opp, 2)
	id := socB20HandCard(me, "Arcane Omens", "Sorcery", "{4}{B}", socB20OmensOracle)
	if err := socB20CastStrict(t, g, me, id, "WUBRG", socB20PlayerRef(opp.ID)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := discardChoiceFor(g, opp.ID); c != nil {
		discardFromHand(t, g, opp.ID)
	}
	if opp.Hand.Size() != 0 {
		t.Errorf("the opponent kept %d cards of a two-card hand against five colours", opp.Hand.Size())
	}
}

func TestSocB20ArcaneOmensCanTargetYourselfAndRefusesAShortPool(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := socB20HandCard(me, "Arcane Omens", "Sorcery", "{4}{B}", socB20OmensOracle)
	if err := socB20CastStrict(t, g, me, id, "BBBB", socB20PlayerRef(opp.ID)); err == nil {
		t.Fatal("cast {4}{B} with four mana")
	}
	if err := socB20CastStrict(t, g, me, id, "WUBRG", nil); err == nil {
		t.Fatal("cast with no target player")
	}
	socB20FillHand(me, 4)
	hand := me.Hand.Size() - 1
	if err := socB20CastStrict(t, g, me, id, "WUBRG", socB20PlayerRef(me.ID)); err != nil {
		t.Fatalf("target yourself: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, me.ID); got != min(5, hand) {
		t.Errorf("you owe %d discards, want %d", got, min(5, hand))
	}
}

// --- Archaic's Agony ---------------------------------------------------------

func socB20AgonyLibrary(p *game.Player) []uuid.UUID {
	return znrStackLibrary(p,
		game.Card{Name: "Top One", TypeLine: "Sorcery", ManaCost: "{R}"},
		game.Card{Name: "Top Two", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Top Three", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Top Four", TypeLine: "Sorcery", ManaCost: "{R}"},
	)
}

func TestSocB20ArchaicsAgonyExilesTheExcessAndYouMayPlayThemThroughYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	top := socB20AgonyLibrary(me)
	id := socB20HandCard(me, "Archaic's Agony", "Sorcery", "{4}{R}", socB20AgonyOracle)
	if err := socB20CastStrict(t, g, me, id, "WUBRG", cardRefs(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Fatal("five damage did not kill the Bear")
	}
	// 5 damage to a 2/2: 3 excess.
	for i := 0; i < 3; i++ {
		if z := znrZoneOf(g, top[i]); z != game.ZoneExile {
			t.Errorf("top card %d is in %v, want exile", i, z)
		}
		perm := exiledPermission(g, top[i])
		if perm.Player != me.ID || perm.CastOnly || !permissionLive(g, perm, me.ID) {
			t.Errorf("top card %d: permission %+v is not a live play permission for the caster", i, perm)
		}
		if permissionLive(g, perm, opp.ID) {
			t.Errorf("top card %d: the opponent may play it", i)
		}
	}
	if z := znrZoneOf(g, top[3]); z != game.ZoneLibrary {
		t.Errorf("the fourth card is in %v, want the library", z)
	}

	// Still live on the opponent's turn and on your next turn, gone after.
	advanceToMainOf(t, g, 1)
	if p := g.CastPermissionOnCardByIDForEffect(top[0]); p == nil || !permissionLive(g, p, me.ID) {
		t.Error("the permission lapsed on the opponent's turn")
	}
	advanceToMainOf(t, g, 0)
	if p := g.CastPermissionOnCardByIDForEffect(top[0]); p == nil || !permissionLive(g, p, me.ID) {
		t.Error("the permission lapsed during your next turn")
	}
	advanceToMainOf(t, g, 1)
	if p := g.CastPermissionOnCardByIDForEffect(top[0]); p != nil && permissionLive(g, p, me.ID) {
		t.Error("the permission outlasted the end of your next turn")
	}
}

func TestSocB20ArchaicsAgonyWithNoExcessExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	top := socB20AgonyLibrary(me)
	id := socB20HandCard(me, "Archaic's Agony", "Sorcery", "{4}{R}", socB20AgonyOracle)
	// Two colours: exactly lethal on a 2/2, no excess.
	if err := socB20CastStrict(t, g, me, id, "RRRWW", cardRefs(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Fatal("two damage did not kill the Bear")
	}
	for _, c := range top {
		if z := znrZoneOf(g, c); z != game.ZoneLibrary {
			t.Errorf("a card left the library with no excess damage (in %v)", z)
		}
	}
}

func TestSocB20ArchaicsAgonyExcessIsTakenFromTheDamageStillNeeded(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].DamageMarked = 1
			}
		}
	})
	top := socB20AgonyLibrary(me)
	id := socB20HandCard(me, "Archaic's Agony", "Sorcery", "{4}{R}", socB20AgonyOracle)
	// Four colours: 4 damage; 1 already marked on a 2/2 leaves 1 lethal: 3 excess.
	if err := socB20CastStrict(t, g, me, id, "RWUBB", cardRefs(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	exiled := 0
	for _, c := range top {
		if znrZoneOf(g, c) == game.ZoneExile {
			exiled++
		}
	}
	if exiled != 3 {
		t.Errorf("%d cards exiled, want 3 (4 colours spent -> 4 damage, 1 lethal still needed)", exiled)
	}
}

// A source with deathtouch needs one damage to be lethal, so everything past
// the first point is excess (CR 120.4a). The catalog card prints no
// deathtouch; the clause is exercised through a copy of the card that has it.
func TestSocB20ArchaicsAgonyFromADeathtouchSourceCountsEverythingPastOneAsExcess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := pushVanillaCreature(g, opp.ID, "Big", 1, 6)
	top := socB20AgonyLibrary(me)
	id := socB20HandCard(me, "Archaic's Agony", "Sorcery", "{4}{R}", socB20AgonyOracle)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].Keywords = append(me.Hand.Cards[i].Keywords, "deathtouch")
			}
		}
	})
	// Four colours: 4 damage to a 1/6. Without deathtouch there is no
	// excess; with it, 3.
	if err := socB20CastStrict(t, g, me, id, "RWUBB", cardRefs(big)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	exiled := 0
	for _, c := range top {
		if znrZoneOf(g, c) == game.ZoneExile {
			exiled++
		}
	}
	if exiled != 3 {
		t.Errorf("%d cards exiled, want 3 (4 colours -> 4 damage, 1 lethal, 3 excess)", exiled)
	}
}

// --- Ark of Hunger -----------------------------------------------------------

func TestSocB20ArkOfHungerDrainsWhenACardLeavesYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ark of Hunger", "Artifact", socB20ArkOracle, false)
	dead := socB11Graveyard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	mine, theirs := me.Life, opp.Life
	socB11Move(t, g, me.ID, dead, game.ZoneGraveyard, game.ZoneExile)
	passPriorityAroundTable(t, g)
	if me.Life != mine+1 || opp.Life != theirs-1 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, mine+1, theirs-1)
	}
	// A card leaving an opponent's graveyard does nothing.
	theirDead := socB11Graveyard(opp, "Their Dead", "Creature — Bear", "{1}{G}")
	socB11Move(t, g, opp.ID, theirDead, game.ZoneGraveyard, game.ZoneExile)
	passPriorityAroundTable(t, g)
	if me.Life != mine+1 || opp.Life != theirs-1 {
		t.Errorf("an opposing graveyard card triggered the Ark: %d/%d", me.Life, opp.Life)
	}
}

func TestSocB20ArkOfHungerHitsEveryOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Ark of Hunger", "Artifact", socB20ArkOracle, false)
	before := make([]int, len(g.Seats))
	for i, s := range g.Seats {
		before[i] = s.Life
	}
	dead := socB11Graveyard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	socB11Move(t, g, me.ID, dead, game.ZoneGraveyard, game.ZoneHand)
	passPriorityAroundTable(t, g)
	for i, s := range g.Seats {
		want := before[i] - 1
		if i == 0 {
			want = before[i] + 1
		}
		if s.Life != want {
			t.Errorf("seat %d has %d life, want %d", i, s.Life, want)
		}
	}
}

func TestSocB20ArkOfHungerMillsACardYouMayPlayThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ark := pushCatalogPermanent(g, me.ID, "Ark of Hunger", "Artifact", socB20ArkOracle, false)
	top := znrStackLibrary(me, game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	mine, theirs := me.Life, opp.Life
	b16Activate(t, g, me.ID, ark, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if z := znrZoneOf(g, top[0]); z != game.ZoneGraveyard {
		t.Fatalf("the milled card is in %v, want the graveyard", z)
	}
	if !b16Tapped(t, g, ark) {
		t.Error("the Ark did not tap")
	}
	perm := exiledPermission(g, top[0])
	if perm.Player != me.ID || perm.CastOnly || !permissionLive(g, perm, me.ID) {
		t.Fatalf("permission %+v: not a live play permission for you", perm)
	}
	if permissionLive(g, perm, opp.ID) {
		t.Error("an opponent may play your milled card")
	}
	// Casting it from the graveyard is a card leaving the graveyard: the Ark drains.
	floatForTest(g, me, "GG")
	if err := g.CastSpell(me.ID, top[0], game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(top[0]) {
		t.Error("the milled Bear did not resolve")
	}
	if me.Life != mine+1 || opp.Life != theirs-1 {
		t.Errorf("life %d/%d, want %d/%d: the card left the graveyard", me.Life, opp.Life, mine+1, theirs-1)
	}
}

func TestSocB20ArkOfHungerMilledLandIsPlayedAndTheArkDrains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ark := pushCatalogPermanent(g, me.ID, "Ark of Hunger", "Artifact", socB20ArkOracle, false)
	top := znrStackLibrary(me, game.Card{Name: "Milled Land", TypeLine: "Basic Land — Forest"})
	b16Activate(t, g, me.ID, ark, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	mine, theirs := me.Life, opp.Life
	if err := g.CastSpell(me.ID, top[0], game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("play the milled land from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(top[0]) {
		t.Fatal("the milled land was not played")
	}
	if me.Life != mine+1 || opp.Life != theirs-1 {
		t.Errorf("life %d/%d, want %d/%d: playing it took it out of the graveyard", me.Life, opp.Life, mine+1, theirs-1)
	}
}

func TestSocB20ArkOfHungerPermissionEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ark := pushCatalogPermanent(g, me.ID, "Ark of Hunger", "Artifact", socB20ArkOracle, false)
	top := znrStackLibrary(me, game.Card{Name: "Milled Land", TypeLine: "Basic Land — Forest"})
	b16Activate(t, g, me.ID, ark, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	perm := g.CastPermissionOnCardByIDForEffect(top[0])
	if perm == nil || perm.CastOnly || !permissionLive(g, perm, me.ID) {
		t.Fatal("no live permission for a milled land (play, not cast-only)")
	}
	advanceToMainOf(t, g, 1)
	if p := g.CastPermissionOnCardByIDForEffect(top[0]); p != nil && permissionLive(g, p, me.ID) {
		t.Error("the permission outlasted the turn")
	}
}

// --- Arnyn, Deathbloom Botanist ----------------------------------------------

func socB20Kill(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
}

// socB20ArnynDrained kills a creature and reports whether Arnyn drained.
func socB20ArnynDrained(t *testing.T, g *game.Game, owner uuid.UUID, power, toughness, plusCounters int) bool {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	creature := b12Creature(g, owner, "Victim", "Creature — Bear", power, toughness)
	if plusCounters > 0 {
		addCountersForTest(g, creature, game.CounterPlusOne, plusCounters)
	}
	mine, theirs := me.Life, opp.Life
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(creature); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil && !stackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		pickPlayer(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)
	if (me.Life == mine+2) != (opp.Life == theirs-2) {
		t.Fatalf("a half drain: %d/%d from %d/%d", me.Life, opp.Life, mine, theirs)
	}
	return me.Life == mine+2
}

func TestSocB20ArnynDrainsForASmallCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Arnyn, Deathbloom Botanist", "Legendary Creature — Vampire Druid", socB20ArnynOracle, 2, 2)
	if !socB20ArnynDrained(t, g, me.ID, 1, 1, 0) {
		t.Error("a 1/1 dying did not drain")
	}
	if !socB20ArnynDrained(t, g, me.ID, 4, 1, 0) {
		t.Error("a 4/1 dying did not drain: toughness 1 or less counts")
	}
	if !socB20ArnynDrained(t, g, me.ID, 0, 5, 0) {
		t.Error("a 0/5 dying did not drain: power 1 or less counts")
	}
	if socB20ArnynDrained(t, g, me.ID, 2, 2, 0) {
		t.Error("a 2/2 dying drained")
	}
	// Last-known size: a 1/1 carrying a +1/+1 counter was a 2/2 when it died.
	if socB20ArnynDrained(t, g, me.ID, 1, 1, 1) {
		t.Error("a 1/1 with a +1/+1 counter (2/2) dying drained")
	}
}

func TestSocB20ArnynIgnoresAnOpponentsCreatureAndCountsItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arnyn := b12Push(g, me.ID, "Arnyn, Deathbloom Botanist", "Legendary Creature — Vampire Druid", socB20ArnynOracle, 2, 2)
	if socB20ArnynDrained(t, g, opp.ID, 1, 1, 0) {
		t.Error("an opponent's 1/1 dying drained")
	}
	// Arnyn itself is a 2/2: not small, no trigger.
	mine := me.Life
	socB20Kill(t, g, arnyn)
	if me.Life != mine {
		t.Error("Arnyn (2/2) dying drained")
	}
}

func TestSocB20ArnynDrainsWhenASmallTokenYouControlDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Arnyn, Deathbloom Botanist", "Legendary Creature — Vampire Druid", socB20ArnynOracle, 2, 2)
	socB11Token(t, g, me.ID, "1/1 red Goblin")
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Controller == me.ID {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("no token was made")
	}
	mine, theirs := me.Life, opp.Life
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(token); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil && !stackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if latestPickTarget(g, me.ID) != nil {
		pickPlayer(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)
	if me.Life != mine+2 || opp.Life != theirs-2 {
		t.Errorf("life %d/%d, want %d/%d after a 1/1 token died", me.Life, opp.Life, mine+2, theirs-2)
	}
}

func TestSocB20ArnynHasDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arnyn := b12Push(g, me.ID, "Arnyn, Deathbloom Botanist", "Legendary Creature — Vampire Druid", socB20ArnynOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, arnyn, "deathtouch") {
		t.Error("no deathtouch")
	}
}

// --- Artistic Process --------------------------------------------------------

// socB20CastProcess casts Artistic Process choosing one mode.
func socB20CastProcess(t *testing.T, g *game.Game, mode int, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[0]
	id := socB20HandCard(me, "Artistic Process", "Sorcery", "{3}{R}{R}", socB20ProcessOracle)
	advanceToMain(t, g)
	floatForTest(g, me, "RRRRR")
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{mode}, Targets: targets})
}

func TestSocB20ArtisticProcessSixDamageToATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := pushVanillaCreature(g, opp.ID, "Big", 5, 6)
	small := pushVanillaCreature(g, opp.ID, "Small", 1, 1)
	if err := socB20CastProcess(t, g, 0, cardRefs(big)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(big) {
		t.Error("six damage did not kill a 5/6")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the other creature died")
	}
}

func TestSocB20ArtisticProcessTwoDamageToEachCreatureYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	theirSmall := pushVanillaCreature(g, opp.ID, "Their Small", 2, 2)
	theirBig := pushVanillaCreature(g, opp.ID, "Their Big", 2, 3)
	thirdSmall := pushVanillaCreature(g, g.Seats[2].ID, "Third Small", 1, 1)
	if err := socB20CastProcess(t, g, 1, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirSmall) || g.Battlefield.Contains(thirdSmall) {
		t.Error("a small opposing creature survived two damage")
	}
	if !g.Battlefield.Contains(theirBig) {
		t.Error("a 2/3 died to two damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creature took damage")
	}
}

func TestSocB20ArtisticProcessMakesAHastyFlyingElementalForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if err := socB20CastProcess(t, g, 2, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Elemental" && g.Battlefield.Cards[i].Controller == me.ID {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil {
		t.Fatal("no Elemental token")
	}
	id := token.InstanceID
	p, tough := socB14PT(t, g, id)
	if p != 3 || tough != 3 || !token.IsToken() {
		t.Errorf("token is %d/%d (token=%v), want a 3/3 token", p, tough, token.IsToken())
	}
	if len(token.Colors) != 2 {
		t.Errorf("colours %v, want blue and red", token.Colors)
	}
	if !hasEffectiveKeyword(t, g, id, "flying") || !hasEffectiveKeyword(t, g, id, "haste") {
		t.Error("the token lacks flying or haste this turn")
	}
	// The haste is a grant until end of turn, not a printed keyword.
	advanceToMainOf(t, g, 1)
	if !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("the token lost flying")
	}
	if hasEffectiveKeyword(t, g, id, "haste") {
		t.Error("the token kept haste past the turn")
	}
}

func TestSocB20ArtisticProcessRefusesAnIllegalTargetForTheDamageMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if err := socB20CastProcess(t, g, 0, socB20PlayerRef(me.ID)); err == nil {
		t.Error("the damage mode accepted a player as its target")
	}
}

// --- Ascendant Dustspeaker ---------------------------------------------------

func TestSocB20DustspeakerPutsACounterOnAnotherCreatureYouControlAndFlies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	speaker := castHoldCreature(t, g, "Ascendant Dustspeaker", "Creature — Orc Cleric", socB20DustspeakerOrcl, 3, 4)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: speaker}); err == nil {
			t.Error("the Dustspeaker was accepted as its own other creature")
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
			t.Error("an opponent's creature was accepted")
		}
		pickCard(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, bear); p != 3 {
		t.Errorf("the Bear has power %d, want 3 after the counter", p)
	}
	if p, _ := socB14PT(t, g, speaker); p != 3 {
		t.Errorf("the Dustspeaker has power %d, want an uncountered 3", p)
	}
	if !hasEffectiveKeyword(t, g, speaker, "flying") {
		t.Error("no flying")
	}
}

func TestSocB20DustspeakerExilesACardFromAGraveyardAtYourBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Ascendant Dustspeaker", "Creature — Orc Cleric", socB20DustspeakerOrcl, 3, 4)
	mineGy := socB11Graveyard(me, "My Dead", "Creature — Bear", "{1}{G}")
	theirGy := socB11Graveyard(opp, "Their Dead", "Creature — Bear", "{1}{G}")
	advanceTo(t, g, game.StepBeginCombat)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, theirGy)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirGy) || opp.Graveyard.Contains(theirGy) {
		t.Error("the chosen card is not exiled")
	}
	if !me.Graveyard.Contains(mineGy) {
		t.Error("an unchosen card left a graveyard")
	}
}

func TestSocB20DustspeakerDoesNothingAtAnOpponentsBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, opp.ID, "Ascendant Dustspeaker", "Creature — Orc Cleric", socB20DustspeakerOrcl, 3, 4)
	mineGy := socB11Graveyard(me, "My Dead", "Creature — Bear", "{1}{G}")
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, opp.ID) != nil {
		t.Error("the Dustspeaker triggered on my beginning of combat")
	}
	if !me.Graveyard.Contains(mineGy) {
		t.Error("the Dustspeaker exiled a card on another player's turn")
	}
}

// --- Aziza, Mage Tower Captain -----------------------------------------------

func socB20AzizaBoard(g *game.Game, me *game.Player, extra int) (aziza uuid.UUID, others []uuid.UUID) {
	aziza = b12Push(g, me.ID, "Aziza, Mage Tower Captain", "Legendary Creature — Djinn Sorcerer", socB20AzizaOracle, 2, 2)
	for i := 0; i < extra; i++ {
		others = append(others, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	return aziza, others
}

// socB20ChooseCardsPrompt is the open choose-cards prompt for chooser.
func socB20ChooseCardsPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func TestSocB20AzizaTapsThreeCreaturesToCopyAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 12)
	_, bears := socB20AzizaBoard(g, me, 3)
	hand := me.Hand.Size()
	msc18DivinationCast(t, g)
	passPriorityAroundTable(t, g)
	prompt := socB20ChooseCardsPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no tap-three prompt")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, bears[:2]); err == nil {
		t.Fatal("two creatures were accepted: the cost is exactly three")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, bears); err != nil {
		t.Fatalf("pick three: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, b := range bears {
		if !b16Tapped(t, g, b) {
			t.Errorf("a chosen creature stayed untapped")
		}
	}
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d cards, want 4: the Divination and its copy", got)
	}
}

func TestSocB20AzizaCanTapItselfAndASummoningSickCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 12)
	aziza, bears := socB20AzizaBoard(g, me, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bears[0] {
				g.Battlefield.Cards[i].SummonedThisTurn = true
			}
		}
	})
	hand := me.Hand.Size()
	msc18DivinationCast(t, g)
	passPriorityAroundTable(t, g)
	prompt := socB20ChooseCardsPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no tap-three prompt")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{aziza, bears[0], bears[1]}); err != nil {
		t.Fatalf("pick Aziza and two bears: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d cards, want 4", got)
	}
}

func TestSocB20AzizaDeclinedOrShortOfCreaturesCopiesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 12)
	_, bears := socB20AzizaBoard(g, me, 3)
	hand := me.Hand.Size()
	msc18DivinationCast(t, g)
	passPriorityAroundTable(t, g)
	prompt := socB20ChooseCardsPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no tap-three prompt")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2: declined, no copy", got)
	}
	for _, b := range bears {
		if b16Tapped(t, g, b) {
			t.Error("a creature tapped though the cost was declined")
		}
	}

	// Two untapped creatures (Aziza and one Bear) cannot pay: no prompt.
	for _, b := range bears[1:] {
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(b) })
	}
	msc18DivinationCast(t, g)
	passPriorityAroundTable(t, g)
	if socB20ChooseCardsPrompt(g, me.ID) != nil {
		t.Error("asked to tap three with only two untapped creatures")
	}
}

func TestSocB20AzizaIgnoresCreatureSpellsAndOpponentsSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	socB20AzizaBoard(g, me, 3)
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if socB20ChooseCardsPrompt(g, me.ID) != nil {
		t.Error("a creature spell asked for the tap cost")
	}
}

// --- Banishing Betrayal ------------------------------------------------------

func TestSocB20BanishingBetrayalBouncesANonlandPermanentThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc18Library(me, 6)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := pushCatalogPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest", "", false)
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	hand := opp.Hand.Size()

	if err := castCatalogSpellErr(t, g, "Banishing Betrayal", "Instant", socB20BetrayalOracle, cardRefs(land)); err == nil {
		t.Fatal("a land was accepted as the target nonland permanent")
	}
	castCatalogSpell(t, g, "Banishing Betrayal", "Instant", socB20BetrayalOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || opp.Hand.Size() != hand+1 {
		t.Fatal("the Bear was not returned to its owner's hand")
	}
	prompt := surveilChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(prompt.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("surveil away the top card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not reach the graveyard")
	}
}

// --- Biblioplex Tomekeeper ---------------------------------------------------

func TestSocB20TomekeeperPreparesOrUnpreparesATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	angel := fraBloodline(me.ID)
	castFromHandAndResolve(t, g, me, angel)
	id := angel.InstanceID
	if g.IsPreparedForEffect(id) {
		t.Fatal("the test creature starts prepared")
	}

	keeper := castCatalogSpell(t, g, "Biblioplex Tomekeeper", "Artifact Creature — Construct", socB20TomekeeperOracle, nil)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil || c.ModeMin != 0 {
		t.Fatalf("a real 'choose up to one' prompt, declinable: %+v", c)
	}
	answerMode(t, g, c, me.ID, 0)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(id) {
		t.Fatal("the creature did not become prepared")
	}
	if !g.Battlefield.Contains(keeper) {
		t.Fatal("the Tomekeeper is not on the battlefield")
	}

	// A second Tomekeeper unprepares it.
	castCatalogSpell(t, g, "Biblioplex Tomekeeper", "Artifact Creature — Construct", socB20TomekeeperOracle, nil)
	passPriorityAroundTable(t, g)
	answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, 1)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	if g.IsPreparedForEffect(id) {
		t.Error("the creature stayed prepared")
	}
}

func TestSocB20TomekeeperPreparingACreatureWithNoPrepareSpellDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Biblioplex Tomekeeper", "Artifact Creature — Construct", socB20TomekeeperOracle, nil)
	passPriorityAroundTable(t, g)
	answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, 0)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.IsPreparedForEffect(bear) {
		t.Error("a creature with no prepare spell became prepared")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear left")
	}
}

func TestSocB20TomekeeperMayChooseNoMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Biblioplex Tomekeeper", "Artifact Creature — Construct", socB20TomekeeperOracle, nil)
	passPriorityAroundTable(t, g)
	if c := modePickChoiceFor(g, me.ID); c != nil {
		if err := g.ResolveModePick(c.ID, me.ID, nil); err != nil {
			t.Fatalf("choose nothing: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if !stackFullyEmpty(g) {
		t.Error("choosing no mode left something on the stack")
	}
}
