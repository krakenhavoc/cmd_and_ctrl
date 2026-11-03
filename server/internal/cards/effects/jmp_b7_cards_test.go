package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b7_cards_test.go — Jumpstart Azorius deck request, seventh batch
// (#4894). Each test plays the real catalog entry through the engine and
// asserts what the printed text says, and what it refuses.

const (
	jmpB7AerialAssaultOracle   = "a4eb06ca-16d8-41c2-8fc6-22e73872b5b0"
	jmpB7AlabasterMageOracle   = "ea7a8acb-f01f-4a69-9b3c-954d8f0d3155"
	jmpB7ArchonJusticeOracle   = "d551931d-4ebe-4acc-9e87-6c628897071a"
	jmpB7ArchonRedemptOracle   = "434e58f4-4237-454f-a2d0-61885c9aafaa"
	jmpB7GagglemasterOracle    = "8c601628-b1ed-467b-ba92-8572343147f8"
	jmpB7CuriousObsessionOracl = "22a23706-a5fe-46f3-b845-6f58d558c723"
	jmpB7DuelistsHeritageOracl = "5cdc4654-b8c5-4599-a1d2-9e9d5f6509cb"
	jmpB7GriffinAerieOracle    = "7cab62dc-a256-4da8-a3ea-b3903ec37077"
	jmpB7InniazOracle          = "b06ebf74-4042-4c8f-a70b-98ba9e74e6ed"
)

func TestJmpB7EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		jmpB7AerialAssaultOracle:   "Aerial Assault",
		jmpB7AlabasterMageOracle:   "Alabaster Mage",
		jmpB7ArchonJusticeOracle:   "Archon of Justice",
		jmpB7ArchonRedemptOracle:   "Archon of Redemption",
		jmpB7GagglemasterOracle:    "Aven Gagglemaster",
		jmpB7CuriousObsessionOracl: "Curious Obsession",
		jmpB7DuelistsHeritageOracl: "Duelist's Heritage",
		jmpB7GriffinAerieOracle:    "Griffin Aerie",
		jmpB7InniazOracle:          "Inniaz, the Gale Force",
	}
	for oracle, name := range want {
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

// --- shared setup ----------------------------------------------------------

// jmpB7Push puts a catalog creature on the battlefield, out of summoning
// sickness, with its printed size and keywords.
func jmpB7Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner, Keywords: keywords,
	})
}

func jmpB7Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Tapped = true })
}

func jmpB7Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func jmpB7GainLife(g *game.Game, who uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, who, n) })
}

func jmpB7Ref(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// --- Aerial Assault --------------------------------------------------------

func TestJmpB7AerialAssaultDestroysATappedCreatureAndGainsOnePerFlier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	victim := pushSizedCreature(g, opp.ID, "Tapped Bear", 2, 2)
	jmpB7Tap(g, victim)
	pushSizedCreature(g, me.ID, "My Flier A", 1, 1, "flying")
	pushSizedCreature(g, me.ID, "My Flier B", 1, 1, "flying")
	pushSizedCreature(g, me.ID, "My Walker", 1, 1)
	pushSizedCreature(g, opp.ID, "Their Flier", 1, 1, "flying")
	life := lifeOf(g, me.ID)

	castCatalogSpell(t, g, "Aerial Assault", "Sorcery", jmpB7AerialAssaultOracle, jmpB7Ref(victim))
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("the tapped creature was not destroyed")
	}
	if got := lifeOf(g, me.ID) - life; got != 2 {
		t.Errorf("gained %d life, want 2 (one for each creature you control with flying)", got)
	}
}

func TestJmpB7AerialAssaultRefusesAnUntappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := jmpB7Seats(g)
	bear := pushSizedCreature(g, opp.ID, "Untapped Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Aerial Assault", "Sorcery", jmpB7AerialAssaultOracle, jmpB7Ref(bear)); err == nil {
		t.Fatal("Aerial Assault targeted an untapped creature")
	}
}

