package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 GAP2 batch 5 (#5948): spell and ability copies and retargeting, and
// spell-on-stack characteristics.

const (
	gap2b05WyllsReversalOracle  = "8d35cef8-a52d-45fb-8f5f-cccea26826d0"
	gap2b05NotOfThisWorldOracle = "5da5d8a3-009e-4d8a-977e-6ebc73ba2682"
	gap2b05AgrusKosOracle       = "3d19823a-dbb8-4384-81f4-5c6818b7f46d"
	gap2b05PyromancerOracle     = "29b3cb3e-aeb1-43eb-8095-9f7881aa34bd"
	gap2b05DynaheirOracle       = "a968524a-3d2d-4c0c-8f66-bfbfb9ff3c36"
	gap2b05WalkingBallistaOra   = "4b515bb0-f275-4400-8032-3173b799ab40"
	gap2b05ZevlorOracle         = "0c85a577-db82-4a36-bc42-49644eba1cf2"
	gap2b05NeeraOracle          = "88fb1dd1-60fb-4f51-ac4d-30b43156a474"
	gap2b05IllusionOracle       = "5af2689d-c45c-4d45-b6a9-f32e8410b336"
	gap2b05EightTailsOracle     = "01c8d29f-4924-4d55-80bc-0c0c06bcb733"
	gap2b05ThiefOracle          = "7abbd5af-a04b-476d-aee5-2b2f5114b019"
	gap2b05JadeOrbOracle        = "5eef3d70-ef16-4722-8c3d-21a0311597bd"
	gap2b05MizzixOracle         = "24a7e1d2-0bc9-40a9-a606-f0017bde3dab"
)

func TestS82Gap2B05CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap2b05WyllsReversalOracle:   "Wyll's Reversal",
		gap2b05NotOfThisWorldOracle:  "Not of This World",
		gap2b05AgrusKosOracle:        "Agrus Kos, Eternal Soldier",
		gap2b05DynaheirOracle:        "Dynaheir, Invoker Adept",
		gap2b05ZevlorOracle:          "Zevlor, Elturel Exile",
		gap2b05NeeraOracle:           "Neera, Wild Mage",
		gap2b05IllusionOracle:        "Illusion",
		gap2b05IllusionOracle + "#1": "Reality",
		gap2b05EightTailsOracle:      "Eight-and-a-Half-Tails",
		gap2b05ThiefOracle:           "Thief of Existence",
		gap2b05JadeOrbOracle:         "Jade Orb of Dragonkind",
		gap2b05MizzixOracle:          "Mizzix, Replica Rider",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

// --- Wyll's Reversal --------------------------------------------------------

// gap2b05Wyll is a table whose next d20 from the Reversal shows `natural`,
// with a Lightning Bolt of the active player's aimed at `victim` and the
// Reversal cast at it; `power` is a creature of the caster's (none for 0).
func gap2b05Wyll(t *testing.T, natural, power int) (g *game.Game, me, victim, other uuid.UUID) {
	t.Helper()
	wyll := uuid.New()
	g = effectstest.D5681Table(t, wyll, natural)
	p, v, o := effectstest.BlbB1Seats(g)
	me, victim, other = p.ID, v.ID, o.ID
	if power > 0 {
		effectstest.B12Creature(g, me, "Brute", "Creature — Giant", power, power)
	}
	bolt := effectstest.CastFor(t, g, p, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtPlayerRef(victim)})
	if err := gapCastWith(t, g, wyll, "Wyll's Reversal", "Instant", gap2b05WyllsReversalOracle, effectstest.RtCardTarget(bolt)); err != nil {
		t.Fatalf("cast Wyll's Reversal: %v", err)
	}
	for i := 0; i < 8 && effectstest.LatestRetarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.LatestRetarget(g, me) == nil {
		t.Fatal("Wyll's Reversal opened no retarget prompt")
	}
	return g, me, victim, other
}

// 14 or less: the Bolt may be redirected, and nothing is copied.
func TestS82Gap2B05WyllsReversalRetargetsOnlyBelowFifteen(t *testing.T) {
	g, me, victim, other := gap2b05Wyll(t, 5, 9) // 5 + 9 = 14
	vBefore, oBefore := effectstest.LifeOf(g, victim), effectstest.LifeOf(g, other)
	effectstest.AnswerRetargetWith(t, g, me, effectstest.RtPlayerRef(other))
	effectstest.SettleThroughCopyRetarget(t, g, me, victim)
	if got := effectstest.LifeOf(g, other); got != oBefore-3 {
		t.Errorf("the new target has %d life, want %d (the Bolt alone)", got, oBefore-3)
	}
	if got := effectstest.LifeOf(g, victim); got != vBefore {
		t.Errorf("the old target has %d life, want %d (no copy on 14)", got, vBefore)
	}
}

