package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b11_cards_test.go — Midnight Hunt batch 11 (#5497): the five
// Visions, Patrician Geist, Ominous Roost, Rootcoil Creeper, Seize the Storm
// and Croaking Counterpart, each driven through a real cast or activation.

const (
	visionsGloryOracle      = "ef47c447-bd7f-43c6-b4db-4a848c39820c"
	visionsDominanceOracle  = "14290772-ca17-4a55-86b2-7a638929b12d"
	visionsDreadOracle      = "0b12fee3-1ad1-4dc3-bd19-e9ccce5cb829"
	visionsRuinOracle       = "3570001c-844f-4319-9850-ab997d62a83b"
	visionsDuplicityOracle  = "ba85f0b0-b08c-4e62-a56e-b3ab7d699175"
	patricianGeistOracle    = "1f65c853-5262-4b75-b324-1fe0c165170d"
	ominousRoostOracle      = "745fc118-ffab-45c6-b51a-7bef2f6931e8"
	rootcoilCreeperOracle   = "101f06f8-7980-41d1-98fa-3cb959380737"
	seizeTheStormOracle     = "d3455e93-3a05-42db-945d-c4c85f4ac827"
	croakingCounterpartOrcl = "142b2f24-49f6-4ef8-9731-9ae2e513f796"
)

func TestS82MidB11AllTenAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{visionsGloryOracle, visionsDominanceOracle, visionsDreadOracle, visionsRuinOracle,
		visionsDuplicityOracle, patricianGeistOracle, ominousRoostOracle, rootcoilCreeperOracle,
		seizeTheStormOracle, croakingCounterpartOrcl} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// s82MidB11Commander puts a commander of the given mana cost on the battlefield.
func s82MidB11Commander(g *game.Game, owner uuid.UUID, manaCost string) uuid.UUID {
	id := uuid.New()
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: id, Name: "Big Commander", TypeLine: "Legendary Creature — Human", ManaCost: manaCost,
		Power: 3, Toughness: 3, Owner: owner, Controller: owner, IsCommander: true,
	})
	return id
}

func s82MidB11Graveyard(t *testing.T, g *game.Game, name, oracle string) uuid.UUID {
	t.Helper()
	return effectstest.SeedGraveyardCard(t, g, name, "Sorcery", oracle)
}

// The flashback discount is the greatest commander mana value, on the
// flashback cast only.
func TestS82MidB11VisionsFlashbackCostsXLessWhereXIsTheCommandersManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	s82MidB11Commander(g, me.ID, "{4}{R}{R}") // mana value 6
	id := s82MidB11Graveyard(t, g, "Visions of Glory", visionsGloryOracle)

	// {8}{W}{W} less 6 is {2}{W}{W}: three mana cannot pay it.
	effectstest.AddPool(t, g, me.ID, "{C}{W}{W}")
	if err := effectstest.S82MidB11Flashback(g, id, game.CastSpellParams{}); err == nil {
		t.Fatal("flashback paid with three mana, but the discounted cost is {2}{W}{W}")
	}
	effectstest.AddPool(t, g, me.ID, "{C}")
	if err := effectstest.S82MidB11Flashback(g, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("flashback with {2}{W}{W}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back Visions should be exiled")
	}
}

func TestS82MidB11VisionsFlashbackWithNoCommanderPaysFullPrice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := s82MidB11Graveyard(t, g, "Visions of Glory", visionsGloryOracle)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{C}{C}{C}{C}{W}{W}")
	if err := effectstest.S82MidB11Flashback(g, id, game.CastSpellParams{}); err == nil {
		t.Fatal("flashback paid with nine mana, but with no commander it costs {8}{W}{W}")
	}
	effectstest.AddPool(t, g, me.ID, "{C}")
	if err := effectstest.S82MidB11Flashback(g, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("flashback with ten mana: %v", err)
	}
}

func TestS82MidB11VisionsDiscountDoesNotReachAHandCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	s82MidB11Commander(g, me.ID, "{4}{R}{R}")
	hand := effectstest.B13HandCard(me, "Visions of Glory", "Sorcery", visionsGloryOracle)
	b13SetManaCost(g, me, hand, "{4}{W}")
	effectstest.AddPool(t, g, me.ID, "{C}{C}{W}{W}")
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a hand cast of {4}{W} was paid with four mana: the flashback discount leaked")
	}
}

func TestS82MidB11VisionsOfGloryMakesAHumanPerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Visions of Glory", "Sorcery", visionsGloryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Human"); got != 2 {
		t.Fatalf("Human tokens = %d, want 2 (one per creature you control, counted at resolution)", got)
	}
}

