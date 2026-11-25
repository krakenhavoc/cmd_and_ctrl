package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b213_cards_test.go — Dominaria cards, S82 OLD batch 213 (#5530),
// played through the real catalog entries.

const (
	oldB213RunnerOracle    = "69e4e5bf-f0e2-4ed1-8b17-c8c7a9381551"
	oldB213TrapperOracle   = "745d37dd-d18e-47cc-81d0-8e61a07342f2"
	oldB213ThornOracle     = "9cdf840d-314e-46ef-9c04-74f145d9e65b"
	oldB213GiantOracle     = "e0efb48b-e7e5-4ed0-a7de-c8f72d357319"
	oldB213RadhaOracle     = "0fb90df6-a62b-4a9f-ae79-db71d40466ce"
	oldB213DruidOracle     = "519fccdb-fde4-42cb-b2ca-35fc36cf984a"
	oldB213ShardOracle     = "769c45ae-5d49-4815-944e-98790ad4efb0"
	oldB213SpiritOracle    = "1e640846-8f50-4a94-b7aa-f228124290d3"
	oldB213WallOracle      = "3c9fcda6-3ebc-4f9d-9f73-2bb4252d9420"
	oldB213SentinelOracle  = "71ecb454-3598-4860-bba1-b3932f906171"
	oldB213KickerSerpent   = "c39c1604-3bae-454d-9985-85101e51ec6e"
	oldB213ShardTypeLine   = "Artifact"
	oldB213LegendTypeLine  = "Legendary Creature — Human"
	oldB213PlainTypeLine   = "Creature — Bear"
	oldB213ArtifactSpellTL = "Artifact"
)

func TestOldB213CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB213RunnerOracle:   "Relic Runner",
		oldB213TrapperOracle:  "D'Avenant Trapper",
		oldB213ThornOracle:    "Thorn Elemental",
		oldB213GiantOracle:    "Two-Headed Giant",
		oldB213RadhaOracle:    "Grand Warlord Radha",
		oldB213DruidOracle:    "Elfhame Druid",
		oldB213ShardOracle:    "Powerstone Shard",
		oldB213SpiritOracle:   "Sanctum Spirit",
		oldB213WallOracle:     "Amaranthine Wall",
		oldB213SentinelOracle: "Drudge Sentinel",
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

func oldB213Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestOldB213PrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Grand Warlord Radha", "Legendary Creature — Elf Warrior", oldB213RadhaOracle, 3, 4, []string{"haste"}},
		{"Sanctum Spirit", "Creature — Spirit", oldB213SpiritOracle, 3, 2, []string{"lifelink"}},
		{"Amaranthine Wall", "Artifact Creature — Wall", oldB213WallOracle, 0, 6, []string{"defender"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
			effectstest.AssertKeywords(t, g, id, c.keywords...)
		})
	}
}

func TestOldB213RelicRunnerIsUnblockableOnlyAfterCastingAHistoricSpell(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		cast           bool
		unblockable    bool
	}{
		{"no spell", "", false, false},
		{"nonhistoric spell", oldB213PlainTypeLine, true, false},
		{"artifact spell", "Artifact", true, true},
		{"legendary spell", oldB213LegendTypeLine, true, true},
		{"saga spell", "Enchantment — Saga", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := oldB213Seats(g)
			runner := effectstest.B12Push(g, me.ID, "Relic Runner", "Creature — Human Rogue", oldB213RunnerOracle, 2, 1)
			blocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 1, 1)
			if tc.cast {
				effectstest.CastCatalogSpell(t, g, "Test Spell", tc.typeLine, "", nil)
				effectstest.PassPriorityAroundTable(t, g)
			}
			effectstest.BrAttack(t, g, runner)
			err := g.DeclareBlocker(blocker, runner)
			if tc.unblockable && err == nil {
				t.Fatal("the Runner was blocked after a historic spell")
			}
			if !tc.unblockable && err != nil {
				t.Fatalf("the Runner could not be blocked: %v", err)
			}
		})
	}
}

func TestOldB213DAvenantTrapperTapsAnOpposingCreatureWhenYouCastAHistoricSpell(t *testing.T) {
	for _, historic := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB213Seats(g)
		effectstest.B12Push(g, me.ID, "D'Avenant Trapper", "Creature — Human Archer", oldB213TrapperOracle, 3, 2)
		victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 2, 2)
		typeLine := oldB213PlainTypeLine
		if historic {
			typeLine = "Artifact"
		}
		effectstest.CastCatalogSpell(t, g, "Test Spell", typeLine, "", nil)
		if historic {
			effectstest.PickTriggerTarget(t, g, me.ID, victim)
		} else if effectstest.LatestPickTarget(g, me.ID) != nil {
			t.Fatal("a nonhistoric spell triggered the Trapper")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.B31Tapped(t, g, victim); got != historic {
			t.Errorf("historic=%v: victim tapped = %v", historic, got)
		}
	}
}

