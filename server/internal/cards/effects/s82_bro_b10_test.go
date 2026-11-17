package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b10_test.go — The Brothers' War, batch 10 (#5482): graveyard hate,
// recursion and reanimation. Each test plays the real catalog entry and
// asserts what its text says, and what it refuses.

const (
	s82b10CarrionLocust  = "04e5ecda-9646-4bc3-addf-cca471553fbe"
	s82b10GnawingVermin  = "33ef3d51-845d-460b-97a0-95e7591cf466"
	s82b10HoardRecluse   = "5fc09704-56cf-4e47-bd96-27ac34434941"
	s82b10SwiftgearDrake = "b7618312-6835-4851-9651-b5402001b3b0"
	s82b10Keeper         = "3b3c5fc6-6c4a-4e23-9625-5954d1986160"
	s82b10Disciples      = "cb7e451b-a9d2-4f50-9651-b5df34e34d99"
	s82b10HeroDunes      = "eb6f936e-8dba-49da-84fc-de240b11b8e4"
	s82b10Recommission   = "e43e8c00-f1e7-49d2-bb27-f1c16d85cc22"
	s82b10NoOneLeft      = "afe8fd01-f0c7-46a5-9e9c-e8b9599cd19a"
	s82b10Dreams         = "8c308ebf-df5e-4f47-9b97-5c31944fe91c"
	s82b10Loran          = "9d1f232c-e904-46f6-a4b7-24ca2ede711c"
	s82b10Tocasia        = "e790f7e6-45f5-45df-b8e5-bcbda4f69f8a"
)

func TestS82Bro10EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82b10CarrionLocust:  "Carrion Locust",
		s82b10GnawingVermin:  "Gnawing Vermin",
		s82b10HoardRecluse:   "Hoarding Recluse",
		s82b10SwiftgearDrake: "Swiftgear Drake",
		s82b10Keeper:         "Keeper of the Cadence",
		s82b10Disciples:      "Disciples of Gix",
		s82b10HeroDunes:      "Hero of the Dunes",
		s82b10Recommission:   "Recommission",
		s82b10NoOneLeft:      "No One Left Behind",
		s82b10Dreams:         "Dreams of Steel and Oil",
		s82b10Loran:          "Loran, Disciple of History",
		s82b10Tocasia:        "Tocasia, Dig Site Mentor",
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

// --- helpers ---------------------------------------------------------------

// s82b10Grave puts a card in a player's graveyard.
func s82b10Grave(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

// s82b10Settle answers a trigger's target prompt, when one was raised, and
// lets the table settle.
func s82b10Settle(t *testing.T, g *game.Game, chooser uuid.UUID, pick *game.TargetRef) {
	t.Helper()
	if pick != nil {
		if p := latestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, *pick); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
	}
	passPriorityAroundTable(t, g)
}

// s82b10Enter casts a creature for free and answers its trigger's target.
func s82b10Enter(t *testing.T, g *game.Game, name, typeLine, oracle string, pick *game.TargetRef) uuid.UUID {
	t.Helper()
	id := castAndResolveCreature(t, g, name, typeLine, oracle)
	s82b10Settle(t, g, g.Seats[g.Turn.ActiveSeat].ID, pick)
	return id
}

func s82b10Kill(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id, game.DestroyOptions{}) })
}

func s82b10OnBottom(p *game.Player, id uuid.UUID) bool {
	return len(p.Library.Cards) > 0 && p.Library.Cards[0].InstanceID == id
}

func s82b10Target(id uuid.UUID) *game.TargetRef {
	r := eoeCDCard(id)
	return &r
}

// --- Carrion Locust --------------------------------------------------------

func TestS82Bro10CarrionLocustExilesAndPunishesACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	dead := s82b10Grave(opp, "Dead Bear", "Creature — Bear", "{1}{G}")
	life := opp.Life
	s82b10Enter(t, g, "Carrion Locust", "Creature — Insect Horror", s82b10CarrionLocust, s82b10Target(dead))
	if opp.Graveyard.Contains(dead) || !g.Exile.Contains(dead) {
		t.Fatal("the creature card was not exiled")
	}
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
}

