package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82b19MarchOracle     = "dd45c284-c34c-4085-8066-dbeee223fb6e"
	s82b19TawnosSolemn    = "871fbbd6-76a0-4f79-b89f-afee6ebe5a59"
	s82b19TawnosToymaker  = "48cce741-cfbd-4ed9-a7ac-38880867a286"
	s82b19DrafnaOracle    = "ed2e8718-4b20-4310-96ce-05a29a24ef32"
	s82b19UrzaOracle      = "6a448862-e8d8-4a08-8e3d-5460afa29b4d"
	s82b19CrucibleOracle  = "d0e80414-abde-4f5e-9539-3efc83aa507f"
	s82b19AshnodOracle    = "dcd5ab1b-adc5-486e-a1e0-636af6ee9470"
	s82b19ThopterShopOrac = "b2f9e07b-64b7-40b3-9a5e-5f1d59b35af7"
)

func TestS82BroB19EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82b19MarchOracle:     "March of Progress",
		s82b19TawnosSolemn:    "Tawnos, Solemn Survivor",
		s82b19TawnosToymaker:  "Tawnos, the Toymaker",
		s82b19DrafnaOracle:    "Drafna, Founder of Lat-Nam",
		s82b19UrzaOracle:      "Urza, Prince of Kroog",
		s82b19CrucibleOracle:  "Wondrous Crucible",
		s82b19AshnodOracle:    "Ashnod the Uncaring",
		s82b19ThopterShopOrac: "Thopter Shop",
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

// s82b19Tokens lists the battlefield tokens a player controls with a name.
func s82b19Tokens(g *game.Game, controller uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.IsToken() && c.Controller == controller {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

// s82b19ArtifactCreature seeds a non-catalog artifact creature.
func s82b19ArtifactCreature(g *game.Game, owner uuid.UUID, name string, p, tough int) uuid.UUID {
	return b12Creature(g, owner, name, "Artifact Creature — Construct", p, tough)
}

// --- Thopter Shop ------------------------------------------------------------

func TestS82BroB19ThopterShopDrawsOncePerTurnForADyingArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Thopter Shop", "Artifact", s82b19ThopterShopOrac, 0, 0)
	a := s82b19ArtifactCreature(g, me.ID, "Construct A", 1, 1)
	b := s82b19ArtifactCreature(g, me.ID, "Construct B", 1, 1)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(a)
		_ = g.DestroyPermanentForEffect(b)
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Fatalf("drew %d cards for two artifact creatures dying, want 1 (once each turn)", got)
	}
	// A third death the same turn draws nothing more.
	c := s82b19ArtifactCreature(g, me.ID, "Construct C", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(c) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards after a third death this turn, want still 1", got)
	}
}

func TestS82BroB19ThopterShopIgnoresNonArtifactsAndOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Thopter Shop", "Artifact", s82b19ThopterShopOrac, 0, 0)
	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := s82b19ArtifactCreature(g, opp.ID, "Their Construct", 1, 1)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(bear)
		_ = g.DestroyPermanentForEffect(theirs)
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d cards, want 0 (a non-artifact and an opponent's artifact creature died)", got)
	}
}

func TestS82BroB19ThopterShopMakesAFlyingThopter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	shop := b12Push(g, me.ID, "Thopter Shop", "Artifact", s82b19ThopterShopOrac, 0, 0)
	if err := activateOn(t, g, me.ID, shop, 0, uuid.Nil); err == nil {
		t.Fatal("the Thopter ability was activated without {2}{W}")
	}
	addPool(t, g, me.ID, "{W}{W}{W}")
	if err := activateOn(t, g, me.ID, shop, 0, uuid.Nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	toks := s82b19Tokens(g, me.ID, "Thopter")
	if len(toks) != 1 {
		t.Fatalf("%d Thopter tokens, want 1", len(toks))
	}
	c := layeredCard(t, g, toks[0])
	if p, tough := ptOf(t, g, toks[0]); !c.IsArtifact() || !c.IsCreature() || p != 1 || tough != 1 {
		t.Errorf("the token is not a 1/1 artifact creature: %v %d/%d", c.Effective().Types, p, tough)
	}
	assertKeywords(t, g, toks[0], "flying")
	if !layeredCard(t, g, shop).Tapped {
		t.Error("the Shop was not tapped to pay")
	}
}

// --- Urza, Prince of Kroog -----------------------------------------------------

