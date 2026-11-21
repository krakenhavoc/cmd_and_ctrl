package testsg

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Every keyword in this file declares a graveyard ability and none of
// them declares a component that needs a permanent — the boot-time
// invariant, asserted here so a future card file that hand-rolls one
// of these fails in this package rather than at server start.
func TestGraveyardKeywordsDeclareAGraveyardAbility(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, mana string
		exiles             bool
	}{
		{"Dregscape Zombie", effectstest.DregscapeZombieOracle, "{B}", false},
		{"Deadbridge Goliath", effectstest.DeadbridgeGoliathOrac, "{4}{G}{G}", true},
		{"Sacred Cat", effectstest.SacredCatOracle, "{W}", true},
	} {
		abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle})
		if len(abilities) != 1 {
			t.Errorf("%s: %d activated abilities, want the one graveyard ability", tc.name, len(abilities))
			continue
		}
		ab := abilities[0]
		if ab.Cost.Mana != tc.mana {
			t.Errorf("%s: cost %q, want %q", tc.name, ab.Cost.Mana, tc.mana)
		}
		if ab.Cost.ExileSelf != tc.exiles {
			t.Errorf("%s: ExileSelf %v, want %v", tc.name, ab.Cost.ExileSelf, tc.exiles)
		}
		if !ab.SorcerySpeed {
			t.Errorf("%s: not sorcery-speed — every keyword here prints \"only as a sorcery\"", tc.name)
		}
		if !game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
			t.Errorf("%s: the ability does not function from a graveyard", tc.name)
		}
		if game.AbilityFunctionsFromZone(ab, game.ZoneBattlefield) {
			t.Errorf("%s: the ability is offered on the battlefield too", tc.name)
		}
		if game.AbilityFunctionsFromZone(ab, game.ZoneHand) {
			t.Errorf("%s: the ability is offered from hand too", tc.name)
		}
		if why := game.AbilityNeedsPermanentSource(ab.Cost); why != "" {
			t.Errorf("%s: declares %s, which needs a permanent", tc.name, why)
		}
	}
}

// Unearth returns the card to the battlefield with haste.
func TestDregscapeZombieUnearthsWithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.ActivateFromGraveyard(t, g, "Dregscape Zombie", "Creature — Zombie",
		effectstest.DregscapeZombieOracle, 2, 1, "{B}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the unearthed Zombie is not on the battlefield")
	}
	if me.Graveyard.Contains(id) {
		t.Error("the unearthed Zombie is still in the graveyard")
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatal("the unearthed Zombie cannot be looked up")
	}
	if !game.HasKeyword(&c, "haste") {
		t.Errorf("the unearthed Zombie has no haste (keywords %v)", c.Effective().Abilities)
	}
}

// CR 702.84a's second sentence: the unearthed creature is exiled at
// the beginning of the next end step, not put back in the graveyard
// for a second unearth.
func TestUnearthedCreatureIsExiledAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.ActivateFromGraveyard(t, g, "Dregscape Zombie", "Creature — Zombie",
		effectstest.DregscapeZombieOracle, 2, 1, "{B}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; g.Turn.Step != game.StepEnd && i < 12; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("the unearthed Zombie survived its own end step")
	}
	if me.Graveyard.Contains(id) {
		t.Error("the unearthed Zombie went to the graveyard — it must be exiled")
	}
	if !g.Exile.Contains(id) {
		t.Error("the unearthed Zombie is not in exile")
	}
}

// And the clause that keeps the loop closed: a destroyed unearthed
// creature is EXILED instead of hitting the graveyard, so it cannot
// be unearthed twice.
func TestUnearthedCreatureDestroyedGoesToExileNotTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.ActivateFromGraveyard(t, g, "Dregscape Zombie", "Creature — Zombie",
		effectstest.DregscapeZombieOracle, 2, 1, "{B}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(id) {
		t.Error("a destroyed unearthed creature reached the graveyard")
	}
	if !g.Exile.Contains(id) {
		t.Error("a destroyed unearthed creature is not in exile")
	}
}

