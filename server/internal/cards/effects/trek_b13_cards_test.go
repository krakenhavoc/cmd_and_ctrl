package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b13_cards_test.go — Star Trek, batch 13 (#4973). Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	trekB13AvulsionOracle   = "f1e28022-b06b-4efb-9f85-c36d8302ef99"
	trekB13SubmatrixOracle  = "523c244b-7292-40c7-8b12-27106fffe63f"
	trekB13DrednokOracle    = "98cbac1d-7896-4c63-a265-03bc5d670fa6"
	trekB13ResistanceOracle = "49975689-85f8-4221-bb11-3d3f68a1a7ac"
	trekB13RogueAIOracle    = "4849f2f8-0c0e-470d-a4f5-76547a9570a0"
	trekB13SaltVampOracle   = "31c7453e-33b8-4eb6-abac-28cdb35fb993"
	trekB13SelaOracle       = "0cbf0f7a-7f5e-4819-9a9f-f655f92fd5f1"
	trekB13VeteranOracle    = "b1012ac3-77f8-4384-b947-84c094d6b59e"
	trekB13StrayOracle      = "adbbb968-5721-427c-a15a-320d49ab4c3e"
)

func TestTrekB13EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB13AvulsionOracle:   "Organic Avulsion Unit",
		trekB13SubmatrixOracle:  "Overflow Submatrix",
		trekB13DrednokOracle:    "Relentless Drednok",
		trekB13ResistanceOracle: "Resistance Is Futile",
		trekB13RogueAIOracle:    "Rogue Artificial Intelligence",
		trekB13SaltVampOracle:   "Salt Vampire",
		trekB13SelaOracle:       "Sela, Tal Shiar Paradox",
		trekB13VeteranOracle:    "Storied Veteran",
		trekB13StrayOracle:      "Stray from History",
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

// trekB13Card puts a card into a zone of p's: a catalog card when oracle is
// set, with its printed cost and P/T.
func trekB13Card(g *game.Game, z *game.Zone, p *game.Player, name, typeLine, cost, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		z.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
			KnownBy: map[uuid.UUID]bool{p.ID: true},
		})
	})
	return id
}

// trekB13Cast casts a card strictly with the given params.
func trekB13Cast(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana string, params game.CastSpellParams) error {
	t.Helper()
	if mana != "" {
		floatMana(t, g, p, mana)
	}
	params.Strict = true
	return g.CastSpell(p.ID, id, params)
}

func trekB13Counter(g *game.Game, id uuid.UUID, kind string) int { return countersOn(g, id, kind) }

// trekB13Tokens is the tokens named `name` that `controller` controls.
func trekB13Tokens(t *testing.T, g *game.Game, controller uuid.UUID, name string) []game.Card {
	t.Helper()
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			if v := layeredCard(t, g, c.InstanceID); v.IsToken() {
				out = append(out, v)
			}
		}
	}
	return out
}

// --- Organic Avulsion Unit ---------------------------------------------------

func TestTrekB13AvulsionGivesYourArtifactCreaturesMenaceOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	unit := b27Push(g, me.ID, "Organic Avulsion Unit", "Artifact Creature — Borg", trekB13AvulsionOracle, "{2}{B}", 2, 3, "B")
	drone := b12Creature(g, me.ID, "Drone", "Artifact Creature — Borg", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := b12Creature(g, opp.ID, "Their Drone", "Artifact Creature — Borg", 1, 1)

	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, unit, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with one mana")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, unit, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	late := b12Creature(g, me.ID, "Late Drone", "Artifact Creature — Borg", 1, 1)
	for id, want := range map[uuid.UUID]bool{unit: true, drone: true, bear: false, rock: false, theirs: false, late: false} {
		if got := hasEffectiveKeyword(t, g, id, "menace"); got != want {
			t.Errorf("%s menace = %v, want %v", layeredCard(t, g, id).Name, got, want)
		}
	}
	advanceTo(t, g, game.StepCleanup)
	powerUpPassTo(t, g, 1)
	if hasEffectiveKeyword(t, g, drone, "menace") {
		t.Error("menace outlasted the turn")
	}
}

// --- Overflow Submatrix ------------------------------------------------------

func trekB13SubmatrixTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b27Push(g, me.ID, "Overflow Submatrix", "Artifact Creature — Borg", trekB13SubmatrixOracle, "{2}{B}", 3, 3, "B")
	return g, me
}

