package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsn "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/n"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b6_cards_test.go — Final Fantasy (FIN), batch 6 (#4462). Each test
// plays the real catalog entry and asserts what its text says.

const (
	finB6GolbezOracle    = "533d5bf9-0912-451f-a3f4-3842b0a75b1d"
	finB6LockeOracle     = "bdcb7aed-3595-4c7e-b9da-543e92de919a"
	finB6NoctisOracle    = "2d44bdd2-aa31-415a-be71-d9e98ef12334"
	finB6RydiaOracle     = "a08f79b7-d66f-47da-9670-afd7409c0c5c"
	finB6UltimeciaOracle = "70773879-b773-4e81-99ca-8f4bba478962"
	finB6MinstrelOracle  = "e6e6e3d2-f2fb-46a1-8798-13a2f2efe8cc"
	finB6YunaOracle      = "3dee768a-612c-4960-bad2-2fe5b83a70c3"
	finB6IronGiantOracle = "325b4288-fe1b-4e29-ae13-72924b4db3dd"
	finB6LionHeartOracle = "bd4151cc-743a-40f0-9e51-1f67555d9ab3"
	finB6PandoraOracle   = "34c65c7f-33ab-4f91-84ce-d0c8f4fc5832"
)

func TestFinB6EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finB6GolbezOracle:           "Golbez, Crystal Collector",
		finB6LockeOracle:            "Locke Cole",
		finB6NoctisOracle:           "Noctis, Prince of Lucis",
		finB6RydiaOracle:            "Rydia, Summoner of Mist",
		finB6UltimeciaOracle:        "Ultimecia, Time Sorceress",
		finB6UltimeciaOracle + "#1": "Ultimecia, Omnipotent",
		finB6MinstrelOracle:         "The Wandering Minstrel",
		finB6YunaOracle:             "Yuna, Hope of Spira",
		finB6IronGiantOracle:        "Iron Giant",
		finB6LionHeartOracle:        "Lion Heart",
		finB6PandoraOracle:          "Lunatic Pandora",
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

// --- Golbez, Crystal Collector -------------------------------------------------------

func finB6Golbez(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Golbez, Crystal Collector", "Legendary Creature — Human Wizard", finB6GolbezOracle, 1, 4)
}

func finB6Artifacts(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Push(g, owner, "Relic", "Artifact", "", 0, 0)
	}
}

func TestFinB6GolbezSurveilsWhenAnArtifactYouControlEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB6Golbez(g, me.ID)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Rock", TypeLine: "Artifact"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil prompt %+v, want one card", c)
	}
}

func TestFinB6GolbezIgnoresNonArtifactsAndOpponentsArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	finB6Golbez(g, me.ID)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Error("a surveil was offered for a non-artifact or an opponent's artifact")
	}
}

func TestFinB6GolbezReturnsACreatureCardWithFourArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], effectstest.FinABOpp(g)
	finB6Golbez(g, me.ID)
	finB6Artifacts(g, me.ID, 4)
	bear := effectstest.PushGraveyardCardWithTypeLineID(me, "Big Bear", "Creature — Bear")
	if c, ok := g.LookupCardForEffect(bear); ok {
		_ = c
	}
	oppLife := opp.Life
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Fatal("the creature card did not return to hand")
	}
	if opp.Life != oppLife {
		t.Errorf("an opponent lost life with only four artifacts (%d → %d)", oppLife, opp.Life)
	}
}

func TestFinB6GolbezDrainsForTheReturnedCardsPowerWithEightArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	finB6Golbez(g, me.ID)
	finB6Artifacts(g, me.ID, 8)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Big Bear", TypeLine: "Creature — Bear", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want -= 3
		}
		if p.Life != want {
			t.Errorf("%s has %d life, want %d", p.Name, p.Life, want)
		}
	}
}

func TestFinB6GolbezDoesNothingWithThreeArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	finB6Golbez(g, me.ID)
	finB6Artifacts(g, me.ID, 3)
	bear := effectstest.PushGraveyardCardWithTypeLineID(me, "Big Bear", "Creature — Bear")
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || !me.Graveyard.Contains(bear) {
		t.Error("Golbez triggered with three artifacts")
	}
}

