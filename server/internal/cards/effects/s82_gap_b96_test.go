package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b96_test.go — set GAP, batch 96: Doctor Who sorceries (S82, #5948).

const (
	gapB96Flux       = "7089512e-f9b6-4d8e-bccc-29e6785cc041"
	gapB96AllHistory = "48f80047-a4f9-4ebb-a97a-cad157b9bb64"
	gapB96Refrain    = "1773c4e8-dc08-482a-9f26-30ef96c5b200"
	gapB96Nanogene   = "c06d1127-a269-453f-9c06-e9c424998eb7"
	gapB96Tardis     = "31b72bcc-2f45-4ff8-8e4b-27c020ff0e32"
	gapB96Traverse   = "ac5c2975-6a56-42c7-97c5-6a0e75c000bb"
	gapB96Twice      = "04dc2e9c-312c-4af8-950c-6a7ed859c9be"
	gapB96Wibbly     = "5207621d-aecf-40ae-8108-2bd571e3a427"
	gapB96Crisis     = "d5862292-9ef9-4bf2-b468-45bfe0d70f7a"
	gapB96Dust       = "39acc3d0-85fc-4814-93f2-e709c3b2c86b"
	gapB96Wedding    = "1ff91cc1-2233-4ceb-91ba-6a49c7147f32"
)

func gapB96TimeCounters(g *game.Game, id uuid.UUID) int {
	c, ok := battlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.Counters[game.CounterTime]
}

func gapB96Exiled(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

// gapB96Hand puts a card with the oracle id into the active seat's hand.
func gapB96Hand(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return id
}

func TestS82GapB96EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB96Flux: "Impending Flux", gapB96AllHistory: "All of History, All at Once", gapB96Refrain: "Inspiring Refrain",
		gapB96Nanogene: "Nanogene Conversion", gapB96Tardis: "Start the TARDIS", gapB96Traverse: "Traverse Eternity",
		gapB96Twice: "Twice Upon a Time", gapB96Twice + "#1": "Unlikely Meeting", gapB96Wibbly: "Wibbly-wobbly, Timey-wimey",
		gapB96Crisis: "Crisis of Conscience", gapB96Dust: "Everything Comes to Dust", gapB96Wedding: "The Wedding of River Song",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec)
		}
	}
}

// --- Impending Flux ----------------------------------------------------------------

func TestS82GapB96ImpendingFluxDealsOnePlusTheSpellsCastFromOutsideYourHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 5)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 5)
	exiled := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: exiled, Name: "Exiled Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		g.GrantCastPermissionOverCardForEffect(exiled, game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Label: "test"})
	})
	if err := g.CastSpell(me.ID, exiled, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	life := opp.Life
	castCatalogSpell(t, g, "Impending Flux", "Sorcery", gapB96Flux, nil)
	passPriorityAroundTable(t, g)
	if got := bnsMarked(g, theirs); got != 2 {
		t.Errorf("their creature was dealt %d, want 2 (1 plus one spell cast from exile)", got)
	}
	if opp.Life != life-2 {
		t.Errorf("their life %d, want %d", opp.Life, life-2)
	}
	if bnsMarked(g, mine) != 0 {
		t.Error("my own creature was dealt damage")
	}
}

func TestS82GapB96ImpendingFluxIsOneDamageWhenEverythingCameFromYourHand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 5)
	castCatalogSpell(t, g, "Impending Flux", "Sorcery", gapB96Flux, nil)
	passPriorityAroundTable(t, g)
	if got := bnsMarked(g, theirs); got != 1 {
		t.Errorf("dealt %d, want 1", got)
	}
}

// --- All of History, All at Once / Wibbly-wobbly ------------------------------------------

func TestS82GapB96AllOfHistoryTimeTravelsAndStormCopiesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	clock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Clock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterTime: 1}})
	// One spell cast before it this turn: storm copies it once, so it time travels twice.
	castCatalogSpell(t, g, "Filler", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "All of History, All at Once", "Sorcery", gapB96AllHistory, nil)
	for i := 0; i < 12; i++ {
		passPriorityAroundTable(t, g)
		if latestChoiceOfKind(g, game.PendingChoiceOptionPick) == nil {
			if stackFullyEmpty(g) {
				break
			}
			continue
		}
		answerOptionPick(t, g, me.ID, 0) // add a time counter
	}
	if got := gapB96TimeCounters(g, clock); got != 3 {
		t.Errorf("the Clock has %d time counters, want 3 (two time travels: the spell and its storm copy)", got)
	}
}

