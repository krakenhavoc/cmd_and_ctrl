package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b31_cards_test.go — Star Trek Commander, batch 31 (#4974). Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB31BorgQueenOracle     = "b03d6bc2-d9f1-4d03-b1b3-77993bec6547"
	trekB31SpockOracle         = "60644afe-9645-45a1-9eac-70f327daa6a7"
	trekB31PicardOracle        = "b95a6001-2773-43c1-a5ad-a37ef7c9f054"
	trekB31WorfOracle          = "a7fcdbab-491f-40d9-be40-409af0d6c3f5"
	trekB31KEhleyrOracle       = "ce60a1a4-1e53-4985-a54f-3627dd35c5f2"
	trekB31BalanceOracle       = "0b035f3c-ba21-4a73-b4df-11a7e8c19056"
	trekB31CrusherOracle       = "7a33edc3-e938-463f-8568-7d7a458427c8"
	trekB31IncursionOracle     = "f00dfc16-c253-4483-94fb-3c2beabd5820"
	trekB31PulaskiOracle       = "686d8449-e132-4478-b4f5-f5cadc8971c8"
	trekB31RegenerationOracle  = "58e9f5db-c369-42c6-a26d-fc31e25d5f83"
	trekB31BorgQueenTypeLine   = "Legendary Artifact Creature — Borg Noble"
	trekB31SpockTypeLine       = "Legendary Creature — Vulcan Human Scientist"
	trekB31PicardTypeLine      = "Legendary Creature — Human Officer"
	trekB31WorfTypeLine        = "Legendary Creature — Klingon Warrior Officer"
	trekB31KEhleyrTypeLine     = "Legendary Creature — Klingon Human Advisor"
	trekB31CrusherTypeLine     = "Legendary Creature — Human Doctor"
	trekB31PulaskiTypeLine     = "Legendary Creature — Human Doctor"
	trekB31BalanceTypeLine     = "Enchantment — Saga"
	trekB31IncursionTypeLine   = "Instant"
	trekB31RegenerationTypeLin = "Instant"
)

func TestTrekB31EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB31BorgQueenOracle:    "The Borg Queen",
		trekB31SpockOracle:        "Commander Spock",
		trekB31PicardOracle:       "Jean-Luc Picard, Luminary",
		trekB31WorfOracle:         "Worf, Defiant Commander",
		trekB31KEhleyrOracle:      "Ambassador K'Ehleyr",
		trekB31BalanceOracle:      "Balance of Terror",
		trekB31CrusherOracle:      "Beverly Crusher, Ship's Doctor",
		trekB31IncursionOracle:    "Borg Incursion",
		trekB31PulaskiOracle:      "Dr. Pulaski, Cynical Clinician",
		trekB31RegenerationOracle: "Emergency Regeneration",
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

// --- shared helpers ------------------------------------------------------------

// trekB31Enter puts a card from owner's hand onto the battlefield (an
// ordinary entry, so enters triggers fire) and returns its ID.
func trekB31Enter(t *testing.T, g *game.Game, owner *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	c.Owner, c.Controller = owner.ID, owner.ID
	g.WithWriteLock(func() { owner.Hand.PushTop(c) })
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		id, err = g.PutFromHandOntoBattlefieldForEffect(c.InstanceID, game.HandEntryOptions{Controller: owner.ID})
		if err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	return id
}

// trekB31Library replaces p's library with the given cards, the first
// one on top.
func trekB31Library(g *game.Game, p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	g.WithWriteLock(func() {
		p.Library.Cards = nil
		for i := len(cards) - 1; i >= 0; i-- {
			c := cards[i]
			if c.InstanceID == uuid.Nil {
				c.InstanceID = uuid.New()
			}
			c.Owner, c.Controller = p.ID, p.ID
			p.Library.PushTop(c)
			ids[i] = c.InstanceID
		}
	})
	return ids
}

// trekB31Draw has p draw n cards through the effect API.
func trekB31Draw(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, n) })
	g.RunStateChecksForTest()
}

// trekB31PlusOnes is the +1/+1 counters on a permanent.
func trekB31PlusOnes(g *game.Game, id uuid.UUID) int {
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[game.CounterPlusOne]
}

// trekB31Counters is the counters of `kind` on a permanent.
func trekB31Counters(g *game.Game, id uuid.UUID, kind string) int {
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[kind]
}

