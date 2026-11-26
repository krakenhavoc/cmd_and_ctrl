package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 488 (#5582): Worldwake, card draw and library, instants and
// sorceries, removal.

const (
	wwk488SpellContortion  = "01003192-8927-4e55-8dd9-e9e088a3fbe4"
	wwk488ShorelineSalv    = "02c08d0b-0d76-4533-a7e9-d40c7c4b3450"
	wwk488Twitch           = "321fd970-a259-448d-bd98-e85e386174f9"
	wwk488SelectiveMemory  = "3cfb4977-5b6d-4220-9c54-70920b3087ed"
	wwk488RicochetTrap     = "15df3467-c67e-4ae2-9e6f-ffb6ad288a5a"
	wwk488MiresToll        = "3e0756d8-c2bd-40f1-a281-0828c5a2da03"
	wwk488StoneIdolTrap    = "47cd98e2-0eaf-4292-a840-c5e8d035c83c"
	wwk488PermafrostTrap   = "f037b319-e559-42ad-bcfe-d43d07ce05e7"
	wwk488RuinGhost        = "169705c3-32c1-4628-b108-c37ca5f27e24"
	wwk488RoilingTerrain   = "5e90f187-42fe-4f32-ad75-302e6709192d"
	wwk488SlingbowTrap     = "66ce0ae5-7000-4e4a-891f-f571f0e58cd7"
	wwk488TuktukScrapper   = "85cf2403-b419-4364-8ac9-67dd1ceddf9e"
	wwk488TrapTypeLine     = "Instant — Trap"
	wwk488BearType         = "Creature — Bear"
	wwk488ForestType       = "Basic Land — Forest"
	wwk488SwampType        = "Basic Land — Swamp"
	wwk488IslandType       = "Basic Land — Island"
	wwk488ArtifactTypeLine = "Artifact"
)

func TestWwkB488CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		wwk488SpellContortion: "Spell Contortion", wwk488ShorelineSalv: "Shoreline Salvager",
		wwk488Twitch: "Twitch", wwk488SelectiveMemory: "Selective Memory",
		wwk488RicochetTrap: "Ricochet Trap", wwk488MiresToll: "Mire's Toll",
		wwk488StoneIdolTrap: "Stone Idol Trap", wwk488PermafrostTrap: "Permafrost Trap",
		wwk488RuinGhost: "Ruin Ghost", wwk488RoilingTerrain: "Roiling Terrain",
		wwk488SlingbowTrap: "Slingbow Trap", wwk488TuktukScrapper: "Tuktuk Scrapper",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

// --- Spell Contortion --------------------------------------------------------

// wwk488Contortion has seat 0 cast a Bear and seat 1 answer with a Spell
// Contortion kicked `kicks` times, returning the table, the Bear and the
// seats' hand size once the Contortion is on the stack.
func wwk488Contortion(t *testing.T, kicks int) (*game.Game, *game.Player, *game.Player, uuid.UUID, int) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Bear", wwk488BearType, "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := effectstest.HandCardFull(them, "Spell Contortion", "Instant", "", wwk488SpellContortion, []string{"U"})
	mana := ""
	var optional []int
	for i := 0; i < kicks; i++ {
		mana += "{C}{U}"
		optional = append(optional, 0)
	}
	if mana != "" {
		if err := g.AddManaForEffect(them.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	}
	if err := g.CastSpell(them.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: optional,
		Targets: effectstest.RtCardTarget(victim)}); err != nil {
		t.Fatalf("cast Spell Contortion: %v", err)
	}
	return g, me, them, victim, them.Hand.Size()
}

func TestWwkB488SpellContortionCountersAndDrawsPerKick(t *testing.T) {
	g, me, them, victim, hand := wwk488Contortion(t, 2)
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) || g.Battlefield.Contains(victim) {
		t.Error("the declined spell was not countered")
	}
	if got := them.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d (two draws for two kicks)", got, hand+2)
	}
}

func TestWwkB488SpellContortionDrawsEvenWhenThePayerPays(t *testing.T) {
	g, me, them, victim, hand := wwk488Contortion(t, 1)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the spell that paid {2} was countered")
	}
	if got := them.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (one draw for one kick)", got, hand+1)
	}
}

