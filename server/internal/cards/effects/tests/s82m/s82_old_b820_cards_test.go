package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b820_cards_test.go — Alliances, graveyard recursion, removal, life,
// lands, creatures with keywords, artifacts and enchantments, batch 820
// (#5634, S82 OLD). Every test plays the real catalog entry through the
// engine.

const (
	oldB820BalduvianDead = "18d6b89a-3588-43d9-a247-4747a7d784fc"
	oldB820Reinforce     = "ff27ff96-afd2-45df-b799-3046179e27c6"
	oldB820Viscerid      = "5fd4be4c-70a3-4c91-a4c6-788a2051ff13"
	oldB820FyndhornDruid = "98e51f1a-ed6e-4286-a41b-e4656d08c934"
	oldB820TasteParadise = "c9a20d03-b90a-43df-8ab4-ff0b38c90a08"
	oldB820Glaciers      = "c6792d9f-8b74-43c4-814f-ba4adab2fdea"
	oldB820Valley        = "cd535fa3-6fd8-4227-97fd-3ef07cb0598d"
	oldB820Escort        = "1c6c1604-0724-40b1-9df6-71e57e5063c9"
	oldB820StormCauldron = "5a51b168-02d1-4eb4-8ccc-614ab6f6cffa"
	oldB820NaturesWrath  = "5a647f28-6e26-4262-8593-d582e0a691f7"
	oldB820Harlequin     = "c8e96a51-24c9-4955-8bf0-6ef448fa8011"
)

func oldB820Mana(p *game.Player, colors ...string) {
	p.ManaPool.AddMana(effectstest.ManaTokens(colors...)...)
}

func TestOldB820CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB820BalduvianDead: "Balduvian Dead", oldB820Reinforce: "Reinforcements", oldB820Viscerid: "Viscerid Drone",
		oldB820FyndhornDruid: "Fyndhorn Druid", oldB820TasteParadise: "Taste of Paradise",
		oldB820Glaciers: "Thawing Glaciers", oldB820Valley: "Sheltered Valley", oldB820Escort: "Kjeldoran Escort",
		oldB820StormCauldron: "Storm Cauldron", oldB820NaturesWrath: "Nature's Wrath", oldB820Harlequin: "Chaos Harlequin",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB820BalduvianDeadMakesAGraveborneThatIsSacrificedAtTheEnd(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.B12Push(g, me.ID, "Balduvian Dead", "Creature — Zombie", oldB820BalduvianDead, 2, 3)
	spell := effectstest.GraveyardCardOf(g, me, "Old Bolt", "Instant", "{R}", "")
	bear := effectstest.GraveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.AdvanceToMain(t, g)
	oldB820Mana(me, "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("a noncreature card paid the exile")
	}
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Error("the creature card was not exiled as the cost")
	}
	toks := effectstest.TokensNamed(g, "Graveborn")
	if len(toks) != 1 || toks[0].Controller != me.ID {
		t.Fatalf("Graveborn tokens %+v, want one of mine", toks)
	}
	id := toks[0].InstanceID
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 3 || tough != 1 || !effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
		t.Errorf("the Graveborn is %d/%d, haste %v", p, tough, effectstest.HasKeywordOnBattlefield(t, g, id, "haste"))
	}
	effectstest.AdvanceToStepOnThisTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("the Graveborn was not sacrificed at the beginning of the next end step")
	}
}