// A creature untapped in response is no longer a legal target, and the
// spell does nothing at all, the life included (CR 608.2b).
func TestJmpB7AerialAssaultDoesNothingWhenTheTargetUntapsInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	victim := pushSizedCreature(g, opp.ID, "Bear", 2, 2)
	jmpB7Tap(g, victim)
	pushSizedCreature(g, me.ID, "My Flier", 1, 1, "flying")
	life := lifeOf(g, me.ID)

	castCatalogSpell(t, g, "Aerial Assault", "Sorcery", jmpB7AerialAssaultOracle, jmpB7Ref(victim))
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, victim).Tapped = false })
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(victim) {
		t.Error("an untapped creature was destroyed")
	}
	if lifeOf(g, me.ID) != life {
		t.Error("life was gained by a spell whose only target was illegal")
	}
}

// --- Alabaster Mage --------------------------------------------------------

func TestJmpB7AlabasterMageGivesYourCreatureLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := jmpB7Seats(g)
	mage := jmpB7Push(g, me.ID, "Alabaster Mage", "Creature — Human Wizard", jmpB7AlabasterMageOracle, 2, 1)
	bear := pushSizedCreature(g, me.ID, "Bear", 2, 2)
	advanceToMain(t, g)
	if err := hobB4Activate(t, g, me.ID, mage, 0, "{W}{C}", jmpB7Ref(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the target did not gain lifelink")
	}
	if hasEffectiveKeyword(t, g, mage, "lifelink") {
		t.Error("the Mage itself gained lifelink")
	}
}

func TestJmpB7AlabasterMageRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	mage := jmpB7Push(g, me.ID, "Alabaster Mage", "Creature — Human Wizard", jmpB7AlabasterMageOracle, 2, 1)
	theirs := pushSizedCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceToMain(t, g)
	if err := hobB4Activate(t, g, me.ID, mage, 0, "{W}{C}", jmpB7Ref(theirs)); err == nil {
		t.Fatal("the Mage targeted a creature its controller does not control")
	}
}

// --- Archon of Justice -----------------------------------------------------

func TestJmpB7ArchonOfJusticeExilesTargetPermanentWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	archon := jmpB7Push(g, me.ID, "Archon of Justice", "Creature — Archon", jmpB7ArchonJusticeOracle, 4, 4, "flying")
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Land", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID,
	})
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(archon) })
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: land}, true)

	if g.Battlefield.Contains(land) {
		t.Fatal("the target permanent is still on the battlefield")
	}
	if !g.Exile.Contains(land) {
		t.Error("the target permanent was not exiled")
	}
}

func TestJmpB7ArchonOfJusticeDoesNothingWhileItLives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	jmpB7Push(g, me.ID, "Archon of Justice", "Creature — Archon", jmpB7ArchonJusticeOracle, 4, 4, "flying")
	bear := pushSizedCreature(g, opp.ID, "Bear", 2, 2)
	advanceToMain(t, g)
	killCreature(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("another creature dying triggered the Archon")
	}
}

// --- Archon of Redemption --------------------------------------------------

func jmpB7RedemptionTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	jmpB7Push(g, me.ID, "Archon of Redemption", "Creature — Archon", jmpB7ArchonRedemptOracle, 3, 4, "flying")
	advanceToMain(t, g)
	return g, me, opp
}

// jmpB7Enter puts a creature onto the battlefield the way a resolution
// does, so its entry triggers.
func jmpB7Enter(t *testing.T, g *game.Game, owner *game.Player, name string, p, tough int, keywords ...string) uuid.UUID {
	t.Helper()
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test",
		Power: p, Toughness: tough, Owner: owner.ID, Controller: owner.ID, Keywords: keywords,
	}
	owner.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("MoveCardByID onto the battlefield: %v", err)
	}
	return c.InstanceID
}

