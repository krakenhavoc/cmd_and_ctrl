package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB87ThirdOracle      = "2a73db99-5e31-49a8-91e7-b28b3434585d"
	gapB87ThirteenthOracle = "3778e06c-62c7-4b8f-90ff-b2fc2d2e213c"
	gapB87AlistairOracle   = "2fe8d739-dee5-4220-8102-6fbae7d4fc19"
	gapB87BentonOracle     = "987a0b61-2ce4-47bd-8f5e-1e7bf01d0869"
	gapB87VrestinOracle    = "1db6a042-303e-4b76-bb6c-63df846fd6c9"
	gapB87AmyOracle        = "aa7d3e99-fc14-4f62-bf00-c0428c11e4f2"
	gapB87BillOracle       = "0ebe67ed-8a71-4bbc-8e58-cebbd7b2d83f"
	gapB87DanOracle        = "b84ab039-901e-4635-b57e-3f2dac239f38"
	gapB87DonnaOracle      = "0d6e76aa-2966-48b7-be4e-2d3e2cc94fc1"
	gapB87TyrannoOracle    = "89505f24-6c96-4062-8171-1851f7e7395b"
	gapB87IraxxaOracle     = "52925539-423a-42ec-91b9-0d39e7df565c"
	gapB87DoctorType       = "Legendary Creature — Time Lord Doctor"
)

func TestGapB87EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB87ThirdOracle:      "The Third Doctor",
		gapB87ThirteenthOracle: "The Thirteenth Doctor",
		gapB87AlistairOracle:   "Alistair, the Brigadier",
		gapB87BentonOracle:     "Sergeant John Benton",
		gapB87VrestinOracle:    "Vrestin, Menoptra Leader",
		gapB87AmyOracle:        "Amy Pond",
		gapB87BillOracle:       "Bill Potts",
		gapB87DanOracle:        "Dan Lewis",
		gapB87DonnaOracle:      "Donna Noble",
		gapB87TyrannoOracle:    "Flaming Tyrannosaurus",
		gapB87IraxxaOracle:     "Iraxxa, Empress of Mars",
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

// gapB87CastFromExile casts a card from exile with a permission for p.
func gapB87CastFromExile(t *testing.T, g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.Cards = append(g.Exile.Cards, game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		_ = g.ExileCardWithPermissionForEffect(id, game.CastPermission{Player: p.ID, Duration: g.UntilEndOfTurnDuration()})
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast %s from exile: %v", name, err)
	}
	return id
}

func gapB87Named(g *game.Game, controller uuid.UUID, subtype string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.HasSubtype(subtype) {
			out = append(out, c)
		}
	}
	return out
}

func TestGapB87IraxxaMakesAnAlienWarriorForAParadoxCastAndHasBattleCry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	iraxxa := apaPush(g, me.ID, me.ID, game.Card{Name: "Iraxxa, Empress of Mars", OracleID: gapB87IraxxaOracle,
		TypeLine: "Legendary Creature — Alien Warrior", Power: 5, Toughness: 4})
	other := apaPush(g, me.ID, me.ID, game.Card{Name: "Other", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	castCatalogSpell(t, g, "Hand Spell", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if n := len(gapB87Named(g, me.ID, "Alien")); n != 1 {
		t.Fatalf("a hand cast made %d Aliens (Iraxxa herself is one)", n)
	}
	gapB87CastFromExile(t, g, me, "Exile Spell", "Instant")
	passPriorityAroundTable(t, g)
	aliens := gapB87Named(g, me.ID, "Alien")
	if len(aliens) != 2 {
		t.Fatalf("an exile cast made %d Aliens in total, want 2", len(aliens))
	}
	var token game.Card
	for _, a := range aliens {
		if a.IsToken() {
			token = a
		}
	}
	if token.Name == "" || token.CurrentPower() != 2 || token.CurrentToughness() != 2 || !token.HasColor("R") || !token.HasSubtype("Warrior") {
		t.Errorf("token %q %d/%d is not a 2/2 red Alien Warrior", token.TypeLine, token.CurrentPower(), token.CurrentToughness())
	}
	declareAttack(t, g, opp.ID, iraxxa, other)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, other); p != 3 {
		t.Errorf("battle cry: the other attacker has power %d, want 3", p)
	}
	if p := effectivePower(t, g, iraxxa); p != 5 {
		t.Errorf("Iraxxa has power %d, want 5", p)
	}
}

func TestGapB87FlamingTyrannosaurusBurnsGrowsAndDiesForItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rex := apaPush(g, me.ID, me.ID, game.Card{Name: "Flaming Tyrannosaurus", OracleID: gapB87TyrannoOracle,
		TypeLine: "Creature — Dinosaur", Power: 5, Toughness: 5})
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 3)
	gapB87CastFromExile(t, g, me, "Exile Spell", "Instant")
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("3 damage kills a 1/3")
	}
	if n := countersOn(g, rex, game.CounterPlusOne); n != 1 {
		t.Fatalf("the Tyrannosaurus has %d counters, want 1", n)
	}
	// It dies as a 6/6: each opponent takes 6.
	lives := lifeOfOpponents(g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rex) })
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != lives[i]-6 {
			t.Errorf("opponent %d took %d, want 6", i, lives[i]-p.Life)
		}
	}
}

