package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 — Innistrad: Crimson Vow, batch 16: cards exiled with a permanent, and
// cards in exile.

const (
	vowB16Desecrator   = "85e9a43e-5676-4d12-ac27-e8cf57307d5e"
	vowB16Gatekeeper   = "fa450433-85b0-4f95-bd4e-f141ba036bee"
	vowB16Illuminator  = "14fe4531-96cd-415b-aa4a-b41fd271d902"
	vowB16Protector    = "34c8fffd-1b23-4f40-ad29-1f0acb519e7e"
	vowB16Circle       = "d1cd7a93-4b83-4755-90bc-17cb64bfb131"
	vowB16Grolnok      = "53029460-50a0-4934-8fa9-0271517cd5d0"
	vowB16Umbris       = "af31febb-af53-4b48-b930-9e1f9e5d8f26"
	vowB16Dreadlight   = "068b9586-53e3-4791-a26c-a6b38d038f66"
	vowB16Scavenger    = "54731e3f-a84a-4fc9-8c02-2f769133c71c"
	vowB16Heirloom     = "4bf0427e-5e87-4d32-ad34-1d8946ab96fd"
	vowB16CreatureLine = "Creature — Bear"
)

func TestVowB16EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		vowB16Desecrator:  "Cemetery Desecrator",
		vowB16Gatekeeper:  "Cemetery Gatekeeper",
		vowB16Illuminator: "Cemetery Illuminator",
		vowB16Protector:   "Cemetery Protector",
		vowB16Circle:      "Circle of Confinement",
		vowB16Grolnok:     "Grolnok, the Omnivore",
		vowB16Umbris:      "Umbris, Fear Manifest",
		vowB16Dreadlight:  "Dreadlight Monstrosity",
		vowB16Scavenger:   "Diregraf Scavenger",
		vowB16Heirloom:    "Honored Heirloom",
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

// vowB16Exiled puts a card owned by `owner` into exile.
func vowB16Exiled(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost, Owner: owner, Controller: owner})
	return id
}

// vowB16OnBattlefield is the permanent with that id, or nil.
func vowB16OnBattlefield(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

// vowB16Activate activates ability `index` of a permanent for `who`.
func vowB16Activate(t *testing.T, g *game.Game, who, id uuid.UUID, index int, targets ...uuid.UUID) error {
	t.Helper()
	toMainForCost(t, g)
	return g.ActivateCatalogAbility(who, id, index, game.ActivateAbilityParams{Targets: cardRefs(targets...), Strict: true})
}

// --- Honored Heirloom ----------------------------------------------------

func TestVowB16HonoredHeirloomExilesACardFromAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b27Push(g, me.ID, "Honored Heirloom", "Artifact", vowB16Heirloom, "{3}", 0, 0)
	mine := b38Graveyard(g, me.ID, "Mine", vowB16CreatureLine, "{1}", 1, 1)
	theirs := b38Graveyard(g, opp.ID, "Theirs", "Instant", "{R}", 0, 0)
	toMainForCost(t, g)
	fillPool(me, 2)
	if err := vowB16Activate(t, g, me.ID, rock, 0, theirs); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) || opp.Graveyard.Contains(theirs) {
		t.Error("the targeted card was not exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("an untargeted card left a graveyard")
	}
	if c := vowB16OnBattlefield(g, rock); c == nil || !c.Tapped {
		t.Error("the Heirloom did not tap")
	}
}

func TestVowB16HonoredHeirloomNeedsTwoManaAndAGraveyardCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b27Push(g, me.ID, "Honored Heirloom", "Artifact", vowB16Heirloom, "{3}", 0, 0)
	card := b38Graveyard(g, me.ID, "Mine", vowB16CreatureLine, "{1}", 1, 1)
	fillPool(me, 1)
	if err := vowB16Activate(t, g, me.ID, rock, 0, card); err == nil {
		t.Error("activated with one mana")
	}
	fillPool(me, 1)
	if err := vowB16Activate(t, g, me.ID, rock, 0, uuid.New()); err == nil {
		t.Error("activated on a card that is in no graveyard")
	}
}

