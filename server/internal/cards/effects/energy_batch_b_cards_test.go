package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_batch_b_cards_test.go — ADR 0129 PR 1 (#1995): the energy
// cards whose activated abilities pay "Pay N {E}" and whose triggers
// give energy. The engine half is game/energy_cost_test.go.

const (
	ebAetherTheorist      = "b07ef53e-46db-457e-a5a2-846bb9805b70"
	ebAethergeodeMiner    = "92045f6c-9ec2-4bf0-826c-e21ef253011c"
	ebAethersphere        = "7b3e74ad-0179-480b-871c-9e3bc30a43ff"
	ebAethersquall        = "47c685b1-6e1e-4e08-8a52-9da6902df113"
	ebAethertideWhale     = "53b2c8f2-db9a-4ab0-a4e8-17385b2fa3bd"
	ebAethertorch         = "7db864fe-e0b7-4b32-8706-2ad01f93f7ca"
	ebAetherwindBasker    = "d9ec9a4f-0644-47c5-bc09-1201d5251bde"
	ebArchitect           = "be6cab52-b3ae-46a2-a581-2468a49bff25"
	ebBespokeBattlewagon  = "cdfcfd2b-966a-4e03-8ed5-4fa7fab423d0"
	ebBristlingHydra      = "b3b23c58-0b7a-4fe4-a8e8-5320a7605724"
	ebConsulateTurret     = "14d71432-c2e6-4c81-92ef-cf98e1fcf5f8"
	ebElectrostatic       = "79676f10-9fca-4bca-acc5-6994955142b4"
	ebJanjeetSentry       = "fb818d43-8e5d-41aa-98f8-5826dfd5dfdb"
	ebLongtuskCub         = "d2e78392-096a-4bd0-be19-9c54dfb8451f"
	ebMinisterOfInquiries = "99894900-771f-4fca-bebf-9f23ec4654c2"
	ebShipwreckMoray      = "6c0b07c4-bb7c-4100-81e5-b0ee95d18625"
	ebSolsticeZealot      = "1ec63280-ae5b-4892-8854-de2d8127ea83"
	ebSpontaneousArtist   = "c0602911-7d44-4a7a-a04c-837036ed3b82"
	ebTempestHarvester    = "5c0ffe37-16ce-4788-a434-58b6fc34bd9c"
	ebWhirlerVirtuoso     = "f82da236-e567-4221-a0bc-439a0c7e2b03"
	ebShieldedAetherThief = "514df71a-138f-460f-a03c-bd1af8e146e2"
)

// ebPush puts a catalog permanent onto me's battlefield, not summoning
// sick.
func ebPush(g *game.Game, me *game.Player, name, oracle, typeLine string, power, toughness int) uuid.UUID {
	return apaPush(g, me.ID, me.ID, game.Card{Name: name, OracleID: oracle, TypeLine: typeLine, Power: power, Toughness: toughness})
}

func ebCardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// Each card that gives energy on entering gives the printed amount, and
// declares it as its purpose.
func TestEnergyBatchBEnterTriggersGiveEnergy(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		n                      int
	}{
		{"Aether Theorist", ebAetherTheorist, "Creature — Vedalken Rogue", 3},
		{"Aethersphere Harvester", ebAethersphere, "Artifact — Vehicle", 2},
		{"Aethertide Whale", ebAethertideWhale, "Creature — Whale", 6},
		{"Aethertorch Renegade", ebAethertorch, "Creature — Human Rogue", 4},
		{"Bristling Hydra", ebBristlingHydra, "Creature — Hydra", 3},
		{"Electrostatic Pummeler", ebElectrostatic, "Artifact Creature — Construct", 3},
		{"Janjeet Sentry", ebJanjeetSentry, "Creature — Vedalken Soldier", 2},
		{"Minister of Inquiries", ebMinisterOfInquiries, "Creature — Vedalken Advisor", 2},
		{"Shipwreck Moray", ebShipwreckMoray, "Creature — Fish", 4},
		{"Solstice Zealot", ebSolsticeZealot, "Creature — Rhino Cleric", 2},
		{"Spontaneous Artist", ebSpontaneousArtist, "Creature — Human Rogue", 1},
		{"Tempest Harvester", ebTempestHarvester, "Creature — Merfolk Wizard", 2},
		{"Whirler Virtuoso", ebWhirlerVirtuoso, "Creature — Vedalken Artificer", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := p7Table(t)
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)
			if got := energyOf(me); got != tc.n {
				t.Errorf("energy after %s entered = %d, want %d", tc.name, got, tc.n)
			}
			spec, _ := Lookup(tc.oracle)
			if spec.Purpose.Energy != tc.n {
				t.Errorf("%s purpose energy = %d, want %d", tc.name, spec.Purpose.Energy, tc.n)
			}
		})
	}
}