func TestGapB87ThirteenthDoctorCountersATargetAndUntapsCounteredCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "The Thirteenth Doctor", gapB87DoctorType, gapB87ThirteenthOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	gapB87CastFromExile(t, g, me, "Exile Spell", "Instant")
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("the target has %d counters, want 1", n)
	}
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, bear).Tapped = true
		findBattlefieldCardForTest(g, plain).Tapped = true
	})
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if tappedForTest(t, g, bear) {
		t.Error("the creature with a counter stayed tapped")
	}
	if !tappedForTest(t, g, plain) {
		t.Error("a creature with no counter was untapped")
	}
}

func TestGapB87BentonMakesBothPlayersDrawTheCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	benton := apaPush(g, me.ID, me.ID, game.Card{Name: "Sergeant John Benton", OracleID: gapB87BentonOracle,
		TypeLine: "Legendary Creature — Human Soldier", Power: 2, Toughness: 4})
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	declareAttack(t, g, opp.ID, benton)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - mine; got != 2 {
		t.Errorf("Benton's controller drew %d, want 2", got)
	}
	if got := opp.Hand.Size() - theirs; got != 2 {
		t.Errorf("the damaged player drew %d, want 2", got)
	}
}

func TestGapB87AlistairMakesSoldiersForHistoricSpellsAndMayPayEightToPump(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	alistair := apaPush(g, me.ID, me.ID, game.Card{Name: "Alistair, the Brigadier", OracleID: gapB87AlistairOracle,
		TypeLine: "Legendary Creature — Human Soldier", Power: 3, Toughness: 3})
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	soldiers := 0
	for _, c := range gapB87Named(g, me.ID, "Soldier") {
		if c.IsToken() {
			soldiers++
		}
	}
	if soldiers != 1 {
		t.Fatalf("%d Soldier tokens, want 1 (only the historic spell counts)", soldiers)
	}
	// Attack and pay {8}: historic permanents are Alistair and the Rock = 2.
	declareAttack(t, g, opp.ID, alistair)
	g.WithWriteLock(func() { fillPool(me, 8) })
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt for Alistair's {8}")
	}
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, alistair); p != 5 {
		t.Errorf("Alistair's power %d, want 5 (3 + 2 historic permanents)", p)
	}
}

func TestGapB87ThirdDoctorGrowsPerNoncreatureTokenAndMakesYourChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	doc := castCatalogSpell(t, g, "The Third Doctor", gapB87DoctorType, gapB87ThirdOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if pick == nil {
		t.Fatal("no choice of Clue, Food or Treasure")
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, 2); err != nil {
		t.Fatalf("choosing a Treasure: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(gapB87Named(g, me.ID, "Treasure")); n != 1 {
		t.Fatalf("%d Treasures, want 1", n)
	}
	// The test card has no printed size, so the whole of its power is the bonus.
	b16Creature(g, me.ID, "Creature Token", "Token Creature — Bear", 1, 1)
	if p := effectivePower(t, g, doc); p != 1 {
		t.Errorf("with one noncreature token and a creature token the Doctor has power %d, want 1", p)
	}
	b12Permanent(g, me.ID, "Second Food", "Token Artifact — Food")
	if p := effectivePower(t, g, doc); p != 2 {
		t.Errorf("with two noncreature tokens the Doctor has power %d, want 2", p)
	}
}

func TestGapB87VrestinEntersWithXCountersAndMakesXInsectsThatGrowWhenTheyAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vrestin := paCast(t, g, "Vrestin, Menoptra Leader", "Legendary Creature — Alien Insect Scout", "{X}{G}{G}{W}", gapB87VrestinOracle,
		game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	if n := countersOn(g, vrestin, game.CounterPlusOne); n != 2 {
		t.Fatalf("Vrestin has %d counters, want 2", n)
	}
	var tokens []uuid.UUID
	for _, c := range gapB87Named(g, me.ID, "Insect") {
		if c.IsToken() {
			tokens = append(tokens, c.InstanceID)
			if c.CurrentPower() != 1 || !c.HasColor("G") || !c.HasColor("W") || !c.HasSubtype("Alien") {
				t.Errorf("token %q %d/%d is not a 1/1 green and white Alien Insect", c.TypeLine, c.CurrentPower(), c.CurrentToughness())
			}
		}
	}
	if len(tokens) != 2 {
		t.Fatalf("%d Insect tokens, want 2", len(tokens))
	}
	g.WithWriteLock(func() {
		for _, id := range append([]uuid.UUID{vrestin}, tokens...) {
			findBattlefieldCardForTest(g, id).SummonedThisTurn = false
		}
	})
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	declareAttack(t, g, opp.ID, tokens[0], vrestin, bear)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, tokens[0], game.CounterPlusOne); n != 1 {
		t.Errorf("an attacking Insect token has %d counters, want 1", n)
	}
	if n := countersOn(g, vrestin, game.CounterPlusOne); n != 3 {
		t.Errorf("Vrestin (an Insect) has %d counters, want 3", n)
	}
	if n := countersOn(g, tokens[1], game.CounterPlusOne); n != 0 {
		t.Errorf("an Insect that did not attack got %d counters", n)
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("a non-Insect got %d counters", n)
	}
}