// trekB31Pick answers the newest pick_target prompt of chooser with refs.
func trekB31Pick(t *testing.T, g *game.Game, chooser uuid.UUID, refs ...game.TargetRef) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTargets(p.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

// --- Emergency Regeneration ------------------------------------------------------

func TestTrekB31EmergencyRegenerationProtectsArtifactsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rock := b27Push(g, me.ID, "My Rock", "Artifact", "", "{2}", 0, 0)
	drone := b27Push(g, me.ID, "My Drone", "Artifact Creature — Borg", "", "{1}", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	addPool(t, g, me.ID, "{C}{W}")
	castFromHandForTest(t, g, me, "Emergency Regeneration", trekB31RegenerationTypeLin, "{1}{W}", trekB31RegenerationOracle,
		game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{rock, drone} {
		if !effectiveAbilitiesContain(t, g, id, "hexproof") || !effectiveAbilitiesContain(t, g, id, "indestructible") {
			t.Errorf("%s lacks hexproof and indestructible", layeredCard(t, g, id).Name)
		}
	}
	if effectiveAbilitiesContain(t, g, bear, "hexproof") || effectiveAbilitiesContain(t, g, bear, "indestructible") {
		t.Error("a nonartifact creature gained the keywords")
	}
	b27Kill(g, drone)
	g.RunStateChecksForTest()
	if b12ZoneOf(g, drone) != game.ZoneBattlefield {
		t.Error("the indestructible artifact creature was destroyed")
	}
}

// --- Dr. Pulaski ----------------------------------------------------------------

func TestTrekB31PulaskiGainsFourOnTheSecondDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Dr. Pulaski, Cynical Clinician", trekB31PulaskiTypeLine, trekB31PulaskiOracle, false)
	life := me.Life
	trekB31Draw(g, me, 1)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("the first draw of the turn gained %d", me.Life-life)
	}
	trekB31Draw(g, me, 1)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatalf("the second draw gains 4: %d → %d", life, me.Life)
	}
	trekB31Draw(g, me, 1)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("the third draw gained life: %d", me.Life-life)
	}
}

// --- Beverly Crusher -------------------------------------------------------------

func TestTrekB31CrusherPutsCountersOnUpToThreeCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Beverly Crusher, Ship's Doctor", trekB31CrusherTypeLine, trekB31CrusherOracle, false)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	d := b12Creature(g, me.ID, "D", "Creature — Bear", 2, 2)
	trekB31Draw(g, me, 1)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the first draw asked for targets")
	}
	trekB31Draw(g, me, 1)
	p := latestPickTarget(g, me.ID)
	if p == nil || p.PickTargetMin != 0 || p.PickTargetMax != 3 {
		t.Fatalf("the second draw asks for up to three creatures: %+v", p)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(a, b, c, d)); err == nil {
		t.Error("four targets were accepted")
	}
	trekB31Pick(t, g, me.ID, cardRefs(a, b, c)...)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if n := trekB31PlusOnes(g, id); n != 1 {
			t.Errorf("a target has %d +1/+1 counters, want 1", n)
		}
	}
	if n := trekB31PlusOnes(g, d); n != 0 {
		t.Errorf("the untargeted creature has %d counters", n)
	}
}

// --- Borg Incursion ---------------------------------------------------------------

func TestTrekB31BorgIncursionMakesTwoBorgAndProtectsArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	drone := b27Push(g, me.ID, "My Drone", "Artifact Creature — Borg", "", "{1}", 1, 1)
	rock := b27Push(g, me.ID, "My Rock", "Artifact", "", "{2}", 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	addPool(t, g, me.ID, "{C}{C}{W}")
	castFromHandForTest(t, g, me, "Borg Incursion", trekB31IncursionTypeLine, "{2}{W}", trekB31IncursionOracle,
		game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "R")
	passPriorityAroundTable(t, g)
	var borg []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Borg" && c.IsToken() && c.Controller == me.ID {
			borg = append(borg, c)
		}
	}
	if len(borg) != 2 {
		t.Fatalf("%d Borg tokens, want 2", len(borg))
	}
	want := game.ProtectionFromColor("R")
	for _, id := range []uuid.UUID{drone, borg[0].InstanceID, borg[1].InstanceID} {
		c := layeredCard(t, g, id)
		if !effectiveAbilitiesContain(t, g, id, want) {
			t.Errorf("%s lacks %s", c.Name, want)
		}
		if c.CurrentPower() != 1 || c.CurrentToughness() != 1 || !c.IsArtifact() {
			t.Errorf("%s is %d/%d artifact=%v", c.Name, c.CurrentPower(), c.CurrentToughness(), c.IsArtifact())
		}
	}
	if effectiveAbilitiesContain(t, g, rock, want) {
		t.Error("a noncreature artifact gained protection")
	}
	if effectiveAbilitiesContain(t, g, bear, want) {
		t.Error("a nonartifact creature gained protection")
	}
}