func TestTrekB13SubmatrixCastsOneArtifactCreatureWithAFinalityCounter(t *testing.T) {
	g, me := trekB13SubmatrixTable(t)
	first := trekB13Card(g, me.Graveyard, me, "Drone A", "Artifact Creature — Borg", "{0}", "", 1, 1)
	second := trekB13Card(g, me.Graveyard, me, "Drone B", "Artifact Creature — Borg", "{0}", "", 1, 1)
	if !grantedPermissionOn(g, me.ID, first, game.ZoneGraveyard).Granted() {
		t.Fatal("an artifact creature card in the graveyard was not opened")
	}
	if err := muldrothaCast(g, me, first, ""); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	resolveAllPT(t, g)
	if !g.Battlefield.Contains(first) {
		t.Fatal("the artifact creature did not resolve onto the battlefield")
	}
	if n := trekB13Counter(g, first, game.CounterFinality); n != 1 {
		t.Errorf("finality counters %d, want 1", n)
	}
	if err := muldrothaCast(g, me, second, ""); err == nil {
		t.Error("a second artifact creature was cast from the graveyard this turn")
	}

	// The budget is per turn: on my next turn it opens again.
	for seat := 1; seat <= 3; seat++ {
		powerUpPassTo(t, g, seat)
	}
	powerUpPassTo(t, g, 0)
	advanceToMain(t, g)
	if err := muldrothaCast(g, me, second, ""); err != nil {
		t.Errorf("my next turn's cast from the graveyard: %v", err)
	}
}

func TestTrekB13SubmatrixOpensOnlyArtifactCreatureSpells(t *testing.T) {
	g, me := trekB13SubmatrixTable(t)
	bear := trekB13Card(g, me.Graveyard, me, "Bear", "Creature — Bear", "{0}", "", 2, 2)
	rock := trekB13Card(g, me.Graveyard, me, "Rock", "Artifact", "{0}", "", 0, 0)
	for _, id := range []uuid.UUID{bear, rock} {
		if grantedPermissionOn(g, me.ID, id, game.ZoneGraveyard).Granted() {
			t.Error("a card that is not an artifact creature was opened")
		}
		if err := muldrothaCast(g, me, id, ""); err == nil {
			t.Error("a card that is not an artifact creature was cast from the graveyard")
		}
	}
}

func TestTrekB13SubmatrixPermissionIsYourTurnOnlyWithAFinalityRider(t *testing.T) {
	spec, _ := Lookup(trekB13SubmatrixOracle)
	if len(spec.CastPermissions) != 1 {
		t.Fatalf("%d cast permissions, want 1", len(spec.CastPermissions))
	}
	p := spec.CastPermissions[0]
	if p.Zone != game.ZoneGraveyard || !p.CastOnly || p.Timing != game.TimingYourTurnOnly {
		t.Errorf("zone %v castOnly %v timing %q, want a cast-only graveyard permission on your turn", p.Zone, p.CastOnly, p.Timing)
	}
	if len(p.PerType) != 1 || p.PerType[0] != game.PermissionTypeCreature {
		t.Errorf("budget %v, want one creature cast", p.PerType)
	}
	if p.EntersWithCounterKind != game.CounterFinality || p.EntersWithCounters != 1 {
		t.Errorf("rider %q x%d, want one finality counter", p.EntersWithCounterKind, p.EntersWithCounters)
	}
}

// --- Relentless Drednok ------------------------------------------------------

func TestTrekB13DrednokExilesItselfFromTheGraveyardToDrawAndLoseOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "Top")
	drednok := trekB13Card(g, me.Graveyard, me, "Relentless Drednok", "Artifact Creature — Robot", "{B}", trekB13DrednokOracle, 1, 1)
	hand, life := me.Hand.Size(), me.Life

	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, drednok, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without black mana")
	}
	if !me.Graveyard.Contains(drednok) {
		t.Fatal("a refused activation moved the card")
	}
	floatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, drednok, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	if got := b12ZoneOf(g, drednok); got != game.ZoneExile {
		t.Fatalf("Drednok is in %q, want exile as a cost", got)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d life %d, want %d and %d", me.Hand.Size(), me.Life, hand+1, life-1)
	}
}

func TestTrekB13DrednokHasLifelinkAndNoBattlefieldActivation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	drednok := b27Push(g, me.ID, "Relentless Drednok", "Artifact Creature — Robot", trekB13DrednokOracle, "{B}", 1, 1, "B")
	assertKeywords(t, g, drednok, "lifelink")
	floatMana(t, g, me, "{B}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, drednok, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the graveyard ability was activated from the battlefield")
	}
}