func TestVowB16HonoredHeirloomTapsForAnyColor(t *testing.T) {
	spec, _ := Lookup(vowB16Heirloom)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|U|B|R|G}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Errorf("mana abilities %+v, want one {T}: any colour", spec.ManaAbilities)
	}
}

// --- Diregraf Scavenger --------------------------------------------------

func vowB16CastScavenger(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Diregraf Scavenger", "Creature — Zombie Bear", vowB16Scavenger, nil)
	passPriorityAroundTable(t, g)
	if target != uuid.Nil {
		pickCard(t, g, g.Seats[0].ID, target)
		passPriorityAroundTable(t, g)
	}
}

func TestVowB16DiregrafScavengerDrainsForACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b38Graveyard(g, opp.ID, "Bear", vowB16CreatureLine, "{1}", 2, 2)
	lifeMe, lifeOpp, lifeThird := me.Life, opp.Life, g.Seats[2].Life
	vowB16CastScavenger(t, g, bear)
	if !g.Exile.Contains(bear) {
		t.Fatal("the creature card was not exiled")
	}
	if me.Life != lifeMe+2 || opp.Life != lifeOpp-2 || g.Seats[2].Life != lifeThird-2 {
		t.Errorf("life %d/%d/%d from %d/%d/%d, want +2 for you, -2 each opponent", me.Life, opp.Life, g.Seats[2].Life, lifeMe, lifeOpp, lifeThird)
	}
}

func TestVowB16DiregrafScavengerIgnoresANoncreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := b38Graveyard(g, opp.ID, "Bolt", "Instant", "{R}", 0, 0)
	lifeMe, lifeOpp := me.Life, opp.Life
	vowB16CastScavenger(t, g, bolt)
	if !g.Exile.Contains(bolt) {
		t.Fatal("the instant was not exiled")
	}
	if me.Life != lifeMe || opp.Life != lifeOpp {
		t.Errorf("life changed to %d/%d for a noncreature card", me.Life, opp.Life)
	}
}

func TestVowB16DiregrafScavengerWithNoTargetDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b38Graveyard(g, opp.ID, "Bear", vowB16CreatureLine, "{1}", 2, 2)
	lifeMe := me.Life
	castCatalogSpell(t, g, "Diregraf Scavenger", "Creature — Zombie Bear", vowB16Scavenger, nil)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{}); err != nil {
			t.Logf("declining the target: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if me.Life != lifeMe {
		t.Errorf("life changed to %d with no card exiled", me.Life)
	}
}

// --- Dreadlight Monstrosity ----------------------------------------------

func TestVowB16DreadlightNeedsACardYouOwnInExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crab := b27Push(g, me.ID, "Dreadlight Monstrosity", "Creature — Crab Horror", vowB16Dreadlight, "{4}{U}{U}", 5, 5, "U")
	toMainForCost(t, g)
	fillPoolColored(me, "U", 5)
	vowB16Exiled(g, opp.ID, "Theirs", vowB16CreatureLine, "{1}")
	if err := vowB16Activate(t, g, me.ID, crab, 0); err == nil {
		t.Fatal("activated with only an opponent's card in exile")
	}
	vowB16Exiled(g, me.ID, "Mine", vowB16CreatureLine, "{1}")
	if err := vowB16Activate(t, g, me.ID, crab, 0); err != nil {
		t.Fatalf("activate with an owned card in exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	var unblockable bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == crab {
				unblockable = c.Effective().Restrictions&game.CantBeBlocked != 0
			}
		}
	})
	if !unblockable {
		t.Error("the Monstrosity can still be blocked")
	}
}

func TestVowB16DreadlightHasWardTwo(t *testing.T) {
	spec, _ := Lookup(vowB16Dreadlight)
	if len(spec.Triggered) != 1 || spec.Triggered[0].Key != "Dreadlight Monstrosity — ward {2}" {
		t.Errorf("triggers %+v, want ward {2}", spec.Triggered)
	}
}