// --- Worf, Defiant Commander -------------------------------------------------------

func TestTrekB31WorfMakesAKlingonWarrior(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	trekB31Enter(t, g, me, game.Card{Name: "Worf, Defiant Commander", TypeLine: trekB31WorfTypeLine,
		OracleID: trekB31WorfOracle, Power: 3, Toughness: 4})
	passPriorityAroundTable(t, g)
	var tok *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Klingon Warrior" && c.IsToken() {
			tok = c
		}
	}
	if tok == nil {
		t.Fatal("no Klingon Warrior token")
	}
	c := layeredCard(t, g, tok.InstanceID)
	if haste := effectiveAbilitiesContain(t, g, tok.InstanceID, "haste"); c.CurrentPower() != 2 || c.CurrentToughness() != 1 || !haste {
		t.Errorf("the token is %d/%d haste=%v", c.CurrentPower(), c.CurrentToughness(), haste)
	}
}

func TestTrekB31WorfExilesOneCardPerOpponentAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Worf, Defiant Commander", trekB31WorfTypeLine, trekB31WorfOracle, false)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	lib := trekB31Library(g, me,
		game.Card{Name: "One", TypeLine: "Sorcery", ManaCost: "{1}"},
		game.Card{Name: "Two", TypeLine: "Sorcery", ManaCost: "{1}"},
		game.Card{Name: "Three", TypeLine: "Sorcery", ManaCost: "{1}"})
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, x := range []struct{ id, def uuid.UUID }{{a, opp1.ID}, {b, opp1.ID}, {c, opp2.ID}} {
		if err := g.DeclareAttacker(x.id, x.def); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for i, id := range lib {
		inExile := g.Exile.Contains(id)
		if want := i < 2; inExile != want {
			t.Errorf("card %d in exile = %v, want %v (two opponents attacked)", i, inExile, want)
		}
		if i < 2 {
			if perm := g.CastPermissionOnCardByIDForEffect(id); perm == nil || perm.Player != me.ID {
				t.Errorf("card %d has no play permission for Worf's controller", i)
			}
		}
	}
}

// --- Ambassador K'Ehleyr -----------------------------------------------------------

func TestTrekB31KEhleyrCountsNonlandRevealsThenEveryoneDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	keh := pushCatalogPermanent(g, me.ID, "Ambassador K'Ehleyr", trekB31KEhleyrTypeLine, trekB31KEhleyrOracle, false)
	tops := map[uuid.UUID]uuid.UUID{}
	for i, p := range g.Seats {
		typeLine := "Sorcery"
		if i%2 == 1 {
			typeLine = "Land"
		}
		ids := trekB31Library(g, p, game.Card{Name: "Top", TypeLine: typeLine}, game.Card{Name: "Next", TypeLine: "Land"})
		tops[p.ID] = ids[0]
	}
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if !p.Hand.Contains(tops[p.ID]) {
			t.Errorf("seat %s did not draw the card it revealed", p.ID)
		}
	}
	nonland := (len(g.Seats) + 1) / 2
	if n := trekB31PlusOnes(g, keh); n != nonland {
		t.Errorf("K'Ehleyr has %d counters, want %d (one per nonland reveal)", n, nonland)
	}
	if !effectiveAbilitiesContain(t, g, keh, "vigilance") {
		t.Error("K'Ehleyr lacks vigilance")
	}
}

// --- Balance of Terror ---------------------------------------------------------------

func TestTrekB31BalanceOfTerrorDestroysThenBlinks(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	relic := b27Push(g, opp.ID, "Their Relic", "Artifact", "", "{2}", 0, 0)
	saga := castCatalogSpell(t, g, "Balance of Terror", trekB31BalanceTypeLine, trekB31BalanceOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("chapter I asks for a target")
	}
	for _, id := range p.PickTargetCards {
		if id == saga {
			t.Error("chapter I offered the Saga itself")
		}
	}
	trekB31Pick(t, g, me.ID, cardRefs(relic)...)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, relic) != game.ZoneGraveyard {
		t.Fatalf("the artifact is in %v, want the graveyard", b12ZoneOf(g, relic))
	}

	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToPrecombatMainOf(t, g, seat)
	trekB31Pick(t, g, me.ID, cardRefs(bear)...)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, bear) != game.ZoneExile {
		t.Fatalf("chapter II: the creature is in %v, want exile", b12ZoneOf(g, bear))
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	back := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "My Bear" && c.Controller == me.ID {
			back = true
		}
	}
	if !back {
		t.Error("the exiled creature did not return at the next end step")
	}
}

