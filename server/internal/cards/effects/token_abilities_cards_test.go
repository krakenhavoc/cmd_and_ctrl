package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_abilities_cards_test.go — ADR 0083 (#1248) at the card level.
//
// The engine seam is pinned hermetically in game/token_abilities_test.go.
// These are the cards the "Triggered and static abilities on non-copy
// tokens" row of docs/engine-seams.md was waiting on, each run through
// the real catalog: three that could not be written at all, and three
// that shipped with a caveat saying their token was a plain body.

const (
	reefWormOracle      = "e3ad5cbc-4245-4e1e-8204-509381fa0c1c"
	nestingDragonOracle = "0acc9372-58b4-43bf-ab82-1f95831c81d4"
	mysidianElderOracle = "10039992-d51a-47e7-9a70-02fe2227c163"
	beledrosOracle      = "90194ff1-db61-463f-b5a3-15cd85311d0e"
)

// destroyAndSettle destroys a permanent and lets the triggers it
// queued resolve, which is two priority rounds: one to put the
// dies-trigger on the stack and resolve it, and one for anything that
// trigger queued in turn.
func destroyAndSettle(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
}

// --- Reef Worm ----------------------------------------------------

// TestReefWormChainsThreeTokenDiesTriggers is the row's flagship: the
// whole card is token triggers, three deep, and before ADR 0083 not
// one of the three could be declared.
func TestReefWormChainsThreeTokenDiesTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	worm := pushDiesCreatureForTest(g, me.ID, "Reef Worm", reefWormOracle, "Creature — Worm", 0, 1)

	destroyAndSettle(t, g, worm)
	fish := findBattlefieldByName(g, "Fish")
	if fish == uuid.Nil {
		t.Fatal("the Worm's dies-trigger made no Fish")
	}
	if c, _ := battlefieldCard(g, fish); c.Power != 3 || c.Toughness != 3 {
		t.Errorf("Fish is %d/%d, want 3/3", c.Power, c.Toughness)
	}

	destroyAndSettle(t, g, fish)
	whale := findBattlefieldByName(g, "Whale")
	if whale == uuid.Nil {
		t.Fatal("the Fish's own dies-trigger made no Whale — the token's trigger did not fire")
	}
	if c, _ := battlefieldCard(g, whale); c.Power != 6 || c.Toughness != 6 {
		t.Errorf("Whale is %d/%d, want 6/6", c.Power, c.Toughness)
	}

	destroyAndSettle(t, g, whale)
	kraken := findBattlefieldByName(g, "Kraken")
	if kraken == uuid.Nil {
		t.Fatal("the Whale's own dies-trigger made no Kraken")
	}
	if c, _ := battlefieldCard(g, kraken); c.Power != 9 || c.Toughness != 9 {
		t.Errorf("Kraken is %d/%d, want 9/9", c.Power, c.Toughness)
	}
	// And the chain stops: a 9/9 Kraken prints nothing, so it is a row
	// in the token table rather than a fourth catalog template.
	if c, _ := battlefieldCard(g, kraken); len(game.TriggersForCard(c)) != 0 {
		t.Error("the Kraken has a trigger it does not print")
	}
}

// TestReefWormTokensCeaseToExistAfterTheirTriggerFires is CR 111.7
// then CR 704.5d, at the card level: the Fish reaches the graveyard,
// its trigger reads it there, and the sweep takes it away — no former
// token is left lying in a graveyard for a "cards in your graveyard"
// count to trip over.
func TestReefWormTokensCeaseToExistAfterTheirTriggerFires(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	worm := pushDiesCreatureForTest(g, me.ID, "Reef Worm", reefWormOracle, "Creature — Worm", 0, 1)

	destroyAndSettle(t, g, worm)
	fish := findBattlefieldByName(g, "Fish")
	if fish == uuid.Nil {
		t.Fatal("no Fish")
	}
	destroyAndSettle(t, g, fish)

	if findBattlefieldByName(g, "Whale") == uuid.Nil {
		t.Fatal("the Fish's trigger did not fire, so this test proves nothing")
	}
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Fish" {
			t.Error("the Fish token is still in the graveyard — CR 704.5d never swept it")
		}
	}
	// The Worm is a real card and stays where a real card goes.
	if !me.Graveyard.Contains(worm) {
		t.Error("Reef Worm itself should be in the graveyard")
	}
}

// --- Nesting Dragon -----------------------------------------------