// 15+ (the greatest power is added): redirect, then a copy the caster aims.
func TestS82Gap2B05WyllsReversalCopiesOnFifteenWithThePowerAdded(t *testing.T) {
	g, me, victim, other := gap2b05Wyll(t, 5, 10) // 5 + 10 = 15
	vBefore, oBefore := effectstest.LifeOf(g, victim), effectstest.LifeOf(g, other)
	effectstest.AnswerRetargetWith(t, g, me, effectstest.RtPlayerRef(other))
	effectstest.SettleThroughCopyRetarget(t, g, me, victim)
	if got := effectstest.LifeOf(g, other); got != oBefore-3 {
		t.Errorf("the redirected Bolt: %d life, want %d", got, oBefore-3)
	}
	if got := effectstest.LifeOf(g, victim); got != vBefore-3 {
		t.Errorf("the copy: %d life, want %d", got, vBefore-3)
	}
}

// A natural 20 with no creatures copies; declining the redirect keeps the
// Bolt on its target.
func TestS82Gap2B05WyllsReversalMayDeclineTheRedirect(t *testing.T) {
	g, me, victim, other := gap2b05Wyll(t, 20, 0)
	vBefore, oBefore := effectstest.LifeOf(g, victim), effectstest.LifeOf(g, other)
	effectstest.AnswerRetargetWith(t, g, me, nil)
	effectstest.SettleThroughCopyRetarget(t, g, me, other)
	if got := effectstest.LifeOf(g, victim); got != vBefore-3 {
		t.Errorf("the Bolt: %d life, want %d", got, vBefore-3)
	}
	if got := effectstest.LifeOf(g, other); got != oBefore-3 {
		t.Errorf("the copy: %d life, want %d", got, oBefore-3)
	}
}

// A spell with no targets is not a legal target.
func TestS82Gap2B05WyllsReversalRefusesASpellWithNoTargets(t *testing.T) {
	wyll := uuid.New()
	g := effectstest.D5681Table(t, wyll, 20)
	me := g.Seats[g.Turn.ActiveSeat]
	draw := effectstest.CastFor(t, g, me, "Divination", "Sorcery", "", game.CastSpellParams{})
	if err := gapCastWith(t, g, wyll, "Wyll's Reversal", "Instant", gap2b05WyllsReversalOracle, effectstest.RtCardTarget(draw)); err == nil {
		t.Fatal("a spell with no targets was accepted")
	}
}

// --- Not of This World -------------------------------------------------------

func gap2b05CastNotOfThisWorld(g *game.Game, me *game.Player, bolt uuid.UUID) error {
	id := uuid.New()
	effectstest.D5681Spell(g, me, id, "Not of This World", "Kindred Instant — Eldrazi", "{7}", gap2b05NotOfThisWorldOracle)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.RtCardTarget(bolt)})
}

// A spell aimed at a power-7 creature of the caster's is countered for {0}.
func TestS82Gap2B05NotOfThisWorldIsFreeAgainstASpellAtABigCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp, _ := effectstest.BlbB1Seats(g)
	titan := effectstest.B12Creature(g, me.ID, "Titan", "Creature — Eldrazi", 7, 7)
	bolt := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: titan})
	if err := gap2b05CastNotOfThisWorld(g, me, bolt); err != nil {
		t.Fatalf("a {0} cast was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, titan).DamageMarked; got != 0 {
		t.Errorf("the Bolt resolved: %d damage marked", got)
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the countered Bolt is not in its owner's graveyard")
	}
}

// Against a smaller creature it costs {7}: an empty pool can't pay, seven can.
func TestS82Gap2B05NotOfThisWorldCostsSevenForASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bolt := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if err := gap2b05CastNotOfThisWorld(g, me, bolt); err == nil {
		t.Fatal("cast for nothing against a 2-power creature")
	}
	effectstest.B10AddMana(me, "C", "C", "C", "C", "C", "C", "C")
	if err := gap2b05CastNotOfThisWorld(g, me, bolt); err != nil {
		t.Fatalf("seven mana was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bolt resolved and killed the Bear")
	}
}

// A spell aimed at the caster, or at another player's permanent, is not a
// legal target.
func TestS82Gap2B05NotOfThisWorldRefusesASpellNotAtYourPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp, other := effectstest.BlbB1Seats(g)
	effectstest.B10AddMana(me, "C", "C", "C", "C", "C", "C", "C")
	theirs := effectstest.B12Creature(g, other.ID, "Giant", "Creature — Giant", 8, 8)
	atMe := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if err := gap2b05CastNotOfThisWorld(g, me, atMe); err == nil {
		t.Error("a spell aimed at a player was accepted")
	}
	atTheirs := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if err := gap2b05CastNotOfThisWorld(g, me, atTheirs); err == nil {
		t.Error("a spell aimed at another player's creature was accepted")
	}
}

