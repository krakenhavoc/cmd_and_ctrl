package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// incubate_test.go — CR 701.53 / CR 111.10i (ADR 0185, #2417): the
// Incubator, the first double-faced token, and the cards that make one.

const (
	sunfallOracle              = "fb3f5097-0da0-458c-8508-60823567e2da"
	traumaticRevelationOracle  = "e47253d1-9b12-4989-80c0-604906c88b5e"
	converterBeastOracle       = "7fbb9547-a805-4768-8574-e283e64600a4"
	eyesOfGitaxiasOracle       = "7ef6ca6e-a8b8-4ea6-b13d-a08eb847c12d"
	mercilessRepurposingOracle = "2e1b5ae6-a29d-4f37-a380-eebee6cdcbd2"
	exciseTheImperfectOracle   = "d3264482-2a5a-45e8-987f-593b8aceac12"
	relmsSketchingOracle       = "dd6601c6-810d-4883-89d2-5d31419fb1cc"
)

// incubatorsOf returns every Incubator token on the battlefield that
// `controller` controls, whichever face is up.
func incubatorsOf(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.TokenKey == game.TokenKey("incubator") && c.Controller == controller {
				out = append(out, c)
			}
		}
	})
	return out
}

// theIncubator is incubatorsOf that insists on exactly one.
func theIncubator(t *testing.T, g *game.Game, controller uuid.UUID) game.Card {
	t.Helper()
	all := incubatorsOf(g, controller)
	if len(all) != 1 {
		t.Fatalf("want exactly one Incubator under %s, have %d", controller, len(all))
	}
	return all[0]
}

// payTwo puts {2} in the seat's pool.
func payTwo(p *game.Player) {
	p.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
}

