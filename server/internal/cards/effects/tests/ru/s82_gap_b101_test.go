package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB101ThornOracle        = "a7894a81-8c0a-4015-8bc1-2698cf5306d1"
	gapB101SlimefootOracle    = "fd541139-de13-4b32-93ba-070ecf711d33"
	gapB101SekKuarOracle      = "94426127-65c2-435e-ba92-423a3c102061"
	gapB101RukarumelOracle    = "84391c32-5bb7-4c36-be50-bbb5f8732156"
	gapB101GravemotherOracle  = "0d59f010-9669-4ad2-9ae9-cd88b85f282c"
	gapB101AniktheaOracle     = "991fd651-0cba-41ee-855a-d01a1a7c8f60"
	gapB101KaradorOracle      = "304c194f-fa0b-42e5-93ca-8c587181d33c"
	gapB101NarciOracle        = "61bbf710-4eb4-4e11-be6e-700a0c6a41ee"
	gapB101HibernationOracle  = "edb8a6d6-9e25-4048-a842-e854eece42c2"
	gapB101SaprolingTokenType = "Token Creature — Saproling"
)

func TestGapB101CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB101ThornOracle:       "Thorn of the Black Rose",
		gapB101SlimefootOracle:   "Slimefoot, the Stowaway",
		gapB101SekKuarOracle:     "Sek'Kuar, Deathkeeper",
		gapB101RukarumelOracle:   "Rukarumel, Biologist",
		gapB101GravemotherOracle: "Sliver Gravemother",
		gapB101AniktheaOracle:    "Anikthea, Hand of Erebos",
		gapB101KaradorOracle:     "Karador, Ghost Chieftain",
		gapB101NarciOracle:       "Narci, Fable Singer",
		gapB101HibernationOracle: "Hibernation Sliver",
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

// Thorn of the Black Rose: deathtouch, and its entry makes you the monarch.
func TestGapB101ThornBecomesTheMonarchOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if g.Monarch != uuid.Nil {
		t.Fatal("somebody is already the monarch")
	}
	id := effectstest.CastCatalogSpell(t, g, "Thorn of the Black Rose", "Creature — Human Assassin", gapB101ThornOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MonSettle(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %s, want the caster %s", g.Monarch, me.ID)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") {
		t.Error("Thorn of the Black Rose has no deathtouch")
	}
}

// Slimefoot drains one for each Saproling of mine that dies, and only mine.
func TestGapB101SlimefootDrainsWhenAYourSaprolingDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.B12Push(g, me.ID, "Slimefoot, the Stowaway", "Legendary Creature — Fungus", gapB101SlimefootOracle, 2, 3)
	mine := effectstest.B12Creature(g, me.ID, "Saproling", gapB101SaprolingTokenType, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Saproling", gapB101SaprolingTokenType, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	myLife, oppLife := me.Life, opp.Life

	effectstest.MidB17Bolt(t, g, effectstest.RprB30Ref(theirs))
	effectstest.MidB17Bolt(t, g, effectstest.RprB30Ref(bear))
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatalf("a non-Saproling or an opponent's Saproling drained: me %d->%d, opp %d->%d", myLife, me.Life, oppLife, opp.Life)
	}
	effectstest.MidB17Bolt(t, g, effectstest.RprB30Ref(mine))
	if me.Life != myLife+1 || opp.Life != oppLife-1 {
		t.Fatalf("life after my Saproling died: me %d->%d, opp %d->%d, want +1 and -1", myLife, me.Life, oppLife, opp.Life)
	}
}

func TestGapB101SlimefootMakesASaprolingForFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	foot := effectstest.B12Push(g, me.ID, "Slimefoot, the Stowaway", "Legendary Creature — Fungus", gapB101SlimefootOracle, 2, 3)
	effectstest.ApaMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, foot, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the {4} ability was activated with three mana")
	}
	effectstest.ApaMana(me, "C")
	effectstest.ApaActivate(t, g, me, foot, 0, game.ActivateAbilityParams{})
	if n := effectstest.B27CountNamed(g, me.ID, "Saproling"); n != 1 {
		t.Fatalf("%d Saprolings, want 1", n)
	}
}

