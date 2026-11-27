package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b478_cards_test.go — Rise of the Eldrazi: Level Up, batch 478
// (#5581, S82 OLD). Every test plays the real catalog entry through the engine.

const (
	oldB478Zulaport  = "0ab5fde8-e7db-4265-81c6-2a6488ba56c4"
	oldB478Wavewatch = "44f01491-cce4-442e-a8d9-345a074805bd"
	oldB478Outrider  = "4c03e0db-ce26-41f6-a12c-861a919e52fb"
	oldB478Cutthroat = "759774d2-abcf-4fc6-8433-5f251e80c70e"
	oldB478Lord      = "d369a02b-c282-454a-9f89-480f4f5d838c"
	oldB478Adept     = "d63725af-be6d-4f91-b5c8-04df82d6f805"
	oldB478Champion  = "dfd2846f-574b-475c-bf2a-dc3e87d13476"
	oldB478Kargan    = "eb4e6d97-f5df-4290-b193-912f44f43a67"
	oldB478Purists   = "ecc3e2ac-6534-4350-92f4-2c1f578d854e"
)

func TestOldB478CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB478Zulaport:  "Zulaport Enforcer",
		oldB478Wavewatch: "Halimar Wavewatch",
		oldB478Outrider:  "Ikiral Outrider",
		oldB478Cutthroat: "Nirkana Cutthroat",
		oldB478Lord:      "Lord of Shatterskull Pass",
		oldB478Adept:     "Skywatcher Adept",
		oldB478Champion:  "Null Champion",
		oldB478Kargan:    "Kargan Dragonlord",
		oldB478Purists:   "Hedron-Field Purists",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

func oldB478Level(t *testing.T, g *game.Game, id uuid.UUID, n int) {
	t.Helper()
	if n == 0 {
		return
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterLevel, n); err != nil {
			t.Fatalf("level counters: %v", err)
		}
	})
}

