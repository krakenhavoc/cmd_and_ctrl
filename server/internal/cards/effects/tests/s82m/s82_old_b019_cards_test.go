package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b019_cards_test.go — Secret Lair Drop: Enchantments (S82 older
// sets, batch 019, #5462). Each test plays the real catalog entry.

const (
	oldB19Iroas      = "36041cf2-159f-4809-9377-930ac4db545f"
	oldB19Storm      = "8586169f-cb8c-4424-ba64-a7c1cc00680e"
	oldB19Pit        = "91e6fb47-59e4-4616-b8dd-3a7e30070074"
	oldB19Puca       = "a8e6655c-e63f-4a60-98f6-9d5bf8631488"
	oldB19Strangle   = "b66c6e3a-daed-43a3-95d6-4f9d6223bbf1"
	oldB19Evening    = "c56ea5ce-c8de-46c2-a8dd-bdd41467a28c"
	oldB19Serpent    = "d31c6043-e005-4efe-9aa7-d82cb122eebe"
	oldB19Mists      = "db25d48e-e5eb-4364-9d7a-f9757ee4ccbe"
	oldB19Invasion   = "db9c0d94-9fa5-4f07-be48-23e19d2f3d68"
	oldB19HiveMind   = "f97e405d-4c24-4d4d-9e19-e349c073113c"
	oldB19Vineyard   = "faa085c3-705f-465d-9290-0e22276ad06d"
	oldB19Ench       = "Enchantment"
	oldB19Mountain   = "Basic Land — Mountain"
	oldB19Instant    = "Instant"
	oldB19Sorcery    = "Sorcery"
	oldB19GodTypeLn  = "Legendary Enchantment Creature — God"
	oldB19Legendary  = "Legendary Creature — Test"
	oldB19Creature   = "Creature — Bear"
	oldB19EnchantNam = "Enchanted Evening"
)

func TestOldB19EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB19Iroas:    "Iroas, God of Victory",
		oldB19Storm:    "Possibility Storm",
		oldB19Pit:      "Bottomless Pit",
		oldB19Puca:     "Puca's Mischief",
		oldB19Strangle: "Stranglehold",
		oldB19Evening:  oldB19EnchantNam,
		oldB19Serpent:  "Serpent of Yawning Depths",
		oldB19Mists:    "Necrogen Mists",
		oldB19Invasion: "Invasion Plans",
		oldB19HiveMind: "Hive Mind",
		oldB19Vineyard: "Eladamri's Vineyard",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB19Devotion puts n red-white permanents ({R}{W} each, two symbols).
func oldB19Devotion(g *game.Game, owner uuid.UUID, pairs int) {
	for i := 0; i < pairs; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Devotee", TypeLine: oldB19Ench, ManaCost: "{R}{W}",
			Owner: owner, Controller: owner,
		})
	}
}

func TestOldB19IroasIsACreatureOnlyAtSevenDevotion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	iroas := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Iroas, God of Victory", TypeLine: oldB19GodTypeLn, ManaCost: "{2}{R}{W}",
		OracleID: oldB19Iroas, Power: 7, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	// Iroas' own cost has two symbols: {R}{W}. Two more pairs make six.
	oldB19Devotion(g, me.ID, 2)
	if effectstest.LayeredCard(t, g, iroas).IsCreature() {
		t.Error("a creature at devotion 6")
	}
	oldB19Devotion(g, me.ID, 1)
	if !effectstest.LayeredCard(t, g, iroas).IsCreature() {
		t.Error("not a creature at devotion 8")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, iroas, "indestructible") {
		t.Error("not indestructible")
	}
}