// Sek'Kuar makes a Graveborn for another nontoken creature of mine that
// dies: not for a token, an opponent's creature or itself.
func TestGapB101SekKuarMakesAGravebornForAnotherNontokenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.B12Push(g, me.ID, "Sek'Kuar, Deathkeeper", "Legendary Creature — Orc Shaman", gapB101SekKuarOracle, 4, 3)
	token := effectstest.B12Creature(g, me.ID, "Saproling", gapB101SaprolingTokenType, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	effectstest.MidB17Bolt(t, g, effectstest.RprB30Ref(token))
	effectstest.MidB17Bolt(t, g, effectstest.RprB30Ref(theirs))
	if n := effectstest.CountBattlefieldByName(g, "Graveborn"); n != 0 {
		t.Fatalf("%d Graveborn after a token and an opponent's creature died, want 0", n)
	}
	effectstest.MidB17Bolt(t, g, effectstest.RprB30Ref(bear))
	if n := effectstest.B27CountNamed(g, me.ID, "Graveborn"); n != 1 {
		t.Fatalf("%d Graveborn after my nontoken creature died, want 1", n)
	}
	var tok game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Graveborn" {
			tok = c
		}
	}
	if tok.Power != 3 || tok.Toughness != 1 || !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
		t.Errorf("Graveborn is %d/%d, want a 3/1 with haste", tok.Power, tok.Toughness)
	}
}

// Hibernation Sliver gives every Sliver "Pay 2 life: Return this permanent
// to its owner's hand"; its controller pays it, an opponent cannot.
func TestGapB101HibernationSliverLetsEverySliverReturnItselfForLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.B12Push(g, me.ID, "Hibernation Sliver", "Creature — Sliver", gapB101HibernationOracle, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Test Sliver", "Creature — Sliver", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Sliver", "Creature — Sliver", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a non-Sliver was given the ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, theirs, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("I activated an opponent's Sliver")
	}
	life := me.Life
	effectstest.ApaActivate(t, g, me, other, 0, game.ActivateAbilityParams{})
	if me.Life != life-2 {
		t.Fatalf("life = %d, want %d after paying 2", me.Life, life-2)
	}
	if g.Battlefield.Contains(other) || !me.Hand.Contains(other) {
		t.Fatal("the Sliver was not returned to its owner's hand")
	}
	me.Life = 1
	hib := effectstest.B12Creature(g, me.ID, "Sliver Two", "Creature — Sliver", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, hib, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("paid 2 life with 1 life")
	}
}

