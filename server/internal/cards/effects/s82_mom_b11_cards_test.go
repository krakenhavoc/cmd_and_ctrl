package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b11_cards_test.go — March of the Machine (and its Commander
// set), S82 batch 11 (#5477): Attentive Skywarden, Seedpod Caretaker,
// Grafted Butcher, Darksteel Splicer, Vulpine Harvester, Phyrexian Censor,
// Phyrexian Pegasus, Ichor Shade, Seer of Stolen Sight, Tenured
// Oilcaster, Terror of Towashi and Filigree Vector.

const (
	s82b11Skywarden = "896213bb-28a0-45a4-abbd-35af51f1d9aa"
	s82b11Caretaker = "257fbe05-b2cb-493c-a116-49f9000b2384"
	s82b11Butcher   = "6c4a0e3d-c63c-4acf-8386-b51f1ec37463"
	s82b11Splicer   = "63a0f0a2-bafc-4cd0-a4ed-42ce58cd3e9b"
	s82b11Harvester = "cc5e3a21-4911-49f8-8e3c-ab23812c4004"
	s82b11Censor    = "4c77de57-29fd-45ba-b2c7-eef231c6b7c7"
	s82b11Pegasus   = "30236ff8-f6d5-462d-9a2f-846b2d8a1ca6"
	s82b11Shade     = "083ea8b0-ff2a-406e-90f1-668f33443747"
	s82b11Seer      = "b2bdccf1-857b-48f8-a9aa-8ef999ef0632"
	s82b11Oilcaster = "7fb837d7-e2c9-40f8-8a7d-156b3aa86830"
	s82b11Terror    = "c4b6dd8f-cdef-413a-a820-ec47e168b3bd"
	s82b11Vector    = "44b90948-e85a-4f69-a9e2-aba63775ba40"
)

func TestS82MomB11CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82b11Skywarden: "Attentive Skywarden", s82b11Caretaker: "Seedpod Caretaker",
		s82b11Butcher: "Grafted Butcher", s82b11Splicer: "Darksteel Splicer",
		s82b11Harvester: "Vulpine Harvester", s82b11Censor: "Phyrexian Censor",
		s82b11Pegasus: "Phyrexian Pegasus", s82b11Shade: "Ichor Shade",
		s82b11Seer: "Seer of Stolen Sight", s82b11Oilcaster: "Tenured Oilcaster",
		s82b11Terror: "Terror of Towashi", s82b11Vector: "Filigree Vector",
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

// s82b11Hand puts a card in seat p's hand and casts it with the active
// seat's main phase open; the spell is left on the stack.
func s82b11Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card) (uuid.UUID, error) {
	t.Helper()
	return msc17Cast(t, g, p, c, game.CastSpellParams{Strict: true})
}

// --- Attentive Skywarden ---------------------------------------------------

func TestS82MomB11SkywardenTransformsAnIncubatorOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sky := b12Push(g, me.ID, "Attentive Skywarden", "Creature — Phyrexian Kor", s82b11Skywarden, 2, 2)
	mine := makeIncubator(t, g, me.ID, 2)
	theirs := makeIncubator(t, g, opp.ID, 2)
	if !effectiveAbilitiesContain(t, g, sky, "flying") {
		t.Fatal("the Skywarden should have flying")
	}
	declareAttack(t, g, opp.ID, sky)
	advanceTo(t, g, game.StepCombatDamage)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("the Skywarden targeted an opponent's Incubator")
	}
	pickCard(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if faceUp(g, mine) != 1 {
		t.Errorf("my Incubator is face %d, want the back face up", faceUp(g, mine))
	}
	if faceUp(g, theirs) != 0 {
		t.Error("an opponent's Incubator turned over")
	}
}

func TestS82MomB11SkywardenMayTargetNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sky := b12Push(g, me.ID, "Attentive Skywarden", "Creature — Phyrexian Kor", s82b11Skywarden, 2, 2)
	mine := makeIncubator(t, g, me.ID, 2)
	declareAttack(t, g, opp.ID, sky)
	advanceTo(t, g, game.StepCombatDamage)
	b04WaitForPick(t, g, me.ID)
	pickTargetRefs(t, g, me.ID, nil)
	passPriorityAroundTable(t, g)
	if faceUp(g, mine) != 0 {
		t.Error("declining the target still turned the Incubator over")
	}
}

