package testsst

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stx_5502_b45_cards_test.go — batch 45 of Strixhaven: School of Mages
// (#5502): ten magecraft cards. Every test plays the real catalog entry
// through the engine.

const (
	stxB45KarokWranglerOracle         = "031f629f-7a9a-4143-9f5d-434ed1c23b7f"
	stxB45PrismariApprenticeOracle    = "0d60ffba-9904-442e-8d4a-a6b1b50a38eb"
	stxB45SymmetrySageOracle          = "16cb916d-bb19-42ec-9a01-969e9d067b40"
	stxB45LoreholdApprenticeOracle    = "3858a2ff-cdc9-4e40-bc2c-42ed691aae08"
	stxB45PrismariPledgemageOracle    = "47ecedec-7bf7-4a3b-ac33-2bf203fa50d9"
	stxB45SpellSatchelOracle          = "912c5e71-554c-4e31-8831-018760752eb1"
	stxB45DragonsguardEliteOracle     = "9bc966b3-1b6e-4bea-ab1d-321bc03d2bc7"
	stxB45JadziOracle                 = "ca79cd3f-13ba-4d11-b885-48a504ae69aa"
	stxB45ElementalExpressionistOracl = "dfc45c58-821c-42f3-ab48-05b912748bdf"
	stxB45SilverquillPledgemageOracle = "ff72b900-ce3c-4cc4-a50e-e501e17631ca"

	stxB45Bear = "Creature — Bear"
)

// stxB45Spell has p cast a vanilla spell of the given type and mana cost,
// paying it exactly from a floated pool, and returns its card. Magecraft
// sees an instant or a sorcery; a creature spell is the negative.
func stxB45Spell(t *testing.T, g *game.Game, p *game.Player, typeLine, cost string) uuid.UUID {
	t.Helper()
	spec := ""
	for _, r := range cost {
		switch {
		case r == '{' || r == '}':
		case r >= '0' && r <= '9': // generic mana, paid with colorless
			spec += strings.Repeat("C", int(r-'0'))
		default:
			spec += string(r)
		}
	}
	effectstest.FloatForTest(g, p, spec)
	return effectstest.CastFromHandForTest(t, g, p, "Vanilla Spell", typeLine, cost, "", effectstest.Strict)
}

// stxB45Settle passes priority until the stack is empty, answering every
// pick_target prompt addressed to chooser with the first of `prefer` that is
// offered, else the first card offered.
func stxB45Settle(t *testing.T, g *game.Game, chooser uuid.UUID, prefer ...uuid.UUID) {
	t.Helper()
	for i := 0; i < 48; i++ {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			pick := uuid.Nil
			for _, id := range prefer {
				if slices.Contains(p.PickTargetCards, id) {
					pick = id
					break
				}
			}
			if pick == uuid.Nil && len(p.PickTargetCards) > 0 {
				pick = p.PickTargetCards[0]
			}
			if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: pick}); err != nil {
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
	t.Fatal("stack did not settle")
}

func stxB45Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
}

// --- every card is in the catalog ----------------------------------------

func TestStx5502B45EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		stxB45KarokWranglerOracle, stxB45PrismariApprenticeOracle, stxB45SymmetrySageOracle,
		stxB45LoreholdApprenticeOracle, stxB45PrismariPledgemageOracle, stxB45SpellSatchelOracle,
		stxB45DragonsguardEliteOracle, stxB45JadziOracle, stxB45JadziOracle + "#1",
		stxB45ElementalExpressionistOracl, stxB45SilverquillPledgemageOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- Karok Wrangler ---------------------------------------------------------

func TestStx5502B45KarokWranglerPutsACounterOnTargetCreatureYouControl(t *testing.T) {
	g, me, opp := stxB45Table(t)
	effectstest.B12Push(g, me.ID, "Karok Wrangler", "Creature — Elf Druid", stxB45KarokWranglerOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", stxB45Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", stxB45Bear, 2, 2)
	stxB45Spell(t, g, me, "Instant", "{R}")
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the magecraft trigger")
	}
	if slices.Contains(p.PickTargetCards, theirs) {
		t.Error("an opponent's creature was offered: the target is a creature you control")
	}
	stxB45Settle(t, g, me.ID, mine)
	if got := effectstest.PlusOneCounters(g, mine); got != 1 {
		t.Errorf("the target has %d +1/+1 counters, want 1", got)
	}
	if got := effectstest.PlusOneCounters(g, theirs); got != 0 {
		t.Errorf("an opponent's creature has %d counters, want 0", got)
	}
}

