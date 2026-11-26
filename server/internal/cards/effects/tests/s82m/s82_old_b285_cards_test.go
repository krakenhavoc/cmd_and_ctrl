package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b285_cards_test.go — S82 older sets, Shadows over Innistrad
// batch 285 (#5543): graveyard recursion.

const (
	oldB285Skaab      = "0b40e537-d385-450d-93a9-b60a25bfe508"
	oldB285Geistblast = "10891656-070f-49f4-80e0-78061b773f17"
	oldB285Shamble    = "11a20648-fdb6-4f52-8fda-c66985e0c044"
	oldB285Chaplain   = "16c587db-d07b-47ab-8b03-c3b2cc7144f5"
	oldB285Harness    = "38eaf0ea-8cdd-45f4-bfb9-0f8d0fcf863a"
	oldB285Moondrakes = "95649140-7978-4684-a1f1-af771e1d345a"
	oldB285Builder    = "cca94004-fe8e-404d-9391-ad33d0a126fb"
	oldB285Masterpiec = "dacab848-89ca-46cc-b796-21fc2c218e78"
	oldB285Ghoulsteed = "dbc34afa-16bc-41f2-b74d-29bf2bfe24a9"
)

func TestOldB285CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB285Skaab: "Stitchwing Skaab", oldB285Geistblast: "Geistblast", oldB285Shamble: "Shamble Back",
		oldB285Chaplain: "Nearheath Chaplain", oldB285Harness: "Harness the Storm", oldB285Moondrakes: "Nephalia Moondrakes",
		oldB285Builder: "Stoic Builder", oldB285Masterpiec: "Geralf's Masterpiece", oldB285Ghoulsteed: "Ghoulsteed",
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

// oldB285Grave puts a creature card with the catalog identity into the graveyard.
func oldB285Grave(me *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	return id
}

func oldB285Hand(me *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, effectstest.PutInHand(me, game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}))
	}
	return ids
}

func TestOldB285SkaabAndGhoulsteedReturnTappedForTwoDiscards(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle, mana string
		colors                       []string
		power, toughness             int
		flying                       bool
	}{
		{"Stitchwing Skaab", "Creature — Zombie Horror", oldB285Skaab, "{1}{U}", []string{"C", "U"}, 3, 1, true},
		{"Ghoulsteed", "Creature — Zombie Horse", oldB285Ghoulsteed, "{2}{B}", []string{"C", "C", "B"}, 4, 4, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		id := oldB285Grave(me, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		hand := oldB285Hand(me, 3)
		effectstest.B06AddMana(me, c.colors...)
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: hand[:1]}); err == nil {
			t.Fatalf("%s: activated with one card to discard", c.name)
		}
		effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{DiscardIDs: hand[:2]})
		effectstest.PassPriorityAroundTable(t, g)
		card, ok := effectstest.BattlefieldCard(g, id)
		if !ok || !card.Tapped {
			t.Fatalf("%s: on the battlefield %v, tapped %v; want back and tapped", c.name, ok, card.Tapped)
		}
		if effectstest.InHandOf(me, hand[0]) || effectstest.InHandOf(me, hand[1]) || !effectstest.InHandOf(me, hand[2]) {
			t.Errorf("%s: exactly the two chosen cards should be discarded", c.name)
		}
		if c.flying && !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
			t.Errorf("%s has no flying", c.name)
		}
	}
}

