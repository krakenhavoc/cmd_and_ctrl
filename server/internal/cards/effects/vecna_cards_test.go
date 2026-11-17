package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vecna_cards_test.go — #5683: The Book of Vile Darkness, Eye of Vecna,
// Hand of Vecna, and the seam under the Book's token: a cost that exiles
// permanents by name (game.AbilityCost.ExileNamed) and a token whose text
// is "all triggered abilities of the exiled cards" (game.TriggersOfKey).
// Every test plays the cards through the engine.

const (
	vecnaBookOracle = "fcc666bc-6fea-44e0-94bc-462c742db528"
	vecnaEyeOracle  = "371a7bbf-5890-46d8-ae57-f020c103b465"
	vecnaHandOracle = "3b7f0dd3-5b4c-4ac0-a43a-ddc62a44663e"

	vecnaBookType = "Legendary Artifact — Book"
	vecnaEyeType  = "Legendary Artifact"
	vecnaHandType = "Legendary Artifact — Equipment"
)

// vecnaTable is a game in the active seat's main phase with the Book
// under that seat, and the Eye and the Hand too when asked.
func vecnaTable(t *testing.T, eye, hand bool) (g *game.Game, me *game.Player, book, eyeID, handID uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	book = gaPut(g, me.ID, "The Book of Vile Darkness", vecnaBookType, vecnaBookOracle)
	if eye {
		eyeID = gaPut(g, me.ID, "Eye of Vecna", vecnaEyeType, vecnaEyeOracle)
	}
	if hand {
		handID = gaPut(g, me.ID, "Hand of Vecna", vecnaHandType, vecnaHandOracle)
	}
	return g, me, book, eyeID, handID
}

// vecnaActivateBook activates the Book's ability and returns the Vecna
// it makes once the ability has resolved. Triggers Vecna's arrival
// queued are still waiting.
func vecnaActivateBook(t *testing.T, g *game.Game, me *game.Player, book uuid.UUID) uuid.UUID {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the Book: %v", err)
	}
	for i := 0; i < 8 && findBattlefieldByName(g, "Vecna") == uuid.Nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	vecna := findBattlefieldByName(g, "Vecna")
	if vecna == uuid.Nil {
		t.Fatal("the Book's ability resolved and made no Vecna")
	}
	return vecna
}

func vecnaInExile(g *game.Game, id uuid.UUID) bool {
	var in bool
	g.ReadSnapshot(func() {
		z := g.FindCardZoneForEffect(id)
		in = z != nil && z.Kind == game.ZoneExile
	})
	return in
}

func vecnaHandOfSize(p *game.Player, n int) {
	for p.Hand.Size() > n {
		_, _ = p.Hand.PopTop()
	}
	for p.Hand.Size() < n {
		putInHand(p, game.NewCard("basic-filler", uuid.Nil))
	}
}

// --- the seam: the cost ---------------------------------------------------