// transformIncubator pays {2} and turns the Incubator over, resolving
// the ability.
func transformIncubator(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) {
	t.Helper()
	payTwo(p)
	if err := g.ActivateCatalogAbility(p.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the Incubator: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// --- the template ---------------------------------------------------

func TestTheIncubatorTemplateIsADoubleFacedToken(t *testing.T) {
	c := IncubatorToken()
	if c.TokenKey != "token:incubator" || c.OracleID != "" {
		t.Fatalf("identity = oracle %q key %q", c.OracleID, c.TokenKey)
	}
	if c.Layout != game.LayoutTransform || len(c.Faces) != 2 || c.ActiveFace != 0 {
		t.Fatalf("layout %q faces %d active %d, want a transform pair with the front up",
			c.Layout, len(c.Faces), c.ActiveFace)
	}
	if c.Name != "Incubator" || c.TypeLine != "Token Artifact — Incubator" || len(c.Colors) != 0 {
		t.Errorf("front = %q / %q / %v", c.Name, c.TypeLine, c.Colors)
	}
	back := c.Faces[1]
	if back.Name != "Phyrexian Token" || back.TypeLine != "Token Artifact Creature — Phyrexian" ||
		back.Power != 0 || back.Toughness != 0 || len(back.Colors) != 0 {
		t.Errorf("back = %+v, want a colourless 0/0 Phyrexian artifact creature named Phyrexian Token", back)
	}
	// The flat fields are Faces[ActiveFace] (card.go's invariant).
	if c.Faces[0].Name != c.Name || c.Faces[0].TypeLine != c.TypeLine {
		t.Errorf("flat fields drifted from Faces[0]: %+v", c.Faces[0])
	}
	if !game.CanTransform(c) {
		t.Error("CanTransform refuses an Incubator: CR 701.27a lets a double-faced token transform")
	}
	abs := game.ActivatedAbilitiesForCard(c)
	if len(abs) != 1 || abs[0].Cost.Mana != "{2}" || abs[0].Cost.Tap || abs[0].Cost.SacrificeSelf {
		t.Fatalf("front abilities = %+v, want one {2} row with no tap or sacrifice", abs)
	}
	if got := game.TokenTextForCard(c); got != "{2}: Transform this token." {
		t.Errorf("printed text = %q", got)
	}
	// The back face is a different catalog entry, and an empty one.
	c.SetFace(1)
	if got := game.CatalogKey(c); got != "token:incubator#1" {
		t.Errorf("back face key = %q", got)
	}
	if abs := game.ActivatedAbilitiesForCard(c); len(abs) != 0 {
		t.Errorf("the Phyrexian Token carries the Incubator's ability: %+v", abs)
	}
	if got := game.TokenTextForCard(c); got != "" {
		t.Errorf("the back face prints %q, want nothing", got)
	}
}

// --- the keyword action ---------------------------------------------

func TestConverterBeastIncubatesFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Converter Beast", "Creature — Phyrexian Beast", converterBeastOracle, nil)
	passPriorityAroundTable(t, g)

	inc := theIncubator(t, g, me.ID)
	if inc.Counters["+1/+1"] != 5 {
		t.Errorf("counters = %v, want five +1/+1", inc.Counters)
	}
	if inc.ActiveFace != 0 || inc.Name != "Incubator" || inc.IsCreature() || !inc.IsArtifact() || !IsToken(inc) {
		t.Errorf("a fresh Incubator should be a front-up noncreature artifact token: %+v", inc)
	}
	if inc.Owner != me.ID {
		t.Errorf("owner = %s", inc.Owner)
	}
}

// Counters are not a characteristic, so the transform keeps all of them
// (CR 712.18): the same object, a 5/5 Phyrexian artifact creature.
func TestTransformingTheIncubatorKeepsItsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Converter Beast", "Creature — Phyrexian Beast", converterBeastOracle, nil)
	passPriorityAroundTable(t, g)
	before := theIncubator(t, g, me.ID)

	transformIncubator(t, g, me, before.InstanceID)

	after := theIncubator(t, g, me.ID)
	if after.InstanceID != before.InstanceID {
		t.Errorf("a transform made a new object: %s -> %s", before.InstanceID, after.InstanceID)
	}
	if after.ActiveFace != 1 || after.Name != "Phyrexian Token" {
		t.Fatalf("after: face %d named %q, want the back face up", after.ActiveFace, after.Name)
	}
	if after.Counters["+1/+1"] != 5 {
		t.Errorf("counters = %v, want five kept across the transform", after.Counters)
	}
	if !after.IsCreature() || !after.IsArtifact() || !IsToken(after) {
		t.Errorf("types = %q, want a token artifact creature", after.TypeLine)
	}
	if p, tg := after.CurrentPower(), after.CurrentToughness(); p != 5 || tg != 5 {
		t.Errorf("size = %d/%d, want 5/5 (0/0 plus five counters)", p, tg)
	}
	if !TransformedPermanent(after) {
		t.Error("CR 701.27g: a double-faced token with its back face up is a transformed permanent")
	}
	if game.CatalogKey(after) != "token:incubator#1" {
		t.Errorf("key = %q", game.CatalogKey(after))
	}
	if len(game.ActivatedAbilitiesForCard(after)) != 0 {
		t.Error("the Phyrexian Token still offers the Incubator's transform")
	}
	if tf := countEvents(g, game.EventTransform); tf != 1 {
		t.Errorf("%d transform events, want 1", tf)
	}
}

// An Incubator with no counters is a real 0/0 once it is a creature, so
// CR 704.5f buries it. That is the printed risk of "incubate 0".
func TestAZeroCounterIncubatorDiesWhenItTransforms(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		_, _ = g.CreateTokensForEffect(me.ID, IncubatorToken(), 1, game.TokenEntryOptions{})
	})
	inc := theIncubator(t, g, me.ID)
	if len(inc.Counters) != 0 {
		t.Fatalf("counters = %v, want none", inc.Counters)
	}
	transformIncubator(t, g, me, inc.InstanceID)
	if got := incubatorsOf(g, me.ID); len(got) != 0 {
		t.Errorf("the 0/0 Phyrexian Token is still on the battlefield: %+v", got)
	}
}

