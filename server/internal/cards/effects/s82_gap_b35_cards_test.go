package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB35HiddenBladeOracle  = "37ec9b95-8c9a-4dd5-9ff2-3d6527984126"
	gapB35HuntersBowOracle   = "810592fe-65f5-4c95-9d75-d6806e004903"
	gapB35PhantomBladeOracle = "58d5d563-cb60-4977-86cd-f32cd811aa77"
	gapB35HemlockVialOracle  = "0d977dc5-807b-4fab-b44b-ccad5e3d63e3"
	gapB35SmokeBombOracle    = "c262e015-f379-484f-894f-1a1deb3c75cc"
	gapB35StaffOfEdenOracle  = "c4fe63e7-bb4c-4aff-b74f-d385d0128def"
	gapB35MjolnirOracle      = "f497e9a5-0ad3-4aec-8422-14be8e0dba4c"
	gapB35AdrestiaOracle     = "94779ca4-c202-4163-9c77-1036921a34bc"
	gapB35YggdrasilOracle    = "7617a6f7-c83b-4ee5-a681-c7e164456526"
)

func TestGapB35EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB35HiddenBladeOracle:  "Hidden Blade",
		gapB35HuntersBowOracle:   "Hunter's Bow",
		gapB35PhantomBladeOracle: "Phantom Blade",
		gapB35HemlockVialOracle:  "Hemlock Vial",
		gapB35SmokeBombOracle:    "Smoke Bomb",
		gapB35StaffOfEdenOracle:  "Staff of Eden, Vault's Key",
		gapB35MjolnirOracle:      "Mjölnir, Storm Hammer",
		gapB35AdrestiaOracle:     "Adrestia",
		gapB35YggdrasilOracle:    "Yggdrasil, Rebirth Engine",
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

func gapB35Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestGapB35HiddenBladeAttachesAndGivesAnAssassinDeathtouch(t *testing.T) {
	for _, tc := range []struct {
		typeLine   string
		deathtouch bool
	}{{"Creature — Human Assassin", true}, {"Creature — Bear", false}} {
		g := newCatalogGame(t)
		me, _ := gapB35Seats(g)
		c := b12Creature(g, me.ID, "Host", tc.typeLine, 2, 2)
		blade := castCatalogSpell(t, g, "Hidden Blade", "Artifact — Equipment", gapB35HiddenBladeOracle, nil)
		passPriorityAroundTable(t, g)
		b04WaitForPick(t, g, me.ID)
		pickCard(t, g, me.ID, c)
		passPriorityAroundTable(t, g)
		if got := e2Card(t, g, blade).AttachedTo.ID; got != c {
			t.Fatalf("%s: attached to %v, want the host", tc.typeLine, got)
		}
		if p, _ := ptOf(t, g, c); p != 3 {
			t.Errorf("%s: power %d, want 3", tc.typeLine, p)
		}
		if !effectiveAbilitiesContain(t, g, c, "first strike") {
			t.Errorf("%s: no first strike", tc.typeLine)
		}
		if got := effectiveAbilitiesContain(t, g, c, "deathtouch"); got != tc.deathtouch {
			t.Errorf("%s: deathtouch = %v, want %v", tc.typeLine, got, tc.deathtouch)
		}
	}
}

func TestGapB35HuntersBowAttachesAndBites(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB35Seats(g)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 5)
	bow := castCatalogSpell(t, g, "Hunter's Bow", "Artifact — Equipment", gapB35HuntersBowOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, mine)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, victim)
	}
	passPriorityAroundTable(t, g)
	if got := e2Card(t, g, bow).AttachedTo.ID; got != mine {
		t.Fatalf("attached to %v, want my creature", got)
	}
	if got := e2Card(t, g, victim).DamageMarked; got != 3 {
		t.Errorf("victim has %d damage, want 3 (the bear's power)", got)
	}
	if !effectiveAbilitiesContain(t, g, mine, "reach") {
		t.Error("the equipped creature has no reach")
	}
}

func TestGapB35PhantomBladeAttachesAndDestroysAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB35Seats(g)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 5)
	blade := castCatalogSpell(t, g, "Phantom Blade", "Artifact — Equipment", gapB35PhantomBladeOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, mine)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, victim)
	}
	passPriorityAroundTable(t, g)
	if got := e2Card(t, g, blade).AttachedTo.ID; got != mine {
		t.Fatalf("attached to %v, want my creature", got)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the other creature survived")
	}
	if p, tough := ptOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("equipped creature is %d/%d, want 3/3", p, tough)
	}
	if !effectiveAbilitiesContain(t, g, mine, "menace") {
		t.Error("no menace")
	}
}

