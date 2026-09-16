package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fetchlands_test.go — the sacrifice-to-fetch land family, the two
// one-drop mana dorks and Lotus Petal, from the frequency-ordered
// Commander staples pass.

const (
	evolvingWildsOracle       = "a75445d3-1303-4bb5-89ad-26ea93fecd48"
	terramorphicExpanseOracle = "1bd3e453-aa21-4ee6-95c2-d6d920ee8e7a"
	pollutedDeltaOracle       = "ef86989d-ce80-4e55-aece-7d11710eeffa"
	llanowarElvesOracle       = "68954295-54e3-4303-a6bc-fc4547a4e3a3"
	elvishMysticOracle        = "3f3b2c10-21f8-4e13-be83-4ef3fa36e123"
	lotusPetalOracle          = "32e5339e-9e4f-46f8-b305-f9d6d3ba8bb5"
	bojukaBogOracle           = "04b7362d-0490-4cb0-b5d7-2a7732f659ce"
)

// tenFetchlands is the whole Zendikar / Onslaught cycle. Listing it
// here rather than spot-checking one card is deliberate: the ten
// files are generated from one template, so the failure mode this
// guards against is a file going missing in a merge, not a typo in
// the logic.
var tenFetchlands = map[string]string{
	"Polluted Delta":    pollutedDeltaOracle,
	"Flooded Strand":    "f3c7af78-a77d-4134-82a2-a5ce84285a84",
	"Misty Rainforest":  "09dd85aa-47bc-4713-a9b9-8b52ff2285ed",
	"Bloodstained Mire": "fc0707c7-d504-4ccf-a0d2-3eb6e26e7a57",
	"Windswept Heath":   "29737a60-3ebd-40d9-b935-c4f54b90d45d",
	"Wooded Foothills":  "6587a463-a108-4854-b6d1-944e89b8c8a4",
	"Verdant Catacombs": "67d60b24-d429-4ded-90d9-06e49f28c396",
	"Scalding Tarn":     "cb027150-848c-4a66-88ad-e20222304dd8",
	"Marsh Flats":       "dab520d0-20b4-4273-ba6b-eb07f85ea433",
	"Arid Mesa":         "c5acf2a5-40f4-433d-a74d-1cb56c521464",
}

// stapleLibraryCard seeds a named land into a player's library.
func stapleLibraryCard(p *game.Player, name, typeLine string) uuid.UUID {
	return pushLibraryCardForTest(p, game.Card{Name: name, TypeLine: typeLine})
}

func stapleCountLands(g *game.Game) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsLand() {
			n++
		}
	}
	return n
}

// --- wiring ------------------------------------------------------

// Every fetchland pays {T}, 1 life and itself — and nothing else.
// A stray mana component or a missing life payment would make the
// card strictly better than printed, which is the failure worth
// pinning.
func TestFetchlandCycleCostIsTapLifeSacrifice(t *testing.T) {
	for name, oracle := range tenFetchlands {
		abs := game.ActivatedAbilitiesForCard(game.Card{OracleID: oracle})
		if len(abs) != 1 {
			t.Errorf("%s: %d activated abilities, want 1", name, len(abs))
			continue
		}
		c := abs[0].Cost
		if !c.Tap || !c.SacrificeSelf || c.Life != 1 {
			t.Errorf("%s: cost %+v, want tap + sacrifice self + 1 life", name, c)
		}
		if c.Mana != "" || c.SacrificeOther != nil {
			t.Errorf("%s: cost has an extra component: %+v", name, c)
		}
		if abs[0].Targets != nil {
			t.Errorf("%s: a fetchland does not target", name)
		}
		if abs[0].SorcerySpeed {
			t.Errorf("%s: fetches are crackable at instant speed", name)
		}
	}
}

// Evolving Wilds and Terramorphic Expanse pay no life — that is the
// entire difference between the two families, so it gets its own
// assertion rather than riding along with the cycle above.
func TestSacrificeFetchBasicsPayNoLife(t *testing.T) {
	for name, oracle := range map[string]string{
		"Evolving Wilds":       evolvingWildsOracle,
		"Terramorphic Expanse": terramorphicExpanseOracle,
	} {
		abs := game.ActivatedAbilitiesForCard(game.Card{OracleID: oracle})
		if len(abs) != 1 {
			t.Fatalf("%s: %d activated abilities, want 1", name, len(abs))
		}
		c := abs[0].Cost
		if !c.Tap || !c.SacrificeSelf {
			t.Errorf("%s: cost %+v, want tap + sacrifice self", name, c)
		}
		if c.Life != 0 {
			t.Errorf("%s: costs %d life, want 0", name, c.Life)
		}
	}
}