func TestBookOfVileDarknessExilesItselfAndEveryArtifactYouControlWithEitherName(t *testing.T) {
	g, me, book, eye, hand := vecnaTable(t, true, true)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirEye := gaPut(g, opp.ID, "Eye of Vecna", vecnaEyeType, vecnaEyeOracle)
	// Not an artifact, whatever it is called.
	impostor := gaPut(g, me.ID, "Hand of Vecna", "Creature — Shapeshifter", "")
	bystander := gaPut(g, me.ID, "Sol Ring", "Artifact", "")

	var set []uuid.UUID
	g.ReadSnapshot(func() {
		set = g.ExileNamedSetForEffect(me.ID, book, &game.ExileNamedCost{CardType: "artifact", Names: []string{"Eye of Vecna", "Hand of Vecna"}})
	})
	if !slices.Equal(set, []uuid.UUID{eye, hand}) {
		t.Fatalf("the cost names %v, want the Eye and the Hand", set)
	}

	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the Book: %v", err)
	}
	// Paid as the ability is activated (CR 602.2b), before it resolves.
	for name, id := range map[string]uuid.UUID{"the Book": book, "the Eye": eye, "the Hand": hand} {
		if !vecnaInExile(g, id) {
			t.Errorf("%s was not exiled as a cost", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"an opponent's Eye": theirEye, "a creature named Hand of Vecna": impostor, "another artifact": bystander} {
		if _, ok := battlefieldCard(g, id); !ok {
			t.Errorf("%s was exiled; the cost names only your artifacts with those names", name)
		}
	}
	var item *game.StackItem
	for _, it := range g.StackMeta {
		if it != nil && it.SourceCardID == book {
			item = it
		}
	}
	if item == nil {
		t.Fatal("no ability on the stack")
	}
	if !slices.Equal(item.Paid.Exiled, []uuid.UUID{eye, hand}) {
		t.Errorf("the payment recorded %v as exiled, want the Eye and the Hand", item.Paid.Exiled)
	}
}

func TestBookOfVileDarknessCannotBeActivatedTapped(t *testing.T) {
	g, me, book, eye, _ := vecnaTable(t, true, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == book {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{})
	if err == nil {
		t.Fatal("a tapped Book paid {T}")
	}
	if errors.Is(err, game.ErrCardNotFound) {
		t.Fatalf("refused for the wrong reason: %v", err)
	}
	if vecnaInExile(g, book) || vecnaInExile(g, eye) {
		t.Error("a refused activation exiled something")
	}
}

// --- the seam: the token --------------------------------------------------

func TestVecnaHasAllTriggeredAbilitiesOfTheExiledCardsAndNoOthers(t *testing.T) {
	g, me, book, _, _ := vecnaTable(t, true, true)
	vecnaHandOfSize(me, 3)
	hand, life := me.Hand.Size(), me.Life
	vecna := vecnaActivateBook(t, g, me, book)

	c := gaLayered(t, g, vecna)
	if !c.IsToken() || c.Name != "Vecna" || c.Power != 8 || c.Toughness != 8 || !c.HasColor("B") || len(c.Colors) != 1 {
		t.Errorf("Vecna is %q %d/%d %v token=%v, want a black 8/8 token named Vecna", c.Name, c.Power, c.Toughness, c.Colors, c.IsToken())
	}
	for _, typ := range []string{"Zombie", "God"} {
		if !c.HasSubtype(typ) {
			t.Errorf("Vecna is not a %s", typ)
		}
	}
	if !c.IsCreature() || !c.IsLegendary() {
		t.Errorf("Vecna's type line is %q, want a legendary creature", c.TypeLine)
	}
	assertKeywords(t, g, vecna, "indestructible")
	want := []string{game.TriggersOfKey(vecnaBookOracle), game.TriggersOfKey(vecnaEyeOracle), game.TriggersOfKey(vecnaHandOracle)}
	if !slices.Equal(c.GrantedAbilities, want) {
		t.Errorf("Vecna's text carries %v, want %v", c.GrantedAbilities, want)
	}
	// "All TRIGGERED abilities": not the Book's own activated ability,
	// nor the Hand's two equips.
	if rows := game.ActivatedAbilitiesForCard(c); len(rows) != 0 {
		t.Errorf("Vecna has %d activated abilities, want none", len(rows))
	}
	if n := len(game.TriggersForCard(c)); n != 4 {
		t.Errorf("Vecna has %d triggered abilities, want the Book's one, the Eye's two and the Hand's one", n)
	}

	// The Eye's "When this enters" is Vecna's now, and Vecna just entered.
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-2 {
		t.Fatalf("hand %d → %d, life %d → %d: want Vecna's arrival to draw a card and lose 2 life", hand, me.Hand.Size(), life, me.Life)
	}

	// The Hand's: at the beginning of combat on your turn, a creature you
	// control named Vecna gets +X/+X.
	vecnaHandOfSize(me, 4)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, vecna); c.Effective().Power != 12 || c.Effective().Toughness != 12 {
		t.Errorf("Vecna is %d/%d at combat with four cards in hand, want 12/12", c.Effective().Power, c.Effective().Toughness)
	}

	// The Book's: at the beginning of your end step, if you lost 2 or
	// more life this turn (the Eye's 2), a Zombie.
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Zombie") == uuid.Nil {
		t.Error("Vecna's end-step ability made no Zombie after 2 life was lost this turn")
	}

	// The Eye's other one: at the beginning of your upkeep, you may pay {2}.
	seat := g.Turn.ActiveSeat
	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("Vecna's upkeep ability did not offer to pay {2}")
	}
	hand, life = me.Hand.Size(), me.Life
	b06AddMana(me, "C", "C")
	answerPayUnless(t, g, me.ID, true)
	if me.Hand.Size() != hand+1 || me.Life != life-2 {
		t.Errorf("hand %d → %d, life %d → %d: want the paid upkeep ability to draw a card and lose 2 life", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestVecnaFromTheBookAloneHasOnlyTheBooksTriggeredAbility(t *testing.T) {
	g, me, book, _, _ := vecnaTable(t, false, false)
	hand, life := me.Hand.Size(), me.Life
	vecna := vecnaActivateBook(t, g, me, book)
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, vecna)
	if want := []string{game.TriggersOfKey(vecnaBookOracle)}; !slices.Equal(c.GrantedAbilities, want) {
		t.Errorf("Vecna's text carries %v, want %v", c.GrantedAbilities, want)
	}
	if n := len(game.TriggersForCard(c)); n != 1 {
		t.Errorf("Vecna has %d triggered abilities, want the Book's one", n)
	}
	if me.Hand.Size() != hand || me.Life != life {
		t.Error("a Vecna made without the Eye drew or lost life on arrival")
	}
}

func TestVecnaDoesNotGetTheAbilitiesOfAnExiledTokenOrACardThatLeftExile(t *testing.T) {
	g, me, book, eye, _ := vecnaTable(t, true, false)
	// A token copy of the Hand is an artifact named Hand of Vecna: the
	// cost exiles it, and it ceases to exist (CR 111.7).
	var token uuid.UUID
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(me.ID, game.Card{
			Name: "Hand of Vecna", TypeLine: "Token " + vecnaHandType, OracleID: vecnaHandOracle,
		}, 1, game.TokenEntryOptions{})
		if err != nil || len(ids) != 1 {
			t.Fatalf("token Hand: %v %v", ids, err)
		}
		token = ids[0]
	})
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the Book: %v", err)
	}
	if _, ok := battlefieldCard(g, token); ok {
		t.Fatal("the token Hand of Vecna was not exiled by the cost")
	}
	// In response, the exiled Eye leaves exile: it is no longer one of
	// the exiled cards when the ability resolves.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, eye); err != nil {
		t.Fatalf("move the Eye out of exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	vecna := findBattlefieldByName(g, "Vecna")
	if vecna == uuid.Nil {
		t.Fatal("no Vecna")
	}
	if got, want := gaLayered(t, g, vecna).GrantedAbilities, []string{game.TriggersOfKey(vecnaBookOracle)}; !slices.Equal(got, want) {
		t.Errorf("Vecna's text carries %v, want only the Book's %v", got, want)
	}
}

