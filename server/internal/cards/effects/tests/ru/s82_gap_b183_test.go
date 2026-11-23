package testsru

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b183_test.go — GAP batch 183 (#5948, Starter Commander Decks):
// every card driven through the engine.

const (
	gap183KazanduOracle    = "03332772-18b0-446a-a7e2-a4bdf38c4f0c"
	gap183ShivanOracle     = "9b64c0f4-b917-4a0d-b033-30ab71aab807"
	gap183ElfhameOracle    = "cade8b94-2998-4d23-87bb-9fbdddd19dea"
	gap183CoastalOracle    = "f0367926-4380-4d22-8cb8-46c2076f102a"
	gap183SejiriOracle     = "73b3e242-075d-4c4d-9b09-6fef1633c348"
	gap183LilianaOracle    = "4f66489c-5a19-40ad-9126-461fa8231f1b"
	gap183DredgeOracle     = "10e286b0-7245-4c30-a7a9-6d4a9c49b0a7"
	gap183GodEternalOracle = "67ee7c8b-afcc-4f09-a184-3db01f2ab175"
)

func TestS82Gap183CardsAreRegisteredFull(t *testing.T) {
	for _, id := range []string{gap183KazanduOracle, gap183ShivanOracle, gap183ElfhameOracle, gap183CoastalOracle,
		gap183SejiriOracle, gap183LilianaOracle, gap183DredgeOracle, gap183GodEternalOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- the lands -------------------------------------------------------------

func TestS82Gap183LandsEnterTappedAndTheRefugesGainALife(t *testing.T) {
	for _, c := range []struct {
		name, oracle string
		life         int
	}{
		{"Shivan Oasis", gap183ShivanOracle, 0},
		{"Elfhame Palace", gap183ElfhameOracle, 0},
		{"Coastal Tower", gap183CoastalOracle, 0},
		{"Kazandu Refuge", gap183KazanduOracle, 1},
		{"Sejiri Refuge", gap183SejiriOracle, 1},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		before := effectstest.LifeOf(g, me.ID)
		land := effectstest.PlayLandFromHand(t, g, c.name, c.oracle)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.LayeredCard(t, g, land).Tapped {
			t.Errorf("%s did not enter tapped", c.name)
		}
		if got := effectstest.LifeOf(g, me.ID); got != before+c.life {
			t.Errorf("%s: life %d -> %d, want %+d", c.name, before, got, c.life)
		}
		spec, _ := effects.Lookup(c.oracle)
		if len(spec.ManaAbilities) != 1 {
			t.Errorf("%s: %d mana abilities, want 1", c.name, len(spec.ManaAbilities))
		}
	}
}

// --- Dredge the Mire --------------------------------------------------------

func TestS82Gap183DredgeTheMireTakesEachOpponentsChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	aKeep := effectstest.PushToGraveyard(g, a, game.Card{Name: "A Keeper", TypeLine: "Creature — Elf", Power: 1, Toughness: 1})
	aTake := effectstest.PushToGraveyard(g, a, game.Card{Name: "A Taken", TypeLine: "Creature — Elf", Power: 2, Toughness: 2})
	bTake := effectstest.PushToGraveyard(g, b, game.Card{Name: "B Taken", TypeLine: "Creature — Goblin", Power: 3, Toughness: 3})
	effectstest.PushToGraveyard(g, b, game.Card{Name: "A Sorcery", TypeLine: "Sorcery"})
	effectstest.PushToGraveyard(g, c, game.Card{Name: "A Land", TypeLine: "Land"})
	mineStays := effectstest.PushToGraveyard(g, me, game.Card{Name: "Mine", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})

	effectstest.CastCatalogSpell(t, g, "Dredge the Mire", "Sorcery", gap183DredgeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestChooseCardsFor(g, c.ID); p != nil {
		t.Error("an opponent with no creature card in their graveyard was asked")
	}
	effectstest.AnswerChooseCards(t, g, a.ID, aTake)
	effectstest.AnswerChooseCards(t, g, b.ID, bTake)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, aTake); got.Controller != me.ID {
		t.Errorf("A's chosen creature is controlled by %s, want the caster", got.Controller)
	}
	if got := effectstest.LayeredCard(t, g, bTake); got.Controller != me.ID {
		t.Errorf("B's chosen creature is controlled by %s, want the caster", got.Controller)
	}
	if g.Battlefield.Contains(aKeep) {
		t.Error("the creature A did not choose came back too")
	}
	if g.Battlefield.Contains(mineStays) {
		t.Error("the caster's own graveyard creature came back")
	}
}

// --- Enter the God-Eternals -------------------------------------------------

func TestS82Gap183EnterTheGodEternalsDamagesDrainsMillsAndAmasses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Big", 2, 6)
	before, oppLib := effectstest.LifeOf(g, me.ID), opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Enter the God-Eternals", "Sorcery", gap183GodEternalOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear, Slot: 0}, {Kind: game.TargetPlayer, ID: opp.ID, Slot: 1}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != before+4 {
		t.Errorf("life %d -> %d, want +4 (the damage dealt)", before, got)
	}
	if got := opp.Library.Size(); got != oppLib-4 {
		t.Errorf("the targeted player's library %d -> %d, want -4", oppLib, got)
	}
	army := effectstest.ArmyOf(t, g, me.ID)
	if n := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); n != 4 {
		t.Errorf("the Army has %d +1/+1 counters, want 4", n)
	}
	if !effectstest.LayeredCard(t, g, army.InstanceID).HasSubtype("Zombie") {
		t.Error("the Army is not a Zombie")
	}
}