// TestNestingDragonEggHatchesIntoADragonWithItsOwnFirebreathing
// exercises BOTH slots on one card: the Egg's trigger and the
// Dragon's activated ability.
func TestNestingDragonEggHatchesIntoADragonWithItsOwnFirebreathing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Nesting Dragon", nestingDragonOracle, "Creature — Dragon")

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)

	egg := findBattlefieldByName(g, "Dragon Egg")
	if egg == uuid.Nil {
		t.Fatal("landfall made no Dragon Egg")
	}
	eggCard, _ := battlefieldCard(g, egg)
	if eggCard.Power != 0 || eggCard.Toughness != 2 {
		t.Errorf("Dragon Egg is %d/%d, want 0/2", eggCard.Power, eggCard.Toughness)
	}
	if !game.HasKeyword(&eggCard, "defender") {
		t.Error("the Dragon Egg lost its printed defender")
	}

	destroyAndSettle(t, g, egg)
	dragon := findBattlefieldByName(g, "Dragon")
	if dragon == uuid.Nil {
		t.Fatal("the Egg's own dies-trigger hatched nothing")
	}
	dragonCard, _ := battlefieldCard(g, dragon)
	if dragonCard.Power != 2 || dragonCard.Toughness != 2 {
		t.Errorf("Dragon is %d/%d, want 2/2", dragonCard.Power, dragonCard.Toughness)
	}
	if !game.HasKeyword(&dragonCard, "flying") {
		t.Error("the hatched Dragon lost its printed flying")
	}

	// The activated ability lives on the same token as the trigger
	// the Egg used, which is the composition #521's two slots could
	// not be asked about.
	abs := game.ActivatedAbilitiesForCard(dragonCard)
	if len(abs) != 1 {
		t.Fatalf("the Dragon has %d activated abilities, want its firebreathing", len(abs))
	}
	if abs[0].Cost.Mana != "{R}" {
		t.Errorf("firebreathing costs %q, want {R}", abs[0].Cost.Mana)
	}

	// And it pumps THE TOKEN. "This token" on a token's own ability is
	// the source of that ability (CR 113.7a), which is
	// item.SourceCardID — the one thing a token's activated ability
	// could plausibly get wrong, since the effect has no target to
	// name it by. Permissive mode waives the unfunded {R}.
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if pumped, _ := battlefieldCard(g, dragon); pumped.CurrentPower() != 3 {
		t.Errorf("the Dragon is %d/%d after firebreathing, want 3/2",
			pumped.CurrentPower(), pumped.CurrentToughness())
	}
}

// --- Mysidian Elder, and the Wizard two cards share ----------------

// TestMysidianElderWizardPingsOnANoncreatureSpell is the token trigger
// that watches an event elsewhere rather than its own death.
func TestMysidianElderWizardPingsOnANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	castCatalogSpell(t, g, "Mysidian Elder", "Creature — Human Wizard", mysidianElderOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g) // the ETB trigger

	wizard := findBattlefieldByName(g, "Wizard")
	if wizard == uuid.Nil {
		t.Fatal("Mysidian Elder made no Wizard")
	}

	before := make(map[uuid.UUID]int, len(g.Seats))
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}

	// Any noncreature spell; Divination is a catalog sorcery.
	castCatalogSpell(t, g, "Divination", "Sorcery", "273b339c-964b-4a18-8eb5-ceb8abcdfd9e", nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		want := before[p.ID]
		if p.ID != me.ID {
			want--
		}
		if p.Life != want {
			t.Errorf("%s life %d -> %d, want %d — the Wizard pings each OPPONENT and not its controller",
				p.Name, before[p.ID], p.Life, want)
		}
	}
}

// --- Fable of the Mirror-Breaker's Goblin Shaman -------------------

// TestFableGoblinShamanMakesATreasureWhenItAttacks clears the caveat
// this card shipped with. A token whose trigger creates ANOTHER token
// is two catalog entries and no new machinery.
func TestFableGoblinShamanMakesATreasureWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := pushToken(g, me.ID, fableGoblinShamanToken())

	if got := countBattlefieldByName(g, "Treasure"); got != 0 {
		t.Fatalf("setup: %d Treasures already out", got)
	}
	attackWith(t, g, opp.ID, shaman)
	passPriorityAroundTable(t, g)

	if got := countBattlefieldByName(g, "Treasure"); got != 1 {
		t.Errorf("Treasures after the Shaman attacked = %d, want 1", got)
	}
}

// --- Avatar Roku's Dragon ------------------------------------------

// TestAvatarRokuDragonTokenHasItsOwnFirebending clears the second of
// that card's two caveats: the token's firebending 4 is its own
// triggered ability, not the Avatar's.
func TestAvatarRokuDragonTokenHasItsOwnFirebending(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := pushToken(g, me.ID, firebendingDragonToken())

	// Declared and locked in, then ONE priority round to resolve the
	// trigger — still inside the declare-attackers step, because the
	// card's standing caveat is that firebending's mana empties with
	// the step instead of lasting until end of combat.
	declareAttack(t, g, opp.ID, dragon)
	passPriorityAroundTable(t, g)

	red := 0
	for _, tok := range me.ManaPool {
		if tok.Color == "R" {
			red++
		}
	}
	if red != 4 {
		t.Errorf("mana pool holds %d red after the Dragon attacked, want the four {R} firebending 4 adds", red)
	}
}

// --- Beledros Witherbloom's and Sedgemoor Witch's Pest -------------