func TestFinB6GolbezDrainsNothingWhenTheTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], effectstest.FinABOpp(g)
	finB6Golbez(g, me.ID)
	finB6Artifacts(g, me.ID, 8)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Big Bear", TypeLine: "Creature — Bear", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	oppLife := opp.Life
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, id)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife {
		t.Errorf("an opponent lost life although the target was gone (%d → %d)", oppLife, opp.Life)
	}
}

// --- Locke Cole ----------------------------------------------------------------------

func TestFinB6LockeColeLootsOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	locke := effectstest.B12Push(g, me.ID, "Locke Cole", "Legendary Creature — Human Rogue", finB6LockeOracle, 2, 3)
	for _, kw := range []string{"deathtouch", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, locke, kw) {
			t.Errorf("Locke Cole lacks %s", kw)
		}
	}
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, locke, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard after the draw")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d before the discard, want +1 (draw first)", hand, me.Hand.Size())
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d after the loot, want unchanged", hand, me.Hand.Size())
	}
}

func TestFinB6LockeColeIgnoresNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	locke := effectstest.B12Push(g, me.ID, "Locke Cole", "Legendary Creature — Human Rogue", finB6LockeOracle, 2, 3)
	hand := me.Hand.Size()
	effectstest.DealNoncombatDamage(g, me.ID, locke, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("noncombat damage triggered the loot")
	}
}

// --- Iron Giant ----------------------------------------------------------------------

func TestFinB6IronGiantHasReachVigilanceAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	giant := effectstest.B12Push(g, me.ID, "Iron Giant", "Artifact Creature — Demon", finB6IronGiantOracle, 6, 6)
	for _, kw := range []string{"reach", "vigilance", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, giant, kw) {
			t.Errorf("Iron Giant lacks %s", kw)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, giant, "flying") {
		t.Error("Iron Giant has flying")
	}
}

// --- Lion Heart ----------------------------------------------------------------------

func TestFinB6LionHeartDealsTwoDamageToAnyTargetWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Lion Heart", TypeLine: "Artifact — Equipment", OracleID: finB6LionHeartOracle}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("2 damage did not kill the 2/2")
	}
	life := opp.Life
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Lion Heart", TypeLine: "Artifact — Equipment", OracleID: finB6LionHeartOracle}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent has %d life, want %d", opp.Life, life-2)
	}
}

func TestFinB6LionHeartEquipGivesPlusTwoPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	heart := effectstest.B12Push(g, me.ID, "Lion Heart", "Artifact — Equipment", finB6LionHeartOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, heart, 0, game.ActivateAbilityParams{Targets: effectstest.TdCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 4 || th != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 4/3", p, th)
	}
}

// --- Lunatic Pandora -----------------------------------------------------------------

func TestFinB6LunaticPandoraSurveilsForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pandora := effectstest.B12Push(g, me.ID, "Lunatic Pandora", "Legendary Artifact", finB6PandoraOracle, 0, 0)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, pandora, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil prompt %+v, want one card", c)
	}
	if !effectstest.IsTapped(g, pandora) {
		t.Error("Pandora did not tap")
	}
}

func TestFinB6LunaticPandoraSacrificesToDestroyANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	pandora := effectstest.B12Push(g, me.ID, "Lunatic Pandora", "Legendary Artifact", finB6PandoraOracle, 0, 0)
	rock := effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	land := effectstest.B12Push(g, opp.ID, "Their Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, pandora, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.TdCard(land)}); err == nil {
		t.Fatal("a land was accepted as a nonland permanent")
	}
	effectstest.B16Activate(t, g, me.ID, pandora, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.TdCard(rock)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pandora) {
		t.Error("Pandora was not sacrificed")
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the targeted permanent was not destroyed")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land was destroyed")
	}
}

