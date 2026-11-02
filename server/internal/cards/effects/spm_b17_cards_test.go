package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_b17_cards_test.go — Marvel's Spider-Man, seventeenth batch (#4460).
// Each test plays the real catalog entry and asserts what its text says.

const (
	spmB17WindsOracle     = "f525cf10-e24c-4c46-9a13-6f8579d09d50"
	spmB17WolverineOracle = "6e53a43e-5a23-4057-9bb5-281508da57af"
	spmB17AgentOracle     = "14c356aa-178d-4c14-a88e-fbe92ee6d89f"
	spmB17ArachneOracle   = "28d500ea-720f-4332-891e-404fb03acb5c"
	spmB17AranaOracle     = "a16037f2-a0f6-465d-8d6e-cec34b4a6320"
	spmB17TentaclesOracle = "a5a18764-eb29-4bd8-a2e9-0ce55f4f6c88"
	spmB17ElectroOracle   = "3db16fb8-5650-4568-b975-42ad3d44ebcd"
	spmB17GwenomOracle    = "c48cf135-b280-4e9f-a9d9-225104870741"
	spmB17KravenOracle    = "e74efcf2-d77f-4fb5-9027-77e462ec829e"
)

func TestSpmB17EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spmB17WindsOracle:         "Winds of Change",
		spmB17WolverineOracle:     "Wolverine, Best There Is",
		spmB17AgentOracle:         "Agent Venom",
		spmB17ArachneOracle:       "Arachne, Psionic Weaver",
		spmB17AranaOracle:         "Araña, Heart of the Spider",
		spmB17TentaclesOracle:     "Doc Ock's Tentacles",
		spmB17ElectroOracle:       "Electro, Assaulting Battery",
		spmB17GwenomOracle:        "Gwenom, Remorseless",
		spmB17KravenOracle:        "Kraven's Last Hunt",
		milesMoralesOracle:        "Miles Morales",
		milesMoralesOracle + "#1": "Ultimate Spider-Man",
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

// spmB17Creature pushes a catalog creature with a printed body.
func spmB17Creature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// spmB17Damage has `source` deal `n` damage to the creature `victim`, then
// lets state-based actions settle.
func spmB17Damage(t *testing.T, g *game.Game, source, victim uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(source, victim, n); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
}

// spmB17Top puts a card on top of p's library, keeping the rest.
func spmB17Top(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	return c.InstanceID
}

func spmB17Marked(g *game.Game, id uuid.UUID) int {
	if c := battlefieldCardFor(g, id); c != nil {
		return c.DamageMarked
	}
	return -1
}

// --- Winds of Change ---

func TestSpmB17WindsOfChangeEachPlayerRedrawsAsManyAsTheyShuffledIn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	hands := map[uuid.UUID]int{}
	libs := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		hands[p.ID] = p.Hand.Size()
		libs[p.ID] = p.Library.Size()
	}
	opp := g.Seats[1]
	// The second seat throws one card away first, so the counts differ.
	g.WithWriteLock(func() {
		c := opp.Hand.Cards[0]
		opp.Hand.Remove(c.InstanceID)
		opp.Graveyard.PushTop(c)
	})
	hands[opp.ID]--
	draws := countDrawEvents(g)

	castCatalogSpell(t, g, "Winds of Change", "Sorcery", spmB17WindsOracle, nil)
	passPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		if p.Hand.Size() != hands[p.ID] {
			t.Errorf("%s hand %d, want %d (drew as many as were shuffled in)", p.Name, p.Hand.Size(), hands[p.ID])
		}
		if p.Library.Size() != libs[p.ID] {
			t.Errorf("%s library %d, want %d", p.Name, p.Library.Size(), libs[p.ID])
		}
	}
	total := 0
	for _, n := range hands {
		total += n
	}
	if got := countDrawEvents(g) - draws; got != total {
		t.Errorf("%d cards drawn, want %d", got, total)
	}
}

// countDrawEvents counts the draws in the log.
func countDrawEvents(g *game.Game) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventDrawCard {
			n++
		}
	}
	return n
}

// --- Wolverine, Best There Is ---

