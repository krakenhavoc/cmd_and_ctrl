package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b144_test.go — S82 GAP batch 144 (Jumpstart 2022), #5948: the
// cards through the real catalog.

const (
	gapAlchemistsVial   = "f574bd9f-4246-42cb-bf4e-d2888daa44c9"
	gapCellarDoor       = "a2f79037-86c0-430c-ba91-62ac807a44ce"
	gapEdificeAuthority = "7c16707e-24da-4a7e-9d24-d3731987f675"
	gapHammerOfRuin     = "78ba3363-0fc8-42a9-b586-69d0795ca65e"
	gapInstrumentsOfWar = "8be46af9-1b70-4da9-b16d-869c3831dc99"
	gapJoustingLance    = "ab747483-2aa1-4b25-b2fa-cf7372a1b1dc"
	gapMonkeyCage       = "33a11c01-89a6-446b-84da-e48fd1f27446"
	gapPlanarAtlas      = "4f14bf38-13ec-461d-8a1b-340843b5d95d"
)

func gapB144PT(g *game.Game, id uuid.UUID) (p, tough int) {
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			p, tough = c.PowerForComparison(), c.CurrentToughness()
		}
	})
	return
}

func TestGapB144Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		gapAlchemistsVial: "Alchemist's Vial", gapCellarDoor: "Cellar Door",
		gapEdificeAuthority: "Edifice of Authority", gapHammerOfRuin: "Hammer of Ruin",
		gapInstrumentsOfWar: "Instruments of War", gapJoustingLance: "Jousting Lance",
		gapMonkeyCage: "Monkey Cage", gapPlanarAtlas: "Planar Atlas",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestGapB144AlchemistsVialDrawsAndSacrificesToStopACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Alchemist's Vial", "Artifact", gapAlchemistsVial, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Fatalf("hand %d, want %d: the enters trigger should draw a card", got, hand+1)
	}
	var vial uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == gapAlchemistsVial {
				vial = c.InstanceID
			}
		}
	})
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.B06AddMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, vial, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vial) {
		t.Error("the Vial was not sacrificed")
	}
	var restricted bool
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(theirs)
		restricted = c.Effective().Restrictions&game.CantAttackOrBlock == game.CantAttackOrBlock
	})
	if !restricted {
		t.Error("the creature can still attack or block")
	}
}

func TestGapB144CellarDoorMillsTheBottomAndMakesAZombieForACreature(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		zombie   bool
	}{{"Creature — Bear", true}, {"Sorcery", false}} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		bottom := effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Bottom", TypeLine: tc.typeLine})
		top := effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Next", TypeLine: "Land"})
		_ = top
		// pushLibraryCardForTest pushes to the bottom, so make "Bottom" the last one pushed.
		g.WithWriteLock(func() {
			opp.Library.Remove(bottom)
			opp.Library.PushBottom(game.Card{InstanceID: bottom, Name: "Bottom", TypeLine: tc.typeLine, Owner: opp.ID, Controller: opp.ID})
		})
		door := effectstest.PushCatalogPermanent(g, me.ID, "Cellar Door", "Artifact", gapCellarDoor, false)
		effectstest.B06AddMana(me, "W", "W", "W")
		if err := g.ActivateCatalogAbility(me.ID, door, 0, game.ActivateAbilityParams{
			Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)},
		}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !opp.Graveyard.Contains(bottom) {
			t.Errorf("%s: the bottom card did not reach the graveyard", tc.typeLine)
		}
		var zombies int
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.Controller == me.ID && c.HasSubtype("Zombie") {
					zombies++
				}
			}
		})
		if (zombies == 1) != tc.zombie || zombies > 1 {
			t.Errorf("%s: %d Zombies, want %v", tc.typeLine, zombies, tc.zombie)
		}
	}
}

func TestGapB144CellarDoorWithAnEmptyLibraryDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	door := effectstest.PushCatalogPermanent(g, me.ID, "Cellar Door", "Artifact", gapCellarDoor, false)
	effectstest.B06AddMana(me, "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, door, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.HasSubtype("Zombie") {
			t.Fatal("a Zombie was made from an empty library")
		}
	}
}

