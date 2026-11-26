package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b137_test.go — Theros Beyond Death, S82 set OLD batch 137 (#5514).
// Every test plays the real catalog entry through the engine.

const (
	oldB137ElspethUH   = "3f9a797b-85e7-474e-aefe-7e2ec08fc5b7"
	oldB137AshiokSoF   = "ce79bbd2-edf0-4d5e-9aaf-717ec3960c48"
	oldB137AshiokNM    = "ebc9ff52-c914-4c48-9d35-0e4cd07741cb"
	oldB137DenyDivine  = "0146de73-29bf-415a-b450-11074553f715"
	oldB137Intervene   = "f7c89d03-606e-4284-90dd-64b1fbebfb4d"
	oldB137Erasure     = "13cf3ef7-d88e-4cc7-a4d8-cedc671f356a"
	oldB137Drag        = "746dc1d8-f245-47be-abd5-50ca6ac0d1af"
	oldB137MirrorShld  = "57f2858e-5793-41fc-9bbb-1fdd3b99fe79"
	oldB137Wings       = "8b56efee-bdc8-4516-9c0d-b28ca1b0a918"
	oldB137NyleaKE     = "642bbdf6-9a8c-4158-83c8-d59feb49101b"
	oldB137PurphorosBB = "6728b9d0-d3ab-4ba3-af13-63cb83e0ed99"
	oldB137Dalakos     = "211814a5-195d-41c6-b96c-4c26b2cbb33c"
)

func TestOldB137EveryCardIsRegisteredWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB137ElspethUH: "Elspeth, Undaunted Hero", oldB137AshiokSoF: "Ashiok, Sculptor of Fears",
		oldB137AshiokNM: "Ashiok, Nightmare Muse", oldB137DenyDivine: "Deny the Divine",
		oldB137Intervene: "Thassa's Intervention", oldB137Erasure: "Ashiok's Erasure",
		oldB137Drag: "Drag to the Underworld", oldB137MirrorShld: "Mirror Shield",
		oldB137Wings: "Wings of Hubris", oldB137NyleaKE: "Nylea, Keen-Eyed",
		oldB137PurphorosBB: "Purphoros, Bronze-Blooded", oldB137Dalakos: "Dalakos, Crafter of Wonders",
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

// oldB137Cast casts a catalog spell from the active player's hand, strictly
// paid from `pool`, and returns the announce error (the spell is left on the
// stack on success).
func oldB137Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle, pool string, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.FloatForTest(g, me, pool)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID,
		})
	})
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

func oldB137Ref(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// --- Deny the Divine ---------------------------------------------------------

func TestOldB137DenyTheDivineExilesACreatureOrEnchantmentSpell(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		legal          bool
	}{
		{"Grizzly Bears", "Creature — Bear", true},
		{"Pacifism", "Enchantment — Aura", true},
		{"Shock", "Instant", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			spell := effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, "", nil)
			err := effectstest.CastCatalogSpellErr(t, g, "Deny the Divine", "Instant", oldB137DenyDivine, []game.TargetRef{oldB137Ref(spell)})
			if tc.legal != (err == nil) {
				t.Fatalf("target legality: err %v, want legal=%v", err, tc.legal)
			}
			if !tc.legal {
				return
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Exile.Contains(spell) {
				t.Error("the countered spell is not in exile")
			}
		})
	}
}

// --- Drag to the Underworld ---------------------------------------------------

func TestOldB137DragToTheUnderworldCostsLessByDevotionToBlack(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rocks int
		pool  string
		ok    bool
	}{
		{"no devotion needs the whole cost", 0, "BB", false},
		{"no devotion, paid in full", 0, "BBBB", true},
		{"devotion two takes off both generic", 1, "BB", true},
		{"devotion never touches the coloured part", 3, "B", false},
		{"a devotion above the generic still pays {B}{B}", 3, "BB", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			for i := 0; i < tc.rocks; i++ {
				effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Black Rock", TypeLine: "Artifact", ManaCost: "{B}{B}"})
			}
			victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			_, err := oldB137Cast(t, g, "Drag to the Underworld", "Instant", "{2}{B}{B}", oldB137Drag, tc.pool, []game.TargetRef{oldB137Ref(victim)})
			if tc.ok != (err == nil) {
				t.Fatalf("cast with %q: err %v, want ok=%v", tc.pool, err, tc.ok)
			}
			if !tc.ok {
				return
			}
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(victim) {
				t.Error("the target creature survived")
			}
		})
	}
}