// --- Resistance Is Futile ----------------------------------------------------

func TestTrekB13ResistanceDestroysANonartifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	drone := b12Creature(g, opp.ID, "Their Drone", "Artifact Creature — Borg", 2, 2)

	spell := trekB13Card(g, me.Hand, me, "Resistance Is Futile", "Sorcery", "{1}{B}", trekB13ResistanceOracle, 0, 0)
	if err := trekB13Cast(t, g, me, spell, "{B}{C}", game.CastSpellParams{
		Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: drone, Mode: 0}},
	}); err == nil {
		t.Fatal("an artifact creature was a legal target for the destroy bullet")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Strict: true, Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear, Mode: 0}},
	}); err != nil {
		t.Fatalf("cast at the bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the nonartifact creature was not destroyed")
	}
	if !g.Battlefield.Contains(drone) {
		t.Error("the artifact creature was touched")
	}
}

func TestTrekB13ResistanceReturnsASmallArtifactCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	small := trekB13Card(g, me.Graveyard, me, "Small Drone", "Artifact Creature — Borg", "{2}{B}", "", 2, 2)
	big := trekB13Card(g, me.Graveyard, me, "Big Drone", "Artifact Creature — Borg", "{3}{B}", "", 4, 4)
	bear := trekB13Card(g, me.Graveyard, me, "Bear", "Creature — Bear", "{1}", "", 2, 2)
	theirs := trekB13Card(g, g.Seats[1].Graveyard, g.Seats[1], "Their Drone", "Artifact Creature — Borg", "{1}", "", 1, 1)

	spell := trekB13Card(g, me.Hand, me, "Resistance Is Futile", "Sorcery", "{1}{B}", trekB13ResistanceOracle, 0, 0)
	floatMana(t, g, me, "{B}{C}")
	for _, bad := range []uuid.UUID{big, bear, theirs} {
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
			Strict: true, Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bad, Mode: 0}},
		}); err == nil {
			t.Fatal("an illegal card was a legal target for the return bullet")
		}
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Strict: true, Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small, Mode: 0}},
	}); err != nil {
		t.Fatalf("cast at the small drone: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the artifact creature card was not returned to the battlefield")
	}
}

// --- Rogue Artificial Intelligence -------------------------------------------

func TestTrekB13RogueAIGainsThreeOrSurveilsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	castAndResolveCreature(t, g, "Rogue Artificial Intelligence", "Artifact Creature — Construct", trekB13RogueAIOracle)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatal("no mode pick from the enters trigger")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}

	g = newCatalogGame(t)
	me = g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	castAndResolveCreature(t, g, "Rogue Artificial Intelligence", "Artifact Creature — Construct", trekB13RogueAIOracle)
	c = latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatal("no mode pick from the enters trigger")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	s := surveilChoiceFor(g, me.ID)
	if s == nil || len(s.ScryCards) != 2 {
		t.Fatalf("surveil prompt %+v, want two cards", s)
	}
	if err := g.ResolveSurveil(s.ID, me.ID, s.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if me.Graveyard.Size() != 2 {
		t.Errorf("graveyard %d, want the two surveilled cards", me.Graveyard.Size())
	}
}

func TestTrekB13RogueAIBasicLandcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	basic := seedSearchLibrary(me, searchTestLand("Swamp", "Basic Land — Swamp"))[0]
	id, _ := cycleFromHand(t, g, "Rogue Artificial Intelligence", "Artifact Creature — Construct", trekB13RogueAIOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the discard cost did not put the card in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("basic landcycling did not search")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{basic}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(basic) {
		t.Error("the basic land did not reach the hand")
	}
}

// --- Salt Vampire ------------------------------------------------------------

func trekB13SaltVampireTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(opp, "Their Top")
	spell := trekB13Card(g, opp.Hand, opp, "Their Bolt", "Instant", "{R}", "", 0, 0)
	land := trekB13Card(g, opp.Hand, opp, "Their Mountain", "Basic Land — Mountain", "", "", 0, 0)
	vamp := castAndResolveCreature(t, g, "Salt Vampire", "Creature — Vampire Shapeshifter", trekB13SaltVampOracle)
	assertKeywords(t, g, vamp, "lifelink")
	if pick := latestPickTarget(g, me.ID); pick != nil {
		for _, p := range pick.PickTargetPlayers {
			if p == me.ID {
				t.Error("I was offered as the target opponent")
			}
		}
	}
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	return g, me, opp, spell, land
}