func TestVecnasAbilitiesAreItsCopiableTextAndGoWithItsOtherAbilities(t *testing.T) {
	g, me, book, _, _ := vecnaTable(t, true, true)
	vecna := vecnaActivateBook(t, g, me, book)
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, vecna)

	// CR 111.3, CR 707.2: what the creating effect gave the token is
	// what a copy of it copies.
	if got := game.CopiableValuesOf(c).GrantedAbilities; !slices.Equal(got, c.GrantedAbilities) || len(got) != 3 {
		t.Errorf("a copy of Vecna would carry %v, want Vecna's %v", got, c.GrantedAbilities)
	}

	// CR 613.1f: an effect that removes all abilities removes these too,
	// whatever its timestamp, because they are the token's own.
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		if err := (ScopedEffectFor{
			Target:   vecna,
			Mods:     []game.Mod{{Kind: game.ModLoseAllAbilities}},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "test: loses all abilities",
		}).Apply(ctx); err != nil {
			t.Fatalf("remove abilities: %v", err)
		}
	})
	if n := len(game.TriggersForCard(gaLayered(t, g, vecna))); n != 0 {
		t.Errorf("Vecna has %d triggered abilities after losing all abilities, want none", n)
	}
}

func TestVecnasTriggeredAbilitiesSurviveASnapshotRoundTrip(t *testing.T) {
	g, me, book, _, _ := vecnaTable(t, true, true)
	hand, life := me.Hand.Size(), me.Life
	vecna := vecnaActivateBook(t, g, me, book)
	// The Eye's enters trigger is waiting: restore with it in flight.
	if kohQueuedTrigger(g, vecna) == nil {
		t.Fatal("Vecna's arrival queued no trigger")
	}
	restored := restoreThroughJSON(t, g)
	rMe := restored.Seats[g.Turn.ActiveSeat]
	if kohQueuedTrigger(restored, vecna) == nil {
		t.Fatal("the restored game lost Vecna's enters trigger")
	}
	passPriorityAroundTable(t, restored)
	if rMe.Hand.Size() != hand+1 || rMe.Life != life-2 {
		t.Fatalf("restored: hand %d → %d, life %d → %d, want the trigger to resolve", hand, rMe.Hand.Size(), life, rMe.Life)
	}
	// And the restored token still has its text: the Hand's combat
	// ability fires for it.
	vecnaHandOfSize(rMe, 2)
	advanceTo(t, restored, game.StepBeginCombat)
	passPriorityAroundTable(t, restored)
	if c := gaLayered(t, restored, vecna); c.Effective().Power != 10 {
		t.Errorf("restored Vecna has power %d at combat with two cards in hand, want 10", c.Effective().Power)
	}
}

