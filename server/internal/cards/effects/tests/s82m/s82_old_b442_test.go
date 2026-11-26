package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dark Ascension batch 442: equipment, damage, removal and activated abilities.

const (
	oldB442Scorch       = "6f9bce77-f7f0-453b-86cd-1e6bd71980a3"
	oldB442DeathsCaress = "aac711d1-2fa0-42f9-9eda-a1642f3ff016"
	oldB442Niblis       = "10391bcb-0932-49a0-8d47-e06e3c10dd68"
	oldB442Collar       = "842e3835-fff3-4684-bfd1-e7064a1fd2e6"
	oldB442Wrack        = "c93c647e-0b9e-4a50-a519-42862d21a0f1"
	oldB442Elbrus       = "da5881ca-b86e-4582-ae0e-1570c59773ae"
	oldB442Demon        = "18932bb9-e054-4ad6-80c8-d2f177312da0"
	oldB442Chosen       = "29328bcb-e991-4095-9439-f493c8fa3f82"
	oldB442Mattock      = "f1f45310-d665-400d-934f-e48e7858ce8c"
	oldB442Disappear    = "dbe28518-d795-4bc5-afa6-63c696c86fca"
	oldB442Quiver       = "f1ebccc2-d12a-4440-8a73-86c1bc901feb"
	oldB442AlphaBrawl   = "cfc023e0-9e9d-48ac-bfb5-74bfd5977dee"
)

func oldB442Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, targets ...game.TargetRef) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, name, typeLine, oracle, targets)
	effectstest.PassPriorityAroundTable(t, g)
}

func oldB442Card(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func TestOldB442CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB442Scorch: "Scorch the Fields", oldB442DeathsCaress: "Death's Caress", oldB442Niblis: "Niblis of the Breath",
		oldB442Collar: "Avacyn's Collar", oldB442Wrack: "Wrack with Madness", oldB442Elbrus: "Elbrus, the Binding Blade",
		oldB442Elbrus + "#1": "Withengar Unbound", oldB442Demon: "Ravenous Demon", oldB442Demon + "#1": "Archdemon of Greed",
		oldB442Chosen: "Chosen of Markov", oldB442Chosen + "#1": "Markov's Servant", oldB442Mattock: "Heavy Mattock",
		oldB442Disappear: "Sudden Disappearance", oldB442Quiver: "Wolfhunter's Quiver", oldB442AlphaBrawl: "Alpha Brawl",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q, completeness %v, caveats %v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB442WrackWithMadnessHitsTheCreatureForItsOwnPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	small := effectstest.B12Creature(g, me.ID, "Tough", "Creature — Beast", 2, 3)
	big := effectstest.B12Creature(g, me.ID, "Brute", "Creature — Beast", 3, 3)
	oldB442Cast(t, g, "Wrack with Madness", "Sorcery", oldB442Wrack, oldB442Card(small))
	if !g.Battlefield.Contains(small) {
		t.Error("a 2/3 dealing 2 to itself should survive")
	}
	oldB442Cast(t, g, "Wrack with Madness", "Sorcery", oldB442Wrack, oldB442Card(big))
	if g.Battlefield.Contains(big) {
		t.Error("a 3/3 dealing 3 to itself should die")
	}
}

func TestOldB442AlphaBrawlHitsOutAndBackSimultaneously(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leader := effectstest.B12Creature(g, opp.ID, "Leader", "Creature — Beast", 3, 3)
	runt := effectstest.B12Creature(g, opp.ID, "Runt", "Creature — Beast", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 1, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Beast", 1, 1)
	oldB442Cast(t, g, "Alpha Brawl", "Sorcery", oldB442AlphaBrawl, oldB442Card(leader))
	if g.Battlefield.Contains(runt) {
		t.Error("the 2/2 takes 3 and dies")
	}
	if !g.Battlefield.Contains(wall) {
		t.Error("the 1/4 takes 3 and survives")
	}
	if g.Battlefield.Contains(leader) {
		t.Error("the leader takes 2 (the Runt, though it died) + 1 (the Wall) = 3 and dies")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creature is out of it")
	}
}

func TestOldB442AlphaBrawlRefusesYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := effectstest.B12Creature(g, g.Seats[0].ID, "Mine", "Creature — Beast", 1, 1)
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Alpha Brawl", "Sorcery", oldB442AlphaBrawl, []game.TargetRef{oldB442Card(mine)})
	if err == nil {
		t.Error("Alpha Brawl accepted a creature its caster controls")
	}
}