// TestPestTokenGainsOneLifeWhenItDies is the example the seam row was
// written around, and the template two cards share.
func TestPestTokenGainsOneLifeWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pest := pushToken(g, me.ID, PestToken())
	lifeBefore := me.Life

	destroyAndSettle(t, g, pest)
	if me.Life != lifeBefore+1 {
		t.Errorf("life %d -> %d, want +1 from the Pest's own dies-trigger", lifeBefore, me.Life)
	}
}

// TestBeledrosPestIsTheSameTokenSedgemoorWitchMakes — two cards that
// print the same token print the same object, so they share one
// template and one catalog key rather than two that can drift.
func TestBeledrosPestIsTheSameTokenSedgemoorWitchMakes(t *testing.T) {
	if got := PestToken().TokenKey; got != game.TokenKey("pest") {
		t.Fatalf("Pest token key = %q, want token:pest", got)
	}
	if spec, ok := Lookup(beledrosOracle); !ok || spec.Completeness != CompletenessFull {
		t.Errorf("Beledros Witherbloom is %v, want full now that the Pest carries its trigger",
			spec.Completeness)
	}
}

// --- the declaration itself ----------------------------------------

// TestTokenTemplateRegistrationRules states what a template must
// satisfy rather than provoking each panic at boot. The rules are the
// reason a token's ability cannot quietly go missing: a template with
// abilities and no text, or with a closure left on the instance, is
// refused.
func TestTokenTemplateRegistrationRules(t *testing.T) {
	ok := tokenTemplate{
		Slug: "example",
		Card: game.Card{Name: "Example", TypeLine: "Token Creature — Example"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Example — nothing", Do()),
		},
		Text: "When this token dies, nothing happens.",
	}
	if why := checkTokenTemplate(ok); why != "" {
		t.Fatalf("a well-formed template was refused: %s", why)
	}

	bad := []struct {
		name   string
		break_ func(*tokenTemplate)
	}{
		{"no slug", func(tt *tokenTemplate) { tt.Slug = "" }},
		{"no abilities", func(tt *tokenTemplate) { tt.Triggered = nil }},
		{"no printed text", func(tt *tokenTemplate) { tt.Text = "" }},
		{"an oracle ID as well", func(tt *tokenTemplate) { tt.Card.OracleID = "oracle-1" }},
		{"its own token key", func(tt *tokenTemplate) { tt.Card.TokenKey = "token:elsewhere" }},
		{"a closure on the instance", func(tt *tokenTemplate) {
			tt.Card.ActivatedAbilities = []game.ActivatedAbilityShape{{Label: "x"}}
		}},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			tt := ok
			tc.break_(&tt)
			if why := checkTokenTemplate(tt); why == "" {
				t.Error("accepted, want a refusal with a reason")
			}
		})
	}
}

// TestBuildTokenDefCarriesEverySlot is the projection itself, stated
// hermetically. It is the one assertion that covers the STATIC slot:
// no card in the catalog prints a token static yet (Avatar Kuruk's
// Spirit is the first that will, and it waits on CardDef.BlockRules),
// so without this a slot could be dropped from the projection and
// every other test would stay green.
func TestBuildTokenDefCarriesEverySlot(t *testing.T) {
	def := buildTokenDef(tokenTemplate{
		Slug:      "example",
		Card:      game.Card{Name: "Example", TypeLine: "Token Creature — Example"},
		Mana:      []game.ManaAbilityShape{{Produced: "{C}"}},
		Activated: []game.ActivatedAbilityShape{{Label: "{T}: Nothing"}},
		Triggered: []game.TriggeredAbility{WhenThisDies("Example — nothing", Do())},
		Static:    []game.StaticAbility{{Layer: game.Layer7PT, SubLayer: game.SubLayer7C_Modify}},
		Text:      "Everything.",
	})
	if len(def.ManaAbilities) != 1 {
		t.Error("the mana slot did not reach the def")
	}
	if len(def.Activated) != 1 {
		t.Error("the activated slot did not reach the def")
	}
	if len(def.Triggered) != 1 {
		t.Error("the triggered slot did not reach the def")
	}
	if len(def.Static) != 1 {
		t.Error("the static slot did not reach the def")
	}
	if def.TokenText != "Everything." {
		t.Errorf("the printed text did not reach the def: %q", def.TokenText)
	}
}

// TestEveryTokenTemplateDeclaresItsPrintedText walks the real list: a
// token nobody can read is the failure mode ADR 0083 decision 6 names,
// and the registration rule above is only worth anything if the list
// it guards is the live one.
func TestEveryTokenTemplateDeclaresItsPrintedText(t *testing.T) {
	for _, build := range tokenTemplates {
		tmpl := build()
		if why := checkTokenTemplate(tmpl); why != "" {
			t.Errorf("%s: %s", tmpl.Slug, why)
		}
		if got := game.TokenTextForCard(tokenFromCatalog(build)); got != tmpl.Text {
			t.Errorf("%s: the wire reads %q, the template declares %q", tmpl.Slug, got, tmpl.Text)
		}
	}
}