// Every "Pay N {E}" row declares the printed amount, refuses a player
// one short with nothing paid, and takes exactly the amount when paid.
func TestEnergyBatchBActivationsPayEnergy(t *testing.T) {
	type row struct {
		name, oracle, typeLine string
		index, energy          int
		tap                    bool
		targets                func(me, opp *game.Player, src, bear uuid.UUID) []game.TargetRef
	}
	creature := func(_, _ *game.Player, _, bear uuid.UUID) []game.TargetRef { return ebCardTarget(bear) }
	player := func(_, opp *game.Player, _, _ uuid.UUID) []game.TargetRef { return p8Player(opp.ID) }
	for _, tc := range []row{
		{"Aether Theorist", ebAetherTheorist, "Creature — Vedalken Rogue", 0, 1, true, nil},
		{"Aethergeode Miner", ebAethergeodeMiner, "Creature — Dwarf Scout", 0, 2, false, nil},
		{"Aethersphere Harvester", ebAethersphere, "Artifact — Vehicle", 0, 1, false, nil},
		{"Aethersquall Ancient", ebAethersquall, "Creature — Leviathan", 0, 8, false, nil},
		{"Aethertide Whale", ebAethertideWhale, "Creature — Whale", 0, 4, false, nil},
		{"Aethertorch Renegade", ebAethertorch, "Creature — Human Rogue", 0, 2, true, creature},
		{"Aethertorch Renegade", ebAethertorch, "Creature — Human Rogue", 1, 8, true, player},
		{"Aetherwind Basker", ebAetherwindBasker, "Creature — Lizard", 0, 1, false, nil},
		{"Architect of the Untamed", ebArchitect, "Creature — Elf Artificer Druid", 0, 8, false, nil},
		{"Bespoke Battlewagon", ebBespokeBattlewagon, "Artifact — Vehicle", 1, 2, true, creature},
		{"Bespoke Battlewagon", ebBespokeBattlewagon, "Artifact — Vehicle", 2, 3, true, nil},
		{"Bespoke Battlewagon", ebBespokeBattlewagon, "Artifact — Vehicle", 3, 4, false, nil},
		{"Bristling Hydra", ebBristlingHydra, "Creature — Hydra", 0, 3, false, nil},
		{"Consulate Turret", ebConsulateTurret, "Artifact", 1, 3, true, player},
		{"Electrostatic Pummeler", ebElectrostatic, "Artifact Creature — Construct", 0, 3, false, nil},
		{"Janjeet Sentry", ebJanjeetSentry, "Creature — Vedalken Soldier", 0, 2, true, creature},
		{"Longtusk Cub", ebLongtuskCub, "Creature — Cat", 0, 2, false, nil},
		{"Minister of Inquiries", ebMinisterOfInquiries, "Creature — Vedalken Advisor", 0, 1, true, player},
		{"Shipwreck Moray", ebShipwreckMoray, "Creature — Fish", 0, 1, false, nil},
		{"Solstice Zealot", ebSolsticeZealot, "Creature — Rhino Cleric", 0, 1, true, creature},
		{"Spontaneous Artist", ebSpontaneousArtist, "Creature — Human Rogue", 0, 1, false, creature},
		{"Tempest Harvester", ebTempestHarvester, "Creature — Merfolk Wizard", 0, 1, true, nil},
		{"Whirler Virtuoso", ebWhirlerVirtuoso, "Creature — Vedalken Artificer", 0, 3, false, nil},
		{"Shielded Aether Thief", ebShieldedAetherThief, "Creature — Vedalken Rogue", 0, 3, true, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if c := specActivatedEnergy(t, tc.oracle, tc.index); c.Energy != tc.energy || c.Tap != tc.tap {
				t.Fatalf("row %d cost energy %d tap %v, want %d %v", tc.index, c.Energy, c.Tap, tc.energy, tc.tap)
			}
			g, me, opp := p7Table(t)
			src := ebPush(g, me, tc.name, tc.oracle, tc.typeLine, 2, 2)
			bear := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
			var targets []game.TargetRef
			if tc.targets != nil {
				targets = tc.targets(me, opp, src, bear)
			}
			setEnergy(t, g, me, tc.energy-1)
			err := g.ActivateCatalogAbility(me.ID, src, tc.index, game.ActivateAbilityParams{Targets: targets})
			if !errors.Is(err, game.ErrInsufficientEnergy) {
				t.Fatalf("one short: err = %v, want ErrInsufficientEnergy", err)
			}
			if c := findBattlefieldCardForTest(g, src); c == nil || c.Tapped {
				t.Fatal("a refused activation tapped the source")
			}
			if energyOf(me) != tc.energy-1 {
				t.Fatalf("a refused activation spent energy: %d", energyOf(me))
			}
			setEnergy(t, g, me, tc.energy+1)
			if err := g.ActivateCatalogAbility(me.ID, src, tc.index, game.ActivateAbilityParams{Targets: targets}); err != nil {
				t.Fatalf("activate with enough energy: %v", err)
			}
			if energyOf(me) != 1 {
				t.Errorf("energy after paying = %d, want 1", energyOf(me))
			}
		})
	}
}

