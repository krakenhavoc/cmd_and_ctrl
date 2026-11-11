package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2S53RepulseOracle   = "6f7ef1d1-c441-41ba-b7d6-6b7727ffcf04"
	cmdr2S53PortRazerOracle = "653e4b68-3184-4352-ac91-127f7e6d7c12"
	cmdr2S53AdrixOracle     = "12a6cad9-eb42-43bd-9e68-aaf862cd83db"
)

func TestCmdr2S53CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdr2S53RepulseOracle:   "Repulse",
		cmdr2S53PortRazerOracle: "Port Razer",
		cmdr2S53AdrixOracle:     "Adrix and Nev, Twincasters",
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

// --- Repulse -----------------------------------------------------------------

func TestCmdr2S53RepulseBouncesACreatureAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Repulse", "Instant", cmdr2S53RepulseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after the spell, want %d (one drawn)", got, hand+1)
	}
}

func TestCmdr2S53RepulseRefusesANoncreatureAndFizzlesWithoutDrawing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Repulse", "Instant", cmdr2S53RepulseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err == nil {
		t.Fatal("an artifact was a legal target")
	}
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Repulse", "Instant", cmdr2S53RepulseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d: a fizzled spell draws nothing", got, hand)
	}
}

// --- Port Razer --------------------------------------------------------------

// Combat damage to a player untaps every creature its controller
// controls and adds a combat, in which it may not attack that player.
func TestCmdr2S53PortRazerUntapsAddsACombatAndCannotAttackTheSamePlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, p1, p2 := g.Seats[0], g.Seats[1], g.Seats[2]
	razer := pushDiesCreatureForTest(g, me.ID, "Port Razer", cmdr2S53PortRazerOracle, "Creature — Orc Pirate", 4, 4)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, p2.ID, "Their Tapped Bear", 2, 2)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, theirs).Tapped = true })
	life := p1.Life

	advanceToDeclareAttackersOf(t, g, 0)
	for _, id := range []uuid.UUID{razer, bear} {
		if err := g.DeclareAttacker(id, p1.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if !isTapped(g, razer) || !isTapped(g, bear) {
		t.Fatal("setup: the attackers are not tapped")
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if p1.Life != life-6 {
		t.Fatalf("defender at %d, want %d", p1.Life, life-6)
	}
	if isTapped(g, razer) || isTapped(g, bear) {
		t.Error("the trigger did not untap each creature you control")
	}
	if !isTapped(g, theirs) {
		t.Error("the trigger untapped an opponent's creature")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after the first combat: %+v, want an additional combat", next)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	second := bloodthirsterTargets(t, g, me.ID, razer)
	if second[p1.ID.String()] || !second[p2.ID.String()] || len(second) != 2 {
		t.Fatalf("extra combat offers %v, want only the two unattacked opponents", second)
	}
	err := g.Clone().DeclareAttacker(razer, p1.ID)
	var refused *game.AttackTargetRestrictionError
	if !errors.As(err, &refused) || !errors.Is(err, game.ErrIllegalAttackTarget) {
		t.Fatalf("re-attacking the same player: err = %v, want an attack-target restriction refusal", err)
	}
	// The Bear has no such restriction.
	if err := g.Clone().DeclareAttacker(bear, p1.ID); err != nil {
		t.Fatalf("the Bear attacking the same player again: %v", err)
	}
	if err := g.DeclareAttacker(razer, p2.ID); err != nil {
		t.Fatalf("attacking a different player: %v", err)
	}
}

// A blocked Port Razer deals no combat damage to a player: no untap and
// no additional combat.
func TestCmdr2S53PortRazerBlockedAddsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, p1 := g.Seats[0], g.Seats[1]
	razer := pushDiesCreatureForTest(g, me.ID, "Port Razer", cmdr2S53PortRazerOracle, "Creature — Orc Pirate", 4, 4)
	wall := pushVanillaCreature(g, p1.ID, "Wall", 0, 8)
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(razer, p1.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, razer); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if !isTapped(g, razer) {
		t.Error("a blocked Port Razer untapped")
	}
	if next := stepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after combat: %+v, want the main phase", next)
	}
}

// --- Adrix and Nev, Twincasters ----------------------------------------------

func TestCmdr2S53AdrixAndNevDoublesOnlyYourTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Adrix and Nev, Twincasters", "Legendary Creature — Merfolk Wizard", cmdr2S53AdrixOracle, 2, 2)
	makeTokens(t, g, me.ID, TreasureToken(), 3)
	if got := eoeCDCount(g, me.ID, "Treasure"); got != 6 {
		t.Errorf("your three Treasures became %d, want 6", got)
	}
	makeTokens(t, g, opp.ID, TreasureToken(), 1)
	if got := eoeCDCount(g, opp.ID, "Treasure"); got != 1 {
		t.Errorf("an opponent's one Treasure became %d, want 1", got)
	}
}