func TestWwkB488SpellContortionUnkickedDrawsNothing(t *testing.T) {
	g, me, them, _, hand := wwk488Contortion(t, 0)
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := them.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d (no kicks, no draw)", got, hand)
	}
}

// --- Shoreline Salvager ------------------------------------------------------

func TestWwkB488ShorelineSalvagerDrawsWithAnIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	salvager := effectstest.B12Push(g, me.ID, "Shoreline Salvager", "Creature — Surrakar", wwk488ShorelineSalv, 3, 3)
	effectstest.FreeSpellPermanent(g, me.ID, "Island", wwk488IslandType)
	effectstest.SeedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, salvager, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestWwkB488ShorelineSalvagerNeedsAnIslandAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	salvager := effectstest.B12Push(g, me.ID, "Shoreline Salvager", "Creature — Surrakar", wwk488ShorelineSalv, 3, 3)
	effectstest.FreeSpellPermanent(g, me.ID, "Swamp", wwk488SwampType)
	effectstest.SeedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, salvager, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Fatal("the trigger asked with no Island")
	}
	effectstest.FreeSpellPermanent(g, me.ID, "Island", wwk488IslandType)
	effectstest.DealCombatDamageToPlayer(g, salvager, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d (nothing drawn)", me.Hand.Size(), hand)
	}
}

// --- Twitch ------------------------------------------------------------------

func TestWwkB488TwitchTapsOrUntapsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", wwk488BearType, 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Twitch", "Instant", wwk488Twitch, effectstest.RtCardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B31Tapped(t, g, bear) {
		t.Error("the first option should tap the target")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the Twitch was cast from it, then one drawn)", me.Hand.Size(), hand+1)
	}
}

func TestWwkB488TwitchAcceptsALandAndAnArtifactButNotAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.FreeSpellPermanent(g, opp.ID, "Forest", wwk488ForestType)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Twitch", "Instant", wwk488Twitch, effectstest.RtCardTarget(land)); err != nil {
		t.Fatalf("a land was refused: %v", err)
	}
	g2 := effectstest.NewCatalogGame(t)
	ench := effectstest.FreeSpellPermanent(g2, g2.Seats[1].ID, "Aura", "Enchantment")
	if err := effectstest.CastCatalogSpellErrStrict(t, g2, "Twitch", "Instant", wwk488Twitch, effectstest.RtCardTarget(ench)); err == nil {
		t.Fatal("an enchantment was accepted")
	}
}

// --- Selective Memory --------------------------------------------------------

func TestWwkB488SelectiveMemoryExilesAnyNumberOfNonlandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	a := effectstest.SeedLibraryTop(me, "A", "Sorcery", "{1}")
	b := effectstest.SeedLibraryTop(me, "B", "Creature — Bear", "{2}")
	land := effectstest.SeedLibraryTop(me, "Forest", wwk488ForestType, "")
	effectstest.CastCatalogSpell(t, g, "Selective Memory", "Sorcery", wwk488SelectiveMemory, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("a chosen card is in %v, want exile", z)
		}
	}
	if !libraryHas(me, land) {
		t.Error("the land should still be in the library")
	}
}

func libraryHas(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Library.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// --- Ricochet Trap -----------------------------------------------------------

func TestWwkB488RicochetTrapChangesTheTargetOfASingleTargetSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim, other := g.Seats[0], g.Seats[1], g.Seats[2]
	bolt := effectstest.CastFor(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtPlayerRef(victim.ID)})
	effectstest.CastFor(t, g, me, "Ricochet Trap", wwk488TrapTypeLine, wwk488RicochetTrap,
		game.CastSpellParams{Targets: effectstest.RtCardTarget(bolt)})
	for i := 0; i < 8 && effectstest.LatestRetarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.LatestRetarget(g, me.ID) == nil {
		t.Fatal("Ricochet Trap opened no retarget prompt")
	}
	vBefore, oBefore := effectstest.LifeOf(g, victim.ID), effectstest.LifeOf(g, other.ID)
	effectstest.AnswerRetargetWith(t, g, me.ID, effectstest.RtPlayerRef(other.ID))
	effectstest.SettleThroughCopyRetarget(t, g, me.ID, victim.ID)
	if got := effectstest.LifeOf(g, other.ID); got != oBefore-3 {
		t.Errorf("the new target has %d life, want %d", got, oBefore-3)
	}
	if got := effectstest.LifeOf(g, victim.ID); got != vBefore {
		t.Errorf("the old target has %d life, want %d", got, vBefore)
	}
}

