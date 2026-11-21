package testss82m

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b20_test.go — Phyrexia: All Will Be One, batch 20 (#5479): five
// planeswalkers, a Hatchery and a flashback sorcery, driven through the real
// engine.

const (
	s82JaceOracle     = "03dafab4-841f-41cb-8f2e-8188f5177837"
	s82KaitoOracle    = "17fd42f6-f388-4ffc-8fec-ca9ada2506ee"
	s82KayaOracle     = "9260ae3d-59bf-4e46-a13e-930754c3f0d2"
	s82KothOracle     = "2c75b870-263b-41ab-9ece-6e3845bc3d8d"
	s82LukkaOracle    = "8e68eba1-65ac-4c03-956f-45e7b5695e2a"
	s82NahiriOracle   = "d71d42e0-cd56-4f53-95ff-ecab7f8c6552"
	s82NissaOracle    = "9c8845d6-fa65-4731-b348-dc724fc2d285"
	s82TyvarOracle    = "848c2b8f-8086-4047-8ad0-24b701dc4d0d"
	s82HatcheryOracle = "c4d65bcd-f6c5-40c3-825d-06b5183e7dbe"
	s82SpawningOracle = "9945c7eb-97a9-42df-a051-1ea514973280"
)

func TestS82OneB20CardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Jace": s82JaceOracle, "Kaito": s82KaitoOracle, "Kaya": s82KayaOracle, "Koth": s82KothOracle,
		"Lukka": s82LukkaOracle, "Nahiri": s82NahiriOracle, "Nissa": s82NissaOracle, "Tyvar": s82TyvarOracle,
		"Goliath Hatchery": s82HatcheryOracle, "Viral Spawning": s82SpawningOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v %v", name, spec, ok)
		}
	}
}

// --- Koth, Fire of Resistance ---------------------------------------------------

func TestS82KothPlusTwoFetchesABasicMountainToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	koth := effectstest.S82Walker(g, me.ID, "Koth, Fire of Resistance", s82KothOracle, 4)
	mountain := effectstest.SeedLibraryTop(me, "Mountain", "Basic Land — Mountain", "")
	forest := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, koth, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, mountain)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mountain) || me.Hand.Contains(forest) {
		t.Errorf("hand has Mountain=%v Forest=%v, want only the Mountain", me.Hand.Contains(mountain), me.Hand.Contains(forest))
	}
	if got := effectstest.LoyaltyOf(g, koth); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestS82KothMinusThreeDealsDamageEqualToMountains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	koth := effectstest.S82Walker(g, me.ID, "Koth, Fire of Resistance", s82KothOracle, 4)
	for i := 0; i < 2; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Mountain", "Basic Land — Mountain", "", false) // not mine
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, koth, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, bear)
	if c.DamageMarked != 2 {
		t.Errorf("Bear has %d damage, want 2 (two Mountains I control)", c.DamageMarked)
	}
	if got := effectstest.LoyaltyOf(g, koth); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

func TestS82KothUltimateEmblemShootsWhenAMountainEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	koth := effectstest.S82Walker(g, me.ID, "Koth, Fire of Resistance", s82KothOracle, 7)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, koth, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.EmblemsOf(g, me.ID)) != 1 {
		t.Fatal("no emblem")
	}
	life := opp.Life
	// A Forest does nothing.
	forest := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, forest); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Fatalf("a Forest entering dealt damage: %d -> %d", life, opp.Life)
	}
	mountain := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mountain, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, mountain); err != nil {
		t.Fatal(err)
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-4)
	}
}

// --- Tyvar, Jubilant Brawler ----------------------------------------------------

func TestS82TyvarPlusOneUntapsUpToOneTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tyvar := effectstest.S82Walker(g, me.ID, "Tyvar, Jubilant Brawler", s82TyvarOracle, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, tyvar, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Tapped {
		t.Error("Bear is still tapped")
	}
}

func TestS82TyvarMinusTwoMillsThenReturnsACheapCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tyvar := effectstest.S82Walker(g, me.ID, "Tyvar, Jubilant Brawler", s82TyvarOracle, 3)
	cheap := effectstest.SeedLibraryTop(me, "Cheap Elf", "Creature — Elf", "{1}{G}")
	dear := effectstest.SeedLibraryTop(me, "Dear Giant", "Creature — Giant", "{3}{G}")
	spell := effectstest.SeedLibraryTop(me, "Bolt", "Instant", "{R}")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, tyvar, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.Cmdr2WaitForChooseCards(t, g, me.ID)
	if len(p.ChooseCards) != 1 || p.ChooseCards[0] != cheap {
		t.Fatalf("offered %v, want only the mana value 2 creature %v (not %v or %v)", p.ChooseCards, cheap, dear, spell)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{cheap}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the Elf did not return")
	}
	if !me.Graveyard.Contains(dear) || !me.Graveyard.Contains(spell) {
		t.Error("the other milled cards should be in the graveyard")
	}
}

func TestS82TyvarLetsAFreshCreatureUseItsTapAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.S82Walker(g, me.ID, "Tyvar, Jubilant Brawler", s82TyvarOracle, 3)
	spec, _ := effects.Lookup(s82TyvarOracle)
	if !spec.ActivateAsThoughHaste {
		t.Error("Tyvar does not declare the as-though-haste static")
	}
}

// --- Kaito, Dancing Shadow --------------------------------------------------------

