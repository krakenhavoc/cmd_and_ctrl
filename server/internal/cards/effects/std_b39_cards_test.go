package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b39_cards_test.go — Standard set batch 39 (#4452): Molten
// Duplication, Bristlebud Farmer, Tarnation Vista, Sanctuary Wall,
// Seasoned Consultant, Tenth District Hero, Wojek Investigator, Wrench,
// Agency Outfitter and Cryptic Coat. Each test plays the real catalog
// entry and asserts what its text says.

const (
	b39MoltenOracle      = "92b91910-9b65-4421-8eef-a33793161089"
	b39FarmerOracle      = "3e44226b-51a2-4099-aff8-a5befab720cb"
	b39VistaOracle       = "b0be3f25-edb1-4299-959c-9aad909730ca"
	b39WallOracle        = "83856c3a-652e-4afc-9dc0-0072324ac36f"
	b39ConsultantOracle  = "9ee64b00-6619-451e-bb48-10af107dcb2f"
	b39HeroOracle        = "0953772c-9374-4616-9c04-5838aaf626c9"
	b39InvestigatorOracl = "ab41243d-b178-4ebe-a7ac-bea37427be99"
	b39WrenchOracle      = "fdbac0a7-18c0-48a5-a174-0d4d1473fd9c"
	b39OutfitterOracle   = "23b2de61-a29c-4a32-b38c-408c723727ce"
	b39CoatOracle        = "73b8adea-1786-44fc-b190-bf240b07f5b6"
	b39GlassOracle       = "f2ce9e55-07c0-4f36-a59e-127693fc0f62"
	b39CapOracle         = "a54fb3de-0581-461e-be33-dc4f4d16a33e"
)

func TestB39EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b39MoltenOracle:      "Molten Duplication",
		b39FarmerOracle:      "Bristlebud Farmer",
		b39VistaOracle:       "Tarnation Vista",
		b39WallOracle:        "Sanctuary Wall",
		b39ConsultantOracle:  "Seasoned Consultant",
		b39HeroOracle:        "Tenth District Hero",
		b39InvestigatorOracl: "Wojek Investigator",
		b39WrenchOracle:      "Wrench",
		b39OutfitterOracle:   "Agency Outfitter",
		b39CoatOracle:        "Cryptic Coat",
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

// b39TokensNamed lists the battlefield tokens with a name.
func b39TokensNamed(g *game.Game, name string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.IsToken() {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

// --- Molten Duplication ------------------------------------------------------

func TestB39MoltenDuplicationCopiesAsAnArtifactWithHasteThenSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Test Bear", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Molten Duplication", "Sorcery", b39MoltenOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	toks := b39TokensNamed(g, "Test Bear")
	if len(toks) != 1 {
		t.Fatalf("%d Test Bear tokens, want 1", len(toks))
	}
	c := layeredCard(t, g, toks[0])
	if !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Bear") {
		t.Errorf("the copy is not an artifact creature Bear: %v %v", c.Effective().Types, c.Effective().Subtypes)
	}
	assertKeywords(t, g, toks[0], "haste")
	if layeredCard(t, g, bear).IsArtifact() {
		t.Error("the original became an artifact")
	}

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(toks[0]) {
		t.Error("the copy was not sacrificed at the beginning of the end step")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the original was sacrificed")
	}
}

func TestB39MoltenDuplicationRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Molten Duplication", "Sorcery", b39MoltenOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Error("an opponent's creature was a legal target")
	}
}

// --- Bristlebud Farmer -------------------------------------------------------

func TestB39BristlebudFarmerEntersWithTwoFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Bristlebud Farmer", "Creature — Plant Druid", b39FarmerOracle, 5, 5)
	if got := tmtB1Count(g, me.ID, "Food"); got != 2 {
		t.Errorf("%d Food tokens, want 2", got)
	}
}