// --- Mirror Shield -------------------------------------------------------------

func TestOldB137MirrorShieldPumpsGrantsHexproofAndKillsDeathtouchBlockers(t *testing.T) {
	for _, tc := range []struct {
		name      string
		deathtch  bool
		wantAlive bool
	}{{"a blocker with deathtouch is destroyed", true, false}, {"an ordinary blocker is left alone", false, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			bear := effectstest.SeedBear(g, me.ID)
			shield := effectstest.SeedEquipment(g, me.ID, "Mirror Shield", oldB137MirrorShld)
			effectstest.EquipTo(t, g, me.ID, shield, bear)
			if p, th := effectstest.PtOf(t, g, bear); p != 2 || th != 4 {
				t.Fatalf("equipped Bear is %d/%d, want 2/4", p, th)
			}
			if !effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
				t.Error("the equipped creature has no hexproof")
			}
			blocker := game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4}
			if tc.deathtch {
				blocker.Keywords = []string{"deathtouch"}
			}
			wall := effectstest.PushTestPermanent(g, opp.ID, blocker)
			effectstest.DeclareAttack(t, g, opp.ID, bear)
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			if err := g.DeclareBlocker(wall, bear); err != nil {
				t.Fatalf("DeclareBlocker: %v", err)
			}
			effectstest.LockInBlocks(t, g)
			want := 0
			if tc.deathtch {
				want = 1
			}
			if n := effectstest.TriggersOnStackFrom(g, shield); n != want {
				t.Fatalf("%d Mirror Shield triggers on the stack, want %d", n, want)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if alive := g.Battlefield.Contains(wall); alive != tc.wantAlive {
				t.Errorf("blocker on the battlefield = %v, want %v", alive, tc.wantAlive)
			}
			if !g.Battlefield.Contains(bear) {
				t.Error("the equipped creature died")
			}
		})
	}
}

func TestOldB137MirrorShieldTriggersWhenTheEquippedCreatureIsBlockedByDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	// The opponent's creature wears the Shield and blocks a deathtouch attacker.
	wall := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4})
	shield := effectstest.SeedEquipment(g, opp.ID, "Mirror Shield", oldB137MirrorShld)
	effectstest.FindBattlefieldCardForTest(g, shield).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: wall}
	snake := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Snake", TypeLine: "Creature — Snake", Power: 1, Toughness: 1, Keywords: []string{"deathtouch"}})
	effectstest.DeclareAttack(t, g, opp.ID, snake)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, snake); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(snake) {
		t.Error("the deathtouch attacker the equipped creature blocked was not destroyed")
	}
}

// --- Wings of Hubris -----------------------------------------------------------

