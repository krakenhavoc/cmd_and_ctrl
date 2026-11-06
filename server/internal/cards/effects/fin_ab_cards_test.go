package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_ab_cards_test.go — Final Fantasy (FIN), the cards A–B (#4462).
// Each test plays the real catalog entry and asserts what its text says.

const (
	finABSteinerOracle    = "657a4500-4b47-4c78-945b-a7791af7d10e"
	finABAirshipOracle    = "e3016295-00f7-4daf-91b0-57653733b103"
	finABAerithOracle     = "dd9a6f6b-f9b7-4292-bc80-b7da326de032"
	finABRescueOracle     = "14c3e48b-4c4e-4787-8c50-0e3aee19aa23"
	finABAhrimanOracle    = "37d1d507-25dc-4cd0-b51e-2f1e9b24832c"
	finABCrashOracle      = "72a0e0f1-e8be-441a-bcaf-029fb1bd1fd2"
	finABSalvagersOracle  = "43f81514-e812-4e17-9b1a-704c136d8ffd"
	finABRealmOracle      = "fd230ab5-874e-4827-9b5c-6d39d079cee6"
	finABAsheOracle       = "db75577d-1001-4088-999f-40684218a07f"
	finABAuronOracle      = "53a4d9ee-6dfa-4a8b-a1e4-7f671a666188"
	finABTRexaurOracle    = "3f678f76-3863-4ab5-a4fd-9b76a282b1a8"
	finABBalthierOracle   = "570cb560-c246-4652-b230-ab34f966c78a"
	finABBardsBowOracle   = "4aceb786-960d-4719-9de8-79dffd491a51"
	finABBartzOracle      = "2d995c47-e548-4fc4-9182-fef46af20ca5"
	finABBattleMenuOracle = "8af57b80-bbc6-47ce-83f6-34ae8e2f5bd5"
	finABBeatrixOracle    = "2f24637d-8d8e-49de-b229-4cdb97595ed7"
	finABBlackRodOracle   = "909b3d90-5762-461b-b49a-004a9e72733d"
	finABBombOracle       = "088acd15-029b-4091-960a-d5060a364ca9"
	finABBlitzballOracle  = "c99779b7-a271-416e-bdc2-0f9ba6ceb144"
	finABQuinjetOracle    = "60105eb8-bdca-44b9-a970-68a7615526e8"
)

func TestFinABEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finABSteinerOracle:        "Adelbert Steiner",
		finABAirshipOracle:        "Adventurer's Airship",
		finABAerithOracle:         "Aerith Gainsborough",
		finABRescueOracle:         "Aerith Rescue Mission",
		finABAhrimanOracle:        "Ahriman",
		finABCrashOracle:          "Airship Crash",
		finABSalvagersOracle:      "Al Bhed Salvagers",
		finABRealmOracle:          "A Realm Reborn",
		finABAsheOracle:           "Ashe, Princess of Dalmasca",
		finABAuronOracle:          "Auron's Inspiration",
		balambGardenOracle:        "Balamb Garden, SeeD Academy",
		balambGardenOracle + "#1": "Balamb Garden, Airborne",
		finABTRexaurOracle:        "Balamb T-Rexaur",
		finABBalthierOracle:       "Balthier and Fran",
		finABBardsBowOracle:       "Bard's Bow",
		finABBartzOracle:          "Bartz and Boko",
		finABBattleMenuOracle:     "Battle Menu",
		finABBeatrixOracle:        "Beatrix, Loyal General",
		finABBlackRodOracle:       "Black Mage's Rod",
		finABBombOracle:           "Blazing Bomb",
		finABBlitzballOracle:      "Blitzball Shot",
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

// --- fixtures ------------------------------------------------------------------

// finABOpp is the seat after the active one.
func finABOpp(g *game.Game) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// finABEquipment drops a plain Equipment under `owner`.
func finABEquipment(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Artifact — Equipment", Owner: owner, Controller: owner,
	})
}

// finABAttack declares each attacker at `defender` and locks the
// declaration in.
func finABAttack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
}