func TestS82Bro10CarrionLocustNonCreatureCostsNoLifeAndOwnGraveyardIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rock := s82b10Grave(opp, "Rock", "Artifact", "{1}")
	mine := s82b10Grave(me, "My Bear", "Creature — Bear", "{1}")
	life := opp.Life
	s82b10Enter(t, g, "Carrion Locust", "Creature — Insect Horror", s82b10CarrionLocust, s82b10Target(rock))
	if !g.Exile.Contains(rock) {
		t.Fatal("the artifact card was not exiled")
	}
	if opp.Life != life {
		t.Errorf("life %d, want %d: it was not a creature card", opp.Life, life)
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard card was exiled")
	}
}

// --- Gnawing Vermin --------------------------------------------------------

func TestS82Bro10GnawingVerminMillsTwoOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	for i := 0; i < 5; i++ {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID})
	}
	lib, yard := len(opp.Library.Cards), len(opp.Graveyard.Cards)
	s82b10Enter(t, g, "Gnawing Vermin", "Creature — Rat", s82b10GnawingVermin, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if len(opp.Library.Cards) != lib-2 || len(opp.Graveyard.Cards) != yard+2 {
		t.Errorf("library %d (was %d), graveyard %d (was %d): want two milled", len(opp.Library.Cards), lib, len(opp.Graveyard.Cards), yard)
	}
}

func TestS82Bro10GnawingVerminDiesShrinksAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vermin := b12Push(g, me.ID, "Gnawing Vermin", "Creature — Rat", s82b10GnawingVermin, 1, 1)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	s82b10Kill(g, vermin)
	s82b10Settle(t, g, me.ID, s82b10Target(theirs))
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("their 1/1 survived -1/-1")
	}
	if _, ok := battlefieldCard(g, mine); !ok {
		t.Error("my own creature was hit")
	}
}

// --- Hoarding Recluse ------------------------------------------------------

func TestS82Bro10HoardingRecluseTucksAnotherCardWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	recluse := b12Push(g, me.ID, "Hoarding Recluse", "Creature — Spider", s82b10HoardRecluse, 2, 3)
	if !hasKeywordOnBattlefield(t, g, recluse, "reach") || !hasKeywordOnBattlefield(t, g, recluse, "deathtouch") {
		t.Error("missing reach or deathtouch")
	}
	other := s82b10Grave(opp, "Their Spell", "Sorcery", "{1}")
	s82b10Kill(g, recluse)
	s82b10Settle(t, g, me.ID, s82b10Target(other))
	if opp.Graveyard.Contains(other) || !s82b10OnBottom(opp, other) {
		t.Error("the card is not on the bottom of its owner's library")
	}
	if !me.Graveyard.Contains(recluse) {
		t.Error("the Recluse itself was tucked: \"other\" was ignored")
	}
}

func TestS82Bro10HoardingRecluseDiesWithNothingElseToTuck(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	recluse := b12Push(g, me.ID, "Hoarding Recluse", "Creature — Spider", s82b10HoardRecluse, 2, 3)
	s82b10Kill(g, recluse)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(recluse) {
		t.Error("the Recluse is not in the graveyard")
	}
}

// --- Swiftgear Drake -------------------------------------------------------

func TestS82Bro10SwiftgearDrakeTucksAGraveyardCard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	dead := s82b10Grave(opp, "Dead Bear", "Creature — Bear", "{1}")
	drake := s82b10Enter(t, g, "Swiftgear Drake", "Artifact Creature — Drake", s82b10SwiftgearDrake, s82b10Target(dead))
	if opp.Graveyard.Contains(dead) || !s82b10OnBottom(opp, dead) {
		t.Error("the card is not on the bottom of its owner's library")
	}
	if !hasKeywordOnBattlefield(t, g, drake, "flying") || !hasKeywordOnBattlefield(t, g, drake, "haste") {
		t.Error("missing flying or haste")
	}
}

// --- Keeper of the Cadence -------------------------------------------------

