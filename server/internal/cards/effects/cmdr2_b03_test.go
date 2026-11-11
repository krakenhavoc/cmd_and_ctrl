package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b03_test.go — Commander-deck wave 2, batch 3 (#5368): simple
// creatures and small permanents.

const (
	cmdr2ShadowmageOracle   = "3a6f886b-2043-47e9-9c0f-f7913a6fa67d"
	cmdr2SlitherBladeOracle = "5a55641b-0818-4418-a2e5-44157056353c"
	cmdr2AbominationOracle  = "139af10d-a3e5-4ade-b876-87324e6c8017"
	cmdr2SkulkOracle        = "490df3b5-ece2-45bb-af38-fcd7f83e37b0"
	cmdr2RestorationOracle  = "dfbd3afc-9905-4cff-a4f4-df08a4d0a7fa"
	cmdr2ChannelerOracle    = "fb220f46-f8b8-4804-baa4-e7d50b4871f7"
	cmdr2TitheTakerOracle   = "c916b81f-700e-4a1a-8c8e-c4685ceaecd2"
	cmdr2BairdOracle        = "69aa6809-5468-413a-94c9-562e6de93957"
	cmdr2BladegriffOracle   = "9fff90b6-98ef-4f78-b08d-7e46dfc03403"
	cmdr2SoulSnareOracle    = "52bd4331-67b6-4eb8-9566-e00359889fc6"
	cmdr2SandwurmOracle     = "429eff2b-635b-43f6-b20f-82e25992c0b9"
)

func TestCmdr2B03CardsAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{
		cmdr2ShadowmageOracle, cmdr2SlitherBladeOracle, cmdr2AbominationOracle, cmdr2SkulkOracle,
		cmdr2RestorationOracle, cmdr2ChannelerOracle, cmdr2TitheTakerOracle, cmdr2BairdOracle,
		cmdr2BladegriffOracle, cmdr2SoulSnareOracle, cmdr2SandwurmOracle,
	} {
		spec, ok := Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v, want Full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Shadowmage Infiltrator -----------------------------------------------

func TestCmdr2B03ShadowmageHasFearAndDrawsOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sm := b12Push(g, me.ID, "Shadowmage Infiltrator", "Creature — Human Wizard", cmdr2ShadowmageOracle, 1, 3)
	if !hasKeywordOnBattlefield(t, g, sm, "fear") {
		t.Error("Shadowmage Infiltrator has no fear")
	}
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, sm, opp.ID, 1)
	assistCardSettle(t, g, me.ID, nil, true)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestCmdr2B03ShadowmageDrawIsOptional(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sm := b12Push(g, me.ID, "Shadowmage Infiltrator", "Creature — Human Wizard", cmdr2ShadowmageOracle, 1, 3)
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, sm, opp.ID, 1)
	assistCardSettle(t, g, me.ID, nil, false)
	if me.Hand.Size() != hand {
		t.Error("declining still drew")
	}
}

// --- Slither Blade -----------------------------------------------------------

func TestCmdr2B03SlitherBladeCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sb := b12Push(g, me.ID, "Slither Blade", "Creature — Snake Rogue", cmdr2SlitherBladeOracle, 1, 2)
	assertRestrictions(t, g, sb, game.CantBeBlocked)
}

// --- Twisted Abomination -----------------------------------------------------

func TestCmdr2B03TwistedAbominationRegeneratesForB(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ta := b12Push(g, me.ID, "Twisted Abomination", "Creature — Zombie Mutant", cmdr2AbominationOracle, 5, 3)
	if err := g.ActivateCatalogAbility(me.ID, ta, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("regenerated with no mana")
	}
	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, ta, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := regenShieldsOn(t, g, ta); n != 1 {
		t.Fatalf("shields = %d, want 1", n)
	}
	destroyIt(t, g, ta)
	if c, ok := battlefieldCard(g, ta); !ok || !c.Tapped {
		t.Fatal("the shield should have saved it, tapped")
	}
}