func TestOldB820ReinforcementsPutsUpToThreeCreatureCardsOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.GraveyardCardOf(g, me, "A", "Creature — Bear", "", "")
	b := effectstest.GraveyardCardOf(g, me, "B", "Creature — Bear", "", "")
	spell := effectstest.GraveyardCardOf(g, me, "Spell", "Instant", "", "")
	spellCard := effectstest.HandCardFull(me, "Reinforcements", "Instant", "{W}", oldB820Reinforce, []string{"W"})
	if err := g.CastSpell(me.ID, spellCard, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(spell)}); err == nil {
		t.Fatal("a noncreature card was a legal target")
	}
	if err := g.CastSpell(me.ID, spellCard, game.CastSpellParams{Targets: effectstest.CardRefs(a, b)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	order := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID)
	if order == nil {
		t.Fatal("no ordering prompt")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{b, a}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	n := len(me.Library.Cards)
	if me.Graveyard.Contains(a) || me.Graveyard.Contains(b) || !me.Graveyard.Contains(spell) || me.Library.Cards[n-1].InstanceID != b || me.Library.Cards[n-2].InstanceID != a {
		t.Error("the two creature cards should be on top in the chosen order and the spell left behind")
	}
}

func TestOldB820VisceridDroneSacrificesACreatureAndASwampToDestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drone := effectstest.B12Push(g, me.ID, "Viscerid Drone", "Creature — Homarid Drone", oldB820Viscerid, 1, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	swamp := effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	snow := effectstest.PushLandFor(g, me.ID, "Snow-Covered Swamp", "Basic Snow Land — Swamp")
	golem := effectstest.B12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	troll := effectstest.B12Creature(g, opp.ID, "Troll", "Creature — Troll", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, drone, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(golem), SacrificeIDs: []uuid.UUID{fodder, swamp}}); err == nil {
		t.Fatal("an artifact creature was a legal target of the first ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, drone, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(troll), SacrificeIDs: []uuid.UUID{fodder, swamp}}); err == nil {
		t.Fatal("a plain Swamp paid for the snow Swamp ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, drone, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(troll), SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("only one permanent was sacrificed")
	}
	if err := g.ActivateCatalogAbility(me.ID, drone, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(troll), SacrificeIDs: []uuid.UUID{fodder, swamp}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(troll) || g.Battlefield.Contains(fodder) || g.Battlefield.Contains(swamp) {
		t.Error("the Troll, the creature and the Swamp should all be gone")
	}
	// the snow ability can take the artifact creature
	fodder2 := effectstest.B12Creature(g, me.ID, "Fodder Two", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, drone).Tapped = false })
	if err := g.ActivateCatalogAbility(me.ID, drone, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(golem), SacrificeIDs: []uuid.UUID{fodder2, snow}}); err != nil {
		t.Fatalf("snow activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) {
		t.Error("the artifact creature survived the snow ability")
	}
}

func TestOldB820FyndhornDruidGainsFourOnlyIfItWasBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	druid := effectstest.B12Push(g, me.ID, "Fyndhorn Druid", "Creature — Elf Druid", oldB820FyndhornDruid, 2, 2)
	life := me.Life
	effectstest.Destroy(t, g, druid)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("an unblocked Druid gained %d life", me.Life-life)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	druid2 := effectstest.B12Push(g2, me2.ID, "Fyndhorn Druid", "Creature — Elf Druid", oldB820FyndhornDruid, 2, 2)
	wall := effectstest.B12Creature(g2, opp2.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.DeclareAttack(t, g2, opp2.ID, druid2)
	effectstest.AdvanceTo(t, g2, game.StepDeclareBlockers)
	if err := g2.DeclareBlocker(wall, druid2); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g2)
	life2 := me2.Life
	effectstest.Destroy(t, g2, druid2)
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Life != life2+4 {
		t.Errorf("life %d → %d, want +4 for a blocked Druid", life2, me2.Life)
	}
	_ = opp
}