// --- Liliana, Untouched by Death --------------------------------------------

func TestS82Gap183LilianaPlusOneDrainsOnlyWhenAZombieIsMilled(t *testing.T) {
	for _, withZombie := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.ToMain(t, g)
		lili := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Liliana, Untouched by Death", "Legendary Planeswalker — Liliana", gap183LilianaOracle, 4)
		top := game.Card{Name: "Plain", TypeLine: "Creature — Bear", Power: 1, Toughness: 1}
		if withZombie {
			top = game.Card{Name: "Walker", TypeLine: "Creature — Zombie", Power: 1, Toughness: 1}
		}
		effectstest.GsSetLibrary(me, top, game.Card{Name: "B", TypeLine: "Land"}, game.Card{Name: "C", TypeLine: "Land"}, game.Card{Name: "D", TypeLine: "Land"})
		myLife, oppLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
		effectstest.B16Activate(t, g, me.ID, lili, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.LoyaltyCount(g, lili); n != 5 {
			t.Errorf("loyalty %d, want 5", n)
		}
		wantMine, wantOpp := myLife, oppLife
		if withZombie {
			wantMine, wantOpp = myLife+2, oppLife-2
		}
		if got := effectstest.LifeOf(g, me.ID); got != wantMine {
			t.Errorf("zombie=%v: my life %d, want %d", withZombie, got, wantMine)
		}
		if got := effectstest.LifeOf(g, opp.ID); got != wantOpp {
			t.Errorf("zombie=%v: their life %d, want %d", withZombie, got, wantOpp)
		}
	}
}

func TestS82Gap183LilianaMinusTwoShrinksByYourZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	lili := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Liliana, Untouched by Death", "Legendary Planeswalker — Liliana", gap183LilianaOracle, 4)
	effectstest.B12Creature(g, me.ID, "Zed", "Creature — Zombie", 1, 1)
	effectstest.B12Creature(g, me.ID, "Zee", "Creature — Zombie Warrior", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 1, 1)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 4, 4)
	effectstest.B16Activate(t, g, me.ID, lili, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.LoyaltyCount(g, lili); n != 2 {
		t.Errorf("loyalty %d, want 2", n)
	}
	if p, tough := effectstest.EffectivePower(t, g, big), effectstest.EffectiveToughness(t, g, big); p != 2 || tough != 2 {
		t.Errorf("target is %d/%d, want 2/2 (-2/-2 for two Zombies of mine)", p, tough)
	}
}

func TestS82Gap183LilianaMinusThreeLetsYouCastZombiesFromTheGraveyardThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	lili := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Liliana, Untouched by Death", "Legendary Planeswalker — Liliana", gap183LilianaOracle, 4)
	zombie := effectstest.PushToGraveyard(g, me, game.Card{Name: "Walker", TypeLine: "Creature — Zombie", ManaCost: "{B}", Power: 1, Toughness: 1})
	bear := effectstest.PushToGraveyard(g, me, game.Card{Name: "Plain", TypeLine: "Creature — Bear", ManaCost: "{B}", Power: 1, Toughness: 1})
	effectstest.B06AddMana(me, "B")
	effectstest.B06AddMana(me, "B")
	if err := g.CastSpell(me.ID, zombie, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("a Zombie was cast from the graveyard before the −3")
	}
	effectstest.B16Activate(t, g, me.ID, lili, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Error("a non-Zombie was cast from the graveyard")
	}
	if err := g.CastSpell(me.ID, zombie, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("a Zombie could not be cast from the graveyard after the −3: %v", err)
	}
}