func TestOldB19IroasGivesMenaceAndPreventsDamageToAttackersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEnchantment(g, oldB19Iroas, "Iroas, God of Victory", oldB19GodTypeLn, me.ID)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 5)
	idler := effectstest.PushVanillaCreature(g, me.ID, "Idler", 2, 5)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 5)
	g.BumpLayerVersionForTest()
	for _, id := range []uuid.UUID{attacker, idler} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "menace") {
			t.Errorf("%s has no menace", effectstest.LayeredCard(t, g, id).Name)
		}
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "menace") {
		t.Error("an opponent's creature has menace")
	}
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	for _, id := range []uuid.UUID{attacker, idler, theirs} {
		g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(theirs, id, 3) })
	}
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{attacker: 0, idler: 3, theirs: 3} {
		if got := effectstest.FindBattlefieldCardByID(g, id).DamageMarked; got != want {
			t.Errorf("damage on %s = %d, want %d", effectstest.FindBattlefieldCardByID(g, id).Name, got, want)
		}
	}
}

func TestOldB19BottomlessPitDiscardsAtRandomAtEveryUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEnchantment(g, oldB19Pit, "Bottomless Pit", oldB19Ench, me.ID)
	for _, p := range []*game.Player{me, opp} {
		for p.Hand.Size() < 3 {
			p.Hand.PushTop(game.NewCard("Filler", p.ID))
		}
	}
	m0, o0 := me.Hand.Size(), opp.Hand.Size()
	effectstest.BeginUpkeep(t, g, opp.ID)
	if opp.Hand.Size() != o0-1 || me.Hand.Size() != m0 {
		t.Errorf("after the opponent's upkeep: their hand %d (was %d), mine %d (was %d)", opp.Hand.Size(), o0, me.Hand.Size(), m0)
	}
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		t.Error("the discard was a choice, not random")
	}
	effectstest.BeginUpkeep(t, g, me.ID)
	if me.Hand.Size() != m0-1 {
		t.Errorf("my hand %d, want %d", me.Hand.Size(), m0-1)
	}
	opp.Hand.Cards = nil
	effectstest.BeginUpkeep(t, g, opp.ID) // an empty hand discards nothing
}

func TestOldB19NecrogenMistsMakesThatPlayerChoose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEnchantment(g, oldB19Mists, "Necrogen Mists", oldB19Ench, me.ID)
	keep := game.NewCard("Keep", opp.ID)
	toss := game.NewCard("Toss", opp.ID)
	opp.Hand.Cards = nil
	opp.Hand.PushTop(keep)
	opp.Hand.PushTop(toss)
	effectstest.BeginUpkeep(t, g, opp.ID)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Fatal("the enchantment's controller was asked")
	}
	if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
		t.Fatal("the upkeep's player was not asked")
	}
	effectstest.AnswerDiscard(t, g, opp.ID, toss.InstanceID)
	if !opp.Hand.Contains(keep.InstanceID) || opp.Hand.Contains(toss.InstanceID) {
		t.Error("the chosen card was not the one discarded")
	}
}

func TestOldB19EnchantedEveningMakesEveryPermanentAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	evening := effectstest.SeedEnchantment(g, oldB19Evening, oldB19EnchantNam, oldB19Ench, me.ID)
	land := effectstest.PushLandFor(g, opp.ID, "Mountain", oldB19Mountain)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	g.BumpLayerVersionForTest()
	for _, id := range []uuid.UUID{evening, land, bear} {
		c := effectstest.LayeredCard(t, g, id)
		if !c.HasCardType("enchantment") {
			t.Errorf("%s is not an enchantment (%v)", c.Name, c.TypeLine)
		}
	}
	if c := effectstest.LayeredCard(t, g, land); !c.IsLand() {
		t.Error("the Mountain stopped being a land")
	}
	if c := effectstest.LayeredCard(t, g, bear); !c.IsCreature() {
		t.Error("the Bear stopped being a creature")
	}
}

func TestOldB19EladamrisVineyardGivesThatPlayerGGAtTheirFirstMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedEnchantment(g, oldB19Vineyard, "Eladamri's Vineyard", oldB19Ench, me.ID)
	for _, p := range []*game.Player{opp, me} {
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginPrecombatMain, Actor: p.ID}) })
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.PoolColors(p); len(got) != 2 || got[0] != "G" || got[1] != "G" {
			t.Errorf("%s's pool = %v, want {G}{G}", p.Name, got)
		}
		other := me
		if p == me {
			other = opp
		}
		if n := len(effectstest.PoolColors(other)); n > 2 {
			t.Errorf("the other player has %d mana", n)
		}
	}
}

