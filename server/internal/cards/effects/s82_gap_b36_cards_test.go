package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB36HeadsplitterOracle = "8333c82f-aae6-470f-a9db-63ab24f9bb46"
	gapB36FootbladeOracle    = "daed1b38-510b-4698-95c3-05c09f5fabe9"
	gapB36SpearOracle        = "dac3a263-cc02-446a-88ec-882a21096b6c"
	gapB36WarAxeOracle       = "d306de95-a80e-4e85-b7d2-36c6ea769b15"
	gapB36GauntletOracle     = "85d06533-3c0b-46a8-8fdb-379db9ae7e2e"
	gapB36HookbladeOracle    = "8df6c4e2-8b62-4c64-a1d2-f17ae9d02bac"
	gapB36TriadOracle        = "e0321049-a3ed-4378-b607-7fb190b6f0fd"
	gapB36InitiateOracle     = "23be5c68-6530-474d-b677-551da4745b3c"
	gapB36DesmondOracle      = "654d84bb-f801-4d03-898c-78952f1c5a73"
	gapB36JacobOracle        = "31af65c3-602b-4ac9-a25f-dccb88f8f751"
)

func TestGapB36EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB36HeadsplitterOracle: "Headsplitter",
		gapB36FootbladeOracle:    "Hidden Footblade",
		gapB36SpearOracle:        "The Spear of Leonidas",
		gapB36WarAxeOracle:       "Raven Clan War-Axe",
		gapB36GauntletOracle:     "Assassin Gauntlet",
		gapB36HookbladeOracle:    "Hookblade",
		gapB36TriadOracle:        "The Capitoline Triad",
		gapB36InitiateOracle:     "Assassin Initiate",
		gapB36DesmondOracle:      "Desmond Miles",
		gapB36JacobOracle:        "Jacob Frye",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB36Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func gapB36Battlefield(g *game.Game) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() { out = g.BattlefieldCardsForEffect() })
	return out
}

func gapB36Find(g *game.Game, name string) (game.Card, bool) {
	for _, c := range gapB36Battlefield(g) {
		if c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}

func TestGapB36HiddenFootbladeAttachesAndGivesFirstStrikeThenHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := gapB36Seats(g)
	c := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	blade := castCatalogSpell(t, g, "Hidden Footblade", "Artifact — Equipment", gapB36FootbladeOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, c)
	passPriorityAroundTable(t, g)
	if got := e2Card(t, g, blade).AttachedTo.ID; got != c {
		t.Fatalf("attached to %v, want the host", got)
	}
	if p, _ := ptOf(t, g, c); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	for _, kw := range []string{"first strike", "haste"} {
		if !effectiveAbilitiesContain(t, g, c, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

func TestGapB36HookbladeGivesFlyingOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := gapB36Seats(g)
	c := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	blade := castCatalogSpell(t, g, "Hookblade", "Artifact — Equipment", gapB36HookbladeOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, c)
	passPriorityAroundTable(t, g)
	if got := e2Card(t, g, blade).AttachedTo.ID; got != c {
		t.Fatalf("attached to %v, want the host", got)
	}
	if p, _ := ptOf(t, g, c); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	if !effectiveAbilitiesContain(t, g, c, "flying") {
		t.Error("no flying on my turn")
	}
	for i := 0; i < 80 && g.Turn.ActiveSeat == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == 0 {
		t.Fatal("the turn never passed")
	}
	if effectiveAbilitiesContain(t, g, c, "flying") {
		t.Error("flying during an opponent's turn")
	}
}

func TestGapB36HeadsplitterMakesAnAssassinAndAttaches(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := gapB36Seats(g)
	blade := castCatalogSpell(t, g, "Headsplitter", "Artifact — Equipment", gapB36HeadsplitterOracle, nil)
	passPriorityAroundTable(t, g)
	token, ok := gapB36Find(g, "Assassin")
	if !ok {
		t.Fatal("no Assassin token was created")
	}
	if token.Controller != me.ID || !token.HasSubtype("Assassin") {
		t.Errorf("token: controller %v, types %q", token.Controller, token.TypeLine)
	}
	if got := e2Card(t, g, blade).AttachedTo.ID; got != token.InstanceID {
		t.Fatalf("attached to %v, want the token", got)
	}
	if p, tough := ptOf(t, g, token.InstanceID); p != 2 || tough != 1 {
		t.Errorf("token is %d/%d, want 2/1", p, tough)
	}
	if !effectiveAbilitiesContain(t, g, token.InstanceID, "menace") {
		t.Error("the token has no menace")
	}
}

func TestGapB36AssassinGauntletAttachesTapsAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB36Seats(g)
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	mine := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	other := b12Creature(g, g.Seats[2].ID, "Other", "Creature — Bear", 2, 2)
	gauntlet := castCatalogSpell(t, g, "Assassin Gauntlet", "Artifact — Equipment", gapB36GauntletOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, mine)
	for i := 0; i < 3 && latestPickTarget(g, me.ID) != nil; i++ {
		p := latestPickTarget(g, me.ID)
		if len(p.PickTargetPlayers) > 0 {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID, Slot: 1}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
	}
	passPriorityAroundTable(t, g)
	if got := e2Card(t, g, gauntlet).AttachedTo.ID; got != mine {
		t.Fatalf("attached to %v, want the host", got)
	}
	if !e2Card(t, g, theirs).Tapped {
		t.Error("the targeted opponent's creature is untapped")
	}
	if e2Card(t, g, other).Tapped {
		t.Error("another opponent's creature was tapped")
	}
	if p, tough := ptOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("equipped creature is %d/%d, want 3/3", p, tough)
	}
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, mine, opp.ID, 3)
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want the same after a loot", hand, me.Hand.Size())
	}
}

