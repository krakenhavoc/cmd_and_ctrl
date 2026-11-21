package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counterTotals returns the +1/+1 counters of every Incubator (either
// face) the player controls, in battlefield order.
func counterTotals(g *game.Game, controller uuid.UUID) []int {
	var out []int
	for _, c := range effectstest.IncubatorsOf(g, controller) {
		out = append(out, c.Counters["+1/+1"])
	}
	return out
}

// --- the repeat ------------------------------------------------------

// "Incubate X twice" is two incubates, each its own Incubator with X
// counters: two tokens of three, never one token of six, and one
// token-created event each.
func TestGlisteningDawnMakesTwoSeparateIncubators(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLands(g, me.ID, 3)
	before := effectstest.CountEvents(g, game.EventTokenCreated)

	effectstest.CastCatalogSpell(t, g, "Glistening Dawn", "Sorcery", effectstest.GlisteningDawnOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := counterTotals(g, me.ID); !effectstest.SameInts(got, []int{3, 3}) {
		t.Fatalf("Incubator counters = %v, want two tokens of three (not one of six)", got)
	}
	if got := effectstest.CountEvents(g, game.EventTokenCreated) - before; got != 2 {
		t.Errorf("%d token-created events, want one per incubate (2)", got)
	}
	for _, c := range effectstest.IncubatorsOf(g, me.ID) {
		if c.Name != "Incubator" || c.ActiveFace != 0 {
			t.Errorf("a fresh Incubator should be front-up, got face %d named %q", c.ActiveFace, c.Name)
		}
	}
	if spec, _ := effects.Lookup(effectstest.GlisteningDawnOracle); spec.Completeness != effects.CompletenessFull {
		t.Errorf("Glistening Dawn completeness = %s, want full", spec.Completeness)
	}
}

// With no lands X is zero: "incubate 0" twice is two bare Incubators.
func TestGlisteningDawnWithNoLandsMakesTwoBareIncubators(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Glistening Dawn", "Sorcery", effectstest.GlisteningDawnOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := counterTotals(g, me.ID); !effectstest.SameInts(got, []int{0, 0}) {
		t.Errorf("Incubator counters = %v, want two with none", got)
	}
}

// Each incubate runs its own creation window, so Doubling Season
// doubles both: four Incubators of twice the counters.
func TestIncubateTwiceEachRunsItsOwnReplacementWindow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLands(g, me.ID, 3)
	_ = effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)

	effectstest.CastCatalogSpell(t, g, "Glistening Dawn", "Sorcery", effectstest.GlisteningDawnOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := counterTotals(g, me.ID); !effectstest.SameInts(got, []int{6, 6, 6, 6}) {
		t.Errorf("Incubator counters = %v, want four tokens of six", got)
	}
}

// --- Progenitor Exarch -----------------------------------------------

// "Incubate 3 X times": X = 2 is two Incubators of three.
func TestProgenitorExarchIncubatesThreeXTimes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	exarch := effectstest.CastXSpell(t, g, "Progenitor Exarch", "Creature — Phyrexian Cat Cleric",
		effectstest.ProgenitorExarchOracle, "{X}{X}{W}", 2, nil)

	effectstest.WaitForXTrigger(t, g, exarch)
	if got := effectstest.IncubatorsOf(g, me.ID); len(got) != 0 {
		t.Fatalf("%d Incubators while the trigger is still on the stack", len(got))
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := counterTotals(g, me.ID); !effectstest.SameInts(got, []int{3, 3}) {
		t.Errorf("Incubator counters = %v, want two tokens of three", got)
	}
}

func TestProgenitorExarchWithXZeroMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastXSpell(t, g, "Progenitor Exarch", "Creature — Phyrexian Cat Cleric",
		effectstest.ProgenitorExarchOracle, "{X}{X}{W}", 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.IncubatorsOf(g, me.ID); len(got) != 0 {
		t.Errorf("X = 0 made %d Incubators, want none", len(got))
	}
}