// --- Seedpod Caretaker -----------------------------------------------------

func TestS82MomB11CaretakerPutsACounterOnAnArtifactOrCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	s82b11Cast(t, g, me, game.Card{Name: "Seedpod Caretaker", TypeLine: "Creature — Phyrexian Cleric", OracleID: s82b11Caretaker, Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt on entry")
	}
	answerMode(t, g, c, me.ID, 0)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("the counter went to an opponent's creature")
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, rock, game.CounterPlusOne); got != 1 {
		t.Errorf("the artifact has %d +1/+1 counters, want 1", got)
	}
}

func TestS82MomB11CaretakerTransformsAnIncubator(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := makeIncubator(t, g, me.ID, 2)
	theirs := makeIncubator(t, g, opp.ID, 2)
	s82b11Cast(t, g, me, game.Card{Name: "Seedpod Caretaker", TypeLine: "Creature — Phyrexian Cleric", OracleID: s82b11Caretaker, Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt on entry")
	}
	answerMode(t, g, c, me.ID, 1)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("the Caretaker targeted an opponent's Incubator")
	}
	pickCard(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if faceUp(g, mine) != 1 {
		t.Errorf("the Incubator is face %d, want the back face up", faceUp(g, mine))
	}
}

// --- Ichor Shade -----------------------------------------------------------

func s82b11ShadeBoard(t *testing.T) (g *game.Game, me, opp *game.Player, shade uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	shade = b12Push(g, me.ID, "Ichor Shade", "Creature — Phyrexian Shade", s82b11Shade, 2, 3)
	return g, me, opp, shade
}

func TestS82MomB11ShadeGrowsWhenACreatureDiedThisTurn(t *testing.T) {
	g, _, opp, shade := s82b11ShadeBoard(t)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	b27Kill(g, victim)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, shade, game.CounterPlusOne); got != 1 {
		t.Errorf("the Shade has %d +1/+1 counters, want 1", got)
	}
}

func TestS82MomB11ShadeCountsAnArtifactThatWasNotACreature(t *testing.T) {
	g, _, opp, shade := s82b11ShadeBoard(t)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	b27Kill(g, rock)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, shade, game.CounterPlusOne); got != 1 {
		t.Errorf("the Shade has %d +1/+1 counters, want 1", got)
	}
}

func TestS82MomB11ShadeDoesNotGrowOnAQuietTurn(t *testing.T) {
	g, _, opp, shade := s82b11ShadeBoard(t)
	// An enchantment going to the graveyard is neither.
	aura := b12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	b27Kill(g, aura)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, shade, game.CounterPlusOne); got != 0 {
		t.Errorf("the Shade has %d +1/+1 counters, want 0", got)
	}
}

func TestS82MomB11ShadeOnlyTriggersOnItsControllersEndStep(t *testing.T) {
	g, _, opp, shade := s82b11ShadeBoard(t)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	b27Kill(g, victim)
	endTurn(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, shade, game.CounterPlusOne); got != 0 {
		t.Errorf("the Shade grew on another player's end step: %d counters", got)
	}
}

// --- Seer of Stolen Sight --------------------------------------------------

func TestS82MomB11SeerSurveilsOnceForSeveralDyingAtOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Seer of Stolen Sight", "Creature — Phyrexian Warlock", s82b11Seer, 2, 3)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := b12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(a)
		_ = g.DestroyPermanentForEffect(b)
		_ = g.DestroyPermanentForEffect(theirs)
	})
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt after my creature and artifact died")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) != nil {
		t.Error("a second surveil: the trigger should fire once per batch")
	}
}

func TestS82MomB11SeerIgnoresAnOpponentsPermanentsAndNonArtifactNoncreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Seer of Stolen Sight", "Creature — Phyrexian Warlock", s82b11Seer, 2, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	aura := b12Permanent(g, me.ID, "Pacifism", "Enchantment — Aura")
	b27Kill(g, theirs)
	b27Kill(g, aura)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) != nil {
		t.Error("surveilled for an opponent's creature or an enchantment")
	}
}

// --- Tenured Oilcaster -----------------------------------------------------

