package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b4_cards_test.go — Jumpstart Selesnya deck request, fourth batch
// (#4893). Each test plays the real catalog entry through the engine and
// asserts what the printed text says, and what it refuses.

const (
	jmpB4AngelOfMercyOracle = "a2daaf32-dbfe-4618-892e-0da24f63a44a"
	jmpB4DireHourOracle     = "8b1e6b7c-892d-4058-b183-9121cdd623f0"
	jmpB4ArbiterOracle      = "34dd789c-4d80-4ccc-a4f7-2e5388a9689d"
	jmpB4BasriKetOracle     = "497d7ab8-07f1-4c8a-a45a-52b9a4d9400f"
	jmpB4AcolyteOracle      = "92c824da-2351-43bd-9fa0-d09709ad8205"
	jmpB4SolidarityOracle   = "68281e06-64e3-490b-8691-8f4169614ec0"
	jmpB4PromotionOracle    = "303d74d0-84db-4694-bb1a-470cf470bc4f"
	jmpB4SanctuaryOracle    = "1b70d7fa-ba2c-4945-b8fd-23623b88c036"
	jmpB4BrightmareOracle   = "32025108-0dba-40bd-88db-5fcc48e3d054"
	jmpB4MantleOracle       = "7417bc73-a75a-4c3a-88d2-17b43e225c2c"
)

func TestJmpB4EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		jmpB4AngelOfMercyOracle: "Angel of Mercy",
		jmpB4DireHourOracle:     "Angel of the Dire Hour",
		jmpB4ArbiterOracle:      "Angelic Arbiter",
		jmpB4BasriKetOracle:     "Basri Ket",
		jmpB4AcolyteOracle:      "Basri's Acolyte",
		jmpB4SolidarityOracle:   "Basri's Solidarity",
		jmpB4PromotionOracle:    "Battlefield Promotion",
		jmpB4SanctuaryOracle:    "Blessed Sanctuary",
		jmpB4BrightmareOracle:   "Brightmare",
		jmpB4MantleOracle:       "Celestial Mantle",
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

// --- shared --------------------------------------------------------------

// jmpB4CastCreature puts a catalog creature card in p's hand and casts it
// with the given params. The caller is at a point where p may cast it.
func jmpB4CastCreature(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// jmpB4Live is the permanent with its layers applied.
func jmpB4Live(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardForTest(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c
}

func jmpB4PlusOneCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return jmpB4Live(t, g, id).Counters[game.CounterPlusOne]
}

// jmpB4Pick answers the latest pick_target prompt of chooser with the cards
// given (none is a legal answer to an "up to" clause).
func jmpB4Pick(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) error {
	t.Helper()
	pick := latestPickTarget(g, chooser)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	var refs []game.TargetRef
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return g.ResolvePickTargets(pick.ID, chooser, refs)
}

// jmpB4CastSandboxSpell has p cast a non-catalog sorcery: a spell, nothing
// more.
func jmpB4CastSandboxSpell(t *testing.T, g *game.Game, p *game.Player) error {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: "Plain Sorcery", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	return g.CastSpell(p.ID, id, game.CastSpellParams{})
}

// --- Angel of Mercy --------------------------------------------------------

func TestJmpB4AngelOfMercyGainsThreeLifeOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	life := me.Life
	angel := jmpB4CastCreature(t, g, me, "Angel of Mercy", "Creature — Angel", jmpB4AngelOfMercyOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(angel) {
		t.Fatal("the Angel did not enter")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if !game.HasKeyword(jmpB4Live(t, g, angel), "flying") {
		t.Error("the Angel has flying")
	}
}

// --- Angel of the Dire Hour ------------------------------------------------

func TestJmpB4DireHourCastFromHandExilesEveryAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := b12Creature(g, me.ID, "Raider", "Creature — Bear", 2, 2)
	homebody := b12Creature(g, me.ID, "Homebody", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, raider)
	if findBattlefieldCardForTest(g, raider).AttackingTarget != opp.ID {
		t.Fatal("the raider is not attacking")
	}
	angel := jmpB4CastCreature(t, g, me, "Angel of the Dire Hour", "Creature — Angel", jmpB4DireHourOracle, 5, 4)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(angel) {
		t.Fatal("the Angel did not enter (flash)")
	}
	if g.Battlefield.Contains(raider) || !g.Exile.Contains(raider) {
		t.Error("the attacking creature is exiled")
	}
	if !g.Battlefield.Contains(homebody) || !g.Battlefield.Contains(theirs) {
		t.Error("creatures that are not attacking stay")
	}
}

func TestJmpB4DireHourNotCastFromHandExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := b12Creature(g, me.ID, "Raider", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, raider)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Angel of the Dire Hour", TypeLine: "Creature — Angel",
		OracleID: jmpB4DireHourOracle, ManaCost: "{5}{W}{W}", Power: 5, Toughness: 4, Owner: me.ID, Controller: me.ID})
	b30Reanimate(t, g, dead, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the reanimated Angel is not on the battlefield")
	}
	if !g.Battlefield.Contains(raider) {
		t.Error("an Angel that was not cast from hand exiles nothing")
	}
}