func TestWwkB488RicochetTrapCostNeedsAnOpponentsBlueSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	check := func() bool {
		var ok bool
		g.ReadSnapshot(func() { ok = effects.S82OldB488AnOpponentCastABlueSpell(g, me.ID) })
		return ok
	}
	if check() {
		t.Fatal("offered before any spell was cast")
	}
	red := effectstest.HandCardFull(opp, "Red Spell", "Instant", "", "", []string{"R"})
	if err := g.CastSpell(opp.ID, red, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the red spell: %v", err)
	}
	if check() {
		t.Error("offered after a red spell")
	}
	blue := effectstest.HandCardFull(opp, "Blue Spell", "Instant", "", "", []string{"U"})
	if err := g.CastSpell(opp.ID, blue, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the blue spell: %v", err)
	}
	if !check() {
		t.Error("not offered after an opponent cast a blue spell")
	}
	mine := effectstest.HandCardFull(me, "Blue Mine", "Instant", "", "", []string{"U"})
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	if err := g2.CastSpell(me2.ID, effectstest.HandCardFull(me2, "Blue Mine", "Instant", "", "", []string{"U"}), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast my own blue spell: %v", err)
	}
	var own bool
	g2.ReadSnapshot(func() { own = effects.S82OldB488AnOpponentCastABlueSpell(g2, me2.ID) })
	if own {
		t.Error("your own blue spell offered the trap cost")
	}
	_ = mine
}

// --- Mire's Toll -------------------------------------------------------------

func TestWwkB488MiresTollRevealsOnePerSwampAndDiscardsTheChosenOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.FreeSpellPermanent(g, me.ID, "Swamp", wwk488SwampType)
	effectstest.FreeSpellPermanent(g, me.ID, "Swamp", wwk488SwampType)
	effectstest.FreeSpellPermanent(g, me.ID, "Forest", wwk488ForestType)
	opp.Hand.Cards = nil
	var hand []uuid.UUID
	for _, name := range []string{"A", "B", "C", "D"} {
		hand = append(hand, effectstest.HandCardForTest(opp, name, "Sorcery", ""))
	}
	effectstest.CastAt(t, g, "Mire's Toll", "Sorcery", wwk488MiresToll, effectstest.PbPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if ask == nil || ask.ChooseMin != 2 || ask.ChooseMax != 2 {
		t.Fatalf("the player reveals exactly two cards (two Swamps), got %+v", ask)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, hand[0], hand[1])
	rp := effectstest.RevealPickFor(g, me.ID)
	if rp == nil || len(rp.ChooseCards) != 2 {
		t.Fatalf("want two revealed cards to choose from, got %+v", rp)
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{hand[1]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if !opp.Graveyard.Contains(hand[1]) || opp.Hand.Size() != 3 {
		t.Errorf("the chosen card was not discarded (hand %d)", opp.Hand.Size())
	}
}

func TestWwkB488MiresTollWithNoSwampsAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.HandCardForTest(opp, "A", "Sorcery", "")
	effectstest.CastAt(t, g, "Mire's Toll", "Sorcery", wwk488MiresToll, effectstest.PbPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) != nil {
		t.Error("no Swamps still raised a prompt")
	}
}

// --- Stone Idol Trap ---------------------------------------------------------

func TestWwkB488StoneIdolTrapMakesAConstructThatLeavesAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Stone Idol Trap", wwk488TrapTypeLine, wwk488StoneIdolTrap, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Construct"); n != 1 {
		t.Fatalf("%d Constructs, want 1", n)
	}
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if c.Name == "Construct" && (c.CurrentPower() != 6 || c.CurrentToughness() != 12 || !game.HasKeyword(&g.Battlefield.Cards[i], "trample") || !c.IsArtifact()) {
			t.Errorf("the Construct is %d/%d or lacks trample or is not an artifact", c.CurrentPower(), c.CurrentToughness())
		}
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Construct"); n != 0 {
		t.Errorf("%d Constructs after the end step, want 0 (exiled)", n)
	}
}

