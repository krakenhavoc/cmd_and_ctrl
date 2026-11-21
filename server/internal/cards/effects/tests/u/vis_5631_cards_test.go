package testsu

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- every card is in the catalog ----------------------------------------

func TestVis5631EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		effectstest.VisArchangelOracle, effectstest.VisFreewindFalconOracle, effectstest.VisLongbowArcherOracle, effectstest.VisPantherWarriorsOracle,
		effectstest.VisPythonOracle, effectstest.VisScalebanesEliteOracle, effectstest.VisTempestDrakeOracle, effectstest.VisWarthogOracle,
		effectstest.VisSuqAtaLancerOracle, effectstest.VisFallenAskariOracle, effectstest.VisHulkingCyclopsOracle, effectstest.VisRiverBoaOracle,
		effectstest.VisWakeOfVulturesOracle, effectstest.VisDarajaGriffinOracle, effectstest.VisViashivanDragonOracle, effectstest.VisKeeperOfKookusOracle,
		effectstest.VisJamuraanLionOracle, effectstest.VisInfantryVeteranOracle, effectstest.VisSisaysRingOracle, effectstest.VisRainbowEfreetOracle,
		effectstest.VisTeferisHonorGuardOracle, effectstest.VisCreepingMoldOracle, effectstest.VisElvenCacheOracle, effectstest.VisInspirationOracle,
		effectstest.VisUndoOracle, effectstest.VisWarriorsHonorOracle, effectstest.VisMiraculousRecoveryOracle, effectstest.VisRetributionOfTheMeekOracle,
		effectstest.VisTremorOracle, effectstest.VisManOWarOracle, effectstest.VisUktabiOrangutanOracle, effectstest.VisNekrataalOracle,
		effectstest.VisCoralAtollOracle, effectstest.VisDormantVolcanoOracle, effectstest.VisEvergladesOracle, effectstest.VisJungleBasinOracle, effectstest.VisKarooOracle,
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

// --- the keyword bodies ----------------------------------------------------

func TestVis5631KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Archangel", "Creature — Angel", effectstest.VisArchangelOracle, []string{"flying", "vigilance"}},
		{"Tempest Drake", "Creature — Drake", effectstest.VisTempestDrakeOracle, []string{"flying", "vigilance"}},
		{"Freewind Falcon", "Creature — Bird", effectstest.VisFreewindFalconOracle, []string{"flying", "protection from red"}},
		{"Longbow Archer", "Creature — Human Soldier Archer", effectstest.VisLongbowArcherOracle, []string{"reach", "first strike"}},
		{"Scalebane's Elite", "Creature — Human Soldier", effectstest.VisScalebanesEliteOracle, []string{"protection from black"}},
		{"Warthog", "Creature — Boar", effectstest.VisWarthogOracle, []string{"swampwalk"}},
		{"Suq'Ata Lancer", effectstest.VisHumanKnightTypeLine, effectstest.VisSuqAtaLancerOracle, []string{"haste"}},
		{"River Boa", "Creature — Snake", effectstest.VisRiverBoaOracle, []string{"islandwalk"}},
		{"Wake of Vultures", "Creature — Bird", effectstest.VisWakeOfVulturesOracle, []string{"flying"}},
		{"Daraja Griffin", "Creature — Griffin", effectstest.VisDarajaGriffinOracle, []string{"flying"}},
		{"Viashivan Dragon", "Creature — Dragon", effectstest.VisViashivanDragonOracle, []string{"flying"}},
		{"Rainbow Efreet", "Creature — Efreet", effectstest.VisRainbowEfreetOracle, []string{"flying"}},
		{"Nekrataal", "Creature — Human Assassin", effectstest.VisNekrataalOracle, []string{"first strike"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Panther Warriors and Python are vanilla: no keyword and no ability rows.
func TestVis5631VanillaBodies(t *testing.T) {
	for _, oracle := range []string{effectstest.VisPantherWarriorsOracle, effectstest.VisPythonOracle} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not in the catalog", oracle)
		}
		if len(spec.PrintedKeywords) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 || len(spec.Triggered) != 0 {
			t.Errorf("%s declares rules: %+v", spec.Name, spec)
		}
	}
}