func TestS82KaitoPlusOneStopsAttackingAndBlockingUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	kaito := effectstest.S82Walker(g, me.ID, "Kaito, Dancing Shadow", s82KaitoOracle, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, kaito, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, bear)
	if r := c.Effective().Restrictions; r&game.CantAttack == 0 || r&game.CantBlock == 0 {
		t.Fatalf("restrictions %v, want can't attack or block", r)
	}
	effectstest.PowerUpPassTo(t, g, 1) // its controller's turn: still held
	c, _ = effectstest.BattlefieldCard(g, bear)
	if r := c.Effective().Restrictions; r&game.CantAttack == 0 {
		t.Error("it can attack on its own controller's turn")
	}
	effectstest.PowerUpPassTo(t, g, 0) // my next turn: over
	c, _ = effectstest.BattlefieldCard(g, bear)
	if r := c.Effective().Restrictions; r&(game.CantAttack|game.CantBlock) != 0 {
		t.Errorf("restrictions %v persisted past my next turn", r)
	}
}

func TestS82KaitoZeroDrawsAndMinusTwoMakesADrone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	kaito := effectstest.S82Walker(g, me.ID, "Kaito, Dancing Shadow", s82KaitoOracle, 3)
	effectstest.AwbPadLibrary(me)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, kaito, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card", hand, me.Hand.Size())
	}
	effectstest.PowerUpPassTo(t, g, 0)
	effectstest.PowerUpPassTo(t, g, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, kaito, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var drone uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Drone" && c.Controller == me.ID {
			drone = c.InstanceID
			e := c.Effective()
			if e.Power != 2 || e.Toughness != 2 || len(c.Colors) != 0 || !c.HasSubtype("Drone") || !game.HasKeyword(&c, "deathtouch") {
				t.Errorf("Drone is %d/%d %v %v, want a colorless 2/2 Drone with deathtouch", e.Power, e.Toughness, c.Colors, e.Subtypes)
			}
			if !c.IsToken() || !c.IsArtifact() || !c.IsCreature() {
				t.Errorf("Drone types %v, want a token artifact creature", e.Types)
			}
		}
	}
	if drone == uuid.Nil {
		t.Fatal("no Drone")
	}
	mine, theirs := me.Life, opp.Life
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(drone)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine+2 || opp.Life != theirs-2 {
		t.Errorf("life %d -> %d / %d -> %d, want +2 and -2 when the Drone left", mine, me.Life, theirs, opp.Life)
	}
}

// Returning a creature that dealt combat damage lets Kaito's loyalty abilities
// be used twice; declining keeps the one-a-turn limit (CR 606.3).
func TestS82KaitoBounceGrantsASecondLoyaltyActivation(t *testing.T) {
	for _, bounce := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		kaito := effectstest.S82Walker(g, me.ID, "Kaito, Dancing Shadow", s82KaitoOracle, 3)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AwbPadLibrary(me)
		effectstest.AdvanceToMain(t, g)
		effectstest.S82Activate(t, g, me.ID, kaito, 1, game.ActivateAbilityParams{}) // 0: draw
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
		p := effectstest.Cmdr2WaitForChooseCards(t, g, me.ID)
		if len(p.ChooseCards) != 1 || p.ChooseCards[0] != bear || p.ChooseMin != 0 {
			t.Fatalf("offered %v min %d, want the Bear, optional", p.ChooseCards, p.ChooseMin)
		}
		pick := []uuid.UUID(nil)
		if bounce {
			pick = []uuid.UUID{bear}
		}
		if err := g.ResolveChooseCards(p.ID, me.ID, pick); err != nil {
			t.Fatal(err)
		}
		effectstest.ToMainForCost(t, g) // the second main phase
		err := g.ActivateCatalogAbility(me.ID, kaito, 1, game.ActivateAbilityParams{Strict: true})
		if bounce {
			if !handHasName(me, "Bear") || g.Battlefield.Contains(bear) {
				t.Error("the Bear was not returned to hand")
			}
			if err != nil {
				t.Fatalf("second activation refused after the bounce: %v", err)
			}
			spent := false
			g.ReadSnapshot(func() { spent = g.LoyaltySpentLocked(kaito) })
			if !spent {
				t.Error("a third activation would be allowed")
			}
		} else if err == nil {
			t.Error("declining the bounce still allowed a second activation")
		}
	}
}

// The permission is Kaito's alone, lasts the turn and goes with the object.
func TestS82KaitoTwicePermissionIsPerPermanentAndEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kaito := effectstest.S82Walker(g, me.ID, "Kaito, Dancing Shadow", s82KaitoOracle, 3)
	other := effectstest.S82Walker(g, me.ID, "Other Walker", "", 3)
	g.WithWriteLock(func() { g.AllowThisPermanentLoyaltyTwiceThisTurnForEffect(kaito) })
	if !g.LoyaltyPermanentTwiceThisTurn[kaito] || g.LoyaltyPermanentTwiceThisTurn[other] {
		t.Fatalf("permission set %v", g.LoyaltyPermanentTwiceThisTurn)
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var snap game.GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !restored.LoyaltyPermanentTwiceThisTurn[kaito] {
		t.Error("the permission did not survive a snapshot")
	}
	effectstest.PowerUpPassTo(t, g, 1)
	if g.LoyaltyPermanentTwiceThisTurn != nil {
		t.Errorf("the turn boundary kept %v", g.LoyaltyPermanentTwiceThisTurn)
	}
}

func handHasName(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}
