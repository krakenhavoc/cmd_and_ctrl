package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 MID batch 3 (#5497): the thirteen plain flashback spells.

const (
	midB03Shadowbeast = "f05d5632-0a82-45bf-b66e-22cf4080ed5e"
	midB03Ants        = "d9d33dcd-3352-478e-afc4-40f2ff369c39"
	midB03Dryad       = "8a69a0f0-5f0f-4098-ab6d-be3233af5cd3"
	midB03Homestead   = "203eb24c-f135-44e2-bdae-80b7b3d0fb15"
	midB03Angelfire   = "29e1c6e3-d76e-4388-b4e9-f732fdec4338"
	midB03Abandon     = "5404c6a5-399d-4237-b03c-41b6a1b79050"
	midB03Sacred      = "812bf52c-37f5-41c9-8a69-009fddbafb3f"
	midB03Path        = "e54d4ecf-ebcc-4374-98e7-c682f83c82a9"
	midB03Tapping     = "d5b80d04-1650-4b5d-9f9b-091ecd17b6e8"
	midB03Arcane      = "d5251d47-11de-41f7-9965-1f921e12d927"
	midB03Hungry      = "3eaf1e6b-bf4c-4ea8-83ea-6827470fd802"
	midB03Winterthorn = "4b4bdbdc-3934-471b-9027-31277e7765ab"
	midB03Secrets     = "55c7ce94-3cd3-42f3-9dd8-0c118fa626c8"
)

func TestMidB03CardsAreRegisteredFullWithAGraveyardFlashback(t *testing.T) {
	for oracle, name := range map[string]string{
		midB03Shadowbeast: "Shadowbeast Sighting", midB03Ants: "Rise of the Ants",
		midB03Dryad: "Dryad's Revival", midB03Homestead: "Homestead Courage",
		midB03Angelfire: "Angelfire Ignition", midB03Abandon: "Abandon the Post",
		midB03Sacred: "Sacred Fire", midB03Path: "Path to the Festival",
		midB03Tapping: "Tapping at the Window", midB03Arcane: "Arcane Infusion",
		midB03Hungry: "Hungry for More", midB03Winterthorn: "Winterthorn Blessing",
		midB03Secrets: "Secrets of the Key",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
		if len(spec.CastableZones) != 1 || spec.CastableZones[0] != game.ZoneGraveyard ||
			len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Key != "flashback" {
			t.Errorf("%s: not a graveyard flashback card", name)
		}
	}
}

// midB03Flash casts the card in the active seat's graveyard for its
// flashback cost, strictly checking nothing about mana (the catalog game
// lets a non-strict cast through), and returns the card's ID.
func midB03Flash(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedGraveyardCard(t, g, name, typeLine, oracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Targets: targets,
	}); err != nil {
		t.Fatalf("flashback cast of %s: %v", name, err)
	}
	return id
}

func midB03Exiled(t *testing.T, g *game.Game, id uuid.UUID, name string) {
	t.Helper()
	if !g.Exile.Contains(id) || g.Seats[g.Turn.ActiveSeat].Graveyard.Contains(id) {
		t.Errorf("the flashed-back %s should be exiled, not back in the graveyard", name)
	}
}

func midB03Tokens(g *game.Game, controller uuid.UUID, subtype string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.HasSubtype(subtype) {
			out = append(out, c)
		}
	}
	return out
}

func TestMidB03ShadowbeastSightingMakesAFourFourBeastTwice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Shadowbeast Sighting", "Sorcery", midB03Shadowbeast, nil)
	passPriorityAroundTable(t, g)
	beasts := midB03Tokens(g, me.ID, "Beast")
	if len(beasts) != 1 || beasts[0].Power != 4 || beasts[0].Toughness != 4 {
		t.Fatalf("beasts = %+v, want one 4/4", beasts)
	}
	id := midB03Flash(t, g, "Shadowbeast Sighting", "Sorcery", midB03Shadowbeast, nil)
	passPriorityAroundTable(t, g)
	if n := len(midB03Tokens(g, me.ID, "Beast")); n != 2 {
		t.Errorf("%d Beasts after the flashback, want 2", n)
	}
	midB03Exiled(t, g, id, "Shadowbeast Sighting")
}

func TestMidB03RiseOfTheAntsMakesTwoInsectsAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	castCatalogSpell(t, g, "Rise of the Ants", "Sorcery", midB03Ants, nil)
	passPriorityAroundTable(t, g)
	ins := midB03Tokens(g, me.ID, "Insect")
	if len(ins) != 2 || ins[0].Power != 3 || ins[0].Toughness != 3 {
		t.Fatalf("insects = %+v, want two 3/3", ins)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	id := midB03Flash(t, g, "Rise of the Ants", "Sorcery", midB03Ants, nil)
	passPriorityAroundTable(t, g)
	if n := len(midB03Tokens(g, me.ID, "Insect")); n != 4 {
		t.Errorf("%d Insects after the flashback, want 4", n)
	}
	midB03Exiled(t, g, id, "Rise of the Ants")
}

func TestMidB03DryadsRevivalReturnsACardAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := seedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	castCatalogSpell(t, g, "Dryad's Revival", "Sorcery", midB03Dryad, b16TargetCard(a))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) {
		t.Fatal("the targeted card did not return to hand")
	}
	b := seedGraveyardCard(t, g, "Wolf", "Creature — Wolf", "")
	id := midB03Flash(t, g, "Dryad's Revival", "Sorcery", midB03Dryad, b16TargetCard(b))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(b) {
		t.Error("the flashback did not return its target")
	}
	midB03Exiled(t, g, id, "Dryad's Revival")
}

func TestMidB03DryadsRevivalRefusesAnOpponentsGraveyardCard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Theirs", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	if err := castCatalogSpellErrStrict(t, g, "Dryad's Revival", "Sorcery", midB03Dryad, b16TargetCard(theirs)); err == nil {
		t.Error("Dryad's Revival targeted a card in an opponent's graveyard")
	}
}

func TestMidB03HomesteadCourageCountersAndVigilanceThenFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Homestead Courage", "Sorcery", midB03Homestead, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if n := counterOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("bear has %d +1/+1 counters, want 1", n)
	}
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the bear did not gain vigilance")
	}
	id := midB03Flash(t, g, "Homestead Courage", "Sorcery", midB03Homestead, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if n := counterOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("bear has %d +1/+1 counters after the flashback, want 2", n)
	}
	midB03Exiled(t, g, id, "Homestead Courage")
}

func TestMidB03HomesteadCourageRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	foe := pushVanillaCreature(g, g.Seats[1].ID, "Foe", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Homestead Courage", "Sorcery", midB03Homestead, b16TargetCard(foe)); err == nil {
		t.Error("Homestead Courage targeted a creature its caster does not control")
	}
}

func TestMidB03AngelfireIgnitionTwoCountersAndFiveKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castCatalogSpell(t, g, "Angelfire Ignition", "Sorcery", midB03Angelfire, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if n := counterOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("bear has %d +1/+1 counters, want 2", n)
	}
	for _, kw := range []string{"vigilance", "trample", "lifelink", "indestructible", "haste"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the bear lacks %s", kw)
		}
	}
	id := midB03Flash(t, g, "Angelfire Ignition", "Sorcery", midB03Angelfire, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if n := counterOn(g, bear, game.CounterPlusOne); n != 4 {
		t.Errorf("bear has %d +1/+1 counters after the flashback, want 4", n)
	}
	midB03Exiled(t, g, id, "Angelfire Ignition")
}

func TestMidB03AbandonThePostStopsUpToTwoCreaturesBlocking(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := pushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "B", 2, 2)
	c := pushVanillaCreature(g, opp.ID, "C", 2, 2)
	castCatalogSpell(t, g, "Abandon the Post", "Sorcery", midB03Abandon,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !midB03CantBlock(t, g, id) {
			t.Errorf("%s can still block", id)
		}
	}
	if midB03CantBlock(t, g, c) {
		t.Error("the untargeted creature can't block")
	}
}

func TestMidB03AbandonThePostFlashbackAndRefusesThreeTargets(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := pushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "B", 2, 2)
	c := pushVanillaCreature(g, opp.ID, "C", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Abandon the Post", "Sorcery", midB03Abandon,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}); err == nil {
		t.Error("Abandon the Post took three targets")
	}
	g2 := newCatalogGame(t)
	foe := pushVanillaCreature(g2, g2.Seats[1].ID, "A", 2, 2)
	id := midB03Flash(t, g2, "Abandon the Post", "Sorcery", midB03Abandon, b16TargetCard(foe))
	passPriorityAroundTable(t, g2)
	if !midB03CantBlock(t, g2, foe) {
		t.Error("the flashback did not stop the creature blocking")
	}
	midB03Exiled(t, g2, id, "Abandon the Post")
}

func midB03CantBlock(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return auraRestrictions(t, g, id)&game.CantBlock != 0
}

func TestMidB03SacredFireDamagesAndGainsLifeThenFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, oppLife := me.Life, opp.Life
	castCatalogSpell(t, g, "Sacred Fire", "Instant", midB03Sacred, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 || me.Life != life+2 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, life+2, oppLife-2)
	}
	id := midB03Flash(t, g, "Sacred Fire", "Instant", midB03Sacred, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-4 || me.Life != life+4 {
		t.Errorf("after the flashback life %d/%d, want %d/%d", me.Life, opp.Life, life+4, oppLife-4)
	}
	midB03Exiled(t, g, id, "Sacred Fire")
}