// finABCastInstantNow casts a catalog instant from the active seat's hand
// in the current step, unpaid.
func finABCastInstantNow(t *testing.T, g *game.Game, name, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushCatalogHandCard(me, name, "Instant", oracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func finABCard(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// --- Adelbert Steiner ------------------------------------------------------------

func TestFinABSteinerGrowsWithEachEquipmentYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	steiner := b12Push(g, me.ID, "Adelbert Steiner", "Legendary Creature — Human Knight", finABSteinerOracle, 2, 1)
	if v := s58p6View(t, g, steiner); v.CurrentPower() != 2 || v.CurrentToughness() != 1 {
		t.Fatalf("Steiner is %d/%d with no Equipment, want 2/1", v.CurrentPower(), v.CurrentToughness())
	}
	finABEquipment(g, me.ID, "Sword")
	finABEquipment(g, me.ID, "Shield")
	finABEquipment(g, opp.ID, "Their Sword")
	if v := s58p6View(t, g, steiner); v.CurrentPower() != 4 || v.CurrentToughness() != 3 {
		t.Errorf("Steiner is %d/%d with two of your Equipment, want 4/3", v.CurrentPower(), v.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, g, steiner, "lifelink") {
		t.Error("Steiner has no lifelink")
	}
}

// --- Adventurer's Airship ---------------------------------------------------------

func TestFinABAirshipCrewsAndLootsWhenItAttacks(t *testing.T) {
	s := newCrewScene(t)
	ship := s.pushVehicle("Adventurer's Airship", finABAirshipOracle, 3, 2)
	pilot := s.pushCreature("Pilot", "Creature — Human", 2, 2)
	s.crewAndResolve(t, ship, 0, pilot)
	if !hasEffectiveKeyword(t, s.g, ship, "flying") {
		t.Error("the Airship has no flying")
	}
	hand := s.me.Hand.Size()
	finABAttack(t, s.g, s.opp.ID, ship)
	passPriorityAroundTable(t, s.g)
	if s.me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d, want the draw first", hand, s.me.Hand.Size())
	}
	discard := s.me.Hand.Cards[0].InstanceID
	answerDiscard(t, s.g, s.me.ID, discard)
	if s.me.Hand.Size() != hand || !s.me.Graveyard.Contains(discard) {
		t.Errorf("hand %d, want %d after the discard", s.me.Hand.Size(), hand)
	}
}

// --- Aerith Gainsborough ----------------------------------------------------------

func TestFinABAerithGrowsOnLifeGainAndPassesHerCountersOnWhenSheDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	aerith := b12Push(g, me.ID, "Aerith Gainsborough", "Legendary Creature — Human Cleric", finABAerithOracle, 2, 2)
	legend := b12Creature(g, me.ID, "Cloud", "Legendary Creature — Human Soldier", 2, 2)
	plain := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Sephiroth", "Legendary Creature — Human Avatar", 2, 2)
	for i := 0; i < 2; i++ {
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
		passPriorityAroundTable(t, g)
	}
	if got := countersOn(g, aerith, game.CounterPlusOne); got != 2 {
		t.Fatalf("Aerith has %d +1/+1 counters after two life gains, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aerith) })
	passPriorityAroundTable(t, g)
	if got := countersOn(g, legend, game.CounterPlusOne); got != 2 {
		t.Errorf("your legendary creature has %d counters, want 2", got)
	}
	if got := countersOn(g, plain, game.CounterPlusOne); got != 0 {
		t.Errorf("a nonlegendary creature has %d counters, want 0", got)
	}
	if got := countersOn(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's legendary creature has %d counters, want 0", got)
	}
}

// --- Aerith Rescue Mission --------------------------------------------------------

func TestFinABRescueMissionMakesThreeHeroes(t *testing.T) {
	g := newCatalogGame(t)
	castModal(t, g, "Aerith Rescue Mission", "Sorcery", finABRescueOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Hero"); n != 3 {
		t.Errorf("%d Hero tokens, want 3", n)
	}
}

func TestFinABRescueMissionTapsThreeAndStunsTheOneChosen(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	castModal(t, g, "Aerith Rescue Mission", "Sorcery", finABRescueOracle, []int{1}, []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Mode: 0},
		{Kind: game.TargetCard, ID: b, Mode: 0},
		{Kind: game.TargetCard, ID: c, Mode: 0},
	})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if !isTapped(g, id) {
			t.Fatal("a target was not tapped")
		}
	}
	me := g.Seats[g.Turn.ActiveSeat]
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("stun prompt = %+v, want one of the three", pick)
	}
	answerChooseCards(t, g, me.ID, b)
	if countersOn(g, b, game.CounterStun) != 1 || countersOn(g, a, game.CounterStun) != 0 || countersOn(g, c, game.CounterStun) != 0 {
		t.Error("the stun counter is not on the chosen creature alone")
	}
}