// Freewind Falcon's protection from red: a red spell can't target it.
func TestVis5631FreewindFalconCantBeTargetedByRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	falcon := effectstest.B12Push(g, g.Seats[1].ID, "Freewind Falcon", "Creature — Bird", effectstest.VisFreewindFalconOracle, 1, 1)
	effectstest.HmlSetColors(g, falcon, "W")
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	visBoltFails(t, g, me, falcon)
}

// visBoltFails casts a Lightning Bolt from `me` at `target` and wants
// the cast refused.
func visBoltFails(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: effectstest.VisLightningBoltForKeeperLabel, TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle,
		ManaCost: "{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
	})
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: effectstest.CardTarget(target)}); err == nil {
		t.Error("Lightning Bolt targeted a creature with protection from red")
	}
}

// --- can't block, and flanking ----------------------------------------------

func TestVis5631CantBlockBodies(t *testing.T) {
	cases := []struct{ name, typeLine, oracle string }{
		{"Hulking Cyclops", "Creature — Cyclops", effectstest.VisHulkingCyclopsOracle},
		{"Fallen Askari", effectstest.VisHumanKnightTypeLine, effectstest.VisFallenAskariOracle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			attacker := effectstest.B12Creature(g, me.ID, "Raider", "Creature — Human", 2, 2)
			body := effectstest.B12Push(g, opp.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			other := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
			effectstest.BrAttack(t, g, attacker)
			effectstest.BrRefusal(t, g.DeclareBlocker(body, attacker), game.BlockReasonCantBlock)
			if err := g.DeclareBlocker(other, attacker); err != nil {
				t.Fatalf("another creature can't block: %v", err)
			}
		})
	}
}

// Flanking (CR 702.25a): each blocker without flanking gets -1/-1 until
// end of turn; one with flanking does not.
func TestVis5631Flanking(t *testing.T) {
	cases := []struct{ name, oracle string }{
		{"Fallen Askari", effectstest.VisFallenAskariOracle},
		{"Suq'Ata Lancer", effectstest.VisSuqAtaLancerOracle},
		{"Teferi's Honor Guard", effectstest.VisTeferisHonorGuardOracle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			knight := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: tc.name, OracleID: tc.oracle, TypeLine: effectstest.VisHumanKnightTypeLine, Power: 2, Toughness: 2, Keywords: []string{"flanking"}})
			plain := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Plain", "", 2, 2))
			flanker := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flanker", TypeLine: "Creature — Test", Power: 2, Toughness: 2, Keywords: []string{"flanking"}})
			g.WithWriteLock(func() {
				for i, b := range []uuid.UUID{plain, flanker} {
					g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: opp.ID, CardID: b, Target: knight, Amount: i + 1})
				}
			})
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.ApaLive(g, plain).CurrentToughness(); got != 1 {
				t.Errorf("the blocker without flanking has toughness %d, want 1", got)
			}
			if got := effectstest.ApaLive(g, flanker).CurrentToughness(); got != 2 {
				t.Errorf("the blocker with flanking has toughness %d, want 2", got)
			}
		})
	}
}

// --- regeneration ----------------------------------------------------------------

func TestVis5631RiverBoaRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	boa := effectstest.VisShield(t, g, me, 2, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(boa); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(boa) {
		t.Fatal("the shielded Boa was destroyed")
	}
	if !effectstest.HmlTapped(t, g, boa) {
		t.Error("regeneration did not tap the Boa")
	}
}

func TestVis5631WakeOfVulturesSacrificesToRegenerate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	wake := effectstest.B12Push(g, me.ID, "Wake of Vultures", "Creature — Bird", effectstest.VisWakeOfVulturesOracle, 3, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "C", "B")
	effectstest.ApaActivate(t, g, me, wake, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear was not sacrificed to pay the cost")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(wake); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(wake) {
		t.Error("the shielded Vultures were destroyed")
	}
}

// --- the other activated abilities --------------------------------------------

func TestVis5631DarajaGriffinDestroysOnlyABlackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	griffin := effectstest.B12Push(g, me.ID, "Daraja Griffin", "Creature — Griffin", effectstest.VisDarajaGriffinOracle, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.HmlSetColors(g, bear, "G")
	zombie := effectstest.B12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2)
	effectstest.HmlSetColors(g, zombie, "B")
	if err := g.ActivateCatalogAbility(me.ID, griffin, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("the Griffin targeted a green creature")
	}
	if !g.Battlefield.Contains(griffin) {
		t.Fatal("a refused activation sacrificed the Griffin")
	}
	effectstest.ApaActivate(t, g, me, griffin, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(zombie)})
	if g.Battlefield.Contains(griffin) {
		t.Error("the Griffin was not sacrificed")
	}
	if g.Battlefield.Contains(zombie) {
		t.Error("the black creature survived")
	}
}