func TestWwkB488StoneIdolTrapCostsOneLessPerAttackingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("A", "", 2, 2))
	b := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("B", "", 2, 2))
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	id := effectstest.HandCardFull(me, "Stone Idol Trap", wwk488TrapTypeLine, "{5}{R}", wwk488StoneIdolTrap, []string{"R"})
	effectstest.B06AddMana(me, "R", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("with two attackers {5}{R} costs {3}{R}: %v", err)
	}
}

// --- Permafrost Trap ---------------------------------------------------------

func TestWwkB488PermafrostTrapTapsUpToTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", wwk488BearType, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", wwk488BearType, 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", wwk488BearType, 2, 2)
	effectstest.CastFor(t, g, me, "Permafrost Trap", wwk488TrapTypeLine, wwk488PermafrostTrap,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B31Tapped(t, g, a) || !effectstest.B31Tapped(t, g, b) {
		t.Error("both targets should be tapped")
	}
	if effectstest.B31Tapped(t, g, c) {
		t.Error("an untargeted creature was tapped")
	}
}

func TestWwkB488PermafrostTrapCostNeedsAnOpponentsGreenCreatureToHaveEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	check := func() bool {
		var ok bool
		g.ReadSnapshot(func() { ok = effects.S82OldB488AnOpponentHadAGreenCreatureEnter(g, me.ID) })
		return ok
	}
	enter := func(who uuid.UUID, typeLine string, colors []string) {
		g.WithWriteLock(func() {
			if _, err := g.CreateTokensForEffect(who, game.Card{Name: "Token", TypeLine: typeLine, Colors: colors, Power: 1, Toughness: 1}, 1, game.TokenEntryOptions{}); err != nil {
				t.Errorf("CreateTokensForEffect: %v", err)
			}
		})
	}
	if check() {
		t.Fatal("offered before any creature entered")
	}
	enter(me.ID, "Token Creature — Bear", []string{"G"})
	enter(opp.ID, "Token Creature — Bear", []string{"R"})
	enter(opp.ID, "Token Land", []string{"G"})
	if check() {
		t.Error("offered with only your own green creature, an opponent's red one and a green noncreature")
	}
	trap := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: trap, Name: "Permafrost Trap", TypeLine: wwk488TrapTypeLine, ManaCost: "{2}{U}{U}",
		OracleID: wwk488PermafrostTrap, Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "U")
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{AlternativeCost: "trap", Strict: true}); err == nil {
		t.Fatal("the trap cost was accepted with no green opponent creature having entered")
	}
	enter(opp.ID, "Token Creature — Elf", []string{"G"})
	if !check() {
		t.Fatal("not offered once an opponent had a green creature enter")
	}
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{AlternativeCost: "trap", Strict: true}); err != nil {
		t.Fatalf("the trap cost was refused once an opponent had a green creature enter: %v", err)
	}
}

// --- Ruin Ghost --------------------------------------------------------------

func TestWwkB488RuinGhostBlinksALandYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ghost := effectstest.B12Push(g, me.ID, "Ruin Ghost", "Creature — Spirit", wwk488RuinGhost, 1, 1)
	land := effectstest.FreeSpellPermanent(g, me.ID, "Forest", wwk488ForestType)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, land).Tapped = true })
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, ghost, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(land)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNamed(g, "Forest") != 1 {
		t.Fatal("the Forest did not come back")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && c.Tapped {
			t.Error("the returned Forest is still tapped (it is a new object)")
		}
	}
}

func TestWwkB488RuinGhostRefusesAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghost := effectstest.B12Push(g, me.ID, "Ruin Ghost", "Creature — Spirit", wwk488RuinGhost, 1, 1)
	land := effectstest.FreeSpellPermanent(g, opp.ID, "Forest", wwk488ForestType)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, ghost, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(land)}); err == nil {
		t.Fatal("an opponent's land was accepted")
	}
}

// --- Roiling Terrain ---------------------------------------------------------

func TestWwkB488RoilingTerrainDestroysALandAndCountsLandCardsInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.FreeSpellPermanent(g, opp.ID, "Forest", wwk488ForestType)
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Forest", TypeLine: wwk488ForestType, Owner: opp.ID, Controller: opp.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "A Bear", TypeLine: wwk488BearType, Owner: opp.ID, Controller: opp.ID})
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Roiling Terrain", "Sorcery", wwk488RoilingTerrain, effectstest.RtCardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("life %d, want %d (two land cards: the old one and the one just destroyed)", got, life-2)
	}
}

func TestWwkB488RoilingTerrainRefusesANonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", wwk488BearType, 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Roiling Terrain", "Sorcery", wwk488RoilingTerrain, effectstest.RtCardTarget(bear)); err == nil {
		t.Fatal("a creature was accepted")
	}
}

// --- Slingbow Trap -----------------------------------------------------------

func TestWwkB488SlingbowTrapDestroysAnAttackingFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bat", TypeLine: "Creature — Bat", Colors: []string{"B"}, Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	ground := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Ground", "", 2, 2))
	effectstest.DeclareAttack(t, g, opp.ID, flier, ground)
	notAFlier := effectstest.HandCardFull(opp, "Slingbow Trap", wwk488TrapTypeLine, "", wwk488SlingbowTrap, []string{"G"})
	if err := g.CastSpell(opp.ID, notAFlier, game.CastSpellParams{Strict: true, Targets: effectstest.RtCardTarget(ground)}); err == nil {
		t.Fatal("an attacker without flying was accepted")
	}
	trap := effectstest.HandCardFull(opp, "Slingbow Trap", wwk488TrapTypeLine, "", wwk488SlingbowTrap, []string{"G"})
	if err := g.CastSpell(opp.ID, trap, game.CastSpellParams{Strict: true, Targets: effectstest.RtCardTarget(flier)}); err != nil {
		t.Fatalf("cast Slingbow Trap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flier) {
		t.Error("the attacking flier survived")
	}
}

func TestWwkB488SlingbowTrapCostNeedsABlackAttackingFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	check := func() bool {
		var ok bool
		g.ReadSnapshot(func() { ok = effects.S82OldB488ABlackCreatureWithFlyingIsAttacking(g, opp.ID) })
		return ok
	}
	white := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Angel", TypeLine: "Creature — Angel", Colors: []string{"W"}, Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	blackGround := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Zombie", TypeLine: "Creature — Zombie", Colors: []string{"B"}, Power: 1, Toughness: 1})
	effectstest.DeclareAttack(t, g, opp.ID, white, blackGround)
	if check() {
		t.Error("offered with only a white flier and a black ground attacker")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	bat := effectstest.ApaPush(g2, me2.ID, me2.ID, game.Card{Name: "Bat", TypeLine: "Creature — Bat", Colors: []string{"B"}, Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	effectstest.DeclareAttack(t, g2, opp2.ID, bat)
	var ok bool
	g2.ReadSnapshot(func() { ok = effects.S82OldB488ABlackCreatureWithFlyingIsAttacking(g2, opp2.ID) })
	if !ok {
		t.Error("not offered with a black attacking flier")
	}
}

// --- Tuktuk Scrapper ---------------------------------------------------------

func TestWwkB488TuktukScrapperDestroysAnArtifactAndHurtsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Creature(g, me.ID, "Other Ally", "Creature — Human Ally", 1, 1)
	rock := effectstest.FreeSpellPermanent(g, opp.ID, "Rock", wwk488ArtifactTypeLine)
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastHoldCreature(t, g, "Tuktuk Scrapper", "Creature — Goblin Artificer Ally", wwk488TuktukScrapper, 2, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("life %d, want %d (two Allies)", got, life-2)
	}
}

func TestWwkB488TuktukScrapperDealsNothingToTheControllerOfAnIndestructibleArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: wwk488ArtifactTypeLine, Keywords: []string{"indestructible"}})
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastHoldCreature(t, g, "Tuktuk Scrapper", "Creature — Goblin Artificer Ally", wwk488TuktukScrapper, 2, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("an indestructible artifact was destroyed")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life {
		t.Errorf("life %d, want %d (it was not put into a graveyard)", got, life)
	}
}

func TestWwkB488TuktukScrapperMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.FreeSpellPermanent(g, opp.ID, "Rock", wwk488ArtifactTypeLine)
	effectstest.CastHoldCreature(t, g, "Tuktuk Scrapper", "Creature — Goblin Artificer Ally", wwk488TuktukScrapper, 2, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("declining still destroyed the artifact")
	}
}