func TestS82Bro10KeeperTucksAnArtifactInstantOrSorceryButNotACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keeper := b12Push(g, me.ID, "Keeper of the Cadence", "Creature — Human Wizard", s82b10Keeper, 2, 5)
	bolt := s82b10Grave(opp, "Bolt", "Instant", "{R}")
	bear := s82b10Grave(opp, "Bear", "Creature — Bear", "{1}{G}")
	advanceToMain(t, g)
	fillPoolColored(me, "C", 6)
	if err := eoeCDActivate(g, me.ID, keeper, 0, eoeCDCard(bear)); err == nil {
		t.Error("a creature card was a legal target")
	}
	if err := eoeCDActivate(g, me.ID, keeper, 0, eoeCDCard(bolt)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(bolt) || !s82b10OnBottom(opp, bolt) {
		t.Error("the instant is not on the bottom of its owner's library")
	}
}

// --- Disciples of Gix ------------------------------------------------------

func TestS82Bro10DisciplesOfGixPutsUpToThreeArtifactsIntoTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	var arts []uuid.UUID
	for i := 0; i < 4; i++ {
		id := uuid.New()
		arts = append(arts, id)
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	}
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	castAndResolveCreature(t, g, "Disciples of Gix", "Creature — Phyrexian Human", s82b10Disciples)
	passPriorityAroundTable(t, g)
	var choice *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID && c.Kind == game.PendingChoiceSearchLibrary {
			choice = c
		}
	}
	if choice == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(choice.ID, me.ID, arts[:3]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	for _, id := range arts[:3] {
		if !me.Graveyard.Contains(id) {
			t.Errorf("%s is not in the graveyard", id)
		}
	}
	if me.Graveyard.Contains(land) {
		t.Error("a land went to the graveyard")
	}
	if me.Graveyard.Contains(arts[3]) {
		t.Error("a fourth artifact was taken")
	}
}

// --- Hero of the Dunes -----------------------------------------------------

func TestS82Bro10HeroOfTheDunesReanimatesAndPumpsSmallCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	small := s82b10Grave(me, "Small Bear", "Creature — Bear", "{1}{G}")
	big := s82b10Grave(me, "Big Bear", "Creature — Bear", "{3}{G}")
	castAndResolveCreature(t, g, "Hero of the Dunes", "Creature — Human Soldier", s82b10HeroDunes)
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.DiscardOptions, big) {
			t.Error("a mana value 4 card was offered")
		}
	}
	s82b10Settle(t, g, me.ID, s82b10Target(small))
	if _, ok := battlefieldCard(g, small); !ok {
		t.Fatal("the small creature did not return")
	}
	if !me.Graveyard.Contains(big) {
		t.Error("the big creature left the graveyard")
	}
	// The anthem: a 3-mana-value creature gets +1/+0, a 4 does not.
	three := b12Push(g, me.ID, "Three", "Creature — Test", "", 2, 2)
	eoeCDSet(g, three, func(c *game.Card) { c.ManaCost = "{2}{G}" })
	four := b12Push(g, me.ID, "Four", "Creature — Test", "", 2, 2)
	eoeCDSet(g, four, func(c *game.Card) { c.ManaCost = "{3}{G}" })
	if p, _ := eoeCDPT(t, g, three); p != 3 {
		t.Errorf("mana value 3 power %d, want 3", p)
	}
	if p, _ := eoeCDPT(t, g, four); p != 2 {
		t.Errorf("mana value 4 power %d, want 2", p)
	}
}

// --- Recommission ----------------------------------------------------------

func TestS82Bro10RecommissionGivesACreatureAPlusOneCounterButNotAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := s82b10Grave(me, "Bear", "Creature — Bear", "{1}{G}")
	rock := s82b10Grave(me, "Rock", "Artifact", "{2}")
	big := s82b10Grave(me, "Big", "Creature — Giant", "{4}{G}")
	if _, err := castFromHandErr(g, me, "Recommission", "Sorcery", "{1}{W}", s82b10Recommission,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{eoeCDCard(big)}}); err == nil {
		t.Fatal("a mana value 5 card was a legal target")
	}
	for _, tc := range []struct {
		id      uuid.UUID
		counter int
	}{{bear, 1}, {rock, 0}} {
		fillPoolColored(me, "W", 2)
		if _, err := castFromHandErr(g, me, "Recommission", "Sorcery", "{1}{W}", s82b10Recommission,
			game.CastSpellParams{Strict: true, Targets: []game.TargetRef{eoeCDCard(tc.id)}}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		c, ok := battlefieldCard(g, tc.id)
		if !ok {
			t.Fatalf("%s did not return", tc.id)
		}
		if got := c.Counters[game.CounterPlusOne]; got != tc.counter {
			t.Errorf("%s has %d +1/+1 counters, want %d", c.Name, got, tc.counter)
		}
	}
}