func TestVis5631ViashivanDragonPumpsEachWay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	dragon := effectstest.B12Push(g, me.ID, "Viashivan Dragon", "Creature — Dragon", effectstest.VisViashivanDragonOracle, 4, 4)
	effectstest.ApaMana(me, "R")
	effectstest.ApaActivate(t, g, me, dragon, 0, game.ActivateAbilityParams{})
	effectstest.ApaMana(me, "G")
	effectstest.ApaActivate(t, g, me, dragon, 1, game.ActivateAbilityParams{})
	effectstest.ApaMana(me, "G")
	effectstest.ApaActivate(t, g, me, dragon, 1, game.ActivateAbilityParams{})
	if p, tt := effectstest.PtOf(t, g, dragon); p != 5 || tt != 6 {
		t.Errorf("after the pumps %d/%d, want 5/6", p, tt)
	}
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p, tt := effectstest.PtOf(t, g, dragon); p != 4 || tt != 4 {
		t.Errorf("next turn %d/%d, want 4/4", p, tt)
	}
}

func TestVis5631KeeperOfKookusGainsProtectionFromRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	keeper := effectstest.B12Push(g, me.ID, "Keeper of Kookus", "Creature — Goblin", effectstest.VisKeeperOfKookusOracle, 1, 1)
	if effectstest.HasEffectiveKeyword(t, g, keeper, "protection from red") {
		t.Fatal("protection before the ability")
	}
	effectstest.ApaMana(me, "R")
	effectstest.ApaActivate(t, g, me, keeper, 0, game.ActivateAbilityParams{})
	if !effectstest.HasEffectiveKeyword(t, g, keeper, "protection from red") {
		t.Fatal("no protection from red after the ability")
	}
	visBoltFails(t, g, me, keeper)
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if effectstest.HasEffectiveKeyword(t, g, keeper, "protection from red") {
		t.Error("protection outlasted the turn")
	}
}

func TestVis5631JamuraanLionStopsATargetFromBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	lion := effectstest.PushCatalogPermanent(g, me.ID, "Jamuraan Lion", "Creature — Cat", effectstest.VisJamuraanLionOracle, false)
	attacker := effectstest.B12Creature(g, me.ID, "Raider", "Creature — Human", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	other := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "W")
	effectstest.ApaActivate(t, g, me, lion, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(wall)})
	if !effectstest.HmlTapped(t, g, lion) {
		t.Error("the Lion was not tapped to pay the cost")
	}
	effectstest.BrAttack(t, g, attacker)
	effectstest.BrRefusal(t, g.DeclareBlocker(wall, attacker), game.BlockReasonCantBlock)
	if err := g.DeclareBlocker(other, attacker); err != nil {
		t.Fatalf("an untargeted creature can't block: %v", err)
	}
}

func TestVis5631InfantryVeteranPumpsOnlyAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	veteran := effectstest.PushCatalogPermanent(g, me.ID, "Infantry Veteran", "Creature — Human Soldier", effectstest.VisInfantryVeteranOracle, false)
	attacker := effectstest.B12Creature(g, me.ID, "Raider", "Creature — Human", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Human", 2, 2)
	effectstest.BrAttack(t, g, attacker)
	if err := g.ActivateCatalogAbility(me.ID, veteran, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("the Veteran targeted a creature that isn't attacking")
	}
	effectstest.ApaActivate(t, g, me, veteran, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(attacker)})
	if p, tt := effectstest.PtOf(t, g, attacker); p != 3 || tt != 3 {
		t.Errorf("the attacker is %d/%d, want 3/3", p, tt)
	}
}

func TestVis5631SisaysRingAddsTwoColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	ring := effectstest.PushCatalogPermanent(g, me.ID, "Sisay's Ring", "Artifact", effectstest.VisSisaysRingOracle, false)
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if n := me.ManaPool.CountColor("C"); n != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("pool holds %d colorless of %d, want 2 of 2", n, me.ManaPool.Total())
	}
	if !effectstest.HmlTapped(t, g, ring) {
		t.Error("the Ring did not tap")
	}
}