// Rukarumel: the chosen type is added to my nontoken creatures and my
// Slivers, to creature spells and creature cards off the battlefield, and
// to nothing of an opponent's, nor to a non-Sliver token.
func TestGapB101RukarumelAddsTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	token := effectstest.B12Creature(g, me.ID, "Saproling", gapB101SaprolingTokenType, 1, 1)
	sliverToken := effectstest.B12Creature(g, me.ID, "Sliver", "Token Creature — Sliver", 1, 1)
	inHand := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: inHand, Name: "Hand Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Hand Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Rukarumel, Biologist", "Legendary Creature — Human Wizard", gapB101RukarumelOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Wizard")

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{
		{"my nontoken creature", mine, true},
		{"my Sliver token", sliverToken, true},
		{"a creature card in my hand", inHand, true},
		{"my non-Sliver token", token, false},
		{"an opponent's creature", theirs, false},
		{"a land card in my hand", land, false},
	} {
		if got := effectstest.B12Card(t, g, tc.id).HasSubtype("Wizard"); got != tc.want {
			t.Errorf("%s is a Wizard = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestGapB101RukarumelMakesASliverForThreeAndTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rukarumel := effectstest.PushCatalogPermanent(g, me.ID, "Rukarumel, Biologist", "Legendary Creature — Human Wizard", gapB101RukarumelOracle, false)
	effectstest.ApaMana(me, "C", "C", "C")
	effectstest.ApaActivate(t, g, me, rukarumel, 0, game.ActivateAbilityParams{})
	if n := effectstest.B27CountNamed(g, me.ID, "Sliver"); n != 1 {
		t.Fatalf("%d Sliver tokens, want 1", n)
	}
	effectstest.ApaMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, rukarumel, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated a {T} ability twice")
	}
}

// Sliver Gravemother: my Sliver creature cards in the graveyard have
// encore {X} (X their mana value), and the Gravemother itself has encore {5}.
func TestGapB101GravemotherGrantsEncoreToSliverCardsInMyGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Sliver Gravemother", "Legendary Creature — Sliver", gapB101GravemotherOracle, false)
	sliver := r37Dead(t, g, me, "Dead Sliver", "Creature — Sliver", "{2}{R}")
	if got, ok := r37EncoreCost(g, sliver); !ok || got != "{3}" {
		t.Errorf("a Sliver card's encore cost is %q (has a row: %v), want {3}", got, ok)
	}
	for _, tc := range []struct{ name, typeLine string }{
		{"Dead Bear", "Creature — Bear"},
		{"Sliver Spell", "Kindred Instant — Sliver"},
	} {
		if _, ok := r37EncoreCost(g, r37Dead(t, g, me, tc.name, tc.typeLine, "{2}")); ok {
			t.Errorf("%s was given encore", tc.name)
		}
	}
	if _, ok := r37EncoreCost(g, r37Dead(t, g, opp, "Their Sliver", "Creature — Sliver", "{2}")); ok {
		t.Error("a Sliver in an opponent's graveyard was given encore")
	}
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, sliver, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("encore {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := r37Tokens(g, "Dead Sliver", me.ID); n != len(g.Seats)-1 {
		t.Fatalf("%d token copies, want one per opponent", n)
	}
}

// "The legend rule doesn't apply to Slivers you control": two legendary
// Slivers of mine stay, two legendary non-Slivers do not.
func TestGapB101GravemotherSuspendsTheLegendRuleForMySlivers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Sliver Gravemother", "Legendary Creature — Sliver", gapB101GravemotherOracle, false)
	a := effectstest.B12Creature(g, me.ID, "Sliver Legend", "Legendary Creature — Sliver", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Sliver Legend", "Legendary Creature — Sliver", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Human Legend", "Legendary Creature — Human", 2, 2)
	d := effectstest.B12Creature(g, me.ID, "Human Legend", "Legendary Creature — Human", 2, 2)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("a legendary Sliver was put into the graveyard by the legend rule")
	}
	if effectstest.LegendPrompt(g, me.ID) == nil {
		t.Error("the legend rule did not apply to two non-Sliver legends")
	}
	_, _ = c, d
}

// Anikthea exiles a non-Aura enchantment card from my graveyard and makes
// a 3/3 black Zombie creature copy of it, on entry and when it attacks.
func TestGapB101AniktheaCopiesAnEnchantmentAsAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aura := effectstest.GyCard(me, "Test Aura", "Enchantment — Aura", "{1}{W}", "")
	ench := effectstest.GyCard(me, "Test Rune", "Enchantment", "{1}{W}", "")
	id := effectstest.CastCatalogSpell(t, g, "Anikthea, Hand of Erebos", "Legendary Enchantment Creature — Demigod", gapB101AniktheaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the entry trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: aura}); err == nil {
		t.Fatal("an Aura card was accepted as the target")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, ench)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ench) {
		t.Fatal("the enchantment card was not exiled")
	}
	var tok game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Test Rune" {
			tok = c
		}
	}
	if tok.InstanceID == uuid.Nil || !effects.IsToken(tok) {
		t.Fatal("no token copy of the enchantment")
	}
	live := effectstest.B12Card(t, g, tok.InstanceID)
	if !live.IsCreature() || !live.IsEnchantment() || !live.HasSubtype("Zombie") {
		t.Errorf("the copy is %q, want an enchantment creature Zombie", live.TypeLine)
	}
	if live.CurrentPower() != 3 || live.CurrentToughness() != 3 {
		t.Errorf("the copy is %d/%d, want 3/3", live.CurrentPower(), live.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "menace") {
		t.Error("an enchantment creature of mine lacks menace")
	}
	if !g.Battlefield.Contains(id) || !me.Graveyard.Contains(aura) {
		t.Error("Anikthea left or the Aura moved")
	}
}