// --- Agrus Kos, Eternal Soldier ---------------------------------------------

type gap2b05AgrusTable struct {
	g                *game.Game
	me, opp          *game.Player
	agrus, pyro      uuid.UUID
	bear, wolf, ogre uuid.UUID
}

// gap2b05Agrus seats Agrus Kos with two other creatures of the active
// player's, an opponent's creature, and a Prodigal Pyromancer of the active
// player's ({T}: 1 damage to any target).
func gap2b05Agrus(t *testing.T) *gap2b05AgrusTable {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp, _ := effectstest.BlbB1Seats(g)
	tb := &gap2b05AgrusTable{g: g, me: me, opp: opp}
	tb.agrus = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Agrus Kos, Eternal Soldier",
		TypeLine: "Legendary Creature — Spirit Soldier", OracleID: gap2b05AgrusKosOracle, Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID})
	tb.pyro = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Prodigal Pyromancer",
		TypeLine: "Creature — Human Wizard", OracleID: gap2b05PyromancerOracle, Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID})
	tb.bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	tb.wolf = effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	tb.ogre = effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	return tb
}

func (tb *gap2b05AgrusTable) ping(t *testing.T, at game.TargetRef) {
	t.Helper()
	if err := tb.g.ActivateCatalogAbility(tb.me.ID, tb.pyro, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{at}}); err != nil {
		t.Fatalf("activate the Pyromancer: %v", err)
	}
}

func (tb *gap2b05AgrusTable) passUntilPay(t *testing.T) bool {
	t.Helper()
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(tb.g, tb.me.ID); i++ {
		if effectstest.StackFullyEmpty(tb.g) {
			return false
		}
		if err := tb.g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return effectstest.HasPayUnlessFor(tb.g, tb.me.ID)
}

func (tb *gap2b05AgrusTable) damage(t *testing.T, id uuid.UUID) int {
	t.Helper()
	return effectstest.FindCardAnywhere(t, tb.g, id).DamageMarked
}

// Paying {1}{R/W} copies the ping once for each other creature of Agrus's
// controller; the opponent's creature is not one of them.
func TestS82Gap2B05AgrusKosCopiesTheAbilityForEachOtherCreature(t *testing.T) {
	tb := gap2b05Agrus(t)
	tb.ping(t, game.TargetRef{Kind: game.TargetCard, ID: tb.agrus})
	if !tb.passUntilPay(t) {
		t.Fatal("Agrus Kos offered no payment")
	}
	effectstest.B10AddMana(tb.me, "C", "W")
	effectstest.AnswerPayUnless(t, tb.g, tb.me.ID, true)
	effectstest.PassPriorityAroundTable(t, tb.g)
	for i := 0; i < 8 && !effectstest.StackFullyEmpty(tb.g); i++ {
		effectstest.PassPriorityAroundTable(t, tb.g)
	}
	for name, id := range map[string]uuid.UUID{"Agrus Kos": tb.agrus, "the Bear": tb.bear, "the Wolf": tb.wolf} {
		if got := tb.damage(t, id); got != 1 {
			t.Errorf("%s has %d damage, want 1", name, got)
		}
	}
	if got := tb.damage(t, tb.ogre); got != 0 {
		t.Errorf("the opponent's Ogre has %d damage, want 0", got)
	}
	if got := tb.damage(t, tb.pyro); got != 1 {
		t.Errorf("the Pyromancer (another creature Agrus's controller controls) has %d damage, want 1", got)
	}
}

// Declining leaves only the original ability.
func TestS82Gap2B05AgrusKosMayDecline(t *testing.T) {
	tb := gap2b05Agrus(t)
	tb.ping(t, game.TargetRef{Kind: game.TargetCard, ID: tb.agrus})
	if !tb.passUntilPay(t) {
		t.Fatal("Agrus Kos offered no payment")
	}
	effectstest.AnswerPayUnless(t, tb.g, tb.me.ID, false)
	effectstest.PassPriorityAroundTable(t, tb.g)
	if got := tb.damage(t, tb.agrus); got != 1 {
		t.Errorf("Agrus Kos has %d damage, want 1", got)
	}
	if got := tb.damage(t, tb.bear); got != 0 {
		t.Errorf("the Bear has %d damage, want 0", got)
	}
}

// A spell, or an ability aimed at another creature, does not trigger it.
func TestS82Gap2B05AgrusKosIgnoresSpellsAndOtherTargets(t *testing.T) {
	tb := gap2b05Agrus(t)
	tb.ping(t, game.TargetRef{Kind: game.TargetCard, ID: tb.bear})
	if tb.passUntilPay(t) {
		t.Fatal("an ability at another creature triggered Agrus Kos")
	}
	effectstest.CastFor(t, tb.g, tb.me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtCardTarget(tb.agrus)})
	if tb.passUntilPay(t) {
		t.Fatal("a spell triggered Agrus Kos")
	}
}

