package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander 2020, OLD batch 124: token makers, life gain and loss.

const (
	oldB124Akim        = "2a20530b-d1d6-4dac-8d08-f583bea55e6e"
	oldB124Gavi        = "716a3bb8-04bd-41f5-a568-131f175e00a6"
	oldB124Drake       = "86d3def3-53fa-4cad-bb25-d3275af1e3f5"
	oldB124Cairn       = "ae896d87-59c7-4d8d-a137-9e471d3e174e"
	oldB124Wort        = "d6b67c7e-3b24-4f60-8e3a-c6ca45229d7f"
	oldB124Wydwen      = "7891ec87-8d1c-42fc-b00a-4c64ee5edbff"
	oldB124Wanderer    = "a549d54f-2640-455c-8a93-bef4df2f9e8b"
	oldB124Selective   = "b68df6d7-d69e-446f-9b1b-7a6ca56efbec"
	oldB124Myriarch    = "c39dca94-3849-4edb-bba9-e7a2eefa5021"
	oldB124Doomsayer   = "e05e70cd-7af3-428e-82e3-73418914d727"
	oldB124Gladiator   = "4da2425d-2118-4116-855d-2ef80693a7b7" // creature with cycling {1}{B}
	oldB124GladTL      = "Creature — Zombie Gladiator"
	oldB124BearTL      = "Creature — Bear"
	oldB124BirdAbility = "flying"
)

func TestOldB124CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB124Akim: "Akim, the Soaring Wind", oldB124Gavi: "Gavi, Nest Warden", oldB124Drake: "Drake Haven",
		oldB124Cairn: "Spirit Cairn", oldB124Wort: "Wort, the Raidmother", oldB124Wydwen: "Wydwen, the Biting Gale",
		oldB124Wanderer: "Cairn Wanderer", oldB124Selective: "Selective Adaptation", oldB124Myriarch: "Majestic Myriarch",
		oldB124Doomsayer: "Thraben Doomsayer",
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

func oldB124Named(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			out = append(out, c)
		}
	}
	return out
}

func oldB124HandCycler(me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Undead Gladiator", TypeLine: oldB124GladTL, OracleID: oldB124Gladiator, Owner: me.ID, Controller: me.ID})
	return id
}

func TestOldB124AkimMakesABirdOnlyForTheFirstTokenEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Akim, the Soaring Wind", "Legendary Creature — Bird Dinosaur", oldB124Akim, 3, 4)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Human"), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(oldB124Named(g, me.ID, "Bird")); n != 1 {
		t.Fatalf("birds after the first tokens of the turn: %d, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Human"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(oldB124Named(g, me.ID, "Bird")); n != 1 {
		t.Errorf("a second creation the same turn made another Bird (%d)", n)
	}
}

func TestOldB124AkimIgnoresTokensMadeBeforeItArrived(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Human"), 1) })
	effectstest.B12Push(g, me.ID, "Akim, the Soaring Wind", "Legendary Creature — Bird Dinosaur", oldB124Akim, 3, 4)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Human"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(oldB124Named(g, me.ID, "Bird")); n != 0 {
		t.Errorf("a token made before Akim arrived let it make %d Birds; this is not the first time this turn", n)
	}
}

func TestOldB124AkimGivesTokensDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	akim := effectstest.B12Push(g, me.ID, "Akim, the Soaring Wind", "Legendary Creature — Bird Dinosaur", oldB124Akim, 3, 4)
	tok := effectstest.B12Creature(g, me.ID, "Plain Bear", oldB124BearTL, 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Human"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ApaMana(me, "C", "C", "C", "U", "R", "W")
	if err := g.ActivateCatalogAbility(me.ID, akim, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range oldB124Named(g, me.ID, "Human") {
		if !effectstest.HasKeywordOf(g, c.InstanceID, "double strike") {
			t.Error("a creature token did not gain double strike")
		}
	}
	if effectstest.HasKeywordOf(g, tok, "double strike") {
		t.Error("a nontoken creature gained double strike")
	}
}

func TestOldB124GaviCyclesTheFirstCardForFreeAndMakesACatOnTheSecondDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Gavi, Nest Warden", "Legendary Creature — Human Shaman", oldB124Gavi, 2, 5)
	first, second := oldB124HandCycler(me), oldB124HandCycler(me)
	twin := game.AltEquipIndexBase + 1
	ref := game.AltEquipRefPrefix + "own:1"
	if err := g.ActivateCatalogAbility(me.ID, first, twin, game.ActivateAbilityParams{Strict: true, Ref: ref}); err != nil {
		t.Fatalf("the first cycle for {0}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(first) {
		t.Fatal("the first card was not cycled")
	}
	if err := g.ActivateCatalogAbility(me.ID, second, twin, game.ActivateAbilityParams{Strict: true, Ref: ref}); err == nil {
		t.Error("a second cycle for {0} in one turn was allowed")
	}
	// The second card drawn this turn makes a Dinosaur Cat, and only the second.
	for drawn := g.TurnTallyFor(me.ID).CardsDrawn; drawn < 4; drawn++ {
		effectstest.DrawN(g, me, 1)
		effectstest.PassPriorityAroundTable(t, g)
		now := g.TurnTallyFor(me.ID).CardsDrawn
		want := 0
		if now >= 2 {
			want = 1
		}
		if n := len(oldB124Named(g, me.ID, "Dinosaur Cat")); n != want {
			t.Fatalf("after %d cards drawn this turn there are %d Cats, want %d", now, n, want)
		}
	}
}

func TestOldB124DrakeHavenPaysOneForADrakeOnACycle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Drake Haven", oldB124Drake, "Enchantment", 0, 0))
	id := oldB124HandCycler(me)
	effectstest.ApaMana(me, "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("cycling did not offer the payment")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(oldB124Named(g, me.ID, "Drake")); n != 1 {
		t.Errorf("drakes: %d, want 1", n)
	}
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a cycled card triggered Drake Haven twice")
	}
}