// --- the fetch itself --------------------------------------------

// A fetchland takes a NONBASIC dual — "an Island or Swamp card" is
// any land with the type, which is the whole reason the cycle is
// played — and the land arrives UNTAPPED.
func TestFetchlandTakesNonbasicDualUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	delta := pushCatalogPermanent(g, me.ID, "Polluted Delta", "Land", pollutedDeltaOracle, false)
	grave := stapleLibraryCard(me, "Watery Grave", "Land — Island Swamp")
	lifeBefore := me.Life

	if err := g.ActivateCatalogAbility(me.ID, delta, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// Costs are paid at announce, before the ability resolves.
	if _, still := aangCardOnBF(g, delta); still {
		t.Error("the fetchland should be sacrificed as a cost, at announce")
	}
	if me.Life != lifeBefore-1 {
		t.Errorf("life %d -> %d, want one less", lifeBefore, me.Life)
	}

	passPriorityAroundTable(t, g)

	got, ok := aangCardOnBF(g, grave)
	if !ok {
		t.Fatal("Watery Grave was not fetched — a fetchland takes any land with the subtype, not just a basic")
	}
	if got.Tapped {
		t.Error("a fetched land arrives untapped; that is what the life pays for")
	}
}

// The subtype filter is real: a Delta cannot find a Taiga or a
// Plains. Getting this wrong would turn every fetch into a
// five-colour tutor.
func TestFetchlandRejectsWrongSubtypes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	delta := pushCatalogPermanent(g, me.ID, "Polluted Delta", "Land", pollutedDeltaOracle, false)
	stapleLibraryCard(me, "Taiga", "Land — Mountain Forest")
	stapleLibraryCard(me, "Plains", "Basic Land — Plains")
	before := stapleCountLands(g)
	lifeBefore := me.Life

	if err := g.ActivateCatalogAbility(me.ID, delta, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	// The Delta itself left, and nothing arrived to replace it.
	if got := stapleCountLands(g); got != before-1 {
		t.Errorf("lands on battlefield %d -> %d, want %d — an illegal land was fetched",
			before, got, before-1)
	}
	// A whiffed fetch still costs the life and the land (CR 601.2h).
	if me.Life != lifeBefore-1 {
		t.Errorf("life %d -> %d: the cost is paid even when the search finds nothing",
			lifeBefore, me.Life)
	}
}

// Evolving Wilds fetches a BASIC, tapped, for no life — and cannot
// take the nonbasic dual sitting next to it in the library.
func TestEvolvingWildsFetchesBasicTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wilds := pushCatalogPermanent(g, me.ID, "Evolving Wilds", "Land", evolvingWildsOracle, false)
	dual := stapleLibraryCard(me, "Watery Grave", "Land — Island Swamp")
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	lifeBefore := me.Life

	if err := g.ActivateCatalogAbility(me.ID, wilds, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	got, ok := aangCardOnBF(g, forest)
	if !ok {
		t.Fatal("Evolving Wilds did not fetch the basic Forest")
	}
	if !got.Tapped {
		t.Error("Evolving Wilds puts the land onto the battlefield TAPPED")
	}
	if _, wrong := aangCardOnBF(g, dual); wrong {
		t.Error("Evolving Wilds fetched a nonbasic — its predicate is basic-land only")
	}
	if me.Life != lifeBefore {
		t.Errorf("life %d -> %d: Evolving Wilds costs no life", lifeBefore, me.Life)
	}
}

// --- mana dorks --------------------------------------------------

func TestLlanowarElvesTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := pushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, false)

	if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "G" {
		t.Errorf("mana pool %+v, want 1xG", me.ManaPool)
	}
	// A fixed colour needs no picker — only pipes queue a choice.
	if len(g.PendingChoices) != 0 {
		t.Errorf("unexpected PendingChoice for a fixed-colour ability: %+v", g.PendingChoices)
	}
}

// A summoning-sick creature cannot tap for mana (CR 302.6).
//
// This test shipped in #243 pinning the OPPOSITE — a known engine
// gap where Game.ActivateManaAbility checked card.Tapped and nothing
// else, so a turn-one Elf tapped for mana. #233 closed the gap in
// the same hour, and #243's own comment named flipping this test as
// the fix's acceptance criterion. The two merged three minutes
// apart without either seeing the other, which is why the flip
// arrives separately.
//
// Validate-before-pay: the refused activation leaves the Elf
// untapped and the pool empty, not a tapped Elf with no mana.
func TestLlanowarElvesCannotTapWhileSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := pushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, true)

	if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{}); err != game.ErrSummoningSick {
		t.Fatalf("summoning-sick Elf: got %v, want ErrSummoningSick", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool %+v, want empty — a sick creature produced mana", me.ManaPool)
	}
	if got, ok := aangCardOnBF(g, elf); !ok || got.Tapped {
		t.Error("a refused activation must not tap the Elf")
	}
}

