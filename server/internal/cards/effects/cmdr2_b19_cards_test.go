package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b19_cards_test.go — Commander wave 2, batch 19 (#5368): simple Auras.

const (
	cmdr2VowOfDuty      = "1e7bddbd-e448-4aea-9d18-b44bd45837c5"
	cmdr2VowOfLightning = "bfed67f3-70d2-49c2-8f25-16e4d10fa5be"
)

func TestCmdr2B19EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		cmdr2VowOfDuty:          "Vow of Duty",
		cmdr2VowOfLightning:     "Vow of Lightning",
		cmdr2VerdantEmbrace:     "Verdant Embrace",
		cmdr2PollenbrightWings:  "Pollenbright Wings",
		cmdr2WarbriarBlessing:   "Warbriar Blessing",
		cmdr2SpectralSteel:      "Spectral Steel",
		cmdr2CarefulCultivation: "Careful Cultivation",
		cmdr2IndomitableMight:   "Indomitable Might",
		cmdr2DaybreakCoronet:    "Daybreak Coronet",
		cmdr2TenuousTruce:       "Tenuous Truce",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// cmdr2B19Attach puts an Aura of the catalog onto the battlefield attached
// to host, controlled by owner.
func cmdr2B19Attach(g *game.Game, owner, host uuid.UUID, name, oracle string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: auraTypeLine,
		OracleID: oracle, Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}})
}

func TestCmdr2B19VowsPumpGrantAndBarAttackingTheirController(t *testing.T) {
	for _, tc := range []struct{ name, oracle, kw string }{
		{"Vow of Duty", cmdr2VowOfDuty, "vigilance"},
		{"Vow of Lightning", cmdr2VowOfLightning, "first strike"},
	} {
		g := newCatalogGame(t)
		me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
		host := b12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
		other := b12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
		cmdr2B19Attach(g, me.ID, host, tc.name, tc.oracle)
		if p, to := effectivePTOf(t, g, host); p != 4 || to != 4 {
			t.Errorf("%s: host is %d/%d, want 4/4", tc.name, p, to)
		}
		if !hasEffectiveKeyword(t, g, host, tc.kw) {
			t.Errorf("%s: host lacks %s", tc.name, tc.kw)
		}
		var walker uuid.UUID
		g.WithWriteLock(func() {
			walker = pushPermanent(g, me.ID, game.Card{Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
				Counters: map[string]int{game.CounterLoyalty: 3}})
		})
		advanceToDeclareAttackersOf(t, g, 1)
		for _, target := range []uuid.UUID{me.ID, walker} {
			if err := g.Clone().DeclareAttacker(host, target); !errors.Is(err, game.ErrIllegalAttackTarget) {
				t.Errorf("%s: enchanted creature attacking %s: err = %v, want ErrIllegalAttackTarget", tc.name, target, err)
			}
		}
		if err := g.Clone().DeclareAttacker(host, third.ID); err != nil {
			t.Errorf("%s: enchanted creature attacking another player: %v", tc.name, err)
		}
		if err := g.Clone().DeclareAttacker(other, me.ID); err != nil {
			t.Errorf("%s: an unenchanted creature attacking me: %v", tc.name, err)
		}
	}
}

const (
	cmdr2VerdantEmbrace    = "f0bbd988-50b7-476c-9183-dfabd4a95079"
	cmdr2PollenbrightWings = "582c8030-1d41-4c63-8ed6-06aad398a96c"
)

func TestCmdr2B19VerdantEmbracePumpsAndMakesSaprolingsForTheHostsControllerEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	cmdr2B19Attach(g, me.ID, host, "Verdant Embrace", cmdr2VerdantEmbrace)
	if p, to := effectivePTOf(t, g, host); p != 5 || to != 5 {
		t.Fatalf("host is %d/%d, want 5/5", p, to)
	}
	advanceTo(t, g, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Saproling", opp.ID); n != 1 {
		t.Errorf("%d Saprolings for the enchanted creature's controller on my upkeep, want 1", n)
	}
	if n := namedOnBattlefield(g, "Saproling", me.ID); n != 0 {
		t.Errorf("%d Saprolings for the Aura's controller, want 0", n)
	}
}