func TestMidB03SacredFireWithNoLegalTargetDoesNotGainLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	foe := pushVanillaCreature(g, g.Seats[1].ID, "Foe", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Sacred Fire", "Instant", midB03Sacred, nil); err == nil {
		t.Error("Sacred Fire was cast with no target")
	}
	if err := castCatalogSpellErrStrict(t, g, "Sacred Fire", "Instant", midB03Sacred, b16TargetCard(foe)); err != nil {
		t.Errorf("a creature is a legal target: %v", err)
	}
	if me.Life != life {
		t.Errorf("life moved to %d before anything resolved", me.Life)
	}
}

func TestMidB03SecretsOfTheKeyInvestigatesOnceThenTwiceFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Secrets of the Key", "Instant", midB03Secrets, nil)
	passPriorityAroundTable(t, g)
	if n := countTokensNamed2(g, me.ID, "Clue"); n != 1 {
		t.Fatalf("%d Clues from the hand cast, want 1", n)
	}
	id := midB03Flash(t, g, "Secrets of the Key", "Instant", midB03Secrets, nil)
	passPriorityAroundTable(t, g)
	if n := countTokensNamed2(g, me.ID, "Clue"); n != 3 {
		t.Errorf("%d Clues after the flashback, want 3 (1 + 2, not 1 + 3)", n)
	}
	midB03Exiled(t, g, id, "Secrets of the Key")
}

func countTokensNamed2(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.Name == name {
			n++
		}
	}
	return n
}

func TestMidB03HungryForMoreMakesAHastyVampireThatIsSacrificedAtEnd(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Hungry for More", "Sorcery", midB03Hungry, nil)
	passPriorityAroundTable(t, g)
	vamps := midB03Tokens(g, me.ID, "Vampire")
	if len(vamps) != 1 || vamps[0].Power != 3 || vamps[0].Toughness != 1 {
		t.Fatalf("vampires = %+v, want one 3/1", vamps)
	}
	for _, kw := range []string{"trample", "lifelink", "haste"} {
		if !hasEffectiveKeyword(t, g, vamps[0].InstanceID, kw) {
			t.Errorf("the Vampire lacks %s", kw)
		}
	}
	advanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := len(midB03Tokens(g, me.ID, "Vampire")); n != 0 {
		t.Errorf("%d Vampires survived the end step, want 0", n)
	}
}

func TestMidB03HungryForMoreFlashbackMakesAnotherVampire(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := midB03Flash(t, g, "Hungry for More", "Sorcery", midB03Hungry, nil)
	passPriorityAroundTable(t, g)
	if n := len(midB03Tokens(g, me.ID, "Vampire")); n != 1 {
		t.Fatalf("%d Vampires after the flashback, want 1", n)
	}
	midB03Exiled(t, g, id, "Hungry for More")
	advanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := len(midB03Tokens(g, me.ID, "Vampire")); n != 0 {
		t.Errorf("%d Vampires survived the end step, want 0", n)
	}
}

func TestMidB03WinterthornBlessingCountersOneAndFreezesAnother(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	foe := pushVanillaCreature(g, opp.ID, "Foe", 2, 2)
	castCatalogSpell(t, g, "Winterthorn Blessing", "Sorcery", midB03Winterthorn,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: foe}})
	passPriorityAroundTable(t, g)
	if n := counterOn(g, mine, game.CounterPlusOne); n != 1 {
		t.Errorf("mine has %d +1/+1 counters, want 1", n)
	}
	c, _ := g.LookupCardForEffect(foe)
	if !c.Tapped {
		t.Error("the opposing creature is not tapped")
	}
	if len(hmlUntapSkips(t, g, foe)) == 0 {
		t.Error("the opposing creature has no next-untap-step mark")
	}
	if n := counterOn(g, foe, game.CounterPlusOne); n != 0 {
		t.Errorf("foe has %d +1/+1 counters, want none", n)
	}
}