// --- the Cemetery creatures ---------------------------------------------

// vowB16Enters casts a creature as the active seat, lets its enters trigger
// ask for the exile, and answers with `pick`; uuid.Nil answers nothing.
func vowB16Enters(t *testing.T, g *game.Game, name, typeLine, oracle string, pick uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := castCatalogSpell(t, g, name, typeLine, oracle, nil)
	passPriorityAroundTable(t, g)
	if pick == uuid.Nil {
		return id
	}
	p := latestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatalf("%s asked for no card", name)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{pick}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func vowB16CountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

func TestVowB16ProtectorExilesAnyCardLinkedToItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b38Graveyard(g, me.ID, "Mine", "Instant", "{R}", 0, 0)
	theirs := b38Graveyard(g, opp.ID, "Theirs", "Land", "", 0, 0)
	id := castCatalogSpell(t, g, "Cemetery Protector", "Creature — Human Soldier", vowB16Protector, nil)
	passPriorityAroundTable(t, g)
	p := latestChooseCardsFor(g, me.ID)
	if p == nil || len(p.ChooseCards) != 2 {
		t.Fatalf("prompt %+v, want a pick among both graveyards' cards", p)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{mine, theirs}); err == nil {
		t.Error("two cards were accepted for one exile")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) || !me.Graveyard.Contains(mine) {
		t.Fatal("the picked card was not the one exiled")
	}
	self := vowB16OnBattlefield(g, id)
	if self == nil {
		t.Fatal("Protector is not on the battlefield")
	}
	if got := vowB16ExiledWithCard(g, self, cemeteryProtectorExileLabel); len(got) != 1 || got[0].InstanceID != theirs {
		t.Errorf("exiled with Protector: %v, want the picked card", got)
	}
}

func TestVowB16ProtectorMakesAHumanForASharedTypeOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := b38Graveyard(g, me.ID, "Bolt", "Instant", "{R}", 0, 0)
	vowB16Enters(t, g, "Cemetery Protector", "Creature — Human Soldier", vowB16Protector, dead)
	if n := vowB16CountNamed(g, me.ID, "Human"); n != 0 {
		t.Fatalf("%d Humans before anything was cast", n)
	}
	// A creature spell shares no type with an instant.
	castCatalogSpell(t, g, "Bear", vowB16CreatureLine, "", nil)
	passPriorityAroundTable(t, g)
	if n := vowB16CountNamed(g, me.ID, "Human"); n != 0 {
		t.Fatalf("%d Humans after a creature spell", n)
	}
	// An instant does, and a land does not.
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if n := vowB16CountNamed(g, me.ID, "Human"); n != 1 {
		t.Fatalf("%d Humans after an instant spell, want 1", n)
	}
	b31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	if n := vowB16CountNamed(g, me.ID, "Human"); n != 1 {
		t.Errorf("%d Humans after a land play, want still 1", n)
	}
}

func TestVowB16ProtectorReadsLandPlays(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := b38Graveyard(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	vowB16Enters(t, g, "Cemetery Protector", "Creature — Human Soldier", vowB16Protector, dead)
	b31PlayLand(t, g, "Island", "Basic Land — Island", "")
	passPriorityAroundTable(t, g)
	if n := vowB16CountNamed(g, me.ID, "Human"); n != 1 {
		t.Errorf("%d Humans after playing a land, want 1", n)
	}
}

func TestVowB16ProtectorIgnoresOpponentsPlays(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := b38Graveyard(g, me.ID, "Bolt", "Instant", "{R}", 0, 0)
	vowB16Enters(t, g, "Cemetery Protector", "Creature — Human Soldier", vowB16Protector, dead)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Shock", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("opponent cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := vowB16CountNamed(g, me.ID, "Human"); n != 0 {
		t.Errorf("%d Humans from an opponent's spell", n)
	}
}

func TestVowB16ProtectorWithEmptyGraveyardsAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vowB16Enters(t, g, "Cemetery Protector", "Creature — Human Soldier", vowB16Protector, uuid.Nil)
	if p := latestChooseCardsFor(g, me.ID); p != nil {
		t.Error("asked to pick from empty graveyards")
	}
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if n := vowB16CountNamed(g, me.ID, "Human"); n != 0 {
		t.Errorf("%d Humans with nothing exiled", n)
	}
}

