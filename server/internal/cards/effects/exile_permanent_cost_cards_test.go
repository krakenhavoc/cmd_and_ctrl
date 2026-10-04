package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_permanent_cost_cards_test.go — #1600's proof cards for "Exile a
// creature you control" as a cost (ADR 0020's 2026-10-03 amendment):
// The Soul Stone, Food Chain, City of Shadows and Altar of Bhaal. The
// engine half is game/exile_permanent_cost_test.go.
//
// The trigger behaviour is pinned here, with real cards, because it is
// what makes exiling different from sacrificing: Circuit Mender's
// "When this creature leaves the battlefield, draw a card" fires,
// Zulaport Cutthroat's "Whenever this creature or another creature you
// control dies" does not, and Juri's "Whenever you sacrifice a
// permanent" does not.

const (
	theSoulStoneOracle   = "92cfba68-12f6-4f97-9187-0f6a39656a0f"
	foodChainOracle      = "5c8e5092-962e-49ef-ab82-8434e475e4e7"
	cityOfShadowsOracle  = "043192d8-6077-46c3-b43f-b7caf6762869"
	circuitMenderOracle  = "1665ca9f-176d-40f1-a4e9-42da4f1236e9"
	juriRevueOracle      = "3364bad6-6d0a-4141-b410-86e3d9e1916e"
	altarOfBhaalOracleTC = "0a364b66-95df-480b-a733-e90f6d5c4d2b"
)

func pushTheSoulStone(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Soul Stone", OracleID: theSoulStoneOracle,
		TypeLine: "Legendary Artifact — Infinity Stone", ManaCost: "{1}{B}",
		Owner: owner, Controller: owner,
	})
}

// pushCircuitMender seats the LTB-trigger creature with its printed
// mana value ({3}), which Food Chain reads.
func pushCircuitMender(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Circuit Mender", OracleID: circuitMenderOracle,
		TypeLine: "Artifact Creature — Construct", ManaCost: "{3}", Power: 2, Toughness: 3,
		Owner: owner, Controller: owner,
	})
}

func isHarnessed(g *game.Game, id uuid.UUID) bool {
	h := false
	g.ReadSnapshot(func() { h = g.IsHarnessed(id) })
	return h
}

// --- The Soul Stone --------------------------------------------------

// The harness pays its whole printed cost — {6}{B}, {T} and a creature
// exiled at announce — and the exile is an exile: the creature's own
// leaves-the-battlefield trigger goes on the stack ABOVE the harness
// and resolves first, while no dies or sacrifice trigger fires.
func TestTheSoulStoneHarnessExilesACreatureYouControl(t *testing.T) {
	g, me, opp := spendTable(t)
	stone := pushTheSoulStone(g, me.ID)
	mender := pushCircuitMender(g, me.ID)
	pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
	juri := pushCatalogPermanent(g, me.ID, "Juri, Master of the Revue", "Legendary Creature — Human Shaman", juriRevueOracle, false)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{B}")
	hand, oppLife := me.Hand.Size(), opp.Life

	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{
		Strict:            true,
		ExilePermanentIDs: []uuid.UUID{mender},
	}); err != nil {
		t.Fatalf("activate the harness: %v", err)
	}
	if !g.Exile.Contains(mender) {
		t.Fatal("Circuit Mender is not in exile — the cost is paid at announce")
	}
	if c, ok := battlefieldCard(g, stone); !ok || !c.Tapped {
		t.Error("the {T} half of the cost did not tap the Stone")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want the {6}{B} spent", me.ManaPool)
	}
	ltb := triggerOnStack(g, mender)
	if ltb == nil {
		t.Fatal("Circuit Mender's leaves-the-battlefield trigger is not on the stack")
	}
	var harness *game.StackItem
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemActivated && item.SourceCardID == stone {
			harness = item
		}
	}
	if harness == nil || ltb.Seq <= harness.Seq {
		t.Error("the leaves-the-battlefield trigger is not above the harness ability (CR 603.3b)")
	}
	if triggerOnStack(g, juri) != nil {
		t.Error("Juri's \"whenever you sacrifice\" fired — exiling is not sacrificing")
	}

	passPriorityAroundTable(t, g)
	if !isHarnessed(g, stone) {
		t.Error("The Soul Stone is not harnessed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want Circuit Mender's draw", hand, me.Hand.Size())
	}
	if opp.Life != oppLife {
		t.Errorf("opponent life %d -> %d — Zulaport Cutthroat drained for a creature that did not die", oppLife, opp.Life)
	}
	if c, _ := battlefieldCard(g, juri); c.Counters["+1/+1"] != 0 {
		t.Error("Juri got a counter for a creature that was not sacrificed")
	}
}

