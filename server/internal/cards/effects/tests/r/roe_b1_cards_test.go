package testsr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsbl "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/bl"
	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	cardsgr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/gr"
	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// roe_b1_cards_test.go — Rise of the Eldrazi, batch 1 (#5581, S82):
// thirty-nine cards built from seams the engine already has. Each test
// plays the real catalog entry.

// roeB1Oracles names every card in the batch by its oracle ID.
var roeB1Oracles = map[string]string{
	"Bloodthrone Vampire":     "aa76cc50-5be0-4844-b3a4-fb33fca3cdd3",
	"Brimstone Mage":          "81dcb85e-d61f-430b-9540-5a3bb1378d5a",
	"Brood Birthing":          "e7d1c762-69d0-401d-8035-6e9744baf9d8",
	"Cadaver Imp":             "7265d768-b79f-44eb-9873-d834e55f6463",
	"Corpsehatch":             "5a6c9376-313b-4420-890e-5aa78fa6382f",
	"Daggerback Basilisk":     "ea4d30c7-7c39-46d5-9b71-a7ebc3e219a1",
	"Death Cultist":           "edbde032-4432-4610-83a2-a52f6ce199e1",
	"Deathless Angel":         "6098ba4b-ebfe-48da-a33d-b94c021ac559",
	"Demystify":               "fd591199-9f7a-4147-a150-13279dbb4498",
	"Deprive":                 "1065eceb-9150-4ea4-bd6d-acaabda14e00",
	"Dread Drone":             "b6821131-9ae1-4a3c-8235-90deae33c05f",
	"Emrakul's Hatcher":       "1ee811fb-c25d-4cf5-b396-539c36e8f90f",
	"Enclave Cryptologist":    "bc0fc535-357a-4d83-8d64-34337222cc4b",
	"Essence Feed":            "c29b997a-5be6-4d51-a81b-d00677ef401d",
	"Goblin Tunneler":         "1005bc16-8a6a-476a-a647-4ed40d2e96de",
	"Gravitational Shift":     "0411aead-9f77-435d-a343-635f23221cf7",
	"Guard Gomazoa":           "7c565975-aebc-4599-ac32-5594c718e2cb",
	"Guul Draz Assassin":      "c6fb6971-112b-4c82-a0a3-4acdde970b30",
	"Joraga Treespeaker":      "a011aefe-5853-4435-b686-179e931ba897",
	"Kabira Vindicator":       "c43c76d5-aa11-4fce-911a-ec60cb06ce2a",
	"Kazandu Tuskcaller":      "ee60ccae-b3ae-47a1-a0bb-2fe51645e95d",
	"Kozilek's Predator":      "452ef7d0-1351-40e6-9379-59cd8bb31da5",
	"Lighthouse Chronologist": "b19ff0dd-2041-461f-aa62-6dc273e4bbe5",
	"Lone Missionary":         "626f1dc3-3b14-4873-902f-ffec5487a97d",
	"Mnemonic Wall":           "f2933616-211b-4d3f-8064-2be61c26079a",
	"Mortician Beetle":        "dc5c4a38-a415-466a-a7ea-dc04eee7f691",
	"Nest Invader":            "a7584a28-0829-4897-abf4-6be1a9232347",
	"Pennon Blade":            "b97e6acd-a676-429e-b875-b734759ebea7",
	"Pestilence Demon":        "2572add6-c0b1-44bd-8713-1e94e3fe69d6",
	"Rapacious One":           "3af03c99-d215-477b-ba49-51f4e87907ff",
	"Repay in Kind":           "24b8177f-5697-4465-9305-cb84307972ae",
	"Runed Servitor":          "4e5e81e1-ab47-4f08-9c42-dba0adc5cf9a",
	"Sea Gate Oracle":         "b333c194-d285-4dfb-984c-57b7e88393af",
	"Skittering Invasion":     "cf890abd-b282-4dbb-ac86-9e7caec1c1b8",
	"Spawning Breath":         "82ce3682-0514-49a5-a330-2591e9789d36",
	"Sporecap Spider":         "e9b5d4f3-b98a-463a-b628-289c2df81204",
	"Stalwart Shield-Bearers": "55aafc25-76c5-4621-941b-754bd8f13e64",
	"Student of Warfare":      "0bc2875b-3a5a-4885-86b9-16ab7330ac10",
	"Vendetta":                "50a16634-ebf7-477f-ad6f-23332599f838",
}