func TestTriggeredTextOfCarriesKeywordTriggersWithoutTheirOtherHalves(t *testing.T) {
	// A triggered keyword that is only a triggered ability is the keyword;
	// one that is also something else lends its trigger alone.
	lender := game.Card{Name: "Test Lender", TypeLine: "Artifact Creature — Construct", Keywords: []string{"prowess", "modular 2", "decayed", "flying"}}
	parts, keywords := game.TriggeredTextOf(lender)
	if !slices.Equal(keywords, []string{"prowess"}) {
		t.Errorf("keywords %v, want prowess alone", keywords)
	}
	if want := []string{game.KeywordTriggerKey("modular 2"), game.KeywordTriggerKey("decayed")}; !slices.Equal(parts, want) {
		t.Fatalf("parts %v, want %v", parts, want)
	}
	for _, p := range parts {
		c := game.Card{Name: "Test Token", TypeLine: "Token Creature — Zombie", GrantedAbilities: []string{p}}
		if n := len(game.CatalogTriggers(game.CatalogAbilityKey(c))); n != 1 {
			t.Errorf("%s resolves to %d triggered rows, want 1", p, n)
		}
	}
	// A face-down card has no abilities to lend (CR 708.2a).
	lender.FaceDown = true
	if parts, keywords := game.TriggeredTextOf(lender); len(parts)+len(keywords) != 0 {
		t.Errorf("a face-down card lent %v %v", parts, keywords)
	}
	// A card with no triggered ability adds nothing.
	if parts, _ := game.TriggeredTextOf(game.Card{OracleID: b22DawnOfHopeOracle}); len(parts) != 1 {
		t.Errorf("Dawn of Hope lends %v, want its one part", parts)
	}
	if parts, _ := game.TriggeredTextOf(game.Card{OracleID: "71c52bf5-2a5d-488e-8b15-7ef290e4b77d"}); len(parts) != 0 {
		t.Errorf("War Room, which has no triggered ability, lent %v", parts)
	}
}

// --- The Book of Vile Darkness's own end step -----------------------------

func TestBookOfVileDarknessMakesAZombieOnlyAfterTwoLifeLostThisTurn(t *testing.T) {
	g, me, _, _, _ := vecnaTable(t, false, false)
	b14Gain(t, g, me.ID, -1)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Zombie") != uuid.Nil {
		t.Fatal("one life lost made a Zombie")
	}

	seat := g.Turn.ActiveSeat
	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	// On another player's turn: not "your end step".
	b14Gain(t, g, me.ID, -3)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Zombie") != uuid.Nil {
		t.Fatal("an opponent's end step made a Zombie")
	}

	advanceToUpkeepOf(t, g, seat)
	advanceToMain(t, g)
	b14Gain(t, g, me.ID, -1)
	b14Gain(t, g, me.ID, -1)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	z := findBattlefieldByName(g, "Zombie")
	if z == uuid.Nil {
		t.Fatal("two life lost this turn made no Zombie at the end step")
	}
	if c := gaLayered(t, g, z); c.Power != 2 || c.Toughness != 2 || !c.HasColor("B") || !c.IsToken() || c.Controller != me.ID {
		t.Errorf("the Zombie is %d/%d %v, want a 2/2 black token under its maker", c.Power, c.Toughness, c.Colors)
	}
}

// --- Eye of Vecna ---------------------------------------------------------

func TestEyeOfVecnaDrawsAndCostsTwoLifeOnEnteringAndOnAPaidUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand, life := me.Hand.Size(), me.Life
	rfdCast(t, g, "Eye of Vecna", vecnaEyeType, vecnaEyeOracle, 0, 0)
	if me.Hand.Size() != hand+1 || me.Life != life-2 {
		t.Fatalf("hand %d → %d, life %d → %d: want the Eye entering to draw a card and lose 2 life", hand, me.Hand.Size(), life, me.Life)
	}

	seat := g.Turn.ActiveSeat
	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("the Eye asked on another player's upkeep")
	}
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("the Eye did not offer to pay {2} on its controller's upkeep")
	}
	hand, life = me.Hand.Size(), me.Life
	answerPayUnless(t, g, me.ID, false)
	if me.Hand.Size() != hand || me.Life != life {
		t.Fatal("declining to pay drew a card or lost life")
	}

	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	hand, life = me.Hand.Size(), me.Life
	b06AddMana(me, "C", "C")
	answerPayUnless(t, g, me.ID, true)
	if me.Hand.Size() != hand+1 || me.Life != life-2 {
		t.Errorf("hand %d → %d, life %d → %d: want the paid upkeep to draw a card and lose 2 life", hand, me.Hand.Size(), life, me.Life)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left after paying {2}", n)
	}
}