func TestJmpB7ArchonOfRedemptionGainsAnotherFliersPower(t *testing.T) {
	g, me, _ := jmpB7RedemptionTable(t)
	life := lifeOf(g, me.ID)
	jmpB7Enter(t, g, me, "Big Flier", 5, 5, "flying")
	assistCardSettle(t, g, me.ID, nil, true)
	if got := lifeOf(g, me.ID) - life; got != 5 {
		t.Errorf("gained %d life, want the flier's power, 5", got)
	}
}

func TestJmpB7ArchonOfRedemptionGainsItsOwnPowerAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := jmpB7Seats(g)
	life := lifeOf(g, me.ID)
	castCatalogSpell(t, g, "Archon of Redemption", "Creature — Archon", jmpB7ArchonRedemptOracle, nil)
	assistCardSettle(t, g, me.ID, nil, true)
	// The cast card carries no printed size here (castCatalogSpell), so
	// its power is 0 and nothing is gained: the clause read the Archon's
	// power. Give it the printed 3/4 and enter it again to see the gain.
	if got := lifeOf(g, me.ID) - life; got != 0 {
		t.Fatalf("a 0-power Archon gained %d life", got)
	}
	life = lifeOf(g, me.ID)
	c := game.Card{
		InstanceID: uuid.New(), Name: "Archon of Redemption", TypeLine: "Creature — Archon",
		OracleID: jmpB7ArchonRedemptOracle, Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID,
	}
	me.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	// Two Archons now: the new one sees itself, the old one sees another
	// creature with flying. Both gain 3.
	assistCardSettle(t, g, me.ID, nil, true)
	assistCardSettle(t, g, me.ID, nil, true)
	if got := lifeOf(g, me.ID) - life; got != 6 {
		t.Errorf("gained %d life, want 3 from each Archon", got)
	}
}

func TestJmpB7ArchonOfRedemptionIgnoresGroundCreaturesAndOpponentsFliers(t *testing.T) {
	g, me, opp := jmpB7RedemptionTable(t)
	life := lifeOf(g, me.ID)
	jmpB7Enter(t, g, me, "Ground Bear", 4, 4)
	jmpB7Enter(t, g, opp, "Their Flier", 4, 4, "flying")
	assistCardSettle(t, g, me.ID, nil, true)
	if lifeOf(g, me.ID) != life {
		t.Error("a creature without flying, or an opponent's flier, triggered the Archon")
	}
}

func TestJmpB7ArchonOfRedemptionIsAMayAbility(t *testing.T) {
	g, me, _ := jmpB7RedemptionTable(t)
	life := lifeOf(g, me.ID)
	jmpB7Enter(t, g, me, "Big Flier", 5, 5, "flying")
	assistCardSettle(t, g, me.ID, nil, false)
	if lifeOf(g, me.ID) != life {
		t.Error("declining still gained life")
	}
}

// The power is read as the trigger resolves: a pump in response counts.
func TestJmpB7ArchonOfRedemptionReadsThePowerOnResolution(t *testing.T) {
	g, me, _ := jmpB7RedemptionTable(t)
	life := lifeOf(g, me.ID)
	flier := jmpB7Enter(t, g, me, "Flier", 2, 2, "flying")
	g.WithWriteLock(func() {
		_ = BoostUntilEOT{Target: flier, Power: 3, Label: "test pump"}.Apply(NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	assistCardSettle(t, g, me.ID, nil, true)
	if got := lifeOf(g, me.ID) - life; got != 5 {
		t.Errorf("gained %d life, want the pumped power 5", got)
	}
}

// --- Aven Gagglemaster -----------------------------------------------------

func TestJmpB7AvenGagglemasterGainsTwoForEachFlierIncludingItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	pushSizedCreature(g, me.ID, "Flier A", 1, 1, "flying")
	pushSizedCreature(g, me.ID, "Walker", 1, 1)
	pushSizedCreature(g, opp.ID, "Their Flier", 1, 1, "flying")
	life := lifeOf(g, me.ID)
	castCatalogSpell(t, g, "Aven Gagglemaster", "Creature — Bird Warrior", jmpB7GagglemasterOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID) - life; got != 4 {
		t.Errorf("gained %d life, want 4 (the Gagglemaster and one other flier, 2 each)", got)
	}
}

// --- Curious Obsession -----------------------------------------------------

func TestJmpB7CuriousObsessionPumpsAndDrawsOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	auraSeedLibrary(me, 3)
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Curious Obsession", jmpB7CuriousObsessionOracl, bear)
	if c, ok := battlefieldCard(g, bear); !ok || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Fatalf("the enchanted creature is not 3/3")
	}
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 3)
	assistCardSettle(t, g, me.ID, nil, true)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards on combat damage, want 1", got)
	}
}

