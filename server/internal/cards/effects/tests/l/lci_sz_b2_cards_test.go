package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_sz_b2_cards_test.go — The Lost Caverns of Ixalan, S–Z batch 2
// (#4470). Each test plays the real catalog entry and asserts what its
// text says.

const (
	lciSz2SaheeliOracle    = "5aa76c5d-8046-42ce-b61d-cc9b4806c11a"
	lciSz2ReflectionOracle = "5c68853e-4fb8-465e-b262-e6dbaf972533"
	lciSz2SentryOracle     = "f05575f8-4961-4780-83fc-3584d9146a14"
	lciSz2BenthisaurOracle = "d3cad867-142c-4e9d-bcd8-0e7ed1fd2f77"
	lciSz2ViperOracle      = "24c3c505-ddaa-4adb-983e-7c78d9933cc7"
	lciSz2EmergenceOracle  = "e468f1fe-d99c-4034-a94f-91c41a129092"
	lciSz2CrewmateOracle   = "8ab49974-3f06-4b6b-a416-d6f872b3887f"
	lciSz2UchbenbakOracle  = "36bfcc4a-7f66-4649-ab6b-18d360299d46"
	lciSz2VitoOracle       = "9768d43b-8b3f-4d03-ac1b-e603af832875"
)

func TestLciSz2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lciSz2SaheeliOracle:    "Saheeli, the Sun's Brilliance",
		lciSz2ReflectionOracle: "Self-Reflection",
		lciSz2SentryOracle:     "Shipwreck Sentry",
		lciSz2BenthisaurOracle: "Sinuous Benthisaur",
		lciSz2ViperOracle:      "Soulcoil Viper",
		lciSz2EmergenceOracle:  "Squirming Emergence",
		lciSz2CrewmateOracle:   "Staunch Crewmate",
		lciSz2UchbenbakOracle:  "Uchbenbak, the Great Mistake",
		lciSz2VitoOracle:       "Vito, Fanatic of Aclazotz",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

// lciSz2GraveyardPermanents puts n vanilla permanent cards into p's
// graveyard (the descend count).
func lciSz2GraveyardPermanents(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.PushCatalogGraveyardCard(p, "Old Rock", "Artifact", "", 0, 0)
	}
}

// --- Saheeli, the Sun's Brilliance -------------------------------------------------

func TestLciSz2SaheeliCopiesAsAnArtifactWithHasteAndSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	saheeli := effectstest.B12Push(g, me.ID, "Saheeli, the Sun's Brilliance", "Legendary Creature — Human Artificer", lciSz2SaheeliOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	if err := effectstest.HobB4Activate(t, g, me.ID, saheeli, 0, "{U}{R}", effectstest.CardRefs(saheeli)); err == nil {
		t.Fatal("Saheeli copied herself; the target must be another creature or artifact")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, saheeli, 0, "", effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, id := range effectstest.TmtB1Named(g, "Grizzly Bears") {
		if id != bear {
			token = id
		}
	}
	if token == uuid.Nil {
		t.Fatal("no token copy of the Bears")
	}
	c := effectstest.TmtB1Card(t, g, token)
	if !c.IsToken() || !c.IsArtifact() || !c.IsCreature() {
		t.Errorf("the copy is %q; want an artifact creature token", c.TypeLine)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, token, "haste") {
		t.Error("the copy has no haste")
	}
	if orig := effectstest.TmtB1Card(t, g, bear); orig.IsArtifact() || effectstest.EffectiveAbilitiesContain(t, g, bear, "haste") {
		t.Error("the original became an artifact or gained haste")
	}
	if !effectstest.TappedNow(g, saheeli) {
		t.Error("Saheeli did not tap")
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the token was not sacrificed at the beginning of the end step")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the original Bears were sacrificed")
	}
}

// --- Self-Reflection ---------------------------------------------------------------

func TestLciSz2SelfReflectionCopiesAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Self-Reflection", "Sorcery", lciSz2ReflectionOracle, effectstest.CardRefs(theirs)); err == nil {
		t.Fatal("Self-Reflection targeted a creature you don't control")
	}
	id := effectstest.SeedGraveyardCard(t, g, "Self-Reflection", "Sorcery", lciSz2ReflectionOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Targets: effectstest.CardRefs(bear),
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	copies := 0
	for _, tid := range effectstest.TmtB1Named(g, "Grizzly Bears") {
		if tid != bear && effects.IsToken(effectstest.TmtB1Card(t, g, tid)) {
			copies++
		}
	}
	if copies != 1 {
		t.Errorf("token copies: %d, want 1", copies)
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back Self-Reflection is not in exile")
	}
}

// --- Shipwreck Sentry --------------------------------------------------------------