func TestOldB137WingsOfHubrisGivesFlyingThenMakesTheCreatureUnblockableAndSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	wings := effectstest.SeedEquipment(g, me.ID, "Wings of Hubris", oldB137Wings)
	effectstest.EquipTo(t, g, me.ID, wings, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Fatal("the equipped creature has no flying")
	}
	wall := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"reach"}})
	if err := g.ActivateCatalogAbility(me.ID, wings, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("sacrifice the Wings: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wings) {
		t.Fatal("the Wings were not sacrificed as the cost")
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err == nil {
		t.Error("the equipped creature was blocked")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("the creature was not sacrificed at the beginning of the next end step")
	}
}

func TestOldB137WingsOfHubrisAttachedToNothingDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	wings := effectstest.SeedEquipment(g, me.ID, "Wings of Hubris", oldB137Wings)
	if err := g.ActivateCatalogAbility(me.ID, wings, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("sacrifice the Wings: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("an unattached Wings sacrificed a creature")
	}
}

// --- Nylea, Keen-Eyed ----------------------------------------------------------

func TestOldB137NyleaKeenEyedIsACreatureOnlyAtFiveDevotionAndDiscountsCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	nylea := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Nylea, Keen-Eyed", OracleID: oldB137NyleaKE, ManaCost: "{3}{G}",
		TypeLine: "Legendary Enchantment Creature — God", Power: 5, Toughness: 6})
	if effectstest.ApaLive(g, nylea).IsCreature() {
		t.Fatal("Nylea is a creature with devotion 1")
	}
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Green Rock", TypeLine: "Artifact", ManaCost: "{G}{G}{G}{G}"})
	if !effectstest.ApaLive(g, nylea).IsCreature() {
		t.Fatal("Nylea is not a creature with devotion 5")
	}
	// A {2}{G} creature spell costs {1}{G}; a noncreature spell is not discounted.
	for _, tc := range []struct {
		name, typeLine, pool string
		ok                   bool
	}{
		{"Bear", "Creature — Bear", "G", false},
		{"Bear", "Creature — Bear", "GG", true},
		{"Growth", "Sorcery", "GG", false},
		{"Growth", "Sorcery", "GGG", true},
	} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Nylea, Keen-Eyed", OracleID: oldB137NyleaKE, ManaCost: "{3}{G}",
			TypeLine: "Legendary Enchantment Creature — God", Power: 5, Toughness: 6})
		if _, err := oldB137Cast(t, g, tc.name, tc.typeLine, "{2}{G}", "", tc.pool, nil); (err == nil) != tc.ok {
			t.Errorf("%s with %q: err %v, want ok=%v", tc.name, tc.pool, err, tc.ok)
		}
	}
}

func TestOldB137NyleaKeenEyedRevealsTheTopCard(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost string
		toHand               bool
		bin                  bool
	}{
		{"Bear", "Creature — Bear", "{1}{G}", true, false},
		{"Forest", "Basic Land — Forest", "", false, true},
		{"Forest", "Basic Land — Forest", "", false, false},
	} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		nylea := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Nylea, Keen-Eyed", OracleID: oldB137NyleaKE, ManaCost: "{3}{G}",
			TypeLine: "Legendary Enchantment Creature — God", Power: 5, Toughness: 6})
		top := effectstest.PlTop(me, tc.name, tc.typeLine, tc.cost)
		effectstest.ApaMana(me, "G", "G", "G")
		effectstest.ApaActivate(t, g, me, nylea, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
		if !tc.toHand {
			effectstest.AnswerMayChoice(t, g, me.ID, tc.bin)
		}
		if me.Hand.Contains(top) != tc.toHand {
			t.Errorf("%s: in hand = %v, want %v", tc.name, me.Hand.Contains(top), tc.toHand)
		}
		if !tc.toHand && me.Graveyard.Contains(top) != tc.bin {
			t.Errorf("%s, answer %v: in graveyard = %v", tc.name, tc.bin, me.Graveyard.Contains(top))
		}
		if !tc.toHand && !tc.bin {
			if got, err := me.Library.Top(); err != nil || got.InstanceID != top {
				t.Errorf("declined, but the card is not on top")
			}
		}
	}
}

// --- Purphoros, Bronze-Blooded --------------------------------------------------