// --- Hand of Vecna --------------------------------------------------------

func TestHandOfVecnaEquipsForOneLifePerCardInHandOrForTwoMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand := gaPut(g, me.ID, "Hand of Vecna", vecnaHandType, vecnaHandOracle)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Ox", "Creature — Ox", 1, 4)
	target := func(id uuid.UUID) game.ActivateAbilityParams {
		return game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
	}

	vecnaHandOfSize(me, 3)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, hand, 0, target(bear)); err != nil {
		t.Fatalf("Equip—Pay 1 life for each card in your hand: %v", err)
	}
	if me.Life != life-3 {
		t.Fatalf("life %d → %d with three cards in hand, want 3 paid", life, me.Life)
	}
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, hand); c.AttachedTo.ID != bear {
		t.Fatal("the Hand is not attached to the Bear")
	}

	// A player can't pay more life than they have (CR 119.4).
	me.Life = 2
	if err := g.ActivateCatalogAbility(me.ID, hand, 0, target(other)); err == nil {
		t.Fatal("paid 3 life at 2 life")
	}
	if me.Life != 2 {
		t.Fatalf("a refused equip changed life to %d", me.Life)
	}
	// An empty hand makes it free.
	vecnaHandOfSize(me, 0)
	if err := g.ActivateCatalogAbility(me.ID, hand, 0, target(other)); err != nil {
		t.Fatalf("equip with an empty hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != 2 || gaLayered(t, g, hand).AttachedTo.ID != other {
		t.Fatalf("life %d, attached to the Ox %v: want a free equip onto the Ox", me.Life, gaLayered(t, g, hand).AttachedTo.ID == other)
	}

	// Equip {2}: no mana, no equip.
	strict := target(bear)
	strict.Strict = true
	if err := g.ActivateCatalogAbility(me.ID, hand, 1, strict); err == nil {
		t.Fatal("Equip {2} was activated with an empty pool")
	}
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, hand, 1, target(bear)); err != nil {
		t.Fatalf("Equip {2}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if gaLayered(t, g, hand).AttachedTo.ID != bear {
		t.Error("Equip {2} did not move the Hand to the Bear")
	}
}

func TestHandOfVecnaPumpsTheEquippedCreatureAtCombatOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand := gaPut(g, me.ID, "Hand of Vecna", vecnaHandType, vecnaHandOracle)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, hand, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("Equip {2}: %v", err)
	}
	passPriorityAroundTable(t, g)

	vecnaHandOfSize(me, 5)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if latestOptionPickFor(g, me.ID) != nil {
		t.Fatal("one creature to pump still asked a question")
	}
	if c := gaLayered(t, g, bear); c.Effective().Power != 7 || c.Effective().Toughness != 7 {
		t.Fatalf("the Bear is %d/%d with five cards in hand, want 7/7", c.Effective().Power, c.Effective().Toughness)
	}

	// Until end of turn, and not at another player's combat.
	seat := g.Turn.ActiveSeat
	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, bear); c.Effective().Power != 2 {
		t.Errorf("the Bear has power %d at an opponent's combat, want 2", c.Effective().Power)
	}
}

func TestHandOfVecnaAsksWhichCreatureWhenAVecnaIsAlsoThere(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	hand := gaPut(g, me.ID, "Hand of Vecna", vecnaHandType, vecnaHandOracle)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Vecna", "Legendary Creature — Zombie God", 8, 8)
	// Someone else's Vecna is not "a creature you control named Vecna".
	b12Creature(g, opp.ID, "Vecna", "Legendary Creature — Zombie God", 8, 8)
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, hand, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("Equip {2}: %v", err)
	}
	passPriorityAroundTable(t, g)

	vecnaHandOfSize(me, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	ask := latestOptionPickFor(g, me.ID)
	if ask == nil {
		t.Fatal("two creatures could get the bonus and nobody was asked")
	}
	if len(ask.PickOptions) != 2 {
		t.Fatalf("%d options, want the equipped creature and your Vecna", len(ask.PickOptions))
	}
	if err := g.ResolveOptionPick(ask.ID, me.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, mine); c.Effective().Power != 10 || c.Effective().Toughness != 10 {
		t.Errorf("the chosen Vecna is %d/%d with two cards in hand, want 10/10", c.Effective().Power, c.Effective().Toughness)
	}
	if c := gaLayered(t, g, bear); c.Effective().Power != 2 {
		t.Errorf("the Bear has power %d, want it left alone", c.Effective().Power)
	}
}