func TestS82BroB19UrzaPumpsArtifactCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Urza, Prince of Kroog", "Legendary Creature — Human Artificer", s82b19UrzaOracle, 2, 3)
	mine := s82b19ArtifactCreature(g, me.ID, "Construct", 1, 1)
	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := s82b19ArtifactCreature(g, opp.ID, "Their Construct", 1, 1)
	for id, want := range map[uuid.UUID][2]int{mine: {3, 3}, bear: {2, 2}, theirs: {1, 1}} {
		if p, tough := ptOf(t, g, id); p != want[0] || tough != want[1] {
			t.Errorf("%v is %d/%d, want %d/%d", id, p, tough, want[0], want[1])
		}
	}
}

func TestS82BroB19UrzaCopiesAnArtifactAsAOneOneSoldier(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	urza := b12Push(g, me.ID, "Urza, Prince of Kroog", "Legendary Creature — Human Artificer", s82b19UrzaOracle, 2, 3)
	rock := b12Permanent(g, me.ID, "Odd Rock", "Artifact")
	addPool(t, g, me.ID, "{G}{G}{G}{G}{G}{G}")
	if err := activateOn(t, g, me.ID, urza, 0, rock); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	toks := s82b19Tokens(g, me.ID, "Odd Rock")
	if len(toks) != 1 {
		t.Fatalf("%d Odd Rock tokens, want 1", len(toks))
	}
	c := layeredCard(t, g, toks[0])
	if !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Soldier") {
		t.Errorf("the copy is not an artifact Soldier creature: %v %v", c.Effective().Types, c.Effective().Subtypes)
	}
	// Base 1/1, then Urza's +2/+2 for being an artifact creature.
	if p, tough := ptOf(t, g, toks[0]); p != 3 || tough != 3 {
		t.Errorf("the copy is %d/%d, want 3/3 (base 1/1 plus Urza's +2/+2)", p, tough)
	}
}

func TestS82BroB19UrzaRefusesAnOpponentsArtifactAndANonArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	urza := b12Push(g, me.ID, "Urza, Prince of Kroog", "Legendary Creature — Human Artificer", s82b19UrzaOracle, 2, 3)
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	addPool(t, g, me.ID, "{G}{G}{G}{G}{G}{G}")
	for _, target := range []uuid.UUID{theirs, bear} {
		if err := activateOn(t, g, me.ID, urza, 0, target); err == nil {
			t.Errorf("target %v was legal", target)
		}
	}
}

// --- March of Progress ---------------------------------------------------------

func s82b19CastMarch(t *testing.T, g *game.Game, overload bool, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "March of Progress", TypeLine: "Sorcery", ManaCost: "{2}{U}",
		OracleID: s82b19MarchOracle, Owner: active.ID, Controller: active.ID,
	})
	toMainForCost(t, g)
	params := game.CastSpellParams{Targets: targets, Strict: true}
	if overload {
		params.AlternativeCost = "overload"
		params.Targets = nil
		addPool(t, g, active.ID, "{U}{U}{U}{U}{U}{U}{U}")
	} else {
		addPool(t, g, active.ID, "{U}{U}{U}")
	}
	return g.CastSpell(active.ID, id, params)
}

func TestS82BroB19MarchOfProgressCopiesOneArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := s82b19ArtifactCreature(g, me.ID, "Alpha Construct", 2, 2)
	s82b19ArtifactCreature(g, me.ID, "Beta Construct", 3, 3)
	if err := s82b19CastMarch(t, g, false, []game.TargetRef{{Kind: game.TargetCard, ID: a}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(s82b19Tokens(g, me.ID, "Alpha Construct")); got != 1 {
		t.Errorf("%d Alpha Construct tokens, want 1", got)
	}
	if got := len(s82b19Tokens(g, me.ID, "Beta Construct")); got != 0 {
		t.Errorf("%d Beta Construct tokens, want 0 (only the target is copied)", got)
	}
}

func TestS82BroB19MarchOfProgressRefusesANonArtifactAndAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := s82b19ArtifactCreature(g, opp.ID, "Their Construct", 1, 1)
	for _, target := range []uuid.UUID{bear, theirs} {
		if err := s82b19CastMarch(t, g, false, []game.TargetRef{{Kind: game.TargetCard, ID: target}}); err == nil {
			t.Errorf("target %v was legal", target)
		}
	}
}