// Elvish Mystic is the same card under another name; assert the
// shapes match rather than repeating the behaviour tests.
func TestElvishMysticMatchesLlanowarElves(t *testing.T) {
	elf := game.ManaAbilitiesForCard(game.Card{OracleID: llanowarElvesOracle})
	mystic := game.ManaAbilitiesForCard(game.Card{OracleID: elvishMysticOracle})
	if len(elf) != 1 || len(mystic) != 1 {
		t.Fatalf("mana abilities: elves %d, mystic %d, want 1 each", len(elf), len(mystic))
	}
	if elf[0].Produced != mystic[0].Produced || mystic[0].Produced != "{G}" {
		t.Errorf("produced: elves %q, mystic %q, want {G}", elf[0].Produced, mystic[0].Produced)
	}
	if !mystic[0].TapCost || mystic[0].SacrificeCost {
		t.Errorf("Elvish Mystic cost: %+v, want tap only", mystic[0])
	}
}

// --- Lotus Petal -------------------------------------------------

// The Petal is the first catalog card whose MANA ability sacrifices
// its source, so this exercises a wire path nothing else in the
// catalog reaches.
func TestLotusPetalCracksForAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	petal := pushArtifactToBattlefieldForTest(g, me.ID, "Lotus Petal", "Artifact", lotusPetalOracle)

	if err := g.ActivateManaAbility(me.ID, petal, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if _, still := aangCardOnBF(g, petal); still {
		t.Error("Lotus Petal should be sacrificed by its own mana ability")
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("PendingChoices %d, want 1 colour pick", len(g.PendingChoices))
	}
	choice := g.PendingChoices[0]
	if len(choice.ColorOptions) != 5 {
		t.Errorf("ColorOptions %v, want all five", choice.ColorOptions)
	}
	if err := g.ResolveManaChoice(choice.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "R" {
		t.Errorf("mana pool %+v, want 1xR", me.ManaPool)
	}
}

// --- Bojuka Bog --------------------------------------------------

// bogPlay moves a Bojuka Bog from hand to battlefield, the path a
// played land takes.
func bogPlay(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	var p *game.Player
	for _, s := range g.Seats {
		if s != nil && s.ID == owner {
			p = s
		}
	}
	if p == nil {
		t.Fatalf("no seat %s", owner)
	}
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Bojuka Bog", TypeLine: "Land",
		OracleID: bojukaBogOracle, Owner: owner, Controller: owner,
	})
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner},
		game.ZoneRef{Kind: game.ZoneBattlefield},
		id,
	); err != nil {
		t.Fatalf("play Bojuka Bog: %v", err)
	}
	return id
}

func TestBojukaBogEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	bog := bogPlay(t, g, g.Seats[0].ID)

	got, ok := aangCardOnBF(g, bog)
	if !ok {
		t.Fatal("Bojuka Bog never reached the battlefield")
	}
	if !got.Tapped {
		t.Error("Bojuka Bog enters tapped; that is its cost for being free removal")
	}
}

// The ETB empties the CHOSEN player's graveyard, all of it, and
// leaves everyone else's alone.
func TestBojukaBogExilesTargetPlayersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	for i := 0; i < 3; i++ {
		pushGraveyardCardForTest(opp, "Buried")
	}
	pushGraveyardCardForTest(other, "Not Yours")
	exileBefore := len(g.Exile.Cards)

	bogPlay(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	if n := len(opp.Graveyard.Cards); n != 0 {
		t.Errorf("target graveyard still holds %d cards, want 0 — the loop skipped cards", n)
	}
	if n := len(other.Graveyard.Cards); n != 1 {
		t.Errorf("an untargeted player's graveyard was touched: %d cards, want 1", n)
	}
	if n := len(g.Exile.Cards); n != exileBefore+3 {
		t.Errorf("exile holds %d cards, want %d", n, exileBefore+3)
	}
}

func TestBojukaBogTapsForBlack(t *testing.T) {
	abs := game.ManaAbilitiesForCard(game.Card{OracleID: bojukaBogOracle})
	if len(abs) != 1 {
		t.Fatalf("%d mana abilities, want 1", len(abs))
	}
	if abs[0].Produced != "{B}" || !abs[0].TapCost {
		t.Errorf("mana ability %+v, want tap for {B}", abs[0])
	}
}