// The creation and the counters both ride the replacement pipeline
// (CR 614.1c, CR 701.7b): Doubling Season turns "incubate 5" into two
// Incubators, each entering with ten counters.
func TestIncubateEntersThroughTheCounterReplacementPipeline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_ = seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	castCatalogSpell(t, g, "Converter Beast", "Creature — Phyrexian Beast", converterBeastOracle, nil)
	passPriorityAroundTable(t, g)
	all := incubatorsOf(g, me.ID)
	if len(all) != 2 {
		t.Fatalf("%d Incubators, want 2 (the token doubled)", len(all))
	}
	for _, c := range all {
		if c.Counters["+1/+1"] != 10 {
			t.Errorf("counters = %v, want ten (the five doubled)", c.Counters)
		}
	}
}

// A transformed Incubator that dies is a creature dying: its death is
// seen as one by a "whenever a creature dies" trigger, through
// last-known information, even though the token is a front-face Incubator
// by the time anything reads the graveyard (CR 712.8a, CR 603.10a).
func TestATransformedIncubatorDyingIsACreatureDying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	castCatalogSpell(t, g, "Converter Beast", "Creature — Phyrexian Beast", converterBeastOracle, nil)
	passPriorityAroundTable(t, g)
	inc := theIncubator(t, g, me.ID)
	transformIncubator(t, g, me, inc.InstanceID)

	oppBefore := opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(inc.InstanceID) })
	for i := 0; i < 6 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppBefore-1 {
		t.Errorf("Blood Artist did not see the Phyrexian Token die: life %d -> %d", oppBefore, opp.Life)
	}
}

// The front face is an artifact, not a creature: an untransformed
// Incubator dying is not a creature dying.
func TestAnUntransformedIncubatorDyingIsNotACreatureDying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	castCatalogSpell(t, g, "Converter Beast", "Creature — Phyrexian Beast", converterBeastOracle, nil)
	passPriorityAroundTable(t, g)
	inc := theIncubator(t, g, me.ID)

	oppBefore := opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(inc.InstanceID) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || opp.Life != oppBefore {
		t.Errorf("Blood Artist reacted to an artifact dying: life %d -> %d", oppBefore, opp.Life)
	}
}

// --- copies (CR 707.8a) ---------------------------------------------

// A token copy of an Incubator is an Incubator: both faces, the front
// up, the transform ability, and none of the original's counters.
func TestATokenCopyOfAnIncubatorIsADoubleFacedToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Converter Beast", "Creature — Phyrexian Beast", converterBeastOracle, nil)
	passPriorityAroundTable(t, g)
	orig := theIncubator(t, g, me.ID)

	castAt(t, g, "Relm's Sketching", "Sorcery", relmsSketchingOracle, pbCard(orig.InstanceID))

	all := incubatorsOf(g, me.ID)
	if len(all) != 2 {
		t.Fatalf("%d Incubators after the copy, want 2", len(all))
	}
	var cp game.Card
	for _, c := range all {
		if c.InstanceID != orig.InstanceID {
			cp = c
		}
	}
	if cp.Layout != game.LayoutTransform || len(cp.Faces) != 2 || cp.ActiveFace != 0 {
		t.Fatalf("the copy lost its faces: layout %q faces %d active %d", cp.Layout, len(cp.Faces), cp.ActiveFace)
	}
	if len(cp.Counters) != 0 {
		t.Errorf("the copy has the original's counters: %v", cp.Counters)
	}
	if len(game.ActivatedAbilitiesForCard(cp)) != 1 {
		t.Errorf("the copy lost the transform ability")
	}
	if !game.CanTransform(cp) {
		t.Error("the copy can't transform")
	}
}