func ebLive(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	c := apaLive(g, id)
	if c == nil {
		t.Fatalf("%v is not on the battlefield", id)
	}
	return c
}

// ebActivate gives me the energy, activates and resolves.
func ebActivate(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID, index, energy int, targets []game.TargetRef) {
	t.Helper()
	setEnergy(t, g, me, energy)
	p7Activate(t, g, me, src, index, game.ActivateAbilityParams{Targets: targets})
}

func TestAethergeodeMinerBlinksItself(t *testing.T) {
	g, me, _ := p7Table(t)
	src := ebPush(g, me, "Aethergeode Miner", ebAethergeodeMiner, "Creature — Dwarf Scout", 3, 1)
	ebActivate(t, g, me, src, 0, 2, nil)
	if g.Battlefield.Contains(src) {
		t.Fatal("the Miner is still the same object")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Aethergeode Miner" && c.Controller == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("the Miner did not return to the battlefield")
	}
}

func TestAethersquallAncientReturnsAllOtherCreatures(t *testing.T) {
	g, me, opp := p7Table(t)
	src := ebPush(g, me, "Aethersquall Ancient", ebAethersquall, "Creature — Leviathan", 6, 6)
	mine := apaPush(g, me.ID, me.ID, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	ebActivate(t, g, me, src, 0, 8, nil)
	if !g.Battlefield.Contains(src) {
		t.Error("the Ancient returned itself")
	}
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("my creature was not returned to my hand")
	}
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the opponent's creature was not returned to its owner's hand")
	}
}

func TestAethertideWhaleReturnsItself(t *testing.T) {
	g, me, _ := p7Table(t)
	src := ebPush(g, me, "Aethertide Whale", ebAethertideWhale, "Creature — Whale", 6, 4)
	ebActivate(t, g, me, src, 0, 4, nil)
	if g.Battlefield.Contains(src) || !me.Hand.Contains(src) {
		t.Error("the Whale is not in its owner's hand")
	}
}

func TestAethertorchRenegadeDealsDamage(t *testing.T) {
	g, me, opp := p7Table(t)
	src := ebPush(g, me, "Aethertorch Renegade", ebAethertorch, "Creature — Human Rogue", 1, 2)
	bear := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	ebActivate(t, g, me, src, 0, 2, ebCardTarget(bear))
	if c := ebLive(t, g, bear); c.DamageMarked != 1 {
		t.Errorf("the Bear has %d damage, want 1", c.DamageMarked)
	}

	g, me, opp = p7Table(t)
	src = ebPush(g, me, "Aethertorch Renegade", ebAethertorch, "Creature — Human Rogue", 1, 2)
	life := opp.Life
	ebActivate(t, g, me, src, 1, 8, p8Player(opp.ID))
	if opp.Life != life-6 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-6)
	}
}