// --- No One Left Behind ----------------------------------------------------

func TestS82Bro10NoOneLeftBehindCostsThreeLessForASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	small := s82b10Grave(me, "Small", "Creature — Bear", "{1}{G}")
	big := s82b10Grave(me, "Big", "Creature — Giant", "{4}{G}")
	fillPoolColored(me, "B", 2) // {1}{B}: {4}{B} less {3}
	if _, err := castFromHandErr(g, me, "No One Left Behind", "Sorcery", "{4}{B}", s82b10NoOneLeft,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{eoeCDCard(big)}}); err == nil {
		t.Fatal("a mana value 5 target was cast for {1}{B}")
	}
	if _, err := castFromHandErr(g, me, "No One Left Behind", "Sorcery", "{4}{B}", s82b10NoOneLeft,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{eoeCDCard(small)}}); err != nil {
		t.Fatalf("cast for the reduced cost: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, small); !ok {
		t.Error("the creature did not return")
	}
}

func TestS82Bro10NoOneLeftBehindReturnsABigCreatureAtFullPrice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	big := s82b10Grave(me, "Big", "Creature — Giant", "{4}{G}")
	fillPoolColored(me, "B", 5)
	if _, err := castFromHandErr(g, me, "No One Left Behind", "Sorcery", "{4}{B}", s82b10NoOneLeft,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{eoeCDCard(big)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, big); !ok {
		t.Error("the creature did not return")
	}
}

// --- Loran, Disciple of History --------------------------------------------

func TestS82Bro10LoranReturnsAnArtifactForItselfAndOtherLegends(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rock := s82b10Grave(me, "Rock", "Artifact", "{1}")
	bear := s82b10Grave(me, "Bear", "Creature — Bear", "{1}")
	s82b10Enter(t, g, "Loran, Disciple of History", "Legendary Creature — Human Artificer", s82b10Loran, s82b10Target(rock))
	if !me.Hand.Contains(rock) {
		t.Fatal("Loran's own entry did not return the artifact")
	}
	if me.Hand.Contains(bear) {
		t.Error("a creature card came back")
	}
	rock2 := s82b10Grave(me, "Rock Two", "Artifact", "{1}")
	s82b10Enter(t, g, "Legend", "Legendary Creature — Elf", "", s82b10Target(rock2))
	if !me.Hand.Contains(rock2) {
		t.Error("another legendary creature entering did not return an artifact")
	}
	rock3 := s82b10Grave(me, "Rock Three", "Artifact", "{1}")
	s82b10Enter(t, g, "Plain", "Creature — Elf", "", s82b10Target(rock3))
	if me.Hand.Contains(rock3) {
		t.Error("a nonlegendary creature triggered Loran")
	}
}

// --- Dreams of Steel and Oil -----------------------------------------------