// --- Angelic Arbiter -------------------------------------------------------

func TestJmpB4ArbiterAnOpponentWhoCastASpellCantAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Angelic Arbiter", "Creature — Angel", jmpB4ArbiterOracle, false)
	raider := b12Creature(g, opp.ID, "Raider", "Creature — Bear", 2, 2)
	other := b12Creature(g, third.ID, "Other Raider", "Creature — Bear", 2, 2)

	// The opponent casts a spell, then may not attack.
	aangAdvanceToMain(t, g, 1)
	if err := jmpB4CastSandboxSpell(t, g, opp); err != nil {
		t.Fatalf("the opponent casts before attacking: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, me.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Fatalf("an opponent who cast a spell attacked: %v", err)
	}

	// An opponent who cast nothing attacks, and then may not cast.
	aangAdvanceToMain(t, g, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(other, me.ID); err != nil {
		t.Fatalf("an opponent who cast nothing may attack: %v", err)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if err := jmpB4CastSandboxSpell(t, g, third); err == nil {
		t.Fatal("an opponent who attacked with a creature cast a spell")
	}
}

func TestJmpB4ArbiterDoesNotBindItsController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Angelic Arbiter", "Creature — Angel", jmpB4ArbiterOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	toMain(t, g)
	if err := jmpB4CastSandboxSpell(t, g, me); err != nil {
		t.Fatalf("its controller casts: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("its controller attacks after casting: %v", err)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if err := jmpB4CastSandboxSpell(t, g, me); err != nil {
		t.Fatalf("its controller casts after attacking: %v", err)
	}
}

func TestJmpB4ArbiterAnOpponentWhoCastNothingYetMayCastBeforeAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Angelic Arbiter", "Creature — Angel", jmpB4ArbiterOracle, false)
	aangAdvanceToMain(t, g, 1)
	if err := jmpB4CastSandboxSpell(t, g, opp); err != nil {
		t.Fatalf("an opponent who has not attacked may cast: %v", err)
	}
}

// --- Basri Ket -------------------------------------------------------------

func jmpB4Basri(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return pushCatalogWalkerWithLine(g, owner, "Basri Ket", "Legendary Planeswalker — Basri", jmpB4BasriKetOracle, loyalty)
}

func TestJmpB4BasriKetPlusOneCounterAndIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	basri := jmpB4Basri(g, me.ID, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, basri, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	passPriorityAroundTable(t, g)
	if loyaltyCount(g, basri) != 4 {
		t.Errorf("loyalty %d, want 4", loyaltyCount(g, basri))
	}
	if n := jmpB4PlusOneCounters(t, g, bear); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
	if !game.HasKeyword(jmpB4Live(t, g, bear), "indestructible") {
		t.Error("the target gains indestructible")
	}
}

func TestJmpB4BasriKetPlusOneWithNoTargetOnlyAddsLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	basri := jmpB4Basri(g, me.ID, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, basri, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if loyaltyCount(g, basri) != 4 {
		t.Errorf("loyalty %d, want 4", loyaltyCount(g, basri))
	}
	if n := jmpB4PlusOneCounters(t, g, bear); n != 0 {
		t.Errorf("no target, yet %d counters", n)
	}
	if game.HasKeyword(jmpB4Live(t, g, bear), "indestructible") {
		t.Error("no target, yet indestructible")
	}
}

func TestJmpB4BasriKetMinusTwoMakesASoldierForEachNontokenAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	basri := jmpB4Basri(g, me.ID, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	spirit := b12Creature(g, me.ID, "Spirit", "Token Creature — Spirit", 1, 1)
	b16Activate(t, g, me.ID, basri, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if loyaltyCount(g, basri) != 1 {
		t.Fatalf("loyalty %d, want 1", loyaltyCount(g, basri))
	}
	declareAttack(t, g, opp.ID, bear, elf, spirit)
	passAndSettleAttackSplit(t, g, me.ID)
	soldiers := b33AttackingTokensNamed(g, me.ID, opp.ID, "Soldier")
	if len(soldiers) != 2 {
		t.Fatalf("two nontoken attackers: %d Soldiers tapped and attacking, want 2", len(soldiers))
	}
	for _, id := range soldiers {
		c := jmpB4Live(t, g, id)
		if !c.Tapped || c.CurrentPower() != 1 || c.CurrentToughness() != 1 || !c.HasColor("W") {
			t.Errorf("a tapped 1/1 white Soldier: tapped %v, %d/%d", c.Tapped, c.CurrentPower(), c.CurrentToughness())
		}
	}
}

func TestJmpB4BasriKetMinusTwoIgnoresAnAttackOfTokensOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	basri := jmpB4Basri(g, me.ID, 3)
	spirit := b12Creature(g, me.ID, "Spirit", "Token Creature — Spirit", 1, 1)
	b16Activate(t, g, me.ID, basri, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, spirit)
	passAndSettleAttackSplit(t, g, me.ID)
	if n := countBattlefieldNamed(g, me.ID, "Soldier"); n != 0 {
		t.Errorf("only a token attacked, yet %d Soldiers", n)
	}
}

func TestJmpB4BasriKetMinusTwoLastsOnlyThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	basri := jmpB4Basri(g, me.ID, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, basri, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, 0)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(bear) })
	declareAttack(t, g, opp.ID, bear)
	passAndSettleAttackSplit(t, g, me.ID)
	if n := countBattlefieldNamed(g, me.ID, "Soldier"); n != 0 {
		t.Errorf("a later turn's attack made %d Soldiers", n)
	}
}