// Scavenge: the cost exiles the card, and the effect reads the power
// it had — five counters off a 5/5 — onto the target.
func TestDeadbridgeGoliathScavengesFiveCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.PushBattlefieldCreature(g, me.ID)
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Deadbridge Goliath", "Creature — Insect",
		effectstest.DeadbridgeGoliathOrac, 5, 5, "{C}{C}{C}{C}{G}{G}", game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
	if !g.Exile.Contains(id) {
		t.Fatal("the scavenged card is not in exile — the cost exiles it at announce")
	}
	if me.Graveyard.Contains(id) {
		t.Fatal("the scavenged card is still in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(target)
	if !ok {
		t.Fatal("the scavenge target vanished")
	}
	if got := c.Counters[game.CounterPlusOne]; got != 5 {
		t.Errorf("%d +1/+1 counters, want 5 (the Goliath's power)", got)
	}
}

// Embalm: the cost exiles the card and the effect creates a token
// copy that is a WHITE ZOMBIE CAT with no mana cost — CR 707.9b's
// "in addition to its other types", which is what separates an
// embalmed Sacred Cat from a plain Zombie.
func TestSacredCatEmbalmsIntoAWhiteZombieCat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.ActivateFromGraveyard(t, g, "Sacred Cat", "Creature — Cat",
		effectstest.SacredCatOracle, 1, 1, "{W}", game.ActivateAbilityParams{})
	if !g.Exile.Contains(id) {
		t.Fatal("the embalmed card is not in exile — the cost exiles it at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	var token *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Sacred Cat" && c.Controller == me.ID {
			token = c
			break
		}
	}
	if token == nil {
		t.Fatal("no embalm token was created")
	}
	if token.InstanceID == id {
		t.Fatal("the exiled card itself came back — embalm creates a COPY")
	}
	if token.ManaCost != "" {
		t.Errorf("token mana cost %q, want none", token.ManaCost)
	}
	if len(token.Colors) != 1 || token.Colors[0] != "W" {
		t.Errorf("token colours %v, want [W]", token.Colors)
	}
	_, _, subs := game.ParseTypeLine(token.TypeLine)
	var zombie, cat bool
	for _, s := range subs {
		switch s {
		case "Zombie":
			zombie = true
		case "Cat":
			cat = true
		}
	}
	if !zombie || !cat {
		t.Errorf("token type line %q, want a Zombie Cat", token.TypeLine)
	}
	if token.Power != 1 || token.Toughness != 1 {
		t.Errorf("token %d/%d, want 1/1 — embalm does not resize", token.Power, token.Toughness)
	}
}

// Eternalize differs from embalm in its "except" clause and in
// nothing else: a 4/4 BLACK Zombie that keeps its other types. No
// card in the catalog prints it yet, so the clause is pinned
// directly — the alternative is a keyword nothing checks until the
// first eternalize card ships and ships it wrong.
func TestEternalizeExceptionIsAFourFourBlackZombie(t *testing.T) {
	tok := game.Card{
		Name: "Test", TypeLine: "Creature — Human Wizard",
		Power: 2, Toughness: 3, ManaCost: "{1}{U}", Colors: []string{"U"},
	}
	effects.EternalizeException(&tok)
	if tok.Power != 4 || tok.Toughness != 4 {
		t.Errorf("%d/%d, want 4/4", tok.Power, tok.Toughness)
	}
	if len(tok.Colors) != 1 || tok.Colors[0] != "B" {
		t.Errorf("colours %v, want [B]", tok.Colors)
	}
	if tok.ManaCost != "" {
		t.Errorf("mana cost %q, want none", tok.ManaCost)
	}
	if tok.TypeLine != "Creature — Zombie Human Wizard" {
		t.Errorf("type line %q, want the Zombie added to the printed types", tok.TypeLine)
	}
}

// addedSubtypeTypeLine is CR 707.9b's ADD form and retypedTypeLine is
// CR 707.9a's SET form; picking the wrong one is the silent failure
// this pair of assertions exists to catch.
func TestAddedSubtypeKeepsThePrintedTypesAndRetypeDoesNot(t *testing.T) {
	const printed = "Legendary Creature — Human Wizard"
	if got := effects.AddedSubtypeTypeLine(printed, "Zombie"); got != "Legendary Creature — Zombie Human Wizard" {
		t.Errorf("added: %q", got)
	}
	if got := effects.RetypedTypeLine(printed, "Zombie"); got != "Legendary Creature — Zombie" {
		t.Errorf("retyped: %q", got)
	}
	// Idempotent: a card that already has the subtype is untouched.
	if got := effects.AddedSubtypeTypeLine("Creature — Zombie", "Zombie"); got != "Creature — Zombie" {
		t.Errorf("duplicate subtype: %q", got)
	}
}
