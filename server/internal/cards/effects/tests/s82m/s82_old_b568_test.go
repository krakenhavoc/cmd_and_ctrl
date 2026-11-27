package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b568_test.go — Coldsnap, batch 568 (#5597): each test plays the
// real catalog entry through the real engine paths.

const (
	b568Wurm      = "49d4aef7-d2d1-4f66-bc2a-79ec8cc96565"
	b568WarCry    = "e45a1626-d659-4eda-bc6b-bd8119ce7386"
	b568Jokul     = "fbf074b2-d650-43b2-b8b8-d7a778dfd27e"
	b568Skyknight = "df344469-1177-4168-91af-3621b2a65182"
	b568Ghoul     = "5ebc166d-98c5-4ab7-9610-bc5bcd231d80"
	b568Thermopod = "c886fa01-bcee-4c73-9c46-2328711fc4f0"
	b568Shackles  = "45ca8431-befd-4542-9dd7-503bf1afaac5"
	b568Transfuse = "d7ff0525-8617-4e0f-92db-660f67d706ca"

	b568DemonicTutor = "82004860-e589-4e38-8d61-8c0210e4ea39"
	b568TemporalAdpt = "fd9b5462-6578-4bf6-8026-23eaf5af3eee"
)

func TestS82OldB568CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b568Wurm: "Panglacial Wurm", b568WarCry: "Kjeldoran War Cry", b568Jokul: "Jokulmorder",
		b568Skyknight: "Tresserhorn Skyknight", b568Ghoul: "Gutless Ghoul", b568Thermopod: "Thermopod",
		b568Shackles: "Gelid Shackles", b568Transfuse: "Rime Transfusion",
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

func b568SnowLand(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, "Snow-Covered "+subtype, "Basic Snow Land — "+subtype, "", 0, 0)
}

func b568Land(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, subtype, "Basic Land — "+subtype, "", 0, 0)
}

func TestS82OldB568PanglacialWurmIsCastFromTheLibraryOnlyWhileSearching(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	wurm := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: wurm, Name: "Panglacial Wurm", TypeLine: "Creature — Wurm", ManaCost: "{5}{G}{G}",
		OracleID: b568Wurm, Power: 9, Toughness: 5, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "G", "G", "G", "G", "G", "G", "G")
	if err := g.CastSpell(me.ID, wurm, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("the Wurm was cast from the library with no search open")
	}
	me.ManaPool = nil
	effectstest.CastCatalogSpell(t, g, "Demonic Tutor", "Sorcery", b568DemonicTutor, nil)
	effectstest.B06AddMana(me, "G", "G", "G", "G", "G", "G", "G")
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.SearchPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("Demonic Tutor opened no search")
	}
	if err := g.CastWhileSearching(me.ID, prompt.ID, wurm, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the Wurm mid-search: %v", err)
	}
	if !g.Stack.Contains(wurm) || me.Library.Contains(wurm) {
		t.Fatal("the Wurm is not on the stack")
	}
	after := effectstest.SearchPromptFor(g, me.ID)
	if after == nil {
		t.Fatal("the search ended with the cast; it goes on")
	}
	var pick uuid.UUID
	for _, id := range after.SearchCards {
		pick = id
		break
	}
	if err := g.ResolveSearchLibrary(after.ID, me.ID, []uuid.UUID{pick}); err != nil {
		t.Fatalf("finishing the search: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) {
		t.Error("the Wurm did not resolve")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, wurm, "trample") {
		t.Error("the Wurm has no trample")
	}
}