func TestSpmB17WolverineDoublesHisDamageAndGrowsAfterHittingACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	wolverine := spmB17Creature(g, me.ID, "Wolverine, Best There Is", "Legendary Creature — Mutant Berserker Hero", spmB17WolverineOracle, 2, 2)
	victim := pushVanillaCreature(g, opp.ID, "Big Victim", 6, 6)

	spmB17Damage(t, g, wolverine, victim, 2)
	if got := spmB17Marked(g, victim); got != 4 {
		t.Fatalf("damage marked %d, want 4 (doubled)", got)
	}
	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(wolverine, opp.ID, 3); err != nil {
			t.Fatalf("damage to player: %v", err)
		}
	})
	if opp.Life != life-6 {
		t.Errorf("life %d -> %d, want -6 (doubled)", life, opp.Life)
	}

	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, wolverine, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
}

func TestSpmB17WolverineDoesNotGrowWithoutDamagingAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	wolverine := spmB17Creature(g, me.ID, "Wolverine, Best There Is", "Legendary Creature — Mutant Berserker Hero", spmB17WolverineOracle, 2, 2)
	other := pushVanillaCreature(g, me.ID, "Other Source", 2, 2)
	victim := pushVanillaCreature(g, opp.ID, "Victim", 6, 6)

	// Another source damaging a creature, Wolverine damaging a player and
	// himself: none of it is "Wolverine dealt damage to another creature".
	spmB17Damage(t, g, other, victim, 2)
	if got := spmB17Marked(g, victim); got != 2 {
		t.Fatalf("another source's damage was changed: %d, want 2", got)
	}
	spmB17Damage(t, g, wolverine, wolverine, 0)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(wolverine, opp.ID, 1) })

	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, wolverine, game.CounterPlusOne); got != 0 {
		t.Errorf("+1/+1 counters %d, want 0", got)
	}
}

func TestSpmB17WolverineRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	wolverine := spmB17Creature(g, me.ID, "Wolverine, Best There Is", "Legendary Creature — Mutant Berserker Hero", spmB17WolverineOracle, 2, 2)
	floatMana(t, g, me, "{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, wolverine, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wolverine) })
	c := battlefieldCardFor(g, wolverine)
	if c == nil {
		t.Fatal("Wolverine was destroyed through a regeneration shield")
	}
	if !c.Tapped {
		t.Error("regeneration did not tap him")
	}
}

// --- Agent Venom ---

func TestSpmB17AgentVenomDrawsAndLosesForAnotherNontokenCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	agent := spmB17Creature(g, me.ID, "Agent Venom", "Legendary Creature — Symbiote Soldier Hero", spmB17AgentOracle, 2, 3)
	if !hasKeywordOnBattlefield(t, g, agent, "menace") || !hasKeywordOnBattlefield(t, g, agent, "flash") {
		t.Error("Agent Venom lacks flash or menace")
	}
	bear := seedBear(g, me.ID)
	hand, life := me.Hand.Size(), me.Life
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d -> %d, life %d -> %d; want +1 card and -1 life", hand, me.Hand.Size(), life, me.Life)
	}

	// A token, an opponent's creature and Agent Venom himself do nothing.
	token := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Token Bear",
		TypeLine: "Token Creature — Bear", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	theirs := seedBear(g, g.Seats[1].ID)
	hand, life = me.Hand.Size(), me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	killCreature(t, g, g.Seats[1].ID, theirs)
	killCreature(t, g, me.ID, agent)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("hand %d -> %d, life %d -> %d; want no change", hand, me.Hand.Size(), life, me.Life)
	}
}

// --- Arachne, Psionic Weaver ---