func TestOldB213ThornElementalMayAssignItsDamageAsThoughUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB213Seats(g)
	thorn := effectstest.B12Push(g, me.ID, "Thorn Elemental", "Creature — Elemental", oldB213ThornOracle, 7, 7)
	if !effectstest.ApaLive(g, thorn).Effective().AssignsAsThoughUnblocked {
		t.Fatal("Thorn Elemental does not assign its combat damage as though it weren't blocked")
	}
}

func TestOldB213TwoHeadedGiantFlipsTwoCoinsForDoubleStrikeOrMenace(t *testing.T) {
	seen := map[string]int{}
	for i := 0; i < 60; i++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB213Seats(g)
		giant := effectstest.B12Push(g, me.ID, "Two-Headed Giant", "Creature — Giant Warrior", oldB213GiantOracle, 4, 4)
		effectstest.DeclareAttack(t, g, opp.ID, giant)
		effectstest.PassPriorityAroundTable(t, g)
		var heads int
		var flips int
		for _, ev := range g.Events {
			if ev.Kind == game.EventFlipCoin {
				flips++
				if ev.Label == "heads" {
					heads++
				}
			}
		}
		if flips != 2 {
			t.Fatalf("flipped %d coins, want 2", flips)
		}
		ds := effectstest.HasEffectiveKeyword(t, g, giant, "double strike")
		menace := effectstest.HasEffectiveKeyword(t, g, giant, "menace")
		switch heads {
		case 2:
			seen["hh"]++
			if !ds || menace {
				t.Fatalf("two heads: double strike %v menace %v, want double strike only", ds, menace)
			}
		case 0:
			seen["tt"]++
			if ds || !menace {
				t.Fatalf("two tails: double strike %v menace %v, want menace only", ds, menace)
			}
		default:
			seen["ht"]++
			if ds || menace {
				t.Fatalf("one head: double strike %v menace %v, want neither", ds, menace)
			}
		}
	}
	if seen["hh"] == 0 || seen["tt"] == 0 || seen["ht"] == 0 {
		t.Errorf("60 trials never produced every outcome: %v", seen)
	}
}

func TestOldB213GrandWarlordRadhaAddsManaPerAttackerAndKeepsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB213Seats(g)
	radha := effectstest.B12Push(g, me.ID, "Grand Warlord Radha", "Legendary Creature — Elf Warrior", oldB213RadhaOracle, 3, 4)
	b1 := effectstest.PushVanillaCreature(g, me.ID, "Bear One", 2, 2)
	b2 := effectstest.PushVanillaCreature(g, me.ID, "Bear Two", 2, 2)
	effectstest.PushVanillaCreature(g, me.ID, "Stay Home", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, radha, b1, b2)
	effectstest.PassPriorityAroundTable(t, g)
	for effectstest.LatestChoiceOfKind(g, game.PendingChoiceMana) != nil {
		effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PoolSize(g, me.ID); got != 3 {
		t.Fatalf("pool = %d mana, want 3 (one per attacking creature)", got)
	}
	for color := range effectstest.MkColors(me) {
		if color != "R" && color != "G" {
			t.Errorf("mana of colour %q in the pool, want {R} and/or {G}", color)
		}
	}
	// "Until end of turn, you don't lose this mana as steps and phases end."
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if got := effectstest.PoolSize(g, me.ID); got != 3 {
		t.Errorf("pool after combat = %d, want 3 (kept through steps and phases)", got)
	}
}