func TestJmpB4BasriKetEmblemMakesASoldierThenCountersEveryCreature(t *testing.T) {
	spec, ok := Lookup(jmpB4BasriKetOracle)
	if !ok || spec.Emblem == nil || len(spec.Emblem.Triggered) != 1 {
		t.Fatal("Basri Ket's emblem is not declared")
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	basri := jmpB4Basri(g, me.ID, 6)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, basri, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	var soldier uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Soldier" && c.IsToken() {
			soldier = c.InstanceID
		}
	}
	if soldier == uuid.Nil {
		t.Fatal("the emblem made no Soldier")
	}
	if n := jmpB4PlusOneCounters(t, g, soldier); n != 1 {
		t.Errorf("the new Soldier has %d counters, want 1", n)
	}
	if n := jmpB4PlusOneCounters(t, g, bear); n != 1 {
		t.Errorf("the Bear has %d counters, want 1", n)
	}
	if n := jmpB4PlusOneCounters(t, g, theirs); n != 0 {
		t.Errorf("an opponent's creature got %d counters", n)
	}
}

// --- Basri's Acolyte -------------------------------------------------------

func TestJmpB4AcolyteCountersUpToTwoOtherCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	acolyte := jmpB4CastCreature(t, g, me, "Basri's Acolyte", "Creature — Cat Cleric", jmpB4AcolyteOracle, 2, 3)
	passPriorityAroundTable(t, g)
	if err := jmpB4Pick(t, g, me.ID, theirs); err == nil {
		t.Fatal("an opponent's creature was accepted")
	}
	if err := jmpB4Pick(t, g, me.ID, acolyte); err == nil {
		t.Fatal("the Acolyte itself was accepted (\"other\")")
	}
	if err := jmpB4Pick(t, g, me.ID, bear, elf); err != nil {
		t.Fatalf("two other creatures you control: %v", err)
	}
	passPriorityAroundTable(t, g)
	if jmpB4PlusOneCounters(t, g, bear) != 1 || jmpB4PlusOneCounters(t, g, elf) != 1 {
		t.Error("each target gets a +1/+1 counter")
	}
	if jmpB4PlusOneCounters(t, g, acolyte) != 0 || jmpB4PlusOneCounters(t, g, theirs) != 0 {
		t.Error("only the targets")
	}
	if !game.HasKeyword(jmpB4Live(t, g, acolyte), "lifelink") {
		t.Error("the Acolyte has lifelink")
	}
}

// --- Basri's Solidarity ----------------------------------------------------

func TestJmpB4SolidarityCountersEachCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Basri's Solidarity", "Sorcery", jmpB4SolidarityOracle, nil)
	passPriorityAroundTable(t, g)
	if jmpB4PlusOneCounters(t, g, bear) != 1 || jmpB4PlusOneCounters(t, g, elf) != 1 {
		t.Error("each creature you control gets a +1/+1 counter")
	}
	if jmpB4PlusOneCounters(t, g, theirs) != 0 {
		t.Error("an opponent's creature got a counter")
	}
}

// --- Battlefield Promotion -------------------------------------------------