func TestS82GapB96WibblyWobblyTimeTravelsThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushLibraryTyped(me, "Top", "Basic Land — Forest")
	clock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Clock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterTime: 1}})
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Wibbly-wobbly, Timey-wimey", "Sorcery", gapB96Wibbly, nil)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if got := gapB96TimeCounters(g, clock); got != 2 {
		t.Errorf("the Clock has %d time counters, want 2", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- Inspiring Refrain ------------------------------------------------------------------

func TestS82GapB96InspiringRefrainDrawsTwoThenSuspendsItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	for i := 0; i < 3; i++ {
		pushLibraryTyped(me, "Filler", "Basic Land — Forest")
	}
	hand := me.Hand.Size()
	id := castCatalogSpell(t, g, "Inspiring Refrain", "Sorcery", gapB96Refrain, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	c, ok := gapB96Exiled(g, id)
	if !ok || c.Counters[game.CounterTime] != 3 {
		t.Fatalf("the spell: in exile=%v counters=%v, want 3 time counters", ok, c.Counters)
	}
	if !game.CardHasSuspend(c) {
		t.Error("the exiled card is not suspended")
	}
}

// --- Nanogene Conversion ------------------------------------------------------------------

func TestS82GapB96NanogeneConversionMakesEveryOtherCreatureACopyThatIsntLegendary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source := b12Creature(g, me.ID, "The Original", "Legendary Creature — Alien Warrior", 5, 5)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	spec, _ := Lookup(gapB96Nanogene)
	rfPermAExpectLegal(t, g, me.ID, spec.Targets, map[uuid.UUID]bool{source: true, mine: true, theirs: false})
	castCatalogSpell(t, g, "Nanogene Conversion", "Sorcery", gapB96Nanogene, cardTarget(source))
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"mine": mine, "theirs": theirs} {
		c, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatalf("%s left the battlefield", name)
		}
		eff := c.Effective()
		if eff.Name != "The Original" || eff.Power != 5 || eff.Toughness != 5 {
			t.Errorf("%s is %s %d/%d, want a copy of The Original", name, eff.Name, eff.Power, eff.Toughness)
		}
		for _, s := range eff.Supertypes {
			if s == "Legendary" {
				t.Errorf("%s is still legendary", name)
			}
		}
	}
	if c, _ := battlefieldCard(g, source); c.Effective().Name != "The Original" {
		t.Error("the chosen creature changed")
	}
	advanceToPrecombatMainOf(t, g, 1)
	if c, _ := battlefieldCard(g, mine); c.Effective().Name != "My Bear" {
		t.Errorf("the copy outlasted the turn: %s", c.Effective().Name)
	}
}

// --- Start the TARDIS --------------------------------------------------------------------

func TestS82GapB96StartTheTardisSurveilsThenDrawsAndJumpStarts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	for i := 0; i < 4; i++ {
		pushLibraryTyped(me, "Filler", "Basic Land — Forest")
	}
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Start the TARDIS", "Sorcery", gapB96Tardis, nil)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	// Jump-start from the graveyard, discarding a card.
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Start the TARDIS", TypeLine: "Sorcery", OracleID: gapB96Tardis, ManaCost: "{1}{U}", Owner: me.ID, Controller: me.ID})
	spec, _ := Lookup(gapB96Tardis)
	if len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Key != "jump-start" {
		t.Errorf("alternative costs %+v, want the jump-start offer", spec.AlternativeCosts)
	}
}

// --- Traverse Eternity ------------------------------------------------------------------

func TestS82GapB96TraverseEternityDrawsTheGreatestHistoricManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	for i := 0; i < 12; i++ {
		pushLibraryTyped(me, "Filler", "Basic Land — Forest")
	}
	blbB1Push(g, me.ID, "Cheap Relic", "Artifact", "", 0, 0)
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, findBattlefieldByName(g, "Cheap Relic")).ManaCost = "{2}" })
	legend := b12Creature(g, me.ID, "Big Legend", "Legendary Creature — Bear", 2, 2)
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, legend).ManaCost = "{4}{G}" })
	plain := b12Creature(g, me.ID, "Huge Plain Bear", "Creature — Bear", 9, 9)
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, plain).ManaCost = "{9}" })
	theirs := b12Creature(g, opp.ID, "Their Legend", "Legendary Creature — Bear", 2, 2)
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, theirs).ManaCost = "{8}" })
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Traverse Eternity", "Sorcery", gapB96Traverse, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 5 {
		t.Errorf("drew %d, want 5 (the legend's mana value; not the plain bear's 9 or their legend's 8)", got)
	}
}