// Rainbow Efreet and Teferi's Honor Guard phase themselves out, and come
// back in their controller's next untap step.
func TestVis5631ThisCreaturePhasesOut(t *testing.T) {
	cases := []struct{ name, typeLine, oracle string }{
		{"Rainbow Efreet", "Creature — Efreet", effectstest.VisRainbowEfreetOracle},
		{"Teferi's Honor Guard", effectstest.VisHumanKnightTypeLine, effectstest.VisTeferisHonorGuardOracle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			seat := g.Turn.ActiveSeat
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.ApaMana(me, "U", "U")
			effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if g.Battlefield.Contains(id) || !g.PhasedOut.Contains(id) {
				t.Fatal("it did not phase out")
			}
			effectstest.AdvanceToMainOf(t, g, (seat+1)%len(g.Seats))
			if !g.PhasedOut.Contains(id) {
				t.Fatal("it phased in during another player's untap step")
			}
			effectstest.AdvanceToMainOf(t, g, seat)
			if !g.Battlefield.Contains(id) {
				t.Error("it did not phase in during its controller's untap step")
			}
		})
	}
}

// --- the spells ------------------------------------------------------------------

func TestVis5631CreepingMoldDestroysALandButNotACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	mold := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mold, Name: "Creeping Mold", TypeLine: "Sorcery", OracleID: effectstest.VisCreepingMoldOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, mold, game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("Creeping Mold targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Creeping Mold", "Sorcery", effectstest.VisCreepingMoldOracle, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived Creeping Mold")
	}
}

func TestVis5631ElvenCacheReturnsACardToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	card := effectstest.PushGraveyardCardForTest(me, "Old Friend")
	effectstest.CastCatalogSpell(t, g, "Elven Cache", "Sorcery", effectstest.VisElvenCacheOracle, effectstest.CardTarget(card))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) || me.Graveyard.Contains(card) {
		t.Error("the card did not return to hand")
	}
}

func TestVis5631InspirationMakesTargetPlayerDrawTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	before := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Inspiration", "Instant", effectstest.VisInspirationOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != before+2 {
		t.Errorf("the target's hand is %d, want %d", got, before+2)
	}
}

func TestVis5631UndoBouncesTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Undo", "Sorcery", effectstest.VisUndoOracle, effectstest.CardTargets(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("the two creatures are not in their owners' hands")
	}
}

func TestVis5631WarriorsHonorPumpsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Warrior's Honor", "Instant", effectstest.VisWarriorsHonorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("yours is %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("theirs is %d/%d, want 2/2", p, tt)
	}
}

func TestVis5631MiraculousRecoveryReturnsWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	card := effectstest.PushGraveyardCardForTest(me, "Fallen Hero")
	effectstest.CastCatalogSpell(t, g, "Miraculous Recovery", "Instant", effectstest.VisMiraculousRecoveryOracle, effectstest.CardTarget(card))
	effectstest.PassPriorityAroundTable(t, g)
	hero, ok := effectstest.VisBattlefieldNamed(g, me.ID, "Fallen Hero")
	if !ok {
		t.Fatal("the creature card did not return to the battlefield")
	}
	if got := hero.Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("%d +1/+1 counters on it, want 1", got)
	}
}

func TestVis5631RetributionOfTheMeekDestroysOnlyTheBig(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	shielded := effectstest.VisShield(t, g, me, 5, 5)
	effectstest.CastCatalogSpell(t, g, "Retribution of the Meek", "Sorcery", effectstest.VisRetributionOfTheMeekOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a 4-power creature survived")
	}
	if g.Battlefield.Contains(shielded) {
		t.Error("a regeneration shield saved a 5-power creature")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("a 3-power creature was destroyed")
	}
}

func TestVis5631TremorSparesFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := effectstest.B12Creature(g, opp.ID, "Walker", "Creature — Human", 1, 1)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Tremor", "Sorcery", effectstest.VisTremorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the creature without flying survived")
	}
	if !g.Battlefield.Contains(flyer) {
		t.Error("the flyer was dealt damage")
	}
	if me.Life != life {
		t.Errorf("the caster lost life: %d -> %d", life, me.Life)
	}
}

// --- the enters triggers ------------------------------------------------------

func TestVis5631ManOWarBouncesATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	jelly := effectstest.CastHoldCreature(t, g, "Man-o'-War", "Creature — Jellyfish", effectstest.VisManOWarOracle, 2, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the target was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(jelly) {
		t.Error("the Man-o'-War left the battlefield")
	}
}

