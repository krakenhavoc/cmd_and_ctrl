package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b177_test.go — GAP batch 177 (#5948): Jumpstart 2022 sorceries.
// Every test plays the real catalog entry through the engine.

const (
	gapB177ActOnImpulse      = "19ad338a-c89f-406d-ad1a-222c22044e1b"
	gapB177Aftershock        = "e6bef0c4-3dfc-49a8-8fec-22b2570b704b"
	gapB177ConeOfFlame       = "62879f3d-ed6d-4004-8196-f3fd7f614695"
	gapB177DestructiveTamper = "e6ae536a-95c0-495d-98b8-200b3564dd79"
	gapB177Firecannon        = "97bff952-3bc8-41d0-b03f-5c76f3c7e0a3"
	gapB177GoblinGrenade     = "20aecb7b-7e37-4efe-ac27-f9f9042203f4"
	gapB177SparkOfCreativity = "22504fd0-0353-4ef0-baf4-4b7be28d1a66"
	gapB177AnchorToAether    = "3e00326d-2bf4-4731-a6cd-1c015f94f553"
)

func TestGapB177EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		gapB177ActOnImpulse, gapB177Aftershock, gapB177ConeOfFlame, gapB177DestructiveTamper,
		gapB177Firecannon, gapB177GoblinGrenade, gapB177SparkOfCreativity, gapB177AnchorToAether,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestGapB177ActOnImpulseExilesThreeAndLetsYouPlayThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	top := effectstest.SeedLibrary(me, "First", "Second", "Third", "Fourth")
	effectstest.CastCatalogSpell(t, g, "Act on Impulse", "Sorcery", gapB177ActOnImpulse, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		if !effectstest.B12ExiledPlayableBy(t, g, top[i], me.ID) {
			t.Errorf("card %d should be exiled and playable", i+1)
		}
	}
	if effectstest.B12ZoneOf(g, top[3]) != game.ZoneLibrary {
		t.Error("the fourth card must stay in the library")
	}
	effectstest.EndTurn(t, g)
	if effectstest.B12ExiledPlayableBy(t, g, top[0], me.ID) {
		t.Error("the permission must end with the turn")
	}
}

func TestGapB177GoblinGrenadeSacrificesAGoblinForFiveDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	goblin := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	before := opp.Life
	effectstest.P1g2Cast(t, g, "Goblin Grenade", "Sorcery", gapB177GoblinGrenade, "", game.CastSpellParams{
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{goblin},
	})
	effectstest.P1WantZone(t, g, goblin, game.ZoneGraveyard, "the sacrificed Goblin")
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-5 {
		t.Errorf("opponent lost %d, want 5", before-opp.Life)
	}
}

func TestGapB177GoblinGrenadeRefusesWithoutAGoblin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Goblin Grenade", TypeLine: "Sorcery", OracleID: gapB177GoblinGrenade,
		Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{bear}, Strict: true,
	})
	if err == nil {
		t.Fatal("Goblin Grenade was cast with a non-Goblin sacrifice")
	}
}