func TestS82MidB11VisionsOfDominancePutsOneCounterThenDoubles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", 2) })
	effectstest.CastCatalogSpell(t, g, "Visions of Dominance", "Sorcery", visionsDominanceOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 6 {
		t.Errorf("two counters, plus one, doubled: %d, want 6", got)
	}
	// Any creature is a legal target, an opponent's included.
	effectstest.CastCatalogSpell(t, g, "Visions of Dominance", "Sorcery", visionsDominanceOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, theirs, "+1/+1"); got != 2 {
		t.Errorf("a bare opposing creature ends on %d counters, want 2", got)
	}
}

func TestS82MidB11VisionsOfDreadTheOpponentChoosesFromTheirOwnGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	weak := effectstest.B17GraveyardCard(opp, "Weak Bear", "Creature — Bear", "{1}{G}")
	strong := effectstest.B17GraveyardCard(opp, "Strong Wurm", "Creature — Wurm", "{6}{G}")
	effectstest.B17GraveyardCard(opp, "Dead Bolt", "Instant", "{R}")
	mine := effectstest.B17GraveyardCard(me, "My Elf", "Creature — Elf", "{G}")
	spell := effectstest.CastCatalogSpell(t, g, "Visions of Dread", "Sorcery", visionsDreadOracle, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsPromptFrom(g, spell)
	if c == nil {
		t.Fatal("no choose-cards prompt from Visions of Dread")
	}
	if c.Chooser != opp.ID {
		t.Errorf("the prompt is for %v, want the targeted opponent", c.Chooser)
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{weak}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := g.LookupCardForEffect(weak)
	if !ok || !g.Battlefield.Contains(weak) {
		t.Fatal("the chosen creature is not on the battlefield")
	}
	if card.Controller != me.ID {
		t.Errorf("the creature is controlled by %v, want the caster", card.Controller)
	}
	if !opp.Graveyard.Contains(strong) || !me.Graveyard.Contains(mine) {
		t.Error("only the chosen card should have left a graveyard")
	}
}

func TestS82MidB11VisionsOfDreadRefusesYouAsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Visions of Dread", "Sorcery", visionsDreadOracle, effectstest.PlayerRefs(g.Seats[0].ID)); err == nil {
		t.Error("Visions of Dread targets an opponent, not you")
	}
}

func TestS82MidB11VisionsOfRuinEachOpponentSacrificesAnArtifactAndYouGetTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a1 := effectstest.B12Permanent(g, g.Seats[1].ID, "Rock", "Artifact")
	a2 := effectstest.B12Permanent(g, g.Seats[2].ID, "Gem", "Artifact")
	keep := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Visions of Ruin", "Sorcery", visionsRuinOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a1) || g.Battlefield.Contains(a2) {
		t.Error("each opponent should have sacrificed their artifact")
	}
	if !g.Battlefield.Contains(keep) || !g.Battlefield.Contains(mine) {
		t.Error("only opponents' artifacts are sacrificed")
	}
	if got := effectstest.B43TokensNamed(g, me.ID, "Treasure"); got != 2 {
		t.Errorf("Treasures = %d, want 2 (one per artifact sacrificed)", got)
	}
}

func TestS82MidB11VisionsOfRuinMakesNoTreasureWhenNoOneHasAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Visions of Ruin", "Sorcery", visionsRuinOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Treasure"); got != 0 {
		t.Errorf("Treasures = %d, want 0", got)
	}
}

func TestS82MidB11VisionsOfDuplicityExchangesTwoCreaturesYouDontControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	a := effectstest.B12Creature(g, o1.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, o2.ID, "Wolf", "Creature — Wolf", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Elf", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Visions of Duplicity", "Sorcery", visionsDuplicityOracle, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	ca, _ := g.LookupCardForEffect(a)
	cb, _ := g.LookupCardForEffect(b)
	if ca.Controller != o2.ID || cb.Controller != o1.ID {
		t.Errorf("controllers after the exchange: %v and %v, want the two opponents swapped", ca.Controller, cb.Controller)
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Visions of Duplicity", "Sorcery", visionsDuplicityOracle, effectstest.CardRefs(a, mine)); err == nil {
		t.Error("a creature you control is not a legal target")
	}
}