func TestSpmB17ArachneLooksAtAnOpponentsHandAndTaxesTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	id := castCatalogSpell(t, g, "Arachne, Psionic Weaver", "Legendary Creature — Spider Human Hero", spmB17ArachneOracle, nil)
	passPriorityAroundTable(t, g)

	pick := latestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no opponent prompt")
	}
	if len(pick.PickOptions) != 3 {
		t.Errorf("%d players offered, want the three opponents", len(pick.PickOptions))
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, optionFor(t, pick, opp.ID)); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	for i := range opp.Hand.Cards {
		if !opp.Hand.Cards[i].IsKnownTo(me.ID) {
			t.Errorf("the chosen opponent's card %d is not known to Arachne's controller", i)
		}
	}
	for i := range other.Hand.Cards {
		if other.Hand.Cards[i].IsKnownTo(me.ID) {
			t.Errorf("another opponent's card %d was revealed too", i)
		}
	}
	types := pickOptionWords(t, g, me.ID)
	for _, w := range types {
		if w == "creature" {
			t.Error("creature is offered")
		}
	}
	answerAnchorWord(t, g, me.ID, "instant")
	if got := g.ChosenOptionOf(id); got != "instant" {
		t.Fatalf("chosen %q, want instant", got)
	}
	if got := priceInHand(t, g, opp, "Bolt", "Instant", "{R}"); got != 2 {
		t.Errorf("an opponent's instant costs %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "My Bolt", "Instant", "{R}"); got != 2 {
		t.Errorf("my own instant costs %d, want 2", got)
	}
	if got := priceInHand(t, g, opp, "Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("a creature spell costs %d, want 2", got)
	}
}

// pickOptionWords lists the open option prompt's words for `chooser`.
func pickOptionWords(t *testing.T, g *game.Game, chooser uuid.UUID) []string {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatal("no option prompt open")
	}
	return optionWords(c)
}

func TestSpmB17ArachneWebSlingsForWhiteWithATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	arachne := handCardFull(me, "Arachne, Psionic Weaver", "Legendary Creature — Spider Human Hero", "{2}{W}", spmB17ArachneOracle, []string{"W"})
	untapped := pushVanillaCreature(g, me.ID, "Untapped Bear", 2, 2)
	tapped := pushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	floatMana(t, g, me, "{W}")
	if err := g.CastSpell(me.ID, arachne, game.CastSpellParams{
		Strict: true, AlternativeCost: "web_slinging", AltCostIDs: []uuid.UUID{untapped},
	}); err == nil {
		t.Fatal("web-slinging accepted an untapped creature")
	}
	if err := g.CastSpell(me.ID, arachne, game.CastSpellParams{
		Strict: true, AlternativeCost: "web_slinging", AltCostIDs: []uuid.UUID{tapped},
	}); err != nil {
		t.Fatalf("web-slinging cast: %v", err)
	}
	if !me.Hand.Contains(tapped) {
		t.Error("the tapped creature did not return to hand")
	}
}

// --- Araña, Heart of the Spider ---

func TestSpmB17AranaCountersAnAttackerAndAModifiedHitImpulses(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	arana := spmB17Creature(g, me.ID, "Araña, Heart of the Spider", "Legendary Creature — Spider Human Hero", spmB17AranaOracle, 3, 3)
	bear := seedBear(g, me.ID)
	stackLibrary(me, "Buried", "Instant", "{R}")
	top := stackLibrary(me, "Top Card", "Instant", "{R}")

	declareAttack(t, g, opp.ID, arana, bear)
	mscM2PickTargets(t, g, bear)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("bear counters %d, want 1", got)
	}
	if got := countersOn(g, arana, game.CounterPlusOne); got != 0 {
		t.Fatalf("Araña got a counter too: %d", got)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the modified Bear's hit exiled nothing")
	}
	if !permissionLive(g, exiledPermission(g, top), me.ID) {
		t.Error("the exiled card is not playable")
	}
	exiled := 0
	for _, c := range g.Exile.Cards {
		if c.Owner == me.ID {
			exiled++
		}
	}
	if exiled != 1 {
		t.Errorf("%d cards exiled, want 1 (Araña is not modified)", exiled)
	}
}

// --- Doc Ock's Tentacles ---

func TestSpmB17DocOcksTentaclesAttachToABigArrivalOnly(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	tentacles := seedEquipment(g, me.ID, "Doc Ock's Tentacles", spmB17TentaclesOracle)

	small := msc12Cast(t, g, "Small One", "Creature — Bear", "{4}", "{C}{C}{C}{C}", "", 2)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, tentacles); host.ID == small {
		t.Fatal("attached to a mana value 4 creature")
	}
	big := msc12Cast(t, g, "Big One", "Creature — Giant", "{5}", "{C}{C}{C}{C}{C}", "", 2)
	c := awaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(c.ID, me.ID, true); err != nil {
		t.Fatalf("accept the attach: %v", err)
	}
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, tentacles); host.ID != big {
		t.Fatalf("the Tentacles are on %v, want the mana value 5 creature", host)
	}
	if p, tough := effectivePTOf(t, g, big); p != 6 || tough != 6 {
		t.Errorf("equipped P/T %d/%d, want 6/6", p, tough)
	}
}

