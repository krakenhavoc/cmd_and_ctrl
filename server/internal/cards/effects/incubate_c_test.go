package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// incubate_c_test.go — the incubate cards that needed small pieces
// (#3568, slice inc-c): the repeat on Incubate, the Incubator-token
// predicate, "transform all Incubator tokens", and the seven cards on
// them.

const (
	eleshNornEtchingsOracle  = "650ed75c-20b0-45b8-a4d4-813e0369aaf1"
	elvishVatkeeperOracle    = "b881d61c-d7cf-49af-bed7-3bc657ea1208"
	progenitorExarchOracle   = "4052f710-6ba7-406f-8fb7-653a684a80e0"
	glissaHeraldOracle       = "9b10e0f7-9414-4769-85d7-c1cd1c74a9a6"
	glisteningDawnOracle     = "13ee85c3-a16a-41ac-bbaa-c7549c700a03"
	nornsInquisitorOracle    = "89a59a6d-e9d4-405c-8518-4da86dae30d1"
	sunderTheGatewayOracle   = "88d49309-e7d7-486d-8696-cd66bec2fe33"
	incCPhyrexianCreatureTyp = "Creature — Phyrexian Test"
)

// makeIncubator puts an Incubator with n +1/+1 counters under owner's
// control through the real creation path.
func makeIncubator(t *testing.T, g *game.Game, owner uuid.UUID, n int) uuid.UUID {
	t.Helper()
	opts := game.TokenEntryOptions{}
	if n > 0 {
		opts.Counters = map[string]int{"+1/+1": n}
	}
	var ids []uuid.UUID
	var err error
	g.WithWriteLock(func() {
		ids, err = g.CreateTokensForEffect(owner, IncubatorToken(), 1, opts)
	})
	if err != nil || len(ids) != 1 {
		t.Fatalf("make an Incubator: %v (%d tokens)", err, len(ids))
	}
	return ids[0]
}

// pushLands puts n basic lands under owner's control.
func pushLands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: owner, Controller: owner,
		})
	}
}

// counterTotals returns the +1/+1 counters of every Incubator (either
// face) the player controls, in battlefield order.
func counterTotals(g *game.Game, controller uuid.UUID) []int {
	var out []int
	for _, c := range incubatorsOf(g, controller) {
		out = append(out, c.Counters["+1/+1"])
	}
	return out
}

func sameInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func faceUp(g *game.Game, id uuid.UUID) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.ActiveFace
		}
	}
	return -1
}

func activateOn(t *testing.T, g *game.Game, who, source uuid.UUID, idx int, target uuid.UUID) error {
	t.Helper()
	params := game.ActivateAbilityParams{Strict: true}
	if target != uuid.Nil {
		params.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: target}}
	}
	return g.ActivateCatalogAbility(who, source, idx, params)
}

// --- the repeat ------------------------------------------------------

// "Incubate X twice" is two incubates, each its own Incubator with X
// counters: two tokens of three, never one token of six, and one
// token-created event each.
func TestGlisteningDawnMakesTwoSeparateIncubators(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushLands(g, me.ID, 3)
	before := countEvents(g, game.EventTokenCreated)

	castCatalogSpell(t, g, "Glistening Dawn", "Sorcery", glisteningDawnOracle, nil)
	passPriorityAroundTable(t, g)

	if got := counterTotals(g, me.ID); !sameInts(got, []int{3, 3}) {
		t.Fatalf("Incubator counters = %v, want two tokens of three (not one of six)", got)
	}
	if got := countEvents(g, game.EventTokenCreated) - before; got != 2 {
		t.Errorf("%d token-created events, want one per incubate (2)", got)
	}
	for _, c := range incubatorsOf(g, me.ID) {
		if c.Name != "Incubator" || c.ActiveFace != 0 {
			t.Errorf("a fresh Incubator should be front-up, got face %d named %q", c.ActiveFace, c.Name)
		}
	}
	if spec, _ := Lookup(glisteningDawnOracle); spec.Completeness != CompletenessFull {
		t.Errorf("Glistening Dawn completeness = %s, want full", spec.Completeness)
	}
}