func TestJmpB7CuriousObsessionDrawIsOptional(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	auraSeedLibrary(me, 3)
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Curious Obsession", jmpB7CuriousObsessionOracl, bear)
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 3)
	assistCardSettle(t, g, me.ID, nil, false)
	if me.Hand.Size() != hand {
		t.Error("declining still drew")
	}
}

func TestJmpB7CuriousObsessionIsSacrificedWithoutAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := jmpB7Seats(g)
	bear := auraBear(g, me.ID)
	aura := auraCast(t, g, "Curious Obsession", jmpB7CuriousObsessionOracl, bear)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura survived an end step with no attack")
	}
}

func TestJmpB7CuriousObsessionStaysAfterAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	bear := auraBear(g, me.ID)
	aura := auraCast(t, g, "Curious Obsession", jmpB7CuriousObsessionOracl, bear)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Error("the Aura was sacrificed although you attacked this turn")
	}
}

// --- Duelist's Heritage ----------------------------------------------------

func TestJmpB7DuelistsHeritageGivesAnAttackerDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	pushCatalogPermanent(g, me.ID, "Duelist's Heritage", "Enchantment", jmpB7DuelistsHeritageOracl, false)
	a := pushSizedCreature(g, me.ID, "Attacker A", 2, 2)
	b := pushSizedCreature(g, me.ID, "Attacker B", 2, 2)
	home := pushSizedCreature(g, me.ID, "Stays Home", 2, 2)
	declareAttack(t, g, opp.ID, a, b)

	rfPermAExpectLegal(t, g, me.ID, jmpB7Spec(t, jmpB7DuelistsHeritageOracl).Triggered[0].Targets,
		map[uuid.UUID]bool{a: true, b: true, home: false})
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: b}, true)
	if !hasEffectiveKeyword(t, g, b, "double strike") {
		t.Error("the chosen attacker lacks double strike")
	}
	if hasEffectiveKeyword(t, g, a, "double strike") {
		t.Error("the other attacker gained double strike")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("two attackers in one declaration triggered twice")
	}
}

// "Whenever one or more creatures attack": an opponent's attack too, and
// the target may be their attacker.
func TestJmpB7DuelistsHeritageTriggersOnAnOpponentsAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	pushCatalogPermanent(g, opp.ID, "Duelist's Heritage", "Enchantment", jmpB7DuelistsHeritageOracl, false)
	a := pushSizedCreature(g, me.ID, "Attacker", 2, 2)
	declareAttack(t, g, opp.ID, a)
	assistCardSettle(t, g, opp.ID, &game.TargetRef{Kind: game.TargetCard, ID: a}, false)
	if hasEffectiveKeyword(t, g, a, "double strike") {
		t.Error("declining still granted double strike")
	}
}

func jmpB7Spec(t *testing.T, oracle string) Spec {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", oracle)
	}
	return spec
}

// --- Griffin Aerie ---------------------------------------------------------