// --- Twice Upon a Time // Unlikely Meeting --------------------------------------------------

func TestS82GapB96TwiceUponATimeNeedsTwoDoctorsThenTakesAnExtraTurnAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Creature(g, me.ID, "Doctor One", "Legendary Creature — Time Lord Doctor", 2, 2)
	id := gapB96Hand(me, "Twice Upon a Time", "Sorcery", gapB96Twice)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with only one Doctor")
	}
	b12Creature(g, me.ID, "Doctor Two", "Legendary Creature — Time Lord Doctor", 2, 2)
	before := len(g.ExtraTurnsQueued())
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast with two Doctors: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(g.ExtraTurnsQueued()); got != before+1 {
		t.Errorf("%d extra turns queued, want %d", got, before+1)
	}
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("the spell did not exile itself")
	}
}

func TestS82GapB96UnlikelyMeetingFindsADoctor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	doctor := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: doctor, Name: "A Doctor", TypeLine: "Legendary Creature — Time Lord Doctor", Owner: me.ID, Controller: me.ID})
	pushLibraryTyped(me, "Not A Doctor", "Basic Land — Forest")
	card := game.Card{
		InstanceID: uuid.New(), OracleID: gapB96Twice, Layout: game.LayoutAdventure, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Twice Upon a Time", TypeLine: "Sorcery", ManaCost: "{4}{U}{U}", Colors: []string{"U"}},
			{Name: "Unlikely Meeting", TypeLine: "Sorcery — Adventure", ManaCost: "{2}{U}", Colors: []string{"U"}},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	floatMana(t, g, me, "{C}{C}{U}")
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Strict: true}); err != nil {
		t.Fatalf("cast Unlikely Meeting: %v", err)
	}
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(doctor) {
		t.Error("Unlikely Meeting did not fetch the Doctor card")
	}
}

// --- Crisis of Conscience ----------------------------------------------------------------

func TestS82GapB96CrisisOfConscienceDestroysAllTokensOrAllNonlandNontokenPermanents(t *testing.T) {
	for mode, want := range map[int]struct{ token, creature, artifact, land bool }{
		0: {token: false, creature: true, artifact: true, land: true},
		1: {token: true, creature: false, artifact: false, land: true},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		var tokenIDs []uuid.UUID
		g.WithWriteLock(func() {
			tokenIDs, _ = g.CreateTokensForEffect(me.ID, TokenCard("1/1 white Soldier"), 1, game.TokenEntryOptions{})
		})
		token := tokenIDs[0]
		creature := b12Creature(g, me.ID, "A Bear", "Creature — Bear", 2, 2)
		artifact := b12Permanent(g, g.Seats[1].ID, "A Rock", "Artifact")
		land := b12Permanent(g, me.ID, "A Forest", "Basic Land — Forest")
		castCatalogSpellWithModes(t, g, "Crisis of Conscience", "Sorcery", gapB96Crisis, []int{mode})
		passPriorityAroundTable(t, g)
		for name, c := range map[string]struct {
			id   uuid.UUID
			kept bool
		}{
			"token": {token, want.token}, "creature": {creature, want.creature}, "artifact": {artifact, want.artifact}, "land": {land, want.land},
		} {
			if got := g.Battlefield.Contains(c.id); got != c.kept {
				t.Errorf("mode %d: %s on the battlefield = %v, want %v", mode, name, got, c.kept)
			}
		}
	}
}

// --- Everything Comes to Dust --------------------------------------------------------------

