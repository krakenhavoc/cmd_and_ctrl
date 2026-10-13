package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conditional_exile_cards_test.go — #3192, ADR 0066's second
// 2026-10-13 amendment. One test per card that exiles with a
// conditional or source-bound permission. The engine model is pinned in
// game/cast_permission_condition_test.go; these pin that each card says
// the right thing with it.

const (
	borosStrikeCaptainOracle    = "30b20932-0d9a-447f-b934-1daa8c44a678"
	goblinResearcherOracle      = "c8762dfa-a027-4e1c-a2fa-bfdab1de102a"
	robberOfTheRichOracle       = "4df0d381-fc05-4d30-9307-48963aa2ecfd"
	pickUpThePaceOracle         = "873e739c-edf4-46a7-9024-5caaf87f507e"
	nerivCracklingVanguardOracl = "c8f827d7-4740-43c8-b494-c9e0479e0bf7"
	possibilityTechnicianOracle = "ebbf84f7-d142-4de5-bfc1-ea86f082d733"
	lightningSecuritySergeantOr = "e1626880-0fa3-441a-9978-b293fc87c418"
	victorManchaRunawayOracle   = "606a83ae-c774-4abc-ba30-c113cbd1c05b"
	nivixAerieOracle            = "9c482f1d-08b4-4882-918c-448a556d3fbe"
)

// nextTurnOfMine rolls the table round to the controller's next turn.
func nextTurnOfMine(t *testing.T, g *game.Game) {
	t.Helper()
	advanceToMainOf(t, g, 1)
	advanceToMainOf(t, g, 0)
}

// topLibraryCard puts a recognisable card on top of the library and
// returns its ID.
func topLibraryCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

func setCommander(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
}

// --- Goblin Researcher ------------------------------------------------

func TestGoblinResearcherCardIsPlayableOnlyInTurnsItAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	top := topLibraryCard(me, "Held Card", "Sorcery")

	researcher := castCatalogSpell(t, g, "Goblin Researcher", "Creature — Goblin Wizard", goblinResearcherOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card is exiled when the Researcher enters")
	}
	perm := exiledPermission(g, top)
	if perm.Condition == nil || perm.Condition.Key != game.CondAttackedWithObject {
		t.Fatalf("%+v: the grant is conditional on attacking with the Researcher", perm)
	}
	if permissionLive(g, perm, me.ID) {
		t.Error("playable the turn it was exiled, before the Researcher attacked")
	}

	nextTurnOfMine(t, g)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if permissionLive(g, exiledPermission(g, top), me.ID) {
		t.Fatal("another creature attacked and opened it")
	}

	nextTurnOfMine(t, g)
	declareAttack(t, g, opp.ID, researcher)
	passPriorityAroundTable(t, g)
	if !permissionLive(g, exiledPermission(g, top), me.ID) {
		t.Error("the Researcher attacked and the card is still shut")
	}
}

// --- Boros Strike-Captain ----------------------------------------------

func TestBorosStrikeCaptainBattalionExilesAndThreeAttackersOpenIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	captain := b12Push(g, me.ID, "Boros Strike-Captain", "Creature — Minotaur Soldier", borosStrikeCaptainOracle, 3, 3)
	a := b12Creature(g, me.ID, "Alpha", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bravo", "Creature — Bear", 2, 2)
	seedSearchLibrary(me,
		game.Card{Name: "Held Card", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 1", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 2", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 3", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 4", TypeLine: "Sorcery"},
	)

	// The Captain and ONE other creature: not a battalion.
	before := me.Library.Size()
	declareAttack(t, g, opp.ID, captain, a)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != before {
		t.Fatal("two attackers triggered battalion")
	}

	nextTurnOfMine(t, g)
	held := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	declareAttack(t, g, opp.ID, captain, a, b)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(held) {
		t.Fatal("the Captain and two others attacked and nothing was exiled")
	}
	if perm := exiledPermission(g, held); !permissionLive(g, perm, me.ID) {
		t.Fatalf("%+v: three attackers opened the card", perm)
	}

	// A later turn with only two attackers leaves it dormant.
	nextTurnOfMine(t, g)
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	perm := exiledPermission(g, held)
	if perm.Player != me.ID {
		t.Fatalf("the grant did not persist: %+v", perm)
	}
	if permissionLive(g, perm, me.ID) {
		t.Error("two attackers opened a three-attacker card")
	}
}

// --- Robber of the Rich -------------------------------------------------