func TestCmdr2B19PollenbrightWingsGivesFlyingAndMakesThatManySaprolings(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 3, 3)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	cmdr2B19Attach(g, me.ID, host, "Pollenbright Wings", cmdr2PollenbrightWings)
	if !hasEffectiveKeyword(t, g, host, "flying") {
		t.Fatal("the enchanted creature lacks flying")
	}
	dealCombatDamageToPlayer(g, other, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Saproling", me.ID); n != 0 {
		t.Fatalf("%d Saprolings after an unenchanted creature connected", n)
	}
	dealCombatDamageToPlayer(g, host, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Saproling", me.ID); n != 3 {
		t.Errorf("%d Saprolings, want 3 (the damage dealt)", n)
	}
}

const cmdr2WarbriarBlessing = "4a58954a-1d20-48b8-847c-3615de67e122"

func TestCmdr2B19WarbriarBlessingPumpsToughnessAndFightsOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	ecl6Cast(t, g, me, "Warbriar Blessing", "Enchantment — Aura", cmdr2WarbriarBlessing, 0, 0,
		game.CastSpellParams{Targets: cardRefs(mine)})
	ecl6ResolveSpell(t, g)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, mine); p != 2 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want 2/4", p, tt)
	}
	if len(battlefieldIDsNamed(g, "Theirs")) != 0 {
		t.Error("the opposing creature survived the fight")
	}
	if len(battlefieldIDsNamed(g, "Mine")) != 1 {
		t.Error("the enchanted creature died")
	}
}

func TestCmdr2B19WarbriarBlessingOnlyEnchantsYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Warbriar Blessing", TypeLine: "Enchantment — Aura",
			OracleID: cmdr2WarbriarBlessing, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(theirs)}); err == nil {
		t.Error("enchanting an opponent's creature was allowed")
	}
}

const cmdr2SpectralSteel = "95f74c0d-3f89-4944-997e-1bdc408bae71"

func TestCmdr2B19SpectralSteelPumpsAndReturnsAnotherAuraOrEquipmentFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	cmdr2B19Attach(g, me.ID, host, "Spectral Steel", cmdr2SpectralSteel)
	if p, tt := ptOf(t, g, host); p != 4 || tt != 4 {
		t.Fatalf("enchanted creature is %d/%d, want 4/4", p, tt)
	}
	toMainForCost(t, g)
	steel := pushCatalogGraveyardCard(me, "Spectral Steel", "Enchantment — Aura", cmdr2SpectralSteel, 0, 0)
	aura := pushCatalogGraveyardCard(me, "Gy Aura", "Enchantment — Aura", "", 0, 0)
	gear := pushCatalogGraveyardCard(me, "Gy Blade", "Artifact — Equipment", "", 0, 0)
	bear := pushCatalogGraveyardCard(me, "Gy Bear", "Creature — Bear", "", 2, 2)
	floatForTest(g, me, "WC")
	for _, bad := range []uuid.UUID{steel, bear} {
		if err := g.ActivateCatalogAbility(me.ID, steel, 0, game.ActivateAbilityParams{Targets: cardTarget(bad)}); err == nil {
			t.Fatalf("%s was accepted; the target is another Aura or Equipment card", bad)
		}
	}
	b16Activate(t, g, me.ID, steel, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(aura)})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("the Aura card did not return to hand")
	}
	if !g.Exile.Contains(steel) {
		t.Error("Spectral Steel was not exiled")
	}
	_ = gear
}

const cmdr2CarefulCultivation = "83959c3d-053c-4d95-a1b0-fe92f2b906db"

func TestCmdr2B19CarefulCultivationOnACreatureAndOnANoncreatureArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	cmdr2B19Attach(g, me.ID, host, "Careful Cultivation", cmdr2CarefulCultivation)
	cmdr2B19Attach(g, me.ID, rock, "Careful Cultivation", cmdr2CarefulCultivation)
	if p, tt := ptOf(t, g, host); p != 3 || tt != 5 {
		t.Errorf("enchanted creature is %d/%d, want 3/5", p, tt)
	}
	if !hasEffectiveKeyword(t, g, host, "reach") {
		t.Error("the enchanted creature lacks reach")
	}
	if grantedManaRowsOf(g, host) != 1 {
		t.Error("the enchanted creature has no granted mana ability")
	}
	if hasEffectiveKeyword(t, g, rock, "reach") || grantedManaRowsOf(g, rock) != 0 {
		t.Error("an enchanted artifact that is not a creature got reach or the mana ability")
	}
	if c, _ := battlefieldCard(g, rock); c.Power != 0 || c.Toughness != 0 {
		t.Errorf("the noncreature artifact got a pump: %d/%d", c.Power, c.Toughness)
	}
}