func TestOldB442ScorchTheFieldsDestroysALandAndPingsEveryHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.SeedPermanentFor(g, opp.ID, "Forest", "Basic Land — Forest")
	myHuman := effectstest.B12Creature(g, me.ID, "My Human", "Creature — Human Soldier", 1, 1)
	theirHuman := effectstest.B12Creature(g, opp.ID, "Their Human", "Creature — Human", 1, 1)
	bigHuman := effectstest.B12Creature(g, opp.ID, "Big Human", "Creature — Human", 2, 2)
	wolf := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 1, 1)
	oldB442Cast(t, g, "Scorch the Fields", "Sorcery", oldB442Scorch, oldB442Card(land))
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
	if g.Battlefield.Contains(myHuman) || g.Battlefield.Contains(theirHuman) {
		t.Error("a 1/1 Human survived the damage")
	}
	if !g.Battlefield.Contains(bigHuman) || !g.Battlefield.Contains(wolf) {
		t.Error("the 2/2 Human or the Wolf died")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Scorch the Fields", "Sorcery", oldB442Scorch, []game.TargetRef{oldB442Card(bigHuman)}); err == nil {
		t.Error("Scorch the Fields accepted a creature as its target land")
	}
}

func TestOldB442DeathsCaressGainsLifeOnlyForAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	human := effectstest.B12Creature(g, opp.ID, "Human", "Creature — Human Cleric", 1, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 5)
	life := me.Life
	oldB442Cast(t, g, "Death's Caress", "Sorcery", oldB442DeathsCaress, oldB442Card(bear))
	if g.Battlefield.Contains(bear) || me.Life != life {
		t.Fatalf("bear on table %v, life %d→%d: want it destroyed and no life", g.Battlefield.Contains(bear), life, me.Life)
	}
	oldB442Cast(t, g, "Death's Caress", "Sorcery", oldB442DeathsCaress, oldB442Card(human))
	if g.Battlefield.Contains(human) || me.Life != life+4 {
		t.Fatalf("human on table %v, life %d→%d: want it destroyed and +4", g.Battlefield.Contains(human), life, me.Life)
	}
	// A Human that survives the destruction still pays: "if that creature was a Human".
	tough := effectstest.B12Push(g, opp.ID, "Indestructible Human", "Creature — Human", "", 1, 3)
	g.WithWriteLock(func() {
		effectstest.BattlefieldCardPtr(t, g, tough).Keywords = append(effectstest.BattlefieldCardPtr(t, g, tough).Keywords, "indestructible")
	})
	oldB442Cast(t, g, "Death's Caress", "Sorcery", oldB442DeathsCaress, oldB442Card(tough))
	if !g.Battlefield.Contains(tough) || me.Life != life+7 {
		t.Errorf("indestructible Human on table %v, life %d→%d: want it alive and +3", g.Battlefield.Contains(tough), life, me.Life)
	}
}

func TestOldB442NiblisOfTheBreathTapsOrUntapsAtAPrice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	niblis := effectstest.B12Push(g, me.ID, "Niblis of the Breath", "Creature — Spirit", oldB442Niblis, 2, 1)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, niblis, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB442Card(foe)}}); err == nil {
		t.Fatal("activated without {U}")
	}
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, niblis, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB442Card(foe)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if !effectstest.HmlTapped(t, g, foe) || !effectstest.HmlTapped(t, g, niblis) {
		t.Error("the foe should be tapped and the Niblis tapped for the cost")
	}
}

func TestOldB442AvacynsCollarPumpsAndLeavesASpiritWhenAHumanDies(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		spirits  int
	}{{"Creature — Human Soldier", 1}, {"Creature — Bear", 0}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		host := effectstest.B12Creature(g, me.ID, "Host", tc.typeLine, 2, 2)
		effectstest.SomB1Equip(t, g, me.ID, "Avacyn's Collar", oldB442Collar, host)
		if p, tough := effectstest.EffectivePTOf(t, g, host); p != 3 || tough != 2 {
			t.Errorf("%s: %d/%d, want 3/2", tc.typeLine, p, tough)
		}
		if !effectstest.EffectiveAbilitiesContain(t, g, host, "vigilance") {
			t.Errorf("%s: no vigilance", tc.typeLine)
		}
		effectstest.KillCreature(t, g, me.ID, host)
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.TokensNamedControlledBy(g, "Spirit", me.ID); n != tc.spirits {
			t.Errorf("%s: %d Spirit tokens, want %d", tc.typeLine, n, tc.spirits)
		}
	}
}

