package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mir_b2_cards_test.go — Mirage, batch 2 (#5633): the flanking creatures
// #5757 made possible, plus four that need only printed keywords or
// regeneration. Each test plays the real catalog entry; the flanking
// blocks use cards as the deck importer hands them to a game
// (flankingImport), so the keyword comes from the real records.

const (
	mirB2BurningShieldOracle = "042fd95a-5e85-4ce5-bb68-066a295c54fb"
	mirB2CadaverousOracle    = "250a027f-2b3e-4eaf-899a-a9da8be20070"
	mirB2FemerefKnightOracle = "4365521e-996d-4eca-bbc9-b0988fa973ae"
	mirB2CentaurOracle       = "39b699e1-79c6-4960-97dc-d58574181bbe"
	mirB2HerderOracle        = "0edf5bac-36a9-4f35-afd1-73f7e1aab51b"
	mirB2SearingSpearOracle  = "d3c68573-95b1-4db5-bbea-363b86f713c1"
	mirB2SidarJabariOracle   = "cb637dfd-4022-43a4-8632-77e614436f38"
	mirB2TelimTorOracle      = "b13a723d-80a9-4116-9857-baf3b373f87a"
	mirB2CommanderOracle     = "1053a778-da6a-4b8e-81fb-537a355d4bfe"
	mirB2ZhalfirinKnOracle   = "a534f06e-dae4-4d1a-96e9-7693bbb21e12"
	mirB2RaidersOracle       = "19edc98c-edc2-47bf-b2da-00f536d7c158"
	mirB2BayFalconOracle     = "709e99f1-288b-49e0-80a1-f817b457a2cf"
	mirB2JungleTrollOracle   = "f1bb17d7-5396-4c99-88bb-f7ab7d47b023"
	mirB2VolcanicOracle      = "994db177-03f5-43dd-bf7b-2994e8d430d3"
)

// mirB2Flankers are the batch's flanking cards.
var mirB2Flankers = map[string]string{
	mirB2BurningShieldOracle: "Burning Shield Askari",
	mirB2CadaverousOracle:    "Cadaverous Knight",
	mirB2FemerefKnightOracle: "Femeref Knight",
	mirB2CentaurOracle:       "Jolrael's Centaur",
	mirB2HerderOracle:        "Mtenda Herder",
	mirB2SearingSpearOracle:  "Searing Spear Askari",
	mirB2SidarJabariOracle:   "Sidar Jabari",
	mirB2TelimTorOracle:      "Telim'Tor",
	mirB2CommanderOracle:     "Zhalfirin Commander",
	mirB2ZhalfirinKnOracle:   "Zhalfirin Knight",
}

func TestMirB2EveryCardIsRegisteredFull(t *testing.T) {
	all := map[string]string{
		mirB2RaidersOracle:     "Merfolk Raiders",
		mirB2BayFalconOracle:   "Bay Falcon",
		mirB2JungleTrollOracle: "Jungle Troll",
		mirB2VolcanicOracle:    "Volcanic Dragon",
	}
	for oracle, name := range mirB2Flankers {
		all[oracle] = name
	}
	for oracle, name := range all {
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

// Every flanking card has the keyword from its catalog row alone, with
// no Keywords slice on the card, and the other printed keywords too.
func TestMirB2PrintedKeywords(t *testing.T) {
	type row struct {
		name, oracle string
		keywords     []string
	}
	rows := []row{
		{"Jolrael's Centaur", mirB2CentaurOracle, []string{"shroud"}},
		{"Merfolk Raiders", mirB2RaidersOracle, []string{"islandwalk", game.KeywordPhasing}},
		{"Bay Falcon", mirB2BayFalconOracle, []string{"flying", "vigilance"}},
		{"Volcanic Dragon", mirB2VolcanicOracle, []string{"flying", "haste"}},
	}
	for oracle, name := range mirB2Flankers {
		rows = append(rows, row{name, oracle, []string{game.KeywordFlanking}})
	}
	for _, tc := range rows {
		g := newCatalogGame(t)
		id := mirB1Push(g, g.Seats[0].ID, tc.name, "Creature", tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !hasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s does not have %s", tc.name, kw)
			}
		}
	}
}

// --- flanking in combat ----------------------------------------------------------

// The #5757 case with this batch's cards: a real Femeref Knight is
// blocked by a real Zhalfirin Knight and a Grizzly Bears. The Zhalfirin
// Knight has flanking and keeps its toughness (CR 702.25a); the Bears
// does not and gets -1/-1 until end of turn.
func TestMirB2FlankingBlockerKeepsItsToughnessAndOtherBlockerShrinks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := apaPush(g, me.ID, me.ID, flankingImport(t, mirB2FemerefKnightOracle, "Femeref Knight",
		"Creature — Human Knight", "{2}{W}", "2", "2", "Flanking"))
	knight := apaPush(g, opp.ID, opp.ID, flankingImport(t, mirB2ZhalfirinKnOracle, "Zhalfirin Knight",
		"Creature — Human Knight", "{2}{W}", "2", "2", "Flanking"))
	bears := apaPush(g, opp.ID, opp.ID, flankingBears(t))
	flankingBlocks(t, g, opp.ID, attacker, knight, bears)
	if p, tough := ptOf(t, g, knight); p != 2 || tough != 2 {
		t.Errorf("Zhalfirin Knight, which has flanking, is %d/%d, want 2/2", p, tough)
	}
	if p, tough := ptOf(t, g, bears); p != 1 || tough != 1 {
		t.Errorf("Grizzly Bears, without flanking, is %d/%d, want 1/1", p, tough)
	}
}

// A catalog flanking card with no Keywords slice at all is spared too:
// the declared PrintedKeywords is what the trigger reads.
func TestMirB2CatalogFlankerBlockingAFlankerIsSpared(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herder := apaPush(g, me.ID, me.ID, apaCreature("Mtenda Herder", mirB2HerderOracle, 1, 1))
	knight := apaPush(g, opp.ID, opp.ID, apaCreature("Cadaverous Knight", mirB2CadaverousOracle, 2, 2))
	flankingBlocks(t, g, opp.ID, herder, knight)
	if got := apaLive(g, knight).CurrentToughness(); got != 2 {
		t.Errorf("Cadaverous Knight, which has flanking, has toughness %d, want 2", got)
	}
}

// A 1/1 without flanking that blocks Mtenda Herder gets -1/-1 before
// damage and dies as a state-based action.
func TestMirB2MtendaHerderKillsAOneToughnessBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herder := apaPush(g, me.ID, me.ID, flankingImport(t, mirB2HerderOracle, "Mtenda Herder",
		"Creature — Human Scout", "{W}", "1", "1", "Flanking"))
	squire := apaPush(g, opp.ID, opp.ID, flankingImport(t, "", "Squire", "Creature — Human Soldier", "{1}{W}", "1", "1"))
	flankingBlocks(t, g, opp.ID, herder, squire)
	if g.Battlefield.Contains(squire) {
		t.Error("a 1/1 without flanking survived blocking Mtenda Herder")
	}
	if !opp.Graveyard.Contains(squire) {
		t.Error("the blocker is not in its owner's graveyard")
	}
}