// With no creature to exile, the harness cannot be paid (CR 118.3) and
// nothing is spent.
func TestTheSoulStoneHarnessNeedsACreature(t *testing.T) {
	g, me, opp := spendTable(t)
	stone := pushTheSoulStone(g, me.ID)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{B}")

	err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{
		Strict:            true,
		ExilePermanentIDs: []uuid.UUID{theirs},
	})
	if !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Fatalf("exiling an opponent's creature: err = %v, want ErrCardCallerMismatch", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the harness activated with no creature exiled")
	}
	if len(me.ManaPool) != 7 || !g.Battlefield.Contains(theirs) {
		t.Error("a refused harness spent mana or moved a creature")
	}
	if c, _ := battlefieldCard(g, stone); c.Tapped {
		t.Error("a refused harness tapped the Stone")
	}
	if isHarnessed(g, stone) {
		t.Error("harnessed without paying")
	}
}

// The ∞ upkeep trigger does not exist until the Stone is harnessed; once
// it is, it returns a creature card from your graveyard to the
// battlefield, under your control.
func TestTheSoulStoneUpkeepReturnsACreatureOnlyOnceHarnessed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stone := pushTheSoulStone(g, me.ID)
	dead := pushGraveyardCardForTest(me, "Dead Bear")
	relic := batch01GraveyardCard(me, "Sol Ring", "Artifact")

	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	if pick := latestPickTarget(g, me.ID); pick != nil || triggerOnStack(g, stone) != nil {
		t.Fatal("the ∞ trigger fired on an unharnessed Soul Stone")
	}

	g.WithWriteLock(func() {
		if err := g.HarnessForEffect(stone); err != nil {
			t.Fatalf("HarnessForEffect: %v", err)
		}
	})
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if hasID(pick.PickTargetCards, relic) {
		t.Error("an artifact card was offered for \"target creature card\"")
	}
	pickCard(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, dead)
	if !ok {
		t.Fatal("the creature card did not return to the battlefield")
	}
	if c.Controller != me.ID {
		t.Error("it returns under your control")
	}
}

// Indestructible, and {T}: Add {B}.
func TestTheSoulStoneIsIndestructibleAndTapsForBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stone := pushTheSoulStone(g, me.ID)

	if !hasString(effectiveAbilities(t, g, stone), "indestructible") {
		t.Error("The Soul Stone is not indestructible")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(stone) })
	if !g.Battlefield.Contains(stone) {
		t.Fatal("a destroy effect removed an indestructible Soul Stone")
	}
	if err := g.ActivateManaAbility(me.ID, stone, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool = %v, want one {B}", got)
	}
}

// --- Food Chain ------------------------------------------------------

// X is 1 plus the exiled creature's mana value, of one colour, and the
// mana is in the pool while the exiled creature's own leaves-the-
// battlefield trigger is still on the stack (CR 605.3a) — no dies
// trigger fires.
func TestFoodChainAddsOnePlusTheExiledCreaturesManaValue(t *testing.T) {
	g, me, opp := spendTable(t)
	chain := pushCatalogPermanent(g, me.ID, "Food Chain", "Enchantment", foodChainOracle, false)
	mender := pushCircuitMender(g, me.ID)
	pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
	oppLife := opp.Life

	if err := g.ActivateManaAbility(me.ID, chain, 0, game.ManaAbilityParams{
		ExilePermanentIDs: []uuid.UUID{mender},
		Colors:            []string{"G"},
	}); err != nil {
		t.Fatalf("activate Food Chain: %v", err)
	}
	if !g.Exile.Contains(mender) {
		t.Fatal("Circuit Mender was not exiled")
	}
	if got := poolColors(me); len(got) != 4 {
		t.Fatalf("pool = %v, want four mana — 1 plus Circuit Mender's mana value 3", got)
	}
	for _, c := range poolColors(me) {
		if c != "G" {
			t.Errorf("pool = %v, want all of the one colour named", poolColors(me))
			break
		}
	}
	if triggerOnStack(g, mender) == nil {
		t.Error("Circuit Mender's leaves-the-battlefield trigger is not on the stack with the mana in the pool")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife {
		t.Error("Zulaport Cutthroat drained for an exiled creature")
	}
}

// A token that copies nothing has mana value 0, so X is 1.
func TestFoodChainOnATokenAddsOne(t *testing.T) {
	g, me, _ := spendTable(t)
	chain := pushCatalogPermanent(g, me.ID, "Food Chain", "Enchantment", foodChainOracle, false)
	token := pushToken(g, me.ID, TokenCard("1/1 green Saproling"))

	if err := g.ActivateManaAbility(me.ID, chain, 0, game.ManaAbilityParams{
		ExilePermanentIDs: []uuid.UUID{token},
		Colors:            []string{"B"},
	}); err != nil {
		t.Fatalf("activate Food Chain: %v", err)
	}
	if got := poolColors(me); len(got) != 1 {
		t.Errorf("pool = %v, want one mana for a mana-value-0 token", got)
	}
}

// "Spend this mana only to cast creature spells": a creature spell, not
// a noncreature spell and not a creature's activated ability (#2059).
func TestFoodChainManaCastsOnlyCreatureSpells(t *testing.T) {
	g, me, _ := spendTable(t)
	chain := pushCatalogPermanent(g, me.ID, "Food Chain", "Enchantment", foodChainOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := g.ActivateManaAbility(me.ID, chain, 0, game.ManaAbilityParams{
		ExilePermanentIDs: []uuid.UUID{bear},
		Colors:            []string{"G"},
	}); err != nil {
		t.Fatalf("activate Food Chain: %v", err)
	}
	cost, _ := game.ParseCost("{G}")
	creature := game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"}
	sorcery := game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(creature)) {
		t.Error("Food Chain mana refused a creature spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(sorcery)) {
		t.Error("Food Chain mana paid for a sorcery")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(creature)) {
		t.Error("Food Chain mana paid for a creature's activated ability")
	}
}

// The auto-tapper never exiles a creature to Food Chain to pay for a
// cast: which creature to give up is the player's decision.
func TestFoodChainIsNeverAutoTapped(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Food Chain", "Enchantment", foodChainOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	spell := handSpell(me, "Green Creature", "Creature — Elf", "{G}")

	refusedForMana(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}),
		"auto-tapping Food Chain")
	if !g.Battlefield.Contains(bear) {
		t.Error("the auto-tapper exiled a creature to Food Chain")
	}
}

