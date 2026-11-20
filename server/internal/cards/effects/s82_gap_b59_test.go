package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82g59HorrificAssault = "4c7e091e-e3c1-446d-a442-4aacb3cdd575"
	s82g59MalevolentRumb  = "d30d8968-ec24-4155-97e2-28efd52c8810"
	s82g59Glimpse         = "7b70696e-b09e-4bc9-8438-d31075d78c0c"
	s82g59Meltdown        = "8fd85a54-f0c7-4e9f-a4b4-b9b513b98d95"
	s82g59Wheel           = "95a7db5d-26f6-4192-9dc7-4342f7e0e9c5"
	s82g59DeemInferior    = "f0bd4a29-ac13-4d4f-8479-e0e80f8d846f"
	s82g59DecreeJustice   = "67a98359-6f02-4556-ad47-c01b643686d6"
)

func TestS82GapB59EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82g59HorrificAssault: "Horrific Assault",
		s82g59MalevolentRumb:  "Malevolent Rumble",
		s82g59Glimpse:         "Glimpse the Impossible",
		s82g59Meltdown:        "Meltdown",
		s82g59Wheel:           "Wheel of Potential",
		s82g59DeemInferior:    "Deem Inferior",
		s82g59DecreeJustice:   "Decree of Justice",
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

// s82g59Card puts a card in a player's hand and returns its id.
func s82g59Card(zone *game.Zone, p *game.Player, name, typeLine, cost string, oracle ...string) uuid.UUID {
	id := uuid.New()
	o := ""
	if len(oracle) > 0 {
		o = oracle[0]
	}
	zone.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: o,
		Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

func s82g59Refs(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func s82g59InZone(g *game.Game, id uuid.UUID, kind game.ZoneKind) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == kind
}

// --- Horrific Assault ------------------------------------------------

func TestS82GapB59HorrificAssaultBitesAndGainsLifeOnlyWithAnEldrazi(t *testing.T) {
	for _, eldrazi := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		typeLine := "Creature — Bear"
		if eldrazi {
			typeLine = "Creature — Eldrazi"
		}
		mine := b12Creature(g, me.ID, "Biter", typeLine, 3, 3)
		theirs := b12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
		life := me.Life
		if err := d1658CastErr(t, g, "Horrific Assault", "Sorcery", s82g59HorrificAssault,
			game.CastSpellParams{Strict: true, Targets: s82g59Refs(theirs, mine)}); err == nil {
			t.Fatal("cast with the targets reversed")
		}
		d1658Cast(t, g, "Horrific Assault", "Sorcery", s82g59HorrificAssault,
			game.CastSpellParams{Strict: true, Targets: s82g59Refs(mine, theirs)})
		passPriorityAroundTable(t, g)
		g.RunStateChecksForTest()
		if g.Battlefield.Contains(theirs) {
			t.Errorf("eldrazi=%v: the victim took %d damage and survived", eldrazi, 3)
		}
		want := life
		if eldrazi {
			want += 3
		}
		if me.Life != want {
			t.Errorf("eldrazi=%v: life %d, want %d", eldrazi, me.Life, want)
		}
	}
}

// --- Meltdown ----------------------------------------------------------

func TestS82GapB59MeltdownDestroysArtifactsWithManaValueXOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	push := func(owner *game.Player, name, typeLine, cost string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
			ManaCost: cost, Owner: owner.ID, Controller: owner.ID})
	}
	free := push(me, "Free Rock", "Artifact", "")
	one := push(opp, "One Rock", "Artifact", "{1}")
	two := push(opp, "Two Golem", "Artifact Creature — Golem", "{2}")
	three := push(me, "Three Rock", "Artifact", "{3}")
	plain := push(opp, "Bear", "Creature — Bear", "{1}{G}")
	id := s82g59Card(me.Hand, me, "Meltdown", "Sorcery", "", s82g59Meltdown)
	g.Seats[g.Turn.ActiveSeat] = me
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	for name, tc := range map[string]struct {
		id   uuid.UUID
		gone bool
	}{"free": {free, true}, "one": {one, true}, "two": {two, true}, "three": {three, false}, "bear": {plain, false}} {
		if got := !g.Battlefield.Contains(tc.id); got != tc.gone {
			t.Errorf("%s destroyed = %v, want %v", name, got, tc.gone)
		}
	}
}

// --- Malevolent Rumble -------------------------------------------------

