package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b15_cards_test.go — Avatar: The Last Airbender batch 15 (#4459).

const (
	avB15InvasionTacticsOracle = "eef63cb4-96de-4c9d-84aa-62a3b52a3c7f"
	avB15KyoshiPlazaOracle     = "c0cdc333-6c5a-401c-9810-220d2027f2fe"
	avB15LeavesOracle          = "7c4abd72-29b0-4c82-9e67-b481cf7e8775"
	avB15MetalbendingOracle    = "177113f8-1188-4e5e-8ad7-34a7bb8ca7f1"
	avB15OstrichHorseOracle    = "2b6cb199-1564-4214-b24f-624dc962c907"
	avB15PillarLaunchOracle    = "a39c758d-82be-4d8c-bb3e-8ee8a6c6788b"
	avB15RaucousAudienceOracle = "5be6f245-31cf-4294-ace0-ca6ce3463fed"
	avB15CaptivesOracle        = "70401db7-edbc-402f-98ef-be74bd316273"
	avB15RockyRebukeOracle     = "fd85db70-8616-42df-9dac-bad7923f88d7"
	avB15MooseLionOracle       = "7af493e3-15a1-42d6-a66c-82aacabe816f"
)

func TestAvatarB15CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB15InvasionTacticsOracle: "Invasion Tactics",
		avB15KyoshiPlazaOracle:     "Kyoshi Island Plaza",
		avB15LeavesOracle:          "Leaves from the Vine",
		avB15MetalbendingOracle:    "Origin of Metalbending",
		avB15OstrichHorseOracle:    "Ostrich-Horse",
		avB15PillarLaunchOracle:    "Pillar Launch",
		avB15RaucousAudienceOracle: "Raucous Audience",
		avB15CaptivesOracle:        "Rebellious Captives",
		avB15RockyRebukeOracle:     "Rocky Rebuke",
		avB15MooseLionOracle:       "Saber-Tooth Moose-Lion",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s is not Full", name)
		}
		if len(spec.Caveats) != 0 {
			t.Errorf("%s has caveats: %v", name, spec.Caveats)
		}
	}
}

// --- shared test helpers ----------------------------------------------------------

// avB15Cast floats `mana` for the active seat and casts a card from its
// hand strictly, so a refusal is the card's own (a target, a mode),
// never an unpaid cost.
func avB15Cast(t *testing.T, g *game.Game, name, typeLine, manaCost, oracle, mana string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := handCardFull(me, name, typeLine, manaCost, oracle, []string{"G"})
	if mana != "" {
		g.WithWriteLock(func() {
			if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
				t.Fatalf("AddManaForEffect(%s): %v", mana, err)
			}
		})
	}
	params.Strict = true
	return id, g.CastSpell(me.ID, id, params)
}

// avB15CastCreature casts a creature with printed P/T (unpaid) and lets
// the spell resolve, leaving any enters trigger to the caller.
func avB15CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// avB15Settle passes priority until the stack is empty or a prompt is
// waiting.
func avB15Settle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 6 && g.Stack.Size() > 0 && len(g.PendingChoices) == 0; i++ {
		passPriorityAroundTable(t, g)
	}
}

func avB15Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func avB15Cards(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func avB15Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// --- Invasion Tactics ---------------------------------------------------------------

func TestAvatarB15InvasionTacticsPumpsYourCreaturesAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	castCatalogSpell(t, g, "Invasion Tactics", "Enchantment", avB15InvasionTacticsOracle, nil)
	avB15Settle(t, g)
	if p, tt := effectivePTOf(t, g, mine); p != 4 || tt != 4 {
		t.Errorf("your creature is %d/%d, want 4/4", p, tt)
	}
	if p, tt := effectivePTOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tt)
	}
}

func TestAvatarB15InvasionTacticsDrawsOncePerPlayerAlliesConnectWith(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Invasion Tactics", "Enchantment", avB15InvasionTacticsOracle, false)
	allyA := b12Creature(g, me.ID, "Ally A", "Creature — Human Ally", 2, 2)
	allyB := b12Creature(g, me.ID, "Ally B", "Creature — Human Ally", 2, 2)
	hand := me.Hand.Size()
	trekB38Attack(t, g, opp.ID, allyA, allyB)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards for two Allies hitting one player, want 1", got)
	}
}

func TestAvatarB15InvasionTacticsIgnoresANonAlly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Invasion Tactics", "Enchantment", avB15InvasionTacticsOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	trekB38Attack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("a non-Ally's damage drew %d cards", got)
	}
}