func TestCmdr2S53AdrixAndNevHasWardTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	twins := pushCatalogPermanent(g, me.ID, "Adrix and Nev, Twincasters", "Legendary Creature — Merfolk Wizard", cmdr2S53AdrixOracle, false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, twins)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {2} payment for the opponent targeting Adrix and Nev")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(twins) || !opp.Graveyard.Contains(blade) {
		t.Error("declining ward did not counter the spell")
	}
}

// --- Endurance ---------------------------------------------------------------

const cmdr2S53EnduranceOracle = "c85d824b-c190-4d04-ab99-918ad0e6516c"

// cmdr2S53BottomOf is the ids of the bottom n cards of p's library.
func cmdr2S53BottomOf(p *game.Player, n int) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	for i := 0; i < n && i < len(p.Library.Cards); i++ {
		out[p.Library.Cards[i].InstanceID] = true
	}
	return out
}

func TestCmdr2S53EnduranceBottomsTheTargetPlayersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := []uuid.UUID{
		pushGraveyardCardForTest(opp, "Their One"),
		pushGraveyardCardForTest(opp, "Their Two"),
		pushGraveyardCardForTest(opp, "Their Three"),
	}
	mine := pushGraveyardCardForTest(me, "My Card")
	lib := opp.Library.Size()
	top := opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID

	endurance := castAndResolveCreature(t, g, "Endurance", "Creature — Elemental Incarnation", cmdr2S53EnduranceOracle)
	assertKeywords(t, g, endurance, "flash", "reach")
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	if opp.Graveyard.Size() != 0 {
		t.Errorf("the target's graveyard has %d cards, want 0", opp.Graveyard.Size())
	}
	if got := opp.Library.Size(); got != lib+3 {
		t.Fatalf("library %d, want %d", got, lib+3)
	}
	bottom := cmdr2S53BottomOf(opp, 3)
	for _, id := range theirs {
		if !bottom[id] {
			t.Errorf("card %s is not among the bottom three of the library", id)
		}
	}
	if got := opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID; got != top {
		t.Error("the library was shuffled: its top card changed")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("another player's graveyard was touched")
	}
	if !g.Battlefield.Contains(endurance) {
		t.Error("a hard-cast Endurance left the battlefield")
	}
}

// "Up to one": choosing no player moves nothing, and the controller's
// own graveyard is a legal choice.
func TestCmdr2S53EnduranceMayChooseNoPlayerOrItsController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushGraveyardCardForTest(me, "My Card")
	castAndResolveCreature(t, g, "Endurance", "Creature — Elemental Incarnation", cmdr2S53EnduranceOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Endurance's trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("choosing no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(mine) {
		t.Fatal("no target was chosen and a graveyard moved")
	}

	castAndResolveCreature(t, g, "Endurance", "Creature — Elemental Incarnation", cmdr2S53EnduranceOracle)
	pickPlayer(t, g, me.ID, me.ID)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(mine) || !cmdr2S53BottomOf(me, 1)[mine] {
		t.Error("the controller's own graveyard did not go to the bottom of their library")
	}
}