func TestFinB6NoctisCastsAnArtifactFromTheGraveyardForThreeLifeAndAFinalityCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Noctis, Prince of Lucis", "Legendary Creature — Human Noble", finB6NoctisOracle, 4, 3)
	rock := effectstest.FinB6GraveyardArtifact(me, "Rock", "Artifact", "{2}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	life := me.Life
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: cardsn.NoctisAltCostKey}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Fatal("the artifact did not enter")
	}
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want 3 paid", life, me.Life)
	}
	if got := effectstest.CountersOn(g, rock, game.CounterFinality); got != 1 {
		t.Errorf("the artifact has %d finality counters, want 1", got)
	}
	// With a finality counter it is exiled instead of going to the graveyard.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, rock); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if me.Graveyard.Contains(rock) {
		t.Error("the artifact went to the graveyard despite its finality counter")
	}
}

func TestFinB6NoctisPaysTheArtifactsOtherCostsToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Noctis, Prince of Lucis", "Legendary Creature — Human Noble", finB6NoctisOracle, 4, 3)
	rock := effectstest.FinB6GraveyardArtifact(me, "Rock", "Artifact", "{2}")
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: cardsn.NoctisAltCostKey}); err == nil {
		t.Error("the artifact was cast without paying its mana cost")
	}
}

func TestFinB6NoctisRefusesAnythingButAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Noctis, Prince of Lucis", "Legendary Creature — Human Noble", finB6NoctisOracle, 4, 3)
	bear := effectstest.FinB6GraveyardArtifact(me, "Bear", "Creature — Bear", "{2}")
	bolt := effectstest.FinB6GraveyardArtifact(me, "Bolt", "Instant", "{1}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	for _, id := range []uuid.UUID{bear, bolt} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: cardsn.NoctisAltCostKey}); err == nil {
			t.Error("a nonartifact card was cast from the graveyard")
		}
	}
}

func TestFinB6NoctisNeedsThreeLifeToPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Noctis, Prince of Lucis", "Legendary Creature — Human Noble", finB6NoctisOracle, 4, 3)
	rock := effectstest.FinB6GraveyardArtifact(me, "Rock", "Artifact", "{2}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	me.Life = 2
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: cardsn.NoctisAltCostKey}); err == nil {
		t.Error("an artifact was cast with 2 life")
	}
	if me.Life != 2 {
		t.Errorf("life changed to %d", me.Life)
	}
}

func TestFinB6AnArtifactIsNotCastableFromTheGraveyardWithoutNoctis(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.FinB6GraveyardArtifact(me, "Rock", "Artifact", "{2}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Error("an artifact was cast from the graveyard with no Noctis")
	}
}

// --- Rydia, Summoner of Mist ---------------------------------------------------------

func finB6Rydia(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Rydia, Summoner of Mist", "Legendary Creature — Human Shaman", finB6RydiaOracle, 1, 2)
}

func TestFinB6RydiaRummagesOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB6Rydia(g, me.ID)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("landfall offered no discard")
	}
	hand := me.Hand.Size()
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want a discard and a draw", hand, me.Hand.Size())
	}
}

func finB6Saga(me *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Enchantment — Saga", ManaCost: cost, Owner: me.ID, Controller: me.ID})
	return id
}

func TestFinB6RydiaReturnsASagaOfManaValueXWithFinalityAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rydia := finB6Rydia(g, me.ID)
	saga := finB6Saga(me, "Summon: Test", "{2}{G}")
	other := finB6Saga(me, "Summon: Other", "{G}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, rydia, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: effectstest.TdCard(other)}); err == nil {
		t.Fatal("a Saga of mana value 1 was accepted for X = 3")
	}
	effectstest.B16Activate(t, g, me.ID, rydia, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: effectstest.TdCard(saga)})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(saga) {
		t.Fatal("the Saga did not return")
	}
	if got := effectstest.CountersOn(g, saga, game.CounterFinality); got != 1 {
		t.Errorf("the Saga has %d finality counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, saga, "haste") {
		t.Error("the returned Saga has no haste")
	}
	if !me.Graveyard.Contains(other) {
		t.Error("the other Saga left the graveyard")
	}
}