// TestOldB478LevelBands walks every card through its printed level symbols:
// size and keywords at the edge of each band.
func TestOldB478LevelBands(t *testing.T) {
	type step struct {
		level, power, toughness int
		keywords, absent        []string
	}
	for _, tc := range []struct {
		name, oracle, typ string
		power, toughness  int
		steps             []step
	}{
		{"Zulaport Enforcer", oldB478Zulaport, "Creature — Human Warrior", 1, 1, []step{{0, 1, 1, nil, nil}, {1, 3, 3, nil, nil}, {2, 3, 3, nil, nil}, {3, 5, 5, nil, nil}, {9, 5, 5, nil, nil}}},
		{"Halimar Wavewatch", oldB478Wavewatch, "Creature — Merfolk Soldier", 0, 3, []step{{0, 0, 3, nil, []string{"islandwalk"}}, {1, 0, 6, nil, []string{"islandwalk"}}, {4, 0, 6, nil, []string{"islandwalk"}}, {5, 6, 6, []string{"islandwalk"}, nil}}},
		{"Ikiral Outrider", oldB478Outrider, "Creature — Human Soldier", 1, 2, []step{{0, 1, 2, nil, []string{"vigilance"}}, {1, 2, 6, []string{"vigilance"}, nil}, {3, 2, 6, []string{"vigilance"}, nil}, {4, 3, 10, []string{"vigilance"}, nil}}},
		{"Nirkana Cutthroat", oldB478Cutthroat, "Creature — Vampire Warrior", 3, 2, []step{{0, 3, 2, nil, []string{"deathtouch", "first strike"}}, {1, 4, 3, []string{"deathtouch"}, []string{"first strike"}}, {2, 4, 3, []string{"deathtouch"}, []string{"first strike"}}, {3, 5, 4, []string{"deathtouch", "first strike"}, nil}}},
		{"Skywatcher Adept", oldB478Adept, "Creature — Merfolk Wizard", 1, 1, []step{{0, 1, 1, nil, []string{"flying"}}, {1, 2, 2, []string{"flying"}, nil}, {2, 2, 2, []string{"flying"}, nil}, {3, 4, 2, []string{"flying"}, nil}}},
		{"Lord of Shatterskull Pass", oldB478Lord, "Creature — Minotaur Shaman", 3, 3, []step{{0, 3, 3, nil, nil}, {1, 6, 6, nil, nil}, {5, 6, 6, nil, nil}, {6, 6, 6, nil, nil}}},
		{"Null Champion", oldB478Champion, "Creature — Zombie Warrior", 1, 1, []step{{0, 1, 1, nil, nil}, {1, 4, 2, nil, nil}, {3, 4, 2, nil, nil}, {4, 7, 3, nil, nil}}},
		{"Kargan Dragonlord", oldB478Kargan, "Creature — Human Warrior", 2, 2, []step{{0, 2, 2, nil, []string{"flying", "trample"}}, {3, 2, 2, nil, []string{"flying", "trample"}}, {4, 4, 4, []string{"flying"}, []string{"trample"}}, {7, 4, 4, []string{"flying"}, []string{"trample"}}, {8, 8, 8, []string{"flying", "trample"}, nil}}},
		{"Hedron-Field Purists", oldB478Purists, "Creature — Human Cleric", 0, 3, []step{{0, 0, 3, nil, nil}, {1, 1, 4, nil, nil}, {4, 1, 4, nil, nil}, {5, 2, 5, nil, nil}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, s := range tc.steps {
				g := effectstest.NewCatalogGame(t)
				me := g.Seats[0]
				id := effectstest.B12Push(g, me.ID, tc.name, tc.typ, tc.oracle, tc.power, tc.toughness)
				oldB478Level(t, g, id, s.level)
				if p, tough := effectstest.EffectivePTOf(t, g, id); p != s.power || tough != s.toughness {
					t.Errorf("level %d: %d/%d, want %d/%d", s.level, p, tough, s.power, s.toughness)
				}
				for _, kw := range s.keywords {
					if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
						t.Errorf("level %d: no %s", s.level, kw)
					}
				}
				for _, kw := range s.absent {
					if effectstest.HasEffectiveKeyword(t, g, id, kw) {
						t.Errorf("level %d: has %s too early", s.level, kw)
					}
				}
			}
		})
	}
}

func TestOldB478LevelUpIsSorcerySpeedAndCostsTheLevelCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Kargan Dragonlord", "Creature — Human Warrior", oldB478Kargan, 2, 2)
	effectstest.AddPool(t, g, me.ID, "{R}{R}")
	for i := 1; i <= 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("level up %d for {R}: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.B12Counter(t, g, id, game.CounterLevel); n != i {
			t.Fatalf("%d level counters, want %d", n, i)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a level up was paid for with no mana left")
	}
}

func TestOldB478ZulaportEnforcerIsBlockedOnlyByBlackCreaturesAtLevelThree(t *testing.T) {
	for _, tc := range []struct {
		level     int
		blackWall bool
		legal     bool
	}{{2, false, true}, {3, false, false}, {3, true, true}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		enforcer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Zulaport Enforcer",
			TypeLine: "Creature — Human Warrior", OracleID: oldB478Zulaport, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Keywords: []string{"haste"}})
		oldB478Level(t, g, enforcer, tc.level)
		// A black blocker is always on the table, so the declaration step is
		// not skipped for want of a legal blocker.
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Black Wall", TypeLine: "Creature — Wall",
			Colors: []string{"B"}, Power: 0, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
		card := game.Card{InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Owner: opp.ID, Controller: opp.ID}
		if tc.blackWall {
			card.Colors = []string{"B"}
		} else {
			card.Colors = []string{"G"}
		}
		wall := effectstest.PushBattlefieldCardWithTimestamp(g, card)
		effectstest.ReqToBlockers(t, g, enforcer)
		err := effectstest.ReqBlock(t, g, enforcer, wall)
		if tc.legal && err != nil {
			t.Errorf("level %d, black %v: refused: %v", tc.level, tc.blackWall, err)
		}
		if !tc.legal && err == nil {
			t.Errorf("level %d, black %v: a nonblack creature blocked it", tc.level, tc.blackWall)
		}
	}
}