func TestB39BristlebudFarmerAttackSacrificesAFoodMillsAndTakesAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	farmer := b12Push(g, me.ID, "Bristlebud Farmer", "Creature — Plant Druid", b39FarmerOracle, 5, 5)
	food := pushToken(g, me.ID, FoodToken())
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		for _, c := range []game.Card{
			{Name: "Top Spell", TypeLine: "Instant"},
			{Name: "Middle Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
			{Name: "Bottom Land", TypeLine: "Basic Land — Forest"},
		} {
			c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
			me.Library.PushTop(c)
		}
	})
	declareAttack(t, g, opp.ID, farmer)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID, food)
	if g.Battlefield.Contains(food) {
		t.Fatal("the Food was not sacrificed")
	}
	if got := me.Library.Size(); got != 0 {
		t.Fatalf("%d cards left in the library, want 0 (three milled)", got)
	}
	pick := pendingOfKind(g, game.PendingChoiceChooseCards)
	if pick == nil {
		t.Fatal("no choice of a permanent card from among the milled cards")
	}
	if len(pick.ChooseCards) != 2 {
		t.Errorf("%d candidates, want the 2 permanent cards (the instant is not one)", len(pick.ChooseCards))
	}
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards[:1]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

func TestB39BristlebudFarmerAttackWithoutAFoodDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	farmer := b12Push(g, me.ID, "Bristlebud Farmer", "Creature — Plant Druid", b39FarmerOracle, 5, 5)
	auraSeedLibrary(me, 5)
	declareAttack(t, g, opp.ID, farmer)
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != 0 {
		t.Errorf("milled %d cards with no Food to sacrifice", got)
	}
}

// --- Tarnation Vista ---------------------------------------------------------

func TestB39TarnationVistaEntersTappedAndTapsForTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vista := b12PlayFromHand(t, g, "Tarnation Vista", "Land", b39VistaOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "R")
	if !b13Tapped(t, g, vista) {
		t.Error("entered untapped")
	}
	untap(g, vista)
	tapForMana(t, g, me.ID, vista)
	if got := sortedPool(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool = %v, want [R]", got)
	}
}

func TestB39TarnationVistaAddsOneManaPerColorAmongMonocoloredPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vista := pushChosenColorPermanent(t, g, me.ID, "Tarnation Vista", "Land", b39VistaOracle, "R")
	b16Creature(g, me.ID, "Red One", "Creature — Goblin", 1, 1, "R")
	b16Creature(g, me.ID, "Red Two", "Creature — Goblin", 1, 1, "R")
	b16Creature(g, me.ID, "Green One", "Creature — Elf", 1, 1, "G")
	b16Creature(g, me.ID, "Gold One", "Creature — Elf", 1, 1, "B", "U")  // multicolored: ignored
	b16Creature(g, me.ID, "Grey One", "Artifact Creature — Golem", 1, 1) // colorless: ignored
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b16Creature(g, opp.ID, "White Foe", "Creature — Human", 1, 1, "W") // not mine

	floatMana(t, g, me, "{C}")
	if err := g.ActivateManaAbility(me.ID, vista, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	got := sortedPool(me)
	if len(got) != 2 || got[0] != "G" || got[1] != "R" {
		t.Errorf("pool = %v, want one {R} and one {G}", got)
	}
}

// --- Sanctuary Wall ----------------------------------------------------------

