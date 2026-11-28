package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b716_cards_test.go — Urza's Legacy, activated abilities and life
// gain and loss, batch 716 (S82, #5623). Each test plays the real catalog
// entry through the engine.

const (
	oldB716SpongeOracle     = "000e5d65-96c3-498b-bd01-72b1a1991850"
	oldB716CrabOracle       = "24982bfe-5276-44a8-99a5-43b497ebf51f"
	oldB716DrakeOracle      = "4f5eaf5a-dd52-4d48-91b5-d48763d13159"
	oldB716ElfOracle        = "73d0b29a-5b7a-4835-9bd4-5d00e685b851"
	oldB716ToolboxOracle    = "a08475a6-1fac-4a3f-b343-aed16a61bc16"
	oldB716SubversionOracle = "0772c5d2-9189-47e2-b0ee-13babd94ea93"
	oldB716SkirgeOracle     = "42bbef66-967b-4922-8b2e-e57d1b8244bb"
	oldB716DelusionsOracle  = "a1fb6dfc-1dab-4337-8311-b4ab2ef336c0"
	oldB716ChanceOracle     = "aa19bbf0-7de3-4ec3-a98b-2ac7533c6b0b"
	oldB716ScrapheapOracle  = "dfed42b8-b35e-4c70-9e75-25a4da158e76"
	oldB716SnareOracle      = "7d011e22-335b-465f-9111-e95a72ebffe9"
)

func TestOldB716EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB716SpongeOracle: "Walking Sponge", oldB716CrabOracle: "King Crab",
		oldB716DrakeOracle: "Vigilant Drake", oldB716ElfOracle: "Weatherseed Elf",
		oldB716ToolboxOracle: "Jhoira's Toolbox", oldB716SubversionOracle: "Subversion",
		oldB716SkirgeOracle: "Tethered Skirge", oldB716DelusionsOracle: "Delusions of Mediocrity",
		oldB716ChanceOracle: "Second Chance", oldB716ScrapheapOracle: "Scrapheap",
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

func oldB716Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB716Activate(t *testing.T, g *game.Game, p *game.Player, src uuid.UUID, mana string, targets []game.TargetRef) error {
	t.Helper()
	effectstest.FloatForTest(g, p, mana)
	return g.ActivateCatalogAbility(p.ID, src, 0, game.ActivateAbilityParams{Strict: true, Targets: targets})
}

func TestOldB716WalkingSpongeRemovesTheKeywordTheControllerPicks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sponge := effectstest.PushCatalogPermanent(g, me.ID, "Walking Sponge", "Creature — Sponge", oldB716SpongeOracle, false)
	foe := effectstest.SfPush(g, opp.ID, "Foe", "Creature — Test", 2, []string{"R"}, "flying", "first strike", "trample")
	if err := oldB716Activate(t, g, me, sponge, "", oldB716Target(foe)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, foe, "first strike") {
		t.Error("the chosen keyword (first strike) was not removed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, foe, "flying") || !effectstest.HasEffectiveKeyword(t, g, foe, "trample") {
		t.Error("a keyword that was not chosen was removed")
	}
}

func TestOldB716KingCrabTucksOnlyAGreenCreatureOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	crab := effectstest.PushCatalogPermanent(g, me.ID, "King Crab", "Creature — Crab", oldB716CrabOracle, false)
	red := effectstest.SfPush(g, opp.ID, "Red", "Creature — Test", 2, []string{"R"})
	green := effectstest.SfPush(g, opp.ID, "Green", "Creature — Test", 2, []string{"G"})
	if err := oldB716Activate(t, g, me, crab, "CU", oldB716Target(red)); err == nil {
		t.Fatal("King Crab targeted a non-green creature")
	}
	if err := oldB716Activate(t, g, me, crab, "CU", oldB716Target(green)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the green creature is still on the battlefield")
	}
	if got := effectstest.LibraryTopIDs(opp, 1); len(got) != 1 || got[0] != green {
		t.Error("the green creature is not on top of its owner's library")
	}
	if !effectstest.TappedOnBattlefield(t, g, crab) {
		t.Error("King Crab did not tap")
	}
}

func TestOldB716VigilantDrakeUntapsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	drake := effectstest.PushCatalogPermanent(g, me.ID, "Vigilant Drake", "Creature — Drake", oldB716DrakeOracle, false)
	if !effectstest.HasEffectiveKeyword(t, g, drake, "flying") {
		t.Error("no flying")
	}
	effectstest.B02bSetTapped(g, drake, true)
	if err := oldB716Activate(t, g, me, drake, "CCU", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedOnBattlefield(t, g, drake) {
		t.Error("the Drake did not untap")
	}
}

func TestOldB716WeatherseedElfGrantsForestwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	elf := effectstest.PushCatalogPermanent(g, me.ID, "Weatherseed Elf", "Creature — Elf", oldB716ElfOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := oldB716Activate(t, g, me, elf, "", oldB716Target(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "forestwalk") {
		t.Error("the target did not gain forestwalk")
	}
	if !effectstest.TappedOnBattlefield(t, g, elf) {
		t.Error("the Elf did not tap")
	}
}

func TestOldB716JhoirasToolboxRegeneratesOnlyAnArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	box := effectstest.PushCatalogPermanent(g, me.ID, "Jhoira's Toolbox", "Artifact Creature — Insect", oldB716ToolboxOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	if err := oldB716Activate(t, g, me, box, "CC", oldB716Target(bear)); err == nil {
		t.Fatal("the Toolbox targeted a nonartifact creature")
	}
	if err := oldB716Activate(t, g, me, box, "CC", oldB716Target(golem)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.RegenShieldsOn(t, g, golem); got != 1 {
		t.Errorf("%d regeneration shields on the Golem, want 1", got)
	}
}

func TestOldB716SubversionDrainsEachOpponentAndGainsTheTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Subversion", "Enchantment", oldB716SubversionOracle, false)
	lives := effectstest.B27Lives(g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.B27Lives(g)
	for i := range got {
		want := lives[i] - 1
		if i == 0 {
			want = lives[i] + len(g.Seats) - 1
		}
		if got[i] != want {
			t.Errorf("seat %d life %d, want %d", i, got[i], want)
		}
	}
}

func TestOldB716TetheredSkirgeCostsItsControllerALifeWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	skirge := effectstest.PushCatalogPermanent(g, opp.ID, "Tethered Skirge", "Creature — Phyrexian Imp", oldB716SkirgeOracle, false)
	if !effectstest.HasEffectiveKeyword(t, g, skirge, "flying") {
		t.Error("no flying")
	}
	lives := effectstest.B27Lives(g)
	effectstest.CastCatalogSpell(t, g, "Snare the Skies", "Instant", oldB716SnareOracle, oldB716Target(skirge))
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.B27Lives(g)
	if got[1] != lives[1]-1 {
		t.Errorf("the Skirge's controller has %d life, want %d", got[1], lives[1]-1)
	}
	if got[0] != lives[0] {
		t.Errorf("the targeting player's life changed: %d -> %d", lives[0], got[0])
	}
	_ = me
}

func TestOldB716DelusionsOfMediocrityGainsOnEnteringAndLosesOnLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life
	id := effectstest.CastHoldCreature(t, g, "Delusions of Mediocrity", "Enchantment", oldB716DelusionsOracle, 0, 0)
	if me.Life != before+10 {
		t.Fatalf("life %d after entering, want %d", me.Life, before+10)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("life %d after leaving, want %d", me.Life, before)
	}
}

func TestOldB716SecondChanceSacrificesAndTakesAnExtraTurnAtFiveOrLess(t *testing.T) {
	for _, tc := range []struct {
		life int
		fire bool
	}{{6, false}, {5, true}, {1, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		chance := effectstest.PushCatalogPermanent(g, me.ID, "Second Chance", "Enchantment", oldB716ChanceOracle, false)
		me.Life = tc.life
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		stillThere := g.Battlefield.Contains(chance)
		turns := len(g.ExtraTurnsQueued())
		if tc.fire && (stillThere || turns != 1) {
			t.Errorf("life %d: on battlefield %v, extra turns %d; want sacrificed and one extra turn", tc.life, stillThere, turns)
		}
		if !tc.fire && (!stillThere || turns != 0) {
			t.Errorf("life %d: on battlefield %v, extra turns %d; want untouched", tc.life, stillThere, turns)
		}
	}
}

func TestOldB716ScrapheapGainsOneForEachOfYourArtifactsAndEnchantmentsThatDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Scrapheap", "Artifact", oldB716ScrapheapOracle, false)
	mineArtifact := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	mineEnchant := effectstest.B12Permanent(g, me.ID, "My Aura", "Enchantment")
	mineCreature := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirArtifact := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	before := me.Life
	bf := game.ZoneRef{Kind: game.ZoneBattlefield}
	for _, mv := range []struct {
		id    uuid.UUID
		owner *game.Player
	}{{mineArtifact, me}, {mineEnchant, me}, {mineCreature, me}, {theirArtifact, opp}} {
		if err := g.MoveCardByID(bf, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: mv.owner.ID}, mv.id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d, want %d (an artifact and an enchantment of mine, not a creature or their artifact)", me.Life, before+2)
	}
}
