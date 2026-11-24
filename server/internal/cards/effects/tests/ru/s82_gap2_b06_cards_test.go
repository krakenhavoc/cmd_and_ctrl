package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b06_cards_test.go — S82 GAP2 batch 6 (#5948, seams #6089 and
// #6090): draw replacements and library-search restrictions. Each test
// plays the real catalog entry.

const (
	gap2b06Alms       = "278e51fe-3cd9-4c82-bd81-2164771f8611"
	gap2b06Abundance  = "eccc9a54-2b56-4a02-9926-258d5b2e25fb"
	gap2b06RiverSong  = "418ecbb7-b809-4639-a531-070b58b5b85e"
	gap2b06Jace       = "7ca99e2f-faf2-4301-a0ae-7cb880ba965b"
	gap2b06Mindcensor = "d9517c5d-66d0-4178-96fb-a8c04f311ad8"
	gap2b06Wraith     = "61849685-b407-4ed0-aafc-ec3a003200b9"
	gap2b06Spinner    = "8fea91ee-5849-4c46-8198-fd383d9a55e8"
	gap2b06Web        = "8b25f353-0798-4a03-a928-6a701a919468"
)

func TestGap2B06EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap2b06Alms:       "Alms Collector",
		gap2b06Abundance:  "Abundance",
		gap2b06RiverSong:  "River Song",
		gap2b06Jace:       "Jace, Mirror Mage",
		gap2b06Mindcensor: "Aven Mindcensor",
		gap2b06Wraith:     "Canoptek Wraith",
		gap2b06Spinner:    "Arachnus Spinner",
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

// gap2b06Library replaces p's library with cards, bottom first, and returns
// their IDs in the same order.
func gap2b06Library(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func gap2b06Basics(p *game.Player, names ...string) []uuid.UUID {
	cards := make([]game.Card, 0, len(names))
	for _, n := range names {
		cards = append(cards, game.Card{Name: n, TypeLine: "Basic Land — " + n})
	}
	return gap2b06Library(p, cards...)
}

func gap2b06Settle(t *testing.T, g *game.Game) {
	t.Helper()
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Alms Collector ---------------------------------------------------------------

func TestGap2B06AlmsCollectorRewritesAnOpponentsMultiCardDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.BfCardWithOracle(g, me, "Alms Collector", "Creature — Cat Cleric", gap2b06Alms)
	gap2b06Basics(me, "Plains", "Plains", "Plains", "Plains")
	gap2b06Basics(opp, "Island", "Island", "Island", "Island", "Island")

	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	effectstest.DrawN(g, opp, 3)
	if got := opp.Hand.Size() - theirs; got != 1 {
		t.Errorf("the opponent drew %d of three, want 1", got)
	}
	if got := me.Hand.Size() - mine; got != 1 {
		t.Errorf("Alms Collector's controller drew %d, want 1", got)
	}

	mine, theirs = me.Hand.Size(), opp.Hand.Size()
	effectstest.DrawN(g, opp, 1)
	effectstest.DrawN(g, me, 2)
	if got := opp.Hand.Size() - theirs; got != 1 {
		t.Errorf("a one-card draw was rewritten: the opponent drew %d", got)
	}
	if got := me.Hand.Size() - mine; got != 2 {
		t.Errorf("the controller's own draw two was rewritten: drew %d", got)
	}
}

// --- Abundance --------------------------------------------------------------------

func TestGap2B06AbundanceRevealsUntilTheChosenKindAndOrdersTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.BfCardWithOracle(g, me, "Abundance", "Enchantment", gap2b06Abundance)
	ids := gap2b06Library(me,
		game.Card{Name: "Deep", TypeLine: "Sorcery"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
	)
	_ = g.DrawCard(me.ID)
	effectstest.AnswerOptional(t, g, me, true)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // land
	if !me.Hand.Contains(ids[1]) {
		t.Fatal("the first land revealed is not in hand")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("the other revealed cards were not offered for ordering")
	}
	// Bolt, then Bear, from the top of the bottom pile: Bear is the bottom card.
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[3], ids[2]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if me.Library.Cards[0].InstanceID != ids[2] || me.Library.Cards[1].InstanceID != ids[3] {
		t.Error("the revealed cards are not on the bottom in the chosen order")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != ids[0] {
		t.Error("the unrevealed card is not on top")
	}
	if g.TurnTallyFor(me.ID).CardsDrawn != 0 {
		t.Error("the replaced draw counted as a draw")
	}
}

func TestGap2B06AbundanceDeclinedDrawsAndIgnoresOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	effectstest.BfCardWithOracle(g, me, "Abundance", "Enchantment", gap2b06Abundance)
	ids := gap2b06Basics(me, "Forest", "Island")
	_ = g.DrawCard(me.ID)
	effectstest.AnswerOptional(t, g, me, false)
	if !me.Hand.Contains(ids[1]) {
		t.Error("declining did not draw the top card")
	}
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("declining still asked land or nonland")
	}
	_ = g.DrawCard(other.ID)
	if effectstest.OptionalOffer(g, other) != nil || effectstest.OptionalOffer(g, me) != nil {
		t.Error("an opponent's draw was offered the replacement")
	}
}

// --- River Song -------------------------------------------------------------------

func TestGap2B06RiverSongDrawsFromTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "River Song", "Legendary Creature — Human Time Lord Rogue", gap2b06RiverSong, 2, 2)
	mine := gap2b06Basics(me, "Bottom", "Middle", "Top")
	theirs := gap2b06Basics(opp, "Bottom", "Middle", "Top")
	effectstest.DrawN(g, me, 1)
	effectstest.DrawN(g, opp, 1)
	if !me.Hand.Contains(mine[0]) || me.Hand.Contains(mine[2]) {
		t.Error("River Song's controller did not draw the bottom card")
	}
	if !opp.Hand.Contains(theirs[2]) {
		t.Error("an opponent did not draw from the top")
	}
}