// A copy of a TRANSFORMED Incubator copies both faces and enters with
// the back face up (CR 707.8a).
func TestATokenCopyOfATransformedIncubatorEntersBackFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Converter Beast", "Creature — Phyrexian Beast", converterBeastOracle, nil)
	passPriorityAroundTable(t, g)
	orig := theIncubator(t, g, me.ID)
	transformIncubator(t, g, me, orig.InstanceID)

	var tmpl game.Card
	var ok bool
	g.ReadSnapshot(func() { tmpl, ok = TokenCopyTemplate(g, orig.InstanceID) })
	if !ok {
		t.Fatal("no template")
	}
	if tmpl.TokenKey != "token:incubator" || tmpl.Layout != game.LayoutTransform ||
		len(tmpl.Faces) != 2 || tmpl.ActiveFace != 1 || tmpl.Name != "Phyrexian Token" {
		t.Errorf("template = key %q layout %q faces %d active %d name %q",
			tmpl.TokenKey, tmpl.Layout, len(tmpl.Faces), tmpl.ActiveFace, tmpl.Name)
	}
	// The template's faces are its own: editing them can't reach the
	// original's.
	tmpl.Faces[0].Name = "edited"
	if theIncubator(t, g, me.ID).Faces[0].Name != "Incubator" {
		t.Error("the copy template shares its Faces with the original")
	}
}

// --- snapshots ------------------------------------------------------

// A turned-over Incubator is a restore point, and comes back on the same
// face with its counters, its key and its (empty) back-face abilities.
func TestATransformedIncubatorSurvivesASnapshotRestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Converter Beast", "Creature — Phyrexian Beast", converterBeastOracle, nil)
	passPriorityAroundTable(t, g)
	front := theIncubator(t, g, me.ID)
	frontSnap := g.CaptureSnapshot()
	transformIncubator(t, g, me, front.InstanceID)

	for name, snap := range map[string]*game.GameSnapshot{"front": frontSnap, "back": g.CaptureSnapshot()} {
		if !snap.Restorable() {
			t.Fatalf("%s: not a restore point: %+v", name, snap.Continuations)
		}
		restored, err := snap.RestoreStrict()
		if err != nil {
			t.Fatalf("%s: RestoreStrict: %v", name, err)
		}
		c := findRestoredCard(t, restored, front.InstanceID)
		wantFace := 0
		if name == "back" {
			wantFace = 1
		}
		if c.ActiveFace != wantFace || len(c.Faces) != 2 || c.Layout != game.LayoutTransform || c.Counters["+1/+1"] != 5 {
			t.Errorf("%s: restored face %d, %d faces, layout %q, counters %v",
				name, c.ActiveFace, len(c.Faces), c.Layout, c.Counters)
		}
		if got := len(game.ActivatedAbilitiesForCard(*c)); got != 1-wantFace {
			t.Errorf("%s: %d activated abilities after restore", name, got)
		}
	}
}

// --- the cards ------------------------------------------------------

func TestSunfallExilesEveryCreatureAndIncubatesThatMany(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	theirs1 := pushCatalogPermanent(g, opp.ID, "Ogre", "Creature — Ogre", "", false)
	theirs2 := pushCatalogPermanent(g, opp.ID, "Ogre", "Creature — Ogre", "", false)
	rock := pushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)

	castCatalogSpell(t, g, "Sunfall", "Sorcery", sunfallOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs1, theirs2} {
		if battlefieldHas(g, id) {
			t.Errorf("creature %s survived Sunfall", id)
		}
	}
	if !battlefieldHas(g, rock) {
		t.Error("Sunfall exiled a noncreature artifact")
	}
	inc := theIncubator(t, g, me.ID)
	if inc.Counters["+1/+1"] != 3 {
		t.Errorf("counters = %v, want three (one per creature exiled)", inc.Counters)
	}
	if n := len(incubatorsOf(g, opp.ID)); n != 0 {
		t.Errorf("the opponent has %d Incubators; the caster incubates", n)
	}
}

// No creatures: incubate 0 still makes the token.
func TestSunfallWithNoCreaturesMakesAnEmptyIncubator(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Sunfall", "Sorcery", sunfallOracle, nil)
	passPriorityAroundTable(t, g)
	inc := theIncubator(t, g, me.ID)
	if len(inc.Counters) != 0 {
		t.Errorf("counters = %v, want none", inc.Counters)
	}
}