func TestCmdr2B19CarefulCultivationRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Careful Cultivation", TypeLine: "Enchantment — Aura",
			OracleID: cmdr2CarefulCultivation, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(land)}); err == nil {
		t.Error("enchanting a land was allowed")
	}
}

func TestCmdr2B19CarefulCultivationChannelsFromHandIntoAManaMonk(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	toMainForCost(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Careful Cultivation", TypeLine: "Enchantment — Aura",
			OracleID: cmdr2CarefulCultivation, Owner: me.ID, Controller: me.ID})
	})
	floatForTest(g, me, "GC")
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(id) || !me.Graveyard.Contains(id) {
		t.Error("the card was not discarded")
	}
	if n := namedOnBattlefield(g, "Human Monk", me.ID); n != 1 {
		t.Fatalf("%d Human Monk tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Human Monk" {
			continue
		}
		var abs []game.ManaAbilityShape
		g.ReadSnapshot(func() {
			card, _ := g.LookupCardForEffect(c.InstanceID)
			abs, _ = game.ManaAbilitiesWithOrigins(card)
		})
		if len(abs) != 1 || abs[0].Produced != "{G}" || !abs[0].TapCost {
			t.Errorf("Human Monk mana abilities = %+v, want {T}: Add {G}", abs)
		}
	}
}

const cmdr2IndomitableMight = "069c3c87-81a8-4d2a-ad2f-8d3243d0e3db"

func TestCmdr2B19IndomitableMightLetsABlockedCreatureAssignAllDamageToThePlayer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	host := apaPush(g, g.Seats[0].ID, g.Seats[0].ID, game.Card{Name: "Host", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	cmdr2B19Attach(g, g.Seats[2].ID, host, "Indomitable Might", cmdr2IndomitableMight)
	wall := mscWlBlocker(g, "Wall", 1, 8)
	mscWlBlockedCombat(t, g, []uuid.UUID{host}, map[uuid.UUID]uuid.UUID{wall: host})

	c := mscWlDamagePrompt(t, g)
	f := c.DamageAssignment
	if !f.AsThoughUnblocked || len(f.BlockerIDs) != 1 || f.AttackerPower != 4 {
		t.Fatalf("the prompt is %+v, want one blocker, power 4 (1+3), as-though-unblocked", f)
	}
	if err := mscWlAnswer(t, g, c, 4, game.DamageAssignmentEntry{BlockerID: wall, Amount: 0}); err != nil {
		t.Fatalf("assigning all 4 to the player: %v", err)
	}
	if lost := 40 - opp.Life; lost != 4 {
		t.Errorf("the defender lost %d, want 4", lost)
	}
}

func TestCmdr2B19IndomitableMightHasFlash(t *testing.T) {
	spec, ok := Lookup(cmdr2IndomitableMight)
	if !ok || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Fatalf("Indomitable Might's printed keywords = %v, want flash", spec.PrintedKeywords)
	}
}

const cmdr2DaybreakCoronet = "0f5ac364-447b-4c43-9aed-6098dc0bab82"

func cmdr2B19CastCoronet(g *game.Game, me *game.Player, host uuid.UUID) (uuid.UUID, error) {
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Daybreak Coronet", TypeLine: "Enchantment — Aura", ManaCost: "{W}{W}",
			Colors: []string{"W"}, OracleID: cmdr2DaybreakCoronet, Owner: me.ID, Controller: me.ID})
	})
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(host)})
}

func TestCmdr2B19DaybreakCoronetNeedsAnotherAuraAndFallsOffWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bare := b12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	other := cmdr2B19Attach(g, me.ID, host, "Other Aura", "")
	advanceToMain(t, g)
	floatForTest(g, me, "WW")
	if _, err := cmdr2B19CastCoronet(g, me, bare); err == nil {
		t.Fatal("Daybreak Coronet was cast on a creature with no other Aura")
	}
	id, err := cmdr2B19CastCoronet(g, me, host)
	if err != nil {
		t.Fatalf("Daybreak Coronet on a creature with another Aura: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, id); !ok || !c.IsAttachedTo(host) {
		t.Fatal("the Coronet did not enter attached to the creature")
	}
	if p, tt := ptOf(t, g, host); p != 5 || tt != 5 {
		t.Errorf("enchanted creature is %d/%d, want 5/5", p, tt)
	}
	for _, kw := range []string{"first strike", "vigilance", "lifelink"} {
		if !hasEffectiveKeyword(t, g, host, kw) {
			t.Errorf("enchanted creature lacks %s", kw)
		}
	}
	// The other Aura leaves: the Coronet has nothing "another" to enchant with.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, other); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if _, ok := battlefieldCard(g, id); ok {
		t.Error("the Coronet stayed on a creature with no other Aura attached")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the Coronet did not go to its owner's graveyard")
	}
}