func TestVowB16GatekeeperHitsAnyPlayerWhoSharesAType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := b38Graveyard(g, me.ID, "Bolt", "Instant", "{R}", 0, 0)
	vowB16Enters(t, g, "Cemetery Gatekeeper", "Creature — Vampire", vowB16Gatekeeper, dead)
	// Its own controller is hit too.
	lifeMe := me.Life
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != lifeMe-2 {
		t.Errorf("life %d, want %d after casting an instant", me.Life, lifeMe-2)
	}
	// So is an opponent.
	lifeOpp := opp.Life
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Zap", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("opponent cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != lifeOpp-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, lifeOpp-2)
	}
	// A creature spell shares nothing with an instant.
	lifeMe = me.Life
	castCatalogSpell(t, g, "Bear", vowB16CreatureLine, "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != lifeMe {
		t.Errorf("life %d after a creature spell, want %d", me.Life, lifeMe)
	}
}

func TestVowB16GatekeeperHitsLandPlaysToo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := b38Graveyard(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	vowB16Enters(t, g, "Cemetery Gatekeeper", "Creature — Vampire", vowB16Gatekeeper, dead)
	lifeMe := me.Life
	b31PlayLand(t, g, "Island", "Basic Land — Island", "")
	passPriorityAroundTable(t, g)
	if me.Life != lifeMe-2 {
		t.Errorf("life %d, want %d after a land play", me.Life, lifeMe-2)
	}
}

func TestVowB16GatekeeperHasFirstStrike(t *testing.T) {
	spec, _ := Lookup(vowB16Gatekeeper)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "first strike" {
		t.Errorf("keywords %v", spec.PrintedKeywords)
	}
}

// vowB16Top puts a card on top of a seat's library.
func vowB16Top(g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	})
	return id
}

func TestVowB16IlluminatorCastsOnlyASharedTypeFromTheTopOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := b38Graveyard(g, me.ID, "Bolt", "Instant", "{R}", 0, 0)
	vowB16Enters(t, g, "Cemetery Illuminator", "Creature — Spirit", vowB16Illuminator, dead)
	// A creature on top shares no type with an instant.
	bear := vowB16Top(g, me, "Bear", vowB16CreatureLine)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("a creature spell was cast from the top of the library")
	}
	// An instant does.
	one := vowB16Top(g, me, "Shock", "Instant")
	if err := g.CastSpell(me.ID, one, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("the instant from the top: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Once each turn.
	two := vowB16Top(g, me, "Zap", "Instant")
	if err := g.CastSpell(me.ID, two, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Error("a second spell was cast from the top in one turn")
	}
}

func TestVowB16IlluminatorNeedsAnExiledCardAndNeverCastsALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := b38Graveyard(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	vowB16Enters(t, g, "Cemetery Illuminator", "Creature — Spirit", vowB16Illuminator, dead)
	spell := vowB16Top(g, me, "Shock", "Instant")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Error("an instant was cast with only a land exiled")
	}
}

func TestVowB16IlluminatorSharesAnyOfTheExiledCardsTypes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := b38Graveyard(g, me.ID, "Golem", "Artifact Creature — Golem", "{2}", 2, 2)
	vowB16Enters(t, g, "Cemetery Illuminator", "Creature — Spirit", vowB16Illuminator, dead)
	bear := vowB16Top(g, me, "Bear", vowB16CreatureLine)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("a creature shares a type with the artifact creature: %v", err)
	}
}

func TestVowB16IlluminatorExilesOnAttackToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := b27Push(g, me.ID, "Cemetery Illuminator", "Creature — Spirit", vowB16Illuminator, "{1}{U}{U}", 2, 3, "U")
	dead := b38Graveyard(g, opp.ID, "Bolt", "Instant", "{R}", 0, 0)
	declareAttack(t, g, opp.ID, id)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	p := latestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("no exile prompt on attack")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{dead}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(dead) {
		t.Error("the card was not exiled on attack")
	}
}

// --- Cemetery Desecrator -------------------------------------------------

func vowB16Desecrate(t *testing.T, g *game.Game, pick uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	p := latestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("no exile prompt")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{pick}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func vowB16PayAmountFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == game.PendingChoicePayAmount && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// vowB16Reflexive answers the reflexive trigger's mode, and its target when
// the mode has one.
func vowB16Reflexive(t *testing.T, g *game.Game, mode int, target uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt for the reflexive trigger")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if target != uuid.Nil {
		pickCard(t, g, me.ID, target)
	}
	passPriorityAroundTable(t, g)
}

func TestVowB16DesecratorShrinksAnOpponentsCreatureByTheManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b27Push(g, opp.ID, "Ogre", vowB16CreatureLine, "", "{3}", 4, 4)
	mine := b27Push(g, me.ID, "Mine", vowB16CreatureLine, "", "{1}", 4, 4)
	dead := b38Graveyard(g, opp.ID, "Three", "Instant", "{1}{R}{R}", 0, 0)
	castCatalogSpell(t, g, "Cemetery Desecrator", "Creature — Zombie", vowB16Desecrator, nil)
	passPriorityAroundTable(t, g)
	vowB16Desecrate(t, g, dead)
	if !g.Exile.Contains(dead) {
		t.Fatal("the card was not exiled")
	}
	if p := latestPickTarget(g, me.ID); p != nil && hasID(p.PickTargetCards, mine) {
		t.Fatal("the creature you control was offered for -X/-X")
	}
	vowB16Reflexive(t, g, 1, victim)
	if pw, tough := effectivePTOf(t, g, victim); pw != 1 || tough != 1 {
		t.Errorf("Ogre is %d/%d, want 1/1 (-3/-3)", pw, tough)
	}
	if pw, tough := effectivePTOf(t, g, mine); pw != 4 || tough != 4 {
		t.Errorf("own creature is %d/%d, want untouched", pw, tough)
	}
}

func TestVowB16DesecratorRemovesExactlyXCountersAcrossKinds(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{"+1/+1": 2, "charge": 5}})
	dead := b38Graveyard(g, opp.ID, "Three", "Instant", "{1}{R}{R}", 0, 0)
	castCatalogSpell(t, g, "Cemetery Desecrator", "Creature — Zombie", vowB16Desecrator, nil)
	passPriorityAroundTable(t, g)
	vowB16Desecrate(t, g, dead)
	vowB16Reflexive(t, g, 0, rock)
	first := vowB16PayAmountFor(g, me.ID)
	if first == nil || first.PayAmount.Min != 0 || first.PayAmount.Max != 2 {
		t.Fatalf("first question %+v, want +1/+1 counters 0..2", first)
	}
	if err := g.ResolvePayAmount(first.ID, me.ID, 1); err != nil {
		t.Fatalf("ResolvePayAmount: %v", err)
	}
	second := vowB16PayAmountFor(g, me.ID)
	if second == nil || second.PayAmount.Min != 2 || second.PayAmount.Max != 2 {
		t.Fatalf("second question %+v, want exactly the 2 still owed", second)
	}
	if err := g.ResolvePayAmount(second.ID, me.ID, 2); err != nil {
		t.Fatalf("ResolvePayAmount: %v", err)
	}
	passPriorityAroundTable(t, g)
	if a, b := countersOn(g, rock, "+1/+1"), countersOn(g, rock, "charge"); a != 1 || b != 3 {
		t.Errorf("counters %d +1/+1 and %d charge, want 1 and 3 (3 removed)", a, b)
	}
}

