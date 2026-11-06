package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_dg_cards_test.go — Foundations, D–G (#4465): the cards from that
// stretch of the alphabet that needed no new seam. Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	fdnDGDauntlessVeteranOracle  = "97b2fb8b-0c69-4ed2-b2d0-8a780f9b5b97"
	fdnDGDawnwingMarshalOracle   = "6640c234-8863-4670-a61f-ba6f59b60d69"
	fdnDGDeathmarkOracle         = "cf09b0af-3cf1-4486-8f65-8cfd2410314a"
	fdnDGDiregrafGhoulOracle     = "6048fc70-0dcc-4b54-977d-16e240225f82"
	fdnDGDiveDownOracle          = "2f990b54-fbf3-4949-85bb-9ba39710e72a"
	fdnDGDragonTrainerOracle     = "328a0691-9453-4f2d-a9c3-5c69da252d20"
	fdnDGDruidOfTheCowlOracle    = "b1793c3b-25d6-4fae-a99d-cfdd2210ca67"
	fdnDGElementalistAdeptOracle = "cf402270-ebc6-481b-b478-0e3acab7002a"
	fdnDGExclusionMageOracle     = "fd28fe55-36cd-4242-8dd4-edb58fbb9895"
	fdnDGFireElementalOracle     = "3912d21e-1ebc-4a81-9dc9-f404248d564a"
	fdnDGGleamingBarrierOracle   = "1eb9e401-8975-447b-839d-f7cd23897465"
	fdnDGGuardedHeirOracle       = "63e79a6f-ed98-480a-a740-d44e111046f7"
)

func TestFdnDGEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnDGDauntlessVeteranOracle:  "Dauntless Veteran",
		fdnDGDawnwingMarshalOracle:   "Dawnwing Marshal",
		fdnDGDeathmarkOracle:         "Deathmark",
		fdnDGDiregrafGhoulOracle:     "Diregraf Ghoul",
		fdnDGDiveDownOracle:          "Dive Down",
		fdnDGDragonTrainerOracle:     "Dragon Trainer",
		fdnDGDruidOfTheCowlOracle:    "Druid of the Cowl",
		fdnDGElementalistAdeptOracle: "Elementalist Adept",
		fdnDGExclusionMageOracle:     "Exclusion Mage",
		fdnDGFireElementalOracle:     "Fire Elemental",
		fdnDGGleamingBarrierOracle:   "Gleaming Barrier",
		fdnDGGuardedHeirOracle:       "Guarded Heir",
	} {
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

// fdnDGTokens returns the P/T and colours of every token named `name`
// that `controller` controls.
func fdnDGTokens(g *game.Game, controller uuid.UUID, name string) (sizes [][2]int, colors [][]string) {
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == name && c.IsToken() {
				eff := c.Effective()
				sizes = append(sizes, [2]int{eff.Power, eff.Toughness})
				colors = append(colors, append([]string(nil), eff.Colors...))
			}
		}
	})
	return sizes, colors
}

// --- Dauntless Veteran --------------------------------------------------------

func TestFdnDGDauntlessVeteranPumpsYourTeamWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vet := b12Push(g, me.ID, "Dauntless Veteran", "Creature — Human Soldier", fdnDGDauntlessVeteranOracle, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(vet, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID][2]int{vet: {3, 3}, mine: {3, 3}, theirs: {2, 2}} {
		if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != want[0] || tough != want[1] {
			t.Errorf("%s is %d/%d after the attack trigger, want %d/%d", id, p, tough, want[0], want[1])
		}
	}
}

// --- Dawnwing Marshal ---------------------------------------------------------

func TestFdnDGDawnwingMarshalFliesAndPumpsYourTeamForFiveMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marshal := b12Push(g, me.ID, "Dawnwing Marshal", "Creature — Cat Soldier", fdnDGDawnwingMarshalOracle, 2, 2)
	assertKeywords(t, g, marshal, "flying")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	b06AddMana(me, "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, marshal, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without white mana")
	}
	b06AddMana(me, "W")
	b16Activate(t, g, me.ID, marshal, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{marshal: 3, mine: 3, theirs: 2} {
		if p := effectivePower(t, g, id); p != want {
			t.Errorf("%s has power %d after the pump, want %d", id, p, want)
		}
	}
}

// --- Deathmark ----------------------------------------------------------------

func TestFdnDGDeathmarkDestroysAGreenOrWhiteCreature(t *testing.T) {
	for _, color := range []string{"G", "W"} {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		target := b27Push(g, opp.ID, "Their Creature", "Creature — Beast", "", "{1}", 2, 2, color)
		castCatalogSpell(t, g, "Deathmark", "Sorcery", fdnDGDeathmarkOracle, cardRefs(target))
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(target) {
			t.Errorf("a %s creature survived Deathmark", color)
		}
	}
}

func TestFdnDGDeathmarkCannotTargetACreatureOfAnotherColor(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	black := b27Push(g, opp.ID, "Their Zombie", "Creature — Zombie", "", "{B}", 2, 2, "B")
	if err := castCatalogSpellErr(t, g, "Deathmark", "Sorcery", fdnDGDeathmarkOracle, cardRefs(black)); err == nil {
		t.Fatal("Deathmark was cast at a black creature")
	}
	if !g.Battlefield.Contains(black) {
		t.Error("the black creature left the battlefield")
	}
}