func TestGapB144EdificeOfAuthorityLocksACreatureOnceItHasThreeBricks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	edifice := effectstest.PushCatalogPermanent(g, me.ID, "Edifice of Authority", "Artifact", gapEdificeAuthority, false)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	untap := func() {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == edifice {
					g.Battlefield.Cards[i].Tapped = false
				}
			}
		})
	}
	// The second ability is refused with no bricks.
	effectstest.B06AddMana(me, "W", "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, edifice, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err == nil {
		t.Fatal("the second ability was activated with no brick counters")
	}
	for i := 0; i < 3; i++ {
		untap()
		if err := g.ActivateCatalogAbility(me.ID, edifice, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err != nil {
			t.Fatalf("first ability %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	var bricks int
	g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(edifice); bricks = c.Counters["brick"] })
	if bricks != 3 {
		t.Fatalf("%d brick counters, want 3", bricks)
	}
	untap()
	if err := g.ActivateCatalogAbility(me.ID, edifice, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("second ability with three bricks: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var r game.Restriction
	g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(theirs); r = c.Effective().Restrictions })
	want := game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana
	if r&want != want {
		t.Errorf("restrictions %v, want can't attack, block or activate", r)
	}
}

func TestGapB144HammerOfRuinPumpsAndDestroysTheVictimsEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	hammer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hammer of Ruin", TypeLine: "Artifact — Equipment",
		OracleID: gapHammerOfRuin, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Sword", TypeLine: "Artifact — Equipment",
		Owner: opp.ID, Controller: opp.ID,
	})
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Sword", TypeLine: "Artifact — Equipment",
		Owner: me.ID, Controller: me.ID,
	})
	attacker := effectstest.BlbB1Push(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, hammer, attacker)
	if p, _ := gapB144PT(g, attacker); p != 4 {
		t.Fatalf("equipped power %d, want 4", p)
	}
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	pick := game.TargetRef{Kind: game.TargetCard, ID: mine}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, p.Chooser, pick); err == nil {
			t.Fatal("the trigger targeted an Equipment its controller owns, not the damaged player's")
		}
	}
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: theirs}, true)
	if g.Battlefield.Contains(theirs) {
		t.Error("the damaged player's Equipment survived")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own Equipment was destroyed")
	}
}

func TestGapB144JoustingLanceHasFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	lance := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jousting Lance", TypeLine: "Artifact — Equipment",
		OracleID: gapJoustingLance, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, lance, bear)
	if p, _ := gapB144PT(g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	if !effectstest.P1g2HasKeyword(g, bear, "first strike") {
		t.Error("no first strike during its controller's turn")
	}
	for i := 0; i < 40 && g.Seats[g.Turn.ActiveSeat].ID == me.ID; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Seats[g.Turn.ActiveSeat].ID == me.ID {
		t.Fatal("the turn never passed")
	}
	if effectstest.P1g2HasKeyword(g, bear, "first strike") {
		t.Error("first strike persisted into another player's turn")
	}
}

func TestGapB144InstrumentsOfWarBuffsTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	elf := effectstest.BlbB1Push(g, me.ID, "Elf", "Creature — Elf", "", 1, 1)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.PushNamedTribePermanent(t, g, me.ID, "Instruments of War", "Artifact", gapInstrumentsOfWar, "Bear")
	if p, th := gapB144PT(g, bear); p != 3 || th != 3 {
		t.Errorf("my Bear %d/%d, want 3/3", p, th)
	}
	if p, _ := gapB144PT(g, elf); p != 1 {
		t.Errorf("my Elf power %d, want 1", p)
	}
	if p, _ := gapB144PT(g, theirs); p != 2 {
		t.Errorf("their Bear power %d, want 2", p)
	}
	spec, _ := effects.Lookup(gapInstrumentsOfWar)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
}

func TestGapB144MonkeyCageSacrificesAndMakesMonkeysForTheEnteringCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	cage := effectstest.PushCatalogPermanent(g, me.ID, "Monkey Cage", "Artifact", gapMonkeyCage, false)
	// An opponent's creature with mana value 3 enters.
	id := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}",
			Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID,
		})
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("enter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cage) {
		t.Error("Monkey Cage was not sacrificed")
	}
	var monkeys int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.HasSubtype("Monkey") {
				monkeys++
				if p, th := c.PowerForComparison(), c.CurrentToughness(); p != 2 || th != 2 || !c.HasColor("G") {
					t.Errorf("a Monkey is %d/%d colors %v, want a 2/2 green one", p, th, c.Colors)
				}
			}
		}
	})
	if monkeys != 3 {
		t.Errorf("%d Monkeys, want 3 (the entering creature's mana value)", monkeys)
	}
}