const cmdr2TenuousTruce = "f0bf2923-aea2-4bac-b061-64bc84437ef1"

// cmdr2B19Truce puts a Tenuous Truce controlled by me onto the battlefield
// attached to the player.
func cmdr2B19Truce(t *testing.T, g *game.Game, me, enchanted *game.Player) uuid.UUID {
	t.Helper()
	aura := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Tenuous Truce", TypeLine: auraTypeLine,
		OracleID: cmdr2TenuousTruce, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetPlayer, ID: enchanted.ID}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	return aura
}

func TestCmdr2B19TenuousTruceDrawsForBothOnTheEnchantedOpponentsEndStepOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	seedLibrary(me, "a", "b", "c")
	seedLibrary(opp, "a", "b", "c")
	seedLibrary(third, "a", "b", "c")
	cmdr2B19Truce(t, g, me, opp)
	advanceToMainOf(t, g, 1)
	mid := [2]int{len(me.Hand.Cards), len(opp.Hand.Cards)}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - mid[0]; got != 1 {
		t.Errorf("I drew %d on the enchanted opponent's end step, want 1", got)
	}
	if got := len(opp.Hand.Cards) - mid[1]; got != 1 {
		t.Errorf("the enchanted opponent drew %d on their end step, want 1", got)
	}
}

func TestCmdr2B19TenuousTruceRefusesAnythingButAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Tenuous Truce", TypeLine: "Enchantment — Aura", ManaCost: "{1}{W}",
			Colors: []string{"W"}, OracleID: cmdr2TenuousTruce, Owner: me.ID, Controller: me.ID})
	})
	floatForTest(g, me, "WC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Error("enchanting myself was allowed")
	}
}

func TestCmdr2B19TenuousTruceIsSacrificedWhenYouAttackTheEnchantedOpponentOrTheyAttackYou(t *testing.T) {
	type attack struct {
		name           string
		attackerSeat   int
		defenderPlayer int // seat attacked; -1 means the planeswalker named by walkerOf
		walkerOf       int // seat whose planeswalker is attacked
		sacrificed     bool
	}
	for _, tc := range []attack{
		{"I attack the enchanted opponent", 0, 1, 0, true},
		{"the enchanted opponent attacks me", 1, 0, 0, true},
		{"I attack another opponent", 0, 2, 0, false},
		{"the enchanted opponent attacks another player", 1, 2, 0, false},
		{"another opponent attacks the enchanted opponent", 2, 1, 0, false},
		{"I attack the enchanted opponent's planeswalker", 0, -1, 1, true},
		{"the enchanted opponent attacks my planeswalker", 1, -1, 0, true},
		{"I attack another opponent's planeswalker", 0, -1, 2, false},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		aura := cmdr2B19Truce(t, g, me, opp)
		attacker := b12Creature(g, g.Seats[tc.attackerSeat].ID, "Attacker", "Creature — Bear", 2, 2)
		defender := uuid.Nil
		if tc.defenderPlayer >= 0 {
			defender = g.Seats[tc.defenderPlayer].ID
		} else {
			g.WithWriteLock(func() {
				defender = pushPermanent(g, g.Seats[tc.walkerOf].ID, game.Card{Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
					Counters: map[string]int{game.CounterLoyalty: 3}})
			})
		}
		advanceToDeclareAttackersOf(t, g, tc.attackerSeat)
		if err := g.DeclareAttacker(attacker, defender); err != nil {
			t.Fatalf("%s: DeclareAttacker: %v", tc.name, err)
		}
		lockInAttacks(t, g)
		passPriorityAroundTable(t, g)
		_, still := battlefieldCard(g, aura)
		if still == tc.sacrificed {
			t.Errorf("%s: Aura on the battlefield = %v, want %v", tc.name, still, !tc.sacrificed)
		}
	}
}