// --- Electro, Assaulting Battery ---

func TestSpmB17ElectroAddsRedOnInstantsAndBurnsAPlayerAsHeLeaves(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	electro := spmB17Creature(g, me.ID, "Electro, Assaulting Battery", "Legendary Creature — Human Villain", spmB17ElectroOracle, 2, 3)

	msc12CastAndSettle(t, g, "Creature Spell", "Creature — Bear")
	if me.ManaPool.Total() != 0 {
		t.Fatalf("a creature spell added mana: %d", me.ManaPool.Total())
	}
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Cheap Instant", TypeLine: "Instant", ManaCost: "{1}",
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := me.ManaPool.Total(); got != 1 {
		t.Fatalf("pool %d after the cast trigger, want {R}", got)
	}

	fillPool(me, 2)
	killCreature(t, g, me.ID, electro)
	passPriorityAroundTable(t, g)
	answerNumber(t, g, me.ID, 3)
	answerPayUnless(t, g, me.ID, true)
	msc12PickPlayer(t, g, opp.ID)
	life := opp.Life
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, opp.Life)
	}
}

// --- Gwenom, Remorseless ---

func TestSpmB17GwenomOpensTheTopOfTheLibraryForLifeAfterAttacking(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	gwenom := spmB17Creature(g, me.ID, "Gwenom, Remorseless", "Legendary Creature — Symbiote Spider Hero", spmB17GwenomOracle, 4, 4)
	top := spmB17Top(me, game.Card{Name: "Top Instant", TypeLine: "Instant", ManaCost: "{1}{B}"})
	if perm := grantedPermissionOn(g, me.ID, top, game.ZoneLibrary); perm.Granted() {
		t.Fatal("the top is open before Gwenom attacks")
	}

	declareAttack(t, g, opp.ID, gwenom)
	passPriorityAroundTable(t, g)
	if perm := grantedPermissionOn(g, me.ID, top, game.ZoneLibrary); !perm.Granted() {
		t.Fatal("the attack opened nothing")
	}
	visible := false
	g.ReadSnapshot(func() { visible = g.LibraryTopVisibleToLocked(me.ID, me.ID) })
	if !visible {
		t.Error("the top card is not visible to its owner")
	}
	// "Rather than pay its mana cost": the printed price is not a way in.
	floatMana(t, g, me, "{C}{B}")
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("the top card was cast for its mana cost")
	}
	life := me.Life
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library", AlternativeCost: "gwenom_remorseless"}); err != nil {
		t.Fatalf("cast from the top for life: %v", err)
	}
	if got := life - me.Life; got != 2 {
		t.Errorf("life paid %d, want 2", got)
	}
	passPriorityAroundTable(t, g)

	nextTurnOfMine(t, g)
	next := spmB17Top(me, game.Card{Name: "Later Instant", TypeLine: "Instant", ManaCost: "{B}"})
	if perm := grantedPermissionOn(g, me.ID, next, game.ZoneLibrary); perm.Granted() {
		t.Error("the permission outlived the turn")
	}
}

// --- Kraven's Last Hunt ---

func TestSpmB17KravensLastHuntMillsThenHitsForTheBiggestGraveyardCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 4; i++ {
		stackLibrary(me, "Filler", "Sorcery", "")
	}
	big := spmB17Top(me, game.Card{Name: "Milled Brute", TypeLine: "Creature — Beast", Power: 5, Toughness: 5})
	victim := pushVanillaCreature(g, opp.ID, "Victim", 6, 6)

	saga := castCatalogSpell(t, g, "Kraven's Last Hunt", "Enchantment — Saga", spmB17KravenOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(big) {
		t.Fatal("chapter I did not mill the creature")
	}
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if got := spmB17Marked(g, victim); got != 5 {
		t.Errorf("damage %d, want 5 (the milled 5/5's power)", got)
	}
	if !g.Battlefield.Contains(saga) {
		t.Fatal("the Saga left early")
	}
}