func TestS82OldB568KjeldoranWarCryCountsEveryWarCryInEveryGraveyard(t *testing.T) {
	for _, tc := range []struct{ inGraveyards, want int }{{0, 3}, {2, 5}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B16Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2, "G")
		theirs := effectstest.B16Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2, "G")
		effectstest.AdvanceToMain(t, g)
		for i := 0; i < tc.inGraveyards; i++ {
			owner := me
			if i%2 == 1 {
				owner = opp
			}
			owner.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Kjeldoran War Cry", TypeLine: "Instant", OracleID: b568WarCry, Owner: owner.ID, Controller: owner.ID})
		}
		effectstest.B06AddMana(me, "W", "W")
		effectstest.CastCatalogSpell(t, g, "Kjeldoran War Cry", "Instant", b568WarCry, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, mine); p != tc.want || tough != tc.want {
			t.Errorf("%d in graveyards: my creature is %d/%d, want %d/%d", tc.inGraveyards, p, tough, tc.want, tc.want)
		}
		if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
			t.Errorf("their creature is %d/%d, want 2/2", p, tough)
		}
	}
}

func TestS82OldB568TresserhornSkyknightIgnoresDamageFromFirstStrikers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sky := effectstest.B12Push(g, me.ID, "Tresserhorn Skyknight", "Creature — Zombie Knight", b568Skyknight, 5, 3)
	striker := effectstest.Pr7bCreature(g, opp.ID, "Striker", 2, "{1}", []string{"R"}, "first strike")
	plain := effectstest.Pr7bCreature(g, opp.ID, "Plain", 2, "{1}", []string{"R"})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(striker, sky, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if got := effectstest.DamageOn(t, g, sky); got != 0 {
		t.Errorf("a first striker dealt %d to the Skyknight, want 0", got)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(plain, sky, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if got := effectstest.DamageOn(t, g, sky); got != 2 {
		t.Errorf("a creature without first strike dealt %d to the Skyknight, want 2", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, sky, "flying") {
		t.Error("the Skyknight has no flying")
	}
}

func TestS82OldB568GutlessGhoulSacrificesAnyCreatureForTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ghoul := effectstest.B12Push(g, me.ID, "Gutless Ghoul", "Snow Creature — Zombie", b568Ghoul, 2, 2)
	fodder := effectstest.B16Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1, "G")
	land := b568Land(g, me.ID, "Forest")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land was accepted as the sacrificed creature")
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || me.Life != life+2 {
		t.Errorf("fodder on battlefield %v, life %d (was %d), want it gone and +2", g.Battlefield.Contains(fodder), me.Life, life)
	}
	// The Ghoul may sacrifice itself.
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{ghoul}}); err != nil {
		t.Fatalf("activate sacrificing itself: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ghoul) || me.Life != life+4 {
		t.Errorf("Ghoul on battlefield %v, life %d, want it gone and +4", g.Battlefield.Contains(ghoul), me.Life)
	}
}

func TestS82OldB568ThermopodGainsHasteForSnowAndSacrificesForRedMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pod := effectstest.B12Push(g, me.ID, "Thermopod", "Snow Creature — Slug", b568Thermopod, 4, 3)
	fodder := effectstest.B16Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1, "G")
	effectstest.AdvanceToMain(t, g)
	b568Land(g, me.ID, "Swamp")
	if err := g.ActivateCatalogAbility(me.ID, pod, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("{S} was paid with a nonsnow Swamp")
	}
	b568SnowLand(g, me.ID, "Mountain")
	if err := g.ActivateCatalogAbility(me.ID, pod, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("{S}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, pod, "haste") {
		t.Error("Thermopod did not gain haste")
	}
	if err := g.ActivateManaAbility(me.ID, pod, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("Sacrifice a creature: Add {R}: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) == 0 || got[len(got)-1] != "R" {
		t.Errorf("pool %v, want it to end with {R}", got)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
}

func TestS82OldB568GelidShacklesStopsBlockingAndActivationAndGrantsDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adept := effectstest.B12Push(g, opp.ID, "Temporal Adept", "Creature — Human Wizard", b568TemporalAdpt, 1, 3)
	attacker := effectstest.B16Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	shackles := effectstest.B12Push(g, opp.ID, "Gelid Shackles", "Snow Enchantment — Aura", b568Shackles, 0, 0)
	effectstest.AttachForTest(t, g, shackles, adept)
	effectstest.B06AddMana(opp, "U", "U", "U")
	if err := g.ActivateCatalogAbility(opp.ID, adept, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(attacker)}); err == nil {
		t.Error("an enchanted creature's activated ability was activated")
	}
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(adept, attacker); err == nil {
		t.Error("an enchanted creature blocked")
	}
}