func TestOldB137PurphorosGivesHasteAndPutsARedOrArtifactCreatureFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	god := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Purphoros, Bronze-Blooded", OracleID: oldB137PurphorosBB, ManaCost: "{4}{R}",
		TypeLine: "Legendary Enchantment Creature — God", Power: 7, Toughness: 6})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectstest.ApaLive(g, god).IsCreature() {
		t.Fatal("Purphoros is a creature with devotion 1")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("another creature has no haste")
	}
	red := effectstest.RfdHandCard(me, "Red Drake", "Creature — Drake", "{3}{R}", "", 3, 3)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == red {
				me.Hand.Cards[i].Colors = []string{"R"}
			}
		}
	})
	golem := effectstest.RfdHandCard(me, "Rock Golem", "Artifact Creature — Golem", "{3}", "", 3, 3)
	green := effectstest.RfdHandCard(me, "Green Bear", "Creature — Bear", "{1}{G}", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == green {
				me.Hand.Cards[i].Colors = []string{"G"}
			}
		}
	})
	effectstest.ApaMana(me, "R", "R", "R")
	effectstest.ApaActivate(t, g, me, god, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the card to put onto the battlefield")
	}
	if effectstest.HasID(pick.ChooseCards, green) || !effectstest.HasID(pick.ChooseCards, red) || !effectstest.HasID(pick.ChooseCards, golem) {
		t.Errorf("offered %v: want the red creature and the artifact creature only", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, golem)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(golem) {
		t.Fatal("the artifact creature did not enter")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) {
		t.Error("the creature was not sacrificed at the beginning of the next end step")
	}
}

func TestOldB137PurphorosDecliningMakesNoSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	god := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Purphoros, Bronze-Blooded", OracleID: oldB137PurphorosBB, ManaCost: "{4}{R}",
		TypeLine: "Legendary Enchantment Creature — God", Power: 7, Toughness: 6})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.RfdHandCard(me, "Rock Golem", "Artifact Creature — Golem", "{3}", "", 3, 3)
	effectstest.ApaMana(me, "R", "R", "R")
	effectstest.ApaActivate(t, g, me, god, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("declining put a sacrifice on a creature")
	}
}

// --- Dalakos, Crafter of Wonders -------------------------------------------------

func TestOldB137DalakosMakesRestrictedManaAndGrantsFlyingAndHasteToEquippedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dal := effectstest.B12Push(g, me.ID, "Dalakos, Crafter of Wonders", "Legendary Creature — Merfolk Artificer", oldB137Dalakos, 2, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.UntapForTest(g, dal)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == dal {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, dal, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Dalakos: %v", err)
	}
	if n := me.ManaPool.CountColor("C"); n != 2 {
		t.Errorf("{C} in pool = %d, want 2", n)
	}
	cost, _ := game.ParseCost("{2}")
	rock := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{2}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rock)) || !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(rock)) {
		t.Error("the mana refused an artifact spell or an artifact's ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) || me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("the mana paid for a nonartifact spell or ability")
	}

	equipped := effectstest.SeedBear(g, me.ID)
	bare := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, g.Seats[1].ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Mirror Shield", oldB137MirrorShld)
	effectstest.EquipTo(t, g, me.ID, blade, equipped)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, equipped, kw) {
			t.Errorf("the equipped creature has no %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, bare, kw) {
			t.Errorf("an unequipped creature has %s", kw)
		}
	}
	theirBlade := effectstest.SeedEquipment(g, g.Seats[1].ID, "Mirror Shield", oldB137MirrorShld)
	effectstest.FindBattlefieldCardForTest(g, theirBlade).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: theirs}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("an opponent's equipped creature got flying")
	}
}

// --- Thassa's Intervention ---------------------------------------------------------

func TestOldB137ThassasInterventionLooksAtTheTopXAndTakesUpToTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PlTop(me, "Card A", "Sorcery", "{1}")
	b := effectstest.PlTop(me, "Card B", "Sorcery", "{1}")
	c := effectstest.PlTop(me, "Card C", "Sorcery", "{1}")
	d := effectstest.PlTop(me, "Card D", "Sorcery", "{1}")
	_, err := oldB137CastModes(t, g, "Thassa's Intervention", "Instant", "{X}{U}{U}", oldB137Intervene, "UUU", 1, []int{0}, nil)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// X = 1: only the top card (D) is looked at.
	effectstest.AnswerChooseCards(t, g, me.ID, d)
	if !me.Hand.Contains(d) || me.Hand.Contains(c) || me.Hand.Contains(b) || me.Hand.Contains(a) {
		t.Errorf("hand after X=1: want only the top card")
	}
}