// --- Kyoshi Island Plaza -------------------------------------------------------------

func TestAvatarB15KyoshiIslandPlazaFetchesOneBasicPerShrine(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "Old Shrine", "Enchantment — Shrine")
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Mountain", "Basic Land — Mountain"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	castCatalogSpell(t, g, "Kyoshi Island Plaza", "Legendary Enchantment — Shrine", avB15KyoshiPlazaOracle, nil)
	avB15Settle(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, ids[:3]); err == nil {
		t.Fatal("took three basics with two Shrines")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, ids[:2]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	for _, id := range ids[:2] {
		if !onBattlefield(g, id) {
			t.Fatalf("a fetched basic is not on the battlefield")
		}
		if !avB15Tapped(t, g, id) {
			t.Error("a fetched basic entered untapped")
		}
	}
}

func TestAvatarB15KyoshiIslandPlazaFetchesWhenAnotherShrineEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Kyoshi Island Plaza", "Legendary Enchantment — Shrine", avB15KyoshiPlazaOracle, false)
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	// Your non-Shrine enchantment triggers nothing.
	castCatalogSpell(t, g, "Plain Enchantment", "Enchantment", "", nil)
	avB15Settle(t, g)
	if searchChoiceFor(g, me.ID) != nil || g.Stack.Size() != 0 {
		t.Fatal("a non-Shrine triggered the Plaza")
	}
	castCatalogSpell(t, g, "New Shrine", "Enchantment — Shrine", "", nil)
	avB15Settle(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("another Shrine entering did not search")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, ids[:1]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !onBattlefield(g, ids[0]) || !avB15Tapped(t, g, ids[0]) {
		t.Error("the Forest is not on the battlefield tapped")
	}
}

// --- Leaves from the Vine -------------------------------------------------------------

func TestAvatarB15LeavesFromTheVineRunsAllThreeChapters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seat := g.Turn.ActiveSeat
	advanceToMain(t, g)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Milled Sorcery", TypeLine: "Sorcery"},
		game.Card{Name: "Milled Instant", TypeLine: "Instant"},
		game.Card{Name: "Spare A", TypeLine: "Sorcery"},
		game.Card{Name: "Spare B", TypeLine: "Sorcery"},
		game.Card{Name: "Spare C", TypeLine: "Sorcery"},
		game.Card{Name: "Spare D", TypeLine: "Sorcery"},
	)
	saga := castCatalogSpell(t, g, "Leaves from the Vine", "Enchantment — Saga", avB15LeavesOracle, nil)
	avB15Settle(t, g)
	for _, id := range ids[:3] {
		if !me.Graveyard.Contains(id) {
			t.Fatal("chapter I did not mill the top three cards")
		}
	}
	if n, _ := trekB38Named(g, me.ID, "Food"); n != 1 {
		t.Errorf("%d Food tokens after chapter I, want 1", n)
	}

	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("chapter II asked for no targets")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, avB15Cards(a, b)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	avB15Settle(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := countersOn(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("chapter II put %d counters on a target, want 1", got)
		}
	}

	hand := me.Hand.Size()
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	avB15Settle(t, g)
	// The draw step drew one; chapter III draws another off the milled Bear.
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand grew by %d over the turn, want 2 (draw step and chapter III)", got)
	}
	if onBattlefield(g, saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestAvatarB15LeavesFromTheVineChapterThreeNeedsACreatureOrLesson(t *testing.T) {
	for _, tc := range []struct {
		name      string
		graveCard game.Card
		draws     int
	}{
		{"sorcery only", game.Card{Name: "Old Sorcery", TypeLine: "Sorcery"}, 0},
		{"a Lesson", game.Card{Name: "Old Lesson", TypeLine: "Sorcery — Lesson"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			g.WithWriteLock(func() {
				me.Graveyard.Cards = nil
				c := tc.graveCard
				c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
				me.Graveyard.PushTop(c)
			})
			hand := me.Hand.Size()
			item := &game.StackItem{Controller: me.ID}
			spec, _ := Lookup(avB15LeavesOracle)
			var err error
			g.WithWriteLock(func() { err = spec.Triggered[2].Effect(g, item) })
			if err != nil {
				t.Fatalf("chapter III: %v", err)
			}
			if got := me.Hand.Size() - hand; got != tc.draws {
				t.Errorf("drew %d, want %d", got, tc.draws)
			}
		})
	}
}

// --- Origin of Metalbending ---------------------------------------------------------

func TestAvatarB15OriginOfMetalbendingDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Pet Rock", "Artifact")
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if _, err := avB15Cast(t, g, "Origin of Metalbending", "Instant — Lesson", "{1}{G}", avB15MetalbendingOracle, "{G}{G}",
		game.CastSpellParams{Modes: []int{0}, Targets: avB15Card(bear)}); err == nil {
		t.Fatal("the destroy mode targeted a creature")
	}
	if _, err := avB15Cast(t, g, "Origin of Metalbending", "Instant — Lesson", "{1}{G}", avB15MetalbendingOracle, "{G}{G}",
		game.CastSpellParams{Modes: []int{0}, Targets: avB15Card(rock)}); err != nil {
		t.Fatalf("destroy mode: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, rock) {
		t.Error("the artifact survived")
	}
}

func TestAvatarB15OriginOfMetalbendingCountersAndProtectsYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if _, err := avB15Cast(t, g, "Origin of Metalbending", "Instant — Lesson", "{1}{G}", avB15MetalbendingOracle, "{G}{G}",
		game.CastSpellParams{Modes: []int{1}, Targets: avB15Card(theirs)}); err == nil {
		t.Fatal("the counter mode targeted an opponent's creature")
	}
	if _, err := avB15Cast(t, g, "Origin of Metalbending", "Instant — Lesson", "{1}{G}", avB15MetalbendingOracle, "{G}{G}",
		game.CastSpellParams{Modes: []int{1}, Targets: avB15Card(mine)}); err != nil {
		t.Fatalf("counter mode: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("your creature did not gain indestructible")
	}
}

// --- Ostrich-Horse -------------------------------------------------------------------

// avB15OstrichLibrary seeds the library with `top` (top first) above
// one more card the mill doesn't reach.
func avB15OstrichLibrary(me *game.Player, top ...game.Card) []uuid.UUID {
	cards := append(append([]game.Card(nil), top...), game.Card{Name: "Bottom", TypeLine: "Sorcery"})
	return seedSearchLibrary(me, cards...)
}

func TestAvatarB15OstrichHorseTakesAMilledLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := avB15OstrichLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Spell", TypeLine: "Sorcery"},
	)
	horse := avB15CastCreature(t, g, "Ostrich-Horse", "Creature — Bird Horse", avB15OstrichHorseOracle, 3, 1)
	avB15Settle(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no land pick after the mill")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != ids[0] {
		t.Fatalf("offered %v, want only the milled Forest", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, ids[0])
	if !me.Hand.Contains(ids[0]) {
		t.Error("the Forest is not in hand")
	}
	if got := countersOn(g, horse, game.CounterPlusOne); got != 0 {
		t.Errorf("took the land and still got %d counters", got)
	}
}