func TestS82MomB11OilcasterGetsThreeMorePowerAgainstAFullGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oil := b12Push(g, me.ID, "Tenured Oilcaster", "Creature — Phyrexian Wizard", s82b11Oilcaster, 2, 4)
	if !effectiveAbilitiesContain(t, g, oil, "menace") {
		t.Error("the Oilcaster should have menace")
	}
	if p := effectivePower(t, g, oil); p != 2 {
		t.Fatalf("power %d with empty graveyards, want 2", p)
	}
	mill := func(n int) {
		g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, n) })
	}
	mill(7)
	if p := effectivePower(t, g, oil); p != 2 {
		t.Errorf("power %d with seven cards, want 2", p)
	}
	mill(1)
	if p := effectivePower(t, g, oil); p != 5 {
		t.Errorf("power %d with eight cards, want 5", p)
	}
}

func TestS82MomB11OilcasterIgnoresItsOwnersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	oil := b12Push(g, me.ID, "Tenured Oilcaster", "Creature — Phyrexian Wizard", s82b11Oilcaster, 2, 4)
	for i := 0; i < 9; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	if p := effectivePower(t, g, oil); p != 2 {
		t.Errorf("power %d from my own graveyard, want 2", p)
	}
}

func s82b11GraveyardSizes(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Graveyard.Size()
	}
	return out
}

func TestS82MomB11OilcasterMillsEveryPlayerWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oil := b12Push(g, me.ID, "Tenured Oilcaster", "Creature — Phyrexian Wizard", s82b11Oilcaster, 2, 4)
	before := s82b11GraveyardSizes(g)
	declareAttack(t, g, opp.ID, oil)
	passPriorityAroundTable(t, g)
	after := s82b11GraveyardSizes(g)
	for i := range before {
		if after[i] != before[i]+1 {
			t.Errorf("seat %d graveyard %d -> %d, want one more", i, before[i], after[i])
		}
	}
}

func TestS82MomB11OilcasterMillsEveryPlayerWhenItBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oil := b12Push(g, opp.ID, "Tenured Oilcaster", "Creature — Phyrexian Wizard", s82b11Oilcaster, 2, 4)
	atk := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	declareAttack(t, g, opp.ID, atk)
	passPriorityAroundTable(t, g)
	before := s82b11GraveyardSizes(g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(oil, atk); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	after := s82b11GraveyardSizes(g)
	for i := range before {
		if after[i] != before[i]+1 {
			t.Errorf("seat %d graveyard %d -> %d, want one more", i, before[i], after[i])
		}
	}
}

// --- Grafted Butcher -------------------------------------------------------

func TestS82MomB11ButcherPumpsOtherPhyrexiansOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	butcher := b12Push(g, me.ID, "Grafted Butcher", "Creature — Phyrexian Samurai", s82b11Butcher, 2, 2)
	ally := b12Creature(g, me.ID, "Ally", "Creature — Phyrexian Zombie", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Phyrexian Zombie", 2, 2)
	if p := effectivePower(t, g, ally); p != 3 {
		t.Errorf("my Phyrexian is %d power, want 3", p)
	}
	if p := effectivePower(t, g, butcher); p != 2 {
		t.Errorf("the Butcher pumped itself: %d", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a Bear got the bonus: %d", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Phyrexian got the bonus: %d", p)
	}
}

func TestS82MomB11ButcherEntersGivingMenaceToYourPhyrexians(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ally := b12Creature(g, me.ID, "Ally", "Creature — Phyrexian Zombie", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	butcher := avB1ResolveCreature(t, g, me, "Grafted Butcher", "Creature — Phyrexian Samurai", s82b11Butcher)
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, ally, "menace") {
		t.Error("my other Phyrexian should have menace")
	}
	if !effectiveAbilitiesContain(t, g, butcher, "menace") {
		t.Error("the Butcher should have menace")
	}
	if effectiveAbilitiesContain(t, g, bear, "menace") {
		t.Error("a Bear gained menace")
	}
	endTurn(t, g)
	if effectiveAbilitiesContain(t, g, ally, "menace") {
		t.Error("menace outlasted the turn")
	}
}

func TestS82MomB11ButcherReturnsFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	butcher := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: butcher, Name: "Grafted Butcher", TypeLine: "Creature — Phyrexian Samurai",
		OracleID: s82b11Butcher, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	floatForTest(g, me, "CCCB")
	if err := g.ActivateCatalogAbility(me.ID, butcher, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without sacrificing anything")
	}
	p7Activate(t, g, me, butcher, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}})
	if z := b12ZoneOf(g, butcher); z != game.ZoneBattlefield {
		t.Errorf("the Butcher is in %v, want the battlefield", z)
	}
	if z := b12ZoneOf(g, fodder); z != game.ZoneGraveyard {
		t.Errorf("the sacrifice is in %v, want the graveyard", z)
	}
}