// --- Dynaheir, Invoker Adept --------------------------------------------------

func gap2b05Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int, sick bool) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		OracleID: oracle, Power: p, Toughness: tough, Owner: owner, Controller: owner, SummonedThisTurn: sick})
}

// A summoning-sick creature's {T} ability can be activated with Dynaheir on
// the battlefield, and not without it.
func TestS82Gap2B05DynaheirLetsOtherSickCreaturesTap(t *testing.T) {
	for _, withDynaheir := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
		me, opp, _ := effectstest.BlbB1Seats(g)
		if withDynaheir {
			gap2b05Push(g, me.ID, "Dynaheir, Invoker Adept", "Legendary Creature — Human Wizard", gap2b05DynaheirOracle, 4, 4, true)
		}
		pyro := effectstest.PushCatalogPermanent(g, me.ID, "Prodigal Pyromancer", "Creature — Human Wizard", gap2b05PyromancerOracle, true)
		err := g.ActivateCatalogAbility(me.ID, pyro, 0, game.ActivateAbilityParams{Targets: effectstest.RtPlayerRef(opp.ID)})
		if withDynaheir && err != nil {
			t.Errorf("with Dynaheir the sick Pyromancer was refused: %v", err)
		}
		if !withDynaheir && err == nil {
			t.Error("without Dynaheir the sick Pyromancer was activated")
		}
	}
}

func gap2b05ActivateDynaheir(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	dyn := gap2b05Push(g, me.ID, "Dynaheir, Invoker Adept", "Legendary Creature — Human Wizard", gap2b05DynaheirOracle, 4, 4, false)
	ballista := gap2b05Push(g, me.ID, "Walking Ballista", "Artifact Creature — Construct", gap2b05WalkingBallistaOra, 0, 0, false)
	effectstest.AddCountersForTest(g, ballista, game.CounterPlusOne, 1)
	if err := g.ActivateCatalogAbility(me.ID, dyn, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Dynaheir: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, ballista
}

// The next ability activated with four or more mana is copied: Walking
// Ballista's {4} counter ability puts two counters on it. A later one is
// not copied (the trigger fired once).
func TestS82Gap2B05DynaheirCopiesTheNextFourManaActivation(t *testing.T) {
	g, me, ballista := gap2b05ActivateDynaheir(t)
	for round, want := range []int{3, 4} {
		effectstest.B10AddMana(me, "C", "C", "C", "C")
		if err := g.ActivateCatalogAbility(me.ID, ballista, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate the Ballista: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		for i := 0; i < 8 && !effectstest.StackFullyEmpty(g); i++ {
			effectstest.PassPriorityAroundTable(t, g)
		}
		if got := effectstest.PlusOneCounters(g, ballista); got != want {
			t.Errorf("round %d: %d counters, want %d", round, got, want)
		}
	}
}

// An activation that spends no mana neither fires nor uses up the trigger.
func TestS82Gap2B05DynaheirIgnoresAnActivationWithoutFourMana(t *testing.T) {
	g, me, ballista := gap2b05ActivateDynaheir(t)
	effectstest.AddCountersForTest(g, ballista, game.CounterPlusOne, 1)
	if err := g.ActivateCatalogAbility(me.ID, ballista, 1, game.ActivateAbilityParams{Targets: effectstest.RtPlayerRef(g.Seats[1].ID)}); err != nil {
		t.Fatalf("activate the Ballista's ping: %v", err)
	}
	before := effectstest.LifeOf(g, g.Seats[1].ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, g.Seats[1].ID); got != before-1 {
		t.Errorf("the ping: life %d -> %d, want -1 (no copy)", before, got)
	}
	effectstest.B10AddMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, ballista, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the Ballista: %v", err)
	}
	for i := 0; i < 8 && !effectstest.StackFullyEmpty(g); i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.PlusOneCounters(g, ballista); got != 3 {
		t.Errorf("%d counters, want 3 (1 left after the ping, then the {4} ability and its copy)", got)
	}
}

// --- Zevlor, Elturel Exile -----------------------------------------------------

// gap2b05Zevlor activates Zevlor for the active player and resolves it.
func gap2b05Zevlor(t *testing.T) (g *game.Game, me *game.Player, opps []*game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me = g.Seats[g.Turn.ActiveSeat]
	for i := 1; i < len(g.Seats); i++ {
		opps = append(opps, g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)])
	}
	zev := gap2b05Push(g, me.ID, "Zevlor, Elturel Exile", "Legendary Creature — Tiefling Warrior", gap2b05ZevlorOracle, 4, 2, false)
	effectstest.B10AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, zev, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Zevlor: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, opps
}