func TestS82BroB19MarchOfProgressOverloadCopiesEveryArtifactCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	s82b19ArtifactCreature(g, me.ID, "Alpha Construct", 2, 2)
	s82b19ArtifactCreature(g, me.ID, "Beta Construct", 3, 3)
	b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	s82b19ArtifactCreature(g, opp.ID, "Their Construct", 1, 1)
	if err := s82b19CastMarch(t, g, true, nil); err != nil {
		t.Fatalf("cast overloaded: %v", err)
	}
	passPriorityAroundTable(t, g)
	for name, want := range map[string]int{"Alpha Construct": 1, "Beta Construct": 1, "Plain Bear": 0, "Their Construct": 0} {
		if got := len(s82b19Tokens(g, me.ID, name)); got != want {
			t.Errorf("%d %s tokens under my control, want %d", got, name, want)
		}
	}
}

// --- Wondrous Crucible ---------------------------------------------------------

func TestS82BroB19WondrousCrucibleGivesEveryPermanentWardTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Wondrous Crucible", "Artifact", s82b19CrucibleOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("targeting a creature the Crucible's controller controls raised no ward {2} payment")
	}
}

func s82b19CrucibleEnd(t *testing.T, g *game.Game, me *game.Player, library, graveyard []game.Card) []uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	b12Push(g, me.ID, "Wondrous Crucible", "Artifact", s82b19CrucibleOracle, 0, 0)
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		for _, c := range library {
			c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
			me.Library.PushTop(c)
		}
		for _, c := range graveyard {
			c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
			me.Graveyard.PushTop(c)
			ids = append(ids, c.InstanceID)
		}
	})
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	return ids
}

func TestS82BroB19WondrousCrucibleMillsThenCopiesAndOffersAFreeCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := []game.Card{
		{Name: "Library Land A", TypeLine: "Basic Land — Forest"},
		{Name: "Library Land B", TypeLine: "Basic Land — Forest"},
		{Name: "Deep Land", TypeLine: "Basic Land — Forest"},
	}
	yard := []game.Card{
		{Name: "Graveyard Land", TypeLine: "Basic Land — Forest"},
		{Name: "Test Spell", TypeLine: "Sorcery", ManaCost: "{4}{R}"},
	}
	ids := s82b19CrucibleEnd(t, g, me, lib, yard)
	if got := me.Library.Size(); got != 1 {
		t.Errorf("library has %d cards, want 1 (two milled)", got)
	}
	if !inExile(g, ids[1]) {
		t.Error("the only nonland card in the graveyard was not exiled")
	}
	if !me.Graveyard.Contains(ids[0]) {
		t.Error("a land card was exiled; the pick is a nonland card")
	}
	cp := copyOf(t, g, "Test Spell")
	if err := castCopy(g, me.ID, cp.InstanceID, true); err != nil {
		t.Fatalf("cast the copy without paying its mana cost: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left after the cast resolved, want none", n)
	}
	if !inExile(g, ids[1]) {
		t.Error("the original left exile; only the copy is cast")
	}
}

func TestS82BroB19WondrousCrucibleWithOnlyLandsDoesNothingAfterTheMill(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := []game.Card{
		{Name: "Library Land A", TypeLine: "Basic Land — Forest"},
		{Name: "Library Land B", TypeLine: "Basic Land — Forest"},
	}
	s82b19CrucibleEnd(t, g, me, lib, nil)
	if got := me.Library.Size(); got != 0 {
		t.Errorf("library has %d cards, want 0 (two milled)", got)
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies, want none: only lands were in the graveyard", n)
	}
}

// --- Tawnos, the Toymaker ------------------------------------------------------

func TestS82BroB19TawnosToymakerCopiesABeastAsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Tawnos, the Toymaker", "Legendary Creature — Human Artificer", s82b19TawnosToymaker, 3, 5)
	s82b19CastCreature(t, g, "Test Beast", "Creature — Beast", 2, 2)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	toks := s82b19Tokens(g, me.ID, "Test Beast")
	if len(toks) != 1 {
		t.Fatalf("%d Test Beast tokens, want 1", len(toks))
	}
	c := layeredCard(t, g, toks[0])
	if !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Beast") {
		t.Errorf("the copy is not an artifact Beast creature: %v %v", c.Effective().Types, c.Effective().Subtypes)
	}
	if got := len(battlefieldIDsNamed(g, "Test Beast")); got != 2 {
		t.Errorf("%d Test Beasts on the battlefield, want the spell and its copy", got)
	}
}