func TestOldB442HeavyMattockGivesAnExtraBonusToHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mattock := effectstest.SomB1Equip(t, g, me.ID, "Heavy Mattock", oldB442Mattock, human)
	if p, tough := effectstest.EffectivePTOf(t, g, human); p != 4 || tough != 4 {
		t.Errorf("Human: %d/%d, want 4/4", p, tough)
	}
	effectstest.AttachForTest(t, g, mattock, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("Bear: %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, human); p != 2 || tough != 2 {
		t.Errorf("unequipped Human: %d/%d, want 2/2", p, tough)
	}
}

func TestOldB442WolfhuntersQuiverGrantsAPingAndAWerewolfSnipe(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B16Creature(g, me.ID, "Host", "Creature — Bear", 2, 5, "G")
	wolf := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Werewolf", 2, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 3)
	effectstest.AdvanceToMain(t, g)
	quiver := effectstest.B12Push(g, me.ID, "Wolfhunter's Quiver", "Artifact — Equipment", oldB442Quiver, 0, 0)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}")
	effectstest.EquipTo(t, g, me.ID, quiver, host)
	abilities, origins := game.ActivatedAbilitiesWithOrigins(effectstest.GaLayered(t, g, host))
	byLabel := map[string]int{}
	for i, a := range abilities {
		if origins[i].Granted() {
			byLabel[a.Label] = i
		}
	}
	snipe, ok1 := byLabel["{T}: This creature deals 3 damage to target Werewolf creature."]
	ping, ok2 := byLabel["{T}: This creature deals 1 damage to any target."]
	if !ok1 || !ok2 {
		t.Fatalf("granted abilities %v", byLabel)
	}
	if err := g.ActivateCatalogAbility(me.ID, host, snipe, game.ActivateAbilityParams{Strict: true, Ref: oldB442RefOf(t, g, host, snipe), Targets: []game.TargetRef{oldB442Card(bear)}}); err == nil {
		t.Error("the Werewolf ability accepted a non-Werewolf")
	}
	before := opp.Life
	effectstest.B16Activate(t, g, me.ID, host, ping, game.ActivateAbilityParams{Ref: oldB442RefOf(t, g, host, ping), Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life %d → %d, want the 1 damage ping", before, opp.Life)
	}
	effectstest.Untap(g, host)
	effectstest.B16Activate(t, g, me.ID, host, snipe, game.ActivateAbilityParams{Ref: oldB442RefOf(t, g, host, snipe), Targets: []game.TargetRef{oldB442Card(wolf)}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wolf) {
		t.Error("3 damage should kill the 2/3 Werewolf")
	}
}

func oldB442RefOf(t *testing.T, g *game.Game, host uuid.UUID, idx int) string {
	t.Helper()
	_, origins := game.ActivatedAbilitiesWithOrigins(effectstest.GaLayered(t, g, host))
	return origins[idx].Ref
}

func TestOldB442SuddenDisappearanceExilesNonlandPermanentsAndReturnsThemAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")
	land := effectstest.SeedPermanentFor(g, opp.ID, "Forest", "Basic Land — Forest")
	oldB442Cast(t, g, "Sudden Disappearance", "Sorcery", oldB442Disappear, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(rock) {
		t.Fatal("their nonland permanents should be exiled")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("lands stay")
	}
	for i := 0; i < 12 && g.Turn.Step != game.StepEnd; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountOnBattlefield(g, "Bear", opp.ID) != 1 || effectstest.CountOnBattlefield(g, "Rock", opp.ID) != 1 {
		t.Error("the exiled permanents did not come back at the beginning of the end step")
	}
}

func oldB442ElbrusRow() cards.Card {
	return cards.Card{
		ID: uuid.New(), Name: "Elbrus, the Binding Blade // Withengar Unbound", Layout: "transform",
		TypeLine: "Legendary Artifact — Equipment // Legendary Creature — Demon", OracleID: uuid.MustParse(oldB442Elbrus),
		CardFaces: []cards.CardFace{
			{Name: "Elbrus, the Binding Blade", TypeLine: "Legendary Artifact — Equipment", ManaCost: "{7}",
				OracleText: "Equipped creature gets +1/+0.\nWhen equipped creature deals combat damage to a player, unattach Elbrus, then transform it.\nEquip {1}"},
			{Name: "Withengar Unbound", TypeLine: "Legendary Creature — Demon", Colors: []string{"B"}, Power: "13", Toughness: "13",
				OracleText: "Flying, intimidate, trample\nWhenever a player loses the game, put thirteen +1/+1 counters on Withengar Unbound."},
		},
	}
}

func TestOldB442ElbrusUnattachesAndTransformsOnCombatDamageThenGrowsWhenAPlayerLoses(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[1]
	id := effectstest.ImportToHand(oldB442ElbrusRow(), me)
	c, _ := effectstest.CardInZone(me.Hand, id)
	me.Hand.Remove(id)
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.AttachForTest(t, g, id, host)
	if p, _ := effectstest.EffectivePTOf(t, g, host); p != 3 {
		t.Fatalf("equipped power %d, want 3", p)
	}
	effectstest.DealCombatDamageToPlayer(g, host, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatal("Elbrus did not transform")
	}
	if effectstest.AttachedTo(t, g, id, host) {
		t.Error("it should have unattached")
	}
	g.WithWriteLock(func() { g.Seats[2].Life = 0 })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, id); got != 13 {
		t.Errorf("%d counters after a player lost, want 13", got)
	}
}