func TestCmdr2B03TwistedAbominationSwampcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swamp := pushLibraryCardForTest(me, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	pushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	p := me
	advanceToMain(t, g)
	cid := pushCatalogHandCard(p, "Twisted Abomination", "Creature — Zombie Mutant", cmdr2AbominationOracle)
	if err := g.AddManaForEffect(p.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(p.ID, cid, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("swampcycle: %v", err)
	}
	if !p.Graveyard.Contains(cid) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, p.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != swamp {
		t.Fatalf("want a search offering only the Swamp, got %+v", c)
	}
	if err := g.ResolveSearchLibrary(c.ID, p.ID, []uuid.UUID{swamp}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !p.Hand.Contains(swamp) {
		t.Error("swampcycling did not fetch the Swamp")
	}
}

// --- Mistmeadow Skulk --------------------------------------------------------

func TestCmdr2B03MistmeadowSkulkHasLifelinkAndProtectionFromBigThings(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	skulk := b12Push(g, opp.ID, "Mistmeadow Skulk", "Creature — Kithkin Rogue", cmdr2SkulkOracle, 1, 1)
	if !hasKeywordOnBattlefield(t, g, skulk, "lifelink") {
		t.Error("no lifelink")
	}
	deal := func(typeCost string, mv int) int {
		src := b12Creature(g, me.ID, "Source "+typeCost, "Creature — Test", 2, 2)
		g.WithWriteLock(func() {
			c := findBattlefieldCardByID(g, src)
			c.ManaCost = typeCost
		})
		before := cmdr2Damage(g, skulk)
		g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, skulk, 1) })
		return cmdr2Damage(g, skulk) - before
	}
	if got := deal("{1}{G}", 2); got != 1 {
		t.Errorf("a mana value 2 source dealt %d, want 1", got)
	}
	for _, cost := range []string{"{2}{G}", "{3}{G}"} {
		if got := deal(cost, 3); got != 0 {
			t.Errorf("a %s source dealt %d, want 0 (protection from mana value 3 or greater)", cost, got)
		}
	}
}

func cmdr2Damage(g *game.Game, id uuid.UUID) int {
	c, ok := battlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.DamageMarked
}

// --- Restoration Angel ---------------------------------------------------------

func TestCmdr2B03RestorationAngelBlinksANonAngelCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castAndResolveCreature(t, g, "Restoration Angel", "Creature — Angel", cmdr2RestorationOracle)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	// It left and came back as a new object (CR 400.7): the same card, new and summoning sick.
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Bear" {
			n++
			if !c.SummonedThisTurn || c.Controller != me.ID {
				t.Errorf("the bear did not re-enter under my control: %+v", c.SummonedThisTurn)
			}
		}
	}
	if n != 1 {
		t.Fatalf("%d bears on the battlefield, want 1", n)
	}
}

func TestCmdr2B03RestorationAngelBlinkIsOptional(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castAndResolveCreature(t, g, "Restoration Angel", "Creature — Angel", cmdr2RestorationOracle)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, false)
	if c, ok := battlefieldCard(g, bear); !ok || c.SummonedThisTurn {
		t.Error("declining still blinked the bear")
	}
}

func TestCmdr2B03RestorationAngelCannotTargetAnAngel(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	angel := b12Creature(g, me.ID, "Other Angel", "Creature — Angel", 2, 2)
	castAndResolveCreature(t, g, "Restoration Angel", "Creature — Angel", cmdr2RestorationOracle)
	sawPrompt := false
	for i := 0; i < 8; i++ {
		if p := latestPickTarget(g, me.ID); p != nil {
			if hasID(p.PickTargetCards, angel) {
				t.Fatalf("an Angel was offered as a target")
			}
			sawPrompt = true
			break
		}
		if stackFullyEmpty(g) {
			break
		}
		passPriorityAroundTable(t, g)
	}
	_ = sawPrompt // with no legal target the trigger is simply removed (CR 603.3d)
}

// --- Aether Channeler ----------------------------------------------------------

func TestCmdr2B03AetherChannelerModes(t *testing.T) {
	t.Run("bird", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		cmdr2ChannelerMode(t, g, me, 0)
		found := false
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Name == "Bird" && c.Controller == me.ID && c.IsCreature() {
				found = true
			}
		}
		if !found {
			t.Error("no 1/1 Bird token")
		}
	})
	t.Run("draw", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		hand := me.Hand.Size()
		cmdr2ChannelerMode(t, g, me, 2)
		if got := me.Hand.Size() - hand; got != 1 {
			t.Errorf("hand grew by %d, want 1", got)
		}
	})
	t.Run("bounce", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
		land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
		cmdr2ChannelerMode(t, g, me, 1)
		pickCard(t, g, me.ID, rock)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
			t.Error("the rock was not returned to its owner's hand")
		}
		if !g.Battlefield.Contains(land) {
			t.Error("the land left")
		}
	})
}