func TestOldB19HiveMindEveryOtherPlayerCopiesAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEnchantment(g, oldB19HiveMind, "Hive Mind", oldB19Ench, me.ID)
	ogre := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Ogre", 3, 30)
	before := len(g.Stack.Cards)
	effectstest.CastCatalogSpell(t, g, "Bombard", oldB19Instant, effectstest.EoeABBombardOracle, effectstest.CardRefs(ogre))
	// The trigger goes on the stack above the spell.
	effectstest.PassPriorityAroundTable(t, g)
	// Each of the three other players is offered the new-targets choice for
	// their copy; the copies and the spell are all on or through the stack.
	asked := 0
	for _, p := range g.Seats[1:] {
		for effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, p.ID) != nil {
			effectstest.AnswerMayChoice(t, g, p.ID, false)
			asked++
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	dmg := effectstest.FindBattlefieldCardByID(g, ogre)
	if dmg == nil {
		t.Fatal("the Ogre died")
	}
	// Four Bombards (the spell and three copies) of 4 damage each.
	if dmg.DamageMarked != 16 {
		t.Errorf("damage on the Ogre = %d, want 16 (the spell and three copies); stack before %d; asked %d", dmg.DamageMarked, before, asked)
	}
	// A creature spell is not copied.
	effectstest.CastCatalogSpell(t, g, "Bear Cub", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB19SerpentOfYawningDepthsOnlySeaMonstersBlockSeaMonsters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	serpent := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Serpent of Yawning Depths", TypeLine: "Enchantment Creature — Serpent",
		OracleID: oldB19Serpent, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	kraken := effectstest.B12Creature(g, me.ID, "Kraken", "Creature — Kraken", 5, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	groundBlocker := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 2)
	octopus := effectstest.B12Creature(g, opp.ID, "Octopus", "Creature — Octopus", 2, 2)
	leviathan := effectstest.B12Creature(g, opp.ID, "Leviathan", "Creature — Leviathan", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, serpent, kraken, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(groundBlocker, serpent); err == nil {
		t.Error("a Bear blocked the Serpent")
	}
	if err := g.DeclareBlocker(groundBlocker, kraken); err == nil {
		t.Error("a Bear blocked the Kraken")
	}
	if err := g.DeclareBlocker(groundBlocker, bear); err != nil {
		t.Errorf("a Bear could not block a Bear: %v", err)
	}
	if err := g.DeclareBlocker(octopus, serpent); err != nil {
		t.Errorf("an Octopus could not block the Serpent: %v", err)
	}
	if err := g.DeclareBlocker(leviathan, kraken); err != nil {
		t.Errorf("a Leviathan could not block the Kraken: %v", err)
	}
}

func TestOldB19StrangleholdStopsOpponentSearchesButNotYours(t *testing.T) {
	for _, tc := range []struct {
		name      string
		searcher  int
		wantFound bool
	}{{"an opponent can't search", 1, false}, {"its controller can", 0, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			effectstest.SeedEnchantment(g, oldB19Strangle, "Stranglehold", oldB19Ench, me.ID)
			p := g.Seats[tc.searcher]
			land := game.NewCard("Mountain", p.ID)
			land.TypeLine = oldB19Mountain
			p.Library.PushTop(land)
			g.WithWriteLock(func() {
				_ = g.SearchLibraryForEffectWithOptions(p.ID, func(c game.Card) bool { return c.IsLand() }, game.ZoneHand, 1, false, true, false)
			})
			if c := effectstest.SearchPromptFor(g, p.ID); c != nil {
				// A search that is allowed asks the searcher to pick.
				if err := g.ResolveSearchLibrary(c.ID, p.ID, []uuid.UUID{land.InstanceID}); err != nil {
					t.Fatalf("ResolveSearchLibrary: %v", err)
				}
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := p.Hand.Contains(land.InstanceID); got != tc.wantFound {
				t.Errorf("land in hand = %v, want %v", got, tc.wantFound)
			}
			searched := 0
			for _, ev := range g.Events {
				if ev.Kind == game.EventSearchLibrary {
					searched++
				}
			}
			if (searched > 0) != tc.wantFound {
				t.Errorf("search events = %d, want a search = %v", searched, tc.wantFound)
			}
			if !tc.wantFound && effectstest.SearchPromptFor(g, p.ID) != nil {
				t.Error("a prompt was opened for a search that can't happen")
			}
		})
	}
}

func TestOldB19StrangleholdStillShufflesWhenTheSearchCant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEnchantment(g, oldB19Strangle, "Stranglehold", oldB19Ench, me.ID)
	for opp.Library.Size() < 40 {
		opp.Library.PushTop(game.NewCard("Filler", opp.ID))
	}
	var order []uuid.UUID
	for _, c := range opp.Library.Cards {
		order = append(order, c.InstanceID)
	}
	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffectWithOptions(opp.ID, func(game.Card) bool { return true }, game.ZoneHand, 1, false, true, false)
	})
	same := true
	for i, c := range opp.Library.Cards {
		if c.InstanceID != order[i] {
			same = false
		}
	}
	if same {
		t.Error("the library was not shuffled")
	}
	if opp.Hand.Size() != 0 && opp.Library.Size() != len(order) {
		t.Error("a card left the library")
	}
}