func TestMidB03WinterthornBlessingTargetsAreEachOptionalAndBoundByControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	foe := pushVanillaCreature(g, opp.ID, "Foe", 2, 2)
	advanceToMain(t, g)
	// Neither target: legal.
	if err := castCatalogSpellErrStrict(t, g, "Winterthorn Blessing", "Sorcery", midB03Winterthorn, nil); err != nil {
		t.Errorf("no targets: %v", err)
	}
	// The counter on an opponent's creature is refused; so is the freeze on mine.
	if err := castCatalogSpellErrStrict(t, g, "Winterthorn Blessing", "Sorcery", midB03Winterthorn,
		[]game.TargetRef{{Kind: game.TargetCard, ID: foe}}); err == nil {
		t.Error("the +1/+1 counter clause accepted a creature its caster does not control")
	}
	if err := castCatalogSpellErrStrict(t, g, "Winterthorn Blessing", "Sorcery", midB03Winterthorn,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: mine}}); err == nil {
		t.Error("the tap clause accepted a creature its caster controls")
	}
}

func TestMidB03WinterthornBlessingFlashbackWithOnlyTheFreeze(t *testing.T) {
	g := newCatalogGame(t)
	foe := pushVanillaCreature(g, g.Seats[1].ID, "Foe", 2, 2)
	id := midB03Flash(t, g, "Winterthorn Blessing", "Sorcery", midB03Winterthorn,
		[]game.TargetRef{{Kind: game.TargetCard, ID: foe}})
	passPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(foe)
	if !c.Tapped {
		t.Error("the flashback did not tap the creature")
	}
	midB03Exiled(t, g, id, "Winterthorn Blessing")
}

func TestMidB03PathToTheFestivalFetchesATappedBasicAndScriesOnThreeTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	domainLands(g, me, opp, "Plains", "Island")
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Path to the Festival", "Sorcery", midB03Path, nil)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	c, ok := g.LookupCardForEffect(forest)
	if !ok || !g.Battlefield.Contains(forest) || !c.Tapped {
		t.Fatal("the Forest is not on the battlefield tapped")
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("three basic land types (Plains, Island, Forest) but no scry 1")
	}
}

func TestMidB03PathToTheFestivalDoesNotScryOnTwoTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	domainLands(g, me, opp, "Plains")
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Path to the Festival", "Sorcery", midB03Path, nil)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the Forest was not fetched")
	}
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("two basic land types, but the spell scried")
	}
}

func TestMidB03PathToTheFestivalCountsTheFetchedLandAndFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	domainLands(g, me, opp, "Plains", "Island", "Swamp")
	id := midB03Flash(t, g, "Path to the Festival", "Sorcery", midB03Path, nil)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("three types already, the failed search still scries")
	}
	if err := g.ResolveScry(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityAroundTable(t, g)
	midB03Exiled(t, g, id, "Path to the Festival")
}

func TestMidB03TappingAtTheWindowTakesACreatureAndMillsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	var top [3]uuid.UUID
	cards := []struct{ name, typeLine string }{
		{"Forest", "Basic Land — Forest"}, {"Bear", "Creature — Bear"}, {"Shock", "Instant"},
	}
	for i := len(cards) - 1; i >= 0; i-- {
		top[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top[i], Name: cards[i].name, TypeLine: cards[i].typeLine, Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Tapping at the Window", "Sorcery", midB03Tapping, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != top[1] || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want only the creature, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[1]) {
		t.Error("the creature card is not in hand")
	}
	if !me.Graveyard.Contains(top[0]) || !me.Graveyard.Contains(top[2]) {
		t.Error("the rest did not go to the graveyard")
	}
}

func TestMidB03TappingAtTheWindowMayTakeNothingAndStillMills(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	ids := seedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Tapping at the Window", "Sorcery", midB03Tapping, nil)
	passPriorityAroundTable(t, g)
	if pick := latestChooseCards(g, me.ID); pick != nil {
		if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("declining the pick: %v", err)
		}
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand grew from %d to %d", hand, me.Hand.Size())
	}
	for _, id := range ids {
		if !me.Graveyard.Contains(id) {
			t.Error("a card of the three did not reach the graveyard")
		}
	}
}

func TestMidB03ArcaneInfusionTakesAnInstantOrSorceryAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	var top [4]uuid.UUID
	cards := []struct{ name, typeLine string }{
		{"Bear", "Creature — Bear"}, {"Shock", "Instant"}, {"Divination", "Sorcery"}, {"Forest", "Basic Land — Forest"},
	}
	for i := len(cards) - 1; i >= 0; i-- {
		top[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top[i], Name: cards[i].name, TypeLine: cards[i].typeLine, Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Arcane Infusion", "Instant", midB03Arcane, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[top[1]] || !offered[top[2]] || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the instant and the sorcery, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[2]) {
		t.Error("the sorcery is not in hand")
	}
	for _, id := range []uuid.UUID{top[0], top[1], top[3]} {
		if me.Hand.Contains(id) || me.Graveyard.Contains(id) {
			t.Error("a card that was not taken left the library other than to the bottom")
		}
	}
}