func TestSpmB17KravensLastHuntChaptersTwoAndThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	saga := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Kraven's Last Hunt",
		TypeLine: "Enchantment — Saga", OracleID: spmB17KravenOracle, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLore: 1}})
	bear := seedBear(g, me.ID)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})

	advanceToPrecombatMainOf(t, g, 0)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("chapter II: bear %d/%d, want 4/4", p, tough)
	}

	advanceToPrecombatMainOf(t, g, 0)
	answerPickTarget(t, g, dead)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("chapter III did not return the creature card")
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after III")
	}
}

// --- Miles Morales // Ultimate Spider-Man ---

func spmB17MilesCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: milesMoralesOracle, Layout: game.LayoutModalDFC,
		Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Miles Morales", TypeLine: "Legendary Creature — Human Citizen Hero", ManaCost: "{1}{G}",
				Colors: []string{"G"}, Power: 1, Toughness: 2},
			{Name: "Ultimate Spider-Man", TypeLine: "Legendary Creature — Spider Human Hero", ManaCost: "{3}{R}{G}{W}",
				Colors: []string{"G", "R", "W"}, Power: 4, Toughness: 3},
		},
	}
	c.SetFace(0)
	return c
}

func TestSpmB17MilesMoralesCountersTwoThenTransforms(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	a, b := seedBear(g, me.ID), seedBear(g, g.Seats[1].ID)
	miles := spmB17MilesCard(me.ID)
	me.Hand.PushTop(miles)
	if err := g.CastSpell(me.ID, miles.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Miles: %v", err)
	}
	passPriorityAroundTable(t, g)
	mscM2PickTargets(t, g, a, b)
	if countersOn(g, a, game.CounterPlusOne) != 1 || countersOn(g, b, game.CounterPlusOne) != 1 {
		t.Fatalf("counters %d and %d, want 1 each", countersOn(g, a, game.CounterPlusOne), countersOn(g, b, game.CounterPlusOne))
	}

	floatMana(t, g, me, "{C}{C}{C}{R}{G}{W}")
	if err := g.ActivateCatalogAbility(me.ID, miles.InstanceID, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("transform: %v", err)
	}
	passPriorityAroundTable(t, g)
	card := battlefieldCardFor(g, miles.InstanceID)
	if card == nil || card.Name != "Ultimate Spider-Man" {
		t.Fatalf("after the transform: %+v", card)
	}
	if !hasKeywordOnBattlefield(t, g, miles.InstanceID, "first strike") || !hasKeywordOnBattlefield(t, g, miles.InstanceID, "haste") {
		t.Error("Ultimate Spider-Man lacks first strike or haste")
	}
}

func TestSpmB17UltimateSpiderManCamouflageAndCounterDoubling(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	c := spmB17MilesCard(me.ID)
	c.SetFace(1)
	spidey := pushBattlefieldCardWithTimestamp(g, c)

	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, spidey, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("camouflage: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, spidey, game.CounterPlusOne); got != 1 {
		t.Errorf("camouflage counters %d, want 1", got)
	}
	if !hasKeywordOnBattlefield(t, g, spidey, "hexproof") {
		t.Error("camouflage gave no hexproof")
	}
	if cols := effectiveColors(t, g, spidey); len(cols) != 0 {
		t.Errorf("colors %v, want colorless", cols)
	}

	spider := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spider Friend",
		TypeLine: "Creature — Spider", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 2, "shield": 1}})
	legend := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Legend",
		TypeLine: "Legendary Creature — Human", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 3}})
	plain := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Plain Bear",
		TypeLine: testCreatureTypeLine, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1}})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Spider",
		TypeLine: "Creature — Spider", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterPlusOne: 1}})

	declareAttack(t, g, opp.ID, spidey)
	passPriorityAroundTable(t, g)
	checks := []struct {
		id   uuid.UUID
		kind string
		want int
	}{
		{spidey, game.CounterPlusOne, 2},
		{spider, game.CounterPlusOne, 4},
		{spider, "shield", 2},
		{legend, game.CounterPlusOne, 6},
		{plain, game.CounterPlusOne, 1},
		{theirs, game.CounterPlusOne, 1},
	}
	for _, ch := range checks {
		if got := countersOn(g, ch.id, ch.kind); got != ch.want {
			t.Errorf("%s %s counters %d, want %d", cardNameOf(g, ch.id), ch.kind, got, ch.want)
		}
	}
}