func TestS82BroB19TawnosToymakerBirdCanBeDeclinedAndABearDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Tawnos, the Toymaker", "Legendary Creature — Human Artificer", s82b19TawnosToymaker, 3, 5)
	s82b19CastCreature(t, g, "Test Bird", "Creature — Bird", 1, 1)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := len(s82b19Tokens(g, me.ID, "Test Bird")); got != 0 {
		t.Errorf("%d Bird tokens after declining, want 0", got)
	}
	s82b19CastCreature(t, g, "Test Bear", "Creature — Bear", 2, 2)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("a Bear spell raised Tawnos's prompt")
	}
	passPriorityAroundTable(t, g)
	if got := len(s82b19Tokens(g, me.ID, "Test Bear")); got != 0 {
		t.Errorf("%d Bear tokens, want 0", got)
	}
}

// s82b19CastCreature casts a free creature spell from the active hand.
func s82b19CastCreature(t *testing.T, g *game.Game, name, typeLine string, p, tough int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Power: p, Toughness: tough, Owner: active.ID, Controller: active.ID,
	})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// --- Drafna, Founder of Lat-Nam -------------------------------------------------

func TestS82BroB19DrafnaBouncesYourArtifactOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	drafna := b12Push(g, me.ID, "Drafna, Founder of Lat-Nam", "Legendary Creature — Human Artificer Advisor", s82b19DrafnaOracle, 2, 1)
	rock := b12Permanent(g, me.ID, "Odd Rock", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	addPool(t, g, me.ID, "{U}{U}{U}{U}")
	for _, target := range []uuid.UUID{theirs, bear} {
		if err := activateOn(t, g, me.ID, drafna, 0, target); err == nil {
			t.Errorf("target %v was legal", target)
		}
	}
	if err := activateOn(t, g, me.ID, drafna, 0, rock); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !me.Hand.Contains(rock) {
		t.Error("the artifact was not returned to its owner's hand")
	}
}

func TestS82BroB19DrafnaCopiesAnArtifactSpellYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	drafna := b12Push(g, me.ID, "Drafna, Founder of Lat-Nam", "Legendary Creature — Human Artificer Advisor", s82b19DrafnaOracle, 2, 1)
	rock := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: rock, Name: "Odd Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Plain Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Bear: %v", err)
	}
	addPool(t, g, me.ID, "{G}{G}{G}")
	if err := activateOn(t, g, me.ID, drafna, 1, bear); err == nil {
		t.Error("a creature spell that isn't an artifact was a legal target")
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Rock: %v", err)
	}
	if err := activateOn(t, g, me.ID, drafna, 1, rock); err != nil {
		t.Fatalf("activate with an artifact spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(battlefieldIDsNamed(g, "Odd Rock")); got != 2 {
		t.Errorf("%d Odd Rocks on the battlefield, want the spell and its copy", got)
	}
	if got := len(s82b19Tokens(g, me.ID, "Odd Rock")); got != 1 {
		t.Errorf("%d Odd Rock tokens, want 1 (the copy becomes a token)", got)
	}
	if !layeredCard(t, g, drafna).Tapped {
		t.Error("Drafna was not tapped to pay")
	}
}

// --- Tawnos, Solemn Survivor ----------------------------------------------------

func s82b19FillLibrary(g *game.Game, me *game.Player, n int) {
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		for i := 0; i < n; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		}
	})
}

func TestS82BroB19TawnosSolemnCopiesAnArtifactTokenAndMills(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tawnos := b12Push(g, me.ID, "Tawnos, Solemn Survivor", "Legendary Creature — Human Artificer", s82b19TawnosSolemn, 1, 3)
	food := pushToken(g, me.ID, FoodToken())
	s82b19FillLibrary(g, me, 5)
	nonToken := b12Permanent(g, me.ID, "Odd Rock", "Artifact")
	addPool(t, g, me.ID, "{U}{U}{U}{U}")
	if err := activateOn(t, g, me.ID, tawnos, 0, nonToken); err == nil {
		t.Fatal("a nontoken artifact was a legal target")
	}
	if err := activateOn(t, g, me.ID, tawnos, 0, food); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(s82b19Tokens(g, me.ID, "Food")); got != 2 {
		t.Errorf("%d Food tokens, want 2 (the original and its copy)", got)
	}
	if got := me.Library.Size(); got != 3 {
		t.Errorf("library has %d cards, want 3 (two milled)", got)
	}
}

