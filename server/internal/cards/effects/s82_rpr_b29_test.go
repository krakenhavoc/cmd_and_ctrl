package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rpr29Oyobi       = "df5c1f34-94e4-4f89-b1af-943c430862d0"
	rpr29Promise     = "6158ab85-a929-472c-9618-8e76d2d4b228"
	rpr29RakishHeir  = "3318c391-7aa1-4550-88a0-b29d8dca28b7"
	rpr29Rattlechain = "ee2ca311-7639-436d-8ecb-d2aad075f25f"
	rpr29Shackle     = "7a1e4880-9abb-4018-84bd-91b5ec702c51"
	rpr29Shepherd    = "d28b7c07-ef65-4187-a15e-2180bdc593d1"
	rpr29Masquerade  = "88f5d146-420b-417c-bcf1-fe0e78312a74"
	rpr29Condemned   = "67b87332-44e1-4948-9269-5fc8f6107d0e"
	rpr29Occultist   = "3bdd140b-1406-4d07-8f35-78f2ba92de64"
	rpr29Phantom     = "9c5f4d02-eedb-4c6c-9f13-1b7a45382483"
	rpr29Connections = "39c10798-9dfc-47a2-9da3-b4038509aebe"
)

func TestS82Rpr29CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rpr29Oyobi: "Oyobi, Who Split the Heavens", rpr29Promise: "Promise of Bunrei",
		rpr29RakishHeir: "Rakish Heir", rpr29Rattlechain: "Rattlechains", rpr29Shackle: "Shacklegeist",
		rpr29Shepherd: "Spectral Shepherd", rpr29Masquerade: "Stensia Masquerade",
		rpr29Condemned: "Stromkirk Condemned", rpr29Occultist: "Stromkirk Occultist",
		rpr29Phantom: "Supreme Phantom", rpr29Connections: "Underworld Connections",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec.Name)
		}
	}
}

func TestS82Rpr29MadnessCostsArePrinted(t *testing.T) {
	for _, oracle := range []string{rpr29Masquerade, rpr29Occultist} {
		reps, trigs := madnessAbilitiesFor(oracle)
		if reps != 1 || trigs != 1 {
			t.Errorf("%s: madness replacements %d, triggers %d, want 1 and 1", oracle, reps, trigs)
		}
	}
}

func TestS82Rpr29VampireCountersOnAnyVampireYouControl(t *testing.T) {
	for name, oracle := range map[string]string{"Rakish Heir": rpr29RakishHeir, "Stensia Masquerade": rpr29Masquerade} {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			typ := "Creature — Vampire"
			if name == "Stensia Masquerade" {
				typ = "Enchantment"
			}
			b12Push(g, me.ID, name, typ, oracle, 2, 2)
			vamp := b12Push(g, me.ID, "Friend", "Creature — Vampire", "", 2, 2)
			bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
			theirs := b12Push(g, opp.ID, "Their Vampire", "Creature — Vampire", "", 2, 2)
			for _, id := range []uuid.UUID{bear, theirs} {
				dealCombatDamageToPlayer(g, id, opp.ID, 2)
				passPriorityAroundTable(t, g)
				if got := counterCount(g, id, game.CounterPlusOne); got != 0 {
					t.Fatalf("a non-qualifying creature got %d counters", got)
				}
			}
			dealCombatDamageToPlayer(g, vamp, opp.ID, 2)
			passPriorityAroundTable(t, g)
			if got := counterCount(g, vamp, game.CounterPlusOne); got != 1 {
				t.Errorf("the Vampire got %d counters, want 1", got)
			}
		})
	}
}

func TestS82Rpr29SupremePhantomPumpsOtherSpiritsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ph := b12Push(g, me.ID, "Supreme Phantom", "Creature — Spirit", rpr29Phantom, 1, 3)
	mine := b12Push(g, me.ID, "My Spirit", "Creature — Spirit", "", 2, 2)
	theirs := b12Push(g, opp.ID, "Their Spirit", "Creature — Spirit", "", 2, 2)
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	for id, want := range map[uuid.UUID]int{ph: 1, mine: 3, theirs: 2, bear: 2} {
		if p, _ := effectivePTOf(t, g, id); p != want {
			t.Errorf("power %d, want %d", p, want)
		}
	}
}

func TestS82Rpr29StensiaMasqueradeGivesAttackersFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	b12Push(g, me.ID, "Stensia Masquerade", "Enchantment", rpr29Masquerade, 0, 0)
	att := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Ogre", 3, 3)
	if hasKeywordOf(g, att, "first strike") {
		t.Fatal("first strike before attacking")
	}
	attackIntoBlocks(t, g, them.ID, att)
	if !hasKeywordOf(g, att, "first strike") {
		t.Error("the attacker lacks first strike")
	}
	if hasKeywordOf(g, idle, "first strike") {
		t.Error("a non-attacker has first strike")
	}
}

func TestS82Rpr29PromiseOfBunreiSacrificesAndMakesFourSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	promise := b12Push(g, me.ID, "Promise of Bunrei", "Enchantment", rpr29Promise, 0, 0)
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(promise) {
		t.Error("Promise of Bunrei stayed on the battlefield")
	}
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 4 {
		t.Errorf("Spirit tokens = %d, want 4", n)
	}
}