func TestTrekB13SaltVampireExilesAChosenNonlandCardAndTheyDraw(t *testing.T) {
	g, me, opp, spell, land := trekB13SaltVampireTable(t)
	pick := openVariant(t, g)
	if pick.Chooser != me.ID || !hasID(pick.DiscardOptions, spell) || hasID(pick.DiscardOptions, land) {
		t.Fatalf("pick chooser %v options %v, want me choosing among nonland cards only", pick.Chooser, pick.DiscardOptions)
	}
	if !pick.PickOptional {
		t.Error("the choice is not optional")
	}
	hand := opp.Hand.Size()
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, spell); got != game.ZoneExile {
		t.Errorf("the chosen card is in %q, want exile", got)
	}
	if rvDiscards(g, spell) != 0 {
		t.Error("the exile was a discard")
	}
	if opp.Hand.Size() != hand {
		t.Errorf("their hand %d, want %d (one exiled, one drawn)", opp.Hand.Size(), hand)
	}
	if !opp.Hand.Contains(land) {
		t.Error("the land left their hand")
	}
}

func TestTrekB13SaltVampireChoosingNothingDrawsNothing(t *testing.T) {
	g, me, opp, spell, _ := trekB13SaltVampireTable(t)
	pick := openVariant(t, g)
	hand := opp.Hand.Size()
	if err := g.ResolvePendingChoice(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(spell) || opp.Hand.Size() != hand {
		t.Errorf("their hand changed (%d, want %d)", opp.Hand.Size(), hand)
	}
}

// --- Sela, Tal Shiar Paradox -------------------------------------------------

func TestTrekB13SelaExilesACardFromAnyGraveyardWhenSheAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sela := b27Push(g, me.ID, "Sela, Tal Shiar Paradox", "Legendary Creature — Romulan Human Rogue", trekB13SelaOracle, "{3}{B}", 3, 3, "B")
	theirs := trekB13Card(g, opp.Graveyard, opp, "Their Dead", "Creature — Bear", "{1}", "", 2, 2)
	declareAttack(t, g, opp.ID, sela)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, theirs); got != game.ZoneExile {
		t.Errorf("their card is in %q, want exile", got)
	}
	if n := len(trekB13Tokens(t, g, me.ID, "Romulan")); n != 0 {
		t.Errorf("%d Romulans from a card leaving another player's graveyard", n)
	}
}

