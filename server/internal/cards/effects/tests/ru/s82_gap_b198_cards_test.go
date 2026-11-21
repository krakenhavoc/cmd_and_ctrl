package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b198_cards_test.go — Warhammer 40,000 Commander instants, a land and
// sorceries, GAP batch 198 (#5948). Each test plays the real catalog entry
// through the engine.

const (
	gap198DenyWitchOracle  = "ddff34d8-715b-4d6a-8ced-7ff5c4c911bc"
	gap198ChaosOracle      = "1a240953-26e0-429f-9663-dcfe5bc2c1c7"
	gap198NoFearOracle     = "0ea6d165-2c30-41a4-8829-6d0510afc03f"
	gap198TombOracle       = "986f510c-e2ec-423e-a443-51a169939558"
	gap198GrenadeOracle    = "ae24b863-974c-4efa-ab41-23cfa1906c54"
	gap198MandateOracle    = "20582c7b-edd3-4189-98ec-8fabd39034b0"
	gap198NameDeathOracle  = "97f9aa5e-e34d-4029-9213-b32ebaf9859c"
	gap198LegionOracle     = "8cd45064-2ad9-4247-902e-aefaf3bc1a43"
	gap198ExterminatusOrcl = "8d90ffb5-63da-48d5-a5a2-7b61062556f6"
	gap198TyranidOracle    = "33a45c4a-1481-4c7e-b823-435cb84231e8"
	gap198GalaxyOracle     = "8c70980b-3c38-4624-8f19-5e29d271acf5"
)

func TestGap198EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap198DenyWitchOracle:  "Deny the Witch",
		gap198ChaosOracle:      "Chaos Mutation",
		gap198NoFearOracle:     "And They Shall Know No Fear",
		gap198TombOracle:       "Tomb Fortress",
		gap198GrenadeOracle:    "Blight Grenade",
		gap198MandateOracle:    "Mandate of Abaddon",
		gap198NameDeathOracle:  "Their Name Is Death",
		gap198LegionOracle:     "Their Number Is Legion",
		gap198ExterminatusOrcl: "Exterminatus",
		gap198TyranidOracle:    "Tyranid Invasion",
		gap198GalaxyOracle:     "Let the Galaxy Burn",
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

// --- Deny the Witch ---------------------------------------------------------------

func TestGap198DenyTheWitchCountersAnAbilityAndItsControllerLosesLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	src := effectstest.PushCatalogPermanent(g, opp.ID, "Their Engine", "Artifact", "", false)
	if err := g.ActivateAbility(opp.ID, src, game.AbilityParams{Label: "Their Engine — an activation"}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	ability := effectstest.AcItemLabelled(g, "Their Engine — an activation")
	fired := false
	g.WithWriteLock(func() {
		g.StackMeta[ability].Effect = func(_ *game.Game, _ *game.StackItem) error { fired = true; return nil }
	})
	before, mine := opp.Life, me.Life
	effectstest.CastCatalogSpell(t, g, "Deny the Witch", "Instant", gap198DenyWitchOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ability}})
	effectstest.PassPriorityAroundTable(t, g)
	if fired {
		t.Error("the countered ability resolved")
	}
	if opp.Life != before-2 {
		t.Errorf("the ability's controller is at %d, want %d (two creatures of mine)", opp.Life, before-2)
	}
	if me.Life != mine {
		t.Errorf("the caster lost life: %d -> %d", mine, me.Life)
	}
}

func TestGap198DenyTheWitchCountersASpellAndItsControllerLosesLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	spell := effectstest.CastCatalogSpell(t, g, "Tyranid Invasion", "Sorcery", gap198TyranidOracle, nil)
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Deny the Witch", "Instant", gap198DenyWitchOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(spell) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
	if me.Life != before-1 {
		t.Errorf("life %d, want %d", me.Life, before-1)
	}
}

// --- Chaos Mutation ---------------------------------------------------------------