// --- Telim'Tor -------------------------------------------------------------------

func TestMirB2TelimTorPumpsAttackingCreaturesWithFlanking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tor := apaPush(g, me.ID, me.ID, apaCreature("Telim'Tor", mirB2TelimTorOracle, 2, 2))
	herder := apaPush(g, me.ID, me.ID, apaCreature("Mtenda Herder", mirB2HerderOracle, 1, 1))
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	home := apaPush(g, me.ID, me.ID, apaCreature("Femeref Knight", mirB2FemerefKnightOracle, 2, 2))
	declareAttack(t, g, opp.ID, tor, herder, bear)
	passPriorityAroundTable(t, g)
	cmrB1WantPT(t, g, tor, "Telim'Tor", 3, 3)
	cmrB1WantPT(t, g, herder, "an attacking creature with flanking", 2, 2)
	cmrB1WantPT(t, g, bear, "an attacker without flanking", 2, 2)
	cmrB1WantPT(t, g, home, "a creature with flanking that did not attack", 2, 2)
}

// --- Sidar Jabari ----------------------------------------------------------------

func TestMirB2SidarJabariTapsADefendersCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sidar := apaPush(g, me.ID, me.ID, apaCreature("Sidar Jabari", mirB2SidarJabariOracle, 2, 2))
	victim := brCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bystander := brCreature(g, g.Seats[2].ID, "Other Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, sidar)
	pickIfAsked(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, victim) {
		t.Error("Sidar Jabari's attack did not tap the defending player's creature")
	}
	if tappedNow(g, bystander) {
		t.Error("a creature of a player who is not defending was tapped")
	}
}

// --- the keyword grants ----------------------------------------------------------