func TestGap2B06RiverSongSpoilersPunishesScrySurveilAndSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	river := effectstest.B12Push(g, me.ID, "River Song", "Legendary Creature — Human Time Lord Rogue", gap2b06RiverSong, 2, 2)
	emit := func(ev game.Event) {
		g.WithWriteLock(func() { g.EmitEvent(ev) })
		gap2b06Settle(t, g)
	}
	life := opp.Life
	emit(game.Event{Kind: game.EventScry, Actor: opp.ID, LookedAt: 1})
	if c, _ := effectstest.BattlefieldCard(g, river); c.Counters[game.CounterPlusOne] != 1 {
		t.Fatalf("River Song has %d +1/+1 counters after a scry, want 1", c.Counters[game.CounterPlusOne])
	}
	if opp.Life != life-3 {
		t.Errorf("the scrying opponent lost %d, want 3 (power after the counter)", life-opp.Life)
	}
	life = opp.Life
	emit(game.Event{Kind: game.EventSurveil, Actor: opp.ID, LookedAt: 1})
	if opp.Life != life-4 {
		t.Errorf("a surveil cost %d, want 4", life-opp.Life)
	}
	life = opp.Life
	emit(game.Event{Kind: game.EventSearchLibrary, Actor: opp.ID, Target: opp.ID})
	if opp.Life != life-5 {
		t.Errorf("a search of their own library cost %d, want 5", life-opp.Life)
	}
	// Not their library, and not an opponent: nothing.
	life, myLife := opp.Life, me.Life
	emit(game.Event{Kind: game.EventSearchLibrary, Actor: opp.ID, Target: me.ID})
	emit(game.Event{Kind: game.EventScry, Actor: me.ID, LookedAt: 1})
	if opp.Life != life || me.Life != myLife {
		t.Error("a search of another library, or the controller's own scry, triggered Spoilers")
	}
	if c, _ := effectstest.BattlefieldCard(g, river); c.Counters[game.CounterPlusOne] != 3 {
		t.Errorf("River Song has %d counters, want 3", c.Counters[game.CounterPlusOne])
	}
}

// --- Jace, Mirror Mage ------------------------------------------------------------

func gap2b06CastJace(t *testing.T, g *game.Game, me *game.Player, kicked bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Jace, Mirror Mage", TypeLine: "Legendary Planeswalker — Jace",
		ManaCost: "{1}{U}{U}", OracleID: gap2b06Jace, Owner: me.ID, Controller: me.ID})
	var optional []int
	if kicked {
		optional = []int{0}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{OptionalCosts: optional}); err != nil {
		t.Fatalf("cast Jace: %v", err)
	}
	gap2b06Settle(t, g)
	gap2b06Settle(t, g)
	return id
}