// {T}: transform target Incubator token you control. A turned-over one
// is a Phyrexian Token, not an Incubator token, so it cannot be picked
// again; neither can an opponent's.
func TestProgenitorExarchTapTransformsAnIncubator(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exarch := effectstest.PushCatalogPermanent(g, me.ID, "Progenitor Exarch", "Creature — Phyrexian Cat Cleric", effectstest.ProgenitorExarchOracle, false)
	mine := effectstest.MakeIncubator(t, g, me.ID, 3)
	theirs := effectstest.MakeIncubator(t, g, opp.ID, 2)

	if err := effectstest.ActivateOn(t, g, me.ID, exarch, 0, theirs); err == nil {
		t.Fatal("the Exarch targeted an opponent's Incubator")
	}
	if err := effectstest.ActivateOn(t, g, me.ID, exarch, 0, mine); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.FaceUp(g, mine) != 1 {
		t.Fatalf("the Incubator did not transform: face %d", effectstest.FaceUp(g, mine))
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == mine && c.Counters["+1/+1"] != 3 {
			t.Errorf("counters = %v, want three kept", c.Counters)
		}
	}
	// The Exarch is tapped now; untap it and try the turned-over token.
	effectstest.UntapForTest(g, exarch)
	if err := effectstest.ActivateOn(t, g, me.ID, exarch, 0, mine); err == nil {
		t.Error("a Phyrexian Token was accepted as an Incubator token")
	}
}

// A tapped-out or summoning-sick Exarch cannot use the {T} ability.
func TestProgenitorExarchIsSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	exarch := effectstest.PushCatalogPermanent(g, me.ID, "Progenitor Exarch", "Creature — Phyrexian Cat Cleric", effectstest.ProgenitorExarchOracle, true)
	inc := effectstest.MakeIncubator(t, g, me.ID, 3)
	if err := effectstest.ActivateOn(t, g, me.ID, exarch, 0, inc); err == nil {
		t.Error("a summoning-sick Exarch tapped for its ability")
	}
}

// --- Elvish Vatkeeper ------------------------------------------------

func TestElvishVatkeeperEntersIncubatingTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Elvish Vatkeeper", "Creature — Phyrexian Elf", effectstest.ElvishVatkeeperOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := counterTotals(g, me.ID); !effectstest.SameInts(got, []int{2}) {
		t.Errorf("Incubator counters = %v, want one of two", got)
	}
}

// {5}: transform, then double the +1/+1 counters: a two-counter
// Incubator becomes a 4/4 Phyrexian Token.
func TestElvishVatkeeperTransformsAndDoublesTheCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Elvish Vatkeeper", "Creature — Phyrexian Elf", effectstest.ElvishVatkeeperOracle, false)
	inc := effectstest.MakeIncubator(t, g, me.ID, 2)
	effectstest.FloatForTest(g, me, "CCCCC")

	if err := effectstest.ActivateOn(t, g, me.ID, keeper, 0, inc); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Elvish Vatkeeper", "Creature — Phyrexian Elf", effectstest.ElvishVatkeeperOracle, false)
	inc := effectstest.MakeIncubator(t, g, me.ID, 2)
	_ = effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	effectstest.FloatForTest(g, me, "CCCCC")

	if err := effectstest.ActivateOn(t, g, me.ID, keeper, 0, inc); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == inc && c.Counters["+1/+1"] != 5 {
			t.Errorf("counters = %v, want five (two, plus two doubled, plus one from Hardened Scales)", c.Counters)
		}
	}
}

// --- Norn's Inquisitor -----------------------------------------------