func TestOldB137ThassasInterventionTakesTwoOfThreeAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PlTop(me, "Card A", "Sorcery", "{1}")
	b := effectstest.PlTop(me, "Card B", "Sorcery", "{1}")
	c := effectstest.PlTop(me, "Card C", "Sorcery", "{1}")
	size := me.Library.Size()
	if _, err := oldB137CastModes(t, g, "Thassa's Intervention", "Instant", "{X}{U}{U}", oldB137Intervene, "UUUUU", 3, []int{0}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, c, b)
	if !me.Hand.Contains(c) || !me.Hand.Contains(b) || me.Hand.Contains(a) {
		t.Error("the two picked cards are not in hand, or the third is")
	}
	if me.Library.Size() != size-2 {
		t.Errorf("library %d, want %d", me.Library.Size(), size-2)
	}
	bottom, err := me.Library.Bottom()
	if err != nil || bottom.InstanceID != a {
		t.Errorf("the unpicked card is not on the bottom (%v)", err)
	}
}

func TestOldB137ThassasInterventionCountersUnlessTheyPayTwiceX(t *testing.T) {
	for _, tc := range []struct {
		name     string
		pool     string
		wantGone bool
	}{{"unable to pay {2}", "", true}, {"able to pay {2}", "CC", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			spell := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
				[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
			if _, err := oldB137CastModes(t, g, "Thassa's Intervention", "Instant", "{X}{U}{U}", oldB137Intervene, "UUU", 1, []int{1}, []game.TargetRef{oldB137Ref(spell)}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			if tc.pool != "" {
				effectstest.FloatForTest(g, me, tc.pool)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerPayUnless(t, g, me.ID, tc.pool != "")
			z := g.FindCardZoneForEffect(spell)
			gone := z == nil || z.Kind == game.ZoneGraveyard
			if gone != tc.wantGone {
				t.Errorf("Bolt countered = %v, want %v", gone, tc.wantGone)
			}
		})
	}
}

// oldB137CastModes casts an X modal instant with the given modes, paid from
// `pool`.
func oldB137CastModes(t *testing.T, g *game.Game, name, typeLine, cost, oracle, pool string, x int, modes []int, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.FloatForTest(g, me, pool)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID,
		})
	})
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: x, Modes: modes, Targets: targets})
}

// --- Elspeth, Undaunted Hero ----------------------------------------------------

func TestOldB137ElspethPlusTwoPutsACounterOnUpToTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	elspeth := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Elspeth, Undaunted Hero", "Legendary Planeswalker — Elspeth", oldB137ElspethUH, 5)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, elspeth, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{oldB137Ref(a), oldB137Ref(b), oldB137Ref(c)}}); err == nil {
		t.Fatal("three targets were accepted for \"up to two\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, elspeth, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{oldB137Ref(a), oldB137Ref(c)}}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyCount(g, elspeth); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
	if effectstest.B12Counter(t, g, a, game.CounterPlusOne) != 1 || effectstest.B12Counter(t, g, c, game.CounterPlusOne) != 1 {
		t.Error("a target did not get its +1/+1 counter")
	}
	if effectstest.B12Counter(t, g, b, game.CounterPlusOne) != 0 {
		t.Error("an unchosen creature got a counter")
	}
}