func TestS82MomB11ButcherCannotBeActivatedAtInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	butcher := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: butcher, Name: "Grafted Butcher", TypeLine: "Creature — Phyrexian Samurai",
		OracleID: s82b11Butcher, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	advanceTo(t, g, game.StepBeginCombat)
	floatForTest(g, me, "CCCB")
	if err := g.ActivateCatalogAbility(me.ID, butcher, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Error("activated during combat")
	}
}

// --- Darksteel Splicer -----------------------------------------------------

func s82b11Golems(g *game.Game, owner uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Phyrexian Golem" {
			out = append(out, c)
		}
	}
	return out
}

func TestS82MomB11SplicerMakesAGolemPerOpponentAndProtectsThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	splicer := avB1ResolveCreature(t, g, me, "Darksteel Splicer", "Creature — Phyrexian Artificer", s82b11Splicer)
	passPriorityAroundTable(t, g)
	golems := s82b11Golems(g, me.ID)
	if len(golems) != 3 {
		t.Fatalf("%d Golems, want one for each of my 3 opponents", len(golems))
	}
	for _, c := range golems {
		if c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !c.IsArtifact() || !c.IsCreature() || len(c.Colors) != 0 {
			t.Errorf("Golem %+v is not a 3/3 colorless artifact creature", c)
		}
		if !effectiveAbilitiesContain(t, g, c.InstanceID, "indestructible") {
			t.Error("a Golem is not indestructible")
		}
	}
	_ = splicer
}

func TestS82MomB11SplicerTriggersForAnotherNontokenPhyrexianOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Darksteel Splicer", "Creature — Phyrexian Artificer", s82b11Splicer, 1, 1)
	s82b11Cast(t, g, me, game.Card{Name: "Zombie", TypeLine: "Creature — Phyrexian Zombie", Power: 1, Toughness: 1})
	passPriorityAroundTable(t, g)
	if n := len(s82b11Golems(g, me.ID)); n != 3 {
		t.Fatalf("%d Golems after another Phyrexian entered, want 3", n)
	}
	// The Golem tokens are Phyrexians but tokens: they did not retrigger it
	// (three tokens, not twelve), and a non-Phyrexian does not trigger it.
	s82b11Cast(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	passPriorityAroundTable(t, g)
	if n := len(s82b11Golems(g, me.ID)); n != 3 {
		t.Errorf("%d Golems after a Bear entered, want still 3", n)
	}
	// An opponent's Phyrexian is not "you control".
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{Name: "Their Phyrexian", TypeLine: "Token Creature — Phyrexian", Power: 1, Toughness: 1}, 1)
	})
	passPriorityAroundTable(t, g)
	if n := len(s82b11Golems(g, me.ID)); n != 3 {
		t.Errorf("%d Golems after an opponent's Phyrexian entered, want still 3", n)
	}
}

func TestS82MomB11SplicerGolemsAreNotIndestructibleWithoutIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	splicer := b12Push(g, me.ID, "Darksteel Splicer", "Creature — Phyrexian Artificer", s82b11Splicer, 1, 1)
	mine := b12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	theirs := b12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	if !effectiveAbilitiesContain(t, g, mine, "indestructible") {
		t.Error("my Golem should be indestructible")
	}
	if effectiveAbilitiesContain(t, g, theirs, "indestructible") {
		t.Error("an opponent's Golem gained indestructible")
	}
	b27Kill(g, splicer)
	if effectiveAbilitiesContain(t, g, mine, "indestructible") {
		t.Error("the Golem kept indestructible after the Splicer left")
	}
}

// --- Vulpine Harvester -----------------------------------------------------

func s82b11HarvesterBoard(t *testing.T, mvs ...string) (g *game.Game, me, opp *game.Player, harvester uuid.UUID, arts []uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	harvester = b12Push(g, me.ID, "Vulpine Harvester", "Creature — Phyrexian Fox", s82b11Harvester, 3, 3)
	for _, mc := range mvs {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Relic " + mc, TypeLine: "Artifact", ManaCost: mc, Owner: me.ID, Controller: me.ID})
		arts = append(arts, id)
	}
	return g, me, opp, harvester, arts
}