func oldB442DemonRow() cards.Card {
	return cards.Card{
		ID: uuid.New(), Name: "Ravenous Demon // Archdemon of Greed", Layout: "transform",
		TypeLine: "Creature — Demon // Creature — Demon", OracleID: uuid.MustParse(oldB442Demon),
		CardFaces: []cards.CardFace{
			{Name: "Ravenous Demon", TypeLine: "Creature — Demon", ManaCost: "{3}{B}{B}", Colors: []string{"B"}, Power: "4", Toughness: "4",
				OracleText: "Sacrifice a Human: Transform this creature. Activate only as a sorcery."},
			{Name: "Archdemon of Greed", TypeLine: "Creature — Demon", Colors: []string{"B"}, Power: "9", Toughness: "9",
				OracleText: "Flying, trample\nAt the beginning of your upkeep, sacrifice a Human. If you can't, tap this creature and it deals 9 damage to you."},
		},
	}
}

func TestOldB442RavenousDemonTransformsBySacrificingAHumanAndTheArchdemonHungers(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.ImportToHand(oldB442DemonRow(), me)
	c, _ := effectstest.CardInZone(me.Hand, id)
	me.Hand.Remove(id)
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("transformed with no Human to sacrifice")
	}
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human", 1, 1)
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{human}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(human) || effectstest.IwFace(g, id) != 1 {
		t.Fatalf("human still there %v, face %d", g.Battlefield.Contains(human), effectstest.IwFace(g, id))
	}
	// Upkeep with no Human: it taps and deals 9 damage to you.
	life := me.Life
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-9 {
		t.Errorf("life %d → %d, want -9", life, me.Life)
	}
	if !effectstest.HmlTapped(t, g, id) {
		t.Error("the Archdemon should have tapped")
	}
	// With a Human: it is sacrificed instead and nothing else happens.
	life = me.Life
	h2 := effectstest.B12Creature(g, me.ID, "Human Two", "Creature — Human", 1, 1)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	oldB442AnswerSacrifice(t, g, me.ID, h2)
	if g.Battlefield.Contains(h2) || me.Life != life {
		t.Errorf("Human on table %v, life %d → %d: want it sacrificed and no damage", g.Battlefield.Contains(h2), life, me.Life)
	}
}

// oldB442AnswerSacrifice answers the sacrifice prompt when one is asked; a
// lone Human is taken without asking.
func oldB442AnswerSacrifice(t *testing.T, g *game.Game, who, pick uuid.UUID) {
	t.Helper()
	effectstest.AnswerSacrifice(t, g, who, pick)
	effectstest.PassPriorityAroundTable(t, g)
}

func oldB442ChosenRow() cards.Card {
	return cards.Card{
		ID: uuid.New(), Name: "Chosen of Markov // Markov's Servant", Layout: "transform",
		TypeLine: "Creature — Human // Creature — Vampire", OracleID: uuid.MustParse(oldB442Chosen),
		CardFaces: []cards.CardFace{
			{Name: "Chosen of Markov", TypeLine: "Creature — Human", ManaCost: "{2}{B}", Colors: []string{"B"}, Power: "2", Toughness: "2",
				OracleText: "{T}, Tap an untapped Vampire you control: Transform this creature."},
			{Name: "Markov's Servant", TypeLine: "Creature — Vampire", Colors: []string{"B"}, Power: "4", Toughness: "4"},
		},
	}
}

func TestOldB442ChosenOfMarkovNeedsAnUntappedVampire(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.ImportToHand(oldB442ChosenRow(), me)
	c, _ := effectstest.CardInZone(me.Hand, id)
	me.Hand.Remove(id)
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("transformed with no Vampire to tap")
	}
	vamp := effectstest.B12Creature(g, me.ID, "Vampire", "Creature — Vampire", 2, 2)
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{vamp}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatal("did not transform")
	}
	if !effectstest.HmlTapped(t, g, vamp) {
		t.Error("the Vampire should be tapped for the cost")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 4 || tough != 4 {
		t.Errorf("Markov's Servant is %d/%d, want 4/4", p, tough)
	}
}