// gap2b05PickObject answers the newest option pick of `chooser` with the
// option naming `id` (a seat or a card).
func gap2b05PickObject(t *testing.T, g *game.Game, chooser, id uuid.UUID) {
	t.Helper()
	for i := 0; i < 8 && effectstest.LatestOptionPickFor(g, chooser) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := effectstest.LatestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatal("no option pick was opened")
	}
	for i, o := range c.PickOptions {
		if o.Player == id || (len(o.Cards) == 1 && o.Cards[0] == id) {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			return
		}
	}
	t.Fatalf("%s is not offered: %+v", id, c.PickOptions)
}

// A Bolt at one opponent is copied once for each other opponent, at the
// player or the permanent chosen for each.
func TestS82Gap2B05ZevlorCopiesTheSpellForEachOtherOpponent(t *testing.T) {
	g, me, opps := gap2b05Zevlor(t)
	bear := effectstest.B12Creature(g, opps[2].ID, "Bear", "Creature — Bear", 2, 2)
	before := []int{opps[0].Life, opps[1].Life, opps[2].Life}
	effectstest.CastFor(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtPlayerRef(opps[0].ID)})
	gap2b05PickObject(t, g, me.ID, opps[1].ID)
	gap2b05PickObject(t, g, me.ID, bear)
	for i := 0; i < 8 && !effectstest.StackFullyEmpty(g); i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if opps[0].Life != before[0]-3 || opps[1].Life != before[1]-3 || opps[2].Life != before[2] {
		t.Errorf("life %v -> [%d %d %d], want the first two -3", before, opps[0].Life, opps[1].Life, opps[2].Life)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the copy aimed at the Bear did not kill it")
	}
}

// A spell aimed at Zevlor's controller does not fire the trigger, and does
// not use it up: the next one at an opponent does.
func TestS82Gap2B05ZevlorWaitsForASpellAtASingleOpponent(t *testing.T) {
	g, me, opps := gap2b05Zevlor(t)
	effectstest.CastFor(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtPlayerRef(me.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Fatal("a spell at its controller fired Zevlor's trigger")
	}
	before := opps[1].Life
	effectstest.CastFor(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtPlayerRef(opps[0].ID)})
	gap2b05PickObject(t, g, me.ID, opps[1].ID)
	gap2b05PickObject(t, g, me.ID, opps[2].ID)
	for i := 0; i < 8 && !effectstest.StackFullyEmpty(g); i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if opps[1].Life != before-3 {
		t.Errorf("the second opponent: %d life, want %d", opps[1].Life, before-3)
	}
}

// A copy whose chosen target is illegal for the spell is not created (CR
// 707.10e): Lightning Bolt can't target a noncreature artifact.
func TestS82Gap2B05ZevlorMakesNoCopyAtAnIllegalChoice(t *testing.T) {
	g, me, opps := gap2b05Zevlor(t)
	before := opps[1].Life
	effectstest.CastFor(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtPlayerRef(opps[0].ID)})
	art := effectstest.B12Permanent(g, opps[1].ID, "Relic", "Artifact")
	gap2b05PickObject(t, g, me.ID, art)
	stackBefore := len(g.Stack.Cards)
	gap2b05PickObject(t, g, me.ID, opps[2].ID)
	if got := len(g.Stack.Cards); got != stackBefore+1 {
		t.Errorf("stack %d -> %d, want one copy (for the third opponent only)", stackBefore, got)
	}
	for i := 0; i < 8 && !effectstest.StackFullyEmpty(g); i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if opps[1].Life != before || !g.Battlefield.Contains(art) {
		t.Error("a copy was made at a noncreature artifact Lightning Bolt can't target")
	}
}

// --- Neera, Wild Mage ------------------------------------------------------------

// gap2b05Neera seats Neera for the active player with a land on top of
// their library and a vanilla creature card under it, casts a Lightning
// Bolt at an opponent and says yes to Neera's trigger.
func gap2b05Neera(t *testing.T) (g *game.Game, me, opp *game.Player, bolt, land, giant uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp, _ = effectstest.BlbB1Seats(g)
	gap2b05Push(g, me.ID, "Neera, Wild Mage", "Legendary Creature — Human Elf Shaman", gap2b05NeeraOracle, 2, 7, false)
	giant = effectstest.PutOnTopOfLibrary(g, me, "Hill Giant", "Creature — Giant", "")
	land = effectstest.PutOnTopOfLibrary(g, me, "Forest", "Basic Land — Forest", "")
	bolt = effectstest.CastFor(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtPlayerRef(opp.ID)})
	if !gap2b05PassUntilTriggerPrompt(t, g, me.ID) {
		t.Fatal("Neera asked nothing")
	}
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	return g, me, opp, bolt, land, giant
}