func TestOldB19StrangleholdSkipsOnlyAnOpponentsExtraTurn(t *testing.T) {
	t.Run("an opponent's extra turn is skipped", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Turn.ActiveSeat
		effectstest.SeedEnchantment(g, oldB19Strangle, "Stranglehold", oldB19Ench, g.Seats[(me+1)%len(g.Seats)].ID)
		effectstest.QueueOwnExtraTurn(t, g)
		effectstest.EndTurn(t, g)
		effectstest.AssertTurnOf(t, g, (me+1)%len(g.Seats), false, "after Stranglehold skipped the opponent's extra turn")
	})
	t.Run("its controller's extra turn is taken", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Turn.ActiveSeat
		effectstest.SeedEnchantment(g, oldB19Strangle, "Stranglehold", oldB19Ench, g.Seats[me].ID)
		effectstest.QueueOwnExtraTurn(t, g)
		effectstest.EndTurn(t, g)
		effectstest.AssertTurnOf(t, g, me, true, "the controller's own extra turn")
	})
}

func TestOldB19PucasMischiefExchangesWithAnEqualOrLesserManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEnchantment(g, oldB19Puca, "Puca's Mischief", oldB19Ench, me.ID)
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mine", TypeLine: oldB19Creature,
		ManaCost: "{2}{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	cheap := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cheap", TypeLine: oldB19Creature,
		ManaCost: "{1}{G}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	pricey := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Pricey", TypeLine: oldB19Creature,
		ManaCost: "{4}{G}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	effectstest.BeginUpkeep(t, g, me.ID)
	pc := effectstest.LatestPickTarget(g, me.ID)
	if pc == nil {
		t.Fatal("no target prompt")
	}
	effectstest.CloudDeckPickTargets(t, g, me.ID, mine)
	pc = effectstest.LatestPickTarget(g, me.ID)
	if pc == nil {
		t.Fatal("no second target prompt")
	}
	for _, id := range pc.PickTargetCards {
		if id == pricey {
			t.Fatal("a permanent with a greater mana value was offered")
		}
	}
	effectstest.CloudDeckPickTargets(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, mine); c.Controller != opp.ID {
		t.Errorf("my permanent is controlled by %v, want the opponent", c.Controller)
	}
	if c := effectstest.FindBattlefieldCardByID(g, cheap); c.Controller != me.ID {
		t.Errorf("their permanent is controlled by %v, want me", c.Controller)
	}
	if c := effectstest.FindBattlefieldCardByID(g, pricey); c.Controller != opp.ID {
		t.Error("the unrelated permanent moved")
	}
}