func TestNornsInquisitorPutsACounterOnWhatTurnsIntoAPhyrexian(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Norn's Inquisitor", "Creature — Phyrexian Knight", effectstest.NornsInquisitorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Norn's Inquisitor", "Creature — Phyrexian Knight", effectstest.NornsInquisitorOracle, false)
	theirs := effectstest.MakeIncubator(t, g, opp.ID, 2)
	vault := effectstest.ImportToBattlefield(t, g, effectstest.StormTheVaultRow(), me)

	g.WithWriteLock(func() {
		if err := g.TransformPermanentForEffect(theirs); err != nil {
			t.Fatalf("transform theirs: %v", err)
		}
		if err := g.TransformPermanentForEffect(vault); err != nil {
			t.Fatalf("transform the vault: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)

	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == theirs && c.Counters["+1/+1"] != 2 {
			t.Errorf("an opponent's Phyrexian Token got a counter: %v", c.Counters)
		}
		if c.InstanceID == vault && len(c.Counters) != 0 {
			t.Errorf("a land got a counter: %v", c.Counters)
		}
	}
}

func TestGlissaIncubatesTwiceForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pick := effectstest.GlissaAtCombat(t, g, me)
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := counterTotals(g, me.ID); !effectstest.SameInts(got, []int{2, 2}) {
		t.Errorf("Incubator counters = %v, want two tokens of two", got)
	}
}

func TestGlissaTransformsEveryIncubatorYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.MakeIncubator(t, g, me.ID, 2)
	b := effectstest.MakeIncubator(t, g, me.ID, 4)
	theirs := effectstest.MakeIncubator(t, g, opp.ID, 3)
	// One already turned over stays turned over (it is not an Incubator
	// token any more, so the transform never reaches it).
	done := effectstest.MakeIncubator(t, g, me.ID, 1)
	g.WithWriteLock(func() { _ = g.TransformPermanentForEffect(done) })

	pick := effectstest.GlissaAtCombat(t, g, me)
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{a, b, done} {
		if effectstest.FaceUp(g, id) != 1 {
			t.Errorf("%s is face %d, want turned over", id, effectstest.FaceUp(g, id))
		}
	}
	if effectstest.FaceUp(g, theirs) != 0 {
		t.Error("an opponent's Incubator was transformed")
	}
}

func TestGlissaGivesPhyrexiansFirstStrikeAndDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phy := effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Friend", effectstest.IncCPhyrexianCreatureTyp, "", false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Phyrexian", effectstest.IncCPhyrexianCreatureTyp, "", false)

	pick := effectstest.GlissaAtCombat(t, g, me)
	if err := g.ResolveModePick(pick.ID, me.ID, []int{2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	for _, kw := range []string{"first strike", "deathtouch"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, phy, kw) {
			t.Errorf("a Phyrexian you control lacks %s", kw)
		}
		if effectstest.EffectiveAbilitiesContain(t, g, bear, kw) {
			t.Errorf("a Bear gained %s", kw)
		}
		if effectstest.EffectiveAbilitiesContain(t, g, theirs, kw) {
			t.Errorf("an opponent's Phyrexian gained %s", kw)
		}
	}
}

// --- Sunder the Gateway ----------------------------------------------

func TestSunderTheGatewayDestroysAndIncubates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushTypedCard(g, opp.ID, "Their Rock", "Artifact", "{2}")
	castCatalogSpellWithModes2(t, g, []int{0}, []game.TargetRef{effectstest.PbCard(rock)})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if got := counterTotals(g, me.ID); !effectstest.SameInts(got, []int{2}) {
		t.Errorf("Incubator counters = %v, want one of two", got)
	}
}

// "Nontoken": an opponent's Incubator is not a legal target for the
// destroy mode.
func TestSunderTheGatewayCannotTargetATokenOrYourOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	token := effectstest.MakeIncubator(t, g, opp.ID, 2)
	mine := effectstest.PushTypedCard(g, me.ID, "My Rock", "Artifact", "{2}")
	for _, target := range []uuid.UUID{token, mine} {
		if err := sunderCast(t, g, []int{0}, []game.TargetRef{effectstest.PbCard(target)}); err == nil {
			t.Errorf("Sunder the Gateway accepted illegal target %s", target)
		}
	}
}