func TestGapB35HemlockVialDrawsLosesAndGivesEquippedThingsDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB35Seats(g)
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	hand, life := me.Hand.Size(), me.Life
	vial := castCatalogSpell(t, g, "Hemlock Vial", "Artifact", gapB35HemlockVialOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1 (the spell left it and one card was drawn)", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
	equipped := b12Creature(g, me.ID, "Equipped", "Creature — Bear", 2, 2)
	bare := b12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	sword := b12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	spare := b12Permanent(g, me.ID, "Spare", "Artifact — Equipment")
	theirSword := b12Permanent(g, opp.ID, "Their Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: equipped})
	})
	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, vial, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vial) {
		t.Error("the Vial was not sacrificed")
	}
	for name, want := range map[string]struct {
		id uuid.UUID
		ok bool
	}{"equipped creature": {equipped, true}, "attached equipment": {sword, true}, "unattached equipment": {spare, true},
		"bare creature": {bare, false}, "their creature": {theirs, false}, "their equipment": {theirSword, false}} {
		if got := effectiveAbilitiesContain(t, g, want.id, "deathtouch"); got != want.ok {
			t.Errorf("%s: deathtouch = %v, want %v", name, got, want.ok)
		}
	}
}

func TestGapB35SmokeBombGivesShroudAndSacrificesAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB35Seats(g)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	bomb := pushCatalogPermanent(g, me.ID, "Smoke Bomb", "Artifact", gapB35SmokeBombOracle, false)
	for _, id := range []uuid.UUID{mine, theirs} {
		if !effectiveAbilitiesContain(t, g, id, "shroud") {
			t.Errorf("creature %v has no shroud", id)
		}
	}
	// Walk to the controller's next upkeep.
	turn := g.Turn.Seq
	for i := 0; i < 80 && !(g.Turn.Seq != turn && g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 0); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Step != game.StepUpkeep || g.Turn.ActiveSeat != 0 {
		t.Fatalf("never reached my upkeep (at %v, seat %d)", g.Turn.Step, g.Turn.ActiveSeat)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bomb) {
		t.Fatal("the Smoke Bomb was not sacrificed at the upkeep")
	}
	for i := 0; i < 4 && latestPickTarget(g, me.ID) == nil; i++ {
		passPriorityAroundTable(t, g)
	}
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, mine)
	}
	passPriorityAroundTable(t, g)
	if restrictionsOf(t, g, mine)&game.CantBeBlocked == 0 {
		t.Error("the chosen creature can still be blocked")
	}
}

func TestGapB35StaffOfEdenReanimatesALegendaryPermanentAndDrawsForStolenOnes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB35Seats(g)
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	legend := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: legend, Name: "Legend", TypeLine: "Legendary Creature — Elf", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	plain := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: plain, Name: "Plain", TypeLine: "Creature — Elf", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Legendary Sorcery", TypeLine: "Legendary Sorcery", Owner: opp.ID, Controller: opp.ID})
	staff := castCatalogSpell(t, g, "Staff of Eden, Vault's Key", "Legendary Artifact", gapB35StaffOfEdenOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); p != nil {
		for _, bad := range []uuid.UUID{plain, spell} {
			if hasID(p.PickTargetCards, bad) {
				t.Errorf("%v is not a legendary permanent card but was offered", bad)
			}
		}
	}
	pickCard(t, g, me.ID, legend)
	passPriorityAroundTable(t, g)
	c := e2Card(t, g, legend)
	if !g.Battlefield.Contains(legend) || c.Controller != me.ID {
		t.Fatalf("the legendary card is not on my side of the battlefield (controller %v)", c.Controller)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1 (one permanent I control but don't own)", got)
	}
}

func TestGapB35MjolnirAttachesToALegendAndPunishesTappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB35Seats(g)
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	legend := b12Creature(g, me.ID, "Legend", "Legendary Creature — Elf", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	target := b12Creature(g, opp.ID, "Target", "Creature — Bear", 2, 2)
	alreadyTapped := b12Creature(g, opp.ID, "Tapped", "Creature — Bear", 2, 2)
	thirdTapped := b12Creature(g, third.ID, "Third", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{alreadyTapped, thirdTapped} {
			if c := findBattlefieldForTest(g, id); c != nil {
				c.Tapped = true
			}
		}
	})
	hammer := castCatalogSpell(t, g, "Mjölnir, Storm Hammer", "Legendary Artifact — Equipment", gapB35MjolnirOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); p != nil && hasID(p.PickTargetCards, plain) {
		t.Error("a nonlegendary creature was offered")
	}
	pickCard(t, g, me.ID, legend)
	passPriorityAroundTable(t, g)
	if got := e2Card(t, g, hammer).AttachedTo.ID; got != legend {
		t.Fatalf("attached to %v, want the legendary creature", got)
	}
	oppLife, thirdLife := opp.Life, third.Life
	declareAttack(t, g, opp.ID, legend)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, target)
	passPriorityAroundTable(t, g)
	tc := e2Card(t, g, target)
	if !tc.Tapped || tc.Counters[game.CounterStun] != 1 {
		t.Errorf("target tapped=%v stun=%d, want tapped with one stun counter", tc.Tapped, tc.Counters[game.CounterStun])
	}
	// The defender controls two tapped creatures (the target and the other one); the third seat one.
	if got := oppLife - opp.Life; got != 2 {
		t.Errorf("the defender took %d, want 2", got)
	}
	if got := thirdLife - third.Life; got != 1 {
		t.Errorf("the third opponent took %d, want 1", got)
	}
}