func TestGap198ChaosMutationReplacesEachCreatureAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	myHit := effectstest.PushLibraryTyped(me, "My Hit", "Creature — Elf")
	mySkip := effectstest.PushLibraryTyped(me, "My Skipped", "Sorcery")
	_ = myHit
	// push order: the top is the last pushed, so the skipped sorcery is revealed first.
	theirHit := effectstest.PushLibraryTyped(opp, "Their Hit", "Creature — Goblin")
	theirSkip := effectstest.PushLibraryTyped(opp, "Their Skipped", "Instant")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Chaos Mutation", "Instant", gap198ChaosOracle, effectstest.CardRefs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 4 && len(g.PendingChoices) > 0; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("a targeted creature was not exiled")
	}
	if !g.Exile.Contains(mine) || !g.Exile.Contains(theirs) {
		t.Error("the targeted creatures are not in exile")
	}
	if !g.Battlefield.Contains(myHit) || !g.Battlefield.Contains(theirHit) {
		t.Errorf("the creature cards did not enter (mine %v, theirs %v)", g.Battlefield.Contains(myHit), g.Battlefield.Contains(theirHit))
	}
	if g.Battlefield.Contains(mySkip) || g.Battlefield.Contains(theirSkip) {
		t.Error("a non-creature revealed card entered")
	}
	// The revealed non-creature cards went to the bottom, not the top and not away.
	if !libraryContains(me, mySkip) || !libraryContains(opp, theirSkip) {
		t.Error("a revealed non-creature card left the library")
	}
	// The bottom of a zone is index 0 (Zone.PushBottom).
	if me.Library.Cards[0].InstanceID != mySkip {
		t.Error("my revealed sorcery is not on the bottom")
	}
	if opp.Library.Cards[0].InstanceID != theirSkip {
		t.Error("their revealed instant is not on the bottom")
	}
}

func libraryContains(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Library.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestGap198ChaosMutationRefusesTwoCreaturesOfOneController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.PushVanillaCreature(g, opp.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "Bear B", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Chaos Mutation", "Instant", gap198ChaosOracle, effectstest.CardRefs(a, b)); err == nil {
		t.Fatal("two creatures with one controller were accepted")
	}
}

func TestGap198ChaosMutationAcceptsNoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Chaos Mutation", "Instant", gap198ChaosOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

// --- And They Shall Know No Fear --------------------------------------------------

func TestGap198NoFearPumpsAndProtectsTheChosenTypeOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marine := effectstest.B12Creature(g, me.ID, "Marine", "Creature — Astartes", 2, 2)
	marine2 := effectstest.B12Creature(g, me.ID, "Marine Two", "Creature — Astartes Warrior", 3, 3)
	ork := effectstest.B12Creature(g, me.ID, "Ork", "Creature — Orc", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Marine", "Creature — Astartes", 2, 2)
	effectstest.CastAndPause(t, g, "And They Shall Know No Fear", "Instant", gap198NoFearOracle)
	effectstest.AnswerCreatureType(t, g, me.ID, "Astartes")
	for _, id := range []uuid.UUID{marine, marine2} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "indestructible") {
			t.Error("a chosen-type creature is not indestructible")
		}
	}
	if p := effectstest.EffectivePower(t, g, marine); p != 3 {
		t.Errorf("marine power %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, marine2); p != 4 {
		t.Errorf("marine two power %d, want 4", p)
	}
	if effectstest.HasEffectiveKeyword(t, g, ork, "indestructible") || effectstest.EffectivePower(t, g, ork) != 2 {
		t.Error("a creature of another type was affected")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") || effectstest.EffectivePower(t, g, theirs) != 2 {
		t.Error("an opponent's creature of the type was affected")
	}
	late := effectstest.B12Creature(g, me.ID, "Late Marine", "Creature — Astartes", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, late, "indestructible") {
		t.Error("a creature that entered later was affected (CR 611.2c)")
	}
}

// --- Tomb Fortress ----------------------------------------------------------------