func TestTraumaticRevelationIncubatesWhenYouChooseNothing(t *testing.T) {
	t.Run("a creature is chosen", func(t *testing.T) {
		g := newCatalogGame(t)
		caster, victim := g.Seats[0], g.Seats[1]
		ids := revealHand(victim, rhForest(), pbBear(), rhBolt())
		castAt(t, g, "Traumatic Revelation", "Sorcery", traumaticRevelationOracle, pbPlayer(victim.ID))
		pick := openVariant(t, g)
		if !pick.PickOptional || !sameIDs(pick.DiscardOptions, ids[1:2]) {
			t.Fatalf("options = %v optional %v, want only the creature card", pick.DiscardOptions, pick.PickOptional)
		}
		if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids[1:2]); err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if !victim.Graveyard.Contains(ids[1]) {
			t.Error("the chosen card was not discarded")
		}
		if n := len(incubatorsOf(g, caster.ID)); n != 0 {
			t.Errorf("incubated although a card was chosen")
		}
	})
	t.Run("you decline", func(t *testing.T) {
		g := newCatalogGame(t)
		caster, victim := g.Seats[0], g.Seats[1]
		ids := revealHand(victim, rhForest(), pbBear())
		castAt(t, g, "Traumatic Revelation", "Sorcery", traumaticRevelationOracle, pbPlayer(victim.ID))
		pick := openVariant(t, g)
		if err := g.ResolvePendingChoice(pick.ID, caster.ID, nil); err != nil {
			t.Fatalf("choose nothing: %v", err)
		}
		passPriorityAroundTable(t, g)
		if victim.Graveyard.Contains(ids[1]) {
			t.Error("a card was discarded although none was chosen")
		}
		if got := theIncubator(t, g, caster.ID).Counters["+1/+1"]; got != 3 {
			t.Errorf("counters = %d, want 3", got)
		}
		if n := len(incubatorsOf(g, victim.ID)); n != 0 {
			t.Error("the discarding player incubated")
		}
	})
	t.Run("nothing to choose", func(t *testing.T) {
		g := newCatalogGame(t)
		caster, victim := g.Seats[0], g.Seats[1]
		revealHand(victim, rhForest(), rhBolt())
		castAt(t, g, "Traumatic Revelation", "Sorcery", traumaticRevelationOracle, pbPlayer(victim.ID))
		noVariant(t, g)
		if got := theIncubator(t, g, caster.ID).Counters["+1/+1"]; got != 3 {
			t.Errorf("counters = %d, want 3", got)
		}
	})
}

func TestEyesOfGitaxiasIncubatesThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Eyes of Gitaxias", "Sorcery", eyesOfGitaxiasOracle, nil)
	passPriorityAroundTable(t, g)
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 3 {
		t.Errorf("counters = %d, want 3", got)
	}
	// castCatalogSpell seeds the spell into the hand and casts it, so the
	// net change is the one card drawn.
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1 from the draw", hand, me.Hand.Size())
	}
}

func TestMercilessRepurposingExilesThenIncubates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := pushCatalogPermanent(g, opp.ID, "Ogre", "Creature — Ogre", "", false)
	castCatalogSpell(t, g, "Merciless Repurposing", "Instant", mercilessRepurposingOracle, []game.TargetRef{pbCard(ogre)})
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, ogre) {
		t.Error("the creature was not exiled")
	}
	if got := theIncubator(t, g, me.ID).Counters["+1/+1"]; got != 3 {
		t.Errorf("counters = %d, want 3", got)
	}
}

// The exiled permanent's CONTROLLER incubates, for its mana value.
func TestExciseTheImperfectGivesTheIncubatorToTheControllerForItsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{4}",
		Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Excise the Imperfect", "Instant", exciseTheImperfectOracle, []game.TargetRef{pbCard(rock)})
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, rock) {
		t.Error("the permanent was not exiled")
	}
	if got := theIncubator(t, g, opp.ID).Counters["+1/+1"]; got != 4 {
		t.Errorf("counters = %d, want 4 (its mana value)", got)
	}
	if n := len(incubatorsOf(g, me.ID)); n != 0 {
		t.Error("the caster incubated; the permanent's controller does")
	}
}