func TestAvatarB15OstrichHorseDecliningTheLandGivesACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := avB15OstrichLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Spell", TypeLine: "Sorcery"},
	)
	horse := avB15CastCreature(t, g, "Ostrich-Horse", "Creature — Bird Horse", avB15OstrichHorseOracle, 3, 1)
	avB15Settle(t, g)
	answerChooseCards(t, g, me.ID)
	if !me.Graveyard.Contains(ids[0]) {
		t.Error("the declined Forest left the graveyard")
	}
	if got := countersOn(g, horse, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after declining, want 1", got)
	}
}

func TestAvatarB15OstrichHorseWithNoLandMilledGetsACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	avB15OstrichLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Spell", TypeLine: "Sorcery"},
		game.Card{Name: "Other Spell", TypeLine: "Instant"},
	)
	horse := avB15CastCreature(t, g, "Ostrich-Horse", "Creature — Bird Horse", avB15OstrichHorseOracle, 3, 1)
	avB15Settle(t, g)
	if latestChooseCards(g, me.ID) != nil {
		t.Error("asked for a land with none milled")
	}
	if got := countersOn(g, horse, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
}

// --- Pillar Launch -------------------------------------------------------------------

func TestAvatarB15PillarLaunchPumpsGrantsReachAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, bear).Tapped = true })
	if _, err := avB15Cast(t, g, "Pillar Launch", "Instant", "{G}", avB15PillarLaunchOracle, "{G}",
		game.CastSpellParams{Targets: avB15Card(bear)}); err != nil {
		t.Fatalf("Pillar Launch: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := effectivePTOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tt)
	}
	if !hasEffectiveKeyword(t, g, bear, "reach") {
		t.Error("the target did not gain reach")
	}
	if avB15Tapped(t, g, bear) {
		t.Error("the target was not untapped")
	}
}

// --- Raucous Audience ----------------------------------------------------------------

