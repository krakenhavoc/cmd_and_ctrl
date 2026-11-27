package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	b720VeiledSentry       = "0b9b8e60-51eb-4d4a-95c1-48ca5b8153df"
	b720HiddenPredators    = "3b143227-f373-4403-b246-ee1c9a0dc85a"
	b720VeiledApparition   = "4187fd6d-5660-4ea7-a9dd-402487ec8721"
	b720HiddenGuerrillas   = "4337f35b-baab-4640-9e1c-05f61c49bbdd"
	b720HiddenStag         = "70c32155-dc07-45cb-bceb-e19d3ebc915d"
	b720OpalArchangel      = "888d12d4-aaa5-406f-9dc8-8c4e856eb2fe"
	b720HiddenHerd         = "8e7c5649-41c2-4a0c-b873-e6f6c5908567"
	b720DefensiveFormation = "9386f4d3-aa72-4849-8921-83740762dc0e"
	b720PlanarVoid         = "994e27e1-0bff-47c9-a1e6-b9e3ae4ffb1e"
)

var b720Colour = map[string]string{
	b720VeiledSentry: "U", b720HiddenPredators: "G", b720VeiledApparition: "U", b720HiddenGuerrillas: "G",
	b720HiddenStag: "G", b720OpalArchangel: "W", b720HiddenHerd: "G", b720DefensiveFormation: "W", b720PlanarVoid: "B",
}

func b720Enchantment(g *game.Game, who *game.Player, name, oracle string) uuid.UUID {
	return effectstest.ApaPush(g, who.ID, who.ID, game.Card{Name: name, OracleID: oracle, TypeLine: "Enchantment", Colors: []string{b720Colour[oracle]}})
}

func b720WantCreature(t *testing.T, g *game.Game, id uuid.UUID, sub string, p, tough int) *game.Card {
	t.Helper()
	c := effectstest.ApaLive(g, id)
	if c == nil {
		t.Fatal("the permanent is gone")
	}
	if !c.IsCreature() || c.HasCardType("enchantment") || !c.HasSubtype(sub) || c.CurrentPower() != p || c.CurrentToughness() != tough {
		t.Fatalf("creature %v enchantment %v subtype %v %d/%d; want a %d/%d %s creature that is not an enchantment",
			c.IsCreature(), c.HasCardType("enchantment"), c.HasSubtype(sub), c.CurrentPower(), c.CurrentToughness(), p, tough, sub)
	}
	return c
}

func b720WantEnchantment(t *testing.T, g *game.Game, id uuid.UUID, why string) {
	t.Helper()
	if c := effectstest.ApaLive(g, id); c == nil || !c.HasCardType("enchantment") || c.IsCreature() {
		t.Fatalf("%s: it is not still an enchantment", why)
	}
}

// b720Cast casts a card from the active seat's hand at its main phase.
func b720Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	id := effectstest.PutInHand(p, c)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return id
}

func b720PlayLand(t *testing.T, g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	id := effectstest.PutInHand(p, c)
	effectstest.AdvanceToMain(t, g)
	p.LandDropsPerTurn++
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", c.Name, err)
	}
	return id
}

func b720AtUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == seat && i > 0 {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestB720VeiledSentryBecomesAnIllusionAsBigAsTheSpellsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	sentry := b720Enchantment(g, owner, "Veiled Sentry", b720VeiledSentry)
	b720Cast(t, g, caster, game.Card{Name: "Big Spell", TypeLine: "Instant", ManaCost: "{2}{R}", Colors: []string{"R"}})
	effectstest.PassPriorityAroundTable(t, g)
	c := b720WantCreature(t, g, sentry, "Illusion", 3, 3)
	if !c.HasColor("U") {
		t.Error("it is no longer blue")
	}
}