func TestS82MomB11HarvesterReturnsAnArtifactWithinTheirTotalPower(t *testing.T) {
	g, me, opp, harvester, arts := s82b11HarvesterBoard(t, "{3}", "{4}")
	declareAttack(t, g, opp.ID, harvester)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, arts[0])
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, arts[0]); z != game.ZoneBattlefield {
		t.Errorf("the mana value 3 artifact is in %v, want the battlefield", z)
	}
	if z := b12ZoneOf(g, arts[1]); z != game.ZoneGraveyard {
		t.Errorf("the other artifact is in %v, want the graveyard", z)
	}
}

func TestS82MomB11HarvesterLeavesAnArtifactCostingMoreThanTheirPower(t *testing.T) {
	g, me, opp, harvester, arts := s82b11HarvesterBoard(t, "{3}", "{4}")
	declareAttack(t, g, opp.ID, harvester)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, arts[1])
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, arts[1]); z != game.ZoneGraveyard {
		t.Errorf("a mana value 4 artifact came back for 3 power: %v", z)
	}
}

func TestS82MomB11HarvesterAddsUpEveryAttackingPhyrexianButNotOthers(t *testing.T) {
	g, me, opp, harvester, arts := s82b11HarvesterBoard(t, "{5}", "{6}")
	ally := b12Creature(g, me.ID, "Ally", "Creature — Phyrexian Zombie", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 4, 4)
	declareAttack(t, g, opp.ID, harvester, ally, bear)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, arts[1])
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, arts[1]); z != game.ZoneGraveyard {
		t.Errorf("the mana value 6 artifact returned for 5 Phyrexian power (the Bear must not count): %v", z)
	}
}

func TestS82MomB11HarvesterTriggersOnceForSeveralPhyrexians(t *testing.T) {
	g, me, opp, harvester, arts := s82b11HarvesterBoard(t, "{5}")
	ally := b12Creature(g, me.ID, "Ally", "Creature — Phyrexian Zombie", 2, 2)
	declareAttack(t, g, opp.ID, harvester, ally)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, arts[0])
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, arts[0]); z != game.ZoneBattlefield {
		t.Errorf("the mana value 5 artifact is in %v; 3 + 2 power should allow it", z)
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Error("a second trigger asked for a target")
	}
}

// --- Phyrexian Censor ------------------------------------------------------

func TestS82MomB11CensorAllowsOneNonPhyrexianSpellEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, opp.ID, "Phyrexian Censor", "Creature — Phyrexian Wizard", s82b11Censor, 3, 3)
	if _, err := s82b11Cast(t, g, me, game.Card{Name: "Instant One", TypeLine: "Instant"}); err != nil {
		t.Fatalf("the first non-Phyrexian spell was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, err := s82b11Cast(t, g, me, game.Card{Name: "Instant Two", TypeLine: "Instant"}); err == nil {
		t.Error("a second non-Phyrexian spell was allowed")
	}
	if _, err := s82b11Cast(t, g, me, game.Card{Name: "Phyrexian Thing", TypeLine: "Creature — Phyrexian Horror", Power: 1, Toughness: 1}); err != nil {
		t.Errorf("a Phyrexian spell was refused: %v", err)
	}
}

func TestS82MomB11CensorDoesNotCountPhyrexianSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Phyrexian Censor", "Creature — Phyrexian Wizard", s82b11Censor, 3, 3)
	for i := 0; i < 2; i++ {
		if _, err := s82b11Cast(t, g, me, game.Card{Name: "Phyrexian Thing", TypeLine: "Creature — Phyrexian Horror", Power: 1, Toughness: 1}); err != nil {
			t.Fatalf("Phyrexian spell %d refused: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
	if _, err := s82b11Cast(t, g, me, game.Card{Name: "Instant One", TypeLine: "Instant"}); err != nil {
		t.Errorf("a non-Phyrexian spell after two Phyrexian ones was refused: %v", err)
	}
}

func TestS82MomB11CensorMakesNonPhyrexianCreaturesEnterTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, opp.ID, "Phyrexian Censor", "Creature — Phyrexian Wizard", s82b11Censor, 3, 3)
	bear := avB1ResolveCreature(t, g, me, "Bear", "Creature — Bear", "")
	zombie := avB1ResolveCreature(t, g, me, "Zombie", "Creature — Phyrexian Zombie", "")
	if !tappedForTest(t, g, bear) {
		t.Error("a non-Phyrexian creature entered untapped")
	}
	if tappedForTest(t, g, zombie) {
		t.Error("a Phyrexian creature entered tapped")
	}
}