func TestRobberOfTheRichNeedsAMoreFullHandAndARogue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	robber := b12Push(g, me.ID, "Robber of the Rich", "Creature — Human Archer Rogue", robberOfTheRichOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// A turn's setup: my hand empty, the opponent's holding three, and a
	// known card on top of their library.
	setUp := func() uuid.UUID {
		g.WithWriteLock(func() { me.Hand.Cards = nil; opp.Hand.Cards = nil })
		for i := 0; i < 3; i++ {
			opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		}
		return topLibraryCard(opp, "Their Bolt", "Instant")
	}

	// Equal hands: the intervening if fails, nothing is exiled.
	theirs := setUp()
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	declareAttack(t, g, opp.ID, robber)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(theirs) {
		t.Fatal("the defender had no more cards in hand and a card was exiled")
	}

	// Only a Bear attacks: Robber's trigger is Robber's own attack.
	nextTurnOfMine(t, g)
	theirs = setUp()
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(theirs) {
		t.Fatal("a Bear attacked and Robber's trigger fired")
	}

	nextTurnOfMine(t, g)
	theirs = setUp()
	declareAttack(t, g, opp.ID, robber)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Fatal("the defender has more cards in hand and nothing was exiled")
	}
	perm := exiledPermission(g, theirs)
	if perm.Player != me.ID || !perm.CastOnly || !perm.AnyColor {
		t.Fatalf("%+v: the Robber's controller may CAST it, spending mana as any color", perm)
	}
	if !permissionLive(g, perm, me.ID) {
		t.Error("a Rogue attacked this turn and the card is shut")
	}

	// A later turn on which only the Bear attacks: dormant.
	nextTurnOfMine(t, g)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if permissionLive(g, exiledPermission(g, theirs), me.ID) {
		t.Error("a turn only a non-Rogue attacked opened the card")
	}
}

// --- Pick Up the Pace ---------------------------------------------------

func TestPickUpThePaceIsLiveOnlyWhileAttackedAndWhileItRemains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pace := b12Push(g, me.ID, "Pick Up the Pace", "Enchantment", pickUpThePaceOracle, 0, 0)
	topLibraryCard(me, "Filler", "Sorcery")
	topLibraryCard(me, "Filler", "Sorcery")
	held := topLibraryCard(me, "Held Card", "Sorcery")
	fresh := b12Creature(g, me.ID, "Fresh Bear", "Creature — Bear", 2, 2)
	// pushBattlefieldCardWithTimestamp emits the zone move only; the
	// entry tally "entered this turn" reads is fed by the ETB event.
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: fresh, Actor: me.ID}) })

	declareAttack(t, g, opp.ID, fresh)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(held) {
		t.Fatal("a creature that entered this turn attacked and nothing was exiled")
	}
	if !permissionLive(g, exiledPermission(g, held), me.ID) {
		t.Fatal("playable the turn its holder attacked")
	}

	// Next turn, before anything attacks: dormant but granted.
	nextTurnOfMine(t, g)
	perm := exiledPermission(g, held)
	if perm.Player != me.ID || permissionLive(g, perm, me.ID) {
		t.Fatalf("%+v: dormant until an attack", perm)
	}
	declareAttack(t, g, opp.ID, fresh)
	passPriorityAroundTable(t, g)
	if !permissionLive(g, exiledPermission(g, held), me.ID) {
		t.Fatal("attacked this turn and the card is shut")
	}

	// The enchantment leaves: the cards "exiled with it" are no longer
	// playable, even on a turn the player attacked.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pace) })
	if permissionLive(g, exiledPermission(g, held), me.ID) {
		t.Error("the card stayed playable after Pick Up the Pace left the battlefield")
	}
}

// --- Neriv, Crackling Vanguard -----------------------------------------

func TestNerivExilesOnePerDifferentlyNamedTokenAndNeedsACommanderToAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neriv := b12Push(g, me.ID, "Neriv, Crackling Vanguard", "Legendary Creature — Spirit Dragon", nerivCracklingVanguardOracl, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b32Token(g, me.ID, "Goblin", 1, 1)
	b32Token(g, me.ID, "Goblin", 1, 1)
	b32Token(g, me.ID, "Soldier", 1, 1)
	ids := seedSearchLibrary(me,
		game.Card{Name: "One", TypeLine: "Sorcery"},
		game.Card{Name: "Two", TypeLine: "Sorcery"},
		game.Card{Name: "Three", TypeLine: "Sorcery"},
		game.Card{Name: "Filler", TypeLine: "Sorcery"},
	)

	declareAttack(t, g, opp.ID, neriv)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || !g.Exile.Contains(ids[1]) || g.Exile.Contains(ids[2]) {
		t.Fatal("two differently named tokens (Goblin, Soldier) exile exactly two cards")
	}
	// Neriv is not (yet) a commander: attacking with her opens nothing.
	if permissionLive(g, exiledPermission(g, ids[0]), me.ID) {
		t.Error("a non-commander attack opened the cards")
	}

	nextTurnOfMine(t, g)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if permissionLive(g, exiledPermission(g, ids[0]), me.ID) {
		t.Error("a Bear attacked and opened the cards")
	}

	setCommander(g, neriv)
	nextTurnOfMine(t, g)
	declareAttack(t, g, opp.ID, neriv)
	passPriorityAroundTable(t, g)
	for _, id := range ids[:2] {
		if !permissionLive(g, exiledPermission(g, id), me.ID) {
			t.Error("a commander attacked and the cards are shut")
		}
	}
}