func TestGap2B06JaceKickedMakesANonlegendaryCopyWithLoyaltyOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := gap2b06CastJace(t, g, me, true)
	if got := effectstest.LoyaltyOf(g, jace); got != 4 {
		t.Errorf("Jace has %d loyalty, want 4", got)
	}
	var copies []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Jace, Mirror Mage" {
			copies = append(copies, c)
		}
	}
	if len(copies) != 1 {
		t.Fatalf("%d token copies, want 1", len(copies))
	}
	tok := copies[0]
	if tok.HasSupertype("Legendary") || !tok.IsPlaneswalker() {
		t.Errorf("the copy's type line is %q", tok.TypeLine)
	}
	if got := tok.Counters[game.CounterLoyalty]; got != 1 {
		t.Errorf("the copy has %d loyalty, want 1", got)
	}
	if _, ok := effectstest.BattlefieldCard(g, jace); !ok {
		t.Error("the legend rule took a Jace: the copy is not legendary")
	}
}

func TestGap2B06JaceUnkickedMakesNoCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gap2b06CastJace(t, g, me, false)
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() {
			t.Fatal("an unkicked Jace made a copy")
		}
	}
}

func TestGap2B06JaceZeroRevealsAndRemovesTheManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := effectstest.S82Walker(g, me.ID, "Jace, Mirror Mage", gap2b06Jace, 4)
	ids := gap2b06Library(me,
		game.Card{Name: "Filler", TypeLine: "Instant", ManaCost: "{U}"},
		game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}"},
	)
	effectstest.AdvanceToMain(t, g)
	from := len(g.Events)
	effectstest.S82Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{})
	gap2b06Settle(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Fatal("the top card was not drawn")
	}
	if got := effectstest.LoyaltyOf(g, jace); got != 1 {
		t.Errorf("Jace has %d loyalty, want 1 (4, +0, then -3)", got)
	}
	revealed := false
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventRevealCards && ev.CardID == ids[1] {
			revealed = true
		}
	}
	if !revealed {
		t.Error("the drawn card was not revealed")
	}
}

func TestGap2B06JaceZeroCanRemoveHisLastLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := effectstest.S82Walker(g, me.ID, "Jace, Mirror Mage", gap2b06Jace, 2)
	gap2b06Library(me, game.Card{Name: "Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}{R}"})
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{})
	gap2b06Settle(t, g)
	if _, ok := effectstest.BattlefieldCard(g, jace); ok {
		t.Error("Jace survived losing more loyalty than he had")
	}
}

func TestGap2B06JacePlusOneScriesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := effectstest.S82Walker(g, me.ID, "Jace, Mirror Mage", gap2b06Jace, 4)
	gap2b06Basics(me, "a", "b", "c")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, jace); got != 5 {
		t.Errorf("Jace has %d loyalty after +1, want 5", got)
	}
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("no scry 2 prompt: %+v", c)
	}
}

// --- Aven Mindcensor --------------------------------------------------------------

func TestGap2B06AvenMindcensorNarrowsAnOpponentsSearchToTheTopFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Aven Mindcensor", "Creature — Bird Wizard", gap2b06Mindcensor, 2, 1)
	ids := gap2b06Basics(opp, "Forest", "Island", "Island", "Island", "Forest", "Island", "Island", "Island")
	search := func(who *game.Player) *game.PendingChoice {
		g.WithWriteLock(func() {
			if err := g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player: who.ID, Pred: func(c game.Card) bool { return c.Name == "Forest" },
				Dest: game.ZoneHand, Limit: 2, Shuffle: true,
			}); err != nil {
				t.Fatal(err)
			}
		})
		return effectstest.SearchPromptFor(g, who.ID)
	}
	c := search(opp)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ids[4] {
		t.Fatalf("the opponent's candidates are %v, want only the Forest in the top four", c.SearchCards)
	}
	if err := g.ResolveSearchLibrary(c.ID, opp.ID, []uuid.UUID{ids[0]}); err == nil {
		t.Error("the Forest below the top four was taken")
	}
	if err := g.ResolveSearchLibrary(c.ID, opp.ID, []uuid.UUID{ids[4]}); err != nil {
		t.Fatalf("take: %v", err)
	}
	if !opp.Hand.Contains(ids[4]) {
		t.Error("the Forest in the top four was not taken")
	}

	mine := gap2b06Basics(me, "Forest", "Island", "Island", "Island", "Island", "Island")
	c = search(me)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != mine[0] {
		t.Errorf("the controller's own search was narrowed: %v", c)
	}
}