func TestJmpB4PromotionCounterFirstStrikeAndLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	castCatalogSpell(t, g, "Battlefield Promotion", "Instant", jmpB4PromotionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if jmpB4PlusOneCounters(t, g, bear) != 1 {
		t.Error("the target gets a +1/+1 counter")
	}
	if !game.HasKeyword(jmpB4Live(t, g, bear), "first strike") {
		t.Error("the target gains first strike")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestJmpB4PromotionWithItsTargetGoneDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	castCatalogSpell(t, g, "Battlefield Promotion", "Instant", jmpB4PromotionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("an illegal target, yet life %d → %d", life, me.Life)
	}
}

// --- Blessed Sanctuary -----------------------------------------------------

func TestJmpB4SanctuaryPreventsNoncombatDamageToYouAndYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Blessed Sanctuary", "Enchantment", jmpB4SanctuaryOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", boltOracleCombat, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("noncombat damage to you: life %d → %d", life, me.Life)
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", boltOracleCombat, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("noncombat damage to a creature you control is prevented")
	}
}

func TestJmpB4SanctuaryLetsCombatDamageAndOtherCreaturesThrough(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "Blessed Sanctuary", "Enchantment", jmpB4SanctuaryOracle, false)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", boltOracleCombat, []game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("a creature its controller doesn't control is not protected")
	}
	raider := b12Creature(g, me.ID, "Raider", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, raider)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("combat damage is not prevented: life %d, want %d", opp.Life, life-2)
	}
}

func TestJmpB4SanctuaryMakesAUnicornForEachNontokenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	pushCatalogPermanent(g, me.ID, "Blessed Sanctuary", "Enchantment", jmpB4SanctuaryOracle, false)
	jmpB4CastCreature(t, g, me, "Plain Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	var unicorns []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Unicorn" {
			unicorns = append(unicorns, c)
		}
	}
	if len(unicorns) != 1 {
		t.Fatalf("one nontoken creature entered: %d Unicorns, want 1 (the Unicorn itself makes none)", len(unicorns))
	}
	u := unicorns[0]
	if !u.IsToken() || u.Power != 2 || u.Toughness != 2 || !u.HasColor("W") || !u.HasSubtype("Unicorn") {
		t.Errorf("a 2/2 white Unicorn token, got %+v", u)
	}
}

// --- Brightmare ------------------------------------------------------------

func TestJmpB4BrightmareTapsAndGainsLifeEqualToPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 4, 4)
	life := me.Life
	jmpB4CastCreature(t, g, me, "Brightmare", "Creature — Unicorn", jmpB4BrightmareOracle, 2, 3)
	passPriorityAroundTable(t, g)
	if err := jmpB4Pick(t, g, me.ID, giant); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, giant) {
		t.Error("the target is tapped")
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func TestJmpB4BrightmareWithNoTargetOrAGoneTargetDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	life := me.Life
	jmpB4CastCreature(t, g, me, "Brightmare", "Creature — Unicorn", jmpB4BrightmareOracle, 2, 3)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		if err := jmpB4Pick(t, g, me.ID); err != nil {
			t.Fatalf("no target is a legal answer: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("no target, yet life %d → %d", life, me.Life)
	}

	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 4, 4)
	jmpB4CastCreature(t, g, me, "Brightmare", "Creature — Unicorn", jmpB4BrightmareOracle, 2, 3)
	passPriorityAroundTable(t, g)
	if err := jmpB4Pick(t, g, me.ID, giant); err != nil {
		t.Fatalf("pick: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(giant) })
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("the target left, yet life %d → %d", life, me.Life)
	}
}

// --- Celestial Mantle ------------------------------------------------------

func TestJmpB4MantlePumpsAndDoublesLifeOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Celestial Mantle", "Enchantment — Aura", jmpB4MantleOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	c := jmpB4Live(t, g, bear)
	if c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Fatalf("enchanted creature is %d/%d, want 5/5", c.CurrentPower(), c.CurrentToughness())
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 7-me.Life) })
	attackWith(t, g, opp.ID, bear, other)
	passPriorityAroundTable(t, g)
	if me.Life != 14 {
		t.Errorf("life %d, want 14 (doubled once, for the enchanted creature only)", me.Life)
	}
}

func TestJmpB4MantleDoublesTheEnchantedCreaturesController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// My Mantle on an opponent's creature: the opponent's life doubles.
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Celestial Mantle", "Enchantment — Aura", jmpB4MantleOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 10-me.Life)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 6-opp.Life)
	})
	aangAdvanceToMain(t, g, 1)
	attackWith(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if opp.Life != 12 {
		t.Errorf("the creature's controller doubles: life %d, want 12", opp.Life)
	}
	if me.Life != 5 {
		t.Errorf("the Aura's controller took 5 and did not double: life %d, want 5", me.Life)
	}
}
