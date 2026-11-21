package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	top := effectstest.TopLibraryCard(me, "Held Card", "Sorcery")

	researcher := effectstest.CastCatalogSpell(t, g, "Goblin Researcher", "Creature — Goblin Wizard", effectstest.GoblinResearcherOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card is exiled when the Researcher enters")
	}
	perm := effectstest.ExiledPermission(g, top)
	if perm.Condition == nil || perm.Condition.Key != game.CondAttackedWithObject {
		t.Fatalf("%+v: the grant is conditional on attacking with the Researcher", perm)
	}
	if effectstest.PermissionLive(g, perm, me.ID) {
		t.Error("playable the turn it was exiled, before the Researcher attacked")
	}

	effectstest.NextTurnOfMine(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, top), me.ID) {
		t.Fatal("another creature attacked and opened it")
	}

	effectstest.NextTurnOfMine(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, researcher)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, top), me.ID) {
		t.Error("the Researcher attacked and the card is still shut")
	}
}

// --- Boros Strike-Captain ----------------------------------------------

func TestBorosStrikeCaptainBattalionExilesAndThreeAttackersOpenIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	captain := effectstest.B12Push(g, me.ID, "Boros Strike-Captain", "Creature — Minotaur Soldier", effectstest.BorosStrikeCaptainOracle, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Alpha", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bravo", "Creature — Bear", 2, 2)
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Held Card", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 1", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 2", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 3", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 4", TypeLine: "Sorcery"},
	)

	// The Captain and ONE other creature: not a battalion.
	before := me.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, captain, a)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != before {
		t.Fatal("two attackers triggered battalion")
	}

	effectstest.NextTurnOfMine(t, g)
	held := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	effectstest.DeclareAttack(t, g, opp.ID, captain, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(held) {
		t.Fatal("the Captain and two others attacked and nothing was exiled")
	}
	if perm := effectstest.ExiledPermission(g, held); !effectstest.PermissionLive(g, perm, me.ID) {
		t.Fatalf("%+v: three attackers opened the card", perm)
	}

	// A later turn with only two attackers leaves it dormant.
	effectstest.NextTurnOfMine(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	perm := effectstest.ExiledPermission(g, held)
	if perm.Player != me.ID {
		t.Fatalf("the grant did not persist: %+v", perm)
	}
	if effectstest.PermissionLive(g, perm, me.ID) {
		t.Error("two attackers opened a three-attacker card")
	}
}

// --- Robber of the Rich -------------------------------------------------

func TestRobberOfTheRichNeedsAMoreFullHandAndARogue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	robber := effectstest.B12Push(g, me.ID, "Robber of the Rich", "Creature — Human Archer Rogue", effectstest.RobberOfTheRichOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// A turn's setup: my hand empty, the opponent's holding three, and a
	// known card on top of their library.
	setUp := func() uuid.UUID {
		g.WithWriteLock(func() { me.Hand.Cards = nil; opp.Hand.Cards = nil })
		for i := 0; i < 3; i++ {
			opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		}
		return effectstest.TopLibraryCard(opp, "Their Bolt", "Instant")
	}

	// Equal hands: the intervening if fails, nothing is exiled.
	theirs := setUp()
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	effectstest.DeclareAttack(t, g, opp.ID, robber)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(theirs) {
		t.Fatal("the defender had no more cards in hand and a card was exiled")
	}

	// Only a Bear attacks: Robber's trigger is Robber's own attack.
	effectstest.NextTurnOfMine(t, g)
	theirs = setUp()
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(theirs) {
		t.Fatal("a Bear attacked and Robber's trigger fired")
	}

	effectstest.NextTurnOfMine(t, g)
	theirs = setUp()
	effectstest.DeclareAttack(t, g, opp.ID, robber)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Fatal("the defender has more cards in hand and nothing was exiled")
	}
	perm := effectstest.ExiledPermission(g, theirs)
	if perm.Player != me.ID || !perm.CastOnly || !perm.AnyColor {
		t.Fatalf("%+v: the Robber's controller may CAST it, spending mana as any color", perm)
	}
	if !effectstest.PermissionLive(g, perm, me.ID) {
		t.Error("a Rogue attacked this turn and the card is shut")
	}

	// A later turn on which only the Bear attacks: dormant.
	effectstest.NextTurnOfMine(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, theirs), me.ID) {
		t.Error("a turn only a non-Rogue attacked opened the card")
	}
}

// --- Pick Up the Pace ---------------------------------------------------

func TestPickUpThePaceIsLiveOnlyWhileAttackedAndWhileItRemains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pace := effectstest.B12Push(g, me.ID, "Pick Up the Pace", "Enchantment", effectstest.PickUpThePaceOracle, 0, 0)
	effectstest.TopLibraryCard(me, "Filler", "Sorcery")
	effectstest.TopLibraryCard(me, "Filler", "Sorcery")
	held := effectstest.TopLibraryCard(me, "Held Card", "Sorcery")
	fresh := effectstest.B12Creature(g, me.ID, "Fresh Bear", "Creature — Bear", 2, 2)
	// pushBattlefieldCardWithTimestamp emits the zone move only; the
	// entry tally "entered this turn" reads is fed by the ETB event.
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: fresh, Actor: me.ID}) })

	effectstest.DeclareAttack(t, g, opp.ID, fresh)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(held) {
		t.Fatal("a creature that entered this turn attacked and nothing was exiled")
	}
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, held), me.ID) {
		t.Fatal("playable the turn its holder attacked")
	}

	// Next turn, before anything attacks: dormant but granted.
	effectstest.NextTurnOfMine(t, g)
	perm := effectstest.ExiledPermission(g, held)
	if perm.Player != me.ID || effectstest.PermissionLive(g, perm, me.ID) {
		t.Fatalf("%+v: dormant until an attack", perm)
	}
	effectstest.DeclareAttack(t, g, opp.ID, fresh)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, held), me.ID) {
		t.Fatal("attacked this turn and the card is shut")
	}

	// The enchantment leaves: the cards "exiled with it" are no longer
	// playable, even on a turn the player attacked.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pace) })
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, held), me.ID) {
		t.Error("the card stayed playable after Pick Up the Pace left the battlefield")
	}
}