func TestOldB285GeralfsMasterpieceShrinksWithYourHandAndReturnsForThreeDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	piece := effectstest.B12Push(g, me.ID, "Geralf's Masterpiece", "Creature — Zombie Horror", oldB285Masterpiec, 7, 7)
	if p := effectstest.EffectivePower(t, g, piece); p != 7 {
		t.Fatalf("empty hand: power %d, want 7", p)
	}
	hand := oldB285Hand(me, 3)
	effectstest.S29Bump(g)
	if p, tou := effectstest.EffectivePower(t, g, piece), effectstest.EffectiveToughness(t, g, piece); p != 4 || tou != 4 {
		t.Errorf("three cards in hand: %d/%d, want 4/4", p, tou)
	}
	if !effectstest.HasEffectiveKeyword(t, g, piece, "flying") {
		t.Error("no flying")
	}
	grave := oldB285Grave(me, "Geralf's Masterpiece", "Creature — Zombie Horror", oldB285Masterpiec, 7, 7)
	effectstest.B06AddMana(me, "C", "C", "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, grave, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: hand[:2]}); err == nil {
		t.Fatal("activated with two discards")
	}
	effectstest.B16Activate(t, g, me.ID, grave, 0, game.ActivateAbilityParams{DiscardIDs: hand})
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, grave)
	if !ok || !card.Tapped {
		t.Fatalf("on the battlefield %v, tapped %v; want back and tapped", ok, card.Tapped)
	}
}

func TestOldB285ShambleBackExilesACreatureCardAndMakesAZombieAndLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dead := oldB285Grave(opp, "Dead Bear", "Creature — Bear", "", 2, 2)
	land := effectstest.GyCard(opp, "Dead Land", "Land", "", "")
	life := me.Life
	if err := effectstest.CastCatalogSpellErr(t, g, "Shamble Back", "Sorcery", oldB285Shamble, effectstest.CardTarget(land)); err == nil {
		t.Fatal("a land card was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Shamble Back", "Sorcery", oldB285Shamble, effectstest.CardTarget(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, dead) != game.ZoneExile {
		t.Errorf("the creature card is in %v, want exile", effectstest.B12ZoneOf(g, dead))
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Zombie"); n != 1 {
		t.Errorf("%d Zombie tokens, want 1", n)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestOldB285NearheathChaplainMakesTwoSpiritsFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := oldB285Grave(me, "Nearheath Chaplain", "Creature — Human Cleric", oldB285Chaplain, 3, 1)
	effectstest.B06AddMana(me, "C", "C", "W")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if effectstest.B12ZoneOf(g, id) != game.ZoneExile {
		t.Fatalf("the Chaplain is in %v, want exile as a cost", effectstest.B12ZoneOf(g, id))
	}
	effectstest.PassPriorityAroundTable(t, g)
	n, _ := effectstest.CountTokensNamed(g, me.ID, "Spirit")
	if n != 2 {
		t.Fatalf("%d Spirit tokens, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.IsToken() {
			if p, tou := effectstest.EffectivePower(t, g, c.InstanceID), effectstest.EffectiveToughness(t, g, c.InstanceID); p != 1 || tou != 1 {
				t.Errorf("a Spirit is %d/%d, want 1/1", p, tou)
			}
			if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Error("a Spirit has no flying")
			}
		}
	}
	// Sorcery speed: refused in an opponent's turn.
	id2 := oldB285Grave(me, "Nearheath Chaplain", "Creature — Human Cleric", oldB285Chaplain, 3, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.B06AddMana(me, "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, id2, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated outside a main phase of my turn")
	}
}

func TestOldB285MoondrakesGrantsFlyingOnEntryAndFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Nephalia Moondrakes", "Creature — Drake", oldB285Moondrakes)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") || effectstest.HasEffectiveKeyword(t, g, other, "flying") {
		t.Error("only the targeted creature should gain flying on entry")
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	effectstest.AdvanceToMain(t, g)
	id := oldB285Grave(me, "Nephalia Moondrakes", "Creature — Drake", oldB285Moondrakes, 5, 5)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "U", "U")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if effectstest.B12ZoneOf(g, id) != game.ZoneExile {
		t.Fatalf("the Drake is in %v, want exile as a cost", effectstest.B12ZoneOf(g, id))
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{bear, other} {
		if !effectstest.HasEffectiveKeyword(t, g, c, "flying") {
			t.Error("a creature I control did not gain flying from the graveyard ability")
		}
	}
}

