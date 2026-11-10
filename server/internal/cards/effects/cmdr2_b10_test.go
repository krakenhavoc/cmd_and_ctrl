package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander wave 2, batch 10 (#5368): Vampires — Blood tokens, partner pairs,
// Athreos.

const (
	c2b10Eulogist  = "cf9190e0-6f53-44fb-8afa-c16f5bf9829e"
	c2b10Estate    = "fb0c0426-f1a6-4e52-9242-627786d3119a"
	c2b10Heart     = "3edd6272-20fa-4c3a-bf34-2f253cc8ac9e"
	c2b10Kamber    = "0b620775-70a5-4423-8453-e403a5b79554"
	c2b10Laurine   = "f2f8847d-17a8-4c26-8a7a-1bb6c02f2b39"
	c2b10Florian   = "d74a2cbf-4d13-401d-9b3f-bfbde4942afd"
	c2b10Timothar  = "82e03151-3990-451c-99cc-c7d386872df2"
	c2b10Athreos   = "9aacc1ce-8d0d-4a24-9a15-74635f2f127d"
	c2b10Cazur     = "ccfc74f7-bf32-4201-ad8f-f552ed7f96f8"
	c2b10Ukkima    = "b8c3cb84-0a30-420d-8fc3-bb85ce69b5e8"
	c2b10BloodName = "Blood"
)

func c2b10Blood(g *game.Game, owner uuid.UUID) int {
	n, _ := countTokensNamed(g, owner, c2b10BloodName)
	return n
}

func TestC2B10MoonstoneEulogistMakesBloodForOpponentDeathsAndGrowsOnArtifactSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eul := b12Push(g, me.ID, "Moonstone Eulogist", "Creature — Bat Warlock", c2b10Eulogist, 4, 4)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b15Destroy(t, g, mine)
	if got := c2b10Blood(g, me.ID); got != 0 {
		t.Fatalf("my own creature's death made %d Blood", got)
	}
	b15Destroy(t, g, theirs)
	if got := c2b10Blood(g, me.ID); got != 1 {
		t.Fatalf("Blood %d, want 1", got)
	}
	life := me.Life
	blood := findBattlefieldByName(g, c2b10BloodName)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(blood) })
	passPriorityAroundTable(t, g)
	if got := countersOn(g, eul, game.CounterPlusOne); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

func TestC2B10VoldarenEstateTapsForColorlessAndRestrictedColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	estate := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Voldaren Estate", TypeLine: "Land",
		OracleID: c2b10Estate, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, estate, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("any-color ability: %v", err)
	}
	b10ResolveAllManaPicks(t, g, me.ID, "R")
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	if me.ManaPool.Total() != 1 {
		t.Fatalf("pool %d, want 1", me.ManaPool.Total())
	}
	mk := func(typeLine string) error {
		id := putInHand(me, game.Card{Name: "Spell", TypeLine: typeLine, ManaCost: "{R}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	if err := mk("Creature — Human Warrior"); err == nil {
		t.Fatal("the Vampire-only mana paid for a Human")
	}
	if err := mk("Creature — Vampire Knight"); err != nil {
		t.Fatalf("the Vampire-only mana could not pay for a Vampire: %v", err)
	}
}

func TestC2B10VoldarenEstateBloodCostsOneLessPerVampire(t *testing.T) {
	try := func(vampires, mana int) error {
		g := newCatalogGame(t)
		me := g.Seats[0]
		estate := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Voldaren Estate", TypeLine: "Land",
			OracleID: c2b10Estate, Owner: me.ID, Controller: me.ID})
		for i := 0; i < vampires; i++ {
			b12Creature(g, me.ID, "Vamp", "Creature — Vampire", 1, 1)
		}
		b12Creature(g, me.ID, "Human", "Creature — Human", 1, 1)
		advanceToMain(t, g)
		b06AddMana(me, "B", "B", "B", "B", "B")
		me.ManaPool = game.ManaPool{}
		for i := 0; i < mana; i++ {
			me.ManaPool.AddMana(game.ManaToken{Color: "B"})
		}
		err := g.ActivateCatalogAbility(me.ID, estate, 0, game.ActivateAbilityParams{Strict: true})
		if err == nil {
			passPriorityAroundTable(t, g)
			if got := c2b10Blood(g, me.ID); got != 1 {
				t.Errorf("%d Vampires: Blood %d, want 1", vampires, got)
			}
		}
		return err
	}
	if err := try(0, 5); err != nil {
		t.Errorf("{5} with no Vampires: %v", err)
	}
	if err := try(0, 4); err == nil {
		t.Error("{4} paid for the undiscounted ability")
	}
	if err := try(2, 3); err != nil {
		t.Errorf("{3} with two Vampires: %v", err)
	}
	if err := try(2, 2); err == nil {
		t.Error("{2} paid with only two Vampires")
	}
	if err := try(5, 0); err != nil {
		t.Errorf("five Vampires make it free: %v", err)
	}
}