func TestOldB124SpiritCairnWatchesEveryPlayersDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Spirit Cairn", oldB124Cairn, "Enchantment", 0, 0))
	// The opponent's discard (here: their cleanup-style discard via an effect).
	card := game.NewCard("Pitched Card", opp.ID)
	opp.Hand.PushTop(card)
	g.WithWriteLock(func() { g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: opp.ID, Source: uuid.New(), N: 1}) })
	effectstest.AnswerDiscard(t, g, opp.ID, card.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("an opponent's discard did not offer the Cairn's controller the payment")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(oldB124Named(g, me.ID, "Spirit")); n != 1 {
		t.Errorf("spirits: %d, want 1", n)
	}
}

func TestOldB124WortMakesTwoGoblinsAndGrantsConspireToRedOrGreenSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Wort, the Raidmother", TypeLine: "Legendary Creature — Goblin Shaman", OracleID: oldB124Wort, ManaCost: "{4}{R/G}{R/G}", Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(oldB124Named(g, me.ID, "Goblin Warrior")); n != 2 {
		t.Errorf("Goblin Warriors: %d, want 2", n)
	}
	spec, _ := effects.Lookup(oldB124Wort)
	if len(spec.GrantedOptionalCosts) != 1 || spec.GrantedOptionalCosts[0].Spells == nil {
		t.Fatalf("granted optional costs: %+v", spec.GrantedOptionalCosts)
	}
	reaches := spec.GrantedOptionalCosts[0].Spells
	for _, tc := range []struct {
		name, typeLine, cost string
		want                 bool
	}{
		{"Red Instant", "Instant", "{1}{R}", true},
		{"Green Sorcery", "Sorcery", "{2}{G}", true},
		{"Blue Instant", "Instant", "{1}{U}", false},
		{"Red Creature", "Creature — Goblin", "{1}{R}", false},
		{"Gold Sorcery", "Sorcery", "{W}{G}", true},
	} {
		if got := reaches(g, me.ID, game.Card{Name: tc.name, TypeLine: tc.typeLine, ManaCost: tc.cost}); got != tc.want {
			t.Errorf("%s: conspire granted = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestOldB124WydwenReturnsToHandForOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	w := effectstest.B12Push(g, me.ID, "Wydwen, the Biting Gale", "Legendary Creature — Faerie Wizard", oldB124Wydwen, 3, 3)
	life := me.Life
	effectstest.ApaMana(me, "U", "B")
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(w) || g.Battlefield.Contains(w) {
		t.Error("Wydwen did not return to its owner's hand")
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	effectstest.ApaMana(me, "U", "B")
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Wydwen's ability was activated from the hand")
	}
}

func TestOldB124ThrabenDoomsayerMakesHumansAndPumpsAtFiveLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	d := effectstest.B12Push(g, me.ID, "Thraben Doomsayer", "Creature — Human Cleric", oldB124Doomsayer, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB124BearTL, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, d, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("tap for a token: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(oldB124Named(g, me.ID, "Human")); n != 1 {
		t.Errorf("Human tokens: %d, want 1", n)
	}
	power := func(id uuid.UUID) int { return effectstest.LayeredCard(t, g, id).PowerForComparison() }
	if power(bear) != 2 {
		t.Fatalf("the Bear is %d power at %d life", power(bear), me.Life)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 5-me.Life) })
	g.BumpLayerVersionForTest()
	if power(bear) != 4 || power(d) != 2 {
		t.Errorf("at 5 life the Bear is %d (want 4) and the Doomsayer %d (want 2: other creatures only)", power(bear), power(d))
	}
}