// --- the template's guard ---------------------------------------------

func TestCheckTokenTemplateHoldsADoubleFacedTokenToItsShape(t *testing.T) {
	good := printedIncubatorToken()
	if why := checkTokenTemplate(good); why != "" {
		t.Fatalf("the Incubator is refused: %s", why)
	}
	for name, mutate := range map[string]func(t *tokenTemplate){
		"a third face":      func(t *tokenTemplate) { t.Back.Back = &tokenTemplate{} },
		"no back type line": func(t *tokenTemplate) { t.Back.Card.TypeLine = "" },
		"a back identity":   func(t *tokenTemplate) { t.Back.Card.OracleID = "x" },
		"a spelled layout":  func(t *tokenTemplate) { t.Card.Layout = game.LayoutTransform },
		"back ability, no text": func(t *tokenTemplate) {
			t.Back.Activated = good.Activated
		},
	} {
		tmpl := printedIncubatorToken()
		mutate(&tmpl)
		if checkTokenTemplate(tmpl) == "" {
			t.Errorf("%s: the template was accepted", name)
		}
	}
}

// --- the wire and the bot's move list --------------------------------

func incubatorOnTheWire(t *testing.T, g *game.Game, viewer uuid.UUID, id uuid.UUID) protocol.CardView {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID == id.String() {
			return c
		}
	}
	t.Fatalf("the Incubator is not on the wire")
	return protocol.CardView{}
}

func offersTransform(g *game.Game, p *game.Player, id uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, p.ID) {
		if m.Type == legal.TypeActivateAbility && m.Source == id {
			return true
		}
	}
	return false
}

// The client gets both faces, which one is up, the printed text and the
// ability row; the bot's move list offers the transform exactly while
// the front is up and {2} is in the pool.
func TestAnIncubatorOnTheWireAndInTheMoveList(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Converter Beast", "Creature — Phyrexian Beast", converterBeastOracle, nil)
	passPriorityAroundTable(t, g)
	inc := theIncubator(t, g, me.ID)

	if offersTransform(g, me, inc.InstanceID) {
		t.Error("the transform is offered with nothing to pay it")
	}
	payTwo(me)
	if !offersTransform(g, me, inc.InstanceID) {
		t.Error("the transform is not offered with {2} in the pool")
	}
	me.ManaPool = nil

	front := incubatorOnTheWire(t, g, me.ID, inc.InstanceID)
	if front.Layout != game.LayoutTransform || len(front.Faces) != 2 || front.ActiveFace != 0 ||
		front.Name != "Incubator" || front.TokenText != "{2}: Transform this token." {
		t.Errorf("front: layout %q faces %d active %d name %q text %q",
			front.Layout, len(front.Faces), front.ActiveFace, front.Name, front.TokenText)
	}
	if len(front.ActivatedAbilities) != 1 || front.Counters["+1/+1"] != 5 {
		t.Errorf("front: %d abilities, counters %v", len(front.ActivatedAbilities), front.Counters)
	}
	if front.Faces[1].Name != "Phyrexian Token" || front.Faces[1].TypeLine != "Token Artifact Creature — Phyrexian" {
		t.Errorf("back face on the wire = %+v", front.Faces[1])
	}

	transformIncubator(t, g, me, inc.InstanceID)
	back := incubatorOnTheWire(t, g, me.ID, inc.InstanceID)
	if back.ActiveFace != 1 || back.Name != "Phyrexian Token" || len(back.ActivatedAbilities) != 0 || back.TokenText != "" {
		t.Errorf("back: active %d name %q abilities %d text %q",
			back.ActiveFace, back.Name, len(back.ActivatedAbilities), back.TokenText)
	}
	payTwo(me)
	if offersTransform(g, me, inc.InstanceID) {
		t.Error("the transform is still offered on the back face")
	}
}