func b39WallSetup(t *testing.T) (g *game.Game, me *game.Player, wall, target uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	wall = b12Push(g, me.ID, "Sanctuary Wall", "Artifact Creature — Wall", b39WallOracle, 0, 4)
	target = b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	floatMana(t, g, me, "{C}{C}{W}")
	p7Activate(t, g, me, wall, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
	passPriorityAroundTable(t, g)
	return
}

func TestB39SanctuaryWallTapsAndMayStunBoth(t *testing.T) {
	g, me, wall, target := b39WallSetup(t)
	answerConfirmFor(t, g, me, true)
	if !b16Tapped(t, g, target) {
		t.Error("the target is untapped")
	}
	if got := countersOn(g, target, game.CounterStun); got != 1 {
		t.Errorf("%d stun counters on the target, want 1", got)
	}
	if got := countersOn(g, wall, game.CounterStun); got != 1 {
		t.Errorf("%d stun counters on the Wall, want 1", got)
	}
}

func TestB39SanctuaryWallDecliningStillTapsButStunsNothing(t *testing.T) {
	g, me, wall, target := b39WallSetup(t)
	answerConfirmFor(t, g, me, false)
	if !b16Tapped(t, g, target) {
		t.Error("the target is untapped")
	}
	if countersOn(g, target, game.CounterStun) != 0 || countersOn(g, wall, game.CounterStun) != 0 {
		t.Error("a stun counter was placed after declining")
	}
}

// --- Seasoned Consultant -----------------------------------------------------

func TestB39SeasonedConsultantPumpsWhenYouAttackWithThree(t *testing.T) {
	for _, tc := range []struct {
		attackers int
		wantPower int
	}{{3, 3}, {2, 1}} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		advanceToMain(t, g)
		consultant := b12Push(g, me.ID, "Seasoned Consultant", "Creature — Human Detective", b39ConsultantOracle, 1, 3)
		var bears []uuid.UUID
		for i := 0; i < tc.attackers; i++ {
			bears = append(bears, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
		}
		declareAttack(t, g, opp.ID, bears...)
		passPriorityAroundTable(t, g)
		if p, _ := mkmB1PT(t, g, consultant); p != tc.wantPower {
			t.Errorf("%d attackers: Consultant power %d, want %d", tc.attackers, p, tc.wantPower)
		}
	}
}

// --- Tenth District Hero -----------------------------------------------------

func b39HeroSetup(t *testing.T, graveyardValue string) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hero := b12Push(g, me.ID, "Tenth District Hero", "Creature — Human", b39HeroOracle, 2, 3)
	ev := mkmAC2PushGraveyard(g, me, "Evidence", "Creature — Bear", "", graveyardValue)
	return g, me, hero, ev
}

func TestB39TenthDistrictHeroBecomesAHumanDetective(t *testing.T) {
	g, me, hero, ev := b39HeroSetup(t, "{1}{G}")
	floatMana(t, g, me, "{C}{W}")
	p7Activate(t, g, me, hero, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{ev}})
	passPriorityAroundTable(t, g)
	if p, tough := mkmB1PT(t, g, hero); p != 4 || tough != 4 {
		t.Errorf("Hero is %d/%d, want 4/4", p, tough)
	}
	assertKeywords(t, g, hero, "vigilance")
	c := layeredCard(t, g, hero)
	if !c.HasSubtype("Human") || !c.HasSubtype("Detective") {
		t.Errorf("subtypes %v, want Human Detective", c.Effective().Subtypes)
	}
	if me.Graveyard.Contains(ev) {
		t.Error("the evidence card was not exiled")
	}
}

func TestB39TenthDistrictHeroCollectEvidenceNeedsEnoughManaValue(t *testing.T) {
	g, me, hero, ev := b39HeroSetup(t, "{G}")
	floatMana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, hero, 0, game.ActivateAbilityParams{
		Strict: true, ExileIDs: []uuid.UUID{ev},
	}); err == nil {
		t.Error("collected evidence 2 with mana value 1")
	}
}