// With no lands X is zero: "incubate 0" twice is two bare Incubators.
func TestGlisteningDawnWithNoLandsMakesTwoBareIncubators(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Glistening Dawn", "Sorcery", glisteningDawnOracle, nil)
	passPriorityAroundTable(t, g)
	if got := counterTotals(g, me.ID); !sameInts(got, []int{0, 0}) {
		t.Errorf("Incubator counters = %v, want two with none", got)
	}
}

// Each incubate runs its own creation window, so Doubling Season
// doubles both: four Incubators of twice the counters.
func TestIncubateTwiceEachRunsItsOwnReplacementWindow(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushLands(g, me.ID, 3)
	_ = seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)

	castCatalogSpell(t, g, "Glistening Dawn", "Sorcery", glisteningDawnOracle, nil)
	passPriorityAroundTable(t, g)

	if got := counterTotals(g, me.ID); !sameInts(got, []int{6, 6, 6, 6}) {
		t.Errorf("Incubator counters = %v, want four tokens of six", got)
	}
}

// --- Progenitor Exarch -----------------------------------------------

// "Incubate 3 X times": X = 2 is two Incubators of three.
func TestProgenitorExarchIncubatesThreeXTimes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	exarch := castXSpell(t, g, "Progenitor Exarch", "Creature — Phyrexian Cat Cleric",
		progenitorExarchOracle, "{X}{X}{W}", 2, nil)

	waitForXTrigger(t, g, exarch)
	if got := incubatorsOf(g, me.ID); len(got) != 0 {
		t.Fatalf("%d Incubators while the trigger is still on the stack", len(got))
	}
	passPriorityAroundTable(t, g)

	if got := counterTotals(g, me.ID); !sameInts(got, []int{3, 3}) {
		t.Errorf("Incubator counters = %v, want two tokens of three", got)
	}
}

func TestProgenitorExarchWithXZeroMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castXSpell(t, g, "Progenitor Exarch", "Creature — Phyrexian Cat Cleric",
		progenitorExarchOracle, "{X}{X}{W}", 0, nil)
	passPriorityAroundTable(t, g)
	if got := incubatorsOf(g, me.ID); len(got) != 0 {
		t.Errorf("X = 0 made %d Incubators, want none", len(got))
	}
}

// {T}: transform target Incubator token you control. A turned-over one
// is a Phyrexian Token, not an Incubator token, so it cannot be picked
// again; neither can an opponent's.
func TestProgenitorExarchTapTransformsAnIncubator(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exarch := pushCatalogPermanent(g, me.ID, "Progenitor Exarch", "Creature — Phyrexian Cat Cleric", progenitorExarchOracle, false)
	mine := makeIncubator(t, g, me.ID, 3)
	theirs := makeIncubator(t, g, opp.ID, 2)

	if err := activateOn(t, g, me.ID, exarch, 0, theirs); err == nil {
		t.Fatal("the Exarch targeted an opponent's Incubator")
	}
	if err := activateOn(t, g, me.ID, exarch, 0, mine); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if faceUp(g, mine) != 1 {
		t.Fatalf("the Incubator did not transform: face %d", faceUp(g, mine))
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == mine && c.Counters["+1/+1"] != 3 {
			t.Errorf("counters = %v, want three kept", c.Counters)
		}
	}
	// The Exarch is tapped now; untap it and try the turned-over token.
	untapForTest(g, exarch)
	if err := activateOn(t, g, me.ID, exarch, 0, mine); err == nil {
		t.Error("a Phyrexian Token was accepted as an Incubator token")
	}
}

// A tapped-out or summoning-sick Exarch cannot use the {T} ability.
func TestProgenitorExarchIsSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	exarch := pushCatalogPermanent(g, me.ID, "Progenitor Exarch", "Creature — Phyrexian Cat Cleric", progenitorExarchOracle, true)
	inc := makeIncubator(t, g, me.ID, 3)
	if err := activateOn(t, g, me.ID, exarch, 0, inc); err == nil {
		t.Error("a summoning-sick Exarch tapped for its ability")
	}
}

// --- Elvish Vatkeeper ------------------------------------------------

func TestElvishVatkeeperEntersIncubatingTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Elvish Vatkeeper", "Creature — Phyrexian Elf", elvishVatkeeperOracle, nil)
	passPriorityAroundTable(t, g)
	if got := counterTotals(g, me.ID); !sameInts(got, []int{2}) {
		t.Errorf("Incubator counters = %v, want one of two", got)
	}
}