// --- Jean-Luc Picard, Luminary ----------------------------------------------------------

func TestTrekB31PicardDrawsForHistoricOfManaValueThreeOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Jean-Luc Picard, Luminary", trekB31PicardTypeLine, trekB31PicardOracle, false)
	trekB31Library(g, me, game.Card{Name: "A", TypeLine: "Land"}, game.Card{Name: "B", TypeLine: "Land"},
		game.Card{Name: "C", TypeLine: "Land"})
	hand := me.Hand.Size()
	trekB31Enter(t, g, me, game.Card{Name: "Cheap Rock", TypeLine: "Artifact", ManaCost: "{2}"})
	trekB31Enter(t, g, me, game.Card{Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Power: 4, Toughness: 4})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("a mana value 2 artifact or a nonhistoric creature drew a card (%d → %d)", hand, me.Hand.Size())
	}
	trekB31Enter(t, g, me, game.Card{Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{3}"})
	passPriorityAroundTable(t, g)
	trekB31Enter(t, g, me, game.Card{Name: "Legend", TypeLine: "Legendary Creature — Human", ManaCost: "{2}{W}", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("two historic permanents of mana value 3 drew %d cards, want 2", me.Hand.Size()-hand)
	}
}

func TestTrekB31PicardSeekOutCountersACreatureAndChargesASpacecraft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	advanceToMain(t, g)
	picard := pushCatalogPermanent(g, me.ID, "Jean-Luc Picard, Luminary", trekB31PicardTypeLine, trekB31PicardOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ship := b27Push(g, me.ID, "My Ship", "Artifact — Spacecraft", "", "{3}", 0, 0)
	trekB31Draw(g, me, 1)
	trekB31Draw(g, me, 1)
	first := latestPickTarget(g, me.ID)
	if first == nil {
		t.Fatal("the second draw asks for targets")
	}
	for _, id := range first.PickTargetCards {
		if id == picard {
			t.Error("the creature clause offered Picard himself")
		}
	}
	trekB31Pick(t, g, me.ID, cardRefs(bear)...)
	second := latestPickTarget(g, me.ID)
	if second == nil {
		t.Fatal("no Spacecraft clause")
	}
	for _, id := range second.PickTargetCards {
		if id == bear {
			t.Error("the Spacecraft clause offered a non-Spacecraft")
		}
	}
	trekB31Pick(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: ship, Slot: 1})
	passPriorityAroundTable(t, g)
	if n := trekB31PlusOnes(g, bear); n != 1 {
		t.Errorf("the creature has %d +1/+1 counters, want 1", n)
	}
	if n := trekB31Counters(g, ship, game.CounterCharge); n != 2 {
		t.Errorf("the Spacecraft has %d charge counters, want 2", n)
	}
}

// --- The Borg Queen -------------------------------------------------------------------

func TestTrekB31BorgQueenAssimilatesTheFirstCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	lib := trekB31Library(g, opp,
		game.Card{Name: "Their Land", TypeLine: "Land"},
		game.Card{Name: "Their Spell", TypeLine: "Sorcery", ManaCost: "{1}"},
		game.Card{Name: "Their Elf", TypeLine: "Creature — Elf Warrior", ManaCost: "{1}{G}", Power: 1, Toughness: 1},
		game.Card{Name: "Their Deep Card", TypeLine: "Land"})
	queen := trekB31Enter(t, g, me, game.Card{Name: "The Borg Queen", TypeLine: trekB31BorgQueenTypeLine,
		OracleID: trekB31BorgQueenOracle, Power: 4, Toughness: 4})
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)

	var elf uuid.UUID
	for _, x := range g.Battlefield.Cards {
		if x.Name == "Their Elf" {
			elf = x.InstanceID
		}
	}
	if elf == uuid.Nil {
		t.Fatalf("the creature card is not on the battlefield")
	}
	c := layeredCard(t, g, elf)
	if c.Controller != me.ID || !c.IsArtifact() || !c.HasSubtype("Borg") || c.HasSubtype("Elf") || c.HasSubtype("Warrior") {
		t.Errorf("assimilated: controller mine %v, artifact %v, subtypes %v", c.Controller == me.ID, c.IsArtifact(), c.Effective().Subtypes)
	}
	if n := trekB31PlusOnes(g, elf); n != 1 {
		t.Errorf("the assimilated card has %d +1/+1 counters, want 1", n)
	}
	// 1/1 + counter + the Queen's +2/+2 anthem.
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("the assimilated card is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if q := layeredCard(t, g, queen); q.CurrentPower() != 6 || q.CurrentToughness() != 6 {
		t.Errorf("the Queen is %d/%d, want 6/6 under her own anthem", q.CurrentPower(), q.CurrentToughness())
	}
	if opp.Library.Size() != 3 {
		t.Fatalf("the opponent's library has %d cards, want 3", opp.Library.Size())
	}
	if top := opp.Library.Cards[opp.Library.Size()-1].InstanceID; top != lib[3] {
		t.Error("the unexiled card is no longer on top")
	}
	for _, id := range lib[:2] {
		if b12ZoneOf(g, id) != game.ZoneLibrary {
			t.Errorf("an exiled noncreature card is in %v, want the bottom of the library", b12ZoneOf(g, id))
		}
	}
}

func TestTrekB31BorgQueenTargetsOnlyAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	trekB31Enter(t, g, me, game.Card{Name: "The Borg Queen", TypeLine: trekB31BorgQueenTypeLine,
		OracleID: trekB31BorgQueenOracle, Power: 4, Toughness: 4})
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Error("the Queen's controller was accepted as target opponent")
	}
}

// --- Commander Spock --------------------------------------------------------------------

func trekB31SpockGame(t *testing.T, lands int, top game.Card) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	for i := 0; i < lands; i++ {
		b27Push(g, me.ID, "Land "+string(rune('A'+i)), "Land", "", "", 0, 0)
	}
	ids := trekB31Library(g, me, top, game.Card{Name: "Second", TypeLine: "Land"})
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	trekB31Enter(t, g, me, game.Card{Name: "Commander Spock", TypeLine: trekB31SpockTypeLine,
		OracleID: trekB31SpockOracle, Power: 3, Toughness: 5})
	passPriorityAroundTable(t, g)
	return g, me, ids[0]
}

func TestTrekB31SpockOffersAFreeCastWithinTheLandCount(t *testing.T) {
	g, me, top := trekB31SpockGame(t, 2, game.Card{Name: "Bear Cub", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, top)
	if err := castThroughResolutionCast(t, g, me.ID, top, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the revealed card: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, top) != game.ZoneBattlefield {
		t.Fatalf("the cast creature is in %v", b12ZoneOf(g, top))
	}
	if me.Hand.Size() != 0 {
		t.Errorf("a card was drawn after the cast: hand %d", me.Hand.Size())
	}
}

func TestTrekB31SpockDrawsTheCardWhenTheCastIsDeclined(t *testing.T) {
	g, me, top := trekB31SpockGame(t, 2, game.Card{Name: "Bear Cub", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Errorf("the declined card is in %v, want drawn into hand", b12ZoneOf(g, top))
	}
}

func TestTrekB31SpockDrawsASpellOverTheLandCount(t *testing.T) {
	g, me, top := trekB31SpockGame(t, 1, game.Card{Name: "Bear Cub", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a mana value 2 spell was offered with one land")
	}
	if !me.Hand.Contains(top) {
		t.Errorf("the too-expensive card is in %v, want drawn", b12ZoneOf(g, top))
	}
}

func TestTrekB31SpockPutsALandOntoTheBattlefieldTapped(t *testing.T) {
	g, me, top := trekB31SpockGame(t, 0, game.Card{Name: "Strange Land", TypeLine: "Land"})
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardByID(g, top)
	if c == nil || !c.Tapped || c.Controller != me.ID {
		t.Fatalf("the land is in %v (tapped %v), want the battlefield tapped", b12ZoneOf(g, top), c != nil && c.Tapped)
	}
	if me.Hand.Size() != 0 {
		t.Errorf("a card was drawn after the land was put: hand %d", me.Hand.Size())
	}
}

func TestTrekB31SpockDrawsALandItLeavesOnTop(t *testing.T) {
	g, me, top := trekB31SpockGame(t, 0, game.Card{Name: "Strange Land", TypeLine: "Land"})
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Errorf("the land left on top is in %v, want drawn", b12ZoneOf(g, top))
	}
}