func TestVowB16DesecratorTakesEveryCounterWhenThereAreFewerThanX(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{"charge": 1}})
	dead := b38Graveyard(g, opp.ID, "Five", "Instant", "{3}{R}{R}", 0, 0)
	castCatalogSpell(t, g, "Cemetery Desecrator", "Creature — Zombie", vowB16Desecrator, nil)
	passPriorityAroundTable(t, g)
	vowB16Desecrate(t, g, dead)
	vowB16Reflexive(t, g, 0, rock)
	if q := vowB16PayAmountFor(g, g.Seats[0].ID); q != nil {
		if err := g.ResolvePayAmount(q.ID, g.Seats[0].ID, q.PayAmount.Min); err != nil {
			t.Fatalf("ResolvePayAmount: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if n := countersOn(g, rock, "charge"); n != 0 {
		t.Errorf("%d counters left, want all removed", n)
	}
}

func TestVowB16DesecratorExilesAnotherCardWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := b27Push(g, me.ID, "Cemetery Desecrator", "Creature — Zombie", vowB16Desecrator, "{4}{B}{B}", 4, 4, "B")
	victim := b27Push(g, opp.ID, "Ogre", vowB16CreatureLine, "", "{3}", 4, 4)
	dead := b38Graveyard(g, opp.ID, "Two", "Instant", "{1}{R}", 0, 0)
	killCreature(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	p := latestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("no exile prompt on death")
	}
	if hasID(p.ChooseCards, id) {
		t.Error("the Desecrator offered itself: it must be another card")
	}
	vowB16Desecrate(t, g, dead)
	vowB16Reflexive(t, g, 1, victim)
	if pw, _ := effectivePTOf(t, g, victim); pw != 2 {
		t.Errorf("Ogre power %d, want 2 (-2/-2)", pw)
	}
}

func TestVowB16DesecratorWithNoOtherCardDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b27Push(g, me.ID, "Cemetery Desecrator", "Creature — Zombie", vowB16Desecrator, "{4}{B}{B}", 4, 4, "B")
	killCreature(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, me.ID) != nil || modePickChoiceFor(g, me.ID) != nil {
		t.Error("asked something with only the Desecrator itself in a graveyard")
	}
}

// --- Circle of Confinement -----------------------------------------------

func TestVowB16CircleExilesACreatureOfManaValueThreeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b27Push(g, opp.ID, "Small", "Creature — Vampire", "", "{2}{B}", 2, 2)
	big := b27Push(g, opp.ID, "Big", vowB16CreatureLine, "", "{3}{B}", 4, 4)
	mine := b27Push(g, me.ID, "Mine", vowB16CreatureLine, "", "{1}", 1, 1)
	castCatalogSpell(t, g, "Circle of Confinement", "Enchantment", vowB16Circle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if hasID(p.PickTargetCards, big) || hasID(p.PickTargetCards, mine) || !hasID(p.PickTargetCards, small) {
		t.Fatalf("offered %v, want only the opponent's creature with mana value 3 or less", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(small) {
		t.Fatal("the creature was not exiled")
	}
}

func TestVowB16CircleReturnsTheCreatureWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b27Push(g, opp.ID, "Small", "Creature — Vampire", "", "{2}{B}", 2, 2)
	circle := castCatalogSpell(t, g, "Circle of Confinement", "Enchantment", vowB16Circle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	exileCreature(t, g, circle)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(small) {
		t.Error("the creature stayed in exile")
	}
	if findBattlefieldByName(g, "Small") == uuid.Nil {
		t.Error("the creature did not come back")
	}
}

func vowB16CircleHolding(t *testing.T, name, typeLine string) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, opp.ID, name, typeLine, "", "{2}{B}", 2, 2)
	castCatalogSpell(t, g, "Circle of Confinement", "Enchantment", vowB16Circle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, findBattlefieldByName(g, name))
	passPriorityAroundTable(t, g)
	return g, opp.ID
}