func TestLciSz2ShipwreckSentryAttacksOnlyAfterAnArtifactEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentry := effectstest.B12Push(g, me.ID, "Shipwreck Sentry", "Creature — Human Pirate", lciSz2SentryOracle, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sentry, opp.ID); err == nil {
		t.Fatal("a defender attacked with no artifact having entered this turn")
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	if err := g.DeclareAttacker(sentry, opp.ID); err != nil {
		t.Fatalf("Shipwreck Sentry could not attack after an artifact entered: %v", err)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, sentry, "defender") {
		t.Error("the waiver removed defender")
	}
}

func TestLciSz2ShipwreckSentryIgnoresAnOpponentsArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentry := effectstest.B12Push(g, me.ID, "Shipwreck Sentry", "Creature — Human Pirate", lciSz2SentryOracle, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TreasureToken(), 1) })
	if err := g.DeclareAttacker(sentry, opp.ID); err == nil {
		t.Error("an artifact that entered under an opponent's control let the Sentry attack")
	}
}

// --- Sinuous Benthisaur ------------------------------------------------------------

func TestLciSz2SinuousBenthisaurLooksAtXAndTakesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Permanent(g, me.ID, "Echoing Deeps", "Land — Cave")
	effectstest.B12Permanent(g, me.ID, "Hidden Cataract", "Land — Cave")
	effectstest.B12Permanent(g, g.Seats[1].ID, "Their Cave", "Land — Cave")
	effectstest.PushCatalogGraveyardCard(me, "Old Cave", "Land — Cave", "", 0, 0)
	top := effectstest.HornTopFive(me, effectstest.HornFive)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Sinuous Benthisaur", TypeLine: "Creature — Dinosaur", OracleID: lciSz2BenthisaurOracle, Power: 4, Toughness: 4})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("the enters trigger raised no prompt")
	}
	// Two Caves you control plus one Cave card in the graveyard: X = 3.
	if len(prompt.ChooseCards) != 3 || prompt.ChooseMin != 2 || prompt.ChooseMax != 2 {
		t.Fatalf("offered %d cards [%d,%d], want the top three, exactly two", len(prompt.ChooseCards), prompt.ChooseMin, prompt.ChooseMax)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[0], top[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[0]) || !me.Hand.Contains(top[2]) {
		t.Error("the two chosen cards are not in hand")
	}
	cards := me.Library.Cards
	if cards[0].InstanceID != top[1] {
		t.Errorf("the bottom card is %q, want the card not taken", cards[0].Name)
	}
}

func TestLciSz2SinuousBenthisaurLooksAtNothingWithoutCaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Sinuous Benthisaur", TypeLine: "Creature — Dinosaur", OracleID: lciSz2BenthisaurOracle, Power: 4, Toughness: 4})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("a prompt was raised with X = 0")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand)
	}
}

// --- Soulcoil Viper ----------------------------------------------------------------

func TestLciSz2SoulcoilViperReturnsACreatureWithFinality(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	viper := effectstest.B12Push(g, me.ID, "Soulcoil Viper", "Creature — Snake", lciSz2ViperOracle, 2, 3)
	dead := effectstest.PushCatalogGraveyardCard(me, "Fallen Hero", "Creature — Human Hero", "", 3, 3)
	g.AddManaForEffect(me.ID, uuid.Nil, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, viper, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(dead),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(viper) {
		t.Error("the Viper was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	bf := effectstest.FindBattlefieldByName(g, "Fallen Hero")
	if bf == uuid.Nil {
		t.Fatal("the creature card did not return")
	}
	if got := effectstest.CountersOn(g, bf, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
}

func TestLciSz2SoulcoilViperIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	viper := effectstest.B12Push(g, me.ID, "Soulcoil Viper", "Creature — Snake", lciSz2ViperOracle, 2, 3)
	dead := effectstest.PushCatalogGraveyardCard(me, "Fallen Hero", "Creature — Human Hero", "", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	g.AddManaForEffect(me.ID, uuid.Nil, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, viper, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead)}); err == nil {
		t.Error("the ability was activated at the beginning of combat")
	}
}

// --- Squirming Emergence -----------------------------------------------------------

func TestLciSz2SquirmingEmergenceCapsTheManaValueByTheDescendCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := uuid.New()
	pricey := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Cheap Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: pricey, Name: "Pricey Dragon", TypeLine: "Creature — Dragon", ManaCost: "{5}{R}{R}", Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spent Spell", TypeLine: "Sorcery", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	})
	// Two permanent cards in the graveyard: mana value 2 or less.
	if err := effectstest.CastCatalogSpellErr(t, g, "Squirming Emergence", "Sorcery", lciSz2EmergenceOracle, effectstest.CardRefs(cheap)); err == nil {
		t.Fatal("a mana value 3 card was a legal target with two permanent cards in the graveyard")
	}
	lciSz2GraveyardPermanents(me, 1)
	effectstest.CastCatalogSpell(t, g, "Squirming Emergence", "Sorcery", lciSz2EmergenceOracle, effectstest.CardRefs(cheap))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the mana value 3 card did not return with three permanent cards in the graveyard")
	}
	if g.Battlefield.Contains(pricey) {
		t.Error("an untargeted card returned")
	}
}