// --- Ahriman ---------------------------------------------------------------------

func TestFinABAhrimanSacrificesAnotherArtifactToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ahriman := pushCatalogPermanent(g, me.ID, "Ahriman", "Creature — Eye Horror", finABAhrimanOracle, false)
	rock := pushPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	if !hasEffectiveKeyword(t, g, ahriman, "deathtouch") || !hasEffectiveKeyword(t, g, ahriman, "flying") {
		t.Error("Ahriman lacks flying or deathtouch")
	}
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, ahriman, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{ahriman}}); err == nil {
		t.Fatal("Ahriman sacrificed itself to its own ability")
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, ahriman, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(rock) {
		t.Fatal("the artifact was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

// --- Airship Crash ---------------------------------------------------------------

func TestFinABAirshipCrashDestroysAFlierButNotAGroundCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ground := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	if err := castCatalogSpellErr(t, g, "Airship Crash", "Instant", finABCrashOracle, finABCard(ground)); err == nil {
		t.Fatal("Airship Crash targeted a creature without flying")
	}
	castCatalogSpell(t, g, "Airship Crash", "Instant", finABCrashOracle, finABCard(flier))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flier) {
		t.Error("the flier survived")
	}
}

func TestFinABAirshipCrashDestroysAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	aura := pushPermanent(g, opp.ID, game.Card{Name: "Glorious Anthem", TypeLine: "Enchantment"})
	castCatalogSpell(t, g, "Airship Crash", "Instant", finABCrashOracle, finABCard(aura))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment survived")
	}
}

func TestFinABAirshipCrashCycles(t *testing.T) {
	g := newCatalogGame(t)
	hand := g.Seats[g.Turn.ActiveSeat].Hand.Size()
	id, me := cycleFromHand(t, g, "Airship Crash", "Instant", finABCrashOracle, "{C}{C}")
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) || me.Hand.Size() != hand+1 {
		t.Errorf("cycling: in graveyard %v, hand %d → %d", me.Graveyard.Contains(id), hand, me.Hand.Size())
	}
}

// --- Al Bhed Salvagers ------------------------------------------------------------

func TestFinABSalvagersDrainWhenAnArtifactOrItselfDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	salvagers := b12Push(g, me.ID, "Al Bhed Salvagers", "Creature — Human Artificer Warrior", finABSalvagersOracle, 2, 3)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, dying := range []uuid.UUID{theirs, rock, salvagers} {
		myLife, oppLife := me.Life, opp.Life
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dying) })
		passPriorityAroundTable(t, g)
		if dying == theirs {
			if latestPickTarget(g, me.ID) != nil {
				t.Fatal("an opponent's creature dying triggered the Salvagers")
			}
			continue
		}
		pickPlayer(t, g, me.ID, opp.ID)
		passPriorityAroundTable(t, g)
		if me.Life != myLife+1 || opp.Life != oppLife-1 {
			t.Errorf("after a death: me %d → %d, opponent %d → %d", myLife, me.Life, oppLife, opp.Life)
		}
	}
}

// --- A Realm Reborn ---------------------------------------------------------------

func TestFinABRealmRebornGivesYourOtherPermanentsAnyColorMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := gaPut(g, me.ID, "Rock", "Artifact", "")
	forest := gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	theirs := gaPut(g, opp.ID, "Their Rock", "Artifact", "")
	realm := gaPut(g, me.ID, "A Realm Reborn", "Enchantment", finABRealmOracle)
	if n := grantedManaRowsOf(g, rock); n != 1 {
		t.Fatalf("your artifact has %d granted mana rows, want 1", n)
	}
	if n := grantedManaRowsOf(g, forest); n != 1 {
		t.Errorf("your land has %d granted mana rows, want 1", n)
	}
	if n := grantedManaRowsOf(g, realm); n != 0 {
		t.Errorf("A Realm Reborn grants itself %d rows; it says other permanents", n)
	}
	if n := grantedManaRowsOf(g, theirs); n != 0 {
		t.Errorf("an opponent's artifact has %d granted rows, want 0", n)
	}
	if err := tapGrantedMana(t, g, me.ID, rock, "B"); err != nil {
		t.Fatalf("tap the artifact: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want [B]", got)
	}
}