func TestS82GapB59MalevolentRumbleTakesAPermanentBinsTheRestAndMakesASpawn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rest := s82g59Card(me.Library, me, "Deep Card", "Creature — Bear", "")
	bolt := s82g59Card(me.Library, me, "Instant", "Instant", "")
	bear := s82g59Card(me.Library, me, "Bear", "Creature — Bear", "")
	land := s82g59Card(me.Library, me, "Forest", "Basic Land — Forest", "")
	sorc := s82g59Card(me.Library, me, "Sorcery", "Sorcery", "")
	castCatalogSpell(t, g, "Malevolent Rumble", "Sorcery", s82g59MalevolentRumb, nil)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the permanent card")
	}
	for _, c := range []uuid.UUID{bolt, sorc, rest} {
		for _, cand := range pick.ChooseCards {
			if cand == c {
				t.Errorf("%v is not a permanent card or not in the top four but was offered", c)
			}
		}
	}
	answerChooseCards(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !s82g59InZone(g, bear, game.ZoneHand) {
		t.Error("the chosen permanent is not in hand")
	}
	for _, id := range []uuid.UUID{land, bolt, sorc} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("%v was not put into the graveyard", id)
		}
	}
	if !s82g59InZone(g, rest, game.ZoneLibrary) {
		t.Error("the fifth card should still be in the library")
	}
	if n, _ := countTokensNamed(g, me.ID, "Eldrazi Spawn"); n != 1 {
		t.Errorf("Eldrazi Spawn tokens = %d, want 1", n)
	}
}

func TestS82GapB59MalevolentRumbleDeclinedStillBinsAllFourAndMakesASpawn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var four []uuid.UUID
	for i := 0; i < 4; i++ {
		four = append(four, s82g59Card(me.Library, me, "Card", "Creature — Bear", ""))
	}
	castCatalogSpell(t, g, "Malevolent Rumble", "Sorcery", s82g59MalevolentRumb, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	for _, id := range four {
		if !me.Graveyard.Contains(id) {
			t.Error("a declined card was not put into the graveyard")
		}
	}
	if n, _ := countTokensNamed(g, me.ID, "Eldrazi Spawn"); n != 1 {
		t.Errorf("Eldrazi Spawn tokens = %d, want 1", n)
	}
}

// --- Glimpse the Impossible --------------------------------------------

func TestS82GapB59GlimpseThePlayedCardsStayAndTheRestBecomeSpawn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	s82g59Card(me.Library, me, "Deep", "Sorcery", "")
	third := s82g59Card(me.Library, me, "Third", "Sorcery", "")
	second := s82g59Card(me.Library, me, "Second", "Sorcery", "")
	first := s82g59Card(me.Library, me, "First", "Sorcery", "")
	castCatalogSpell(t, g, "Glimpse the Impossible", "Sorcery", s82g59Glimpse, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{first, second, third} {
		if !s82g59InZone(g, id, game.ZoneExile) {
			t.Fatalf("%v was not exiled", id)
		}
	}
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play an exiled card this turn: %v", err)
	}
	passPriorityAroundTable(t, g)
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{second, third} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("%v remained exiled", id)
		}
	}
	if s82g59InZone(g, first, game.ZoneExile) {
		t.Error("the played card was put back into exile")
	}
	if n, _ := countTokensNamed(g, me.ID, "Eldrazi Spawn"); n != 2 {
		t.Errorf("Eldrazi Spawn tokens = %d, want 2 (one per card put into the graveyard)", n)
	}
}

func TestS82GapB59GlimpseThePlayingAllThreeLeavesNoSpawn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var top []uuid.UUID
	for i := 0; i < 3; i++ {
		top = append(top, s82g59Card(me.Library, me, "Spell", "Sorcery", ""))
	}
	castCatalogSpell(t, g, "Glimpse the Impossible", "Sorcery", s82g59Glimpse, nil)
	passPriorityAroundTable(t, g)
	for _, id := range top {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Eldrazi Spawn"); n != 0 {
		t.Errorf("Eldrazi Spawn tokens = %d, want 0", n)
	}
}

// --- Wheel of Potential ------------------------------------------------

func s82g59FillLibrary(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		s82g59Card(p.Library, p, "Filler", "Sorcery", "")
	}
}

func TestS82GapB59WheelOfPotentialSevenEnergyDrawsAndOpensThePlayWindow(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	s82g59FillLibrary(me, 10)
	s82g59FillLibrary(opp, 10)
	mine := s82g59Card(me.Hand, me, "My Spell", "Sorcery", "")
	theirs := s82g59Card(opp.Hand, opp, "Their Spell", "Sorcery", "")
	setEnergy(t, g, me, 4)
	castCatalogSpell(t, g, "Wheel of Potential", "Sorcery", s82g59Wheel, nil)
	passPriorityAroundTable(t, g)
	answerNumber(t, g, me.ID, 7)
	if energyOf(me) != 0 {
		t.Errorf("energy = %d after paying 7", energyOf(me))
	}
	answerMayChoice(t, g, me.ID, true)
	if !s82g59InZone(g, mine, game.ZoneExile) {
		t.Error("my hand was not exiled")
	}
	if got := me.Hand.Size(); got != 7 {
		t.Errorf("I drew %d, want 7", got)
	}
	answerMayChoice(t, g, opp.ID, true)
	for _, p := range g.Seats {
		if p.ID != me.ID && p.ID != opp.ID {
			answerMayChoice(t, g, p.ID, false)
		}
	}
	passPriorityAroundTable(t, g)
	if !s82g59InZone(g, theirs, game.ZoneExile) {
		t.Error("their hand was not exiled")
	}
	if got := opp.Hand.Size(); got != 7 {
		t.Errorf("they drew %d, want 7", got)
	}
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play a card I own exiled this way: %v", err)
	}
	if err := g.CastSpell(me.ID, theirs, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("played a card another player owns")
	}
}