func TestB39TenthDistrictHeroBecomesMilevaOnlyIfADetective(t *testing.T) {
	g, me, hero, ev := b39HeroSetup(t, "{2}{G}{G}")
	floatMana(t, g, me, "{C}{C}{W}")
	// Not a Detective yet: the ability resolves and does nothing.
	p7Activate(t, g, me, hero, 1, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{ev}})
	passPriorityAroundTable(t, g)
	if p, tough := mkmB1PT(t, g, hero); p != 2 || tough != 3 {
		t.Fatalf("a Hero that is no Detective became %d/%d", p, tough)
	}
	if layeredCard(t, g, hero).Effective().Name == "Mileva, the Stalwart" {
		t.Fatal("renamed without being a Detective")
	}

	ev2 := mkmAC2PushGraveyard(g, me, "Evidence", "Creature — Bear", "", "{2}{G}")
	ev3 := mkmAC2PushGraveyard(g, me, "Evidence", "Creature — Bear", "", "{1}{G}")
	floatMana(t, g, me, "{C}{W}")
	p7Activate(t, g, me, hero, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{ev3}})
	passPriorityAroundTable(t, g)
	ev4 := mkmAC2PushGraveyard(g, me, "Evidence", "Creature — Bear", "", "{2}{G}")
	floatMana(t, g, me, "{C}{C}{W}")
	p7Activate(t, g, me, hero, 1, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{ev2, ev4}})
	passPriorityAroundTable(t, g)
	_ = ev3
	c := layeredCard(t, g, hero)
	if c.Effective().Name != "Mileva, the Stalwart" {
		t.Errorf("name %q, want Mileva, the Stalwart", c.Effective().Name)
	}
	if !c.HasSupertype("Legendary") || !c.HasSubtype("Detective") {
		t.Errorf("supertypes %v subtypes %v, want a legendary Detective", c.Effective().Supertypes, c.Effective().Subtypes)
	}
	if p, tough := mkmB1PT(t, g, hero); p != 5 || tough != 5 {
		t.Errorf("Mileva is %d/%d, want 5/5", p, tough)
	}
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	assertKeywords(t, g, other, "indestructible")
	assertKeywords(t, g, hero, "vigilance")
}

// --- Wojek Investigator ------------------------------------------------------

func TestB39WojekInvestigatorInvestigatesPerOpponentWithMoreCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seat := g.Turn.ActiveSeat
	b12Push(g, me.ID, "Wojek Investigator", "Creature — Angel Detective", b39InvestigatorOracl, 2, 4)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	stepToSeat(t, g, seat, game.StepUpkeep)
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			p.Hand.Cards = nil
		}
		fill := func(p *game.Player, n int) {
			for i := 0; i < n; i++ {
				p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
			}
		}
		fill(me, 2)
		for i, p := range g.Seats {
			if p == me {
				continue
			}
			switch i % 3 {
			case 0:
				fill(p, 3) // more
			case 1:
				fill(p, 2) // equal: not more
			default:
				fill(p, 5) // more
			}
		}
	})
	want := 0
	for i, p := range g.Seats {
		if p != me && (i%3 == 0 || i%3 == 2) {
			want++
		}
	}
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Clue"); got != want {
		t.Errorf("%d Clues, want %d (one per opponent with more cards in hand)", got, want)
	}
}

// --- Wrench ------------------------------------------------------------------

func TestB39WrenchPumpsGivesVigilanceAndTapAbilityThenSacrificesToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	wrench := b12Push(g, me.ID, "Wrench", "Artifact — Clue Equipment", b39WrenchOracle, 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	foe := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)

	floatMana(t, g, me, "{C}{C}")
	p7Activate(t, g, me, wrench, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if p, tough := mkmB1PT(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped creature is %d/%d, want 3/3", p, tough)
	}
	assertKeywords(t, g, bear, "vigilance")

	idx, ref := grantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the equipped creature has no granted ability")
	}
	floatMana(t, g, me, "{C}{C}{C}")
	b16Activate(t, g, me.ID, bear, idx, game.ActivateAbilityParams{
		Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: foe}},
	})
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, foe) {
		t.Error("the target creature was not tapped")
	}

	auraSeedLibrary(me, 2)
	hand := me.Hand.Size()
	floatMana(t, g, me, "{C}{C}")
	p7Activate(t, g, me, wrench, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(wrench) {
		t.Error("Wrench was not sacrificed")
	}
	if me.Hand.Size()-hand != 1 {
		t.Errorf("drew %d cards, want 1", me.Hand.Size()-hand)
	}
}

// --- Agency Outfitter --------------------------------------------------------