func roeB1Oracle(t *testing.T, name string) string {
	t.Helper()
	o, ok := roeB1Oracles[name]
	if !ok {
		t.Fatalf("%s is not in the batch", name)
	}
	return o
}

func TestRoeB1CardsAreRegisteredFull(t *testing.T) {
	if len(roeB1Oracles) != 39 {
		t.Errorf("%d cards in the batch, want 39", len(roeB1Oracles))
	}
	for name, oracle := range roeB1Oracles {
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

func TestRoeB1PrintedKeywords(t *testing.T) {
	for name, want := range map[string][]string{
		"Sporecap Spider":         {"reach"},
		"Daggerback Basilisk":     {"deathtouch"},
		"Guard Gomazoa":           {"defender", "flying"},
		"Mnemonic Wall":           {"defender"},
		"Stalwart Shield-Bearers": {"defender"},
		"Cadaver Imp":             {"flying"},
		"Pestilence Demon":        {"flying"},
		"Deathless Angel":         {"flying"},
		"Rapacious One":           {"trample"},
	} {
		spec, _ := effects.Lookup(roeB1Oracle(t, name))
		if len(spec.PrintedKeywords) != len(want) {
			t.Errorf("%s: keywords %v, want %v", name, spec.PrintedKeywords, want)
			continue
		}
		for i := range want {
			if spec.PrintedKeywords[i] != want[i] {
				t.Errorf("%s: keywords %v, want %v", name, spec.PrintedKeywords, want)
			}
		}
	}
}

// --- helpers -----------------------------------------------------------------

// roeB1Push puts a catalog permanent onto seat `owner`'s battlefield,
// untapped and able to attack and tap.
func roeB1Push(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	t.Helper()
	return effectstest.B12Push(g, owner, name, typeLine, roeB1Oracle(t, name), power, toughness)
}

// roeB1Level adds level counters to a leveler.
func roeB1Level(t *testing.T, g *game.Game, id uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterLevel, n); err != nil {
			t.Fatalf("level counters: %v", err)
		}
	})
}

// roeB1Size checks a creature's power and toughness after the layers.
func roeB1Size(t *testing.T, g *game.Game, id uuid.UUID, power, toughness int, what string) {
	t.Helper()
	if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != power || tg != toughness {
		t.Errorf("%s is %d/%d, want %d/%d", what, p, tg, power, toughness)
	}
}

// roeB1ActivateGrant activates a leveler's granted bundle row: the
// leveler's own level up is row 0, so the one granted row is row 1.
func roeB1ActivateGrant(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, key string, params game.ActivateAbilityParams) error {
	t.Helper()
	params.Ref = "grant:" + key + ":0:0"
	return g.ActivateCatalogAbility(p.ID, source, 1, params)
}

// roeB1SpawnCount counts a player's Eldrazi Spawn tokens.
func roeB1SpawnCount(g *game.Game, controller uuid.UUID) int {
	return effectstest.CountOnBattlefield(g, "Eldrazi Spawn", controller)
}

// roeB1Cast casts a catalog spell from the active seat's hand with the
// given parameters, without resolving it.
func roeB1Cast(t *testing.T, g *game.Game, name, typeLine string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: roeB1Oracle(t, name),
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	return id, g.CastSpell(me.ID, id, params)
}

// --- levelers ----------------------------------------------------------------