func TestAetherwindBaskerCountsCreaturesAndPumps(t *testing.T) {
	g, me, _ := p7Table(t)
	apaPush(g, me.ID, me.ID, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	castCatalogSpell(t, g, "Aetherwind Basker", "Creature — Lizard", ebAetherwindBasker, nil)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 2 {
		t.Errorf("energy after the Basker entered beside one creature = %d, want 2", energyOf(me))
	}

	g, me, _ = p7Table(t)
	src := ebPush(g, me, "Aetherwind Basker", ebAetherwindBasker, "Creature — Lizard", 7, 7)
	ebActivate(t, g, me, src, 0, 1, nil)
	if c := ebLive(t, g, src); c.CurrentPower() != 8 || c.CurrentToughness() != 8 {
		t.Errorf("Basker is %d/%d, want 8/8", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestArchitectOfTheUntamedMakesABeast(t *testing.T) {
	g, me, _ := p7Table(t)
	src := ebPush(g, me, "Architect of the Untamed", ebArchitect, "Creature — Elf Artificer Druid", 2, 3)
	ebActivate(t, g, me, src, 0, 8, nil)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Beast" && c.Controller == me.ID {
			n++
			if c.Power != 6 || c.Toughness != 6 || !c.HasCardType("artifact") || len(c.Colors) != 0 {
				t.Errorf("Beast = %d/%d %q %v, want a colorless 6/6 artifact", c.Power, c.Toughness, c.TypeLine, c.Colors)
			}
		}
	}
	if n != 1 {
		t.Errorf("%d Beasts, want 1", n)
	}
}

func TestBespokeBattlewagonRows(t *testing.T) {
	g, me, _ := p7Table(t)
	src := ebPush(g, me, "Bespoke Battlewagon", ebBespokeBattlewagon, "Artifact — Vehicle", 5, 6)
	p7Activate(t, g, me, src, 0, game.ActivateAbilityParams{})
	if energyOf(me) != 2 {
		t.Errorf("energy after {T}: you get {E}{E} = %d, want 2", energyOf(me))
	}

	g, me, _ = p7Table(t)
	src = ebPush(g, me, "Bespoke Battlewagon", ebBespokeBattlewagon, "Artifact — Vehicle", 5, 6)
	hand := len(me.Hand.Cards)
	ebActivate(t, g, me, src, 2, 3, nil)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d, want %d", len(me.Hand.Cards), hand+1)
	}

	g, me, _ = p7Table(t)
	src = ebPush(g, me, "Bespoke Battlewagon", ebBespokeBattlewagon, "Artifact — Vehicle", 5, 6)
	ebActivate(t, g, me, src, 3, 4, nil)
	if c := ebLive(t, g, src); !c.IsCreature() || !c.HasCardType("artifact") {
		t.Errorf("the Battlewagon is %q, want an artifact creature", c.TypeLine)
	}
}

func TestBristlingHydraCounterAndHexproof(t *testing.T) {
	g, me, _ := p7Table(t)
	src := ebPush(g, me, "Bristling Hydra", ebBristlingHydra, "Creature — Hydra", 4, 3)
	ebActivate(t, g, me, src, 0, 3, nil)
	c := ebLive(t, g, src)
	if c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("+1/+1 counters %d, want 1", c.Counters[game.CounterPlusOne])
	}
	if !game.HasKeyword(c, "hexproof") {
		t.Error("the Hydra does not have hexproof")
	}
}

func TestConsulateTurretRows(t *testing.T) {
	g, me, opp := p7Table(t)
	src := ebPush(g, me, "Consulate Turret", ebConsulateTurret, "Artifact", 0, 0)
	p7Activate(t, g, me, src, 0, game.ActivateAbilityParams{})
	if energyOf(me) != 1 {
		t.Errorf("energy after {T}: you get {E} = %d, want 1", energyOf(me))
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, src).Tapped = false })
	life := opp.Life
	ebActivate(t, g, me, src, 1, 3, p8Player(opp.ID))
	if opp.Life != life-2 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-2)
	}
}