// --- Ashe, Princess of Dalmasca ---------------------------------------------------

func TestFinABAsheOffersOnlyArtifactCardsFromTheTopFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	top := hornTopFive(me, hornFive)
	ashe := pushCatalogPermanent(g, me.ID, "Ashe, Princess of Dalmasca", "Legendary Creature — Human Rebel Noble", finABAsheOracle, false)
	finABAttack(t, g, opp.ID, ashe)
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("Ashe's attack raised no prompt")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != top[4] || prompt.ChooseMax != 1 || prompt.ChooseMin != 0 {
		t.Fatalf("offered %v [%d,%d], want only the artifact card, optional", prompt.ChooseCards, prompt.ChooseMin, prompt.ChooseMax)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[4]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[4]) {
		t.Error("the artifact card is not in hand")
	}
}

// --- Auron's Inspiration ----------------------------------------------------------

func TestFinABAuronsInspirationPumpsAttackersAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	home := b12Creature(g, me.ID, "Home", "Creature — Bear", 2, 2)
	finABAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	spell := finABCastInstantNow(t, g, "Auron's Inspiration", finABAuronOracle, nil)
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, attacker); v.CurrentPower() != 4 || v.CurrentToughness() != 2 {
		t.Errorf("the attacker is %d/%d, want 4/2", v.CurrentPower(), v.CurrentToughness())
	}
	if v := s58p6View(t, g, home); v.CurrentPower() != 2 {
		t.Errorf("a creature that is not attacking has power %d, want 2", v.CurrentPower())
	}
	if !me.Graveyard.Contains(spell) {
		t.Fatal("the spell is not in the graveyard")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, attacker); v.CurrentPower() != 6 {
		t.Errorf("after flashback the attacker has power %d, want 6", v.CurrentPower())
	}
	if !inExile(g, spell) {
		t.Error("the flashed-back spell was not exiled")
	}
}

// --- Balamb Garden ------------------------------------------------------------------

// finABGarden puts the front face of Balamb Garden onto the battlefield.
func finABGarden(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: balambGardenOracle, Layout: game.LayoutTransform,
		Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Balamb Garden, SeeD Academy", TypeLine: "Land — Town"},
			{Name: "Balamb Garden, Airborne", TypeLine: "Legendary Artifact — Vehicle", Power: 5, Toughness: 4},
		},
	}
	c.SetFace(0)
	return pushBattlefieldCardWithTimestamp(g, c)
}

func TestFinABBalambGardenEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	land := playLandFromHand(t, g, "Balamb Garden, SeeD Academy", balambGardenOracle)
	if c := findBattlefieldCardByID(g, land); c == nil || !c.Tapped {
		t.Error("Balamb Garden should enter tapped")
	}
}

func TestFinABBalambGardenCostsOneLessPerOtherTown(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	garden := finABGarden(g, me.ID)
	pushPermanent(g, me.ID, game.Card{Name: "Starting Town", TypeLine: "Land — Town"})
	// One other Town: {4}{G}{U}. Five mana is one short, which it would
	// not be if the Garden counted itself.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{U}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, garden, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("with one other Town and five mana: err = %v, want insufficient mana", err)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Crossroads Village", TypeLine: "Land — Town"})
	if err := g.ActivateCatalogAbility(me.ID, garden, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("with two other Towns and {3}{G}{U}: %v", err)
	}
	passPriorityAroundTable(t, g)
	got := rvBattlefieldCard(g, garden)
	if got == nil || got.Name != "Balamb Garden, Airborne" {
		t.Fatalf("the Garden did not transform: %+v", got)
	}
	if !got.Tapped || got.IsCreature() {
		t.Errorf("Balamb Garden, Airborne: tapped %v, creature %v; want tapped, not a creature until crewed", got.Tapped, got.IsCreature())
	}
}