// {5}: transform, then double the +1/+1 counters: a two-counter
// Incubator becomes a 4/4 Phyrexian Token.
func TestElvishVatkeeperTransformsAndDoublesTheCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	keeper := pushCatalogPermanent(g, me.ID, "Elvish Vatkeeper", "Creature — Phyrexian Elf", elvishVatkeeperOracle, false)
	inc := makeIncubator(t, g, me.ID, 2)
	floatForTest(g, me, "CCCCC")

	if err := activateOn(t, g, me.ID, keeper, 0, inc); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	var tok game.Card
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == inc {
			tok = c
		}
	}
	if tok.ActiveFace != 1 || tok.Name != "Phyrexian Token" {
		t.Fatalf("token = face %d %q, want the back face up", tok.ActiveFace, tok.Name)
	}
	if tok.Counters["+1/+1"] != 4 {
		t.Errorf("counters = %v, want four (two doubled)", tok.Counters)
	}
	if p, tg := tok.CurrentPower(), tok.CurrentToughness(); p != 4 || tg != 4 {
		t.Errorf("size = %d/%d, want 4/4", p, tg)
	}
}

// The doubling goes through the counter pipeline: Hardened Scales
// makes the doubling place one more.
func TestElvishVatkeeperDoublingUsesTheCounterPipeline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	keeper := pushCatalogPermanent(g, me.ID, "Elvish Vatkeeper", "Creature — Phyrexian Elf", elvishVatkeeperOracle, false)
	inc := makeIncubator(t, g, me.ID, 2)
	_ = seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	floatForTest(g, me, "CCCCC")

	if err := activateOn(t, g, me.ID, keeper, 0, inc); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == inc && c.Counters["+1/+1"] != 5 {
			t.Errorf("counters = %v, want five (two, plus two doubled, plus one from Hardened Scales)", c.Counters)
		}
	}
}

// --- Norn's Inquisitor -----------------------------------------------

func TestNornsInquisitorPutsACounterOnWhatTurnsIntoAPhyrexian(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Norn's Inquisitor", "Creature — Phyrexian Knight", nornsInquisitorOracle, nil)
	passPriorityAroundTable(t, g)
	inc := theIncubator(t, g, me.ID)
	if inc.Counters["+1/+1"] != 2 {
		t.Fatalf("Incubator counters = %v, want two from the enters trigger", inc.Counters)
	}

	transformIncubator(t, g, me, inc.InstanceID)

	var tok game.Card
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == inc.InstanceID {
			tok = c
		}
	}
	if tok.ActiveFace != 1 {
		t.Fatalf("the Incubator did not transform")
	}
	if tok.Counters["+1/+1"] != 3 {
		t.Errorf("counters = %v, want three: the Inquisitor adds one as it becomes a Phyrexian", tok.Counters)
	}
}

// An opponent's permanent turning Phyrexian is not "a permanent you
// control", and a permanent that turns over into something that is not
// a Phyrexian does not count either.
func TestNornsInquisitorIgnoresOpponentsAndNonPhyrexianFaces(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Norn's Inquisitor", "Creature — Phyrexian Knight", nornsInquisitorOracle, false)
	theirs := makeIncubator(t, g, opp.ID, 2)
	vault := importToBattlefield(t, g, stormTheVaultRow(), me)

	g.WithWriteLock(func() {
		if err := g.TransformPermanentForEffect(theirs); err != nil {
			t.Fatalf("transform theirs: %v", err)
		}
		if err := g.TransformPermanentForEffect(vault); err != nil {
			t.Fatalf("transform the vault: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == theirs && c.Counters["+1/+1"] != 2 {
			t.Errorf("an opponent's Phyrexian Token got a counter: %v", c.Counters)
		}
		if c.InstanceID == vault && len(c.Counters) != 0 {
			t.Errorf("a land got a counter: %v", c.Counters)
		}
	}
}

// --- Glissa, Herald of Predation -------------------------------------

// glissaAtCombat puts Glissa under seat 0's control and walks to the
// beginning of combat, returning the open mode prompt.
func glissaAtCombat(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	pushCatalogPermanent(g, me.ID, "Glissa, Herald of Predation", "Legendary Creature — Phyrexian Zombie Elf", glissaHeraldOracle, false)
	advanceTo(t, g, game.StepBeginCombat)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt at the beginning of combat: %+v", g.PendingChoices)
	}
	return pick
}

func TestGlissaIncubatesTwiceForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pick := glissaAtCombat(t, g, me)
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterTotals(g, me.ID); !sameInts(got, []int{2, 2}) {
		t.Errorf("Incubator counters = %v, want two tokens of two", got)
	}
}

func TestGlissaTransformsEveryIncubatorYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := makeIncubator(t, g, me.ID, 2)
	b := makeIncubator(t, g, me.ID, 4)
	theirs := makeIncubator(t, g, opp.ID, 3)
	// One already turned over stays turned over (it is not an Incubator
	// token any more, so the transform never reaches it).
	done := makeIncubator(t, g, me.ID, 1)
	g.WithWriteLock(func() { _ = g.TransformPermanentForEffect(done) })

	pick := glissaAtCombat(t, g, me)
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{a, b, done} {
		if faceUp(g, id) != 1 {
			t.Errorf("%s is face %d, want turned over", id, faceUp(g, id))
		}
	}
	if faceUp(g, theirs) != 0 {
		t.Error("an opponent's Incubator was transformed")
	}
}

func TestGlissaGivesPhyrexiansFirstStrikeAndDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phy := pushCatalogPermanent(g, me.ID, "Phyrexian Friend", incCPhyrexianCreatureTyp, "", false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Phyrexian", incCPhyrexianCreatureTyp, "", false)

	pick := glissaAtCombat(t, g, me)
	if err := g.ResolveModePick(pick.ID, me.ID, []int{2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, kw := range []string{"first strike", "deathtouch"} {
		if !effectiveAbilitiesContain(t, g, phy, kw) {
			t.Errorf("a Phyrexian you control lacks %s", kw)
		}
		if effectiveAbilitiesContain(t, g, bear, kw) {
			t.Errorf("a Bear gained %s", kw)
		}
		if effectiveAbilitiesContain(t, g, theirs, kw) {
			t.Errorf("an opponent's Phyrexian gained %s", kw)
		}
	}
}

// --- Sunder the Gateway ----------------------------------------------

func TestSunderTheGatewayDestroysAndIncubates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := pushTypedCard(g, opp.ID, "Their Rock", "Artifact", "{2}")
	castCatalogSpellWithModes2(t, g, []int{0}, []game.TargetRef{pbCard(rock)})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if got := counterTotals(g, me.ID); !sameInts(got, []int{2}) {
		t.Errorf("Incubator counters = %v, want one of two", got)
	}
}

// "Nontoken": an opponent's Incubator is not a legal target for the
// destroy mode.
func TestSunderTheGatewayCannotTargetATokenOrYourOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	token := makeIncubator(t, g, opp.ID, 2)
	mine := pushTypedCard(g, me.ID, "My Rock", "Artifact", "{2}")
	for _, target := range []uuid.UUID{token, mine} {
		if err := sunderCast(t, g, []int{0}, []game.TargetRef{pbCard(target)}); err == nil {
			t.Errorf("Sunder the Gateway accepted illegal target %s", target)
		}
	}
}

// Second mode: incubate 2, then choose an Incubator you control and
// turn it over. With an older Incubator on the table the controller may
// choose either; choosing the old one leaves the new one front-up.
func TestSunderTheGatewayIncubatesThenTransformsAChosenIncubator(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	old := makeIncubator(t, g, me.ID, 5)
	castCatalogSpellWithModes2(t, g, []int{1}, nil)
	passPriorityAroundTable(t, g)

	pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatalf("no choice of an Incubator was offered: %+v", g.PendingChoices)
	}
	if len(pick.ChooseCards) != 2 {
		t.Fatalf("%d candidates, want both Incubators", len(pick.ChooseCards))
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{old}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	passPriorityAroundTable(t, g)

	if faceUp(g, old) != 1 {
		t.Error("the chosen Incubator did not transform")
	}
	turned := 0
	for _, c := range incubatorsOf(g, me.ID) {
		if c.ActiveFace == 1 {
			turned++
		}
	}
	if turned != 1 {
		t.Errorf("%d Incubators are turned over, want exactly the chosen one", turned)
	}
}

func castCatalogSpellWithModes2(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) {
	t.Helper()
	if err := sunderCast(t, g, modes, targets); err != nil {
		t.Fatalf("cast Sunder the Gateway: %v", err)
	}
}