func TestB720VeiledSentryIgnoresItsControllersSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sentry := b720Enchantment(g, me, "Veiled Sentry", b720VeiledSentry)
	b720Cast(t, g, me, game.Card{Name: "Spell", TypeLine: "Instant", ManaCost: "{1}"})
	effectstest.PassPriorityAroundTable(t, g)
	b720WantEnchantment(t, g, sentry, "its controller's spell")
}

func TestB720HiddenPredatorsNeedsAnOpponentsCreatureWithPowerFourOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	preds := b720Enchantment(g, me, "Hidden Predators", b720HiddenPredators)
	effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	b720WantEnchantment(t, g, preds, "my own big creature and their 3-power creature")
	effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	b720WantCreature(t, g, preds, "Beast", 4, 4)
}

func TestB720VeiledApparitionBecomesAFlyingIllusionWithAnUpkeepTax(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	app := b720Enchantment(g, owner, "Veiled Apparition", b720VeiledApparition)
	b720Cast(t, g, caster, game.Card{Name: "Spell", TypeLine: "Instant", ManaCost: "{1}"})
	effectstest.PassPriorityAroundTable(t, g)
	c := b720WantCreature(t, g, app, "Illusion", 3, 3)
	if !game.HasKeyword(c, "flying") {
		t.Fatal("it has no flying")
	}
	b720AtUpkeepOf(t, g, 1)
	choice := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, owner.ID)
	if choice == nil {
		t.Fatal("its controller was not asked to pay {1}{U} at their upkeep")
	}
	if err := g.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, app) {
		t.Error("the creature was not sacrificed")
	}
}

func TestB720HiddenGuerrillasWaitsForAnArtifactSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	guer := b720Enchantment(g, owner, "Hidden Guerrillas", b720HiddenGuerrillas)
	b720Cast(t, g, caster, game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	effectstest.PassPriorityAroundTable(t, g)
	b720WantEnchantment(t, g, guer, "a noncreature nonartifact spell")
	b720Cast(t, g, caster, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	effectstest.PassPriorityAroundTable(t, g)
	c := b720WantCreature(t, g, guer, "Soldier", 5, 3)
	if !game.HasKeyword(c, "trample") {
		t.Error("it has no trample")
	}
}

func TestB720OpalArchangelBecomesAFlyingVigilantAngelOnACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	angel := b720Enchantment(g, owner, "Opal Archangel", b720OpalArchangel)
	b720Cast(t, g, caster, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	c := b720WantCreature(t, g, angel, "Angel", 5, 5)
	if !game.HasKeyword(c, "flying") || !game.HasKeyword(c, "vigilance") {
		t.Error("it lacks flying or vigilance")
	}
}

func TestB720HiddenHerdWaitsForAnOpponentsNonbasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	player, owner := g.Seats[0], g.Seats[1]
	herd := b720Enchantment(g, owner, "Hidden Herd", b720HiddenHerd)
	b720PlayLand(t, g, player, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PassPriorityAroundTable(t, g)
	b720WantEnchantment(t, g, herd, "a basic land")
	b720PlayLand(t, g, player, game.Card{Name: "Strip Mine", TypeLine: "Land"})
	effectstest.PassPriorityAroundTable(t, g)
	b720WantCreature(t, g, herd, "Beast", 3, 3)
}

func TestB720HiddenHerdIgnoresItsControllersLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herd := b720Enchantment(g, me, "Hidden Herd", b720HiddenHerd)
	b720PlayLand(t, g, me, game.Card{Name: "Strip Mine", TypeLine: "Land"})
	effectstest.PassPriorityAroundTable(t, g)
	b720WantEnchantment(t, g, herd, "its controller's nonbasic land")
}