func TestTrekB13SelaMakesOneTappedRomulanPerBatchLeavingYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b27Push(g, me.ID, "Sela, Tal Shiar Paradox", "Legendary Creature — Romulan Human Rogue", trekB13SelaOracle, "{3}{B}", 3, 3, "B")
	a := trekB13Card(g, me.Graveyard, me, "Dead A", "Creature — Bear", "{1}", "", 2, 2)
	b := trekB13Card(g, me.Graveyard, me, "Dead B", "Creature — Bear", "{1}", "", 2, 2)

	// Two cards leave together: one trigger.
	stray := trekB13Card(g, me.Hand, me, "Stray from History", "Sorcery", "{2}{B}", trekB13StrayOracle, 0, 0)
	if err := trekB13Cast(t, g, me, stray, "{B}{C}{C}", game.CastSpellParams{
		Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a, Mode: 0}, {Kind: game.TargetCard, ID: b, Mode: 0}},
	}); err != nil {
		t.Fatalf("cast Stray from History: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	tokens := trekB13Tokens(t, g, me.ID, "Romulan")
	if len(tokens) != 1 {
		t.Fatalf("%d Romulans, want one for one batch", len(tokens))
	}
	r := tokens[0]
	if !r.Tapped || r.Power != 2 || r.Toughness != 2 || !r.HasColor("B") || !r.HasSubtype("Romulan") {
		t.Errorf("Romulan %+v, want a tapped 2/2 black Romulan", r)
	}
}

// --- Storied Veteran ---------------------------------------------------------

func TestTrekB13VeteranCastFromTheGraveyardEntersWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	vet := trekB13Card(g, me.Graveyard, me, "Storied Veteran", "Creature — Klingon Advisor", "{B}", trekB13VeteranOracle, 2, 1)
	if err := trekB13Cast(t, g, me, vet, "{C}", game.CastSpellParams{FromZone: "graveyard"}); err == nil {
		t.Fatal("cast from the graveyard without black mana")
	}
	floatMana(t, g, me, "{B}")
	if err := g.CastSpell(me.ID, vet, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := trekB13Counter(g, vet, game.CounterFinality); n != 1 {
		t.Fatalf("finality counters %d, want 1", n)
	}

	// The finality counter exiles it instead of letting it die.
	b27Kill(g, vet)
	g.RunStateChecksForTest()
	if got := b12ZoneOf(g, vet); got != game.ZoneExile {
		t.Errorf("the destroyed Veteran is in %q, want exile", got)
	}
}

func TestTrekB13VeteranCastFromHandHasNoFinalityCounterAndCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	vet := trekB13Card(g, me.Hand, me, "Storied Veteran", "Creature — Klingon Advisor", "{B}", trekB13VeteranOracle, 2, 1)
	if err := trekB13Cast(t, g, me, vet, "{B}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := trekB13Counter(g, vet, game.CounterFinality); n != 0 {
		t.Errorf("finality counters %d after a hand cast, want 0", n)
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	attacker := b12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := b27Push(g2, opp2.ID, "Storied Veteran", "Creature — Klingon Advisor", trekB13VeteranOracle, "{B}", 2, 1, "B")
	advanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(attacker, opp2.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g2, game.StepDeclareBlockers)
	if err := g2.DeclareBlocker(blocker, attacker); err == nil {
		t.Error("Storied Veteran blocked")
	}
}

// --- Stray from History ------------------------------------------------------

func TestTrekB13StrayReturnsCreatureAndSpacecraftCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := trekB13Card(g, me.Graveyard, me, "Bear", "Creature — Bear", "{1}", "", 2, 2)
	ship := trekB13Card(g, me.Graveyard, me, "Ship", "Artifact — Spacecraft", "{3}", "", 0, 0)
	rock := trekB13Card(g, me.Graveyard, me, "Rock", "Artifact", "{1}", "", 0, 0)
	theirs := trekB13Card(g, g.Seats[1].Graveyard, g.Seats[1], "Their Bear", "Creature — Bear", "{1}", "", 2, 2)

	stray := trekB13Card(g, me.Hand, me, "Stray from History", "Sorcery", "{2}{B}", trekB13StrayOracle, 0, 0)
	floatMana(t, g, me, "{B}{C}{C}")
	for _, bad := range []uuid.UUID{rock, theirs} {
		if err := g.CastSpell(me.ID, stray, game.CastSpellParams{
			Strict: true, Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bad, Mode: 0}},
		}); err == nil {
			t.Fatal("an illegal card was a legal target for Rewrite the Past")
		}
	}
	if err := g.CastSpell(me.ID, stray, game.CastSpellParams{
		Strict: true, Modes: []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear, Mode: 0}, {Kind: game.TargetCard, ID: ship, Mode: 0}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || !me.Hand.Contains(ship) {
		t.Error("the creature and Spacecraft cards did not return to hand")
	}
	if !me.Graveyard.Contains(rock) {
		t.Error("the plain artifact left the graveyard")
	}
}

func TestTrekB13StrayExilesANonlandCardFromTheRevealedHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	spell := trekB13Card(g, opp.Hand, opp, "Their Bolt", "Instant", "{R}", "", 0, 0)
	land := trekB13Card(g, opp.Hand, opp, "Their Mountain", "Basic Land — Mountain", "", "", 0, 0)

	stray := trekB13Card(g, me.Hand, me, "Stray from History", "Sorcery", "{2}{B}", trekB13StrayOracle, 0, 0)
	if err := trekB13Cast(t, g, me, stray, "{B}{C}{C}", game.CastSpellParams{
		Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID, Mode: 0}},
	}); err == nil {
		t.Fatal("I was a legal target opponent")
	}
	if err := g.CastSpell(me.ID, stray, game.CastSpellParams{
		Strict: true, Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := openVariant(t, g)
	if pick.Chooser != me.ID || !hasID(pick.DiscardOptions, spell) || hasID(pick.DiscardOptions, land) || pick.PickDestination != game.PickExile {
		t.Fatalf("pick chooser %v options %v destination %v", pick.Chooser, pick.DiscardOptions, pick.PickDestination)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if got := b12ZoneOf(g, spell); got != game.ZoneExile {
		t.Errorf("the chosen card is in %q, want exile", got)
	}
}