// Level 1-4: the Treespeaker taps for {G}{G}. Level 5+: every Elf its
// controller controls does, and nothing else (CR 711.2a, 711.2b, 711.5).
func TestRoeB1JoragaTreespeakerGrantsItsManaBySize(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	joraga := roeB1Push(t, g, me.ID, "Joraga Treespeaker", "Creature — Elf Druid", 1, 1)
	elf := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf Warrior", 1, 1)
	theirElf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf Warrior", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if n := effectstest.GrantedManaRowsOf(g, joraga); n != 0 {
		t.Fatalf("level 0: %d granted mana rows, want 0", n)
	}
	roeB1Size(t, g, joraga, 1, 1, "a level 0 Treespeaker")

	roeB1Level(t, g, joraga, 1)
	roeB1Size(t, g, joraga, 1, 2, "a level 1 Treespeaker")
	if n := effectstest.GrantedManaRowsOf(g, joraga); n != 1 {
		t.Fatalf("level 1: the Treespeaker has %d granted mana rows, want 1", n)
	}
	if n := effectstest.GrantedManaRowsOf(g, elf); n != 0 {
		t.Errorf("level 1: the other Elf has %d granted mana rows, want 0", n)
	}
	if err := effectstest.TapGrantedMana(t, g, me.ID, joraga, "G"); err != nil {
		t.Fatalf("tap the Treespeaker: %v", err)
	}
	if got := effectstest.PoolCount(me, "G"); got != 2 {
		t.Errorf("pool holds %d {G}, want 2", got)
	}

	roeB1Level(t, g, joraga, 4)
	roeB1Size(t, g, joraga, 1, 4, "a level 5 Treespeaker")
	for _, c := range []struct {
		id   uuid.UUID
		what string
		want int
	}{{joraga, "the Treespeaker", 1}, {elf, "my other Elf", 1}, {theirElf, "an opponent's Elf", 0}, {bear, "my Bear", 0}} {
		if n := effectstest.GrantedManaRowsOf(g, c.id); n != c.want {
			t.Errorf("level 5: %s has %d granted mana rows, want %d", c.what, n, c.want)
		}
	}
}

// LEVEL 7+: at each end step that isn't its controller's turn, its
// controller takes an extra turn; its own end step gives nothing, and
// neither does a lower level.
func TestRoeB1LighthouseChronologistTakesExtraTurnsOnOtherTurns(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	mine := roeB1Push(t, g, me.ID, "Lighthouse Chronologist", "Creature — Human Wizard", 1, 3)
	theirs := roeB1Push(t, g, opp.ID, "Lighthouse Chronologist", "Creature — Human Wizard", 1, 3)
	low := roeB1Push(t, g, g.Seats[2].ID, "Lighthouse Chronologist", "Creature — Human Wizard", 1, 3)
	roeB1Level(t, g, mine, 7)
	roeB1Level(t, g, theirs, 7)
	roeB1Level(t, g, low, 6)
	roeB1Size(t, g, theirs, 3, 5, "a level 7 Chronologist")
	roeB1Size(t, g, low, 2, 4, "a level 6 Chronologist")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.ExtraTurns) != 1 {
		t.Fatalf("%d extra turns queued, want 1: %+v", len(g.ExtraTurns), g.ExtraTurns)
	}
	if seat := g.ExtraTurns[0].Seat; g.Seats[seat].ID != opp.ID {
		t.Errorf("the extra turn is seat %d's, want the opponent's", seat)
	}
}