// --- Phyrexian Pegasus -----------------------------------------------------

func TestS82MomB11PegasusGivesAnotherAttackerFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	peg := b12Push(g, me.ID, "Phyrexian Pegasus", "Creature — Phyrexian Pegasus", s82b11Pegasus, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, peg, bear)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: peg}); err == nil {
		t.Fatal("the Pegasus targeted itself")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("the Bear should have flying")
	}
	endTurn(t, g)
	if effectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("flying outlasted the turn")
	}
}

// --- Terror of Towashi -----------------------------------------------------

func s82b11TerrorBoard(t *testing.T) (g *game.Game, me, opp *game.Player, terror, dead uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	terror = b12Push(g, me.ID, "Terror of Towashi", "Creature — Phyrexian Ogre", s82b11Terror, 4, 3)
	dead = uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	return g, me, opp, terror, dead
}

func TestS82MomB11TerrorPaysToReturnACreatureAsAPhyrexian(t *testing.T) {
	g, me, opp, terror, dead := s82b11TerrorBoard(t)
	if !effectiveAbilitiesContain(t, g, terror, "deathtouch") {
		t.Error("Terror of Towashi should have deathtouch")
	}
	declareAttack(t, g, opp.ID, terror)
	passPriorityAroundTable(t, g)
	floatForTest(g, me, "CCCB")
	answerPayUnless(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, dead); z != game.ZoneBattlefield {
		t.Fatalf("the creature card is in %v, want the battlefield", z)
	}
	subs := map[string]bool{}
	for _, st := range effectiveSubtypes(t, g, dead) {
		subs[st] = true
	}
	if !subs["Phyrexian"] || !subs["Bear"] {
		t.Errorf("the returned creature has subtypes %v, want Bear and Phyrexian", subs)
	}
}

func TestS82MomB11TerrorDoesNothingWhenTheCostIsDeclined(t *testing.T) {
	g, me, opp, terror, dead := s82b11TerrorBoard(t)
	declareAttack(t, g, opp.ID, terror)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, dead); z != game.ZoneGraveyard {
		t.Errorf("the creature card is in %v after declining, want the graveyard", z)
	}
}

// --- Filigree Vector -------------------------------------------------------

func TestS82MomB11VectorPutsCountersOnAnyNumberOfTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	s82b11Cast(t, g, me, game.Card{Name: "Filigree Vector", TypeLine: "Artifact Creature — Phyrexian Construct", OracleID: s82b11Vector, Power: 1, Toughness: 1})
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickTargetRefs(t, g, me.ID, nil,
		game.TargetRef{Kind: game.TargetCard, ID: mine}, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if latestPickTarget(g, me.ID) != nil {
		pickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: rock})
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("my creature has %d +1/+1 counters, want 1", got)
	}
	if got := counterCount(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("their creature has %d +1/+1 counters, want 1", got)
	}
	if got := counterCount(g, rock, game.CounterCharge); got != 1 {
		t.Errorf("the artifact has %d charge counters, want 1", got)
	}
	if got := counterCount(g, rock, game.CounterPlusOne); got != 0 {
		t.Errorf("a non-creature artifact got %d +1/+1 counters", got)
	}
}

func TestS82MomB11VectorProliferatesBySacrificingAnotherArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vec := b12Push(g, me.ID, "Filigree Vector", "Artifact Creature — Phyrexian Construct", s82b11Vector, 1, 1)
	grown := b12Creature(g, me.ID, "Grown", "Creature — Bear", 2, 2)
	b28AddCounter(t, g, grown, 1)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	advanceToMain(t, g)
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, vec, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{vec}}); err == nil {
		t.Fatal("the Vector sacrificed itself")
	}
	p7Activate(t, g, me, vec, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}})
	settleAnsweringProliferate(t, g)
	if z := b12ZoneOf(g, rock); z != game.ZoneGraveyard {
		t.Errorf("the sacrificed artifact is in %v", z)
	}
	if got := counterCount(g, grown, game.CounterPlusOne); got != 2 {
		t.Errorf("proliferate: %d counters, want 2", got)
	}
}