// cmdr2ChannelerMode casts Aether Channeler and answers its ETB mode prompt.
func cmdr2ChannelerMode(t *testing.T, g *game.Game, me *game.Player, mode int) {
	t.Helper()
	castAndResolveCreature(t, g, "Aether Channeler", "Creature — Human Wizard", cmdr2ChannelerOracle)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt from the ETB: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if mode != 1 {
		passPriorityAroundTable(t, g)
	}
}

// --- Tithe Taker ---------------------------------------------------------------

func TestCmdr2B03TitheTakerTaxesOpponentsSpellsOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Tithe Taker", "Creature — Human Soldier", cmdr2TitheTakerOracle, false)
	if got := priceInHand(t, g, opp, "Their Bolt", "Instant", "{R}"); got != 2 {
		t.Errorf("an opponent's spell on my turn costs %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "My Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("my own spell costs %d, want 1", got)
	}
}

func TestCmdr2B03TitheTakerDoesNothingOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "Tithe Taker", "Creature — Human Soldier", cmdr2TitheTakerOracle, false)
	// It is seat 0's turn, so seat 1's Tithe Taker is not "your turn".
	if got := priceInHand(t, g, me, "My Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("a spell on its caster's own turn costs %d, want 1", got)
	}
}

func TestCmdr2B03TitheTakerTaxesOpponentsAbilitiesButNotManaAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Tithe Taker", "Creature — Human Soldier", cmdr2TitheTakerOracle, false)
	snare := pushCatalogPermanent(g, opp.ID, "Soul Snare", "Enchantment", cmdr2SoulSnareOracle, false)
	mine := pushCatalogPermanent(g, me.ID, "Soul Snare", "Enchantment", cmdr2SoulSnareOracle, false)
	price := func(activator uuid.UUID, id uuid.UUID) int {
		var q game.ParsedCost
		var err error
		g.ReadSnapshot(func() {
			c, _ := g.LookupCardForEffect(id)
			q, err = g.AbilityManaCostForEffect(activator, c, game.ZoneBattlefield, game.ActivatedAbilitiesForCard(c)[0])
		})
		if err != nil {
			t.Fatalf("AbilityManaCostForEffect: %v", err)
		}
		return q.ManaValue()
	}
	if got := price(opp.ID, snare); got != 2 {
		t.Errorf("an opponent's ability costs %d, want 2", got)
	}
	if got := price(me.ID, mine); got != 1 {
		t.Errorf("my own ability costs %d, want 1", got)
	}
}

func TestCmdr2B03TitheTakerAfterlifeMakesASpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tt := pushCatalogPermanent(g, me.ID, "Tithe Taker", "Creature — Human Soldier", cmdr2TitheTakerOracle, false)
	destroyIt(t, g, tt)
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Spirit" && c.Controller == me.ID && c.Power == 1 && c.Toughness == 1 && len(c.Colors) == 2 {
			found = true
		}
	}
	if !found {
		t.Error("no 1/1 white and black Spirit token")
	}
}

// --- Baird, Steward of Argive --------------------------------------------------

func TestCmdr2B03BairdTaxesEachAttackerOneAndCoversPlaneswalkers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	pushCatalogPermanent(g, opp.ID, "Baird, Steward of Argive", "Legendary Creature — Human Soldier", cmdr2BairdOracle, false)
	pw := pushWalkerForTest(g, opp.ID, "Their Walker", "", 3)
	p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: a, Target: opp.ID}, {Attacker: b, Target: opp.ID}})
	if p.Cost != "{1}{1}" {
		t.Errorf("two attackers cost %q, want {1}{1}", p.Cost)
	}
	p = g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: a, Target: pw}})
	if p.Cost != "{1}" {
		t.Errorf("an attack on their planeswalker costs %q, want {1}", p.Cost)
	}
	p = g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: a, Target: g.Seats[2].ID}})
	if p.Cost != "" {
		t.Errorf("an attack on another player costs %q, want nothing", p.Cost)
	}
}

// --- Sandwurm Convergence ------------------------------------------------------