// Second mode: incubate 2, then choose an Incubator you control and
// turn it over. With an older Incubator on the table the controller may
// choose either; choosing the old one leaves the new one front-up.
func TestSunderTheGatewayIncubatesThenTransformsAChosenIncubator(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	old := effectstest.MakeIncubator(t, g, me.ID, 5)
	castCatalogSpellWithModes2(t, g, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatalf("no choice of an Incubator was offered: %+v", g.PendingChoices)
	}
	if len(pick.ChooseCards) != 2 {
		t.Fatalf("%d candidates, want both Incubators", len(pick.ChooseCards))
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{old}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.FaceUp(g, old) != 1 {
		t.Error("the chosen Incubator did not transform")
	}
	turned := 0
	for _, c := range effectstest.IncubatorsOf(g, me.ID) {
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
		InstanceID: id, Name: "Sunder the Gateway", TypeLine: "Sorcery", OracleID: effectstest.SunderTheGatewayOracle,
		Owner: active.ID, Controller: active.ID,
	})
	effectstest.ToMainForCost(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

// --- Elesh Norn // The Argent Etchings -------------------------------

func eleshNornCard(owner uuid.UUID, back bool) game.Card {
	return effectstest.PraetorCard(owner, effectstest.EleshNornEtchingsOracle, "Elesh Norn", "Legendary Creature — Phyrexian Praetor", 3, 5,
		"The Argent Etchings", "Enchantment — Saga", back)
}

func TestEleshNornRegistersBothFacesFull(t *testing.T) {
	for _, key := range []string{effectstest.EleshNornEtchingsOracle, effectstest.EleshNornEtchingsOracle + "#1"} {
		spec, ok := effects.Lookup(key)
		if !ok || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: registered %v, completeness %s, want full", key, ok, spec.Completeness)
		}
	}
}

// The front's third ability: sacrifice three other creatures, flip.
// The Saga enters with chapter I on the stack; five Incubators of two
// appear and every Incubator the controller has turns over.
func TestEleshNornFlipsAndChapterOneMakesFiveAndTransformsAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	elesh := effectstest.PushBattlefieldCardWithTimestamp(g, eleshNornCard(me.ID, false))
	fodder := []uuid.UUID{
		effectstest.PushVanillaCreature(g, me.ID, "Fodder A", 1, 1),
		effectstest.PushVanillaCreature(g, me.ID, "Fodder B", 1, 1),
		effectstest.PushVanillaCreature(g, me.ID, "Fodder C", 1, 1),
	}
	older := effectstest.MakeIncubator(t, g, me.ID, 1)
	effectstest.ToMainForCost(t, g)
	effectstest.FloatForTest(g, me, "WCC")

	if err := g.ActivateCatalogAbility(me.ID, elesh, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: fodder,
	}); err != nil {
		// Index 0 is the first Activated row; the damage ability is a trigger.
		t.Fatalf("flip: %v", err)
	}
	for i := 0; i < 12; i++ {
		if saga := effectstest.PraetorOnBattlefield(g, effectstest.EleshNornEtchingsOracle); saga != nil && saga.Name == "The Argent Etchings" &&
			effectstest.TriggerOnStack(g, saga.InstanceID) != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)

	got := counterTotals(g, me.ID)
	if !effectstest.SameInts(sorted(got), []int{1, 2, 2, 2, 2, 2}) {
		t.Fatalf("Incubator counters = %v, want the old one of one and five new ones of two", got)
	}
	for _, c := range effectstest.IncubatorsOf(g, me.ID) {
		if c.ActiveFace != 1 {
			t.Errorf("an Incubator (%v counters) was left face-up after chapter I", c.Counters)
		}
	}
	if effectstest.FaceUp(g, older) != 1 {
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
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	saga := effectstest.PushBattlefieldCardWithTimestamp(g, eleshNornCard(me.ID, true))
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == saga {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterLore: 1}
			}
		}
	})
	mineBear := effectstest.PushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)
	phy := effectstest.PushCatalogPermanent(g, me.ID, "My Phyrexian", effectstest.IncCPhyrexianCreatureTyp, "", false)

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mineBear, phy} {
		if !effectstest.EffectiveAbilitiesContain(t, g, id, "double strike") {
			t.Errorf("chapter II: %s lacks double strike", id)
		}
	}
	if got := effectstest.EffectivePower(t, g, mineBear); got != 2 {
		t.Errorf("chapter II: Bear power = %d, want 2", got)
	}

	theirBear := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	theirEnch := effectstest.PushTypedCard(g, opp.ID, "Their Anthem", "Enchantment", "{1}{W}")
	theirRock := effectstest.PushTypedCard(g, opp.ID, "Their Rock", "Artifact", "{2}")
	theirPhy := effectstest.PushCatalogPermanent(g, opp.ID, "Their Phyrexian", effectstest.IncCPhyrexianCreatureTyp, "", false)
	theirLand := effectstest.PushTypedCard(g, opp.ID, "Their Forest", "Basic Land — Forest", "")

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

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
	back := effectstest.PraetorOnBattlefield(g, effectstest.EleshNornEtchingsOracle)
	if back == nil || back.Name != "Elesh Norn" || back.ActiveFace != 0 || back.Counters[game.CounterLore] != 0 {
		t.Fatalf("the Saga did not return front face up as a new object (found %v)", back != nil)
	}
}