func TestOldB478LordOfShatterskullPassSweepsOnlyAtLevelSix(t *testing.T) {
	for _, tc := range []struct {
		level int
		dead  bool
	}{{5, false}, {6, true}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		lord := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Lord of Shatterskull Pass",
			TypeLine: "Creature — Minotaur Shaman", OracleID: oldB478Lord, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID, Keywords: []string{"haste"}})
		oldB478Level(t, g, lord, tc.level)
		victim := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		effectstest.DeclareAttack(t, g, opp.ID, lord)
		effectstest.PassPriorityAroundTable(t, g)
		if got := !g.Battlefield.Contains(victim); got != tc.dead {
			t.Errorf("level %d: the defending player's 5/5 died = %v, want %v", tc.level, got, tc.dead)
		}
		if !g.Battlefield.Contains(mine) {
			t.Errorf("level %d: the attacker's own creature died", tc.level)
		}
	}
}

func TestOldB478NullChampionRegeneratesOnlyAtLevelFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Null Champion", "Creature — Zombie Warrior", oldB478Champion, 1, 1)
	oldB478Level(t, g, id, 3)
	effectstest.AddPool(t, g, me.ID, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Ref: "grant:null-champion/regenerate:0:0", Strict: true}); err == nil {
		t.Fatal("a level 3 Null Champion regenerated")
	}
	oldB478Level(t, g, id, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Ref: "grant:null-champion/regenerate:0:0", Strict: true}); err != nil {
		t.Fatalf("regenerate at level 4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.RegenShieldsOn(t, g, id); n != 1 {
		t.Errorf("%d regeneration shields, want 1", n)
	}
}

func TestOldB478KarganDragonlordPumpsOnlyAtLevelEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Kargan Dragonlord", "Creature — Human Warrior", oldB478Kargan, 2, 2)
	oldB478Level(t, g, id, 7)
	effectstest.AddPool(t, g, me.ID, "{R}{R}{R}")
	params := game.ActivateAbilityParams{Ref: "grant:kargan-dragonlord/pump:0:0", Strict: true}
	if err := g.ActivateCatalogAbility(me.ID, id, 1, params); err == nil {
		t.Fatal("a level 7 Dragonlord pumped")
	}
	oldB478Level(t, g, id, 1)
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, id, 1, params); err != nil {
			t.Fatalf("pump %d at level 8: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 10 || tough != 8 {
		t.Errorf("twice pumped is %d/%d, want 10/8", p, tough)
	}
}

func TestOldB478HedronFieldPuristsPreventsDamageToYouAndYourCreatures(t *testing.T) {
	for _, tc := range []struct {
		level, prevent int
	}{{0, 0}, {1, 1}, {4, 1}, {5, 2}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		purists := effectstest.B12Push(g, opp.ID, "Hedron-Field Purists", "Creature — Human Cleric", oldB478Purists, 0, 3)
		oldB478Level(t, g, purists, tc.level)
		ally := effectstest.B12Creature(g, opp.ID, "Ally", "Creature — Bear", 2, 9)
		other := effectstest.B12Creature(g, me.ID, "Not Theirs", "Creature — Bear", 2, 9)
		life := opp.Life
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		if got, want := life-opp.Life, 3-tc.prevent; got != want {
			t.Errorf("level %d: the player lost %d life, want %d", tc.level, got, want)
		}
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTarget(ally))
		effectstest.PassPriorityAroundTable(t, g)
		if got, want := effectstest.DamageOn(t, g, ally), 3-tc.prevent; got != want {
			t.Errorf("level %d: your creature was dealt %d damage, want %d", tc.level, got, want)
		}
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTarget(other))
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.DamageOn(t, g, other); got != 3 {
			t.Errorf("level %d: another player's creature was dealt %d damage, want 3", tc.level, got)
		}
	}
}