func TestS82Bro10DreamsExilesAHandCardAndAGraveyardCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := revealHand(opp,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Bolt", TypeLine: "Instant"})
	opp.Graveyard.Cards = nil
	gyBear := s82b10Grave(opp, "Dead Bear", "Creature — Bear", "{1}")
	gyBolt := s82b10Grave(opp, "Dead Bolt", "Instant", "{R}")
	gyRock := s82b10Grave(opp, "Dead Rock", "Artifact", "{1}")

	castAt(t, g, "Dreams of Steel and Oil", "Sorcery", s82b10Dreams, pbPlayer(opp.ID))
	pick := openVariant(t, g)
	if !sameIDs(pick.DiscardOptions, ids[1:2]) {
		t.Fatalf("hand options %v, want only the creature", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, ids[1:2]); err != nil {
		t.Fatalf("resolve hand pick: %v", err)
	}
	g.SettleResolution()
	pick = openVariant(t, g)
	if !sameIDs(pick.DiscardOptions, []uuid.UUID{gyBear, gyRock}) || !pick.PickFromGraveyard {
		t.Fatalf("graveyard options %v, want the creature and the artifact", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{gyRock}); err != nil {
		t.Fatalf("resolve graveyard pick: %v", err)
	}
	g.SettleResolution()
	if !g.Exile.Contains(ids[1]) || !g.Exile.Contains(gyRock) {
		t.Error("the chosen cards were not exiled")
	}
	if !opp.Graveyard.Contains(gyBear) || !opp.Graveyard.Contains(gyBolt) || !opp.Hand.Contains(ids[0]) || !opp.Hand.Contains(ids[2]) {
		t.Error("an unchosen card moved")
	}
}

func TestS82Bro10DreamsStillTakesFromTheGraveyardWhenTheHandHasNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	revealHand(opp, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	opp.Graveyard.Cards = nil
	gyBear := s82b10Grave(opp, "Dead Bear", "Creature — Bear", "{1}")
	castAt(t, g, "Dreams of Steel and Oil", "Sorcery", s82b10Dreams, pbPlayer(opp.ID))
	pick := openVariant(t, g)
	if !sameIDs(pick.DiscardOptions, []uuid.UUID{gyBear}) {
		t.Fatalf("options %v", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{gyBear}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	g.SettleResolution()
	if !g.Exile.Contains(gyBear) {
		t.Error("the graveyard card was not exiled")
	}
}

// --- Tocasia, Dig Site Mentor ----------------------------------------------

func TestS82Bro10TocasiaGivesCreaturesVigilanceAndASurveil(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tocasia, Dig Site Mentor", "Legendary Creature — Human Artificer", s82b10Tocasia, 4, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	if !hasKeywordOnBattlefield(t, g, bear, "vigilance") {
		t.Error("my creature has no vigilance")
	}
	if hasKeywordOnBattlefield(t, g, theirs, "vigilance") {
		t.Error("their creature has vigilance")
	}
	advanceToMain(t, g)
	eoeCDSet(g, bear, func(c *game.Card) { c.SummonedThisTurn = false })
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the granted {T}: Surveil 1 was not activatable: %v", err)
	}
	if c, _ := battlefieldCard(g, bear); !c.Tapped {
		t.Error("the {T} cost did not tap the creature")
	}
	passPriorityAroundTable(t, g)
}

func TestS82Bro10TocasiaReturnsArtifactsUpToTotalManaValueTenFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tocasia := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: tocasia, Name: "Tocasia, Dig Site Mentor", TypeLine: "Legendary Creature — Human Artificer",
		OracleID: s82b10Tocasia, ManaCost: "{1}{G}{W}{U}", Owner: me.ID, Controller: me.ID})
	a := s82b10Grave(me, "Six", "Artifact", "{6}")
	b := s82b10Grave(me, "Four", "Artifact", "{4}")
	c := s82b10Grave(me, "Two", "Artifact", "{2}")
	advanceToMain(t, g)
	activate := func(targets ...uuid.UUID) error {
		fillPoolColored(me, "G", 2)
		fillPoolColored(me, "W", 2)
		fillPoolColored(me, "U", 2)
		fillPoolColored(me, "C", 2)
		var refs []game.TargetRef
		for _, id := range targets {
			refs = append(refs, eoeCDCard(id))
		}
		return g.ActivateCatalogAbility(me.ID, tocasia, 0, game.ActivateAbilityParams{Strict: true, Targets: refs})
	}
	if err := activate(a, b, c); err == nil {
		t.Fatal("a total mana value of 12 was accepted")
	}
	if err := activate(a, b); err != nil {
		t.Fatalf("activate with total 10: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if _, ok := battlefieldCard(g, id); !ok {
			t.Errorf("%s did not return", id)
		}
	}
	if !me.Graveyard.Contains(c) {
		t.Error("the untargeted artifact left the graveyard")
	}
	if !g.Exile.Contains(tocasia) {
		t.Error("Tocasia was not exiled as a cost")
	}
}