// Evoke: a green card is exiled as the cost, the trigger still resolves,
// and Endurance is sacrificed. A nongreen card does not pay for it.
func TestCmdr2S53EnduranceEvokePitchesAGreenCardAndIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	theirs := pushGraveyardCardForTest(opp, "Their Card")
	blue := handCardFull(me, "Opt", "Instant", "{U}", "", []string{"U"})
	green := handCardFull(me, "Rampant Growth", "Sorcery", "{1}{G}", "", []string{"G"})
	endurance := handCardFull(me, "Endurance", "Creature — Elemental Incarnation", "{1}{G}{G}", cmdr2S53EnduranceOracle, []string{"G"})

	if err := g.CastSpell(me.ID, endurance, game.CastSpellParams{
		Strict: true, AlternativeCost: "evoke", AltCostIDs: []uuid.UUID{blue},
	}); err == nil {
		t.Fatal("a blue card paid for the evoke")
	}
	if err := g.CastSpell(me.ID, endurance, game.CastSpellParams{
		Strict: true, AlternativeCost: "evoke", AltCostIDs: []uuid.UUID{endurance},
	}); err == nil {
		t.Fatal("Endurance paid for its own evoke")
	}
	if err := g.CastSpell(me.ID, endurance, game.CastSpellParams{
		Strict: true, AlternativeCost: "evoke", AltCostIDs: []uuid.UUID{green},
	}); err != nil {
		t.Fatalf("evoking Endurance: %v", err)
	}
	if !g.Exile.Contains(green) {
		t.Error("the evoke pitch was not exiled")
	}
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	// The enters trigger and the evoke sacrifice are one CR 603.3b
	// batch; the enters trigger resolves first.
	if !answerTriggerOrderLastQueuedFirst(t, g) {
		t.Error("the evoke sacrifice and the enters trigger were not offered for ordering")
	}
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(theirs) || !cmdr2S53BottomOf(opp, 1)[theirs] {
		t.Error("the evoked Endurance's trigger did not resolve")
	}
	if g.Battlefield.Contains(endurance) || !me.Graveyard.Contains(endurance) {
		t.Error("the evoked Endurance was not sacrificed")
	}
}

// --- Siphon Insight ----------------------------------------------------------

const (
	cmdr2S53SiphonOracle = "aee82122-d13e-4175-b499-9dde718e4da9"
	cmdr2S53WealthOracle = "f2fae95d-074b-4c0c-99cd-f9385a424c05"
)

func TestCmdr2S53LateCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdr2S53EnduranceOracle: "Endurance",
		cmdr2S53SiphonOracle:    "Siphon Insight",
		cmdr2S53WealthOracle:    "Villainous Wealth",
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
	if ac := game.AlternativeCostByKey(cmdr2S53SiphonOracle, "flashback"); ac == nil || ac.ManaCost != "{1}{U}{B}" {
		t.Errorf("Siphon Insight: flashback %+v, want {1}{U}{B}", ac)
	}
}

// cmdr2S53Library replaces p's library with these cards, the first on top.
func cmdr2S53Library(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, len(cards))
	for i, c := range cards {
		// pushLibraryCardForTest pushes on the bottom.
		ids[i] = pushLibraryCardForTest(p, c)
	}
	return ids
}