func TestOldB137ElspethMinusTwoFindsSunlitHoplite(t *testing.T) {
	for _, tc := range []struct {
		name         string
		zone         string
		libraryShuff bool
	}{{"from the library, which shuffles", game.SearchZoneLibrary, true}, {"from the graveyard, which does not", game.SearchZoneGraveyard, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.ToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			elspeth := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Elspeth, Undaunted Hero", "Legendary Planeswalker — Elspeth", oldB137ElspethUH, 5)
			hoplite := game.Card{Name: "Sunlit Hoplite", TypeLine: "Creature — Human Soldier", ManaCost: "{1}{W}", Power: 2, Toughness: 1}
			other := game.Card{Name: "Sunlit Rider", TypeLine: "Creature — Human Soldier", ManaCost: "{1}{W}", Power: 2, Toughness: 1}
			ids := effectstest.SeedSearchLibrary(me, other)
			effectstest.LgKnowLibrary(me)
			var want uuid.UUID
			if tc.zone == game.SearchZoneLibrary {
				want = effectstest.PushLibraryCardForTest(me, hoplite)
			} else {
				hoplite.Owner, hoplite.Controller, hoplite.InstanceID = me.ID, me.ID, uuid.New()
				want = hoplite.InstanceID
				me.Graveyard.PushTop(hoplite)
			}
			_ = ids
			if err := g.ActivateCatalogAbility(me.ID, elspeth, 1, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("-2: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerSearchZones(t, g, me.ID, tc.zone)
			effectstest.AnswerSearchByID(t, g, me.ID, want)
			if !g.Battlefield.Contains(want) {
				t.Fatal("Sunlit Hoplite did not enter")
			}
			shuffled := effectstest.LibraryKnownTo(me, me.ID) == 0
			if shuffled != tc.libraryShuff {
				t.Errorf("library shuffled = %v, want %v", shuffled, tc.libraryShuff)
			}
			if got := effectstest.LoyaltyCount(g, elspeth); got != 3 {
				t.Errorf("loyalty %d, want 3", got)
			}
		})
	}
}

func TestOldB137ElspethMinusEightPumpsByDevotionAndGrantsFlying(t *testing.T) {
	// X is read as the ability resolves (CR 608.2h): with 8 loyalty she is gone by then
	// (CR 704.5i) and counts nothing; with 9 she stays and her {W}{W}{W} count.
	for _, tc := range []struct{ loyalty, want int }{{8, 4}, {9, 7}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMain(t, g)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		elspeth := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Elspeth, Undaunted Hero", "Legendary Planeswalker — Elspeth", oldB137ElspethUH, tc.loyalty)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == elspeth {
					g.Battlefield.Cards[i].ManaCost = "{2}{W}{W}{W}"
				}
			}
		})
		effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "White Rock", TypeLine: "Artifact", ManaCost: "{W}{W}"})
		mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		if err := g.ActivateCatalogAbility(me.ID, elspeth, 2, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("-8: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p, th := effectstest.PtOf(t, g, mine); p != tc.want || th != tc.want {
			t.Errorf("loyalty %d: your Bear is %d/%d, want %d/%d", tc.loyalty, p, th, tc.want, tc.want)
		}
		if !effectstest.HasEffectiveKeyword(t, g, mine, "flying") {
			t.Error("your creature did not gain flying")
		}
		if p, _ := effectstest.PtOf(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "flying") {
			t.Error("an opponent's creature was affected")
		}
	}
}

// --- Ashiok, Sculptor of Fears --------------------------------------------------------

func TestOldB137AshiokSculptorPlusTwoDrawsAndEveryoneMills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ashiok := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Ashiok, Sculptor of Fears", "Legendary Planeswalker — Ashiok", oldB137AshiokSoF, 4)
	libs := map[uuid.UUID]int{}
	hand := me.Hand.Size()
	for _, p := range g.Seats {
		libs[p.ID] = p.Library.Size()
	}
	if err := g.ActivateCatalogAbility(me.ID, ashiok, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	for _, p := range g.Seats {
		want := libs[p.ID] - 2
		if p.ID == me.ID {
			want--
		}
		if p.Library.Size() != want {
			t.Errorf("a library has %d cards, want %d", p.Library.Size(), want)
		}
	}
	if got := effectstest.LoyaltyCount(g, ashiok); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestOldB137AshiokSculptorMinusFiveReanimatesFromAnyGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	ashiok := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Ashiok, Sculptor of Fears", "Legendary Planeswalker — Ashiok", oldB137AshiokSoF, 5)
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Beast", "Creature — Beast", "{3}")
	land := effectstest.PushGraveyardPermanent(opp, "Their Rock", "Artifact", "{2}")
	if err := g.ActivateCatalogAbility(me.ID, ashiok, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB137Ref(land)}}); err == nil {
		t.Fatal("a noncreature card was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, ashiok, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB137Ref(theirs)}}); err != nil {
		t.Fatalf("-5: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, theirs)
	if !ok {
		t.Fatal("the creature did not enter")
	}
	if c.Controller != me.ID {
		t.Error("the creature is not under your control")
	}
}

func TestOldB137AshiokSculptorMinusElevenTakesEveryCreatureTheOpponentControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp, third := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4], g.Seats[(g.Turn.ActiveSeat+2)%4]
	ashiok := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Ashiok, Sculptor of Fears", "Legendary Planeswalker — Ashiok", oldB137AshiokSoF, 11)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, third.ID, "Other Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, ashiok, 2, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("yourself was accepted as the target opponent")
	}
	if err := g.ActivateCatalogAbility(me.ID, ashiok, 2, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("-11: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if c, _ := effectstest.BattlefieldCard(g, id); c.Controller != me.ID {
			t.Errorf("%s is not under your control", c.Name)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, other); c.Controller != third.ID {
		t.Error("a creature of another opponent was taken")
	}
	if c, _ := effectstest.BattlefieldCard(g, rock); c.Controller != opp.ID {
		t.Error("a noncreature permanent was taken")
	}
	// No stated duration: it outlasts the turn.
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, a); c.Controller != me.ID {
		t.Error("the creature went home at end of turn")
	}
}