// gap2b05PassUntilTriggerPrompt passes priority until `chooser` owes a
// trigger's "you may" or the stack is empty.
func gap2b05PassUntilTriggerPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) bool {
	t.Helper()
	for i := 0; i < 8; i++ {
		if effectstest.LatestTriggerPromptFor(g, chooser) != nil {
			return true
		}
		if effectstest.StackFullyEmpty(g) {
			return false
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return effectstest.LatestTriggerPromptFor(g, chooser) != nil
}

func gap2b05LibraryBottom(p *game.Player, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n && i < len(p.Library.Cards); i++ {
		out = append(out, p.Library.Cards[i].InstanceID)
	}
	return out
}

// The spell goes to the bottom of the library uncountered, the nonland card
// revealed is cast free, the land goes to the bottom; and the ability does
// not trigger a second time this turn.
func TestS82Gap2B05NeeraTucksTheSpellAndCastsTheRevealedCardFree(t *testing.T) {
	g, me, opp, bolt, land, giant := gap2b05Neera(t)
	before := opp.Life
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, giant)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, giant, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the revealed card: %v", err)
	}
	for i := 0; i < 8 && !effectstest.StackFullyEmpty(g); i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(giant) {
		t.Error("the revealed creature was not cast")
	}
	if opp.Life != before {
		t.Errorf("the tucked Bolt resolved: life %d -> %d", before, opp.Life)
	}
	bottom := gap2b05LibraryBottom(me, 2)
	if !gap2b05Has(bottom, bolt) || !gap2b05Has(bottom, land) {
		t.Errorf("the Bolt and the land should be the bottom two cards: %v (bolt %s, land %s)", bottom, bolt, land)
	}
	effectstest.CastFor(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtPlayerRef(opp.ID)})
	if gap2b05PassUntilTriggerPrompt(t, g, me.ID) {
		t.Error("Neera triggered a second time in one turn")
	}
}

// Declining the cast puts the revealed cards on the bottom too.
func TestS82Gap2B05NeeraDeclinedCastBottomsTheRevealedCards(t *testing.T) {
	g, me, _, bolt, land, giant := gap2b05Neera(t)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, giant)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	for i := 0; i < 8 && !effectstest.StackFullyEmpty(g); i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	bottom := gap2b05LibraryBottom(me, 3)
	for name, id := range map[string]uuid.UUID{"the Bolt": bolt, "the land": land, "the creature": giant} {
		if !gap2b05Has(bottom, id) {
			t.Errorf("%s is not among the bottom three cards", name)
		}
	}
}

func gap2b05Has(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// --- Illusion // Reality ----------------------------------------------------------

func gap2b05IllusionReality(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: gap2b05IllusionOracle, Layout: game.LayoutSplit,
		Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Illusion", TypeLine: "Instant", ManaCost: "{U}",
				OracleText: "Target spell or permanent becomes the color of your choice until end of turn."},
			{Name: "Reality", TypeLine: "Instant", ManaCost: "{2}{G}", OracleText: "Destroy target artifact."},
		},
	}
	c.SettleImported()
	return c
}

func gap2b05CastHalf(t *testing.T, g *game.Game, p *game.Player, face int, target uuid.UUID) {
	t.Helper()
	c := gap2b05IllusionReality(p.ID)
	p.Hand.PushTop(c)
	if err := g.CastSpell(p.ID, c.InstanceID, game.CastSpellParams{Face: face, Targets: effectstest.RtCardTarget(target)}); err != nil {
		t.Fatalf("cast face %d: %v", face, err)
	}
}

func gap2b05StackColors(g *game.Game, id uuid.UUID) ([]string, bool) {
	for _, c := range g.Stack.Cards {
		if c.InstanceID == id {
			return c.EffectiveColors(), true
		}
	}
	return nil, false
}

// A red Lightning Bolt made blue can no longer hit a creature with
// protection from blue: it is removed as it resolves (CR 608.2b, 702.16b).
func TestS82Gap2B05IllusionRecoloursASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Warded Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Keywords: []string{"protection from blue"}, Owner: opp.ID, Controller: opp.ID})
	bolt := effectstest.CastFor(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtCardTarget(bear)})
	gap2b05CastHalf(t, g, me, 0, bolt)
	for i := 0; i < 8 && effectstest.PendingOfKind(g, game.PendingChoiceColor) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerColor(t, g, me.ID, "U")
	if got, ok := gap2b05StackColors(g, bolt); !ok || len(got) != 1 || got[0] != "U" {
		t.Fatalf("the Bolt on the stack is %v (on the stack: %v), want blue", got, ok)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the blue Bolt hit a creature with protection from blue")
	}
}