// Karador: costs {1} less for each creature card in my graveyard, and one
// creature spell a turn may be cast from the graveyard.
func TestGapB101KaradorCostsLessAndCastsOneCreatureFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Karador, Ghost Chieftain", "Legendary Creature — Centaur Spirit", gapB101KaradorOracle, false)
	bear := effectstest.PtGraveCard(me, "Bear", "Creature — Bear")
	wolf := effectstest.PtGraveCard(me, "Wolf", "Creature — Wolf")
	rock := effectstest.PtGraveCard(me, "Rock", "Artifact")
	if err := effectstest.MuldrothaCast(g, me, rock, ""); err == nil {
		t.Error("a noncreature spell was cast from the graveyard")
	}
	if err := effectstest.MuldrothaCast(g, me, bear, ""); err != nil {
		t.Fatalf("a creature from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.MuldrothaCast(g, me, wolf, ""); err == nil {
		t.Error("a second creature spell was cast from the graveyard in one turn")
	}
}

func TestGapB101KaradorSelfCostReduction(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 5; i++ {
		effectstest.PtGraveCard(me, "Bear", "Creature — Bear")
	}
	effectstest.PtGraveCard(me, "Rock", "Artifact")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Karador, Ghost Chieftain", OracleID: gapB101KaradorOracle,
		TypeLine: "Legendary Creature — Centaur Spirit", ManaCost: "{5}{W}{B}{G}", Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID})
	// {5}{W}{B}{G} less five creature cards is {W}{B}{G}.
	effectstest.ApaMana(me, "W", "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast Karador for two mana with five creature cards in the graveyard")
	}
	effectstest.ApaMana(me, "G")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Karador for {W}{B}{G}: %v", err)
	}
}

// Narci: sacrificing an enchantment draws; a Saga's final chapter ability
// resolving drains each opponent for the Saga's mana value.
func TestGapB101NarciDrawsWhenYouSacrificeAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Narci, Fable Singer", "Legendary Creature — Human Bard", gapB101NarciOracle, 3, 3)
	ench := effectstest.B12Permanent(g, me.ID, "Test Rune", "Enchantment")
	rock := effectstest.B12Permanent(g, me.ID, "Test Rock", "Artifact")
	hand := me.Hand.Size()
	if err := g.SacrificePermanent(me.ID, rock); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatal("sacrificing an artifact drew a card")
	}
	if err := g.SacrificePermanent(me.ID, ench); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestGapB101NarciDrainsForTheSagasManaValueOnItsFinalChapter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Narci, Fable Singer", "Legendary Creature — Human Bard", gapB101NarciOracle, 3, 3)
	saga := r37Benalia(g, me, 1) // History of Benalia: {1}{W}{W}, three chapters
	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}
	r37Advance(t, g, saga) // chapter II: not the final chapter
	for _, p := range g.Seats {
		if p.Life != life[p.ID] {
			t.Fatalf("%s's life moved on chapter II", p.Name)
		}
	}
	r37Advance(t, g, saga) // chapter III
	effectstest.MonSettle(t, g)
	for _, p := range g.Seats {
		want := life[p.ID] - 3
		if p.ID == me.ID {
			want = life[p.ID] + 3
		}
		if p.Life != want {
			t.Errorf("%s's life = %d, want %d", p.Name, p.Life, want)
		}
	}
}

func TestGapB101NarciIgnoresAnOpponentsSaga(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Narci, Fable Singer", "Legendary Creature — Human Bard", gapB101NarciOracle, 3, 3)
	saga := r37Benalia(g, opp, 2)
	before := me.Life
	r37Advance(t, g, saga)
	if me.Life != before {
		t.Fatalf("my life went %d -> %d for an opponent's Saga", before, me.Life)
	}
}