func TestStx5502B45KarokWranglerIgnoresACreatureSpell(t *testing.T) {
	g, me, _ := stxB45Table(t)
	effectstest.B12Push(g, me.ID, "Karok Wrangler", "Creature — Elf Druid", stxB45KarokWranglerOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", stxB45Bear, 2, 2)
	stxB45Spell(t, g, me, stxB45Bear, "{G}")
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a creature spell triggered magecraft")
	}
	stxB45Settle(t, g, me.ID, mine)
	if got := effectstest.PlusOneCounters(g, mine); got != 0 {
		t.Errorf("%d counters after a creature spell, want 0", got)
	}
}

// --- Symmetry Sage ----------------------------------------------------------

func TestStx5502B45SymmetrySageSetsBasePowerTwoUntilEndOfTurn(t *testing.T) {
	g, me, opp := stxB45Table(t)
	sage := effectstest.B12Push(g, me.ID, "Symmetry Sage", "Creature — Human Wizard", stxB45SymmetrySageOracle, 0, 2)
	if !effectstest.HasEffectiveKeyword(t, g, sage, "flying") {
		t.Error("Symmetry Sage has no flying")
	}
	brute := effectstest.B12Creature(g, me.ID, "Brute", stxB45Bear, 5, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", stxB45Bear, 4, 4)
	effectstest.AddCounters(t, g, brute, game.CounterPlusOne, 1)
	stxB45Spell(t, g, me, "Instant", "{R}")
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || slices.Contains(p.PickTargetCards, theirs) {
		t.Fatalf("want a target prompt that offers only my creatures, got %+v", p)
	}
	stxB45Settle(t, g, me.ID, brute)
	// Base power 2, plus the counter: 3 (the +1 counter is applied after the base set).
	if pw, tough := effectstest.PtOf(t, g, brute); pw != 3 || tough != 6 {
		t.Errorf("the target is %d/%d, want 3/6 (base power 2 plus a counter)", pw, tough)
	}
	if pw, _ := effectstest.PtOf(t, g, theirs); pw != 4 {
		t.Errorf("an opponent's creature has power %d, want 4", pw)
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	if pw, _ := effectstest.PtOf(t, g, brute); pw != 6 {
		t.Errorf("after the turn the target has power %d, want 6", pw)
	}
}

func TestStx5502B45SymmetrySageIgnoresACreatureSpell(t *testing.T) {
	g, me, _ := stxB45Table(t)
	effectstest.B12Push(g, me.ID, "Symmetry Sage", "Creature — Human Wizard", stxB45SymmetrySageOracle, 0, 2)
	brute := effectstest.B12Creature(g, me.ID, "Brute", stxB45Bear, 5, 5)
	stxB45Spell(t, g, me, stxB45Bear, "{G}")
	stxB45Settle(t, g, me.ID, brute)
	if pw, _ := effectstest.PtOf(t, g, brute); pw != 5 {
		t.Errorf("power %d after a creature spell, want 5", pw)
	}
}

// --- Prismari Pledgemage ----------------------------------------------------

func TestStx5502B45PrismariPledgemageAttacksAfterMagecraft(t *testing.T) {
	g, me, opp := stxB45Table(t)
	mage := effectstest.B12Push(g, me.ID, "Prismari Pledgemage", "Creature — Orc Wizard", stxB45PrismariPledgemageOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, mage, "defender") {
		t.Fatal("Prismari Pledgemage has no defender")
	}
	stxB45Spell(t, g, me, "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := effectstest.DeclareResult(g, mage, opp.ID); err != nil {
		t.Errorf("attack after magecraft: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mage, "defender") {
		t.Error("it still has defender")
	}
}

func TestStx5502B45PrismariPledgemageCantAttackWithoutMagecraft(t *testing.T) {
	g, me, opp := stxB45Table(t)
	mage := effectstest.B12Push(g, me.ID, "Prismari Pledgemage", "Creature — Orc Wizard", stxB45PrismariPledgemageOracle, 3, 3)
	stxB45Spell(t, g, me, stxB45Bear, "{G}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.WantDefenderRefused(t, g, mage, opp.ID)
}

func TestStx5502B45PrismariPledgemageWaiverEndsWithTheTurn(t *testing.T) {
	g, me, opp := stxB45Table(t)
	mage := effectstest.B12Push(g, me.ID, "Prismari Pledgemage", "Creature — Orc Wizard", stxB45PrismariPledgemageOracle, 3, 3)
	stxB45Spell(t, g, me, "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	effectstest.WantDefenderRefused(t, g, mage, opp.ID)
}

// --- Prismari Apprentice ----------------------------------------------------

func stxB45ApprenticeAttack(t *testing.T, cost, typeLine string) (counters int, blockErr error) {
	t.Helper()
	g, me, opp := stxB45Table(t)
	app := effectstest.B12Push(g, me.ID, "Prismari Apprentice", "Creature — Human Shaman", stxB45PrismariApprenticeOracle, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", stxB45Bear, 2, 2)
	stxB45Spell(t, g, me, typeLine, cost)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BrAttack(t, g, app)
	return effectstest.PlusOneCounters(g, app), g.DeclareBlocker(blocker, app)
}

func TestStx5502B45PrismariApprenticeCantBeBlockedAndGrowsOnABigSpell(t *testing.T) {
	counters, err := stxB45ApprenticeAttack(t, "{4}{R}", "Sorcery")
	if err == nil {
		t.Error("the unblockable Apprentice was blocked")
	}
	if counters != 1 {
		t.Errorf("%d counters after a mana value 5 spell, want 1", counters)
	}
}

func TestStx5502B45PrismariApprenticeSmallSpellIsUnblockableWithoutACounter(t *testing.T) {
	counters, err := stxB45ApprenticeAttack(t, "{3}{R}", "Instant")
	if err == nil {
		t.Error("the unblockable Apprentice was blocked")
	}
	if counters != 0 {
		t.Errorf("%d counters after a mana value 4 spell, want 0", counters)
	}
}

func TestStx5502B45PrismariApprenticeCanBeBlockedWithoutMagecraft(t *testing.T) {
	counters, err := stxB45ApprenticeAttack(t, "{5}{G}", stxB45Bear)
	if err != nil {
		t.Errorf("a creature spell made the Apprentice unblockable: %v", err)
	}
	if counters != 0 {
		t.Errorf("%d counters after a creature spell, want 0", counters)
	}
}

// --- Dragonsguard Elite -----------------------------------------------------

func TestStx5502B45DragonsguardEliteGrowsAndDoubles(t *testing.T) {
	g, me, _ := stxB45Table(t)
	elite := effectstest.B12Push(g, me.ID, "Dragonsguard Elite", "Creature — Human Druid", stxB45DragonsguardEliteOracle, 2, 2)
	stxB45Spell(t, g, me, "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	stxB45Spell(t, g, me, "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	stxB45Spell(t, g, me, "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, elite); got != 3 {
		t.Fatalf("%d counters after three instants, want 3", got)
	}
	effectstest.FloatForTest(g, me, "GGCCCC")
	effectstest.ApaActivate(t, g, me, elite, 0, game.ActivateAbilityParams{})
	if got := effectstest.PlusOneCounters(g, elite); got != 6 {
		t.Errorf("%d counters after doubling three, want 6", got)
	}
	if pw, tough := effectstest.PtOf(t, g, elite); pw != 8 || tough != 8 {
		t.Errorf("the Elite is %d/%d, want 8/8", pw, tough)
	}
}

func TestStx5502B45DragonsguardEliteDoublingNothingAddsNothing(t *testing.T) {
	g, me, _ := stxB45Table(t)
	elite := effectstest.B12Push(g, me.ID, "Dragonsguard Elite", "Creature — Human Druid", stxB45DragonsguardEliteOracle, 2, 2)
	stxB45Spell(t, g, me, stxB45Bear, "{G}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, elite); got != 0 {
		t.Fatalf("%d counters after a creature spell, want 0", got)
	}
	effectstest.FloatForTest(g, me, "GGCCCC")
	effectstest.ApaActivate(t, g, me, elite, 0, game.ActivateAbilityParams{})
	if got := effectstest.PlusOneCounters(g, elite); got != 0 {
		t.Errorf("%d counters after doubling none, want 0", got)
	}
	// Short of mana, the activation is refused.
	if err := g.ActivateCatalogAbility(me.ID, elite, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the doubling was activated with no mana")
	}
}

// --- Spell Satchel ----------------------------------------------------------

func TestStx5502B45SpellSatchelStoresBooksAndSpendsThem(t *testing.T) {
	g, me, _ := stxB45Table(t)
	satchel := effectstest.PushCatalogPermanent(g, me.ID, "Spell Satchel", "Artifact", stxB45SpellSatchelOracle, false)
	if err := g.ActivateManaAbility(me.ID, satchel, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("a Satchel with no book counter made mana")
	}
	stxB45Spell(t, g, me, stxB45Bear, "{G}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, satchel, "book"); got != 0 {
		t.Fatalf("%d book counters after a creature spell, want 0", got)
	}
	stxB45Spell(t, g, me, "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, satchel, "book"); got != 1 {
		t.Fatalf("%d book counters after an instant, want 1", got)
	}
	if err := g.ActivateManaAbility(me.ID, satchel, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping the Satchel for mana: %v", err)
	}
	if got := effectstest.CountersOn(g, satchel, "book"); got != 0 {
		t.Errorf("%d book counters after the mana ability, want 0", got)
	}
	if n := me.ManaPool.Total(); n != 1 {
		t.Errorf("%d mana in the pool, want {C}", n)
	}
}

func TestStx5502B45SpellSatchelDrawsForThreeBooks(t *testing.T) {
	g, me, _ := stxB45Table(t)
	satchel := effectstest.PushCatalogPermanent(g, me.ID, "Spell Satchel", "Artifact", stxB45SpellSatchelOracle, false)
	effectstest.AddCounters(t, g, satchel, "book", 2)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, satchel, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("two book counters paid for the draw")
	}
	if got := effectstest.CountersOn(g, satchel, "book"); got != 2 {
		t.Fatalf("a refused activation spent counters: %d left", got)
	}
	effectstest.AddCounters(t, g, satchel, "book", 1)
	hand := me.Hand.Size()
	effectstest.ApaActivate(t, g, me, satchel, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
	if got := effectstest.CountersOn(g, satchel, "book"); got != 0 {
		t.Errorf("%d book counters left, want 0", got)
	}
}

// --- Silverquill Pledgemage -------------------------------------------------

func TestStx5502B45SilverquillPledgemageGainsTheChosenKeyword(t *testing.T) {
	for i, kw := range []string{"flying", "lifelink"} {
		other := []string{"lifelink", "flying"}[i]
		g, me, _ := stxB45Table(t)
		mage := effectstest.B12Push(g, me.ID, "Silverquill Pledgemage", "Creature — Vampire Cleric", stxB45SilverquillPledgemageOracle, 3, 1)
		stxB45Spell(t, g, me, "Instant", "{R}")
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerOptionPick(t, g, me.ID, i)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, mage, kw) {
			t.Errorf("choice %d: no %s", i, kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, mage, other) {
			t.Errorf("choice %d: it also has %s", i, other)
		}
		effectstest.B12ToMyNextUpkeep(t, g)
		if effectstest.HasEffectiveKeyword(t, g, mage, kw) {
			t.Errorf("choice %d: %s outlasted the turn", i, kw)
		}
	}
}

func TestStx5502B45SilverquillPledgemageIgnoresACreatureSpell(t *testing.T) {
	g, me, _ := stxB45Table(t)
	mage := effectstest.B12Push(g, me.ID, "Silverquill Pledgemage", "Creature — Vampire Cleric", stxB45SilverquillPledgemageOracle, 3, 1)
	stxB45Spell(t, g, me, stxB45Bear, "{G}")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("a creature spell asked for a keyword")
	}
	if effectstest.HasEffectiveKeyword(t, g, mage, "flying") || effectstest.HasEffectiveKeyword(t, g, mage, "lifelink") {
		t.Error("it gained a keyword from a creature spell")
	}
}

// --- Lorehold Apprentice ----------------------------------------------------

func TestStx5502B45LoreholdApprenticeGivesSpiritsAPing(t *testing.T) {
	g, me, opp := stxB45Table(t)
	effectstest.B12Push(g, me.ID, "Lorehold Apprentice", "Creature — Human Cleric", stxB45LoreholdApprenticeOracle, 2, 2)
	spirit := effectstest.B12Creature(g, me.ID, "My Spirit", "Creature — Spirit", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", stxB45Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Spirit", "Creature — Spirit", 1, 1)
	stxB45Spell(t, g, me, "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, spirit)
	if idx < 0 {
		t.Fatal("my Spirit did not gain the ping")
	}
	if i, _ := effectstest.GrantedActivatedIndex(t, g, bear); i >= 0 {
		t.Error("a non-Spirit gained the ping")
	}
	if i, _ := effectstest.GrantedActivatedIndex(t, g, theirs); i >= 0 {
		t.Error("an opponent's Spirit gained the ping")
	}
	var lives []int
	for _, s := range g.Seats {
		lives = append(lives, s.Life)
	}
	effectstest.B16Activate(t, g, me.ID, spirit, idx, game.ActivateAbilityParams{Ref: ref})
	effectstest.PassPriorityAroundTable(t, g)
	for i, s := range g.Seats {
		want := lives[i] - 1
		if s.ID == me.ID {
			want = lives[i]
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
	if c, ok := effectstest.BattlefieldCard(g, spirit); !ok || !c.Tapped {
		t.Error("the Spirit did not tap for the ping")
	}
	// Until end of turn: gone by my next turn.
	effectstest.B12ToMyNextUpkeep(t, g)
	if i, _ := effectstest.GrantedActivatedIndex(t, g, spirit); i >= 0 {
		t.Error("the ping outlasted the turn")
	}
}

func TestStx5502B45LoreholdApprenticeSpiritThatArrivesLaterHasNoPing(t *testing.T) {
	g, me, _ := stxB45Table(t)
	effectstest.B12Push(g, me.ID, "Lorehold Apprentice", "Creature — Human Cleric", stxB45LoreholdApprenticeOracle, 2, 2)
	stxB45Spell(t, g, me, stxB45Bear, "{G}")
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.B12Creature(g, me.ID, "Late Spirit", "Creature — Spirit", 1, 1)
	if i, _ := effectstest.GrantedActivatedIndex(t, g, late); i >= 0 {
		t.Error("a Spirit gained the ping from a creature spell")
	}
	stxB45Spell(t, g, me, "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	after := effectstest.B12Creature(g, me.ID, "Later Spirit", "Creature — Spirit", 1, 1)
	if i, _ := effectstest.GrantedActivatedIndex(t, g, late); i < 0 {
		t.Error("the Spirit that was there did not gain the ping")
	}
	if i, _ := effectstest.GrantedActivatedIndex(t, g, after); i >= 0 {
		t.Error("a Spirit that arrived after the trigger resolved gained the ping (CR 611.2c)")
	}
}

// --- Elemental Expressionist ------------------------------------------------

func stxB45Expressionist(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g, me, opp := stxB45Table(t)
	effectstest.B12Push(g, me.ID, "Elemental Expressionist", "Creature — Orc Wizard", stxB45ElementalExpressionistOracl, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", stxB45Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", stxB45Bear, 2, 2)
	return g, me, bear, theirs
}

func stxB45Elementals(g *game.Game, me *game.Player) int {
	n := 0
	for _, id := range effectstest.BattlefieldIDsNamed(g, "Elemental") {
		if c := effectstest.CardByID(g, id); c.Controller == me.ID && c.Power == 4 && c.Toughness == 4 {
			n++
		}
	}
	return n
}

func TestStx5502B45ElementalExpressionistExilesInsteadAndMakesAnElemental(t *testing.T) {
	g, me, bear, theirs := stxB45Expressionist(t)
	stxB45Spell(t, g, me, "Instant", "{R}")
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || slices.Contains(p.PickTargetCards, theirs) {
		t.Fatalf("want a target prompt that offers only my creatures, got %+v", p)
	}
	stxB45Settle(t, g, me.ID, bear)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, bear); z != game.ZoneExile {
		t.Errorf("the destroyed creature went to %q, want exile", z)
	}
	if got := stxB45Elementals(g, me); got != 1 {
		t.Errorf("%d 4/4 Elementals, want 1", got)
	}
	c := effectstest.CardByID(g, effectstest.BattlefieldIDsNamed(g, "Elemental")[0])
	if !c.IsToken() || len(c.Colors) != 2 {
		t.Errorf("Elemental = %+v, want a blue and red token", c)
	}
}

func TestStx5502B45ElementalExpressionistEachInstanceTriggersSeparately(t *testing.T) {
	g, me, bear, _ := stxB45Expressionist(t)
	stxB45Spell(t, g, me, "Instant", "{R}")
	stxB45Settle(t, g, me.ID, bear)
	stxB45Spell(t, g, me, "Instant", "{R}")
	stxB45Settle(t, g, me.ID, bear)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	// Two identical exile-instead effects apply to one event: the controller
	// orders them (CR 616.1), and either order exiles it once.
	if c := effectstest.B06ReplacementOrderFor(g, me.ID); c != nil {
		if err := g.ResolveReplacementOrder(c.ID, me.ID, c.ReplacementEffectIDs); err != nil {
			t.Fatalf("ordering the replacements: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, bear); z != game.ZoneExile {
		t.Errorf("the creature went to %q, want exile", z)
	}
	if got := stxB45Elementals(g, me); got != 2 {
		t.Errorf("%d 4/4 Elementals after two magecraft triggers, want 2", got)
	}
}

func TestStx5502B45ElementalExpressionistDoesNothingToACreatureThatWasNotChosen(t *testing.T) {
	g, me, bear, _ := stxB45Expressionist(t)
	other := effectstest.B12Creature(g, me.ID, "Other", stxB45Bear, 2, 2)
	stxB45Spell(t, g, me, "Instant", "{R}")
	stxB45Settle(t, g, me.ID, bear)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(other); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, other); z != game.ZoneGraveyard {
		t.Errorf("an unchosen creature went to %q, want the graveyard", z)
	}
	if got := stxB45Elementals(g, me); got != 0 {
		t.Errorf("%d Elementals, want 0", got)
	}
}

func TestStx5502B45ElementalExpressionistLastsOnlyThisTurn(t *testing.T) {
	g, me, bear, _ := stxB45Expressionist(t)
	stxB45Spell(t, g, me, "Instant", "{R}")
	stxB45Settle(t, g, me.ID, bear)
	effectstest.B12ToMyNextUpkeep(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, bear); z != game.ZoneGraveyard {
		t.Errorf("next turn the creature went to %q, want the graveyard", z)
	}
	if got := stxB45Elementals(g, me); got != 0 {
		t.Errorf("%d Elementals, want 0", got)
	}
}

// --- Jadzi, Oracle of Arcavios // Journey to the Oracle ---------------------

func stxB45JadziCard(owner uuid.UUID) game.Card {
	return effectstest.MdfcCard(owner, stxB45JadziOracle, game.LayoutModalDFC,
		game.Face{Name: "Jadzi, Oracle of Arcavios", TypeLine: "Legendary Creature — Human Wizard", ManaCost: "{6}{U}{U}",
			Power: 5, Toughness: 5, Colors: []string{"U"}},
		game.Face{Name: "Journey to the Oracle", TypeLine: "Sorcery", ManaCost: "{2}{G}{G}", Colors: []string{"G"}},
	)
}

func stxB45JadziOnBattlefield(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := stxB45JadziCard(owner)
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	return c.InstanceID
}

func TestStx5502B45JadziDiscardsToReturnToHand(t *testing.T) {
	g, me, _ := stxB45Table(t)
	jadzi := stxB45JadziOnBattlefield(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, jadzi, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Jadzi was returned without discarding a card")
	}
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, jadzi, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("discard to return Jadzi: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(jadzi) || !me.Hand.Contains(jadzi) {
		t.Error("Jadzi did not return to its owner's hand")
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discarded card is not in the graveyard")
	}
}

// jadziMagecraft seeds `top` on the library, puts Jadzi on the battlefield
// and casts a vanilla instant.
func stxB45JadziMagecraft(t *testing.T, top game.Card) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g, me, _ := stxB45Table(t)
	jadzi := stxB45JadziOnBattlefield(g, me.ID)
	id := effectstest.SeedLibraryTop(me, top.Name, top.TypeLine, top.ManaCost)
	for i := range me.Library.Cards {
		if me.Library.Cards[i].InstanceID == id {
			me.Library.Cards[i].Power, me.Library.Cards[i].Toughness = top.Power, top.Toughness
		}
	}
	stxB45Spell(t, g, me, "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, jadzi, id
}

func TestStx5502B45JadziCastsTheRevealedCardForOneMana(t *testing.T) {
	g, me, _, top := stxB45JadziMagecraft(t, game.Card{Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{G}{G}", Power: 6, Toughness: 6})
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, top)
	// It costs {1}, not its mana cost and not nothing: with an empty pool the cast is refused.
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, top, effectstest.Strict); err == nil {
		t.Fatal("the revealed card was cast for nothing")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, top, effectstest.Strict); err != nil {
		t.Fatalf("casting the revealed card for {1}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, top) != game.ZoneBattlefield {
		t.Errorf("the cast creature is in %v, want the battlefield", effectstest.B12ZoneOf(g, top))
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left over, want 0", n)
	}
}

func TestStx5502B45JadziDecliningLeavesTheCardOnTop(t *testing.T) {
	g, me, _, top := stxB45JadziMagecraft(t, game.Card{Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{G}{G}", Power: 6, Toughness: 6})
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Library.Cards) == 0 || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != top {
		// The library's top is its last element (PushTop); a card not cast stays there.
		if z := effectstest.B12ZoneOf(g, top); z != game.ZoneLibrary {
			t.Errorf("the declined card is in %v, want the library", z)
		}
	}
}

func TestStx5502B45JadziPutsARevealedLandOntoTheBattlefield(t *testing.T) {
	g, me, _, top := stxB45JadziMagecraft(t, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a land was offered as a cast")
	}
	if !g.Battlefield.Contains(top) {
		t.Errorf("the revealed land is in %v, want the battlefield", effectstest.B12ZoneOf(g, top))
	}
	if g.LandsPlayedThisTurnFor(me.ID) != 0 {
		t.Error("the put used the land drop")
	}
}

func TestStx5502B45JadziIgnoresACreatureSpell(t *testing.T) {
	g, me, _ := stxB45Table(t)
	stxB45JadziOnBattlefield(g, me.ID)
	top := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	stxB45Spell(t, g, me, stxB45Bear, "{G}")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, top) != game.ZoneLibrary {
		t.Errorf("a creature spell triggered magecraft: the top card is in %v", effectstest.B12ZoneOf(g, top))
	}
}

// Journey to the Oracle, cast from the back face of the card in hand.
func stxB45Journey(t *testing.T, onBattlefield, inHand int) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID, uuid.UUID) {
	t.Helper()
	g, me, _ := stxB45Table(t)
	for i := 0; i < onBattlefield; i++ {
		effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	var hand []uuid.UUID
	for i := 0; i < inHand; i++ {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		hand = append(hand, id)
	}
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	card := stxB45JadziCard(me.ID)
	card.SetFace(1)
	me.Hand.PushTop(card)
	effectstest.FloatForTest(g, me, "CCGG")
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Strict: true}); err != nil {
		t.Fatalf("cast Journey to the Oracle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, card.InstanceID, hand, pitch
}

func TestStx5502B45JourneyReturnsToHandWithEightLandsAndADiscard(t *testing.T) {
	g, me, journey, hand, pitch := stxB45Journey(t, 5, 3)
	effectstest.AnswerChooseCards(t, g, me.ID, hand...)
	for _, id := range hand {
		if !g.Battlefield.Contains(id) {
			t.Fatalf("a land from hand did not enter")
		}
	}
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("eight lands and no discard offer")
	}
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(journey) {
		t.Errorf("Journey to the Oracle is in %v, want its owner's hand", effectstest.B12ZoneOf(g, journey))
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discarded card is not in the graveyard")
	}
}

func TestStx5502B45JourneyStaysInTheGraveyardWithSevenLands(t *testing.T) {
	g, me, journey, hand, _ := stxB45Journey(t, 5, 3)
	effectstest.AnswerChooseCards(t, g, me.ID, hand[:2]...)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("a discard was offered with seven lands")
	}
	if !me.Graveyard.Contains(journey) {
		t.Errorf("Journey to the Oracle is in %v, want the graveyard", effectstest.B12ZoneOf(g, journey))
	}
}

func TestStx5502B45JourneyDecliningTheDiscardLeavesItInTheGraveyard(t *testing.T) {
	g, me, journey, hand, _ := stxB45Journey(t, 5, 3)
	effectstest.AnswerChooseCards(t, g, me.ID, hand...)
	pick := effectstest.DiscardChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no discard offer with eight lands")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining the discard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(journey) {
		t.Errorf("Journey to the Oracle is in %v, want the graveyard", effectstest.B12ZoneOf(g, journey))
	}
}