func sunderCast(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Sunder the Gateway", TypeLine: "Sorcery", OracleID: sunderTheGatewayOracle,
		Owner: active.ID, Controller: active.ID,
	})
	toMainForCost(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

// --- Elesh Norn // The Argent Etchings -------------------------------

func eleshNornCard(owner uuid.UUID, back bool) game.Card {
	return praetorCard(owner, eleshNornEtchingsOracle, "Elesh Norn", "Legendary Creature — Phyrexian Praetor", 3, 5,
		"The Argent Etchings", "Enchantment — Saga", back)
}

func TestEleshNornRegistersBothFacesFull(t *testing.T) {
	for _, key := range []string{eleshNornEtchingsOracle, eleshNornEtchingsOracle + "#1"} {
		spec, ok := Lookup(key)
		if !ok || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered %v, completeness %s, want full", key, ok, spec.Completeness)
		}
	}
}

// The front's third ability: sacrifice three other creatures, flip.
// The Saga enters with chapter I on the stack; five Incubators of two
// appear and every Incubator the controller has turns over.
func TestEleshNornFlipsAndChapterOneMakesFiveAndTransformsAll(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	elesh := pushBattlefieldCardWithTimestamp(g, eleshNornCard(me.ID, false))
	fodder := []uuid.UUID{
		pushVanillaCreature(g, me.ID, "Fodder A", 1, 1),
		pushVanillaCreature(g, me.ID, "Fodder B", 1, 1),
		pushVanillaCreature(g, me.ID, "Fodder C", 1, 1),
	}
	older := makeIncubator(t, g, me.ID, 1)
	toMainForCost(t, g)
	floatForTest(g, me, "WCC")

	if err := g.ActivateCatalogAbility(me.ID, elesh, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: fodder,
	}); err != nil {
		// Index 0 is the first Activated row; the damage ability is a trigger.
		t.Fatalf("flip: %v", err)
	}
	for i := 0; i < 12; i++ {
		if saga := praetorOnBattlefield(g, eleshNornEtchingsOracle); saga != nil && saga.Name == "The Argent Etchings" &&
			triggerOnStack(g, saga.InstanceID) != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	passPriorityAroundTable(t, g)

	got := counterTotals(g, me.ID)
	if !sameInts(sorted(got), []int{1, 2, 2, 2, 2, 2}) {
		t.Fatalf("Incubator counters = %v, want the old one of one and five new ones of two", got)
	}
	for _, c := range incubatorsOf(g, me.ID) {
		if c.ActiveFace != 1 {
			t.Errorf("an Incubator (%v counters) was left face-up after chapter I", c.Counters)
		}
	}
	if faceUp(g, older) != 1 {
		t.Error("the older Incubator was not transformed")
	}
}

func sorted(in []int) []int {
	out := append([]int(nil), in...)
	for i := range out {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// Chapter II then chapter III, a turn apiece.
func TestTheArgentEtchingsChaptersTwoAndThree(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	saga := pushBattlefieldCardWithTimestamp(g, eleshNornCard(me.ID, true))
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == saga {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterLore: 1}
			}
		}
	})
	mineBear := pushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)
	phy := pushCatalogPermanent(g, me.ID, "My Phyrexian", incCPhyrexianCreatureTyp, "", false)

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mineBear, phy} {
		if !effectiveAbilitiesContain(t, g, id, "double strike") {
			t.Errorf("chapter II: %s lacks double strike", id)
		}
	}
	if got := effectivePower(t, g, mineBear); got != 2 {
		t.Errorf("chapter II: Bear power = %d, want 2", got)
	}

	theirBear := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	theirEnch := pushTypedCard(g, opp.ID, "Their Anthem", "Enchantment", "{1}{W}")
	theirRock := pushTypedCard(g, opp.ID, "Their Rock", "Artifact", "{2}")
	theirPhy := pushCatalogPermanent(g, opp.ID, "Their Phyrexian", incCPhyrexianCreatureTyp, "", false)
	theirLand := pushTypedCard(g, opp.ID, "Their Forest", "Basic Land — Forest", "")

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)

	for _, gone := range []uuid.UUID{theirBear, theirEnch, mineBear} {
		if g.Battlefield.Contains(gone) {
			t.Errorf("chapter III left %s on the battlefield", gone)
		}
	}
	for _, kept := range []uuid.UUID{theirRock, theirPhy, theirLand, phy} {
		if !g.Battlefield.Contains(kept) {
			t.Errorf("chapter III destroyed %s, which is an artifact, a land or a Phyrexian", kept)
		}
	}
	back := praetorOnBattlefield(g, eleshNornEtchingsOracle)
	if back == nil || back.Name != "Elesh Norn" || back.ActiveFace != 0 || back.Counters[game.CounterLore] != 0 {
		t.Fatalf("the Saga did not return front face up as a new object (found %v)", back != nil)
	}
}