func TestS82GapB96EverythingComesToDustSparesCreaturesSharingATypeWithAConvoker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	convoker := b12Creature(g, me.ID, "Convoking Elf", "Creature — Elf Warrior", 1, 1)
	sharesWarrior := b12Creature(g, opp.ID, "Their Warrior", "Creature — Human Warrior", 2, 2)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	myBear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	golem := b12Creature(g, opp.ID, "Artifact Elf", "Artifact Creature — Elf", 2, 2)
	rock := b12Permanent(g, opp.ID, "A Rock", "Artifact")
	ench := b12Permanent(g, opp.ID, "An Aura", "Enchantment — Aura")
	land := b12Permanent(g, opp.ID, "A Forest", "Basic Land — Forest")
	id := gapB96Hand(me, "Everything Comes to Dust", "Sorcery", gapB96Dust)
	g.WithWriteLock(func() { me.Hand.Cards[len(me.Hand.Cards)-1].ManaCost = "{7}{W}{W}{W}" })
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{W}{W}{W}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: []uuid.UUID{convoker}}); err != nil {
		t.Fatalf("cast with convoke: %v", err)
	}
	passPriorityAroundTable(t, g)
	for name, c := range map[string]struct {
		id   uuid.UUID
		kept bool
	}{
		"the convoker": {convoker, true}, "a Warrior sharing a type": {sharesWarrior, true}, "an unrelated Bear": {bear, false},
		"my own unrelated Bear": {myBear, false}, "an artifact creature": {golem, false}, "an artifact": {rock, false},
		"an enchantment": {ench, false}, "a land": {land, true},
	} {
		if got := g.Battlefield.Contains(c.id); got != c.kept {
			t.Errorf("%s on the battlefield = %v, want %v", name, got, c.kept)
		}
	}
}

// --- The Wedding of River Song ----------------------------------------------------------------

func TestS82GapB96WeddingDrawsExilesFromBothHandsWithSuspendAndTimeTravels(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	for i := 0; i < 3; i++ {
		pushLibraryTyped(me, "Filler", "Basic Land — Forest")
	}
	mine := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mine, Name: "My Three-Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Owner: me.ID, Controller: me.ID})
	theirs := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: theirs, Name: "Their Two-Drop", TypeLine: "Sorcery", ManaCost: "{1}{R}", Owner: opp.ID, Controller: opp.ID})
	theirLand := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: theirLand, Name: "Their Land", TypeLine: "Basic Land — Mountain", Owner: opp.ID, Controller: opp.ID})
	castCatalogSpell(t, g, "The Wedding of River Song", "Sorcery", gapB96Wedding, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	for name, want := range map[string]struct {
		id uuid.UUID
		n  int
	}{"mine": {mine, 3}, "theirs": {theirs, 2}} {
		c, ok := gapB96Exiled(g, want.id)
		if !ok || c.Counters[game.CounterTime] != want.n || !game.CardHasSuspend(c) {
			t.Errorf("%s: in exile=%v time counters=%v suspend=%v, want %d counters and suspend", name, ok, c.Counters, game.CardHasSuspend(c), want.n)
		}
	}
	if !opp.Hand.Contains(theirLand) {
		t.Error("the opponent's land left their hand")
	}
	// Time travel now asks about the caster's suspended card(s).
	if latestChoiceOfKind(g, game.PendingChoiceOptionPick) == nil {
		t.Fatal("the caster did not time travel after both exiles")
	}
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if c, _ := gapB96Exiled(g, mine); c.Counters[game.CounterTime] != 4 {
		t.Errorf("my suspended card has %d time counters after time travel, want 4", c.Counters[game.CounterTime])
	}
}

// In a Planechase game Start the TARDIS offers the planeswalk, and a yes takes it.
func TestS82GapB96StartTheTardisOffersAPlaneswalkInPlanechase(t *testing.T) {
	g, _ := newPlaneGame(t, planeAmysHome)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 4; i++ {
		pushLibraryTyped(me, "Filler", "Basic Land — Forest")
	}
	if _, ok := faceUpPlane(g, planeAmysHome.Oracle); !ok {
		t.Fatal("Amy's Home is not the face-up plane to start")
	}
	castCatalogSpell(t, g, "Start the TARDIS", "Sorcery", gapB96Tardis, nil)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	confirm := pr7bPrompt(t, g, game.PendingChoiceConfirm)
	if err := g.ResolveConfirm(confirm.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := faceUpPlane(g, planeAmysHome.Oracle); ok {
		t.Error("the plane did not change after saying yes")
	}
}