// --- Canoptek Wraith --------------------------------------------------------------

func TestGap2B06CanoptekWraithPaysSacrificesAndFetchesTwoMatchingBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	wraith := effectstest.B12Push(g, me.ID, "Canoptek Wraith", "Artifact Creature — Wraith", gap2b06Wraith, 2, 1)
	island := effectstest.BfCardWithOracle(g, me, "Island", "Basic Land — Island", "")
	effectstest.BfCardWithOracle(g, me, "Forest", "Basic Land — Forest", "")
	ids := gap2b06Basics(me, "Island", "Forest", "Island", "Island", "Swamp")

	effectstest.DealCombatDamageToPlayer(g, wraith, opp.ID, 2)
	gap2b06Settle(t, g)
	effectstest.B06AddMana(me, "C", "C", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if _, ok := effectstest.BattlefieldCard(g, wraith); ok {
		t.Fatal("the Wraith was not sacrificed")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, island)
	c := effectstest.SearchPromptFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMax != 2 || len(c.SearchCards) != 3 {
		t.Fatalf("search offers %d of %d, want up to 2 of the 3 Islands", c.SearchMax, len(c.SearchCards))
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[1]}); err == nil {
		t.Error("a Forest was taken for a chosen Island")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0], ids[2]}); err != nil {
		t.Fatalf("take: %v", err)
	}
	for _, id := range []uuid.UUID{ids[0], ids[2]} {
		if tapped, ok := effectstest.BattlefieldCardTapped(g, id); !ok || !tapped {
			t.Error("a fetched Island is not on the battlefield tapped")
		}
	}
}

func TestGap2B06CanoptekWraithDeclinedKeepsTheWraith(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	wraith := effectstest.B12Push(g, me.ID, "Canoptek Wraith", "Artifact Creature — Wraith", gap2b06Wraith, 2, 1)
	effectstest.AssertRestrictions(t, g, wraith, game.CantBeBlocked)
	effectstest.DealCombatDamageToPlayer(g, wraith, opp.ID, 2)
	gap2b06Settle(t, g)
	effectstest.B06AddMana(me, "C", "C", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if _, ok := effectstest.BattlefieldCard(g, wraith); !ok {
		t.Error("declining sacrificed the Wraith")
	}
	if effectstest.SearchPromptFor(g, me.ID) != nil {
		t.Error("declining still searched")
	}
}

// --- Arachnus Spinner -------------------------------------------------------------

func TestGap2B06ArachnusSpinnerFetchesTheWebFromTheGraveyardOntoATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	spinner := effectstest.B12Push(g, me.ID, "Arachnus Spinner", "Creature — Spider", gap2b06Spinner, 5, 7)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	web := effectstest.GyCardWithOracle(me, "Arachnus Web", "Enchantment — Aura", gap2b06Web)
	gap2b06Basics(me, "Forest", "Forest")

	if err := g.ActivateCatalogAbility(me.ID, spinner, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.S82TargetCard(bear), TapIDs: []uuid.UUID{spinner},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, spinner); !tapped {
		t.Error("the Spinner was not tapped to pay the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchPromptFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ChooseSearchZones(c.ID, me.ID, []string{game.SearchZoneGraveyard}); err != nil {
		t.Fatalf("ChooseSearchZones: %v", err)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{web}); err != nil {
		t.Fatalf("take: %v", err)
	}
	w, ok := effectstest.BattlefieldCard(g, web)
	if !ok || !w.IsAttachedTo(bear) {
		t.Fatal("Arachnus Web is not on the battlefield attached to the target")
	}
}

func TestGap2B06ArachnusSpinnerNeedsAnUntappedSpider(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	spinner := effectstest.B12Push(g, me.ID, "Arachnus Spinner", "Creature — Spider", gap2b06Spinner, 5, 7)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	target := effectstest.B12Creature(g, opp.ID, "Target", "Creature — Elf", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, spinner, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.S82TargetCard(target), TapIDs: []uuid.UUID{bear},
	}); err == nil {
		t.Error("a Bear paid the Spider cost")
	}
}