func TestJmpB7GriffinAerieMakesAGriffinAfterThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := jmpB7Seats(g)
	pushCatalogPermanent(g, me.ID, "Griffin Aerie", "Enchantment", jmpB7GriffinAerieOracle, false)
	advanceToMain(t, g)
	jmpB7GainLife(g, me.ID, 2)
	jmpB7GainLife(g, me.ID, 1)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	var griffins []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Griffin" && c.Controller == me.ID {
			griffins = append(griffins, c)
		}
	}
	if len(griffins) != 1 {
		t.Fatalf("%d Griffin tokens, want 1", len(griffins))
	}
	gr := griffins[0]
	if gr.CurrentPower() != 2 || gr.CurrentToughness() != 2 || !slices.Contains(gr.Colors, "W") {
		t.Errorf("the Griffin is %d/%d %v, want a 2/2 white", gr.CurrentPower(), gr.CurrentToughness(), gr.Colors)
	}
	if !hasEffectiveKeyword(t, g, gr.InstanceID, "flying") {
		t.Error("the Griffin lacks flying")
	}
}

func TestJmpB7GriffinAerieNeedsThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := jmpB7Seats(g)
	pushCatalogPermanent(g, me.ID, "Griffin Aerie", "Enchantment", jmpB7GriffinAerieOracle, false)
	advanceToMain(t, g)
	jmpB7GainLife(g, me.ID, 2)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Griffin", me.ID); n != 0 {
		t.Errorf("%d Griffins after gaining only 2 life", n)
	}
}

// --- Inniaz, the Gale Force ------------------------------------------------

func jmpB7Inniaz(g *game.Game, owner uuid.UUID) uuid.UUID {
	return jmpB7Push(g, owner, "Inniaz, the Gale Force", "Legendary Creature — Djinn", jmpB7InniazOracle, 4, 4, "flying")
}

func TestJmpB7InniazPumpsOnlyAttackingFliers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	inniaz := jmpB7Inniaz(g, me.ID)
	flier := pushSizedCreature(g, me.ID, "Flier", 1, 1, "flying")
	walker := pushSizedCreature(g, me.ID, "Walker", 1, 1)
	idle := pushSizedCreature(g, me.ID, "Idle Flier", 1, 1, "flying")
	declareAttack(t, g, opp.ID, flier, walker)
	passPriorityAroundTable(t, g)
	if err := hobB4Activate(t, g, me.ID, inniaz, 0, "{W}{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range []struct {
		id   uuid.UUID
		want int
	}{{flier, 2}, {walker, 1}, {idle, 1}} {
		if got, ok := battlefieldCard(g, c.id); !ok || got.CurrentPower() != c.want {
			t.Errorf("%s has power %d, want %d", got.Name, got.CurrentPower(), c.want)
		}
	}
}

// Each player gains a permanent from the player to their right: in turn
// order A, B, C, D, A gains D's, B gains A's, C gains B's and D gains C's.
func TestJmpB7InniazPassesAPermanentFromEachPlayersRight(t *testing.T) {
	g := newCatalogGame(t)
	n := len(g.Seats)
	seat := func(k int) *game.Player { return g.Seats[(g.Turn.ActiveSeat+k)%n] }
	me := seat(0)
	inniaz := jmpB7Inniaz(g, me.ID)
	f1 := pushSizedCreature(g, me.ID, "Flier 1", 1, 1, "flying")
	f2 := pushSizedCreature(g, me.ID, "Flier 2", 1, 1, "flying")
	mine := pushSizedCreature(g, me.ID, "My Bear", 2, 2)
	owned := map[int]uuid.UUID{0: mine}
	for k := 1; k < n; k++ {
		owned[k] = pushSizedCreature(g, seat(k).ID, "Bear", 2, 2)
		// A land is never a choice.
		pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Land", TypeLine: "Land",
			Owner: seat(k).ID, Controller: seat(k).ID})
	}
	declareAttack(t, g, seat(1).ID, inniaz, f1, f2)
	passPriorityAroundTable(t, g)

	// The controller is asked once per player, in turn order from
	// themselves, each time among the right-hand neighbour's nonland
	// permanents.
	for k := 0; k < n; k++ {
		right := (k + n - 1) % n
		p := chooseCardsChoiceFor(g, me.ID)
		if p == nil {
			t.Fatalf("no choice open for seat %d", k)
		}
		for _, id := range p.ChooseCards {
			c, _ := battlefieldCard(g, id)
			if c.Controller != seat(right).ID || c.IsLand() {
				t.Fatalf("seat %d's choice offered %s, controlled by someone else or a land", k, c.Name)
			}
		}
		answerChooseCards(t, g, me.ID, owned[right])
		if k < n-1 {
			// Nothing changes hands until every choice is made.
			if controllerOf(t, g, owned[right]) != seat(right).ID {
				t.Fatal("control changed before the last choice")
			}
		}
	}
	passPriorityAroundTable(t, g)
	for k := 0; k < n; k++ {
		right := (k + n - 1) % n
		if got := controllerOf(t, g, owned[right]); got != seat(k).ID {
			t.Errorf("seat %d's bear is controlled by %v, want seat %d", right, got, k)
		}
	}
}