func TestMirB2KnightsGainTheirKeywordUntilEndOfTurn(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, mana, keyword string
	}{
		{"Burning Shield Askari", mirB2BurningShieldOracle, "{R}{R}", "first strike"},
		{"Zhalfirin Knight", mirB2ZhalfirinKnOracle, "{W}{W}", "first strike"},
		{"Femeref Knight", mirB2FemerefKnightOracle, "{W}", "vigilance"},
		{"Searing Spear Askari", mirB2SearingSpearOracle, "{C}{R}", "menace"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := mirB1Push(g, me.ID, tc.name, "Creature — Human Knight", tc.oracle, 2, 2)
			if hasKeywordOnBattlefield(t, g, id, tc.keyword) {
				t.Fatalf("%s has %s before the ability", tc.name, tc.keyword)
			}
			dcMana(t, g, me.ID, tc.mana)
			if err := mirB1Activate(g, me.ID, id, 0); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			if !hasKeywordOnBattlefield(t, g, id, tc.keyword) {
				t.Errorf("%s did not gain %s", tc.name, tc.keyword)
			}
			if !hasKeywordOnBattlefield(t, g, id, game.KeywordFlanking) {
				t.Errorf("%s lost flanking", tc.name)
			}
		})
	}
}

// Searing Spear Askari with menace can't be blocked by one creature
// (CR 702.111b).
func TestMirB2SearingSpearAskariWithMenaceRefusesALoneBlocker(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	askari := apaPush(g, me.ID, me.ID, apaCreature("Searing Spear Askari", mirB2SearingSpearOracle, 2, 2))
	bear := brCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	dcMana(t, g, me.ID, "{C}{R}")
	if err := mirB1Activate(g, me.ID, askari, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(askari, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, askari); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("a lone block on Searing Spear Askari with menace = %v, want a refusal", err)
	}
}

// --- Zhalfirin Commander ---------------------------------------------------------

func TestMirB2ZhalfirinCommanderPumpsOnlyAKnight(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	commander := mirB1Push(g, me.ID, "Zhalfirin Commander", "Creature — Human Knight", mirB2CommanderOracle, 2, 2)
	knight := mirB1Push(g, opp.ID, "Their Knight", "Creature — Human Knight", "", 2, 2)
	bear := mirB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	dcMana(t, g, me.ID, "{C}{W}{W}")
	if err := mirB1Activate(g, me.ID, commander, 0, cardTarget(bear)...); err == nil {
		t.Fatal("a creature that isn't a Knight was accepted as the target")
	}
	if err := mirB1Activate(g, me.ID, commander, 0, cardTarget(knight)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, knight); p != 3 || tough != 3 {
		t.Errorf("the target Knight is %d/%d, want 3/3", p, tough)
	}
	if p, tough := ptOf(t, g, commander); p != 2 || tough != 2 {
		t.Errorf("Zhalfirin Commander is %d/%d, want 2/2", p, tough)
	}
}

// --- Jolrael's Centaur -----------------------------------------------------------

// Shroud stops its controller's own ability too (CR 702.18a): a Harmattan
// Efreet beside it can't give it flying.
func TestMirB2JolraelsCentaurCantBeTargeted(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	centaur := mirB1Push(g, me.ID, "Jolrael's Centaur", "Creature — Centaur Archer", mirB2CentaurOracle, 2, 2)
	efreet := mirB1Push(g, me.ID, "Harmattan Efreet", "Creature — Efreet", mirB1HarmattanOracle, 2, 2)
	dcMana(t, g, me.ID, "{C}{U}{U}")
	if err := mirB1Activate(g, me.ID, efreet, 0, cardTarget(centaur)...); err == nil {
		t.Error("Jolrael's Centaur, which has shroud, was accepted as a target")
	}
}

// --- regeneration ----------------------------------------------------------------

func TestMirB2RegeneratorsSurviveOneDestruction(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, mana string
		row                int
	}{
		{"Cadaverous Knight", mirB2CadaverousOracle, "{C}{B}{B}", 0},
		{"Jungle Troll", mirB2JungleTrollOracle, "{R}", 0},
		{"Jungle Troll", mirB2JungleTrollOracle, "{G}", 1},
	} {
		t.Run(tc.name+" "+tc.mana, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := mirB1Push(g, me.ID, tc.name, "Creature", tc.oracle, 2, 1)
			dcMana(t, g, me.ID, tc.mana)
			if err := mirB1Activate(g, me.ID, id, tc.row); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
			if !g.Battlefield.Contains(id) {
				t.Fatalf("the regenerated %s was destroyed", tc.name)
			}
			if !findCardAnywhere(t, g, id).Tapped {
				t.Errorf("regeneration did not tap %s", tc.name)
			}
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
			if g.Battlefield.Contains(id) {
				t.Errorf("one regeneration shield saved %s twice", tc.name)
			}
		})
	}
}

// --- Merfolk Raiders -------------------------------------------------------------

func TestMirB2MerfolkRaidersPhaseOutInTheirControllersUntapStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := mirB1Push(g, me.ID, "Merfolk Raiders", "Creature — Merfolk Soldier", mirB2RaidersOracle, 2, 3)
	walkToUpkeepOfSeat0(t, g)
	if !g.IsPhasedOutForEffect(id) {
		t.Error("Merfolk Raiders did not phase out in its controller's untap step")
	}
}