func TestCmdr2B03SandwurmConvergenceStopsFliersAttackingYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := b12Creature(g, me.ID, "Flier", "Creature — Bird", 1, 1)
	g.WithWriteLock(func() { findBattlefieldCardByID(g, flier).Keywords = []string{"flying"} })
	walker := pushVanillaCreature(g, me.ID, "Walker", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	pushCatalogPermanent(g, opp.ID, "Sandwurm Convergence", "Enchantment", cmdr2SandwurmOracle, false)
	pw := pushWalkerForTest(g, opp.ID, "Their Walker", "", 3)
	if err := g.DeclareAttacker(flier, opp.ID); err == nil {
		t.Error("a flier attacked the Convergence's controller")
	}
	if _, err := g.DeclareAttackersWith([]game.AttackDeclaration{{Attacker: flier, Target: pw}}, game.DeclareAttackersParams{}); err == nil {
		t.Error("a flier attacked the controller's planeswalker")
	}
	if err := g.DeclareAttacker(walker, opp.ID); err != nil {
		t.Errorf("a ground creature could not attack: %v", err)
	}
	if err := g.DeclareAttacker(flier, g.Seats[2].ID); err != nil {
		t.Errorf("a flier could not attack another player: %v", err)
	}
}

func TestCmdr2B03SandwurmConvergenceMakesAWurmAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Sandwurm Convergence", "Enchantment", cmdr2SandwurmOracle, false)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Wurm" && c.Controller == me.ID && c.Power == 5 && c.Toughness == 5 {
			found = true
		}
	}
	if !found {
		t.Error("no 5/5 Wurm")
	}
}

// --- Soul Snare ------------------------------------------------------------------

func TestCmdr2B03SoulSnareExilesAnAttackerOfYouOrYourPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pw := pushWalkerForTest(g, opp.ID, "Their Walker", "", 3)
	atkYou := pushVanillaCreature(g, me.ID, "Attacks You", 2, 2)
	atkWalker := pushVanillaCreature(g, me.ID, "Attacks Walker", 2, 2)
	atkOther := pushVanillaCreature(g, me.ID, "Attacks Other", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	snare := pushCatalogPermanent(g, opp.ID, "Soul Snare", "Enchantment", cmdr2SoulSnareOracle, false)
	if _, err := g.DeclareAttackersWith([]game.AttackDeclaration{
		{Attacker: atkYou, Target: opp.ID}, {Attacker: atkWalker, Target: pw}, {Attacker: atkOther, Target: g.Seats[2].ID},
	}, game.DeclareAttackersParams{}); err != nil {
		t.Fatalf("DeclareAttackersWith: %v", err)
	}
	lockInAttacks(t, g)

	fillPoolColored(opp, "W", 1)
	if err := g.ActivateCatalogAbility(opp.ID, snare, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: atkOther}}}); err == nil {
		t.Fatal("exiled a creature attacking someone else")
	}
	if err := g.ActivateCatalogAbility(opp.ID, snare, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: atkWalker}}}); err != nil {
		t.Fatalf("exile a creature attacking my planeswalker: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(atkWalker) || g.Battlefield.Contains(snare) {
		t.Error("the attacker should be exiled and the Snare sacrificed")
	}
}

// --- Bladegriff Prototype --------------------------------------------------------

func TestCmdr2B03BladegriffTheDamagedPlayerChoosesWhatDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	griff := b12Push(g, me.ID, "Bladegriff Prototype", "Artifact Creature — Griffin", cmdr2BladegriffOracle, 3, 2)
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirRock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	thirdRock := b12Permanent(g, third.ID, "Third's Rock", "Artifact")
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	dealCombatDamageToPlayer(g, griff, opp.ID, 3)

	// The damaged player (opp), not Bladegriff's controller, is asked,
	// and the options are every nonland permanent of any of my opponents.
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("the controller was asked to choose")
	}
	p := latestPickTarget(g, opp.ID)
	if p == nil {
		t.Fatal("the damaged player was not asked to choose")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range p.PickTargetCards {
		offered[id] = true
	}
	if !offered[theirRock] || !offered[thirdRock] {
		t.Errorf("opponents' nonland permanents not all offered: %v", offered)
	}
	if offered[mine] || offered[land] {
		t.Errorf("offered a permanent of mine or a land: %v", offered)
	}
	if err := g.ResolvePickTarget(p.ID, p.Chooser, game.TargetRef{Kind: game.TargetCard, ID: thirdRock}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(thirdRock) {
		t.Error("the chosen permanent survived")
	}
	if !g.Battlefield.Contains(theirRock) {
		t.Error("an unchosen permanent was destroyed")
	}
}