func TestFinABBalambGardenAirborneCrewsAndDrawsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	advanceToMain(t, g)
	garden := finABGarden(g, me.ID)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{U}{C}{C}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, garden, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("transform: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(garden) })
	pilot := b12Creature(g, me.ID, "Pilot", "Creature — Human", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, garden, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{pilot}}); err != nil {
		t.Fatalf("crew 1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, garden); !v.IsCreature() || v.CurrentPower() != 5 || v.CurrentToughness() != 4 {
		t.Fatalf("crewed Garden: creature %v, %d/%d, want a 5/4 creature", v.IsCreature(), v.CurrentPower(), v.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, g, garden, "flying") {
		t.Error("Balamb Garden, Airborne has no flying")
	}
	hand := me.Hand.Size()
	finABAttack(t, g, opp.ID, garden)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

// --- Balamb T-Rexaur ----------------------------------------------------------------

func TestFinABTRexaurGainsThreeOnEntryAndForestcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castCatalogSpell(t, g, "Balamb T-Rexaur", "Creature — Dinosaur", finABTRexaurOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	ids := seedSearchLibrary(me2,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
	)
	cycleFromHand(t, g2, "Balamb T-Rexaur", "Creature — Dinosaur", finABTRexaurOracle, "{C}{C}")
	passPriorityAroundTable(t, g2)
	c := searchChoiceFor(g2, me2.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("search = %+v, want only the Forest", c)
	}
}

// --- Balthier and Fran --------------------------------------------------------------

func TestFinABBalthierAndFranPumpsVehiclesAndPaysForASecondCombat(t *testing.T) {
	for _, pay := range []bool{true, false} {
		t.Run(map[bool]string{true: "pays", false: "declines"}[pay], func(t *testing.T) {
			s := newCrewScene(t)
			bf := pushCatalogPermanent(s.g, s.me.ID, "Balthier and Fran", "Legendary Creature — Human Rabbit", finABBalthierOracle, false)
			g := s.g
			g.WithWriteLock(func() {
				if c := findBattlefieldCardByID(g, bf); c != nil {
					c.Power, c.Toughness = 4, 3
				}
			})
			jet := s.pushVehicle("Dependable Quinjet", finABQuinjetOracle, 3, 3)
			if v := s58p6View(t, g, jet); v.CurrentPower() != 4 || v.CurrentToughness() != 4 {
				t.Errorf("the Vehicle is %d/%d, want 4/4", v.CurrentPower(), v.CurrentToughness())
			}
			for _, kw := range []string{"reach", "vigilance"} {
				if !hasEffectiveKeyword(t, g, jet, kw) {
					t.Errorf("the Vehicle has no %s", kw)
				}
			}
			s.crewAndResolve(t, jet, 0, bf)
			finABAttack(t, g, s.opp.ID, jet)
			passPriorityAroundTable(t, g)
			if isTapped(g, jet) {
				t.Error("the Vehicle has vigilance but tapped to attack")
			}
			if !hasPayUnlessFor(g, s.me.ID) {
				t.Fatal("no pay prompt when the Vehicle Balthier and Fran crewed attacked")
			}
			fillPool(s.me, 1)
			fillPoolColored(s.me, "R", 1)
			fillPoolColored(s.me, "G", 1)
			answerPayUnless(t, g, s.me.ID, pay)
			advanceTo(t, g, game.StepEndCombat)
			next := stepOnce(t, g)
			if pay && next.Step != game.StepBeginCombat {
				t.Errorf("paid: after combat %+v, want another combat", next)
			}
			if !pay && next.Step != game.StepPostcombatMain {
				t.Errorf("declined: after combat %+v, want the main phase", next)
			}
		})
	}
}

func TestFinABBalthierAndFranIgnoresAVehicleSomeoneElseCrewed(t *testing.T) {
	s := newCrewScene(t)
	pushCatalogPermanent(s.g, s.me.ID, "Balthier and Fran", "Legendary Creature — Human Rabbit", finABBalthierOracle, false)
	jet := s.pushVehicle("Dependable Quinjet", finABQuinjetOracle, 3, 3)
	crew := s.pushCreature("Crew", "Creature — Human", 4, 4)
	s.crewAndResolve(t, jet, 0, crew)
	finABAttack(t, s.g, s.opp.ID, jet)
	passPriorityAroundTable(t, s.g)
	if hasPayUnlessFor(s.g, s.me.ID) {
		t.Error("a Vehicle another creature crewed offered the extra combat")
	}
}

// --- Bard's Bow -------------------------------------------------------------------

func TestFinABBardsBowMakesAReachBardHero(t *testing.T) {
	g := newCatalogGame(t)
	bow := castCatalogSpell(t, g, "Bard's Bow", "Artifact — Equipment", finABBardsBowOracle, nil)
	passPriorityAroundTable(t, g)
	hero := ysBHero(t, g, bow)
	v := s58p6View(t, g, hero)
	if v.CurrentPower() != 3 || v.CurrentToughness() != 3 || !v.HasSubtype("Bard") || !v.HasSubtype("Hero") {
		t.Errorf("Hero is %d/%d %v, want a 3/3 Hero Bard", v.CurrentPower(), v.CurrentToughness(), v.Effective().Subtypes)
	}
	if !hasEffectiveKeyword(t, g, hero, "reach") {
		t.Error("the Hero has no reach")
	}
}

// --- Bartz and Boko ---------------------------------------------------------------

func TestFinABBartzAndBokoCostsLessPerBird(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, finABBartzOracle, "Legendary Creature — Human Bird", "{3}{G}{G}"); got != 5 {
		t.Errorf("with no Bird it costs %d, want 5", got)
	}
	b12Creature(g, me.ID, "Chocobo", "Creature — Bird", 2, 2)
	b12Creature(g, me.ID, "Moogle", "Creature — Moogle", 1, 1)
	if got := selfPricedMV(t, g, me, finABBartzOracle, "Legendary Creature — Human Bird", "{3}{G}{G}"); got != 4 {
		t.Errorf("with one Bird it costs %d, want 4", got)
	}
}

func TestFinABBartzAndBokoHasEachOtherBirdStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	b12Creature(g, me.ID, "Chocobo", "Creature — Bird", 2, 2)
	b12Creature(g, me.ID, "Fat Chocobo", "Creature — Bird", 3, 3)
	b12Creature(g, me.ID, "Moogle", "Creature — Moogle", 4, 4)
	big := b12Creature(g, opp.ID, "Behemoth", "Creature — Beast", 6, 6)
	bartz := castCatalogSpell(t, g, "Bartz and Boko", "Legendary Creature — Human Bird", finABBartzOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bartz) {
		t.Fatal("Bartz and Boko did not resolve")
	}
	pickTriggerTarget(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, big); got != 5 {
		t.Errorf("the target has %d damage marked, want 5 (the two other Birds only)", got)
	}
}