func TestAvatarB15RaucousAudienceAddsMoreWithAFourPowerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	aud := pushCatalogPermanent(g, me.ID, "Raucous Audience", "Creature — Human Citizen", avB15RaucousAudienceOracle, false)
	if err := g.ActivateManaAbility(me.ID, aud, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := poolCount(me, "G"); got != 1 {
		t.Fatalf("added %d {G} with no big creature, want 1", got)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, aud).Tapped = false })
	pushVanillaCreature(g, me.ID, "Big Bear", 4, 4)
	if err := g.ActivateManaAbility(me.ID, aud, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := poolCount(me, "G"); got != 2 {
		t.Errorf("added %d {G} with a 4-power creature, want 2", got)
	}
}

// --- Rebellious Captives -------------------------------------------------------------

func TestAvatarB15RebelliousCaptivesGrowsAndEarthbendsOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	captives := pushCatalogPermanent(g, me.ID, "Rebellious Captives", "Creature — Human Peasant Ally", avB15CaptivesOracle, false)
	land := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	other := pushEarthbendLand(g, me.ID, "Mountain", "Basic Land — Mountain")
	theirLand := pushEarthbendLand(g, opp.ID, "Island", "Basic Land — Island")

	if err := trekB38Activate(t, g, me.ID, captives, 0, "{C}{C}{C}{C}{C}{C}",
		game.ActivateAbilityParams{Targets: avB15Card(theirLand)}); err == nil {
		t.Fatal("earthbent an opponent's land")
	}
	if err := trekB38Activate(t, g, me.ID, captives, 0, "",
		game.ActivateAbilityParams{Targets: avB15Card(land)}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, captives, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters on Rebellious Captives, want 2", got)
	}
	if got := countersOn(g, land, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters on the earthbent land, want 2", got)
	}
	if p, tt, creatureLand, hasty := earthbentBody(t, g, land); p != 2 || tt != 2 || !creatureLand || !hasty {
		t.Errorf("earthbent land is %d/%d creature=%v haste=%v, want a 2/2 hasty land creature", p, tt, creatureLand, hasty)
	}
	err := trekB38Activate(t, g, me.ID, captives, 0, "{C}{C}{C}{C}{C}{C}",
		game.ActivateAbilityParams{Targets: avB15Card(other)})
	if !errors.Is(err, game.ErrAbilityExhausted) {
		t.Errorf("second activation: %v, want ErrAbilityExhausted", err)
	}
}

// --- Rocky Rebuke --------------------------------------------------------------------

func TestAvatarB15RockyRebukeBitesAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	biter := pushVanillaCreature(g, me.ID, "Big Bear", 4, 4)
	mine := pushVanillaCreature(g, me.ID, "My Other Bear", 3, 3)
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 3, 4)
	if _, err := avB15Cast(t, g, "Rocky Rebuke", "Instant", "{1}{G}", avB15RockyRebukeOracle, "{G}{G}",
		game.CastSpellParams{Targets: avB15Cards(biter, mine)}); err == nil {
		t.Fatal("bit your own creature")
	}
	if _, err := avB15Cast(t, g, "Rocky Rebuke", "Instant", "{1}{G}", avB15RockyRebukeOracle, "{G}{G}",
		game.CastSpellParams{Targets: avB15Cards(victim, biter)}); err == nil {
		t.Fatal("an opponent's creature was the biter")
	}
	if _, err := avB15Cast(t, g, "Rocky Rebuke", "Instant", "{1}{G}", avB15RockyRebukeOracle, "{G}{G}",
		game.CastSpellParams{Targets: avB15Cards(biter, victim)}); err != nil {
		t.Fatalf("Rocky Rebuke: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, victim) {
		t.Error("the 3/4 survived 4 damage")
	}
	if c, ok := battlefieldCard(g, biter); !ok || c.DamageMarked != 0 {
		t.Error("the biter took damage back")
	}
}

// --- Saber-Tooth Moose-Lion ----------------------------------------------------------

func TestAvatarB15SaberToothMooseLionHasReachAndForestcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lion := pushCatalogPermanent(g, me.ID, "Saber-Tooth Moose-Lion", "Creature — Elk Cat", avB15MooseLionOracle, false)
	if !hasEffectiveKeyword(t, g, lion, "reach") {
		t.Error("Saber-Tooth Moose-Lion has no reach")
	}
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id, _ := cycleFromHand(t, g, "Saber-Tooth Moose-Lion", "Creature — Elk Cat", avB15MooseLionOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("search = %+v, want only the Forest", c)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the Forest did not reach the hand")
	}
}