func b39OutfitterSetup(t *testing.T) (g *game.Game, me *game.Player, glassGrave, capHand, capLib, glassLib uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	card := func(name, oracle string) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Artifact", OracleID: oracle, Owner: me.ID, Controller: me.ID}
	}
	g1, c1, c2, g2 := card("Magnifying Glass", b39GlassOracle), card("Thinking Cap", b39CapOracle),
		card("Thinking Cap", b39CapOracle), card("Magnifying Glass", b39GlassOracle)
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(g1)
		me.Hand.PushTop(c1)
		me.Library.PushTop(c2)
		me.Library.PushTop(g2)
	})
	return g, me, g1.InstanceID, c1.InstanceID, c2.InstanceID, g2.InstanceID
}

func TestB39AgencyOutfitterFindsOneOfEachNameAcrossZonesAndShuffles(t *testing.T) {
	g, me, glassGrave, capHand, capLib, _ := b39OutfitterSetup(t)
	castHoldCreature(t, g, "Agency Outfitter", "Creature — Sphinx Detective", b39OutfitterOracle, 4, 3)
	answerConfirmFor(t, g, me, true)
	pick := pendingOfKind(g, game.PendingChoiceChooseCards)
	if pick == nil {
		t.Fatal("no search prompt")
	}
	if len(pick.ChooseCards) != 4 {
		t.Errorf("%d candidates, want 4 (graveyard, hand and library copies)", len(pick.ChooseCards))
	}
	// Two cards of one name is not a legal set.
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{capHand, capLib}); err == nil {
		t.Error("two cards named Thinking Cap were accepted")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{glassGrave, capLib}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(glassGrave) || !g.Battlefield.Contains(capLib) {
		t.Error("the chosen Magnifying Glass and Thinking Cap are not on the battlefield")
	}
	if !me.Hand.Contains(capHand) {
		t.Error("the unchosen Thinking Cap left the hand")
	}
}

func TestB39AgencyOutfitterWithoutSearchingTheLibraryOnlySeesGraveyardAndHand(t *testing.T) {
	g, me, glassGrave, capHand, _, _ := b39OutfitterSetup(t)
	castHoldCreature(t, g, "Agency Outfitter", "Creature — Sphinx Detective", b39OutfitterOracle, 4, 3)
	answerConfirmFor(t, g, me, false)
	pick := pendingOfKind(g, game.PendingChoiceChooseCards)
	if pick == nil {
		t.Fatal("no search prompt")
	}
	if len(pick.ChooseCards) != 2 {
		t.Errorf("%d candidates, want 2 (graveyard and hand only)", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{glassGrave, capHand}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(glassGrave) || !g.Battlefield.Contains(capHand) {
		t.Error("the graveyard Glass and hand Cap are not on the battlefield")
	}
}

// --- Cryptic Coat ------------------------------------------------------------

func TestB39CrypticCoatCloaksTheTopCardAndAttachesToIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: top, Name: "Secret Bear", TypeLine: "Creature — Bear",
			ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	coat := castCatalogSpell(t, g, "Cryptic Coat", "Artifact — Equipment", b39CoatOracle, nil)
	passPriorityAroundTable(t, g)

	c := findCardAnywhere(t, g, top)
	if !g.Battlefield.Contains(top) || !c.FaceDown || c.FaceDownKind != game.FaceDownCloaked {
		t.Fatalf("the top card is not a cloaked permanent: %+v", c.FaceDownKind)
	}
	if !attachedTo(t, g, coat, top) {
		t.Fatal("the Coat is not attached to the cloaked creature")
	}
	// 2/2 + 1/+0, can't be blocked.
	if p, tough := mkmB1PT(t, g, top); p != 3 || tough != 2 {
		t.Errorf("cloaked creature is %d/%d, want 3/2", p, tough)
	}
	_ = opp

	floatMana(t, g, me, "{C}{U}")
	p7Activate(t, g, me, coat, 0, game.ActivateAbilityParams{})
	if !me.Hand.Contains(coat) {
		t.Error("{1}{U} did not return the Coat to hand")
	}
}

func TestB39CrypticCoatWithAnEmptyLibraryCloaksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Library.Cards = nil })
	coat := castCatalogSpell(t, g, "Cryptic Coat", "Artifact — Equipment", b39CoatOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(coat) {
		t.Error("the Coat did not enter")
	}
}