func TestOldB285StoicBuilderMayReturnALandCard(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		land := effectstest.GyCard(me, "Dead Land", "Land", "", "")
		creature := oldB285Grave(me, "Dead Bear", "Creature — Bear", "", 2, 2)
		effectstest.CastAndResolveCreature(t, g, "Stoic Builder", "Creature — Human", oldB285Builder)
		effectstest.AnswerPickTargetIfOpen(t, g, land)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, accept)
		effectstest.PassPriorityAroundTable(t, g)
		want := game.ZoneGraveyard
		if accept {
			want = game.ZoneHand
		}
		if got := effectstest.B12ZoneOf(g, land); got != want {
			t.Errorf("accept=%v: the land is in %v, want %v", accept, got, want)
		}
		if effectstest.B12ZoneOf(g, creature) != game.ZoneGraveyard {
			t.Error("a creature card left the graveyard")
		}
	}
}

func TestOldB285GeistblastBurnsAndCopiesFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Geistblast", "Instant", oldB285Geistblast, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("life %d, want %d after two damage", opp.Life, life-2)
	}
	// Now the graveyard ability: copy a spell I control.
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Geistblast", TypeLine: "Instant", OracleID: oldB285Geistblast,
		ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	bolt := effectstest.CastCatalogSpell(t, g, "Geistblast", "Instant", oldB285Geistblast, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.B06AddMana(me, "C", "C", "U")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bolt)})
	if effectstest.B12ZoneOf(g, id) != game.ZoneExile {
		t.Fatalf("the graveyard Geistblast is in %v, want exile as a cost", effectstest.B12ZoneOf(g, id))
	}
	life = opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to choose new targets for the copy")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d, want %d: the spell and its copy each deal 2", opp.Life, life-4)
	}
}

func TestOldB285HarnessTheStormOffersASameNameCastFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Harness the Storm", "Enchantment", oldB285Harness, 0, 0)
	zapInYard := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: zapInYard, Name: "Zap Test", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	other := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: other, Name: "Another Spell", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Zap Test", "Instant", "", []game.TargetRef{})
	effectstest.AnswerPickTargetIfOpen(t, g, zapInYard)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.KeywordCastPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no offer to cast the same-name card from the graveyard")
	}
	if offered := effectstest.KeywordCastCard(prompt); offered != zapInYard {
		t.Errorf("offered %v, want the card named like the spell", offered)
	}
	effectstest.AnswerKeywordCast(t, g, prompt, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, zapInYard) != game.ZoneGraveyard {
		t.Error("declining moved the card")
	}
	_, _ = life, opp
}

func TestOldB285HarnessTheStormCastsTheSameNameCardPayingItsCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Harness the Storm", "Enchantment", oldB285Harness, 0, 0)
	inYard := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: inYard, Name: "Zap Test", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Zap Test", "Instant", "", nil)
	effectstest.AnswerPickTargetIfOpen(t, g, inYard)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.KeywordCastPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no offer")
	}
	_ = prompt
	granted := game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, inYard, granted); err == nil {
		t.Fatal("cast it for free")
	}
	effectstest.FloatMana(t, g, me, "{R}")
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, inYard, granted); err != nil {
		t.Fatalf("cast for {R}: %v", err)
	}
	if z := effectstest.B12ZoneOf(g, inYard); z != game.ZoneStack {
		t.Errorf("the graveyard copy is in %v, want the stack", z)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("its cost was not paid: pool %+v", me.ManaPool)
	}
}

func TestOldB285HarnessTheStormNeedsASameNameCardAndACastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Harness the Storm", "Enchantment", oldB285Harness, 0, 0)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Another Spell", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Zap Test", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.KeywordCastPromptFor(g, me.ID) != nil {
		t.Error("an offer was made with no card of the same name in the graveyard")
	}
}