// --- Ashiok, Nightmare Muse -----------------------------------------------------------

func TestOldB137AshiokNightmareMuseMakesANightmareThatExilesOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	ashiok := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Ashiok, Nightmare Muse", "Legendary Planeswalker — Ashiok", oldB137AshiokNM, 5)
	if err := g.ActivateCatalogAbility(me.ID, ashiok, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var token game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Nightmare" {
			token = c
		}
	}
	if token.InstanceID == uuid.Nil {
		t.Fatal("no Nightmare token")
	}
	if p, th := effectstest.PtOf(t, g, token.InstanceID); p != 2 || th != 3 {
		t.Errorf("Nightmare is %d/%d, want 2/3", p, th)
	}
	if len(token.Colors) != 2 {
		t.Errorf("Nightmare colours %v, want blue and black", token.Colors)
	}
	// Attack with it (it is summoning sick, so let it be a creature that has been there).
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == token.InstanceID {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	lib := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lib[p.ID] = p.Library.Size()
	}
	exile := g.Exile.Size()
	effectstest.DeclareAttack(t, g, opp.ID, token.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lib[p.ID]
		if p.ID != me.ID {
			want -= 2
		}
		if p.Library.Size() != want {
			t.Errorf("a library has %d cards, want %d", p.Library.Size(), want)
		}
	}
	if got := g.Exile.Size() - exile; got != 6 {
		t.Errorf("%d cards exiled, want 6 (three opponents, two each)", got)
	}
}

func TestOldB137AshiokNightmareMuseMinusThreeBouncesThenTheOwnerExilesFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	ashiok := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Ashiok, Nightmare Muse", "Legendary Planeswalker — Ashiok", oldB137AshiokNM, 5)
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	keep := effectstest.RfdHandCard(opp, "Kept Card", "Sorcery", "{1}", "", 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, ashiok, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB137Ref(land)}}); err == nil {
		t.Fatal("a land was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, ashiok, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB137Ref(rock)}}); err != nil {
		t.Fatalf("-3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Fatal("the permanent was not returned to its owner's hand")
	}
	pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("its owner was not asked to exile a card from their hand")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, keep)
	if opp.Hand.Contains(keep) || !g.Exile.Contains(keep) {
		t.Error("the chosen card was not exiled")
	}
	if !opp.Hand.Contains(rock) {
		t.Error("the returned permanent was exiled instead")
	}
}