func TestGapB177FirecannonBlastIsThreeWithoutRaidAndSixWithIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	victim := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.CastCatalogSpell(t, g, "Firecannon Blast", "Sorcery", gapB177Firecannon, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, victim); got != 3 {
		t.Errorf("without raid: %d damage, want 3", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := effectstest.Top100ActiveSeat(g2), effectstest.Top100Opponent(g2)
	attacker := effectstest.B12Creature(g2, me2.ID, "Attacker", "Creature — Human", 1, 1)
	victim2 := effectstest.B12Creature(g2, opp2.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.BrAttack(t, g2, attacker)
	effectstest.CastCatalogSpell(t, g2, "Firecannon Blast", "Sorcery", gapB177Firecannon, effectstest.CardTarget(victim2))
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.DamageMarkedOn(g2, victim2); got != 6 {
		t.Errorf("with raid: %d damage, want 6", got)
	}
}

func TestGapB177AnchorToTheAetherTucksThenScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.SeedLibrary(opp, "Theirs")
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Anchor to the Aether", "Sorcery", gapB177AnchorToAether, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, victim) != game.ZoneLibrary {
		t.Fatalf("the creature is in %v, want its owner's library", effectstest.B12ZoneOf(g, victim))
	}
	if top, err := opp.Library.Top(); err != nil || top.InstanceID != victim {
		t.Error("the creature is not on top of its owner's library")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

func TestGapB177DestructiveTamperingDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.P1g2Cast(t, g, "Destructive Tampering", "Sorcery", gapB177DestructiveTamper, "", game.CastSpellParams{
		Modes: []int{0}, Targets: effectstest.CardTarget(rock),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestGapB177DestructiveTamperingGroundsNonFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	walker := effectstest.B12Creature(g, me.ID, "Walker", "Creature — Human", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	effectstest.CastCatalogSpellWithModes(t, g, "Destructive Tampering", "Sorcery", gapB177DestructiveTamper, []int{1})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BrAttack(t, g, walker)
	if err := g.DeclareBlocker(bear, walker); err == nil {
		t.Error("a creature without flying blocked")
	}
	if err := g.DeclareBlocker(flyer, walker); err != nil {
		t.Errorf("a creature with flying can't block: %v", err)
	}
}

func TestGapB177AftershockDestroysAndHurtsYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Aftershock", "Sorcery", gapB177Aftershock, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
	if me.Life != before-3 {
		t.Errorf("you lost %d life, want 3", before-me.Life)
	}
}

func TestGapB177AftershockRefusesAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	ench := effectstest.B12Permanent(g, opp.ID, "Aura", "Enchantment")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Aftershock", "Sorcery", gapB177Aftershock, effectstest.CardTarget(ench)); err == nil {
		t.Error("Aftershock targeted an enchantment")
	}
}

func TestGapB177ConeOfFlameDealsOneTwoAndThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 9)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 9)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Wall", 0, 9)
	effectstest.CastCatalogSpell(t, g, "Cone of Flame", "Sorcery", gapB177ConeOfFlame, effectstest.CardTargets(a, b, c))
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{a: 1, b: 2, c: 3} {
		if got := effectstest.DamageMarkedOn(g, id); got != want {
			t.Errorf("target took %d, want %d", got, want)
		}
	}
}

func TestGapB177ConeOfFlameNeedsDifferentTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 9)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Cone of Flame", "Sorcery", gapB177ConeOfFlame, effectstest.CardTargets(a, a, a)); err == nil {
		t.Error("Cone of Flame accepted one object in all three slots")
	}
}

func TestGapB177SparkOfCreativityDamageEqualToManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	victim := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	top := effectstest.SeedLibraryTop(me, "Big Spell", "Sorcery", "{2}{R}")
	effectstest.CastCatalogSpell(t, g, "Spark of Creativity", "Sorcery", gapB177SparkOfCreativity, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, victim); got != 3 {
		t.Errorf("the creature took %d, want 3 (the exiled card's mana value)", got)
	}
	if effectstest.B12ZoneOf(g, top) != game.ZoneExile {
		t.Error("the exiled card should stay in exile")
	}
	if effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("choosing the damage must not also grant the play")
	}
}

func TestGapB177SparkOfCreativityDecliningLetsYouPlayTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	victim := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	top := effectstest.SeedLibraryTop(me, "Big Spell", "Sorcery", "{2}{R}")
	effectstest.CastCatalogSpell(t, g, "Spark of Creativity", "Sorcery", gapB177SparkOfCreativity, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, victim); got != 0 {
		t.Errorf("the creature took %d, want 0", got)
	}
	if !effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the exiled card should be playable this turn")
	}
	effectstest.EndTurn(t, g)
	if effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the permission must end with the turn")
	}
}