// A creature spell made green stays green as the creature it becomes (CR
// 400.7a); a permanent is recoloured directly.
func TestS82Gap2B05IllusionRecoloursAPermanentSpellAndAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	giant := effectstest.CastFor(t, g, me, "Hill Giant", "Creature — Giant", "", game.CastSpellParams{})
	gap2b05CastHalf(t, g, me, 0, giant)
	for i := 0; i < 8 && effectstest.PendingOfKind(g, game.PendingChoiceColor) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerColor(t, g, me.ID, "G")
	for i := 0; i < 8 && !effectstest.StackFullyEmpty(g); i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(giant) {
		t.Fatal("the creature spell did not resolve")
	}
	if got := effectstest.EffectiveColors(t, g, giant); len(got) != 1 || got[0] != "G" {
		t.Errorf("the creature is %v, want green", got)
	}

	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	gap2b05CastHalf(t, g, me, 0, rock)
	for i := 0; i < 8 && effectstest.PendingOfKind(g, game.PendingChoiceColor) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerColor(t, g, me.ID, "B")
	if got := effectstest.EffectiveColors(t, g, rock); len(got) != 1 || got[0] != "B" {
		t.Errorf("the artifact is %v, want black", got)
	}
}

func TestS82Gap2B05RealityDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp, _ := effectstest.BlbB1Seats(g)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	gap2b05CastHalf(t, g, me, 1, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("Reality did not destroy the artifact")
	}
}

// --- Eight-and-a-Half-Tails ------------------------------------------------------

func TestS82Gap2B05EightAndAHalfTailsWhitensASpellAndProtects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp, _ := effectstest.BlbB1Seats(g)
	fox := gap2b05Push(g, me.ID, "Eight-and-a-Half-Tails", "Legendary Creature — Fox Cleric", gap2b05EightTailsOracle, 2, 2, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)

	// {1}{W}: only a permanent its controller controls.
	effectstest.B10AddMana(me, "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, fox, 0, game.ActivateAbilityParams{Targets: effectstest.RtCardTarget(theirs)}); err == nil {
		t.Fatal("protection was given to an opponent's creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, fox, 0, game.ActivateAbilityParams{Targets: effectstest.RtCardTarget(bear)}); err != nil {
		t.Fatalf("activate the protection: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "protection from white") {
		t.Fatal("the Bear did not gain protection from white")
	}

	// {1}: a red Bolt at the Bear becomes white and fizzles.
	bolt := effectstest.CastFor(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtCardTarget(bear)})
	effectstest.B10AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, fox, 1, game.ActivateAbilityParams{Targets: effectstest.RtCardTarget(bolt)}); err != nil {
		t.Fatalf("activate the whitening: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	for i := 0; i < 8; i++ {
		if got, ok := gap2b05StackColors(g, bolt); ok && len(got) == 1 && got[0] == "W" {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if got, _ := gap2b05StackColors(g, bolt); len(got) != 1 || got[0] != "W" {
		t.Fatalf("the Bolt is %v, want white", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindCardAnywhere(t, g, bear).DamageMarked != 0 || !g.Battlefield.Contains(bear) {
		t.Error("the white Bolt hit a creature with protection from white")
	}
}

// --- Thief of Existence ----------------------------------------------------------

// gap2b05Settle passes priority to an empty stack, answering every target
// pick of `chooser` with `card` (when not nil) or else `player`.
func gap2b05Settle(t *testing.T, g *game.Game, chooser, card, player uuid.UUID) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			ref := game.TargetRef{Kind: game.TargetPlayer, ID: player}
			if card != uuid.Nil {
				ref = game.TargetRef{Kind: game.TargetCard, ID: card}
				card = uuid.Nil
			}
			if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}

func gap2b05CastThief(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Thief of Existence", TypeLine: "Creature — Eldrazi",
		OracleID: gap2b05ThiefOracle, Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Thief of Existence: %v", err)
	}
	return id
}

// The cast trigger exiles the opponent's artifact, and the creature then
// has "when this creature leaves the battlefield, target opponent draws a
// card".
func TestS82Gap2B05ThiefOfExistenceExilesAndGainsTheLeaveTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp, _ := effectstest.BlbB1Seats(g)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	thief := gap2b05CastThief(t, g, me)
	gap2b05Settle(t, g, me.ID, rock, opp.ID)
	if z := g.FindCardZoneForEffect(rock); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the artifact is in %v, want exile", z)
	}
	if !g.Battlefield.Contains(thief) {
		t.Fatal("Thief of Existence did not resolve")
	}
	hand := len(opp.Hand.Cards)
	effectstest.D5681Dies(t, g, me.ID, thief)
	gap2b05Settle(t, g, me.ID, uuid.Nil, opp.ID)
	if got := len(opp.Hand.Cards); got != hand+1 {
		t.Errorf("the opponent has %d cards, want %d", got, hand+1)
	}
}

// With nothing exiled, the creature gains nothing.
func TestS82Gap2B05ThiefOfExistenceGainsNothingWithoutAnExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	thief := gap2b05CastThief(t, g, me)
	gap2b05Settle(t, g, me.ID, uuid.Nil, opp.ID)
	hand := len(opp.Hand.Cards)
	effectstest.D5681Dies(t, g, me.ID, thief)
	gap2b05Settle(t, g, me.ID, uuid.Nil, opp.ID)
	if got := len(opp.Hand.Cards); got != hand {
		t.Errorf("the opponent drew: %d cards, want %d", got, hand)
	}
}