func TestRoeB1StudentOfWarfareLevels(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	student := roeB1Push(t, g, me.ID, "Student of Warfare", "Creature — Human Knight", 1, 1)
	have := 0
	for _, step := range []struct {
		level, pt  int
		first, dbl bool
	}{{0, 1, false, false}, {1, 1, false, false}, {2, 3, true, false}, {6, 3, true, false}, {7, 4, false, true}} {
		roeB1Level(t, g, student, step.level-have)
		have = step.level
		roeB1Size(t, g, student, step.pt, step.pt, "the Student")
		if got := effectstest.HasEffectiveKeyword(t, g, student, "first strike"); got != step.first {
			t.Errorf("level %d: first strike %v, want %v", step.level, got, step.first)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, student, "double strike"); got != step.dbl {
			t.Errorf("level %d: double strike %v, want %v", step.level, got, step.dbl)
		}
	}
	// Level up costs {W} and is sorcery speed (CR 702.87a).
	effectstest.ApaMana(me, "W")
	effectstest.P7Activate(t, g, me, student, 0, game.ActivateAbilityParams{Strict: true})
	if got := effectstest.CountersOn(g, student, game.CounterLevel); got != 8 {
		t.Errorf("%d level counters after leveling up, want 8", got)
	}
}

func TestRoeB1KabiraVindicatorAnthemFollowsItsBand(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	kabira := roeB1Push(t, g, me.ID, "Kabira Vindicator", "Creature — Human Knight", 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	roeB1Size(t, g, bear, 2, 2, "my Bear at level 0")
	roeB1Level(t, g, kabira, 2)
	roeB1Size(t, g, kabira, 3, 6, "a level 2 Vindicator")
	roeB1Size(t, g, bear, 3, 3, "my Bear at level 2")
	roeB1Level(t, g, kabira, 3)
	roeB1Size(t, g, kabira, 4, 8, "a level 5 Vindicator")
	roeB1Size(t, g, bear, 4, 4, "my Bear at level 5")
	roeB1Size(t, g, theirs, 2, 2, "an opponent's Bear")
}

func TestRoeB1GuulDrazAssassinShrinksByBand(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	assassin := roeB1Push(t, g, me.ID, "Guul Draz Assassin", "Creature — Vampire Assassin", 1, 1)
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
	effectstest.ApaMana(me, "B")
	if err := roeB1ActivateGrant(t, g, me, assassin, cardsgr.GuulDrazAssassinTwo, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ogre)}); err == nil {
		t.Fatal("a level 0 Assassin activated its LEVEL 2-3 ability")
	}
	roeB1Level(t, g, assassin, 2)
	if err := roeB1ActivateGrant(t, g, me, assassin, cardsgr.GuulDrazAssassinTwo, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ogre)}); err != nil {
		t.Fatalf("activate -2/-2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	roeB1Size(t, g, ogre, 3, 3, "the Ogre after -2/-2")
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(assassin) })
	roeB1Level(t, g, assassin, 2)
	roeB1Size(t, g, assassin, 4, 4, "a level 4 Assassin")
	effectstest.ApaMana(me, "B")
	if err := roeB1ActivateGrant(t, g, me, assassin, cardsgr.GuulDrazAssassinFour, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ogre)}); err != nil {
		t.Fatalf("activate -4/-4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the Ogre survived -2/-2 and -4/-4")
	}
}