// --- Battle Menu --------------------------------------------------------------------

func TestFinABBattleMenuModes(t *testing.T) {
	t.Run("attack", func(t *testing.T) {
		g := newCatalogGame(t)
		castModal(t, g, "Battle Menu", "Instant", finABBattleMenuOracle, []int{0}, nil)
		passPriorityAroundTable(t, g)
		knight := findBattlefieldByName(g, "Knight")
		if knight == uuid.Nil {
			t.Fatal("no Knight token")
		}
		if v := s58p6View(t, g, knight); v.CurrentPower() != 2 || v.CurrentToughness() != 2 {
			t.Errorf("the Knight is %d/%d, want 2/2", v.CurrentPower(), v.CurrentToughness())
		}
	})
	t.Run("ability", func(t *testing.T) {
		g := newCatalogGame(t)
		bear := b12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
		castModal(t, g, "Battle Menu", "Instant", finABBattleMenuOracle, []int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: bear, Mode: 0}})
		passPriorityAroundTable(t, g)
		if v := s58p6View(t, g, bear); v.CurrentPower() != 2 || v.CurrentToughness() != 6 {
			t.Errorf("the creature is %d/%d, want 2/6", v.CurrentPower(), v.CurrentToughness())
		}
	})
	t.Run("magic", func(t *testing.T) {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
		big := b12Creature(g, opp.ID, "Big", "Creature — Beast", 4, 4)
		advanceToMain(t, g)
		if _, err := castModalErr(t, g, "Battle Menu", "Instant", finABBattleMenuOracle, []int{2}, []game.TargetRef{{Kind: game.TargetCard, ID: small, Mode: 0}}); err == nil {
			t.Fatal("the Magic mode targeted a creature with power 3")
		}
		castModal(t, g, "Battle Menu", "Instant", finABBattleMenuOracle, []int{2}, []game.TargetRef{{Kind: game.TargetCard, ID: big, Mode: 0}})
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(big) {
			t.Error("the power-4 creature survived")
		}
	})
	t.Run("item", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		life := me.Life
		castModal(t, g, "Battle Menu", "Instant", finABBattleMenuOracle, []int{3}, nil)
		passPriorityAroundTable(t, g)
		if me.Life != life+4 {
			t.Errorf("life %d → %d, want +4", life, me.Life)
		}
	})
}

// --- Beatrix, Loyal General ---------------------------------------------------------