// --- Diregraf Ghoul -----------------------------------------------------------

func TestFdnDGDiregrafGhoulEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	ghoul := castHoldCreature(t, g, "Diregraf Ghoul", "Creature — Zombie", fdnDGDiregrafGhoulOracle, 2, 2)
	if !b27Tapped(t, g, ghoul) {
		t.Error("Diregraf Ghoul entered untapped")
	}
}

// --- Dive Down ----------------------------------------------------------------

func TestFdnDGDiveDownGivesToughnessAndHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Dive Down", "Instant", fdnDGDiveDownOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2 || tough != 5 {
		t.Errorf("the target is %d/%d, want 2/5", p, tough)
	}
	if !effectiveAbilitiesContain(t, g, bear, "hexproof") {
		t.Error("the target did not gain hexproof")
	}
}

func TestFdnDGDiveDownCannotTargetAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Dive Down", "Instant", fdnDGDiveDownOracle, cardRefs(theirs)); err == nil {
		t.Fatal("Dive Down was cast at an opponent's creature")
	}
}

// --- Dragon Trainer -----------------------------------------------------------

func TestFdnDGDragonTrainerMakesAFourFourFlyingDragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castHoldCreature(t, g, "Dragon Trainer", "Creature — Human", fdnDGDragonTrainerOracle, 1, 1)
	sizes, colors := fdnDGTokens(g, me.ID, "Dragon")
	if len(sizes) != 1 {
		t.Fatalf("%d Dragon tokens after the Trainer entered, want 1", len(sizes))
	}
	if sizes[0] != [2]int{4, 4} || len(colors[0]) != 1 || colors[0][0] != "R" {
		t.Errorf("the Dragon is %d/%d %v, want a 4/4 red token", sizes[0][0], sizes[0][1], colors[0])
	}
	var dragon uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Dragon" {
				dragon = c.InstanceID
			}
		}
	})
	assertKeywords(t, g, dragon, "flying")
}

// --- Druid of the Cowl --------------------------------------------------------

func TestFdnDGDruidOfTheCowlTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	druid := pushCatalogPermanent(g, me.ID, "Druid of the Cowl", "Creature — Elf Druid", fdnDGDruidOfTheCowlOracle, false)
	if err := g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "G" {
		t.Errorf("mana pool %+v, want 1xG", me.ManaPool)
	}
}

func TestFdnDGDruidOfTheCowlCannotTapWhileSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	druid := pushCatalogPermanent(g, me.ID, "Druid of the Cowl", "Creature — Elf Druid", fdnDGDruidOfTheCowlOracle, true)
	if err := g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{}); err != game.ErrSummoningSick {
		t.Fatalf("summoning-sick Druid: got %v, want ErrSummoningSick", err)
	}
}

// --- Elementalist Adept -------------------------------------------------------

func TestFdnDGElementalistAdeptHasFlashAndProwess(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	adept := b27Push(g, me.ID, "Elementalist Adept", "Creature — Human Wizard", fdnDGElementalistAdeptOracle, "{1}{U}", 2, 1, "U")
	assertKeywords(t, g, adept, "flash", "prowess")
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, adept), effectiveToughness(t, g, adept); p != 3 || tough != 2 {
		t.Errorf("the Adept is %d/%d after a noncreature spell, want 3/2", p, tough)
	}
}

// --- Exclusion Mage -----------------------------------------------------------

func TestFdnDGExclusionMageBouncesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mage := castHoldCreature(t, g, "Exclusion Mage", "Creature — Human Wizard", fdnDGExclusionMageOracle, 2, 2)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{mine, mage} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as a target", bad)
		}
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the opponent's creature was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(mage) {
		t.Error("your creature or the Mage left the battlefield")
	}
}

// --- Gleaming Barrier ---------------------------------------------------------

func TestFdnDGGleamingBarrierMakesATreasureWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	barrier := b27Push(g, me.ID, "Gleaming Barrier", "Artifact Creature — Wall", fdnDGGleamingBarrierOracle, "{2}", 0, 4)
	assertKeywords(t, g, barrier, "defender")
	before := tmtB1Count(g, me.ID, "Treasure")
	b27Kill(g, barrier)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Treasure"); got != before+1 {
		t.Errorf("%d Treasure tokens after the Barrier died, want %d", got, before+1)
	}
}

// --- Guarded Heir -------------------------------------------------------------

func TestFdnDGGuardedHeirMakesTwoWhiteKnights(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	heir := castHoldCreature(t, g, "Guarded Heir", "Creature — Human Noble", fdnDGGuardedHeirOracle, 1, 1)
	assertKeywords(t, g, heir, "lifelink")
	sizes, colors := fdnDGTokens(g, me.ID, "Knight")
	if len(sizes) != 2 {
		t.Fatalf("%d Knight tokens after the Heir entered, want 2", len(sizes))
	}
	for i := range sizes {
		if sizes[i] != [2]int{3, 3} || len(colors[i]) != 1 || colors[i][0] != "W" {
			t.Errorf("a Knight is %d/%d %v, want a 3/3 white token", sizes[i][0], sizes[i][1], colors[i])
		}
	}
}