func TestOldB820TasteOfParadiseGainsThreePlusThreePerPayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.HandCardFull(me, "Taste of Paradise", "Sorcery", "{3}{G}", oldB820TasteParadise, []string{"G"})
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{OptionalCosts: []int{0, 0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+9 {
		t.Errorf("life %d → %d, want +9 (3 plus 3 for each of 2 payments)", life, me.Life)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	plain := effectstest.HandCardFull(me2, "Taste of Paradise", "Sorcery", "{3}{G}", oldB820TasteParadise, []string{"G"})
	life2 := me2.Life
	if err := g2.CastSpell(me2.ID, plain, game.CastSpellParams{}); err != nil {
		t.Fatalf("plain cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Life != life2+3 {
		t.Errorf("unkicked: life %d → %d, want +3", life2, me2.Life)
	}
}

func TestOldB820ThawingGlaciersFetchesABasicTappedAndReturnsAtCleanup(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	glaciers := effectstest.PushCatalogPermanent(g, me.ID, "Thawing Glaciers", "Land", oldB820Glaciers, false)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	oldB820Mana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, glaciers, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	forests := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && c.Controller == me.ID {
			forests++
			if !c.Tapped {
				t.Error("the fetched land entered untapped")
			}
		}
	}
	if forests != 1 {
		t.Fatalf("basic lands fetched: %d, want 1", forests)
	}
	effectstest.AdvanceToStepOnThisTurn(t, g, game.StepEnd)
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	for i := 0; i < 4 && g.Battlefield.Contains(glaciers); i++ {
		_, _ = g.AdvanceStep()
		effectstest.AnswerOpenOrderingIfAny(t, g, me.ID)
	}
	if g.Battlefield.Contains(glaciers) || !me.Hand.Contains(glaciers) {
		t.Error("Thawing Glaciers was not returned to its owner's hand at the cleanup step")
	}
}

func TestOldB820ShelteredValleySacrificesItsTwinAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	first := effectstest.PushCatalogPermanent(g, me.ID, "Sheltered Valley", "Land", oldB820Valley, false)
	effectstest.AdvanceToMain(t, g)
	second := effectstest.PlayLandFromHand(t, g, "Sheltered Valley", oldB820Valley)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(first) || !g.Battlefield.Contains(second) {
		t.Errorf("the older Valley should be sacrificed and the new one stay (old %v, new %v)", g.Battlefield.Contains(first), g.Battlefield.Contains(second))
	}
	life := me.Life
	effectstest.BeginUpkeep(t, g, me.ID)
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1 with three or fewer lands", life, me.Life)
	}
	for i := 0; i < 3; i++ {
		effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	}
	life = me.Life
	effectstest.BeginUpkeep(t, g, me.ID)
	if me.Life != life {
		t.Errorf("with four lands life %d → %d, want no gain", life, me.Life)
	}
}

func TestOldB820KjeldoranEscortHasBanding(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Kjeldoran Escort", "Creature — Human Soldier", oldB820Escort, 2, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "banding") {
		t.Error("no banding")
	}
}

func TestOldB820StormCauldronGivesEveryoneALandAndReturnsTappedLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := 0
	g.ReadSnapshot(func() { before = g.EffectiveLandDropsLocked(opp) })
	effectstest.PushCatalogPermanent(g, me.ID, "Storm Cauldron", "Artifact", oldB820StormCauldron, false)
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := g.ActivateManaAbility(opp.ID, forest, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) || !opp.Hand.Contains(forest) {
		t.Error("the tapped land did not return to its owner's hand")
	}
	if len(opp.ManaPool) != 1 {
		t.Errorf("the mana is lost: pool %v", opp.ManaPool)
	}
	g.ReadSnapshot(func() {
		if n := g.EffectiveLandDropsLocked(opp); n != before+1 {
			t.Errorf("an opponent's land drops = %d, want %d: 'each player'", n, before+1)
		}
		if n := g.EffectiveLandDropsLocked(me); n != before+1 {
			t.Errorf("the controller's land drops = %d, want %d", n, before+1)
		}
	})
}

func TestOldB820NaturesWrathPunishesIslandsBlueSwampsAndBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Nature's Wrath", "Enchantment", oldB820NaturesWrath, false)
	effectstest.AdvanceToMain(t, g)
	island := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: island, Actor: opp.ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, island)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(island) {
		t.Error("the Island that entered was not sacrificed")
	}
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: forest, Actor: opp.ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Error("a Forest was sacrificed")
	}
	bat := effectstest.B12Creature(g, opp.ID, "Bat", "Creature — Bat", 1, 1)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, bat).Colors = []string{"B"}
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: bat, Actor: opp.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, bat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bat) {
		t.Error("the black creature that entered was not sacrificed")
	}
}

func TestOldB820ChaosHarlequinPumpsOrShrinksByTheExiledCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	harlequin := effectstest.B12Push(g, me.ID, "Chaos Harlequin", "Creature — Human", oldB820Harlequin, 2, 4)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	oldB820Mana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, harlequin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, harlequin); p != -2 {
		t.Errorf("power %d after exiling a land, want -2", p)
	}
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	oldB820Mana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, harlequin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, harlequin); p != 0 {
		t.Errorf("power %d after exiling a nonland (-4 then +2), want 0", p)
	}
}