func TestOldB19InvasionPlansForcesBlocksAndTheAttackerChoosesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active, other, defender := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.SeedEnchantment(g, oldB19Invasion, "Invasion Plans", oldB19Ench, other.ID)
	bear := effectstest.B12Creature(g, active.ID, "Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, defender.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.DeclareAttack(t, g, defender.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if got := g.ControllerOfPlayer(defender.ID); got != active.ID {
		t.Fatalf("the defender's block declaration is made by %v, want the attacking player", got)
	}
	if err := g.FinishBlocks(defender.ID); err == nil {
		t.Error("the defender's wall was allowed not to block")
	}
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("the attacking player's block: %v", err)
	}
	if err := g.FinishBlocks(defender.ID); err != nil {
		t.Fatal(err)
	}
	if c := effectstest.MustBattlefieldCard(t, g, wall); c.BlockingTarget != bear {
		t.Fatalf("the wall blocks %v, want the bear", c.BlockingTarget)
	}
}

func TestOldB19InvasionPlansEndsWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active, other, defender := g.Seats[0], g.Seats[1], g.Seats[2]
	plans := effectstest.SeedEnchantment(g, oldB19Invasion, "Invasion Plans", oldB19Ench, other.ID)
	bear := effectstest.B12Creature(g, active.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, defender.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.DestroyNow(t, g, plans)
	effectstest.DeclareAttack(t, g, defender.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if got := g.ControllerOfPlayer(defender.ID); got != defender.ID {
		t.Errorf("the defender's declaration is made by %v without Invasion Plans", got)
	}
}

func TestOldB19PossibilityStormExilesTheSpellAndLetsThemCastASharedTypeCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEnchantment(g, oldB19Storm, "Possibility Storm", oldB19Ench, me.ID)
	top := effectstest.HornTopFive(opp, []struct{ Name, TypeLine string }{
		{"Mountain", oldB19Mountain}, {"Divination", oldB19Sorcery}, {"Counterspell", oldB19Instant}, {"Bystander", oldB19Creature},
	})
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Zap", "", "{0}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.KeywordCastPromptFor(g, opp.ID)
	if prompt == nil {
		t.Fatal("no free-cast offer for the caster")
	}
	if effectstest.KeywordCastCard(prompt) != top[2] {
		t.Fatalf("offered %v, want the Counterspell (the first card that shares a type with the Instant)", effectstest.KeywordCastCard(prompt))
	}
	if g.Exile.Contains(top[3]) || !opp.Library.Contains(top[3]) {
		t.Error("the walk went past the hit")
	}
	if !g.Exile.Contains(spell) {
		t.Error("the spell is not in exile")
	}
	effectstest.AnswerMayCastPrompt(t, g, opp.ID, false)
	// Declined: the spell, the Mountain, the Divination and the hit are all on the library bottom.
	for _, id := range []uuid.UUID{spell, top[0], top[1], top[2]} {
		if g.Exile.Contains(id) || !opp.Library.Contains(id) {
			t.Errorf("card %s is not back in the library", id)
		}
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range opp.Library.Cards[:4] {
		bottom[c.InstanceID] = true
	}
	for _, id := range []uuid.UUID{spell, top[0], top[1], top[2]} {
		if !bottom[id] {
			t.Errorf("card %s is not among the bottom four", id)
		}
	}
}

func TestOldB19PossibilityStormCastHitLeavesTheRestOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEnchantment(g, oldB19Storm, "Possibility Storm", oldB19Ench, me.ID)
	top := effectstest.HornTopFive(opp, []struct{ Name, TypeLine string }{
		{"Mountain", oldB19Mountain}, {"Counterspell", oldB19Instant},
	})
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Zap", "", "{0}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCastPrompt(t, g, opp.ID, true)
	if !g.Stack.Contains(top[1]) {
		t.Error("the hit was not cast")
	}
	for _, id := range []uuid.UUID{spell, top[0]} {
		if !opp.Library.Contains(id) {
			t.Errorf("card %s is not back in the library", id)
		}
	}
}