func TestGapB36AssassinInitiateGainsTheChosenKeyword(t *testing.T) {
	for i, kw := range []string{"flying", "deathtouch", "lifelink"} {
		g := newCatalogGame(t)
		toMain(t, g)
		me, _ := gapB36Seats(g)
		id := pushCatalogPermanent(g, me.ID, "Assassin Initiate", "Creature — Human Assassin", gapB36InitiateOracle, false)
		fillPool(me, 1)
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		c := latestOptionPick(g, me.ID)
		if c == nil {
			t.Fatal("no keyword question")
		}
		if err := g.ResolveOptionPick(c.ID, me.ID, i); err != nil {
			t.Fatalf("ResolveOptionPick: %v", err)
		}
		for _, other := range []string{"flying", "deathtouch", "lifelink"} {
			if got := effectiveAbilitiesContain(t, g, id, other); got != (other == kw) {
				t.Errorf("choice %q: has %s = %v", kw, other, got)
			}
		}
	}
}

func latestOptionPick(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func TestGapB36DesmondMilesCountsOtherAssassinsAndSurveilsForTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB36Seats(g)
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	desmond := pushCatalogPermanent(g, me.ID, "Desmond Miles", "Legendary Creature — Human Assassin", gapB36DesmondOracle, false)
	g.WithWriteLock(func() {
		if c := findBattlefieldForTest35(g, desmond); c != nil {
			c.Power, c.Toughness = 1, 3
		}
	})
	if p, _ := ptOf(t, g, desmond); p != 1 {
		t.Fatalf("alone, power %d, want 1", p)
	}
	b12Creature(g, me.ID, "Friend", "Creature — Human Assassin", 1, 1)
	b12Creature(g, opp.ID, "Their Assassin", "Creature — Human Assassin", 1, 1)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Assassin", TypeLine: "Creature — Human Assassin", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if p, _ := ptOf(t, g, desmond); p != 3 {
		t.Errorf("with one other Assassin and one in the graveyard, power %d, want 3", p)
	}
	dealCombatDamageToPlayer(g, desmond, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if c := latestChooseCardsFor(g, me.ID); c == nil && !hasSurveilPrompt(g, me.ID) {
		t.Error("no surveil prompt after combat damage")
	}
}

func hasSurveilPrompt(g *game.Game, who uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == who && c.Kind == game.PendingChoiceSurveil {
			return true
		}
	}
	return false
}