// gapB144AtlasLibrary sets the library so the top four cards, top first, are
// the given ones, with a filler underneath.
func gapB144AtlasLibrary(g *game.Game, p *game.Player, topFirst ...game.Card) (filler uuid.UUID) {
	filler = uuid.New()
	g.WithWriteLock(func() {
		p.Library.Cards = []game.Card{{InstanceID: filler, Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID}}
		for i := len(topFirst) - 1; i >= 0; i-- {
			c := topFirst[i]
			c.Owner, c.Controller = p.ID, p.ID
			p.Library.Cards = append(p.Library.Cards, c)
		}
	})
	return filler
}

func gapB144TopCard(g *game.Game, p *game.Player) (id uuid.UUID) {
	g.ReadSnapshot(func() { id = p.Library.Cards[len(p.Library.Cards)-1].InstanceID })
	return id
}

func TestGapB144PlanarAtlasEntersTappedAndKeepsALandOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	land := uuid.New()
	other := uuid.New()
	filler := gapB144AtlasLibrary(g, me,
		game.Card{InstanceID: uuid.New(), Name: "A", TypeLine: "Sorcery"},
		game.Card{InstanceID: other, Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{InstanceID: land, Name: "Swamp", TypeLine: "Basic Land — Swamp"},
		game.Card{InstanceID: uuid.New(), Name: "B", TypeLine: "Sorcery"},
	)
	atlas := effectstest.CastCatalogSpell(t, g, "Planar Atlas", "Artifact", gapPlanarAtlas, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(atlas)
		if !ok || !c.Tapped {
			t.Errorf("Planar Atlas on the battlefield: %v, tapped %v; want it to enter tapped", ok, c.Tapped)
		}
	})
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to reveal a land card")
	}
	if len(pick.ChooseCards) != 2 {
		t.Fatalf("offered %v: want only the two land cards", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := gapB144TopCard(g, me); got != land {
		t.Error("the revealed land is not on top of the library")
	}
	if got := len(me.Library.Cards); got != 5 {
		t.Errorf("library has %d cards, want 5", got)
	}
	// The other three looked-at cards went under the filler.
	if me.Library.Cards[0].InstanceID == filler {
		t.Error("the looked-at cards were not put on the bottom")
	}
}

func TestGapB144PlanarAtlasDecliningTheLookOrTheLandLeavesTheOrder(t *testing.T) {
	// Declining "you may look" leaves the library alone.
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	top := uuid.New()
	gapB144AtlasLibrary(g, me, game.Card{InstanceID: top, Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	effectstest.CastCatalogSpell(t, g, "Planar Atlas", "Artifact", gapPlanarAtlas, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, nil, false)
	if got := gapB144TopCard(g, me); got != top {
		t.Error("declining the look changed the library")
	}

	// No land among the top four: every card goes to the bottom, no prompt.
	g2 := effectstest.NewCatalogGame(t)
	me2, _, _ := effectstest.BlbB1Seats(g2)
	filler := gapB144AtlasLibrary(g2, me2,
		game.Card{InstanceID: uuid.New(), Name: "A", TypeLine: "Sorcery"},
		game.Card{InstanceID: uuid.New(), Name: "B", TypeLine: "Sorcery"},
		game.Card{InstanceID: uuid.New(), Name: "C", TypeLine: "Sorcery"},
		game.Card{InstanceID: uuid.New(), Name: "D", TypeLine: "Sorcery"},
	)
	effectstest.CastCatalogSpell(t, g2, "Planar Atlas", "Artifact", gapPlanarAtlas, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AssistCardSettle(t, g2, me2.ID, nil, true)
	if effectstest.LatestChooseCardsFor(g2, me2.ID) != nil {
		t.Error("a prompt was asked with no land among the four")
	}
	if got := gapB144TopCard(g2, me2); got != filler {
		t.Error("with no land, the filler should now be on top of the library")
	}
}