func TestFinB6RydiaSummonRefusesANonSaga(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rydia := finB6Rydia(g, me.ID)
	bear := effectstest.FinB6GraveyardArtifact(me, "Bear", "Creature — Bear", "{2}{G}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, rydia, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: effectstest.TdCard(bear)}); err == nil {
		t.Error("a non-Saga card was accepted")
	}
}

func TestFinB6RydiaSummonIsSorceryOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rydia := finB6Rydia(g, me.ID)
	saga := finB6Saga(me, "Summon: Test", "{2}{G}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	// A spell on the stack closes the sorcery window.
	effectstest.CastCatalogSpell(t, g, "Bear Cub", "Instant", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, rydia, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: effectstest.TdCard(saga)}); err == nil {
		t.Error("the Summon ability was activated with a spell on the stack")
	}
}

// --- Ultimecia, Time Sorceress // Ultimecia, Omnipotent ------------------------------

func finB6UltimeciaRow() cards.Card {
	row := effectstest.TransformRow(finB6UltimeciaOracle, "Ultimecia, Time Sorceress", "Legendary Creature — Human Warlock", "{3}{U}{B}",
		"Ultimecia, Omnipotent", "Legendary Creature — Nightmare Warlock", "7", "7", []string{"U", "B"})
	row.CardFaces[0].Power, row.CardFaces[0].Toughness = "4", "5"
	return row
}

func finB6UltimeciaMana(me *game.Player) {
	effectstest.B10AddMana(me, "C", "C", "C", "C", "U", "U", "B", "B")
}

func TestFinB6UltimeciaSurveilsTwoWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.AdvanceToMain(t, g)
	effectstest.ImportToBattlefield(t, g, finB6UltimeciaRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil prompt %+v, want two cards", c)
	}
}

func TestFinB6UltimeciaSurveilsTwoWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.ImportToBattlefield(t, g, finB6UltimeciaRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SurveilChoiceFor(g, me.ID); c != nil {
		if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
			t.Fatalf("ResolveSurveil: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.SeedLibrary(me, "A", "B", "C")
	finB5Attack(t, g, id, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil prompt on attack %+v, want two cards", c)
	}
}

func finB6UltimeciaEndStep(t *testing.T, g *game.Game, me *game.Player, graveyard int, pay bool) uuid.UUID {
	t.Helper()
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceToMain(t, g)
	id := effectstest.ImportToBattlefield(t, g, finB6UltimeciaRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SurveilChoiceFor(g, me.ID); c != nil {
		if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
			t.Fatalf("ResolveSurveil: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < graveyard; i++ {
		effectstest.PushGraveyardCardWithTypeLineID(me, "Filler", "Sorcery")
	}
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if graveyard < 8 {
		if effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("the payment was offered with fewer than eight cards in the graveyard")
		}
		return id
	}
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment offered at the end step")
	}
	finB6UltimeciaMana(me) // the pool empties between steps: fund it as the question is asked
	effectstest.AnswerPayUnless(t, g, me.ID, pay)
	return id
}

func TestFinB6UltimeciaPaysAndExilesEightToTransformAndTakeAnExtraTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := finB6UltimeciaEndStep(t, g, me, 9, true)
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c == nil || c.ChooseMin != 8 || c.ChooseMax != 8 {
		t.Fatalf("exile pick %+v, want exactly eight", c)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID).ChooseCards[:8]
	exiled := len(g.Exile.Cards)
	effectstest.AnswerChooseCards(t, g, me.ID, pick...)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(g.Exile.Cards) - exiled; got != 8 {
		t.Errorf("%d cards exiled, want 8", got)
	}
	if me.Graveyard.Size() != 1 {
		t.Errorf("the graveyard holds %d cards, want 1", me.Graveyard.Size())
	}
	if c := effectstest.FindBattlefieldCardByID(g, id); c == nil || c.ActiveFace != 1 {
		t.Fatal("Ultimecia did not transform")
	}
	if p, th := effectstest.EffectivePTOf(t, g, id); p != 7 || th != 7 {
		t.Errorf("Ultimecia, Omnipotent is %d/%d, want 7/7", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "menace") {
		t.Error("Ultimecia, Omnipotent lacks menace")
	}
	if len(g.ExtraTurns) != 1 {
		t.Errorf("%d extra turns queued, want 1 (Time Compression)", len(g.ExtraTurns))
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left in the pool, want 0 (all of {4}{U}{U}{B}{B} paid)", got)
	}
}

func TestFinB6UltimeciaDeclinedStaysAsItIs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := finB6UltimeciaEndStep(t, g, me, 9, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("declining still asked for cards to exile")
	}
	if c := effectstest.FindBattlefieldCardByID(g, id); c == nil || c.ActiveFace != 0 {
		t.Error("Ultimecia transformed although the payment was declined")
	}
	if me.Graveyard.Size() < 9 {
		t.Errorf("the graveyard shrank to %d", me.Graveyard.Size())
	}
	if len(g.ExtraTurns) != 0 {
		t.Error("an extra turn was queued")
	}
}

func TestFinB6UltimeciaOffersNothingWithSevenCardsInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := finB6UltimeciaEndStep(t, g, me, 7, true)
	if c := effectstest.FindBattlefieldCardByID(g, id); c == nil || c.ActiveFace != 0 {
		t.Error("Ultimecia transformed with only seven cards in the graveyard")
	}
}