func TestB720HiddenStagBecomesACreatureAndBackAgain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	stag := b720Enchantment(g, me, "Hidden Stag", b720HiddenStag)
	// My own land while it is an enchantment: nothing.
	b720PlayLand(t, g, me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PassPriorityAroundTable(t, g)
	b720WantEnchantment(t, g, stag, "its controller's land")
	// An opponent's land turns it into a 3/2 Elk Beast.
	g2 := effectstest.NewCatalogGame(t)
	stag2 := b720Enchantment(g2, g2.Seats[1], "Hidden Stag", b720HiddenStag)
	b720PlayLand(t, g2, g2.Seats[0], game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PassPriorityAroundTable(t, g2)
	c := b720WantCreature(t, g2, stag2, "Elk", 3, 2)
	if !c.HasSubtype("Beast") {
		t.Error("it is not a Beast")
	}
	// Its controller's land turns the creature back into an enchantment with no
	// creature subtypes, on a later turn (one land drop each turn).
	for g2.Turn.ActiveSeat != 1 {
		if _, err := g2.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	b720PlayLand(t, g2, g2.Seats[1], game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PassPriorityAroundTable(t, g2)
	b720WantEnchantment(t, g2, stag2, "after its controller played a land")
	if back := effectstest.ApaLive(g2, stag2); back.HasSubtype("Elk") || back.HasSubtype("Beast") {
		t.Error("it kept its creature subtypes as an enchantment")
	}
}

func TestB720DefensiveFormationLetsTheDefenderDivideTheAttackersDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b720Enchantment(g, opp, "Defensive Formation", b720DefensiveFormation)
	beast := effectstest.B12Creature(g, me.ID, "Beast", "Creature — Beast", 4, 4)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, beast)
	if err := g.DeclareBlocker(a, beast); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareBlocker(b, beast); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	var choice *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDamageAssignment {
			choice = c
		}
	}
	if choice == nil {
		t.Fatal("no prompt for the attacker's damage")
	}
	if choice.Chooser != opp.ID {
		t.Fatalf("the prompt is for %v, want the defending player", choice.Chooser)
	}
	if err := g.ResolveDamageAssignment(choice.ID, me.ID, []game.DamageAssignmentEntry{{BlockerID: a, Amount: 2}, {BlockerID: b, Amount: 2}}, 0); err == nil {
		t.Fatal("the attacking player answered the defending player's division")
	}
	if err := g.ResolveDamageAssignment(choice.ID, opp.ID, []game.DamageAssignmentEntry{{BlockerID: a, Amount: 4}, {BlockerID: b, Amount: 0}}, 0); err != nil {
		t.Fatalf("the defender's division: %v", err)
	}
	if effects.OnBattlefield(g, a) {
		t.Error("Bear A took 4 and is still on the battlefield")
	}
	if got := effectstest.DamageOn(t, g, b); got != 0 {
		t.Errorf("Bear B has %d damage, want 0", got)
	}
}

func TestB720PlanarVoidExilesEveryOtherCardPutIntoAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	void := b720Enchantment(g, me, "Planar Void", b720PlanarVoid)
	// A creature dying.
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B18Kill(t, g, bear)
	if opp.Graveyard.Contains(bear) || !g.Exile.Contains(bear) {
		t.Error("the dead creature was not exiled")
	}
	// A discard, from a hand.
	discarded := effectstest.PutInHand(opp, game.Card{Name: "Discarded", TypeLine: "Sorcery"})
	g.WithWriteLock(func() {
		if err := g.DiscardCardsKeyedForEffect(opp.ID, []uuid.UUID{discarded}, uuid.Nil, game.KeyedTail{}); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(discarded) || !g.Exile.Contains(discarded) {
		t.Error("the discarded card was not exiled")
	}
	// A resolved instant.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Graveyard.Cards) != 0 {
		t.Errorf("the resolved Bolt sits in a graveyard: %d cards", len(me.Graveyard.Cards))
	}
	if !effects.OnBattlefield(g, void) {
		t.Error("Planar Void exiled itself")
	}
}

func TestB720PlanarVoidDoesNotExileItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	void := b720Enchantment(g, me, "Planar Void", b720PlanarVoid)
	effectstest.B18Kill(t, g, void)
	if !me.Graveyard.Contains(void) {
		t.Error("Planar Void did not stay in the graveyard it was put into")
	}
}