func TestGapB87DonnaDealsTheDamageDoneToHerOrHerPartner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 5)
	other := b12Creature(g, me.ID, "Unpaired", "Creature — Bear", 3, 5)
	src := b16Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5, "R")
	donna := sbondPairCard(t, g, me, "Donna Noble", "Legendary Creature — Human", gapB87DonnaOracle, 2, 4, bear, game.KeywordSoulbond)
	life := opp.Life
	cmdr2B08Damage(g, src, bear, 2)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("target opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("damage to the partner: opponent life %d → %d, want -2", life, opp.Life)
	}
	cmdr2B08Damage(g, src, donna, 3)
	b04WaitForPick(t, g, me.ID)
	p = latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("target opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Fatalf("damage to Donna: opponent life %d, want %d", opp.Life, life-5)
	}
	cmdr2B08Damage(g, src, other, 4)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || opp.Life != life-5 {
		t.Error("damage to an unpaired creature triggered Donna")
	}
}

func TestGapB87AmyPondRemovesTimeCountersFromASuspendedCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	amy := apaPush(g, me.ID, me.ID, game.Card{Name: "Amy Pond", OracleID: gapB87AmyOracle,
		TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2})
	suspended := uuid.New()
	notMine := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.Cards = append(g.Exile.Cards,
			game.Card{InstanceID: suspended, Name: "Suspended", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterTime: 4}},
			game.Card{InstanceID: notMine, Name: "Theirs", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterTime: 4}})
	})
	declareAttack(t, g, opp.ID, amy)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	for _, c := range g.Exile.Cards {
		switch c.InstanceID {
		case suspended:
			if c.Counters[game.CounterTime] != 2 {
				t.Errorf("my suspended card has %d time counters, want 4 - 2", c.Counters[game.CounterTime])
			}
		case notMine:
			if c.Counters[game.CounterTime] != 4 {
				t.Errorf("an opponent's suspended card lost counters: %d", c.Counters[game.CounterTime])
			}
		}
	}
}

func TestGapB87BillPottsCopiesASpellThatTargetsOnlyHimOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bill := apaPush(g, me.ID, me.ID, game.Card{Name: "Bill Potts", OracleID: gapB87BillOracle,
		TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 4})
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	counts := func() int {
		n := 0
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemSpell {
				n++
			}
		}
		return n
	}
	// A spell that targets another creature does not trigger.
	handCardOf(g, me, "Shrink", "Instant", "", "")
	id := handCardOf(g, me, "Shrink", "Instant", "", "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(other)}); err != nil {
		t.Fatal(err)
	}
	if got := len(g.Stack.Cards); got != 1 {
		t.Fatalf("a spell not aimed at Bill: %d spells on the stack, want 1", got)
	}
	passPriorityAroundTable(t, g)
	_ = counts
	// A spell that targets only Bill is copied; the copy keeps its target.
	id = handCardOf(g, me, "Boost", "Instant", "", "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(bill)}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4 && len(g.Stack.Cards) < 2; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if len(g.Stack.Cards) < 2 {
		t.Fatalf("a spell targeting only Bill was not copied: %d on the stack", len(g.Stack.Cards))
	}
	passPriorityAroundTable(t, g)
	// The second one this turn is not copied.
	id = handCardOf(g, me, "Boost Two", "Instant", "", "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(bill)}); err != nil {
		t.Fatal(err)
	}
	passPriorityOnceNoResolve(t, g)
	if got := len(g.Stack.Cards); got != 1 {
		t.Errorf("the second Bill spell this turn was copied: %d spells on the stack", got)
	}
}

func passPriorityOnceNoResolve(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 2; i++ {
		if len(g.PendingTriggers) == 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGapB87DanLewisMakesNoncreatureArtifactsEquipmentWithEquipOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Dan Lewis", "Legendary Creature — Human", gapB87DanOracle, false)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	vehicle := b12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	advanceToMain(t, g)
	c := battlefieldCardOrFail(t, g, rock)
	var eff game.Card
	g.ReadSnapshot(func() { eff, _ = g.LookupCardForEffect(rock) })
	if !eff.HasSubtype("Equipment") {
		t.Fatalf("the artifact is %q, want an Equipment in addition", eff.TypeLine)
	}
	_ = c
	g.ReadSnapshot(func() { eff, _ = g.LookupCardForEffect(vehicle) })
	if eff.HasSubtype("Equipment") {
		t.Error("an artifact creature became an Equipment")
	}
	g.WithWriteLock(func() { fillPool(me, 1) })
	beEquip(t, g, me.ID, rock, bear, "Equip {1}")
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("the equipped creature has power %d, want 3 (+1/+0)", p)
	}
}

func battlefieldCardOrFail(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("card %s not on the battlefield", id)
	}
	return c
}
