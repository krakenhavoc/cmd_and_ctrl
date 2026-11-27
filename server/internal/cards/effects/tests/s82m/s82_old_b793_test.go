package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b793_test.go — Mirage, batch 793 (#5633). Every test plays the real
// catalog entry through the engine.

const (
	oldB793Granger   = "0230fdda-d916-457e-9dfd-61783fcf4652"
	oldB793Darts     = "0c978da4-9caa-4b4b-be52-889677962ec9"
	oldB793Sling     = "10a52a73-157f-475e-975a-eb00c80573d4"
	oldB793Cobra     = "115c7ca7-c27b-43ca-9281-be6913e90059"
	oldB793Unicorn   = "2f91818f-a6fe-4472-b52b-ab574f790905"
	oldB793Feast     = "331233ba-0ddc-4cb2-b3b7-1b782e03529f"
	oldB793Cage      = "38ec5b13-b143-4818-9243-b1715dab3bef"
	oldB793Efreet    = "55cbc0e8-eeec-433a-b70c-3501238b658b"
	oldB793Embermage = "5771296a-16f8-423e-9b83-8e6f687cbb34"
	oldB793Spirit    = "58917163-584d-4c74-8243-f56c5d4b49e7"
	oldB793Bolt      = "5c19db44-0eae-4d22-9bf4-7e6cdc6b9d8f"
)

func TestOldB793EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB793Granger: "Granger Guildmage", oldB793Darts: "Telim'Tor's Darts", oldB793Sling: "Unerring Sling",
		oldB793Cobra: "Sabertooth Cobra", oldB793Unicorn: "Zebra Unicorn", oldB793Feast: "Grim Feast",
		oldB793Cage: "Paupers' Cage", oldB793Efreet: "Burning Palm Efreet", oldB793Embermage: "Reckless Embermage",
		oldB793Spirit: "Subterranean Spirit", oldB793Bolt: "Energy Bolt",
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

func oldB793Player(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

func oldB793Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func TestOldB793GrangerGuildmageHurtsItsTargetAndYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	m := effectstest.BlbB1Push(g, me.ID, "Granger Guildmage", "Creature — Human Wizard", oldB793Granger, 1, 1)
	effectstest.FloatMana(t, g, me, "{R}")
	myLife, oppLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	effectstest.B16Activate(t, g, me.ID, m, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB793Player(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife-1 {
		t.Errorf("opponent life %d, want %d", got, oppLife-1)
	}
	if got := effectstest.LifeOf(g, me.ID); got != myLife-1 {
		t.Errorf("your life %d, want %d", got, myLife-1)
	}
}

func TestOldB793GrangerGuildmageGivesFirstStrikeForWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	m := effectstest.BlbB1Push(g, me.ID, "Granger Guildmage", "Creature — Human Wizard", oldB793Granger, 1, 1)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.FloatMana(t, g, me, "{W}")
	effectstest.B16Activate(t, g, me.ID, m, 1, game.ActivateAbilityParams{Strict: true, Targets: oldB793Card(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, bear, "first strike") {
		t.Error("the Bear did not gain first strike")
	}
}

func TestOldB793TelimTorsDartsPingsAPlayerForTwoAndTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	d := effectstest.PushCatalogPermanent(g, me.ID, "Telim'Tor's Darts", "Artifact", oldB793Darts, false)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.B16Activate(t, g, me.ID, d, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB793Player(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-1 {
		t.Errorf("life %d, want %d", got, before-1)
	}
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, d, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB793Player(opp.ID)}); err == nil {
		t.Error("activated while tapped")
	}
}

func TestOldB793TelimTorsDartsCannotTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	d := effectstest.PushCatalogPermanent(g, me.ID, "Telim'Tor's Darts", "Artifact", oldB793Darts, false)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, d, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB793Card(bear)}); err == nil {
		t.Error("targeted a creature")
	}
}