func vowB16OppCasts(t *testing.T, g *game.Game, opp uuid.UUID, name, typeLine string) {
	t.Helper()
	p := g.PlayerByID(opp)
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Keywords: []string{"flash"}, Owner: opp, Controller: opp})
	if err := g.CastSpell(opp, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
}

func TestVowB16CircleGainsLifeForAVampireSpellWithTheExiledName(t *testing.T) {
	g, opp := vowB16CircleHolding(t, "Same Name", "Creature — Vampire")
	me := g.Seats[0]
	life := me.Life
	// Another name: nothing.
	vowB16OppCasts(t, g, opp, "Other Name", "Creature — Vampire")
	if me.Life != life {
		t.Fatalf("life %d after a different name, want %d", me.Life, life)
	}
	// The same name, but not a Vampire spell: nothing.
	vowB16OppCasts(t, g, opp, "Same Name", "Instant")
	if me.Life != life {
		t.Fatalf("life %d after a non-Vampire spell of that name, want %d", me.Life, life)
	}
	vowB16OppCasts(t, g, opp, "Same Name", "Creature — Vampire")
	if me.Life != life+2 {
		t.Errorf("life %d after the matching Vampire spell, want %d", me.Life, life+2)
	}
}

func TestVowB16CircleIgnoresYourOwnVampireSpell(t *testing.T) {
	g, _ := vowB16CircleHolding(t, "Same Name", "Creature — Vampire")
	me := g.Seats[0]
	life := me.Life
	castCatalogSpell(t, g, "Same Name", "Creature — Vampire", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d after your own spell, want %d", me.Life, life)
	}
}

// --- Grolnok, the Omnivore -----------------------------------------------

func TestVowB16GrolnokMillsThreeWhenAFrogAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Grolnok, the Omnivore", "Legendary Creature — Frog", vowB16Grolnok, "{2}{G}{U}", 3, 3, "G", "U")
	frog := b27Push(g, me.ID, "Frog", "Creature — Frog", "", "{1}", 1, 1)
	toad := b27Push(g, me.ID, "Toad", "Creature — Toad", "", "{1}", 1, 1)
	before := me.Library.Size()
	declareAttack(t, g, opp.ID, toad)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != before {
		t.Fatalf("milled for a Toad: library %d, was %d", me.Library.Size(), before)
	}
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	b27Push(g2, me2.ID, "Grolnok, the Omnivore", "Legendary Creature — Frog", vowB16Grolnok, "{2}{G}{U}", 3, 3, "G", "U")
	frog2 := b27Push(g2, me2.ID, "Frog", "Creature — Frog", "", "{1}", 1, 1)
	before = me2.Library.Size()
	declareAttack(t, g2, opp2.ID, frog2)
	lockInAttacks(t, g2)
	passPriorityAroundTable(t, g2)
	if me2.Library.Size() != before-3 {
		t.Errorf("library %d, want %d after a Frog attacked", me2.Library.Size(), before-3)
	}
	_ = frog
}

func TestVowB16GrolnokExilesMilledPermanentCardsWithCroakCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Grolnok, the Omnivore", "Legendary Creature — Frog", vowB16Grolnok, "{2}{G}{U}", 3, 3, "G", "U")
	spell := vowB16Top(g, me, "Bolt", "Instant")
	thing := vowB16Top(g, me, "Rock", "Artifact")
	theirs := vowB16Top(g, opp, "TheirRock", "Artifact")
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 2) })
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(thing) {
		t.Fatal("the permanent card was not exiled")
	}
	if c := exiledCard(g, thing); c == nil || c.Counters["croak"] != 1 {
		t.Errorf("exiled card %+v, want one croak counter", c)
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the instant card should stay in the graveyard")
	}
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(theirs) {
		t.Error("an opponent's milled card was exiled")
	}
}