func findBattlefieldForTest(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func TestGapB35AdrestiaDrawsOnlyWhenAnAssassinCrewedIt(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		draws    bool
	}{{"Creature — Human Assassin", true}, {"Creature — Bear", false}} {
		g := newCatalogGame(t)
		toMain(t, g)
		me, opp := gapB35Seats(g)
		for i := 0; i < 3; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		}
		ship := pushVehicleForTest(g, me.ID, "Adrestia", gapB35AdrestiaOracle, 4, 3)
		crewer := b12Creature(g, me.ID, "Crew", tc.typeLine, 1, 1)
		crewBy(t, g, me.ID, ship, crewer)
		passPriorityAroundTable(t, g)
		hand := me.Hand.Size()
		declareAttack(t, g, opp.ID, ship)
		passPriorityAroundTable(t, g)
		got := me.Hand.Size() - hand
		if tc.draws && got != 1 || !tc.draws && got != 0 {
			t.Errorf("%s crew: drew %d, want draws=%v", tc.typeLine, got, tc.draws)
		}
		if isAssassin := e2Card(t, g, ship).HasSubtype("Assassin"); isAssassin != tc.draws {
			t.Errorf("%s crew: Adrestia is an Assassin = %v, want %v", tc.typeLine, isAssassin, tc.draws)
		}
	}
}

func TestGapB35YggdrasilExilesLinksAndReturnsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB35Seats(g)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Elf", TypeLine: "Creature — Elf", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Elf", TypeLine: "Creature — Elf", Owner: opp.ID, Controller: opp.ID})
	for i := 0; i < 4; i++ {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Lib", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	}
	ygg := castCatalogSpell(t, g, "Yggdrasil, Rebirth Engine", "Legendary Artifact", gapB35YggdrasilOracle, nil)
	passPriorityAroundTable(t, g)
	if inZone(me.Graveyard, dead) || !inZone(me.Graveyard, spell) || !inZone(opp.Graveyard, theirs) {
		t.Error("the entry trigger exiled the wrong cards")
	}
	linked := func() int {
		n := 0
		g.ReadSnapshot(func() {
			if c, ok := g.LookupCardForEffect(ygg); ok {
				n = len(g.ExiledWithForEffect(game.PermissionCardRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}))
			}
		})
		return n
	}
	if linked() != 1 {
		t.Fatalf("%d cards exiled with Yggdrasil, want 1", linked())
	}
	lib := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, ygg, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("tap ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lib - me.Library.Size(); got != 3 {
		t.Errorf("exiled %d cards from the library, want 3", got)
	}
	if linked() != 4 {
		t.Errorf("%d cards exiled with Yggdrasil, want 4", linked())
	}
	// Untap and return the dead elf: {4}, {T}, sorcery speed.
	g.WithWriteLock(func() {
		if c := findBattlefieldForTest(g, ygg); c != nil {
			c.Tapped = false
		}
	})
	fillPool(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, ygg, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("return ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choice of a creature card exiled with Yggdrasil")
	}
	if hasID(pick.ChooseCards, theirs) {
		t.Error("a card not exiled with Yggdrasil was offered")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{dead}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	returned := uuid.Nil
	for _, bc := range gapB35Battlefield(g) {
		if bc.Name == "Dead Elf" {
			returned = bc.InstanceID
		}
	}
	if returned == uuid.Nil {
		t.Fatal("the creature card did not enter the battlefield")
	}
	if got := e2Card(t, g, returned).Controller; got != me.ID {
		t.Errorf("it entered under %v, want me", got)
	}
	if !effectiveAbilitiesContain(t, g, returned, "haste") {
		t.Error("the returned creature has no haste")
	}
}

func gapB35Battlefield(g *game.Game) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() { out = g.BattlefieldCardsForEffect() })
	return out
}