// The damage ability: pay {1} or lose 2.
func TestEleshNornMakesTheSourcesControllerPayOrLoseTwo(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushBattlefieldCardWithTimestamp(g, eleshNornCard(me.ID, false))
		bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
		lifeBefore := opp.Life

		g.WithWriteLock(func() {
			if err := g.DealDamageToPlayerForEffect(bear, me.ID, 2); err != nil {
				t.Fatalf("damage: %v", err)
			}
		})
		for i := 0; i < 8 && effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless) == nil; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		prompt := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
		if prompt == nil || prompt.Chooser != opp.ID {
			t.Fatalf("pay=%v: no pay prompt for the source's controller: %+v", pay, prompt)
		}
		if pay {
			effectstest.FloatForTest(g, opp, "C")
		}
		if err := g.ResolvePayUnless(prompt.ID, opp.ID, pay); err != nil {
			t.Fatalf("ResolvePayUnless: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, eleshNornCard(me.ID, false))
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.ToMainForCost(t, g)

	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(mine, theirs, 1) // my source, their permanent
		_ = g.DealDamageToCreatureForEffect(mine, mine, 1)   // my source, my permanent
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless) != nil {
		t.Fatal("Elesh Norn triggered on damage that is not an opponent's source hurting her side")
	}

	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(theirs, mine, 1) })
	for i := 0; i < 8 && effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if prompt := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless); prompt == nil || prompt.Chooser != opp.ID {
		t.Fatalf("no pay prompt for the opponent after their source damaged my creature: %+v", prompt)
	}
}

// The imported row is the shape the deck importer produces: the back
// face is a Saga, which is the reason this card needed checking.
func TestEleshNornImportedBackFaceIsASaga(t *testing.T) {
	row := cards.Card{
		ID: uuid.New(), Name: "Elesh Norn // The Argent Etchings", Layout: "transform",
		TypeLine: "Legendary Creature — Phyrexian Praetor // Enchantment — Saga",
		OracleID: uuid.MustParse(effectstest.EleshNornEtchingsOracle),
		CardFaces: []cards.CardFace{
			{Name: "Elesh Norn", TypeLine: "Legendary Creature — Phyrexian Praetor", ManaCost: "{2}{W}{W}", Power: "3", Toughness: "5", Colors: []string{"W"}},
			{Name: "The Argent Etchings", TypeLine: "Enchantment — Saga", Colors: []string{"W"}},
		},
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.ImportToBattlefield(t, g, row, me)
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