func TestVis5631UktabiOrangutanDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Uktabi Orangutan", "Creature — Ape", effectstest.VisUktabiOrangutanOracle, 2, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a creature was accepted as the artifact target")
	}
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestVis5631NekrataalDestroysPastRegeneration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	shielded := effectstest.VisShield(t, g, opp, 2, 1)
	zombie := effectstest.B12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2)
	effectstest.HmlSetColors(g, zombie, "B")
	golem := effectstest.B12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	effectstest.CastHoldCreature(t, g, "Nekrataal", "Creature — Human Assassin", effectstest.VisNekrataalOracle, 2, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{zombie, golem} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as a target", bad)
		}
	}
	effectstest.AnswerPickTarget(t, g, shielded)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shielded) {
		t.Error("a regeneration shield saved Nekrataal's target")
	}
}

// --- the land cycle ---------------------------------------------------------------

type visKaroo struct {
	name, oracle, landType, color string
}

var visKaroos = []visKaroo{
	{"Coral Atoll", effectstest.VisCoralAtollOracle, "Island", "U"},
	{"Dormant Volcano", effectstest.VisDormantVolcanoOracle, "Mountain", "R"},
	{"Everglades", effectstest.VisEvergladesOracle, "Swamp", "B"},
	{"Jungle Basin", effectstest.VisJungleBasinOracle, "Forest", "G"},
	{"Karoo", effectstest.VisKarooOracle, "Plains", "W"},
}

func TestVis5631KarooTapsForColorlessAndItsColor(t *testing.T) {
	for _, tc := range visKaroos {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			land := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			if me.ManaPool.CountColor("C") != 1 || me.ManaPool.CountColor(tc.color) != 1 || me.ManaPool.Total() != 2 {
				t.Errorf("pool holds %d, want {C}{%s}", me.ManaPool.Total(), tc.color)
			}
		})
	}
}

// Only an untapped land of the type is offered; returning it keeps the
// land, which entered tapped.
func TestVis5631KarooReturnsAnUntappedLandOfItsType(t *testing.T) {
	for _, tc := range visKaroos {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			basic := effectstest.SeedLandOnBattlefield(g, me.ID, tc.landType, "Basic Land — "+tc.landType)
			tapped := effectstest.SeedLandOnBattlefield(g, me.ID, tc.landType, "Basic Land — "+tc.landType)
			g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, tapped).Tapped = true })
			wrong := effectstest.SeedLandOnBattlefield(g, me.ID, "Wastes", "Basic Land")
			id := effectstest.PlayLandFromHand(t, g, tc.name, tc.oracle)
			effectstest.Top100AssertEnteredTapped(t, g, id, tc.name)
			effectstest.PassPriorityAroundTable(t, g)
			pick := effectstest.LatestChooseCardsFor(g, me.ID)
			if pick == nil {
				t.Fatal("no prompt to return a land")
			}
			if !slices.Contains(pick.ChooseCards, basic) || slices.Contains(pick.ChooseCards, tapped) || slices.Contains(pick.ChooseCards, wrong) || slices.Contains(pick.ChooseCards, id) {
				t.Fatalf("offered %v; want only the untapped %s", pick.ChooseCards, tc.landType)
			}
			if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{basic}); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Errorf("%s was sacrificed after the land was returned", tc.name)
			}
			if !me.Hand.Contains(basic) {
				t.Errorf("the %s is not in its owner's hand", tc.landType)
			}
		})
	}
}

// Declining, or having no untapped land of the type, sacrifices it.
func TestVis5631KarooIsSacrificedWithoutThePayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	island := effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	atoll := effectstest.PlayLandFromHand(t, g, "Coral Atoll", effectstest.VisCoralAtollOracle)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to return an Island")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(atoll) {
		t.Error("Coral Atoll stayed after the payment was declined")
	}
	if !g.Battlefield.Contains(island) {
		t.Error("the Island moved")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	tapped := effectstest.SeedLandOnBattlefield(g2, me2.ID, "Plains", "Basic Land — Plains")
	g2.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g2, tapped).Tapped = true })
	karoo := effectstest.PlayLandFromHand(t, g2, "Karoo", effectstest.VisKarooOracle)
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(karoo) {
		t.Error("Karoo stayed with only a tapped Plains")
	}
}