func gap198ActivateTomb(t *testing.T, g *game.Game, me *game.Player, land uuid.UUID) {
	t.Helper()
	me.ManaPool.EmptyPool()
	effectstest.FillPool(me, 2)
	effectstest.FillPoolColored(me, "B", 3)
	effectstest.S58p6Activate(t, g, me.ID, land, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGap198TombFortressEntersTappedAndTapsForBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12PlayFromHand(t, g, "Tomb Fortress", "Land", gap198TombOracle, game.CastSpellParams{})
	if !effectstest.IsTapped(g, id) {
		t.Fatal("Tomb Fortress entered untapped")
	}
	effectstest.FindBattlefieldCardByID(g, id).Tapped = false
	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for B: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "B" {
		t.Errorf("pool = %v, want one B", me.ManaPool.Units())
	}
}

func TestGap198TombFortressExilesItselfMillsFourAndReturnsAnUntappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.S58p6Push(g, me.ID, "Tomb Fortress", "Land", gap198TombOracle, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	milled := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: milled, Name: "Milled Ogre", TypeLine: "Creature — Ogre",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	older := effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	sorcery := effectstest.PushGraveyardCardTyped(me, "Old Sorcery", "Sorcery")
	lib := len(me.Library.Cards)

	gap198ActivateTomb(t, g, me, land)

	if g.Battlefield.Contains(land) || !g.Exile.Contains(land) {
		t.Error("the land was not exiled as a cost")
	}
	if got := lib - len(me.Library.Cards); got != 4 {
		t.Errorf("milled %d, want 4", got)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no creature-card prompt after the mill")
	}
	for _, id := range pick.ChooseCards {
		if id == sorcery {
			t.Fatal("a non-creature card was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{sorcery}); err == nil {
		t.Fatal("a sorcery was accepted as the pick")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, milled)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(milled) {
		t.Fatal("the chosen creature did not enter")
	}
	if effectstest.IsTapped(g, milled) {
		t.Error("the returned creature entered tapped")
	}
	if !me.Graveyard.Contains(older) {
		t.Error("the unchosen creature left the graveyard")
	}
}

func TestGap198TombFortressIsSorceryOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.S58p6Push(g, me.ID, "Tomb Fortress", "Land", gap198TombOracle, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.CastCatalogSpell(t, g, "Tyranid Invasion", "Sorcery", gap198TyranidOracle, nil) // leaves the stack non-empty
	me.ManaPool.EmptyPool()
	effectstest.FillPool(me, 2)
	effectstest.FillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with a spell on the stack")
	}
}

// --- Blight Grenade ---------------------------------------------------------------

func TestGap198BlightGrenadeDestroysTheTargetAndShrinksEveryOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 6, 6)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Rat", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Blight Grenade", "Sorcery", gap198GrenadeOracle, effectstest.CardRefs(target))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) {
		t.Error("the target survived")
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 3/3 survived -3/-3")
	}
	if !g.Battlefield.Contains(mine) || effectstest.EffectivePower(t, g, mine) != 1 {
		t.Error("my 4/4 should be a 1/1")
	}
	late := effectstest.B12Creature(g, opp.ID, "Late", "Creature — Rat", 3, 3)
	if effectstest.EffectivePower(t, g, late) != 3 {
		t.Error("a creature that entered later was shrunk (CR 611.2c)")
	}
}

func TestGap198BlightGrenadeRefusesANonCreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.B12Permanent(g, g.Seats[1].ID, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Blight Grenade", "Sorcery", gap198GrenadeOracle, effectstest.CardRefs(rock)); err == nil {
		t.Fatal("an artifact is not a legal target")
	}
}

// --- Mandate of Abaddon -----------------------------------------------------------

func TestGap198MandateDestroysEveryCreatureWithLessPowerThanTheChosenOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chosen := effectstest.B12Creature(g, me.ID, "Chosen", "Creature — Giant", 4, 4)
	mineSmall := effectstest.B12Creature(g, me.ID, "Mine Small", "Creature — Bear", 3, 3)
	equal := effectstest.B12Creature(g, opp.ID, "Equal", "Creature — Giant", 4, 1)
	bigger := effectstest.B12Creature(g, opp.ID, "Bigger", "Creature — Giant", 5, 5)
	theirSmall := effectstest.B12Creature(g, opp.ID, "Their Small", "Creature — Rat", 0, 5)
	effectstest.CastCatalogSpell(t, g, "Mandate of Abaddon", "Sorcery", gap198MandateOracle, effectstest.CardRefs(chosen))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{chosen, equal, bigger} {
		if !g.Battlefield.Contains(id) {
			t.Error("a creature with power 4 or more was destroyed")
		}
	}
	for _, id := range []uuid.UUID{mineSmall, theirSmall} {
		if g.Battlefield.Contains(id) {
			t.Error("a creature with power less than 4 survived")
		}
	}
}

func TestGap198MandateRefusesACreatureYouDoNotControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Giant", 4, 4)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Mandate of Abaddon", "Sorcery", gap198MandateOracle, effectstest.CardRefs(theirs)); err == nil {
		t.Fatal("an opponent's creature is not a legal target")
	}
}

// --- Their Name Is Death ----------------------------------------------------------

func TestGap198TheirNameIsDeathSparesArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	golem := effectstest.B12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Their Name Is Death", "Sorcery", gap198NameDeathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a nonartifact creature survived")
	}
	if !g.Battlefield.Contains(golem) || !g.Battlefield.Contains(rock) {
		t.Error("an artifact was destroyed")
	}
}

// --- Their Number Is Legion -------------------------------------------------------

func TestGap198TheirNumberIsLegionMakesTappedArtifactWarriorsGainsLifeAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Pebble", "Land")
	before := me.Life
	effectstest.AdvanceToMain(t, g)
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Their Number Is Legion", TypeLine: "Sorcery",
		OracleID: gap198LegionOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	warriors := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Necron Warrior" && c.Controller == me.ID {
			warriors++
			if !c.Tapped {
				t.Error("a Necron Warrior entered untapped")
			}
			if !c.IsArtifact() || !c.IsCreature() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
				t.Errorf("token is %s %d/%d", c.TypeLine, c.CurrentPower(), c.CurrentToughness())
			}
		}
	}
	if warriors != 3 {
		t.Errorf("created %d warriors, want 3", warriors)
	}
	if me.Life != before+4 {
		t.Errorf("life %d, want %d (the rock and three tokens)", me.Life, before+4)
	}
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("the spell did not exile itself")
	}
}

func TestGap198TheirNumberIsLegionCanBeCastFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Their Number Is Legion", TypeLine: "Sorcery",
		ManaCost: "{X}{B}{B}{B}{B}", OracleID: gap198LegionOracle, Owner: me.ID, Controller: me.ID})
	me.ManaPool.EmptyPool()
	effectstest.FillPoolColored(me, "B", 5)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 1, FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("the spell cast from the graveyard did not end up in exile")
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Necron Warrior" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("created %d warriors, want 1", n)
	}
}

// --- Exterminatus -----------------------------------------------------------------

func TestGap198ExterminatusStripsOpponentsIndestructibleAndDestroysEveryNonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their God", TypeLine: "Creature — God",
		Power: 5, Toughness: 5, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID})
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My God", TypeLine: "Creature — God",
		Power: 5, Toughness: 5, Keywords: []string{"indestructible"}, Owner: me.ID, Controller: me.ID})
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Exterminatus", "Sorcery", gap198ExterminatusOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{theirs, rock, bear} {
		if g.Battlefield.Contains(id) {
			t.Error("a nonland permanent survived")
		}
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own indestructible creature was destroyed; only opponents' lose it")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was destroyed")
	}
}

// --- Tyranid Invasion -------------------------------------------------------------

func TestGap198TyranidInvasionMakesOneTrampleWarriorPerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Tyranid Invasion", "Sorcery", gap198TyranidOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Tyranid Warrior" && c.Controller == me.ID {
			n++
			if c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "trample") {
				t.Errorf("token is %d/%d", c.CurrentPower(), c.CurrentToughness())
			}
		}
	}
	if want := len(g.Seats) - 1; n != want {
		t.Errorf("created %d tokens, want %d", n, want)
	}
}

// --- Let the Galaxy Burn ----------------------------------------------------------

func TestGap198LetTheGalaxyBurnSparesCreaturesThatEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	old := effectstest.PushVanillaCreature(g, opp.ID, "Old Ogre", 5, 5)
	tough := effectstest.PushVanillaCreature(g, opp.ID, "Tough Wall", 0, 9)
	fresh := effectstest.PushVanillaCreature(g, opp.ID, "Fresh Ogre", 1, 1)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	g.WithWriteLock(func() {
		if g.TurnTally.Entered == nil {
			g.TurnTally.Entered = map[uuid.UUID]int{}
		}
		g.TurnTally.Entered[fresh]++
	})
	effectstest.Pr8CastX(t, g, "Let the Galaxy Burn", "Sorcery", gap198GalaxyOracle, 3, nil)
	for i := 0; i < 6 && len(g.PendingChoices) > 0; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(old) || g.Battlefield.Contains(mine) {
		t.Error("a creature with toughness at most 5 survived X+2 = 5 damage")
	}
	if !g.Battlefield.Contains(tough) {
		t.Error("a 0/9 should survive 5 damage")
	}
	if !g.Battlefield.Contains(fresh) {
		t.Error("a creature that entered this turn was damaged")
	}
	if got := effectstest.DamageMarkedOn(g, tough); got != 5 {
		t.Errorf("the wall has %d damage, want 5", got)
	}
}