// --- Jade Orb of Dragonkind ------------------------------------------------------

// A Dragon cast with the Orb's mana enters with a +1/+1 counter and has
// hexproof until its caster's next turn; a Goblin gets neither.
func TestS82Gap2B05JadeOrbCountersAndHexproofsADragon(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		dragon   bool
	}{
		{"Creature — Dragon", true},
		{"Creature — Goblin", false},
	} {
		t.Run(tc.typeLine, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			orb := effectstest.SeedPermanentWithOracle(g, me.ID, "Jade Orb of Dragonkind", "Artifact", gap2b05JadeOrbOracle)
			effectstest.ActivateManaFor(t, g, me.ID, orb, 0, game.ManaAbilityParams{})
			id := effectstest.CastFromHandForTest(t, g, me, "Test Creature", tc.typeLine, "{G}", "", effectstest.Strict)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatal("the creature did not resolve")
			}
			want := 0
			if tc.dragon {
				want = 1
			}
			if got := effectstest.PlusOneCounters(g, id); got != want {
				t.Errorf("%d +1/+1 counters, want %d", got, want)
			}
			if got := effectstest.EffectiveAbilitiesContain(t, g, id, "hexproof"); got != tc.dragon {
				t.Fatalf("hexproof = %v, want %v", got, tc.dragon)
			}
			if !tc.dragon {
				return
			}
			for i := 0; i < len(g.Seats)-1; i++ {
				effectstest.EndTurn(t, g)
				if !effectstest.EffectiveAbilitiesContain(t, g, id, "hexproof") {
					t.Fatalf("hexproof ended on an opponent's turn (%d)", i+1)
				}
			}
			effectstest.EndTurn(t, g)
			if g.Seats[g.Turn.ActiveSeat].ID != me.ID {
				t.Fatal("the test did not reach its caster's next turn")
			}
			if effectstest.EffectiveAbilitiesContain(t, g, id, "hexproof") {
				t.Error("hexproof lasted into its caster's next turn")
			}
		})
	}
}

// --- Mizzix, Replica Rider -------------------------------------------------------

// gap2b05MizzixCast seats Mizzix and casts a 3/3 creature card from the
// active player's graveyard (or hand), returning it.
func gap2b05MizzixCast(t *testing.T, fromGraveyard bool) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	gap2b05Push(g, me.ID, "Mizzix, Replica Rider", "Legendary Creature — Goblin Wizard", gap2b05MizzixOracle, 4, 5, false)
	giant := game.Card{InstanceID: uuid.New(), Name: "Hill Giant", TypeLine: "Creature — Giant", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID}
	params := game.CastSpellParams{}
	if fromGraveyard {
		me.Graveyard.PushTop(giant)
		g.WithWriteLock(func() {
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{Player: me.ID, Zone: game.ZoneGraveyard, CastOnly: true,
				Duration: g.UntilEndOfTurnDuration(), Cost: "{0}", Label: "test — cast it from your graveyard"}, []game.Card{giant})
		})
		params.FromZone = "graveyard"
	} else {
		me.Hand.PushTop(giant)
	}
	if err := g.CastSpell(me.ID, giant.InstanceID, params); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	return g, me, giant.InstanceID
}

// Paying {1}{U/R} copies a spell cast from the graveyard; the token copy
// has haste and is sacrificed at its controller's end step.
func TestS82Gap2B05MizzixCopiesASpellCastFromElsewhere(t *testing.T) {
	g, me, giant := gap2b05MizzixCast(t, true)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.B10AddMana(me, "C", "R")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	for i := 0; i < 8 && !effectstest.StackFullyEmpty(g); i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Hill Giant" && c.InstanceID != giant {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil || !g.Battlefield.Contains(giant) {
		t.Fatal("want the card and a token copy on the battlefield")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, token, "haste") {
		t.Error("the token copy has no haste")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, giant, "haste") {
		t.Error("the original gained haste")
	}
	effectstest.AdvanceToEndStep(t, g)
	for i := 0; i < 8 && !effectstest.StackFullyEmpty(g); i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(token) {
		t.Error("the token copy was not sacrificed at the end step")
	}
	if !g.Battlefield.Contains(giant) {
		t.Error("the original was sacrificed")
	}
}

// A spell cast from hand does not trigger it.
func TestS82Gap2B05MizzixIgnoresASpellFromHand(t *testing.T) {
	g, me, _ := gap2b05MizzixCast(t, false)
	for i := 0; i < 8 && !effectstest.StackFullyEmpty(g); i++ {
		if effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("a spell from hand triggered Mizzix")
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}