func TestS82OldB568GelidShacklesSnowAbilityGivesDefenderUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	shackles := effectstest.B12Push(g, me.ID, "Gelid Shackles", "Snow Enchantment — Aura", b568Shackles, 0, 0)
	effectstest.AttachForTest(t, g, shackles, bear)
	b568Land(g, me.ID, "Plains")
	if err := g.ActivateCatalogAbility(me.ID, shackles, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("{S} was paid with a nonsnow Plains")
	}
	b568SnowLand(g, me.ID, "Plains")
	if err := g.ActivateCatalogAbility(me.ID, shackles, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("{S}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "defender") {
		t.Error("the enchanted creature did not gain defender")
	}
}

func TestS82OldB568RimeTransfusionPumpsAndGrantsSnowEvasion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	plainBlocker := effectstest.B16Creature(g, opp.ID, "Plain Blocker", "Creature — Bear", 2, 2, "G")
	snowBlocker := effectstest.B16Creature(g, opp.ID, "Snow Blocker", "Snow Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	aura := effectstest.B12Push(g, me.ID, "Rime Transfusion", "Snow Enchantment — Aura", b568Transfuse, 0, 0)
	effectstest.AttachForTest(t, g, aura, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 3 {
		t.Fatalf("enchanted Bear is %d/%d, want 4/3", p, tough)
	}
	b568SnowLand(g, me.ID, "Swamp")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("the granted {S} ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(plainBlocker, bear); err == nil {
		t.Error("a nonsnow creature blocked the evasive Bear")
	}
	if err := g.DeclareBlocker(snowBlocker, bear); err != nil {
		t.Errorf("a snow creature could not block it: %v", err)
	}
}

func TestS82OldB568JokulmorderSacrificesItselfWithoutFiveLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 4; i++ {
		b568Land(g, me.ID, "Island")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "U", "U", "U", "U", "U", "U")
	id := effectstest.HandCardFull(me, "Jokulmorder", "Creature — Leviathan", "{4}{U}{U}{U}", b568Jokul, []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("Jokulmorder stayed with only four lands to sacrifice")
	}
}

func TestS82OldB568JokulmorderEntersTappedPaysFiveLandsAndUntapsForIslands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var lands []uuid.UUID
	for i := 0; i < 6; i++ {
		lands = append(lands, b568Land(g, me.ID, "Island"))
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "U", "U", "U", "U", "U", "U")
	id := effectstest.HandCardFull(me, "Jokulmorder", "Creature — Leviathan", "{4}{U}{U}{U}", b568Jokul, []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.AnswerChooseCards(t, g, me.ID, lands[:5]...)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Jokulmorder did not stay after the five lands were paid")
	}
	for _, l := range lands[:5] {
		if g.Battlefield.Contains(l) {
			t.Error("a chosen land was not sacrificed")
		}
	}
	if !g.Battlefield.Contains(lands[5]) {
		t.Error("the sixth land was sacrificed")
	}
	if !effectstest.B31Tapped(t, g, id) {
		t.Error("Jokulmorder did not enter tapped")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "trample") {
		t.Error("Jokulmorder has no trample")
	}
	// Playing a non-Island does nothing; an Island lets it untap, as a may.
	me.LandDropsPerTurn += 2
	effectstest.B31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B31Tapped(t, g, id) {
		t.Error("a Forest untapped Jokulmorder")
	}
	effectstest.B31PlayLand(t, g, "Island", "Basic Land — Island", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B31Tapped(t, g, id) {
		t.Error("playing an Island did not untap Jokulmorder")
	}
	// It does not untap during its controller's untap step.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, id).Tapped = true })
	effectstest.AdvanceToDrawStepOfSeat(t, g, 0)
	if !effectstest.B31Tapped(t, g, id) {
		t.Error("Jokulmorder untapped during its controller's untap step")
	}
}