// The damage ability: pay {1} or lose 2.
func TestEleshNornMakesTheSourcesControllerPayOrLoseTwo(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushBattlefieldCardWithTimestamp(g, eleshNornCard(me.ID, false))
		bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
		lifeBefore := opp.Life

		g.WithWriteLock(func() {
			if err := g.DealDamageToPlayerForEffect(bear, me.ID, 2); err != nil {
				t.Fatalf("damage: %v", err)
			}
		})
		for i := 0; i < 8 && latestChoiceOfKind(g, game.PendingChoicePayUnless) == nil; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		prompt := latestChoiceOfKind(g, game.PendingChoicePayUnless)
		if prompt == nil || prompt.Chooser != opp.ID {
			t.Fatalf("pay=%v: no pay prompt for the source's controller: %+v", pay, prompt)
		}
		if pay {
			floatForTest(g, opp, "C")
		}
		if err := g.ResolvePayUnless(prompt.ID, opp.ID, pay); err != nil {
			t.Fatalf("ResolvePayUnless: %v", err)
		}
		passPriorityAroundTable(t, g)

		want := lifeBefore - 2
		if pay {
			want = lifeBefore
		}
		if opp.Life != want {
			t.Errorf("pay=%v: opponent life %d -> %d, want %d", pay, lifeBefore, opp.Life, want)
		}
	}
}

// Damage to a permanent you control triggers it too; damage from your
// own source, or to an opponent's, does not.
func TestEleshNornWatchesDamageToYourPermanentsFromOpponentsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, eleshNornCard(me.ID, false))
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	toMainForCost(t, g)

	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(mine, theirs, 1) // my source, their permanent
		_ = g.DealDamageToCreatureForEffect(mine, mine, 1)   // my source, my permanent
	})
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoicePayUnless) != nil {
		t.Fatal("Elesh Norn triggered on damage that is not an opponent's source hurting her side")
	}

	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(theirs, mine, 1) })
	for i := 0; i < 8 && latestChoiceOfKind(g, game.PendingChoicePayUnless) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if prompt := latestChoiceOfKind(g, game.PendingChoicePayUnless); prompt == nil || prompt.Chooser != opp.ID {
		t.Fatalf("no pay prompt for the opponent after their source damaged my creature: %+v", prompt)
	}
}

// The imported row is the shape the deck importer produces: the back
// face is a Saga, which is the reason this card needed checking.
func TestEleshNornImportedBackFaceIsASaga(t *testing.T) {
	row := cards.Card{
		ID: uuid.New(), Name: "Elesh Norn // The Argent Etchings", Layout: "transform",
		TypeLine: "Legendary Creature — Phyrexian Praetor // Enchantment — Saga",
		OracleID: uuid.MustParse(eleshNornEtchingsOracle),
		CardFaces: []cards.CardFace{
			{Name: "Elesh Norn", TypeLine: "Legendary Creature — Phyrexian Praetor", ManaCost: "{2}{W}{W}", Power: "3", Toughness: "5", Colors: []string{"W"}},
			{Name: "The Argent Etchings", TypeLine: "Enchantment — Saga", Colors: []string{"W"}},
		},
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := importToBattlefield(t, g, row, me)
	g.WithWriteLock(func() {
		if err := g.TransformPermanentForEffect(id); err != nil {
			t.Fatalf("transform: %v", err)
		}
	})
	var card game.Card
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			card = c
		}
	}
	if !game.IsSaga(card) || game.SagaFinalChapter(card) != 3 {
		t.Errorf("back face: saga=%v final chapter=%d, want a Saga with three chapters", game.IsSaga(card), game.SagaFinalChapter(card))
	}
}