func findBattlefieldForTest35(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func TestGapB36SpearOfLeonidasOffersThreeModes(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me, opp := gapB36Seats(g)
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	bear := pushCrewerForTest(g, me.ID, "Bear", 2)
	spear := pushCatalogPermanent(g, me.ID, "The Spear of Leonidas", "Legendary Artifact — Equipment", gapB36SpearOracle, false)
	g.WithWriteLock(func() { _ = g.AttachForEffect(spear, game.TargetRef{Kind: game.TargetCard, ID: bear}) })
	declareAttack(t, g, opp.ID, bear)
	mode := latestModePick(g, me.ID)
	if mode == nil {
		t.Fatal("no mode prompt on attack")
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, bear, "double strike") {
		t.Error("Bull Rush: no double strike")
	}
}

func latestModePick(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceModePick && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func TestGapB36SpearOfLeonidasSummonsPhobosAndRevelationDiscardsThenDraws(t *testing.T) {
	for mode := 1; mode <= 2; mode++ {
		g := newCatalogGame(t)
		toMain(t, g)
		me, opp := gapB36Seats(g)
		for i := 0; i < 4; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		}
		bear := pushCrewerForTest(g, me.ID, "Bear", 2)
		spear := pushCatalogPermanent(g, me.ID, "The Spear of Leonidas", "Legendary Artifact — Equipment", gapB36SpearOracle, false)
		g.WithWriteLock(func() { _ = g.AttachForEffect(spear, game.TargetRef{Kind: game.TargetCard, ID: bear}) })
		hand := me.Hand.Size()
		declareAttack(t, g, opp.ID, bear)
		mp := latestModePick(g, me.ID)
		if mp == nil {
			t.Fatal("no mode prompt on attack")
		}
		if err := g.ResolveModePick(mp.ID, me.ID, []int{mode}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		passPriorityAroundTable(t, g)
		if mode == 1 {
			phobos, ok := gapB36Find(g, "Phobos")
			if !ok || phobos.Controller != me.ID || !phobos.IsLegendary() {
				t.Errorf("Phobos token missing or wrong: %+v", phobos)
			}
			continue
		}
		discardFromHand(t, g, me.ID)
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size(); got != hand {
			t.Errorf("hand %d → %d, want the same (discard two, draw two)", hand, got)
		}
	}
}

func TestGapB36RavenClanWarAxeSearchesForEivor(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := gapB36Seats(g)
	seedSearchLibrary(me,
		game.Card{Name: "Eivor, Battle-Ready", TypeLine: "Legendary Creature — Human Warrior"},
		game.Card{Name: "Someone Else", TypeLine: "Creature — Bear"},
	)
	castCatalogSpell(t, g, "Raven Clan War-Axe", "Artifact — Equipment", gapB36WarAxeOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Someone Else") != uuid.Nil {
		t.Error("a card with another name was offered")
	}
	answerSearchNamed(t, g, me.ID, "Eivor, Battle-Ready")
	if !b02bHandHasNamed(me, "Eivor, Battle-Ready") {
		t.Error("Eivor did not reach the hand")
	}
}

func TestGapB36CapitolineTriadEmblemSetsBasePowerAndToughness(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me, opp := gapB36Seats(g)
	triad := pushCatalogPermanent(g, me.ID, "The Capitoline Triad", "Legendary Creature — God Artificer", gapB36TriadOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	a := gyCard(me, "Big Artifact A", "Artifact", "{10}", "")
	b := gyCard(me, "Big Artifact B", "Legendary Artifact", "{10}", "")
	c := gyCard(me, "Big Artifact C", "Artifact", "{9}", "")
	plain := gyCard(me, "Plain Creature", "Creature — Bear", "{10}", "")
	d := gyCard(me, "Tiny Artifact", "Artifact", "{1}", "")
	if err := g.ActivateCatalogAbility(me.ID, triad, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{a, b, c, plain}}); err == nil {
		t.Fatal("exiled a nonhistoric card to pay")
	}
	if err := g.ActivateCatalogAbility(me.ID, triad, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{a, b, c}}); err == nil {
		t.Fatal("paid with total mana value 29")
	}
	if err := g.ActivateCatalogAbility(me.ID, triad, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{a, b, c, d}}); err != nil {
		t.Fatalf("activate with total mana value 30: %v", err)
	}
	for _, id := range []uuid.UUID{a, b, c, d} {
		if inZone(me.Graveyard, id) {
			t.Error("a paid card is still in the graveyard")
		}
	}
	passPriorityAroundTable(t, g)
	if len(emblemsOf(g, me.ID)) != 1 {
		t.Fatalf("%d emblems, want 1", len(emblemsOf(g, me.ID)))
	}
	if p, tough := ptOf(t, g, bear); p != 9 || tough != 9 {
		t.Errorf("my creature is %d/%d, want 9/9", p, tough)
	}
	if p, _ := ptOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

func TestGapB36JacobFryeExilesAnAssassinCardAndCopiesIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := gapB36Seats(g)
	jacob := pushCatalogPermanent(g, me.ID, "Jacob Frye", "Legendary Creature — Human Assassin", gapB36JacobOracle, false)
	dead := gyCard(me, "Dead Assassin", "Creature — Human Assassin", "{1}{B}", "")
	bear := gyCard(me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	dealCombatDamageToPlayer(g, jacob, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, bear) {
			t.Error("a nonassassin card was offered")
		}
		pickCard(t, g, me.ID, dead)
	}
	passPriorityAroundTable(t, g)
	if inZone(me.Graveyard, dead) {
		t.Error("the Assassin card is still in the graveyard")
	}
	if len(cardCopies(g)) == 0 && latestChooseCardsFor(g, me.ID) == nil && !hasAnyPending(g, me.ID) {
		t.Error("no copy was made or offered")
	}
}

func hasAnyPending(g *game.Game, who uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == who {
			return true
		}
	}
	return false
}