func TestJmpB7InniazNeedsThreeAttackingFliers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	inniaz := jmpB7Inniaz(g, me.ID)
	f1 := pushSizedCreature(g, me.ID, "Flier", 1, 1, "flying")
	walker := pushSizedCreature(g, me.ID, "Walker", 1, 1)
	pushSizedCreature(g, opp.ID, "Their Bear", 2, 2)
	declareAttack(t, g, opp.ID, inniaz, f1, walker)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("two fliers and a walker triggered Inniaz")
	}
}

// With two players left it is an exchange: each gains one of the other's.
func TestJmpB7InniazIsAnExchangeWithTwoPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	for _, p := range g.Seats {
		if p.ID != me.ID && p.ID != opp.ID {
			p.Eliminated = true
		}
	}
	inniaz := jmpB7Inniaz(g, me.ID)
	f1 := pushSizedCreature(g, me.ID, "Flier 1", 1, 1, "flying")
	f2 := pushSizedCreature(g, me.ID, "Flier 2", 1, 1, "flying")
	mine := pushSizedCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushSizedCreature(g, opp.ID, "Their Bear", 2, 2)
	declareAttack(t, g, opp.ID, inniaz, f1, f2)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, theirs)
	answerChooseCards(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if controllerOf(t, g, theirs) != me.ID || controllerOf(t, g, mine) != opp.ID {
		t.Error("the two bears were not exchanged")
	}
}

// A player whose right-hand neighbour controls no nonland permanent gains
// nothing, and the controller is not asked for them.
func TestJmpB7InniazSkipsAPlayerWhoseNeighbourHasNothing(t *testing.T) {
	g := newCatalogGame(t)
	n := len(g.Seats)
	seat := func(k int) *game.Player { return g.Seats[(g.Turn.ActiveSeat+k)%n] }
	me := seat(0)
	inniaz := jmpB7Inniaz(g, me.ID)
	f1 := pushSizedCreature(g, me.ID, "Flier 1", 1, 1, "flying")
	f2 := pushSizedCreature(g, me.ID, "Flier 2", 1, 1, "flying")
	theirs := pushSizedCreature(g, seat(1).ID, "Seat 1 Bear", 2, 2)
	declareAttack(t, g, seat(2).ID, inniaz, f1, f2)
	passPriorityAroundTable(t, g)

	// Only seat 0 (Inniaz's controller) and seat 1 have anything; the
	// first question is seat 1's (from seat 0) and offers my permanents;
	// the second is seat 2's, from seat 1.
	p := chooseCardsChoiceFor(g, me.ID)
	if p == nil {
		t.Fatal("no choice opened")
	}
	if !slices.Contains(p.ChooseCards, f1) {
		t.Fatal("the first choice is not among my permanents (seat 0 has nothing to its right)")
	}
	answerChooseCards(t, g, me.ID, f1)
	answerChooseCards(t, g, me.ID, theirs)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("asked for a player whose neighbour has nothing")
	}
	passPriorityAroundTable(t, g)
	if controllerOf(t, g, f1) != seat(1).ID {
		t.Error("seat 1 did not gain the flier from seat 0")
	}
	if controllerOf(t, g, theirs) != seat(2).ID {
		t.Error("seat 2 did not gain seat 1's bear")
	}
}