func TestVowB16GrolnokPlaysLandsAndCastsSpellsFromCroakedExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Grolnok, the Omnivore", "Legendary Creature — Frog", vowB16Grolnok, "{2}{G}{U}", 3, 3, "G", "U")
	land := vowB16Exiled(g, me.ID, "Forest", "Basic Land — Forest", "")
	bear := vowB16Exiled(g, me.ID, "Bear", vowB16CreatureLine, "")
	plain := vowB16Exiled(g, me.ID, "Plain", vowB16CreatureLine, "")
	theirs := vowB16Exiled(g, opp.ID, "Theirs", vowB16CreatureLine, "")
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{land, bear, theirs} {
			if c := exiledCard(g, id); c != nil {
				c.Counters = map[string]int{"croak": 1}
			}
		}
	})
	toMainForCost(t, g)
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("cast a card with no croak counter")
	}
	if err := g.CastSpell(opp.ID, theirs, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("cast a card the casting player does not own")
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast a croaked card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Bear") == uuid.Nil {
		t.Error("the croaked creature did not resolve")
	}
	var perm *game.CastPermission
	g.ReadSnapshot(func() { perm = g.CastPermissionOnCardByIDForEffect(land) })
	if perm == nil || perm.Zone != game.ZoneExile {
		t.Errorf("no permission to play the croaked land from exile: %+v", perm)
	}
}

// --- Umbris, Fear Manifest -----------------------------------------------

func TestVowB16UmbrisGrowsForEachCardOpponentsOwnInExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	umbris := b27Push(g, me.ID, "Umbris, Fear Manifest", "Legendary Creature — Nightmare Horror", vowB16Umbris, "{3}{U}{B}", 1, 1, "U", "B")
	if pw, tough := effectivePTOf(t, g, umbris); pw != 1 || tough != 1 {
		t.Fatalf("Umbris is %d/%d with an empty exile, want 1/1", pw, tough)
	}
	for _, owner := range []uuid.UUID{opp.ID, g.Seats[2].ID, me.ID} {
		id := b38Graveyard(g, owner, "Card", vowB16CreatureLine, "{1}", 1, 1)
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
			t.Fatalf("exile: %v", err)
		}
	}
	if pw, tough := effectivePTOf(t, g, umbris); pw != 3 || tough != 3 {
		t.Errorf("Umbris is %d/%d, want 3/3 for two opposing exiled cards", pw, tough)
	}
}

func TestVowB16UmbrisExilesFromTheTopUntilALand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	// The top is the last card pushed: Bear, Bolt, Forest, then Island below.
	island := vowB16Top(g, opp, "Island", "Basic Land — Island")
	forest := vowB16Top(g, opp, "Forest", "Basic Land — Forest")
	bolt := vowB16Top(g, opp, "Bolt", "Instant")
	bear := vowB16Top(g, opp, "Bear", vowB16CreatureLine)
	castCatalogSpell(t, g, "Umbris, Fear Manifest", "Legendary Creature — Nightmare Horror", vowB16Umbris, nil)
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, bolt, forest} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s was not exiled", id)
		}
	}
	if !opp.Library.Contains(island) {
		t.Error("the card under the first land was exiled")
	}
}

func TestVowB16UmbrisExilesAWholeLibraryWithNoLand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	vowB16Top(g, opp, "Bolt", "Instant")
	vowB16Top(g, opp, "Bear", vowB16CreatureLine)
	castCatalogSpell(t, g, "Umbris, Fear Manifest", "Legendary Creature — Nightmare Horror", vowB16Umbris, nil)
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != 0 {
		t.Errorf("%d cards left, want the library exiled", opp.Library.Size())
	}
}

func TestVowB16UmbrisTriggersForAnotherNightmareOrHorrorOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Umbris, Fear Manifest", "Legendary Creature — Nightmare Horror", vowB16Umbris, "{3}{U}{B}", 1, 1, "U", "B")
	castCatalogSpell(t, g, "Bear", vowB16CreatureLine, "", nil)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Error("a Bear entering triggered Umbris")
	}
	castCatalogSpell(t, g, "Scary", "Creature — Horror", "", nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) == nil {
		t.Error("another Horror entering did not trigger Umbris")
	}
}