func TestFinB6UltimeciaEnteringOnItsBackFaceTakesNoExtraTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.ImportToHand(finB6UltimeciaRow(), me)
	var err error
	g.WithWriteLock(func() { err = g.ExileAndReturnTransformedForEffect(id, uuid.Nil) })
	if err != nil {
		t.Fatalf("enter transformed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.ExtraTurns) != 0 {
		t.Errorf("%d extra turns after entering on the back face, want 0 (it never transformed)", len(g.ExtraTurns))
	}
}

// --- The Wandering Minstrel ----------------------------------------------------------

func finB6Minstrel(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "The Wandering Minstrel", "Legendary Creature — Human Bard", finB6MinstrelOracle, 1, 3)
}

func finB6Towns(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Push(g, owner, "Gilded Town", "Land — Town", "", 0, 0)
	}
}

func TestFinB6MinstrelMakesLandsEnterUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB6Minstrel(g, me.ID)
	land := tdCastFromHandLand(t, g, "Forest")
	if effectstest.IsTapped(g, land) {
		t.Error("a land entered tapped under the Minstrel")
	}
}

// tdCastFromHandLand plays a land that would enter tapped: a Town-like
// "enters tapped" land cannot be fixtured by type line alone, so this
// checks the replacement on an entry that asks to be tapped.
func tdCastFromHandLand(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	var err error
	g.WithWriteLock(func() {
		_, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{Tapped: true})
	})
	if err != nil {
		t.Fatalf("put %s: %v", name, err)
	}
	return id
}

func TestFinB6MinstrelDoesNotAffectAnOpponentsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	finB6Minstrel(g, me.ID)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	var err error
	g.WithWriteLock(func() {
		_, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{Tapped: true})
	})
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if !effectstest.IsTapped(g, id) {
		t.Error("an opponent's land entered untapped")
	}
}

func TestFinB6MinstrelMakesAnAllColorsElementalAtCombatWithFiveTowns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	finB6Minstrel(g, me.ID)
	finB6Towns(g, me.ID, 5)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Elemental", me.ID); got != 1 {
		t.Fatalf("%d Elementals, want 1", got)
	}
	v := effectstest.S58p6View(t, g, effectstest.FindBattlefieldByName(g, "Elemental"))
	if p, th := v.CurrentPower(), v.CurrentToughness(); p != 2 || th != 2 {
		t.Errorf("the Elemental is %d/%d, want 2/2", p, th)
	}
	if got := len(v.Effective().Colors); got != 5 {
		t.Errorf("the Elemental has %d colors, want all five", got)
	}
}