func TestS82GapB59WheelOfPotentialSixEnergyGivesNoPlayWindowAndAPlayerMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	s82g59FillLibrary(me, 10)
	s82g59FillLibrary(opp, 10)
	mine := s82g59Card(me.Hand, me, "My Spell", "Sorcery", "")
	kept := s82g59Card(opp.Hand, opp, "Kept Spell", "Sorcery", "")
	setEnergy(t, g, me, 3)
	castCatalogSpell(t, g, "Wheel of Potential", "Sorcery", s82g59Wheel, nil)
	passPriorityAroundTable(t, g)
	answerNumber(t, g, me.ID, 6)
	answerMayChoice(t, g, me.ID, true)
	for _, p := range g.Seats {
		if p.ID != me.ID {
			answerMayChoice(t, g, p.ID, false)
		}
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("with fewer than seven {E} paid the exiled card was playable")
	}
	if !s82g59InZone(g, kept, game.ZoneHand) {
		t.Error("the player who declined lost their hand")
	}
	if got := me.Hand.Size(); got != 6 {
		t.Errorf("I drew %d, want 6", got)
	}
}

// --- Deem Inferior -----------------------------------------------------

func TestS82GapB59DeemInferiorCostsLessPerCardDrawnAndOwnerChoosesLane(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	s82g59FillLibrary(me, 5)
	lib := seedLibrary(opp, "Top", "Next")
	perm := seedCreature(g, "Their Creature", opp.ID)
	advanceToMain(t, g)
	id := s82g59Card(me.Hand, me, "Deem Inferior", "Sorcery", "{3}{U}", s82g59DeemInferior)
	g.WithWriteLock(func() {
		if have := g.TurnTallyFor(me.ID).CardsDrawn; have < 2 {
			_ = g.DrawNForEffect(me.ID, 2-have)
		}
	})
	if got := g.TurnTallyFor(me.ID).CardsDrawn; got != 2 {
		t.Skipf("the turn already counts %d drawn cards", got)
	}
	fillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: s82g59Refs(perm)}); err == nil {
		t.Fatal("{U} alone paid for {1}{U}")
	}
	fillPoolColored(me, "C", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: s82g59Refs(perm)}); err != nil {
		t.Fatalf("two cards drawn should make it {1}{U}: %v", err)
	}
	passPriorityAroundTable(t, g)
	ask := putInLibraryChoiceFor(g, opp.ID)
	if ask == nil || ask.LibraryTopDepth != 2 {
		t.Fatalf("the owner is not asked for second from the top or the bottom: %+v", ask)
	}
	if err := g.ResolvePutInLibrary(ask.ID, opp.ID, nil, []uuid.UUID{perm}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := libraryTopIDs(opp, 3); !sameIDs(got, []uuid.UUID{lib[0], perm, lib[1]}) {
		t.Errorf("top three %v, want [Top perm Next]", got)
	}
}

// --- Decree of Justice -------------------------------------------------

func TestS82GapB59DecreeOfJusticeMakesXAngels(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := s82g59Card(me.Hand, me, "Decree of Justice", "Sorcery", "", s82g59DecreeJustice)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Angel" && c.Controller == me.ID {
			n++
			if c.Power != 4 || c.Toughness != 4 || !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Errorf("Angel is %d/%d without flying", c.Power, c.Toughness)
			}
		}
	}
	if n != 2 {
		t.Errorf("Angels = %d, want 2", n)
	}
}

func TestS82GapB59DecreeOfJusticeCycledPaysXForSoldiers(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		id, me := cycleFromHand(t, g, "Decree of Justice", "Sorcery", s82g59DecreeJustice, "{C}{C}{W}")
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		if !me.Graveyard.Contains(id) {
			t.Fatal("the cycled card is not in the graveyard")
		}
		answerNumber(t, g, me.ID, 3)
		answerPayUnless(t, g, me.ID, pay)
		passPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 3
		}
		if n, _ := countTokensNamed(g, me.ID, "Soldier"); n != want {
			t.Errorf("pay=%v: Soldiers = %d, want %d", pay, n, want)
		}
	}
}