func TestRoeB1EnclaveCryptologistLootsThenDraws(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	crypt := roeB1Push(t, g, me.ID, "Enclave Cryptologist", "Creature — Merfolk Wizard", 0, 1)
	roeB1Level(t, g, crypt, 1)
	before := effectstest.HandSize(me)
	if err := roeB1ActivateGrant(t, g, me, crypt, cardse.EnclaveCryptologistLoot, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("loot: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerChooseCards(t, g, me.ID, me.Hand.Cards[0].InstanceID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.HandSize(me); got != before {
		t.Errorf("hand %d after a loot, want %d", got, before)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(crypt) })
	roeB1Level(t, g, crypt, 2)
	if err := roeB1ActivateGrant(t, g, me, crypt, cardse.EnclaveCryptologistDraw, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != before+1 {
		t.Errorf("hand %d after a draw, want %d", got, before+1)
	}
}

func TestRoeB1KazanduTuskcallerMakesElephantsByBand(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	tusk := roeB1Push(t, g, me.ID, "Kazandu Tuskcaller", "Creature — Human Shaman", 1, 1)
	roeB1Level(t, g, tusk, 2)
	if err := roeB1ActivateGrant(t, g, me, tusk, cardsj.KazanduTuskcallerOne, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("one Elephant: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Elephant", me.ID); n != 1 {
		t.Fatalf("%d Elephants, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(tusk) })
	roeB1Level(t, g, tusk, 4)
	if err := roeB1ActivateGrant(t, g, me, tusk, cardsj.KazanduTuskcallerTwo, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("two Elephants: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Elephant", me.ID); n != 3 {
		t.Errorf("%d Elephants, want 3", n)
	}
}

func TestRoeB1BrimstoneMagePingsByBand(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	mage := roeB1Push(t, g, me.ID, "Brimstone Mage", "Creature — Human Shaman", 2, 2)
	roeB1Level(t, g, mage, 3)
	roeB1Size(t, g, mage, 2, 4, "a level 3 Mage")
	before := opp.Life
	if err := roeB1ActivateGrant(t, g, me, mage, cardsbl.BrimstoneMageThree, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("ping: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("opponent at %d, want %d", opp.Life, before-3)
	}
}

// --- Eldrazi Spawn -------------------------------------------------------------

func TestRoeB1SpawnMakersEnter(t *testing.T) {
	for _, c := range []struct {
		name, typeLine string
		n              int
	}{
		{"Emrakul's Hatcher", "Creature — Eldrazi Drone", 3},
		{"Kozilek's Predator", "Creature — Eldrazi Drone", 2},
		{"Nest Invader", "Creature — Eldrazi Drone", 1},
		{"Dread Drone", "Creature — Eldrazi Drone", 2},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		effectstest.EnterFromHand(t, g, me.ID, c.name, c.typeLine, roeB1Oracle(t, c.name))
		effectstest.PassPriorityAroundTable(t, g)
		if got := roeB1SpawnCount(g, me.ID); got != c.n {
			t.Errorf("%s: %d Eldrazi Spawn, want %d", c.name, got, c.n)
		}
	}
}

func TestRoeB1SkitteringInvasionMakesFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Skittering Invasion", "Kindred Sorcery — Eldrazi", roeB1Oracle(t, "Skittering Invasion"), nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := roeB1SpawnCount(g, me.ID); got != 5 {
		t.Errorf("%d Eldrazi Spawn, want 5", got)
	}
}

func TestRoeB1BroodBirthingMakesThreeWithASpawn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Brood Birthing", "Sorcery", roeB1Oracle(t, "Brood Birthing"), nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := roeB1SpawnCount(g, me.ID); got != 1 {
		t.Fatalf("with no Spawn: %d Eldrazi Spawn, want 1", got)
	}
	effectstest.CastCatalogSpell(t, g, "Brood Birthing", "Sorcery", roeB1Oracle(t, "Brood Birthing"), nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := roeB1SpawnCount(g, me.ID); got != 4 {
		t.Errorf("with a Spawn: %d Eldrazi Spawn, want 4", got)
	}
}

func TestRoeB1SpawningBreathPingsAndMakesASpawn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Ons2Opponent(g)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Spawning Breath", "Instant", roeB1Oracle(t, "Spawning Breath"), []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 || roeB1SpawnCount(g, me.ID) != 1 {
		t.Errorf("opponent at %d (want %d), %d Spawn (want 1)", opp.Life, before-1, roeB1SpawnCount(g, me.ID))
	}
}

func TestRoeB1CorpsehatchDestroysANonblackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Ons2Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	zombie := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Zombie", TypeLine: "Creature — Zombie", ManaCost: "{1}{B}",
		Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Corpsehatch", "Sorcery", roeB1Oracle(t, "Corpsehatch"), effectstest.CardRefs(zombie)); err == nil {
		t.Error("Corpsehatch targeted a black creature")
	}
	effectstest.CastCatalogSpell(t, g, "Corpsehatch", "Sorcery", roeB1Oracle(t, "Corpsehatch"), effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || roeB1SpawnCount(g, me.ID) != 2 {
		t.Errorf("bear alive %v, %d Spawn (want dead, 2)", g.Battlefield.Contains(bear), roeB1SpawnCount(g, me.ID))
	}
}

func TestRoeB1EssenceFeedDrainsAndMakesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Ons2Opponent(g)
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Essence Feed", "Sorcery", roeB1Oracle(t, "Essence Feed"), []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-3 || me.Life != mine+3 || roeB1SpawnCount(g, me.ID) != 3 {
		t.Errorf("opponent %d (want %d), me %d (want %d), %d Spawn (want 3)", opp.Life, theirs-3, me.Life, mine+3, roeB1SpawnCount(g, me.ID))
	}
}

func TestRoeB1RapaciousOneMakesASpawnPerDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	one := roeB1Push(t, g, me.ID, "Rapacious One", "Creature — Eldrazi Drone", 5, 4)
	effectstest.AttackWith(t, g, opp.ID, one)
	effectstest.PassPriorityAroundTable(t, g)
	if got := roeB1SpawnCount(g, me.ID); got != 5 {
		t.Errorf("%d Eldrazi Spawn, want 5", got)
	}
}

// --- the rest --------------------------------------------------------------------

func TestRoeB1StalwartShieldBearersPumpsOtherDefenders(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	bearers := roeB1Push(t, g, me.ID, "Stalwart Shield-Bearers", "Creature — Human Soldier", 0, 3)
	wall := roeB1Push(t, g, me.ID, "Mnemonic Wall", "Creature — Wall", 0, 4)
	theirWall := roeB1Push(t, g, opp.ID, "Mnemonic Wall", "Creature — Wall", 0, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	roeB1Size(t, g, bearers, 0, 3, "the Shield-Bearers")
	roeB1Size(t, g, wall, 0, 6, "my other defender")
	roeB1Size(t, g, theirWall, 0, 4, "an opponent's defender")
	roeB1Size(t, g, bear, 2, 2, "my Bear")
}

func TestRoeB1GravitationalShiftSplitsByFlying(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	roeB1Push(t, g, me.ID, "Gravitational Shift", "Enchantment", 0, 0)
	imp := roeB1Push(t, g, opp.ID, "Cadaver Imp", "Creature — Imp", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	roeB1Size(t, g, imp, 3, 1, "a flier")
	roeB1Size(t, g, bear, 0, 2, "a creature without flying")
}

func TestRoeB1MnemonicWallReturnsAnInstantOrSorcery(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bolt := effectstest.PushGraveyardCardTyped(me, "Bolt", "Instant")
		bear := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
		effectstest.CastHoldCreature(t, g, "Mnemonic Wall", "Creature — Wall", roeB1Oracle(t, "Mnemonic Wall"), 0, 4)
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil || !effectstest.HasID(p.PickTargetCards, bolt) || effectstest.HasID(p.PickTargetCards, bear) {
			t.Fatalf("pick: %+v", p)
		}
		effectstest.PickCard(t, g, me.ID, bolt)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, accept)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Contains(bolt) != accept {
			t.Errorf("accept=%v: instant in hand = %v", accept, me.Hand.Contains(bolt))
		}
	}
}

func TestRoeB1CadaverImpReturnsACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bolt := effectstest.PushGraveyardCardTyped(me, "Bolt", "Instant")
	bear := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	effectstest.CastHoldCreature(t, g, "Cadaver Imp", "Creature — Imp", roeB1Oracle(t, "Cadaver Imp"), 1, 1)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || !effectstest.HasID(p.PickTargetCards, bear) || effectstest.HasID(p.PickTargetCards, bolt) {
		t.Fatalf("pick: %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature card did not return")
	}
}