func TestOldB793UnerringSlingDealsTheTappedCreaturesPowerToAnAttackingFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sling := effectstest.PushCatalogPermanent(g, me.ID, "Unerring Sling", "Artifact", oldB793Sling, false)
	ogre := effectstest.BlbB1Push(g, me.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 3, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	effectstest.Pr7bAttack(t, g, opp.ID, flyer)
	effectstest.FloatMana(t, g, me, "{R}{R}{R}")
	effectstest.B16Activate(t, g, me.ID, sling, 0, game.ActivateAbilityParams{Strict: true,
		Targets: oldB793Card(flyer), TapIDs: []uuid.UUID{ogre}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flyer) {
		t.Error("the flyer survived 3 damage")
	}
	if !effectstest.B16Tapped(t, g, ogre) {
		t.Error("the Ogre was not tapped")
	}
}

func TestOldB793UnerringSlingRefusesAGroundedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sling := effectstest.PushCatalogPermanent(g, me.ID, "Unerring Sling", "Artifact", oldB793Sling, false)
	ogre := effectstest.BlbB1Push(g, me.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.FloatMana(t, g, me, "{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, sling, 0, game.ActivateAbilityParams{Strict: true,
		Targets: oldB793Card(bear), TapIDs: []uuid.UUID{ogre}}); err == nil {
		t.Error("targeted a creature that is neither attacking nor blocking nor flying")
	}
}

func TestOldB793ZebraUnicornGainsLifeEqualToAnyDamageItDeals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	u := effectstest.BlbB1Push(g, me.ID, "Zebra Unicorn", "Creature — Unicorn", oldB793Unicorn, 2, 2)
	before := effectstest.LifeOf(g, me.ID)
	effectstest.P2Ping(t, g, u, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != before+3 {
		t.Errorf("life %d, want %d", got, before+3)
	}
}

func TestOldB793GrimFeastHurtsYouAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Grim Feast", "Enchantment", oldB793Feast, false)
	before := effectstest.LifeOf(g, me.ID)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != before-1 {
		t.Errorf("life %d, want %d", got, before-1)
	}
}

func TestOldB793GrimFeastGainsToughnessForAnOpponentsDeadCreatureOnly(t *testing.T) {
	for _, tc := range []struct {
		name string
		mine bool
		want int
	}{{"an opponent's creature", false, 3}, {"your own creature", true, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			effectstest.PushCatalogPermanent(g, me.ID, "Grim Feast", "Enchantment", oldB793Feast, false)
			owner := opp.ID
			if tc.mine {
				owner = me.ID
			}
			victim := effectstest.BlbB1Push(g, owner, "Victim", "Creature — Bear", "", 1, 3)
			before := effectstest.LifeOf(g, me.ID)
			effectstest.P1Destroy(t, g, victim)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.LifeOf(g, me.ID); got != before+tc.want {
				t.Errorf("life %d, want %d", got, before+tc.want)
			}
		})
	}
}

func TestOldB793PaupersCageNeedsAnOpponentWithTwoOrFewerCards(t *testing.T) {
	for _, tc := range []struct {
		name string
		hand int
		want int
	}{{"two cards", 2, 2}, {"three cards", 3, 0}, {"empty hand", 0, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			effectstest.PushCatalogPermanent(g, me.ID, "Paupers' Cage", "Artifact", oldB793Cage, false)
			opp.Hand.Cards = nil
			for i := 0; i < tc.hand; i++ {
				effectstest.PushCatalogHandCard(opp, "Filler", "Sorcery", "")
			}
			before := effectstest.LifeOf(g, opp.ID)
			effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.LifeOf(g, opp.ID); got != before-tc.want {
				t.Errorf("life %d, want %d", got, before-tc.want)
			}
		})
	}
}

func TestOldB793BurningPalmEfreetHurtsAFlyerAndGroundsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	e := effectstest.BlbB1Push(g, me.ID, "Burning Palm Efreet", "Creature — Efreet", oldB793Efreet, 2, 2)
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 3, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.FloatMana(t, g, me, "{R}{R}{R}")
	effectstest.B16Activate(t, g, me.ID, e, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB793Card(bird)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, bird); got != 2 {
		t.Errorf("damage %d, want 2", got)
	}
	if effectstest.P1g2HasKeyword(g, bird, "flying") {
		t.Error("the Bird kept flying")
	}
}