func TestS82MidB11PatricianGeistLordsOtherSpiritsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geist := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Patrician Geist", TypeLine: "Creature — Spirit Knight", OracleID: patricianGeistOracle,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	mySpirit := effectstest.B12Creature(g, me.ID, "My Spirit", "Creature — Spirit", 1, 1)
	theirSpirit := effectstest.B12Creature(g, opp.ID, "Their Spirit", "Creature — Spirit", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p := effectstest.EffectivePower(t, g, mySpirit); p != 2 {
		t.Errorf("your other Spirit has power %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, theirSpirit); p != 1 {
		t.Errorf("an opposing Spirit has power %d, want 1", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Spirit has power %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, geist); p != 2 {
		t.Errorf("the Geist itself is a %d/x, want 2 (other Spirits only)", p)
	}
}

func TestS82MidB11PatricianGeistDiscountsAGraveyardCastByOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedPermanentWithOracle(g, me.ID, "Patrician Geist", "Creature — Spirit Knight", patricianGeistOracle)
	id := s82MidB11Graveyard(t, g, "Seize the Storm", seizeTheStormOracle) // flashback {6}{R}
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{C}{C}{R}")
	if err := effectstest.S82MidB11Flashback(g, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("flashback for {5}{R} under the Geist: %v", err)
	}
}

func TestS82MidB11PatricianGeistDoesNotDiscountAHandCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedPermanentWithOracle(g, me.ID, "Patrician Geist", "Creature — Spirit Knight", patricianGeistOracle)
	hand := effectstest.B13HandCard(me, "Seize the Storm", "Sorcery", seizeTheStormOracle)
	b13SetManaCost(g, me, hand, "{4}{R}")
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{R}")
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a hand cast got the graveyard discount")
	}
}

func TestS82MidB11OminousRoostMakesABirdOnEnteringAndOnAGraveyardCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Ominous Roost", "Enchantment", ominousRoostOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Bird"); got != 1 {
		t.Fatalf("Birds after the Roost entered = %d, want 1", got)
	}
	// A cast from hand does not trigger it.
	effectstest.CastCatalogSpell(t, g, "Seize the Storm", "Sorcery", seizeTheStormOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Bird"); got != 1 {
		t.Errorf("a hand cast made a Bird: %d", got)
	}
	// A cast from your graveyard does.
	id := s82MidB11Graveyard(t, g, "Croaking Counterpart", croakingCounterpartOrcl)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Bird"); got != 2 {
		t.Errorf("Birds after a graveyard cast = %d, want 2", got)
	}
}

func TestS82MidB11OminousRoostBirdCanBlockOnlyCreaturesWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Ominous Roost", "Enchantment", ominousRoostOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var bird game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Bird" {
			bird = c
		}
	}
	if bird.InstanceID == uuid.Nil {
		t.Fatal("no Bird")
	}
	if !bird.IsCreature() || !effectstest.HasKeywordOf(g, bird.InstanceID, "flying") || !bird.HasColor("U") {
		t.Errorf("the Bird is %q %v, want a blue Bird creature with flying", bird.TypeLine, bird.Colors)
	}
	if bird.Power != 1 || bird.Toughness != 1 {
		t.Errorf("the Bird is %d/%d, want 1/1", bird.Power, bird.Toughness)
	}
}

func TestS82MidB11RootcoilCreeperMakesAnyColourAndGraveyardOnlyMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedPermanentWithOracle(g, me.ID, "Rootcoil Creeper", "Creature — Plant Horror", rootcoilCreeperOracle)
	spec, _ := effects.Lookup(rootcoilCreeperOracle)
	if len(spec.ManaAbilities) != 2 || len(spec.Activated) != 1 {
		t.Fatalf("Creeper has %d mana abilities and %d activated, want 2 and 1", len(spec.ManaAbilities), len(spec.Activated))
	}
	if got := spec.ManaAbilities[1].Restrictions; len(got) != 1 || got[0] != game.ManaRestrictCastFromGraveyard {
		t.Errorf("the two-mana ability's restrictions = %v, want the graveyard-cast restriction", got)
	}
	if len(spec.ManaAbilities[0].Restrictions) != 0 {
		t.Error("the one-mana ability is unrestricted")
	}
}

func TestS82MidB11RootcoilCreeperReturnsAFlashbackCardFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	creeper := effectstest.SeedPermanentWithOracle(g, me.ID, "Rootcoil Creeper", "Creature — Plant Horror", rootcoilCreeperOracle)
	mine := s82MidB11ExiledCard(g, me.ID, "Seize the Storm", seizeTheStormOracle)
	noFlash := s82MidB11ExiledCard(g, me.ID, "Plain Bolt", "")
	theirs := s82MidB11ExiledCard(g, opp.ID, "Seize the Storm", seizeTheStormOracle)

	for _, bad := range []uuid.UUID{noFlash, theirs} {
		effectstest.AddPool(t, g, me.ID, "{G}{U}")
		if err := g.ActivateCatalogAbility(me.ID, creeper, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bad), Strict: true}); err == nil {
			t.Fatal("targeted a card that has no flashback or that you don't own")
		}
	}
	effectstest.AddPool(t, g, me.ID, "{G}{U}")
	if err := g.ActivateCatalogAbility(me.ID, creeper, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mine), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) {
		t.Error("the flashback card should be back in your hand")
	}
	if g.Battlefield.Contains(creeper) {
		t.Error("the Creeper is exiled as a cost")
	}
	if !g.Exile.Contains(creeper) {
		t.Error("the Creeper should be in exile")
	}
}