func TestRoeB1RepayInKindLevelsEveryLifeTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	for i, life := range []int{20, 7, 31, 12} {
		g.Seats[i].Life = life
	}
	effectstest.CastCatalogSpell(t, g, "Repay in Kind", "Sorcery", roeB1Oracle(t, "Repay in Kind"), nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != 7 {
			t.Errorf("seat %d at %d life, want 7", i, p.Life)
		}
	}
}

func TestRoeB1SeaGateOracleTakesOneAndBottomsTheOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	me.Library.Cards = nil
	keep, bottom, deep := uuid.New(), uuid.New(), uuid.New()
	for _, c := range []struct {
		id   uuid.UUID
		name string
	}{{deep, "Deep"}, {keep, "Keep"}, {bottom, "Bottom"}} {
		me.Library.PushTop(game.Card{InstanceID: c.id, Name: c.name, TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.EnterFromHand(t, g, me.ID, "Sea Gate Oracle", "Creature — Human Wizard", roeB1Oracle(t, "Sea Gate Oracle"))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, keep)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(keep) {
		t.Fatal("the chosen card is not in hand")
	}
	if me.Library.Size() != 2 || !me.Library.Contains(bottom) {
		t.Errorf("library holds %d cards, want Deep and Bottom", me.Library.Size())
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != deep {
		t.Errorf("top of library %v (%v), want Deep", top.Name, err)
	}
}

// Bloodthrone Vampire's sacrifice feeds Mortician Beetle, which counts any
// player's sacrifice of a creature.
func TestRoeB1BloodthroneVampireFeedsMorticianBeetle(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	vampire := roeB1Push(t, g, me.ID, "Bloodthrone Vampire", "Creature — Vampire", 1, 1)
	beetle := roeB1Push(t, g, me.ID, "Mortician Beetle", "Creature — Insect", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, vampire, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("sacrifice the Bear: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	roeB1Size(t, g, vampire, 3, 3, "the Vampire")
	if got := effectstest.CountersOn(g, beetle, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the Beetle, want 1", got)
	}
}

func TestRoeB1DeprivesReturnsALandAndCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if _, err := roeB1Cast(t, g, "Deprive", "Instant", game.CastSpellParams{}); err == nil {
		t.Error("Deprive was cast with no land to return")
	}
	land := effectstest.PushLand(g, me.ID, "Island")
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	if _, err := roeB1Cast(t, g, "Deprive", "Instant", game.CastSpellParams{
		Targets: effectstest.CardRefs(spell), ReturnIDs: []uuid.UUID{land},
	}); err != nil {
		t.Fatalf("cast Deprive: %v", err)
	}
	if !me.Hand.Contains(land) {
		t.Error("the land did not return to hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(spell) || g.Battlefield.Contains(spell) {
		t.Error("the spell was not countered")
	}
}

func TestRoeB1GuardGomazoaTakesNoCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	gomazoa := roeB1Push(t, g, opp.ID, "Guard Gomazoa", "Creature — Jellyfish", 1, 3)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, ogre, gomazoa)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if got := effectstest.Pr6Marked(g, gomazoa); got != 0 {
		t.Errorf("Gomazoa has %d combat damage marked, want 0", got)
	}
	if got := effectstest.Pr6Marked(g, ogre); got != 1 {
		t.Errorf("the Ogre has %d damage marked, want 1", got)
	}
	effectstest.Pr6Damage(t, g, ogre, gomazoa, 1)
	if got := effectstest.Pr6Marked(g, gomazoa); got != 1 {
		t.Errorf("Gomazoa has %d noncombat damage marked, want 1", got)
	}
}

func TestRoeB1PestilenceDemonHitsEveryCreatureAndPlayer(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	demon := roeB1Push(t, g, me.ID, "Pestilence Demon", "Creature — Demon", 7, 6)
	elf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	mine, theirs := me.Life, opp.Life
	effectstest.ApaMana(me, "B")
	effectstest.P7Activate(t, g, me, demon, 0, game.ActivateAbilityParams{Strict: true})
	if g.Battlefield.Contains(elf) {
		t.Error("the 1/1 survived")
	}
	if got := effectstest.Pr6Marked(g, demon); got != 1 {
		t.Errorf("the Demon has %d damage marked, want 1", got)
	}
	if me.Life != mine-1 || opp.Life != theirs-1 {
		t.Errorf("life %d / %d, want %d / %d", me.Life, opp.Life, mine-1, theirs-1)
	}
}

func TestRoeB1VendettaCostsTheToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Ons2Opponent(g)
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 5)
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Vendetta", "Instant", roeB1Oracle(t, "Vendetta"), effectstest.CardRefs(wall))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the creature survived")
	}
	if me.Life != before-5 {
		t.Errorf("life %d, want %d", me.Life, before-5)
	}
}