func TestC2B10GlassCastHeartMakesOneBloodPerAttackOfVampires(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Glass-Cast Heart", "Artifact", c2b10Heart, 0, 0)
	v1 := b12Creature(g, me.ID, "Vamp One", "Creature — Vampire", 1, 1)
	v2 := b12Creature(g, me.ID, "Vamp Two", "Creature — Vampire", 1, 1)
	declareAttack(t, g, opp.ID, v1, v2)
	passPriorityAroundTable(t, g)
	if got := c2b10Blood(g, me.ID); got != 1 {
		t.Fatalf("Blood %d, want 1 for two attacking Vampires", got)
	}
}

func TestC2B10GlassCastHeartIgnoresNonVampireAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Glass-Cast Heart", "Artifact", c2b10Heart, 0, 0)
	human := b12Creature(g, me.ID, "Human", "Creature — Human", 1, 1)
	declareAttack(t, g, opp.ID, human)
	passPriorityAroundTable(t, g)
	if got := c2b10Blood(g, me.ID); got != 0 {
		t.Fatalf("Blood %d, want 0", got)
	}
}

func TestC2B10GlassCastHeartMakesAVampireForOneLife(t *testing.T) {
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[0]
	heart := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Glass-Cast Heart", TypeLine: "Artifact",
		OracleID: c2b10Heart, Owner: me.ID, Controller: me.ID})
	life := me.Life
	b06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, heart, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Vampire"); n != 1 {
		t.Fatalf("Vampire tokens %d, want 1", n)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Vampire" && (!p1g2HasKeyword(g, c.InstanceID, "lifelink") || len(c.Colors) != 2) {
			t.Errorf("token colors %v, want white and black with lifelink", c.Colors)
		}
	}
}

func TestC2B10GlassCastHeartDrainsThirteenWithThirteenBloodTokens(t *testing.T) {
	setup := func(blood int) (*game.Game, uuid.UUID) {
		g := newCatalogGame(t)
		advanceTo(t, g, game.StepPrecombatMain)
		me := g.Seats[0]
		heart := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Glass-Cast Heart", TypeLine: "Artifact",
			OracleID: c2b10Heart, Owner: me.ID, Controller: me.ID})
		g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, BloodToken(), blood) })
		b06AddMana(me, "B", "B")
		return g, heart
	}
	// Twelve is not enough.
	g, heart := setup(12)
	if err := g.ActivateCatalogAbility(g.Seats[0].ID, heart, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with twelve Blood tokens")
	}
	g, heart = setup(13)
	me := g.Seats[0]
	var ids []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == c2b10BloodName {
			ids = append(ids, c.InstanceID)
		}
	}
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, heart, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: ids}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(heart) || c2b10Blood(g, me.ID) != 0 {
		t.Error("the Heart and the Blood tokens should all be sacrificed")
	}
	for _, s := range g.Seats {
		want := lives[s.ID] - 13
		if s.ID == me.ID {
			want = lives[s.ID] + 13
		}
		if s.Life != want {
			t.Errorf("seat %s life %d, want %d", s.ID, s.Life, want)
		}
	}
}

func TestC2B10KamberGainsLifeAndMakesBloodWhenAnOpponentsCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Kamber, the Plunderer", "Legendary Creature — Vampire Rogue", c2b10Kamber, 3, 4)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	b15Destroy(t, g, mine)
	if c2b10Blood(g, me.ID) != 0 || me.Life != life {
		t.Fatal("my own creature's death triggered Kamber")
	}
	b15Destroy(t, g, theirs)
	if got := c2b10Blood(g, me.ID); got != 1 {
		t.Errorf("Blood %d, want 1", got)
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

func c2b10PartnerFetch(t *testing.T, cast func(g *game.Game, me *game.Player), partnerName, partnerOracle, partnerType string) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 2; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Instant"})
	}
	partner := pushLibraryCardForTest(me, game.Card{Name: partnerName, TypeLine: partnerType, OracleID: partnerOracle})
	cast(g, me)
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	answerSearchNamed(t, g, me.ID, partnerName)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(partner) {
		t.Fatalf("%s did not reach the hand", partnerName)
	}
}