// cmdr2S53SiphonPick casts Siphon Insight at opp from the hand and
// exiles `pick` of the two cards looked at.
func cmdr2S53SiphonPick(t *testing.T, g *game.Game, me, opp *game.Player, pick uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Siphon Insight", "Instant", cmdr2S53SiphonOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	ch := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if ch == nil || ch.ChooseMin != 1 || ch.ChooseMax != 1 {
		t.Fatalf("want a choose-exactly-one prompt, got %+v", ch)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{pick}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestCmdr2S53SiphonInsightExilesOneFaceDownAndBottomsTheOther(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := cmdr2S53Library(opp,
		game.Card{Name: "Top", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Second", TypeLine: "Creature — Goblin", ManaCost: "{R}", Colors: []string{"R"}},
		game.Card{Name: "Third", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Fourth", TypeLine: "Artifact", ManaCost: "{1}"},
	)
	if err := castCatalogSpellErr(t, g, "Siphon Insight", "Instant", cmdr2S53SiphonOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("Siphon Insight targeted its own caster")
	}
	castCatalogSpell(t, g, "Siphon Insight", "Instant", cmdr2S53SiphonOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	ch := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if ch == nil || ch.ChooseMin != 1 || ch.ChooseMax != 1 {
		t.Fatalf("want a choose-exactly-one prompt, got %+v", ch)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[2]}); err == nil {
		t.Fatal("a card below the top two was accepted")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)

	perm := e2Permission(g, lib[1])
	if perm == nil || perm.Player != me.ID || perm.CastOnly || !perm.AnyColor || perm.AnyType {
		t.Fatalf("permission %+v, want an any-color play grant to the caster", perm)
	}
	e2AssertFaceDownFor(t, g, lib[1], me.ID)
	if got := b12ZoneOf(g, lib[1]); got != game.ZoneExile {
		t.Errorf("picked card is in %q, want exile", got)
	}
	if n := len(opp.Library.Cards); n != 3 {
		t.Fatalf("library has %d cards, want 3", n)
	}
	if bottom := opp.Library.Cards[0].InstanceID; bottom != lib[0] {
		t.Error("the other card is not on the bottom of that library")
	}
	if top := opp.Library.Cards[2].InstanceID; top != lib[2] {
		t.Error("the third card is not the new top of the library")
	}

	// Mana as though it were mana of any color: blue pays for {R}. An
	// empty pool does not.
	if err := g.CastSpell(me.ID, lib[1], game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Fatal("the exiled card was cast with an empty pool")
	}
	fillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, lib[1], game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("casting the exiled {R} card with blue mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardForTest(g, lib[1]); c == nil || c.Controller != me.ID {
		t.Error("the stolen creature did not enter under the caster's control")
	}
	// The opponent holds no permission over their own exiled card.
	if err := g.Clone().CastSpell(opp.ID, lib[1], game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("the card's owner cast it from exile")
	}
}

// "Play": an exiled land is the caster's land drop.
func TestCmdr2S53SiphonInsightLetsItsCasterPlayAnExiledLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := cmdr2S53Library(opp,
		game.Card{Name: "Their Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Second", TypeLine: "Artifact", ManaCost: "{1}"},
	)
	cmdr2S53SiphonPick(t, g, me, opp, lib[0])
	if err := g.CastSpell(me.ID, lib[0], game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("playing the exiled land: %v", err)
	}
	if c := findBattlefieldCardForTest(g, lib[0]); c == nil || c.Controller != me.ID {
		t.Error("the exiled land is not on the battlefield under the caster's control")
	}
}

// A library of one card: it is exiled with no choice to make.
func TestCmdr2S53SiphonInsightWithOneCardExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := cmdr2S53Library(opp, game.Card{Name: "Only", TypeLine: "Artifact", ManaCost: "{1}"})
	castCatalogSpell(t, g, "Siphon Insight", "Instant", cmdr2S53SiphonOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, lib[0]); got != game.ZoneExile {
		t.Fatalf("the only card is in %q, want exile", got)
	}
	if perm := e2Permission(g, lib[0]); perm == nil || perm.Player != me.ID {
		t.Errorf("permission %+v, want the caster's", perm)
	}
}

// Flashback {1}{U}{B} from the graveyard: the same resolution, and the
// card is exiled afterwards.
func TestCmdr2S53SiphonInsightFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := cmdr2S53Library(opp,
		game.Card{Name: "Top", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Second", TypeLine: "Artifact", ManaCost: "{1}"},
	)
	id := seedGraveyardCard(t, g, "Siphon Insight", "Instant", cmdr2S53SiphonOracle)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	fillPoolColored(me, "U", 1)
	fillPoolColored(me, "B", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: "flashback", Targets: target,
	}); err == nil {
		t.Fatal("flashback was paid with {U}{B}; it costs {1}{U}{B}")
	}
	fillPool(me, 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: "flashback", Targets: target,
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	ch := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if ch == nil {
		t.Fatal("the flashed-back Siphon Insight asked nothing")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[0]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, lib[0]); got != game.ZoneExile {
		t.Errorf("picked card is in %q, want exile", got)
	}
	if got := b12ZoneOf(g, id); got != game.ZoneExile {
		t.Errorf("the flashed-back Siphon Insight is in %q, want exile", got)
	}
}

// --- Villainous Wealth -------------------------------------------------------

func cmdr2S53CastWealth(t *testing.T, g *game.Game, me, opp *game.Player, x int) {
	t.Helper()
	toMainForCost(t, g)
	id := handCardFull(me, "Villainous Wealth", "Sorcery", "{X}{B}{G}{U}", cmdr2S53WealthOracle, []string{"B", "G", "U"})
	fillPoolColored(me, "B", 1)
	fillPoolColored(me, "G", 1)
	fillPoolColored(me, "U", 1)
	fillPool(me, x)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, XValue: x, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("casting Villainous Wealth for X=%d: %v", x, err)
	}
	passPriorityAroundTable(t, g)
}