func TestFinABBeatrixAttachesAnyNumberOfYourEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	pushCatalogPermanent(g, me.ID, "Beatrix, Loyal General", "Legendary Creature — Human Soldier", finABBeatrixOracle, false)
	host := b12Creature(g, me.ID, "Host", "Creature — Human", 2, 2)
	sword := finABEquipment(g, me.ID, "Sword")
	shield := finABEquipment(g, me.ID, "Shield")
	theirs := finABEquipment(g, opp.ID, "Their Sword")
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, host)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to attach Equipment")
	}
	if len(pick.ChooseCards) != 2 || hasID(pick.ChooseCards, theirs) || pick.ChooseMin != 0 {
		t.Fatalf("offered %v [min %d], want your two Equipment, optional", pick.ChooseCards, pick.ChooseMin)
	}
	answerChooseCards(t, g, me.ID, sword, shield)
	for _, id := range []uuid.UUID{sword, shield} {
		if c, ok := battlefieldCard(g, id); !ok || c.AttachedTo.ID != host {
			t.Errorf("an Equipment is not attached to the target: %+v", c.AttachedTo)
		}
	}
}

// --- Black Mage's Rod ---------------------------------------------------------------

func TestFinABBlackMagesRodPingsEachOpponentOnANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rod := castCatalogSpell(t, g, "Black Mage's Rod", "Artifact — Equipment", finABBlackRodOracle, nil)
	passPriorityAroundTable(t, g)
	hero := ysBHero(t, g, rod)
	v := s58p6View(t, g, hero)
	if v.CurrentPower() != 2 || v.CurrentToughness() != 1 || !v.HasSubtype("Wizard") {
		t.Errorf("Hero is %d/%d %v, want a 2/1 Hero Wizard", v.CurrentPower(), v.CurrentToughness(), v.Effective().Subtypes)
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	b24CastNoncreature(t, g, "Shock", "Instant", "{R}")
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID] - 1
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat life %d → %d, want %d", lives[p.ID], p.Life, want)
		}
	}
	creatures := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		creatures[p.ID] = p.Life
	}
	b24CastNoncreature(t, g, "Bear", "Creature — Bear", "{1}{G}")
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != creatures[p.ID] {
			t.Error("a creature spell set off the Rod's trigger")
		}
	}
}

// --- Blazing Bomb -------------------------------------------------------------------

func TestFinABBlazingBombGrowsOnlyOnFourManaNoncreatureSpells(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, pool string
		want                       int
	}{
		{"three mana", "Sorcery", "{2}{R}", "RRR", 0},
		{"four mana", "Sorcery", "{3}{R}", "RRRR", 1},
		{"four mana creature", "Creature — Bear", "{3}{G}", "GGGG", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bomb := pushCatalogPermanent(g, me.ID, "Blazing Bomb", "Creature — Elemental", finABBombOracle, false)
			advanceToMain(t, g)
			id := socB17Hand(me, "Test Spell", tc.typeLine, tc.cost, socB17PlainSorceryOracle)
			floatForTest(g, me, tc.pool)
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast %s: %v", tc.cost, err)
			}
			passPriorityAroundTable(t, g)
			if got := countersOn(g, bomb, game.CounterPlusOne); got != tc.want {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.want)
			}
		})
	}
}

func TestFinABBlazingBombBlowsUpForItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	advanceToMain(t, g)
	bomb := b12Push(g, me.ID, "Blazing Bomb", "Creature — Elemental", finABBombOracle, 3, 3)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{Targets: finABCard(victim)}); err != nil {
		t.Fatalf("Blow Up: %v", err)
	}
	if g.Battlefield.Contains(bomb) {
		t.Fatal("the Bomb was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 3/3 survived three damage")
	}
}

func TestFinABBlazingBombBlowsUpOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	bomb := b12Push(g, me.ID, "Blazing Bomb", "Creature — Elemental", finABBombOracle, 1, 1)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 1)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{Targets: finABCard(victim)}); err == nil {
		t.Error("Blow Up was activated in the beginning of combat step")
	}
}

// --- Blitzball Shot -----------------------------------------------------------------

func TestFinABBlitzballShotPumpsAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Blitzball Shot", "Instant", finABBlitzballOracle, finABCard(bear))
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, bear); v.CurrentPower() != 5 || v.CurrentToughness() != 5 {
		t.Errorf("the creature is %d/%d, want 5/5", v.CurrentPower(), v.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the creature has no trample")
	}
}
