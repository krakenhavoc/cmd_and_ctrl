package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b528_cards_test.go — Shadowmoor, S82 older sets batch 528 (#5591):
// pump and combat tricks.

const (
	shmSpawn    = "05ad7ec4-f48d-4cbe-be58-f871dc19702b"
	shmShield   = "0734272d-5cce-476d-a663-fbf0bd7cd3a7"
	shmSentry   = "22534fab-2c62-4113-bcea-f305ae9f664f"
	shmSeedcrad = "278ea996-3094-4a76-b6e1-db1d002e5502"
	shmEmber    = "44295009-9f85-4022-a685-a96079ba3e26"
	shmMerrow   = "4e9a63d2-b48b-4740-95ba-7966ada95baf"
	shmSafeDuo  = "76168d1b-266d-4d27-9e97-248b0c2c7e8f"
	shmNurturer = "78064301-f81e-4737-9cfe-b1e8757d0dd6"
	shmTatDuo   = "a56d13b5-a84d-4629-b9ec-4f407ee62596"
	shmGravel   = "a9e28134-382b-45e5-8fd0-cc0e7d56fd0c"
	shmTatWitch = "b1696989-7f00-41ec-82a4-fd10aed7a7f9"
	shmHorn     = "d80d87f5-af9e-404b-88e3-7a4b419dc32c"
	shmThistle  = "f6777aee-c39b-49d8-9049-5d2b1e888dd0"
)

func TestS82OldB528EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		shmSpawn: "Isleback Spawn", shmShield: "Kithkin Shielddare", shmSentry: "Safehold Sentry",
		shmSeedcrad: "Seedcradle Witch", shmEmber: "Emberstrike Duo", shmMerrow: "Merrow Grimeblotter",
		shmSafeDuo: "Safehold Duo", shmNurturer: "Nurturer Initiate", shmTatDuo: "Tattermunge Duo",
		shmGravel: "Gravelgill Duo", shmTatWitch: "Tattermunge Witch", shmHorn: "Revelsong Horn",
		shmThistle: "Thistledown Duo",
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

// shmCast casts a spell of the given colours from my hand (free) and settles the stack.
func shmCast(t *testing.T, g *game.Game, me *game.Player, name string, colors ...string) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", Colors: colors, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
}

func shmStats(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id)
}

func TestS82OldB528DuosPumpOnOneColourAndGrantOnTheOther(t *testing.T) {
	for _, row := range []struct {
		name, oracle, types string
		pump, grant, kw     string
		p, tough            int
	}{
		{"Emberstrike Duo", shmEmber, "Creature — Elemental Warrior Shaman", "B", "R", "first strike", 1, 1},
		{"Safehold Duo", shmSafeDuo, "Creature — Elf Warrior Shaman", "G", "W", "vigilance", 2, 4},
		{"Tattermunge Duo", shmTatDuo, "Creature — Goblin Warrior Shaman", "R", "G", "forestwalk", 2, 3},
		{"Gravelgill Duo", shmGravel, "Creature — Merfolk Rogue Warrior", "U", "B", "fear", 2, 1},
		{"Thistledown Duo", shmThistle, "Creature — Kithkin Soldier Wizard", "W", "U", "flying", 2, 2},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		duo := effectstest.B12Push(g, me.ID, row.name, row.types, row.oracle, row.p, row.tough)
		effectstest.AdvanceToMain(t, g)
		shmCast(t, g, me, "Pump colour", row.pump)
		if p, tough := shmStats(t, g, duo); p != row.p+1 || tough != row.tough+1 {
			t.Errorf("%s after a %s spell is %d/%d, want %d/%d", row.name, row.pump, p, tough, row.p+1, row.tough+1)
		}
		if got := effectstest.EffectiveAbilities(t, g, duo); effectstest.ContainsString(got, row.kw) {
			t.Errorf("%s has %s before a %s spell: %v", row.name, row.kw, row.grant, got)
		}
		shmCast(t, g, me, "Grant colour", row.grant)
		effectstest.AssertKeywords(t, g, duo, row.kw)
		if p, _ := shmStats(t, g, duo); p != row.p+1 {
			t.Errorf("%s: the grant colour changed power to %d", row.name, p)
		}
		// A gold spell of both colours triggers both abilities.
		shmCast(t, g, me, "Both", row.pump, row.grant)
		if p, _ := shmStats(t, g, duo); p != row.p+2 {
			t.Errorf("%s after a spell of both colours has power %d, want %d", row.name, p, row.p+2)
		}
	}
}

func TestS82OldB528IslebackSpawnGrowsWhenAnyLibraryIsTwentyOrFewer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spawn := effectstest.B12Push(g, me.ID, "Isleback Spawn", "Creature — Kraken", shmSpawn, 4, 8)
	effectstest.AdvanceToMain(t, g)
	trim := func(p *game.Player, n int) {
		g.WithWriteLock(func() {
			for len(p.Library.Cards) > n {
				p.Library.Cards = p.Library.Cards[1:]
			}
			for len(p.Library.Cards) < n {
				p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
			}
		})
	}
	for _, s := range g.Seats {
		trim(s, 30)
	}
	if p, tough := shmStats(t, g, spawn); p != 4 || tough != 8 {
		t.Fatalf("with every library at 30 the Spawn is %d/%d, want 4/8", p, tough)
	}
	trim(opp, 21)
	if p, _ := shmStats(t, g, spawn); p != 4 {
		t.Errorf("at 21 cards the Spawn has power %d, want 4", p)
	}
	// A draw takes the opponent's library to 20 cards: any library counts.
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	if p, tough := shmStats(t, g, spawn); p != 8 || tough != 16 {
		t.Errorf("with a library at 20 the Spawn is %d/%d, want 8/16", p, tough)
	}
	effectstest.AssertKeywords(t, g, spawn, "shroud")
	trim(opp, 25)
	if p, _ := shmStats(t, g, spawn); p != 4 {
		t.Errorf("after the library regrows the Spawn has power %d, want 4", p)
	}
}