// --- Possibility Technician ----------------------------------------------

func TestPossibilityTechnicianCardIsPlayableOnlyWhileYouControlAKavu(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := topLibraryCard(me, "Held Card", "Sorcery")

	tech := castCatalogSpell(t, g, "Possibility Technician", "Creature — Kavu Artificer", possibilityTechnicianOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the Technician entering exiled nothing")
	}
	if !permissionLive(g, exiledPermission(g, top), me.ID) {
		t.Fatal("the Technician is a Kavu and the card is shut")
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tech) })
	perm := exiledPermission(g, top)
	if perm.Player != me.ID {
		t.Fatalf("the grant did not persist: %+v", perm)
	}
	if permissionLive(g, perm, me.ID) {
		t.Error("no Kavu left and the card is still playable")
	}

	b12Creature(g, me.ID, "Another Kavu", "Creature — Kavu", 1, 1)
	if !permissionLive(g, exiledPermission(g, top), me.ID) {
		t.Error("a Kavu came back and the card stayed shut")
	}
}

// --- Lightning, Security Sergeant ----------------------------------------

func TestLightningCardIsPlayableForAsLongAsYouControlHer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lightning := b12Push(g, me.ID, "Lightning, Security Sergeant", "Legendary Creature — Human Soldier", lightningSecuritySergeantOr, 2, 3)
	top := topLibraryCard(me, "Held Card", "Sorcery")

	attackWith(t, g, opp.ID, lightning)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("combat damage to a player exiled nothing")
	}
	if !permissionLive(g, exiledPermission(g, top), me.ID) {
		t.Fatal("not playable while its controller controls Lightning")
	}
	// Still playable on later turns, with no condition beyond her.
	nextTurnOfMine(t, g)
	if !permissionLive(g, exiledPermission(g, top), me.ID) {
		t.Error("the grant lapsed with the turn; it lasts as long as you control her")
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(lightning) })
	if permissionLive(g, exiledPermission(g, top), me.ID) {
		t.Error("playable after Lightning left the battlefield")
	}
}

// --- Victor Mancha, Runaway ------------------------------------------------

func TestVictorManchaGraveyardCardIsPlayableForAsLongAsYouControlHim(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bolt := seedGraveyardCard(t, g, "Filler Bolt", "Instant", "test-victor-bolt")

	victor := castCatalogSpell(t, g, "Victor Mancha, Runaway", "Legendary Artifact Creature — Human Hero", victorManchaRunawayOracle, nil)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(bolt) {
		t.Fatal("the targeted graveyard card was not exiled")
	}
	if !permissionLive(g, exiledPermission(g, bolt), me.ID) {
		t.Fatal("not playable while its controller controls Victor Mancha")
	}
	nextTurnOfMine(t, g)
	if !permissionLive(g, exiledPermission(g, bolt), me.ID) {
		t.Error("the grant lapsed with the turn")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victor) })
	if permissionLive(g, exiledPermission(g, bolt), me.ID) {
		t.Error("playable after Victor Mancha left the battlefield")
	}
}

// --- Nivix, Aerie of the Firemind --------------------------------------------

func TestNivixOnlyOpensAnInstantOrSorcery(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		castable       bool
	}{
		{"Held Bolt", "Instant", true},
		{"Held Divination", "Sorcery", true},
		{"Held Bear", "Creature — Bear", false},
		{"Held Forest", "Basic Land — Forest", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			top := topLibraryCard(me, tc.name, tc.typeLine)
			land := pushStampedCatalogCard(g, me.ID, "Nivix, Aerie of the Firemind", "Land", nivixAerieOracle)
			advanceToStepInTurn(t, g, game.StepPrecombatMain)
			me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "R"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
			if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate Nivix: %v", err)
			}
			passPriorityAroundTable(t, g)
			if !g.Exile.Contains(top) {
				t.Fatal("the top card was not exiled")
			}
			live := permissionLive(g, exiledPermission(g, top), me.ID)
			if live != tc.castable {
				t.Errorf("%s: playable = %v, want %v", tc.typeLine, live, tc.castable)
			}
			if tc.castable {
				if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile"}); err != nil {
					t.Errorf("cast from exile: %v", err)
				}
			}
		})
	}
}