func s82MidB11ExiledCard(g *game.Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	id := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", OracleID: oracle, Owner: owner, Controller: owner})
	return id
}

func TestS82MidB11SeizeTheStormElementalCountsSpellsInYourGraveyardAndFlashbackInExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B17GraveyardCard(me, "Bolt", "Instant", "{R}")
	effectstest.B17GraveyardCard(me, "Divination", "Sorcery", "{2}{U}")
	effectstest.B17GraveyardCard(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.B17GraveyardCard(opp, "Their Bolt", "Instant", "{R}")
	s82MidB11ExiledCard(g, me.ID, "Seize the Storm", seizeTheStormOracle)
	s82MidB11ExiledCard(g, opp.ID, "Seize the Storm", seizeTheStormOracle)
	s82MidB11ExiledCard(g, me.ID, "Plain Bolt", "")
	effectstest.CastCatalogSpell(t, g, "Seize the Storm", "Sorcery", seizeTheStormOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var tok game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Elemental" && c.IsToken() {
			tok = c
		}
	}
	if tok.InstanceID == uuid.Nil {
		t.Fatal("no Elemental token")
	}
	if !effectstest.HasKeywordOf(g, tok.InstanceID, "trample") || !tok.HasColor("R") {
		t.Errorf("the Elemental should be red with trample: %v %v", tok.Colors, tok.Keywords)
	}
	// Your graveyard: Bolt, Divination and (once resolved) Seize the Storm
	// itself = 3 instants and sorceries; exile: one of your cards with flashback.
	if p := effectstest.EffectivePower(t, g, tok.InstanceID); p != 4 {
		t.Errorf("the Elemental's power is %d, want 4 (three instants/sorceries + one flashback card in exile)", p)
	}
}

func TestS82MidB11SeizeTheStormFlashbackExilesItAndTheTokenCountsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := s82MidB11Graveyard(t, g, "Seize the Storm", seizeTheStormOracle)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{C}{C}{C}{R}")
	if err := effectstest.S82MidB11Flashback(g, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Fatal("the flashed-back spell should be exiled")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Elemental" {
			if p := effectstest.EffectivePower(t, g, c.InstanceID); p != 1 {
				t.Errorf("the Elemental is %d/x, want 1 (the exiled card with flashback)", p)
			}
			return
		}
	}
	t.Fatal("no Elemental")
}

func TestS82MidB11CroakingCounterpartCopiesAsAOneOneGreenFrog(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wurm := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Wurm", TypeLine: "Artifact Creature — Wurm Construct", Colors: []string{"R"},
		Power: 6, Toughness: 6, Keywords: []string{"trample"}, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	effectstest.CastCatalogSpell(t, g, "Croaking Counterpart", "Sorcery", croakingCounterpartOrcl, effectstest.B16TargetCard(wurm))
	effectstest.PassPriorityAroundTable(t, g)
	var tok game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() {
			tok = c
		}
	}
	if tok.InstanceID == uuid.Nil {
		t.Fatal("no token copy")
	}
	if tok.Name != "Big Wurm" {
		t.Errorf("the copy is named %q, want Big Wurm", tok.Name)
	}
	if p := effectstest.EffectivePower(t, g, tok.InstanceID); p != 1 {
		t.Errorf("the copy's power is %d, want 1", p)
	}
	if !tok.HasSubtype("Frog") || tok.HasSubtype("Wurm") || tok.HasSubtype("Construct") {
		t.Errorf("the copy's type line is %q, want its creature types replaced by Frog", tok.TypeLine)
	}
	if !tok.HasCardType("artifact") || !tok.IsCreature() {
		t.Errorf("the copy should keep its card types: %q", tok.TypeLine)
	}
	if len(tok.Colors) != 1 || tok.Colors[0] != "G" {
		t.Errorf("the copy's colours = %v, want green only", tok.Colors)
	}
}

func TestS82MidB11CroakingCounterpartRefusesAFrog(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	frog := effectstest.B12Creature(g, g.Seats[1].ID, "Pond Frog", "Creature — Frog", 1, 1)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Croaking Counterpart", "Sorcery", croakingCounterpartOrcl, effectstest.B16TargetCard(frog)); err == nil {
		t.Error("a Frog is not a legal target")
	}
}

func TestS82MidB11CroakingCounterpartFlashback(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := s82MidB11Graveyard(t, g, "Croaking Counterpart", croakingCounterpartOrcl)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{G}{U}")
	if err := effectstest.S82MidB11Flashback(g, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("flashback for {3}{G}{U}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back sorcery should be exiled")
	}
}

// b13SetManaCost gives a hand card its printed mana cost, which the bare
// hand-card helper leaves empty.
func b13SetManaCost(g *game.Game, p *game.Player, id uuid.UUID, cost string) {
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID == id {
			p.Hand.Cards[i].ManaCost = cost
		}
	}
}