func TestOldB793BurningPalmEfreetRefusesAGroundedTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	e := effectstest.BlbB1Push(g, me.ID, "Burning Palm Efreet", "Creature — Efreet", oldB793Efreet, 2, 2)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.FloatMana(t, g, me, "{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, e, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB793Card(bear)}); err == nil {
		t.Error("targeted a creature without flying")
	}
}

func TestOldB793RecklessEmbermageHurtsItsTargetAndItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	e := effectstest.BlbB1Push(g, me.ID, "Reckless Embermage", "Creature — Human Wizard", oldB793Embermage, 2, 2)
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	effectstest.B16Activate(t, g, me.ID, e, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB793Player(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-1 {
		t.Errorf("life %d, want %d", got, before-1)
	}
	if got := effectstest.DamageOn(t, g, e); got != 1 {
		t.Errorf("the Embermage has %d damage, want 1", got)
	}
}

func TestOldB793SubterraneanSpiritHurtsGroundCreaturesButNotFlyersOrItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spirit := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Subterranean Spirit",
		TypeLine: "Creature — Elemental Spirit", OracleID: oldB793Spirit, Power: 3, Toughness: 3, Colors: []string{"R"},
		Owner: me.ID, Controller: me.ID})
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 3, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.B16Activate(t, g, me.ID, spirit, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, bear); got != 1 {
		t.Errorf("the Bear has %d damage, want 1", got)
	}
	if got := effectstest.DamageOn(t, g, bird); got != 0 {
		t.Errorf("the flyer has %d damage, want 0", got)
	}
	if got := effectstest.DamageOn(t, g, spirit); got != 0 {
		t.Errorf("the Spirit has %d damage, want 0 (protection from red)", got)
	}
}

func oldB793CastBolt(t *testing.T, g *game.Game, mode, x int, target game.TargetRef) {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	c := game.NewCard("Energy Bolt", me.ID)
	c.TypeLine, c.OracleID = "Sorcery", oldB793Bolt
	me.Hand.PushTop(c)
	target.Mode = 0
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Modes: []int{mode}, XValue: x, Targets: []game.TargetRef{target}}); err != nil {
		t.Fatalf("cast Energy Bolt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB793EnergyBoltDealsXDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := effectstest.LifeOf(g, opp.ID)
	oldB793CastBolt(t, g, 0, 4, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if got := effectstest.LifeOf(g, opp.ID); got != before-4 {
		t.Errorf("life %d, want %d", got, before-4)
	}
}

func TestOldB793EnergyBoltGainsXLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := effectstest.LifeOf(g, me.ID)
	oldB793CastBolt(t, g, 1, 3, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if got := effectstest.LifeOf(g, me.ID); got != before+3 {
		t.Errorf("life %d, want %d", got, before+3)
	}
}

func TestOldB793SabertoothCobraPoisonsAndPoisonsAgainUnlessPaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		name := "unpaid"
		if pay {
			name = "paid"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			cobra := effectstest.BlbB1Push(g, me.ID, "Sabertooth Cobra", "Creature — Snake", oldB793Cobra, 2, 2)
			effectstest.SeedLand(g, opp.ID, "Forest", "Basic Land — Forest", "")
			effectstest.SeedLand(g, opp.ID, "Forest", "Basic Land — Forest", "")
			effectstest.P2Ping(t, g, cobra, opp.ID, 2)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.PoisonOn(opp); got != 1 {
				t.Fatalf("poison %d after the damage, want 1", got)
			}
			if pay {
				var id uuid.UUID
				for _, dt := range g.DelayedTriggers {
					if dt != nil && dt.PayToCancel != nil {
						id = dt.ID
					}
				}
				if id == uuid.Nil {
					t.Fatal("no delayed trigger with a payment")
				}
				if err := g.PerformSpecialAction(opp.ID, id, game.SpecialActionPayToCancel, game.SpecialActionParams{Strict: true, AutoTap: true}); err != nil {
					t.Fatalf("pay {2}: %v", err)
				}
			}
			effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
			effectstest.PassPriorityAroundTable(t, g)
			want := 2
			if pay {
				want = 1
			}
			if got := effectstest.PoisonOn(opp); got != want {
				t.Errorf("poison %d after their upkeep, want %d", got, want)
			}
		})
	}
}