func TestS82BroB19TawnosSolemnMillsEvenWithNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tawnos := b12Push(g, me.ID, "Tawnos, Solemn Survivor", "Legendary Creature — Human Artificer", s82b19TawnosSolemn, 1, 3)
	s82b19FillLibrary(g, me, 5)
	addPool(t, g, me.ID, "{U}{U}")
	if err := activateOn(t, g, me.ID, tawnos, 0, uuid.Nil); err != nil {
		t.Fatalf("activate with up to zero targets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != 3 {
		t.Errorf("library has %d cards, want 3 (two milled)", got)
	}
}

func TestS82BroB19TawnosSolemnSecondAbilityCopiesTheExiledCardAsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tawnos := b12Push(g, me.ID, "Tawnos, Solemn Survivor", "Legendary Creature — Human Artificer", s82b19TawnosSolemn, 1, 3)
	t1 := pushToken(g, me.ID, FoodToken())
	t2 := pushToken(g, me.ID, ClueToken())
	bear := gyCard(me, "Grave Bear", "Creature — Bear", "{1}{G}", "")
	land := gyCard(me, "Grave Land", "Basic Land — Forest", "", "")
	addPool(t, g, me.ID, "{G}{W}{U}{B}")
	err := g.ActivateCatalogAbility(me.ID, tawnos, 1, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{t1, t2}, ExileIDs: []uuid.UUID{land}})
	if err == nil {
		t.Fatal("a land card was a legal card to exile")
	}
	err = g.ActivateCatalogAbility(me.ID, tawnos, 1, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{t1, t2}, ExileIDs: []uuid.UUID{bear}})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(t1) || g.Battlefield.Contains(t2) {
		t.Error("the artifact tokens were not sacrificed")
	}
	if !inExile(g, bear) {
		t.Error("the creature card was not exiled")
	}
	toks := s82b19Tokens(g, me.ID, "Grave Bear")
	if len(toks) != 1 {
		t.Fatalf("%d Grave Bear tokens, want 1", len(toks))
	}
	c := layeredCard(t, g, toks[0])
	if !c.IsArtifact() || !c.IsCreature() {
		t.Errorf("the copy is not an artifact creature: %v", c.Effective().Types)
	}
	// Sorcery speed only.
	advanceTo(t, g, game.StepEnd)
	addPool(t, g, me.ID, "{G}{W}{U}{B}")
	if err := g.ActivateCatalogAbility(me.ID, tawnos, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated at the end step, but it is a sorcery-speed ability")
	}
}

// --- Ashnod the Uncaring ---------------------------------------------------------

func TestS82BroB19AshnodCopiesAnAbilityActivatedWithASacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Ashnod the Uncaring", "Legendary Creature — Human Artificer", s82b19AshnodOracle, 1, 4)
	food := pushToken(g, me.ID, FoodToken())
	addPool(t, g, me.ID, "{G}{G}")
	life := me.Life
	if err := activateOn(t, g, me.ID, food, 0, uuid.Nil); err != nil {
		t.Fatalf("activate the Food: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d -> %d, want +6 (the Food's ability and its copy)", life, me.Life)
	}
}

func TestS82BroB19AshnodMayDeclineTheCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Ashnod the Uncaring", "Legendary Creature — Human Artificer", s82b19AshnodOracle, 1, 4)
	food := pushToken(g, me.ID, FoodToken())
	addPool(t, g, me.ID, "{G}{G}")
	life := me.Life
	if err := activateOn(t, g, me.ID, food, 0, uuid.Nil); err != nil {
		t.Fatalf("activate the Food: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3 (no copy)", life, me.Life)
	}
}

func TestS82BroB19AshnodIgnoresAnActivationWithNoSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Ashnod the Uncaring", "Legendary Creature — Human Artificer", s82b19AshnodOracle, 1, 4)
	shop := b12Push(g, me.ID, "Thopter Shop", "Artifact", s82b19ThopterShopOrac, 0, 0)
	addPool(t, g, me.ID, "{W}{W}{W}")
	if err := activateOn(t, g, me.ID, shop, 0, uuid.Nil); err != nil {
		t.Fatalf("activate the Shop: %v", err)
	}
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("Ashnod asked about copying an ability that sacrificed nothing")
	}
	passPriorityAroundTable(t, g)
	if got := len(s82b19Tokens(g, me.ID, "Thopter")); got != 1 {
		t.Errorf("%d Thopters, want 1", got)
	}
}