// X = 3: the top three are exiled, and any number of spells with mana
// value 3 or less among them are cast free; a land and a bigger spell
// stay in exile.
func TestCmdr2S53VillainousWealthCastsAnyNumberAtMostX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := cmdr2S53Library(opp,
		game.Card{Name: "Three", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Card{Name: "Four", TypeLine: "Artifact", ManaCost: "{4}"},
		game.Card{Name: "One", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Below", TypeLine: "Artifact", ManaCost: "{1}"},
	)
	cmdr2S53CastWealth(t, g, me, opp, 3)
	for _, id := range lib[:3] {
		if !inExile(g, id) {
			t.Fatalf("%v was among the top three and is not in exile", id)
		}
	}
	if got := b12ZoneOf(g, lib[3]); got != game.ZoneLibrary {
		t.Fatalf("the fourth card is in %q, want the library", got)
	}
	if err := g.CastSpell(me.ID, lib[1], game.CastSpellParams{FromZone: "exile"}); !errors.Is(err, game.ErrSpellManaValueTooHigh) {
		t.Fatalf("cast the mana value 4 card: err = %v, want ErrSpellManaValueTooHigh", err)
	}
	if err := g.Clone().CastSpell(opp.ID, lib[0], game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("the cards' owner cast one of them")
	}
	if err := g.CastSpell(me.ID, lib[0], game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast the mana value 3 card free: %v", err)
	}
	if err := g.CastSpell(me.ID, lib[2], game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast the mana value 1 card after the 3: %v", err)
	}
	if !inExile(g, lib[1]) {
		t.Error("the uncast card left exile")
	}
}

// A land among the exiled cards is not cast, and X = 0 exiles nothing.
func TestCmdr2S53VillainousWealthStrandsALandAndXZeroDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := cmdr2S53Library(opp,
		game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Below", TypeLine: "Artifact", ManaCost: "{1}"},
	)
	cmdr2S53CastWealth(t, g, me, opp, 0)
	if n := len(opp.Library.Cards); n != 2 {
		t.Fatalf("X = 0 left %d cards in the library, want 2", n)
	}
	cmdr2S53CastWealth(t, g, me, opp, 1)
	if !inExile(g, lib[0]) || len(opp.Library.Cards) != 1 {
		t.Fatal("X = 1 did not exile exactly the top card")
	}
	if err := g.CastSpell(me.ID, lib[0], game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("a land was played off Villainous Wealth")
	}
}