func TestS82OldB528KithkinShielddarePumpsOnlyABlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	shield := effectstest.B12Push(g, opp.ID, "Kithkin Shielddare", "Creature — Kithkin Soldier", shmShield, 1, 1)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	effectstest.AddPool(t, g, opp.ID, "{W}{W}")
	if err := g.ActivateCatalogAbility(opp.ID, shield, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(attacker)}); err == nil {
		t.Fatal("targeted a creature that is not blocking")
	}
	effectstest.B16Activate(t, g, opp.ID, shield, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(blocker)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := shmStats(t, g, blocker); p != 4 || tough != 4 {
		t.Errorf("the blocker is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.B16Tapped(t, g, shield) {
		t.Error("{T} was not paid")
	}
}

func TestS82OldB528TattermungeWitchPumpsEveryBlockedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	witch := effectstest.B12Push(g, me.ID, "Tattermunge Witch", "Creature — Goblin Shaman", shmTatWitch, 2, 1)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	unblocked := effectstest.B12Creature(g, me.ID, "Unblocked", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{attacker, unblocked} {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatal(err)
	}
	effectstest.AddPool(t, g, me.ID, "{R}{G}")
	effectstest.B16Activate(t, g, me.ID, witch, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := shmStats(t, g, attacker); p != 3 {
		t.Errorf("the blocked attacker has power %d, want 3", p)
	}
	effectstest.AssertKeywords(t, g, attacker, "trample")
	if p, _ := shmStats(t, g, unblocked); p != 2 {
		t.Errorf("the unblocked attacker has power %d, want 2", p)
	}
	if p, _ := shmStats(t, g, blocker); p != 2 {
		t.Errorf("the blocker itself has power %d, want 2", p)
	}
}

func TestS82OldB528UntapSymbolCreaturesNeedToBeTapped(t *testing.T) {
	for _, row := range []struct {
		name, oracle, types, pool string
		p, tough                  int
		check                     func(t *testing.T, g *game.Game, self, other uuid.UUID)
	}{
		{"Safehold Sentry", shmSentry, "Creature — Elf Warrior", "{C}{C}{W}", 2, 2,
			func(t *testing.T, g *game.Game, self, _ uuid.UUID) {
				if p, tough := shmStats(t, g, self); p != 2 || tough != 4 {
					t.Errorf("Sentry is %d/%d, want 2/4", p, tough)
				}
			}},
		{"Merrow Grimeblotter", shmMerrow, "Creature — Merfolk Wizard", "{C}{U}", 2, 2,
			func(t *testing.T, g *game.Game, _, other uuid.UUID) {
				if p, _ := shmStats(t, g, other); p != 0 {
					t.Errorf("the target has power %d, want 0", p)
				}
			}},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		self := effectstest.B12Push(g, me.ID, row.name, row.types, row.oracle, row.p, row.tough)
		other := effectstest.B12Creature(g, g.Seats[1].ID, "Target", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		params := game.ActivateAbilityParams{Strict: true}
		if row.oracle == shmMerrow {
			params.Targets = effectstest.CardRefs(other)
		}
		effectstest.AddPool(t, g, me.ID, row.pool)
		if err := g.ActivateCatalogAbility(me.ID, self, 0, params); err == nil {
			t.Fatalf("%s: activated {Q} while untapped", row.name)
		}
		effectstest.TapForTest(t, g, self)
		effectstest.B16Activate(t, g, me.ID, self, 0, params)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.B16Tapped(t, g, self) {
			t.Errorf("%s: {Q} did not untap it", row.name)
		}
		row.check(t, g, self, other)
	}
}

func TestS82OldB528SeedcradleWitchPumpsAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	witch := effectstest.B12Push(g, me.ID, "Seedcradle Witch", "Creature — Elf Shaman", shmSeedcrad, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, bear)
	effectstest.AdvanceToMain(t, g)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{G}{W}")
	effectstest.B16Activate(t, g, me.ID, witch, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := shmStats(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the Bear was not untapped")
	}
}

func TestS82OldB528NurturerInitiateOffersThePumpOnAnyPlayersGreenSpell(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Nurturer Initiate", "Creature — Elf Shaman", shmNurturer, 1, 1)
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		effectstest.AddPool(t, g, me.ID, "{C}")
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Green Spell", TypeLine: "Instant", Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatal(err)
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			effectstest.B17PickCards(t, g, me.ID, bear)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.HasPayUnlessFor(g, me.ID) {
			effectstest.AnswerPayUnless(t, g, me.ID, pay)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := 2
		if pay {
			want = 3
		}
		if p, _ := shmStats(t, g, bear); p != want {
			t.Errorf("pay=%v: the Bear has power %d, want %d", pay, p, want)
		}
	}
}

func TestS82OldB528RevelsongHornTapsACreatureAndPumpsATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	horn := effectstest.B12Push(g, me.ID, "Revelsong Horn", "Artifact", shmHorn, 0, 0)
	payer := effectstest.B12Creature(g, me.ID, "Payer", "Creature — Bear", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.AddPool(t, g, me.ID, "{C}")
	effectstest.B16Activate(t, g, me.ID, horn, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear),
		TapIDs: []uuid.UUID{payer}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := shmStats(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.B16Tapped(t, g, payer) || !effectstest.B16Tapped(t, g, horn) {
		t.Error("the Horn and the paying creature should both be tapped")
	}
}