// --- City of Shadows -------------------------------------------------

// The banking ability exiles a creature (not a sacrifice) for a storage
// counter, and the mana ability taps for {C} per counter without
// spending them.
func TestCityOfShadowsBanksACreatureAndTapsForItsCounters(t *testing.T) {
	g, me, _ := spendTable(t)
	city := pushCatalogPermanent(g, me.ID, "City of Shadows", "Land", cityOfShadowsOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	if err := g.ActivateManaAbility(me.ID, city, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap with no counters: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v with no storage counters, want nothing", me.ManaPool)
	}
	untap(g, city)

	b16Activate(t, g, me.ID, city, 0, game.ActivateAbilityParams{ExilePermanentIDs: []uuid.UUID{bear}})
	if !g.Exile.Contains(bear) {
		t.Fatal("the creature was not exiled")
	}
	if c, _ := battlefieldCard(g, city); c.Counters[game.CounterStorage] != 1 {
		t.Fatalf("storage counters = %d, want 1", c.Counters[game.CounterStorage])
	}
	untap(g, city)
	if err := g.ActivateManaAbility(me.ID, city, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool = %v, want one {C}", got)
	}
	if c, _ := battlefieldCard(g, city); c.Counters[game.CounterStorage] != 1 {
		t.Error("tapping for mana spent the storage counter")
	}
}

// --- Altar of Bhaal // Bone Offering ---------------------------------

func altarOfBhaal(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   altarOfBhaalOracleTC,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Altar of Bhaal", TypeLine: "Artifact", ManaCost: "{1}{B}", Colors: []string{"B"}},
			{Name: "Bone Offering", TypeLine: "Sorcery — Adventure", ManaCost: "{2}{B}", Colors: []string{"B"}},
		},
	}
	c.SetFace(0)
	return c
}

// The reanimation exiles a creature you control at announce and returns
// the target from your graveyard — the exiled creature never reaches
// the graveyard, so it cannot be its own target.
func TestAltarOfBhaalExilesACreatureToReturnOne(t *testing.T) {
	g, me, _ := spendTable(t)
	altar := altarOfBhaal(me.ID)
	g.Battlefield.PushTop(altar)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	dead := pushGraveyardCardForTest(me, "Dead Bear")
	floatMana(t, g, me, "{C}{C}{B}")

	b16Activate(t, g, me.ID, altar.InstanceID, 0, game.ActivateAbilityParams{
		Strict:            true,
		ExilePermanentIDs: []uuid.UUID{bear},
		Targets:           []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	})
	if !g.Exile.Contains(bear) || me.Graveyard.Contains(bear) {
		t.Error("the creature paid for the cost was not exiled")
	}
	if c, ok := battlefieldCard(g, dead); !ok || c.Controller != me.ID {
		t.Error("the target did not return to the battlefield under your control")
	}
}

// Bone Offering makes a tapped 4/1 black Skeleton with menace, then the
// card goes on an adventure.
func TestBoneOfferingCreatesATappedSkeleton(t *testing.T) {
	g, me, _ := spendTable(t)
	altar := altarOfBhaal(me.ID)
	me.Hand.PushTop(altar)

	if err := g.CastSpell(me.ID, altar.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Bone Offering: %v", err)
	}
	passPriorityAroundTable(t, g)
	skeleton := findBattlefieldByName(g, "Skeleton")
	if skeleton == uuid.Nil {
		t.Fatal("no Skeleton token")
	}
	c, _ := battlefieldCard(g, skeleton)
	if !c.Tapped || c.Power != 4 || c.Toughness != 1 || !hasString(effectiveAbilities(t, g, skeleton), "menace") {
		t.Errorf("Skeleton = tapped %v %d/%d %v, want a tapped 4/1 with menace", c.Tapped, c.Power, c.Toughness, effectiveAbilities(t, g, skeleton))
	}
	if !g.Exile.Contains(altar.InstanceID) {
		t.Error("the card did not go on an adventure")
	}
}