func TestRoeB1GoblinTunnelerUnblocksASmallCreature(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	tunneler := roeB1Push(t, g, me.ID, "Goblin Tunneler", "Creature — Goblin Rogue", 1, 1)
	big := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	small := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, tunneler, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(big)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("targeting power 3: %v, want ErrIllegalTarget", err)
	}
	effectstest.P7Activate(t, g, me, tunneler, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(small)})
	if effectstest.EffectiveOf(t, g, small).Restrictions&game.CantBeBlocked == 0 {
		t.Error("the Bear can be blocked")
	}
}

func TestRoeB1DemystifyDestroysAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Ons2Opponent(g)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Demystify", "Instant", roeB1Oracle(t, "Demystify"), effectstest.CardRefs(ench))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
}

func TestRoeB1PennonBladeCountsYourCreatures(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	blade := roeB1Push(t, g, me.ID, "Pennon Blade", "Artifact — Equipment", 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	effectstest.ApaMana(me, "C", "C", "C", "C")
	effectstest.P7Activate(t, g, me, blade, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true})
	roeB1Size(t, g, bear, 4, 4, "the equipped Bear")
}

func TestRoeB1RunedServitorEachPlayerDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	servitor := roeB1Push(t, g, me.ID, "Runed Servitor", "Artifact Creature — Construct", 2, 2)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = effectstest.HandSize(p)
	}
	effectstest.KillCreature(t, g, me.ID, servitor)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if effectstest.HandSize(p) != before[i]+1 {
			t.Errorf("seat %d hand %d, want %d", i, effectstest.HandSize(p), before[i]+1)
		}
	}
}

func TestRoeB1DeathCultistDrains(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	cultist := roeB1Push(t, g, me.ID, "Death Cultist", "Creature — Human Wizard", 1, 1)
	mine, theirs := me.Life, opp.Life
	effectstest.P7Activate(t, g, me, cultist, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if g.Battlefield.Contains(cultist) {
		t.Error("the Cultist was not sacrificed")
	}
	if me.Life != mine+1 || opp.Life != theirs-1 {
		t.Errorf("life %d / %d, want %d / %d", me.Life, opp.Life, mine+1, theirs-1)
	}
}

func TestRoeB1DeathlessAngelGrantsIndestructible(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	angel := roeB1Push(t, g, me.ID, "Deathless Angel", "Creature — Angel", 5, 7)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "W", "W")
	effectstest.P7Activate(t, g, me, angel, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true})
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the Bear is not indestructible")
	}
}

func TestRoeB1LoneMissionaryGainsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	before := me.Life
	effectstest.EnterFromHand(t, g, me.ID, "Lone Missionary", "Creature — Kor Monk", roeB1Oracle(t, "Lone Missionary"))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+4 {
		t.Errorf("life %d, want %d", me.Life, before+4)
	}
}