func TestOldB213ElfhameDruidMakesGAndKickerOnlyGG(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	druid := effectstest.PushCatalogPermanent(g, me.ID, "Elfhame Druid", "Creature — Elf Druid", oldB213DruidOracle, false)
	effectstest.ActivateManaFor(t, g, me.ID, druid, 0, game.ManaAbilityParams{})
	effectstest.MkWantPool(t, me, map[string]int{"G": 1})
	g2, me2, _ := effectstest.SpendTable(t)
	druid2 := effectstest.PushCatalogPermanent(g2, me2.ID, "Elfhame Druid", "Creature — Elf Druid", oldB213DruidOracle, false)
	effectstest.ActivateManaFor(t, g2, me2.ID, druid2, 1, game.ManaAbilityParams{})
	effectstest.MkWantPool(t, me2, map[string]int{"G": 2})
	// The {G}{G} pays for a kicked spell but not for an unkicked one.
	for _, kicked := range []bool{false, true} {
		g3, me3, _ := effectstest.SpendTable(t)
		for i := 0; i < 2; i++ {
			d := effectstest.PushCatalogPermanent(g3, me3.ID, "Elfhame Druid", "Creature — Elf Druid", oldB213DruidOracle, false)
			effectstest.ActivateManaFor(t, g3, me3.ID, d, 1, game.ManaAbilityParams{})
		}
		var opt []int
		if kicked {
			opt = []int{0}
		}
		id := uuid.New()
		me3.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Serpent", TypeLine: "Creature — Serpent", ManaCost: "{G}",
			OracleID: oldB213KickerSerpent, Owner: me3.ID, Controller: me3.ID})
		err := g3.CastSpell(me3.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: opt})
		if kicked && err != nil {
			t.Errorf("the restricted mana could not pay for a kicked spell: %v", err)
		}
		if !kicked {
			effectstest.RefusedForMana(t, err, "the restricted mana for an unkicked spell")
		}
	}
}

func TestOldB213PowerstoneShardAddsOneColorlessPerShardYouControl(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	shard := effectstest.PushCatalogPermanent(g, me.ID, "Powerstone Shard", oldB213ShardTypeLine, oldB213ShardOracle, false)
	effectstest.ActivateManaFor(t, g, me.ID, shard, 0, game.ManaAbilityParams{})
	effectstest.MkWantPool(t, me, map[string]int{"C": 1})

	g2, me2, opp2 := effectstest.SpendTable(t)
	s1 := effectstest.PushCatalogPermanent(g2, me2.ID, "Powerstone Shard", oldB213ShardTypeLine, oldB213ShardOracle, false)
	effectstest.PushCatalogPermanent(g2, me2.ID, "Powerstone Shard", oldB213ShardTypeLine, oldB213ShardOracle, false)
	effectstest.PushCatalogPermanent(g2, opp2.ID, "Powerstone Shard", oldB213ShardTypeLine, oldB213ShardOracle, false)
	effectstest.PushCatalogPermanent(g2, me2.ID, "Other Rock", "Artifact", "", false)
	effectstest.ActivateManaFor(t, g2, me2.ID, s1, 0, game.ManaAbilityParams{})
	effectstest.MkWantPool(t, me2, map[string]int{"C": 2})
	_, _ = opp, g
}

func TestOldB213SanctumSpiritDiscardsAHistoricCardForIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB213Seats(g)
	effectstest.ToMain(t, g)
	spirit := effectstest.PushCatalogPermanent(g, me.ID, "Sanctum Spirit", "Creature — Spirit", oldB213SpiritOracle, false)
	plain := effectstest.PushHandCardOfType(me, "Plain Card", "Creature — Bear")
	if err := g.ActivateCatalogAbility(me.ID, spirit, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{plain}}); err == nil {
		t.Fatal("paid the cost with a card that is not historic")
	}
	historic := effectstest.PushHandCardOfType(me, "Old Relic", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, spirit, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{historic}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.InGraveyard(me, historic) {
		t.Error("the historic card was not discarded as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, spirit, "indestructible") {
		t.Error("the Spirit did not gain indestructible")
	}
	effectstest.EndTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, spirit, "indestructible") {
		t.Error("indestructible outlasted the turn")
	}
}

func TestOldB213AmaranthineWallPaysTwoForIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB213Seats(g)
	effectstest.ToMain(t, g)
	wall := effectstest.PushCatalogPermanent(g, me.ID, "Amaranthine Wall", "Artifact Creature — Wall", oldB213WallOracle, false)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, wall, "indestructible") {
		t.Error("the Wall did not gain indestructible")
	}
}

func TestOldB213DrudgeSentinelTapsItselfAndGainsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB213Seats(g)
	effectstest.ToMain(t, g)
	sentinel := effectstest.PushCatalogPermanent(g, me.ID, "Drudge Sentinel", "Creature — Skeleton Warrior", oldB213SentinelOracle, false)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sentinel, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B31Tapped(t, g, sentinel) {
		t.Error("the Sentinel is not tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, sentinel, "indestructible") {
		t.Error("the Sentinel did not gain indestructible")
	}
}