// --- Staunch Crewmate --------------------------------------------------------------

func TestLciSz2StaunchCrewmateOffersArtifactsAndPirates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{
		{"Mountain", "Basic Land — Mountain"},
		{"Rusty Relic", "Artifact"},
		{"Deckhand", "Creature — Human Pirate"},
		{"Bear", "Creature — Bear"},
		{"Under Relic", "Artifact"},
	})
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Staunch Crewmate", TypeLine: "Creature — Human Pirate", OracleID: lciSz2CrewmateOracle, Power: 2, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("the enters trigger raised no prompt")
	}
	if len(prompt.ChooseCards) != 2 || prompt.ChooseMax != 1 || prompt.ChooseMin != 0 {
		t.Fatalf("offered %v [%d,%d], want the artifact and the Pirate from the top four, optional", prompt.ChooseCards, prompt.ChooseMin, prompt.ChooseMax)
	}
	for _, id := range prompt.ChooseCards {
		if id == top[4] {
			t.Error("a card below the top four is offered")
		}
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[2]) {
		t.Error("the Pirate card is not in hand")
	}
}

// --- Uchbenbak, the Great Mistake --------------------------------------------------

func TestLciSz2UchbenbakReturnsWithDescendEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogGraveyardCard(me, "Uchbenbak, the Great Mistake", "Legendary Creature — Skeleton Horror", lciSz2UchbenbakOracle, 6, 4)
	lciSz2GraveyardPermanents(me, 6)
	g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{U}{B}")
	// Uchbenbak and six more: seven permanent cards.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the ability was activated with seven permanent cards in the graveyard")
	}
	lciSz2GraveyardPermanents(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with eight permanent cards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Uchbenbak did not return")
	}
	if got := effectstest.CountersOn(g, id, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
	for _, kw := range []string{"vigilance", "menace"} {
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Uchbenbak has no %s", kw)
		}
	}
}

// --- Vito, Fanatic of Aclazotz -----------------------------------------------------

func TestLciSz2VitoEscalatesWithEachSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vito := effectstest.B12Push(g, me.ID, "Vito, Fanatic of Aclazotz", "Legendary Creature — Vampire Demon", lciSz2VitoOracle, 4, 4)
	if !effectstest.HasEffectiveKeyword(t, g, vito, "flying") {
		t.Error("Vito has no flying")
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}

	effectstest.AvB9Sacrifice(t, g, effectstest.B12Permanent(g, me.ID, "Junk", "Artifact"))
	if me.Life != lives[me.ID]+2 || opp.Life != lives[opp.ID] {
		t.Errorf("first: life %d (want %d), opponent %d (want %d)", me.Life, lives[me.ID]+2, opp.Life, lives[opp.ID])
	}
	effectstest.AvB9Sacrifice(t, g, effectstest.B12Permanent(g, me.ID, "Old Land", "Land"))
	for _, p := range g.Seats[1:] {
		if p.Life != lives[p.ID]-2 {
			t.Errorf("second: %s at %d, want %d", p.Name, p.Life, lives[p.ID]-2)
		}
	}
	if me.Life != lives[me.ID]+2 {
		t.Errorf("second: controller life %d, want unchanged at %d", me.Life, lives[me.ID]+2)
	}
	effectstest.AvB9Sacrifice(t, g, effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1))
	if got := effectstest.TokensNamedControlledBy(g, "Vampire Demon", me.ID); got != 1 {
		t.Fatalf("third: %d Vampire Demons, want 1", got)
	}
	effectstest.AvB9Sacrifice(t, g, effectstest.B12Permanent(g, me.ID, "More Junk", "Artifact"))
	if got := effectstest.TokensNamedControlledBy(g, "Vampire Demon", me.ID); got != 1 || me.Life != lives[me.ID]+2 {
		t.Error("a fourth resolution did something")
	}
	// An opponent's sacrifice is not yours.
	effectstest.AvB9Sacrifice(t, g, effectstest.B12Permanent(g, opp.ID, "Their Junk", "Artifact"))
	if me.Life != lives[me.ID]+2 {
		t.Error("an opponent's sacrifice triggered Vito")
	}
}