func TestC2B10PartnerPairsFetchEachOther(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, partner, partnerOracle, partnerType string
	}{
		{"Kamber, the Plunderer", c2b10Kamber, "Legendary Creature — Vampire Rogue", "Laurine, the Diversion", c2b10Laurine, "Legendary Creature — Human Rogue"},
		{"Laurine, the Diversion", c2b10Laurine, "Legendary Creature — Human Rogue", "Kamber, the Plunderer", c2b10Kamber, "Legendary Creature — Vampire Rogue"},
		{"Cazur, Ruthless Stalker", c2b10Cazur, "Legendary Creature — Human Warrior", "Ukkima, Stalking Shadow", c2b10Ukkima, "Legendary Creature — Whale Wolf"},
		{"Ukkima, Stalking Shadow", c2b10Ukkima, "Legendary Creature — Whale Wolf", "Cazur, Ruthless Stalker", c2b10Cazur, "Legendary Creature — Human Warrior"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c2b10PartnerFetch(t, func(g *game.Game, me *game.Player) {
				castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			}, tc.partner, tc.partnerOracle, tc.partnerType)
		})
	}
}

func TestC2B10LaurineSacrificesToGoad(t *testing.T) {
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	laurine := b12Push(g, me.ID, "Laurine, the Diversion", "Legendary Creature — Human Rogue", c2b10Laurine, 3, 3)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	target := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b06AddMana(me, "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, laurine, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice was not paid")
	}
	if b33Goaded(t, g, target) != me.ID {
		t.Error("the target is not goaded by Laurine's controller")
	}
}

func TestC2B10CazurGrowsAnyCreatureThatHitsAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Cazur, Ruthless Stalker", "Legendary Creature — Human Warrior", c2b10Cazur, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters on the attacker %d, want 1", got)
	}
}

func TestC2B10UkkimaCantBeBlockedAndHitsForItsPowerWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ukkima := b12Push(g, me.ID, "Ukkima, Stalking Shadow", "Legendary Creature — Whale Wolf", c2b10Ukkima, 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, ukkima)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, ukkima); err == nil {
		t.Fatal("Ukkima was blocked")
	}
	// Leaves: X is its last-known power, counters included.
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	u2 := b12Push(g2, me2.ID, "Ukkima, Stalking Shadow", "Legendary Creature — Whale Wolf", c2b10Ukkima, 5, 5)
	life, theirs := me2.Life, opp2.Life
	b15Destroy(t, g2, u2)
	answerPickTargetPlayer(t, g2, opp2.ID)
	passPriorityAroundTable(t, g2)
	if opp2.Life != theirs-5 {
		t.Errorf("target life %d, want %d", opp2.Life, theirs-5)
	}
	if me2.Life != life+5 {
		t.Errorf("my life %d, want %d", me2.Life, life+5)
	}
	_ = me
	_ = opp
}

func TestC2B10FlorianExilesOneOfTheTopXAndMayPlayIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flo := b12Push(g, me.ID, "Florian, Voldaren Scion", "Legendary Creature — Vampire Noble", c2b10Florian, 3, 3)
	me.Library.Cards = nil
	var top []uuid.UUID
	for _, n := range []string{"One", "Two", "Three", "Four"} {
		top = append(top, pushLibraryCardForTest(me, game.Card{Name: n, TypeLine: "Basic Land — Forest"}))
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(flo, opp.ID, -3) })
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no exile prompt after the opponent lost 3 life")
	}
	if len(prompt.ChooseCards) != 3 || hasID(prompt.ChooseCards, top[3]) {
		t.Fatalf("offered %v, want exactly the top three cards", prompt.ChooseCards)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top[1]) {
		t.Fatal("the chosen card is not exiled")
	}
	if me.Library.Size() != 3 {
		t.Fatalf("library %d cards, want 3", me.Library.Size())
	}
	if topCard, err := me.Library.Top(); err != nil || topCard.InstanceID != top[3] {
		t.Error("the card that was not looked at is not on top any more")
	}
	if err := g.CastSpell(me.ID, top[1], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Errorf("the exiled land can't be played this turn: %v", err)
	}
}

func TestC2B10FlorianDoesNothingWhenNoOpponentLostLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Florian, Voldaren Scion", "Legendary Creature — Vampire Noble", c2b10Florian, 3, 3)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if latestChooseCards(g, me.ID) != nil {
		t.Fatal("asked to exile a card with X = 0")
	}
}