func TestOldB124CairnWandererTakesKeywordsFromGraveyardCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := effectstest.B12Push(g, me.ID, "Cairn Wanderer", "Creature — Shapeshifter", oldB124Wanderer, 4, 4)
	has := func(kw string) bool { return effectstest.HasKeywordOf(g, w, kw) }
	if has("flying") {
		t.Fatal("flying with empty graveyards")
	}
	effectstest.PushGraveyardTyped(opp, "Their Drake", "Creature — Drake", "flying", "vigilance")
	effectstest.PushGraveyardTyped(me, "My Relic", "Artifact", "trample")
	g.BumpLayerVersionForTest()
	if !has("flying") || !has("vigilance") {
		t.Error("a creature card with flying and vigilance in a graveyard did not lend them")
	}
	if has("trample") {
		t.Error("a noncreature card lent trample")
	}
	effectstest.PushGraveyardTyped(me, "Walker", "Creature — Elf", "swampwalk", "protection from red")
	g.BumpLayerVersionForTest()
	if !has("swampwalk") || !has("protection from red") {
		t.Error("landwalk and protection were not lent as the same ability")
	}
}

func TestOldB124MajesticMyriarchIsTwiceTheCreaturesAndCopiesKeywordsAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	m := effectstest.B12Push(g, me.ID, "Majestic Myriarch", "Creature — Chimera", oldB124Myriarch, 0, 0)
	effectstest.B12Creature(g, me.ID, "Bear", oldB124BearTL, 1, 1)
	flier := effectstest.B12Creature(g, me.ID, "Bird", "Creature — Bird", 1, 1)
	_ = flier
	g.BumpLayerVersionForTest()
	if p := effectstest.LayeredCard(t, g, m).PowerForComparison(); p != 6 {
		t.Errorf("power %d with three creatures, want 6", p)
	}
	// A creature with trample and deathtouch makes the Myriarch have both at combat.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Bear" {
				g.Battlefield.Cards[i].Keywords = []string{"trample", "deathtouch"}
			}
		}
	})
	g.BumpLayerVersionForTest()
	if effectstest.HasKeywordOf(g, m, "trample") {
		t.Fatal("trample before combat")
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, m, "trample") || !effectstest.HasKeywordOf(g, m, "deathtouch") {
		t.Error("the Myriarch did not gain trample and deathtouch at the beginning of combat")
	}
	if effectstest.HasKeywordOf(g, m, "flying") {
		t.Error("it gained flying with no flier")
	}
}

func TestOldB124SelectiveAdaptationSortsTheTopSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	me.Library.Cards = nil
	mk := func(name, typeLine string, kws ...string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Keywords: kws, Owner: me.ID, Controller: me.ID, Power: 2, Toughness: 2})
		return id
	}
	bottom := mk("Bottom Card", "Creature — Bear")
	deep := mk("Seventh Card", "Creature — Bear")
	flier := mk("Flier", "Creature — Bird", "flying")
	striker := mk("Striker", "Creature — Knight", "first strike", "vigilance")
	junk := mk("Junk", "Creature — Bear")
	// Three cards more are revealed than chosen from; seven in all.
	extra1 := mk("Extra One", "Creature — Bear")
	extra2 := mk("Extra Two", "Creature — Bear")
	extra3 := mk("Extra Three", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Selective Adaptation", "Sorcery", oldB124Selective, nil)
	for i := 0; i < 6; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		open := effectstest.PendingOfKind(g, game.PendingChoiceChooseCards)
		if open == nil {
			break
		}
		// Whatever is asked for: take the first candidate, and the Flier onto the battlefield.
		cands := open.ChooseCards
		choice := cands[0]
		for _, id := range cands {
			if id == flier {
				choice = id
			}
		}
		if err := g.ResolveChooseCards(open.ID, open.Chooser, []uuid.UUID{choice}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(flier) && !g.Battlefield.Contains(striker) {
		t.Error("none of the chosen cards entered the battlefield")
	}
	inHand := me.Hand.Contains(flier) || me.Hand.Contains(striker)
	if !inHand {
		t.Error("the other chosen card did not reach the hand")
	}
	for _, id := range []uuid.UUID{junk, extra1, extra2, extra3, deep} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("an unchosen revealed card %v is not in the graveyard", id)
		}
	}
	if !me.Library.Contains(bottom) {
		t.Error("the eighth card was touched")
	}
}