func TestOldB137AshiokNightmareMuseMinusSevenCastsUpToThreeSpellsFromOpponentsExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	ashiok := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Ashiok, Nightmare Muse", "Legendary Planeswalker — Ashiok", oldB137AshiokNM, 7)
	exiled := func(owner uuid.UUID, name, typeLine string, faceDown bool) uuid.UUID {
		id := uuid.New()
		g.Exile.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: effectstest.LightningBoltOracle,
			ManaCost: "{R}", Owner: owner, Controller: owner, FaceDown: faceDown})
		return id
	}
	bolts := []uuid.UUID{
		exiled(opp.ID, "Lightning Bolt", "Instant", false), exiled(opp.ID, "Lightning Bolt", "Instant", false),
		exiled(g.Seats[(g.Turn.ActiveSeat+2)%4].ID, "Lightning Bolt", "Instant", false), exiled(opp.ID, "Lightning Bolt", "Instant", false),
	}
	hidden := exiled(opp.ID, "Hidden Bolt", "Instant", true)
	mine := exiled(me.ID, "My Bolt", "Instant", false)
	land := exiled(opp.ID, "Island", "Basic Land — Island", false)
	if err := g.ActivateCatalogAbility(me.ID, ashiok, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-7: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, bolts...)
	for _, bad := range []uuid.UUID{hidden, mine, land} {
		if effectstest.HasID(offer.ChooseCards, bad) {
			t.Error("a card that is not a face-up nonland card an opponent owns was offered")
		}
	}
	target := game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, bolts[0], target); err != nil {
		t.Fatalf("first free cast: %v", err)
	}
	target.FromZone = "exile"
	for i, id := range bolts[1:3] {
		if err := g.CastSpell(me.ID, id, target); err != nil {
			t.Fatalf("free cast %d: %v", i+2, err)
		}
	}
	if err := g.CastSpell(me.ID, bolts[3], target); err == nil {
		t.Error("a fourth spell was cast")
	}
	for _, bad := range []uuid.UUID{hidden, mine, land} {
		if err := g.CastSpell(me.ID, bad, target); err == nil {
			t.Error("a card that is not a face-up nonland card an opponent owns was cast")
		}
	}
}

func TestOldB137AshiokNightmareMuseMinusSevenDecliningWithdrawsTheCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	ashiok := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Ashiok, Nightmare Muse", "Legendary Planeswalker — Ashiok", oldB137AshiokNM, 7)
	bolt := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle,
		ManaCost: "{R}", Owner: opp.ID, Controller: opp.ID})
	if err := g.ActivateCatalogAbility(me.ID, ashiok, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-7: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, bolt)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "exile", Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Error("a card was cast after the offer was declined")
	}
}

// --- Ashiok's Erasure -----------------------------------------------------------------------

func TestOldB137AshioksErasureExilesTheSpellBansItsNameAndReturnsItWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1; g.Turn.PriorityHolder = 1 })
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle,
		ManaCost: "{R}", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}}); err != nil {
		t.Fatalf("cast Bolt: %v", err)
	}
	g.WithWriteLock(func() { g.Turn.PriorityHolder = 0 })
	erasure := effectstest.CastFromSeat(t, g, me, "Ashiok's Erasure", "Enchantment", oldB137Erasure, nil)
	effectstest.PassUntilOnBattlefield(t, g, erasure)
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bolt) {
		t.Fatal("the spell was not exiled")
	}
	if me.Life != 40 {
		t.Errorf("the exiled spell resolved: life %d", me.Life)
	}
	sameName := game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle, Owner: opp.ID, Controller: opp.ID}
	otherName := game.Card{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID}
	if err := effectstest.CastGateFor(g, opp.ID, sameName, game.ZoneHand); err == nil {
		t.Error("an opponent cast a spell with the exiled card's name")
	}
	if err := effectstest.CastGateFor(g, opp.ID, otherName, game.ZoneHand); err != nil {
		t.Errorf("an opponent was barred from a different name: %v", err)
	}
	mineSame := sameName
	mineSame.Owner, mineSame.Controller = me.ID, me.ID
	if err := effectstest.CastGateFor(g, me.ID, mineSame, game.ZoneHand); err != nil {
		t.Errorf("its controller was barred: %v", err)
	}
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{erasure}, game.DestroyOptions{}) })
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bolt) {
		t.Error("the exiled card did not return to its owner's hand")
	}
	if err := effectstest.CastGateFor(g, opp.ID, sameName, game.ZoneHand); err != nil {
		t.Errorf("the ban outlived the enchantment: %v", err)
	}
}