func TestElectrostaticPummelerDoublesItsSize(t *testing.T) {
	g, me, _ := p7Table(t)
	src := ebPush(g, me, "Electrostatic Pummeler", ebElectrostatic, "Artifact Creature — Construct", 1, 1)
	ebActivate(t, g, me, src, 0, 3, nil)
	if c := ebLive(t, g, src); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("Pummeler is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	ebActivate(t, g, me, src, 0, 3, nil)
	if c := ebLive(t, g, src); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("Pummeler twice is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestLongtuskCubCounter(t *testing.T) {
	g, me, _ := p7Table(t)
	src := ebPush(g, me, "Longtusk Cub", ebLongtuskCub, "Creature — Cat", 2, 2)
	ebActivate(t, g, me, src, 0, 2, nil)
	if c := ebLive(t, g, src); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("+1/+1 counters %d, want 1", c.Counters[game.CounterPlusOne])
	}
}

func TestMinisterOfInquiriesMillsThree(t *testing.T) {
	g, me, opp := p7Table(t)
	src := ebPush(g, me, "Minister of Inquiries", ebMinisterOfInquiries, "Creature — Vedalken Advisor", 1, 2)
	lib, gy := len(opp.Library.Cards), len(opp.Graveyard.Cards)
	ebActivate(t, g, me, src, 0, 1, p8Player(opp.ID))
	if len(opp.Library.Cards) != lib-3 || len(opp.Graveyard.Cards) != gy+3 {
		t.Errorf("library %d graveyard %d, want %d and %d", len(opp.Library.Cards), len(opp.Graveyard.Cards), lib-3, gy+3)
	}
}

func TestShipwreckMorayPump(t *testing.T) {
	g, me, _ := p7Table(t)
	src := ebPush(g, me, "Shipwreck Moray", ebShipwreckMoray, "Creature — Fish", 0, 5)
	ebActivate(t, g, me, src, 0, 1, nil)
	if c := ebLive(t, g, src); c.CurrentPower() != 2 || c.CurrentToughness() != 3 {
		t.Errorf("Moray is %d/%d, want 2/3", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSolsticeZealotTaps(t *testing.T) {
	g, me, opp := p7Table(t)
	src := ebPush(g, me, "Solstice Zealot", ebSolsticeZealot, "Creature — Rhino Cleric", 2, 3)
	bear := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	ebActivate(t, g, me, src, 0, 1, ebCardTarget(bear))
	if !ebLive(t, g, bear).Tapped {
		t.Error("the Bear is untapped")
	}
}

func TestSpontaneousArtistGrantsHaste(t *testing.T) {
	g, me, _ := p7Table(t)
	src := ebPush(g, me, "Spontaneous Artist", ebSpontaneousArtist, "Creature — Human Rogue", 3, 3)
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	ebActivate(t, g, me, src, 0, 1, ebCardTarget(bear))
	if !game.HasKeyword(ebLive(t, g, bear), "haste") {
		t.Error("the Bear does not have haste")
	}
}

func TestWhirlerVirtuosoMakesAThopter(t *testing.T) {
	g, me, _ := p7Table(t)
	src := ebPush(g, me, "Whirler Virtuoso", ebWhirlerVirtuoso, "Creature — Vedalken Artificer", 2, 3)
	ebActivate(t, g, me, src, 0, 3, nil)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" && c.Controller == me.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Thopters, want 1", n)
	}
}

func TestShieldedAetherThiefDraws(t *testing.T) {
	g, me, _ := p7Table(t)
	src := ebPush(g, me, "Shielded Aether Thief", ebShieldedAetherThief, "Creature — Vedalken Rogue", 0, 4)
	hand := len(me.Hand.Cards)
	ebActivate(t, g, me, src, 0, 3, nil)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d, want %d", len(me.Hand.Cards), hand+1)
	}
}

func TestAethersphereHarvesterGainsLifelink(t *testing.T) {
	g, me, _ := p7Table(t)
	src := ebPush(g, me, "Aethersphere Harvester", ebAethersphere, "Artifact — Vehicle", 3, 5)
	ebActivate(t, g, me, src, 0, 1, nil)
	if !game.HasKeyword(ebLive(t, g, src), "lifelink") {
		t.Error("the Harvester does not have lifelink")
	}
}