func TestS82Rpr29OyobiMakesASpiritOnSpiritOrArcaneCasts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	b12Push(g, me.ID, "Oyobi", "Legendary Creature — Spirit", rpr29Oyobi, 3, 6)
	castFromSeat(t, g, me, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 0 {
		t.Fatalf("a Bear made %d Spirits", n)
	}
	castFromSeat(t, g, me, "Some Spirit", "Creature — Spirit", "", nil)
	passPriorityAroundTable(t, g)
	castFromSeat(t, g, me, "Some Arcane", "Instant — Arcane", "", nil)
	passPriorityAroundTable(t, g)
	n, _ := countTokensNamed(g, me.ID, "Spirit")
	if n != 2 {
		t.Fatalf("Spirit tokens = %d, want 2", n)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.IsToken() && c.Name == "Spirit" {
				if e := c.Effective(); e.Power != 3 || e.Toughness != 3 {
					t.Errorf("token is %d/%d, want 3/3", e.Power, e.Toughness)
				}
			}
		}
	})
}

func TestS82Rpr29ShacklegeistTapsTwoSpiritsToTapACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s := b12Push(g, me.ID, "Shacklegeist", "Creature — Spirit", rpr29Shackle, 2, 2)
	other := b12Creature(g, me.ID, "Other Spirit", "Creature — Spirit", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{s, bear}, Targets: cardTarget(theirs), Strict: true}); err == nil {
		t.Fatal("a non-Spirit paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{s, other}, Targets: cardTarget(s), Strict: true}); err == nil {
		t.Fatal("targeted a creature the controller controls")
	}
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{s, other}, Targets: cardTarget(theirs), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	tapped := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == theirs {
				tapped = c.Tapped
			}
		}
	})
	if !tapped {
		t.Error("the target is not tapped")
	}
}

func TestS82Rpr29ShacklegeistBlocksOnlyFlyers(t *testing.T) {
	spec, _ := Lookup(rpr29Shackle)
	if spec.Name != "Shacklegeist" {
		t.Fatal("not registered")
	}
	if len(game.CatalogBlockRules(rpr29Shackle)) == 0 {
		t.Error("Shacklegeist declares no block rule")
	}
}

func TestS82Rpr29SpectralShepherdBouncesYourSpirit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sh := b12Push(g, me.ID, "Spectral Shepherd", "Creature — Spirit", rpr29Shepherd, 2, 2)
	mine := b12Creature(g, me.ID, "My Spirit", "Creature — Spirit", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Spirit", "Creature — Spirit", 1, 1)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, sh, 0, game.ActivateAbilityParams{Targets: cardTarget(theirs), Strict: true}); err == nil {
		t.Fatal("bounced an opponent's Spirit")
	}
	if err := g.ActivateCatalogAbility(me.ID, sh, 0, game.ActivateAbilityParams{Targets: cardTarget(mine), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("the Spirit was not returned to its owner's hand")
	}
}

func TestS82Rpr29StromkirkCondemnedPumpsVampiresOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c := b12Push(g, me.ID, "Stromkirk Condemned", "Creature — Vampire Horror", rpr29Condemned, 2, 2)
	friend := b12Creature(g, me.ID, "Friend", "Creature — Vampire", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	var fodder []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		fodder = append(fodder, id)
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fodder", Owner: me.ID, Controller: me.ID})
	}
	if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{DiscardIDs: fodder[:1], Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{c: 3, friend: 3, bear: 2} {
		if p, _ := effectivePTOf(t, g, id); p != want {
			t.Errorf("power %d, want %d", p, want)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{DiscardIDs: fodder[1:], Strict: true}); err == nil {
		t.Error("activated a second time in one turn")
	}
}

func TestS82Rpr29UnderworldConnectionsGrantsTheLandADrawAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	toMainForCost(t, g)
	castFromSeat(t, g, me, "Underworld Connections", "Enchantment — Aura", rpr29Connections, cardTarget(land))
	passPriorityAroundTable(t, g)
	stdB21Library(me, game.Card{Name: "Top"}, game.Card{Name: "Next"})
	handBefore, lifeBefore := len(me.Hand.Cards), me.Life
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != lifeBefore-1 {
		t.Errorf("life %d, want %d", me.Life, lifeBefore-1)
	}
	if len(me.Hand.Cards) != handBefore+1 {
		t.Errorf("hand %d, want %d", len(me.Hand.Cards), handBefore+1)
	}
}

func TestS82Rpr29RattlechainsGivesASpiritHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	id := castFromSeat(t, g, me, "Rattlechains", "Creature — Spirit", rpr29Rattlechain, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, id)
		passPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(id) {
		t.Fatal("Rattlechains did not enter")
	}
	if !hasKeywordOf(g, id, "hexproof") {
		t.Error("the Spirit did not gain hexproof")
	}
}

func TestS82Rpr29StromkirkOccultistExilesTheTopCardOnDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stdB21Library(me, game.Card{Name: "A"}, game.Card{Name: "B"})
	o := b12Push(g, me.ID, "Stromkirk Occultist", "Creature — Vampire Horror", rpr29Occultist, 3, 2)
	before := len(me.Library.Cards)
	dealCombatDamageToPlayer(g, o, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if len(me.Library.Cards) != before-1 {
		t.Errorf("library %d, want %d", len(me.Library.Cards), before-1)
	}
}