// --- Neriv, Crackling Vanguard -----------------------------------------

func TestNerivExilesOnePerDifferentlyNamedTokenAndNeedsACommanderToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neriv := effectstest.B12Push(g, me.ID, "Neriv, Crackling Vanguard", "Legendary Creature — Spirit Dragon", effectstest.NerivCracklingVanguardOracl, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B32Token(g, me.ID, "Goblin", 1, 1)
	effectstest.B32Token(g, me.ID, "Goblin", 1, 1)
	effectstest.B32Token(g, me.ID, "Soldier", 1, 1)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "One", TypeLine: "Sorcery"},
		game.Card{Name: "Two", TypeLine: "Sorcery"},
		game.Card{Name: "Three", TypeLine: "Sorcery"},
		game.Card{Name: "Filler", TypeLine: "Sorcery"},
	)

	effectstest.DeclareAttack(t, g, opp.ID, neriv)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || !g.Exile.Contains(ids[1]) || g.Exile.Contains(ids[2]) {
		t.Fatal("two differently named tokens (Goblin, Soldier) exile exactly two cards")
	}
	// Neriv is not (yet) a commander: attacking with her opens nothing.
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, ids[0]), me.ID) {
		t.Error("a non-commander attack opened the cards")
	}

	effectstest.NextTurnOfMine(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, ids[0]), me.ID) {
		t.Error("a Bear attacked and opened the cards")
	}

	setCommander(g, neriv)
	effectstest.NextTurnOfMine(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, neriv)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:2] {
		if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, id), me.ID) {
			t.Error("a commander attacked and the cards are shut")
		}
	}
}

// --- Possibility Technician ----------------------------------------------

func TestPossibilityTechnicianCardIsPlayableOnlyWhileYouControlAKavu(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := effectstest.TopLibraryCard(me, "Held Card", "Sorcery")

	tech := effectstest.CastCatalogSpell(t, g, "Possibility Technician", "Creature — Kavu Artificer", effectstest.PossibilityTechnicianOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the Technician entering exiled nothing")
	}
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, top), me.ID) {
		t.Fatal("the Technician is a Kavu and the card is shut")
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tech) })
	perm := effectstest.ExiledPermission(g, top)
	if perm.Player != me.ID {
		t.Fatalf("the grant did not persist: %+v", perm)
	}
	if effectstest.PermissionLive(g, perm, me.ID) {
		t.Error("no Kavu left and the card is still playable")
	}

	effectstest.B12Creature(g, me.ID, "Another Kavu", "Creature — Kavu", 1, 1)
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, top), me.ID) {
		t.Error("a Kavu came back and the card stayed shut")
	}
}

// --- Lightning, Security Sergeant ----------------------------------------

func TestLightningCardIsPlayableForAsLongAsYouControlHer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lightning := effectstest.B12Push(g, me.ID, "Lightning, Security Sergeant", "Legendary Creature — Human Soldier", effectstest.LightningSecuritySergeantOr, 2, 3)
	top := effectstest.TopLibraryCard(me, "Held Card", "Sorcery")

	effectstest.AttackWith(t, g, opp.ID, lightning)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("combat damage to a player exiled nothing")
	}
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, top), me.ID) {
		t.Fatal("not playable while its controller controls Lightning")
	}
	// Still playable on later turns, with no condition beyond her.
	effectstest.NextTurnOfMine(t, g)
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, top), me.ID) {
		t.Error("the grant lapsed with the turn; it lasts as long as you control her")
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(lightning) })
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, top), me.ID) {
		t.Error("playable after Lightning left the battlefield")
	}
}

// --- Victor Mancha, Runaway ------------------------------------------------

func TestVictorManchaGraveyardCardIsPlayableForAsLongAsYouControlHim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bolt := effectstest.SeedGraveyardCard(t, g, "Filler Bolt", "Instant", "test-victor-bolt")

	victor := effectstest.CastCatalogSpell(t, g, "Victor Mancha, Runaway", "Legendary Artifact Creature — Human Hero", effectstest.VictorManchaRunawayOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(bolt) {
		t.Fatal("the targeted graveyard card was not exiled")
	}
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, bolt), me.ID) {
		t.Fatal("not playable while its controller controls Victor Mancha")
	}
	effectstest.NextTurnOfMine(t, g)
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, bolt), me.ID) {
		t.Error("the grant lapsed with the turn")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victor) })
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, bolt), me.ID) {
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			top := effectstest.TopLibraryCard(me, tc.name, tc.typeLine)
			land := effectstest.PushStampedCatalogCard(g, me.ID, "Nivix, Aerie of the Firemind", "Land", effectstest.NivixAerieOracle)
			effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
			me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "R"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
			if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate Nivix: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Exile.Contains(top) {
				t.Fatal("the top card was not exiled")
			}
			live := effectstest.PermissionLive(g, effectstest.ExiledPermission(g, top), me.ID)
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