func TestC2B10AthreosReturnsCreaturesWithCoinCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Athreos, Shroud-Veiled", "Legendary Enchantment Creature — God", c2b10Athreos, 4, 7)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no coin-counter target prompt")
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, victim, "coin"); got != 1 {
		t.Fatalf("coin counters %d, want 1", got)
	}
	// The return half, from a main phase.
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Athreos, Shroud-Veiled", "Legendary Enchantment Creature — God", c2b10Athreos, 4, 7)
	victim = b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	findBattlefieldCardForTest(g, victim).Counters = map[string]int{"coin": 1}
	b15Destroy(t, g, victim)
	if got := onBattlefieldNamed(g, "Their Bear"); got != 1 {
		t.Fatalf("Their Bear on the battlefield %d times, want 1 (returned)", got)
	}
	var back game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" {
			back = c
		}
	}
	if back.Controller != me.ID {
		t.Errorf("the returned creature is controlled by %s, want Athreos' controller", back.Controller)
	}
	if got := countersOn(g, back.InstanceID, "coin"); got != 0 {
		t.Errorf("the returned creature kept %d coin counters", got)
	}
}

func TestC2B10AthreosIgnoresCreaturesWithoutACoinCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Athreos, Shroud-Veiled", "Legendary Enchantment Creature — God", c2b10Athreos, 4, 7)
	plain := b12Creature(g, opp.ID, "Plain Bear", "Creature — Bear", 2, 2)
	b15Destroy(t, g, plain)
	if onBattlefieldNamed(g, "Plain Bear") != 0 {
		t.Fatal("a creature with no coin counter came back")
	}
}

func TestC2B10AthreosIsACreatureOnlyWithSevenDevotion(t *testing.T) {
	for _, tc := range []struct {
		pad  string
		want bool
	}{{"{W}{B}{B}", false}, {"{W}{W}{B}{B}{B}", true}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		// Athreos is {4}{W}{B}: two symbols of its own.
		id := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Athreos, Shroud-Veiled",
			TypeLine: "Legendary Enchantment Creature — God", OracleID: c2b10Athreos, ManaCost: "{4}{W}{B}",
			Power: 4, Toughness: 7, Owner: me.ID, Controller: me.ID})
		pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Devotion Pad", TypeLine: "Enchantment",
			ManaCost: tc.pad, Owner: me.ID, Controller: me.ID})
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		c := avB1Card(t, g, id)
		if c.IsCreature() != tc.want {
			t.Errorf("pad %s: creature=%v, want %v", tc.pad, c.IsCreature(), tc.want)
		}
	}
}

func TestC2B10TimotharExilesADeadVampireForABatThatReturnsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Timothar, Baron of Bats", "Legendary Creature — Vampire Noble", c2b10Timothar, 4, 4)
	vamp := b12Creature(g, me.ID, "Dead Vampire", "Creature — Vampire", 2, 2)
	b15Destroy(t, g, vamp)
	b06AddMana(me, "B") // after the table passed: passing moves the step and empties the pool
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(vamp) {
		t.Fatal("the Vampire card was not exiled")
	}
	bat := findBattlefieldByName(g, "Bat")
	if bat == uuid.Nil {
		t.Fatal("no Bat token")
	}
	if c, _ := battlefieldCard(g, bat); c.Power != 1 || !p1g2HasKeyword(g, bat, "flying") {
		t.Errorf("the Bat is %d/%d without flying", c.Power, c.Toughness)
	}
	// The Bat connects: it is sacrificed and the card returns tapped.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	declareAttack(t, g, opp.ID, bat)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bat) {
		t.Error("the Bat was not sacrificed")
	}
	back, ok := findCardByName(g, "Dead Vampire")
	if !ok || !back.Tapped {
		t.Fatalf("the Vampire did not return tapped (found=%v)", ok)
	}
}

func TestC2B10TimotharDeclinedOrTokenOrOtherPlayersVampireDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Timothar, Baron of Bats", "Legendary Creature — Vampire Noble", c2b10Timothar, 4, 4)
	token := b12Creature(g, me.ID, "Vampire", "Token Creature — Vampire", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Vampire", "Creature — Vampire", 2, 2)
	b15Destroy(t, g, token)
	b15Destroy(t, g, theirs)
	if latestPayUnlessOpen(g, me.ID) {
		t.Fatal("asked to pay for a token or an opponent's Vampire")
	}
	vamp := b12Creature(g, me.ID, "Mine", "Creature — Vampire", 2, 2)
	b15Destroy(t, g, vamp)
	b06AddMana(me, "B")
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(vamp) || findBattlefieldByName(g, "Bat") != uuid.Nil {
		t.Error("declining the payment still exiled the card or made a Bat")
	}
}

func findCardByName(g *game.Game, name string) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}

func latestPayUnlessOpen(g *game.Game, chooser uuid.UUID) bool {
	return payUnlessFor(g, chooser) != nil
}