func TestFinB6MinstrelNeedsFiveTowns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	finB6Minstrel(g, me.ID)
	finB6Towns(g, me.ID, 4)
	effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Elemental", me.ID); got != 0 {
		t.Errorf("%d Elementals with four Towns, want 0", got)
	}
}

func TestFinB6MinstrelDoesNotTriggerOnAnOpponentsCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToNextSeatsTurn(t, g) // past your own combat
	finB6Minstrel(g, me.ID)
	finB6Towns(g, me.ID, 5)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Elemental", me.ID); got != 0 {
		t.Errorf("%d Elementals on an opponent's combat, want 0", got)
	}
}

func TestFinB6MinstrelPumpsOtherCreaturesByTheTowns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	minstrel := finB6Minstrel(g, me.ID)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	finB6Towns(g, me.ID, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{W}{U}{B}{R}{G}")
	effectstest.B16Activate(t, g, me.ID, minstrel, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 5 || th != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, minstrel); p != 1 || th != 3 {
		t.Errorf("the Minstrel is %d/%d, want 1/3 (other creatures only)", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, theirs); p != 2 || th != 2 {
		t.Errorf("their Bear is %d/%d, want 2/2", p, th)
	}
}

// --- Yuna, Hope of Spira -------------------------------------------------------------

func finB6Yuna(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Yuna, Hope of Spira", "Legendary Creature — Human Cleric", finB6YunaOracle, 3, 5)
}

func TestFinB6YunaGrantsTrampleAndLifelinkDuringYourTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	yuna := finB6Yuna(g, me.ID)
	sage := effectstest.B12Creature(g, me.ID, "Sage", "Enchantment Creature — Human", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, kw := range []string{"trample", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, yuna, kw) || !effectstest.HasEffectiveKeyword(t, g, sage, kw) {
			t.Errorf("Yuna or the enchantment creature lacks %s during your turn", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("a non-enchantment creature has %s", kw)
		}
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	for _, kw := range []string{"trample", "lifelink"} {
		if effectstest.HasEffectiveKeyword(t, g, yuna, kw) || effectstest.HasEffectiveKeyword(t, g, sage, kw) {
			t.Errorf("%s is granted on an opponent's turn", kw)
		}
	}
}

func TestFinB6YunaGrantsNothingToAnOpponentsEnchantmentCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	finB6Yuna(g, me.ID)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Sage", "Enchantment Creature — Human", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's enchantment creature has trample")
	}
}

func TestFinB6YunaHasWardTwoDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	yuna := finB6Yuna(g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastAtWardedCreature(t, g, opp, yuna)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(yuna) {
		t.Error("an unpaid ward did not counter the spell")
	}
}

func TestFinB6YunaReturnsAnEnchantmentWithAFinalityCounterAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	finB6Yuna(g, me.ID)
	ench := effectstest.PushGraveyardCardWithTypeLineID(me, "Old Enchantment", "Enchantment")
	bear := effectstest.PushGraveyardCardWithTypeLineID(me, "Bear", "Creature — Bear")
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	for _, id := range effectstest.LatestPickTarget(g, me.ID).PickTargetCards {
		if id == bear {
			t.Error("a non-enchantment card was a legal target")
		}
	}
	effectstest.PickCard(t, g, me.ID, ench)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ench) {
		t.Fatal("the enchantment did not return")
	}
	if got := effectstest.CountersOn(g, ench, game.CounterFinality); got != 1 {
		t.Errorf("the enchantment has %d finality counters, want 1", got)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, ench); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if me.Graveyard.Contains(ench) {
		t.Error("the returned enchantment went to the graveyard despite its finality counter")
	}
}

func TestFinB6YunaDoesNotTriggerOnAnOpponentsEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	finB6Yuna(g, me.ID)
	effectstest.AdvanceToNextSeatsTurn(t, g) // past your own end step
	ench := effectstest.PushGraveyardCardWithTypeLineID(me, "Old Enchantment", "Enchantment")
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || !me.Graveyard.Contains(ench) {
		t.Error("Yuna triggered on an opponent's end step")
	}
}
