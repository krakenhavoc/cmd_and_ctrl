package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b119_cards_test.go — OLD batch 119 (Ikoria: Lair of Behemoths,
// #5512): eleven cards. Each test plays the real catalog entry through the
// engine and asserts what its text says.

const (
	oldB119WolfbearOracle   = "40f5f8b9-5c2c-41d4-8b83-55a2e51ced64"
	oldB119WolverineOracle  = "a3b226b9-f149-4578-882b-77447608380a"
	oldB119DolphinOracle    = "b70b739b-b147-4abe-8e48-26e7ea02a9f6"
	oldB119TacticianOracle  = "efee89a6-b730-45b6-bb84-33de0f25ad47"
	oldB119SnapdaxOracle    = "4b0b4d16-1d2b-497e-a1ae-bfb89eece665"
	oldB119AnticipateOracle = "a97da5a9-17fc-4367-81e1-17ff81601a63"
	oldB119ConnectionOracle = "d74de74f-be0f-4809-8b7a-1bd58ee85551"
	oldB119StampedeOracle   = "ea91be73-40f5-461a-bb39-d60216829a53"
	oldB119CommandoOracle   = "04cffc41-05da-45fb-82da-2c4d4108c7e1"
	oldB119IllunaOracle     = "42b10f0f-19cd-45d2-a2db-d5463ddc5e5b"
	oldB119RescuerOracle    = "c9de9f0b-bf07-4171-b54d-69033424a212"
)

func TestOldB119EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB119WolfbearOracle:   "Exuberant Wolfbear",
		oldB119WolverineOracle:  "Spelleater Wolverine",
		oldB119DolphinOracle:    "Phase Dolphin",
		oldB119TacticianOracle:  "Snare Tactician",
		oldB119SnapdaxOracle:    "Mythos of Snapdax",
		oldB119AnticipateOracle: "Anticipate",
		oldB119ConnectionOracle: "Tentative Connection",
		oldB119StampedeOracle:   "Lead the Stampede",
		oldB119CommandoOracle:   "Nightsquad Commando",
		oldB119IllunaOracle:     "Mythos of Illuna",
		oldB119RescuerOracle:    "Valiant Rescuer",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v, want Full with no caveats", name, spec, ok)
		}
	}
}

func oldB119PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := effectstest.GaLayered(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

func TestOldB119WolfbearGivesAHumanItsSizeUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolf := effectstest.B12Push(g, me.ID, "Exuberant Wolfbear", "Creature — Wolf Bear", oldB119WolfbearOracle, 4, 4)
	human := effectstest.B12Creature(g, me.ID, "Villager", "Creature — Human", 1, 1)
	effectstest.AttackWithTrigger(t, g, wolf, opp.ID, human)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := oldB119PT(t, g, human); p != 4 || tg != 4 {
		t.Errorf("the Human is %d/%d, want 4/4", p, tg)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tg := oldB119PT(t, g, human); p != 1 || tg != 1 {
		t.Errorf("the Human is %d/%d next turn, want 1/1 again", p, tg)
	}
}

func TestOldB119WolfbearIsOptionalAndOnlyTargetsYourHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolf := effectstest.B12Push(g, me.ID, "Exuberant Wolfbear", "Creature — Wolf Bear", oldB119WolfbearOracle, 4, 4)
	human := effectstest.B12Creature(g, me.ID, "Villager", "Creature — Human", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Human", "Creature — Human", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wolf, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	for _, bad := range []uuid.UUID{bear, theirs} {
		ch := effectstest.PickTargetPrompt(t, g)
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("the trigger accepted %s", bad)
		}
	}
	effectstest.AnswerPickTarget(t, g, human)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := oldB119PT(t, g, human); p != 1 || tg != 1 {
		t.Errorf("the Human is %d/%d after declining, want 1/1", p, tg)
	}
}

// oldB119CastBlank casts and resolves an uncatalogued spell of the given
// type line, which leaves it in its owner's graveyard.
func oldB119CastBlank(t *testing.T, g *game.Game, p *game.Player, typeLine string) uuid.UUID {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(p, "Blank "+typeLine, typeLine, "", "", nil)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", typeLine, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !p.Graveyard.Contains(id) {
		t.Fatalf("the %s is not in the graveyard", typeLine)
	}
	return id
}

func TestOldB119WolverineHasDoubleStrikeWithThreeInstantsOrSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolv := effectstest.B12Push(g, me.ID, "Spelleater Wolverine", "Creature — Wolverine", oldB119WolverineOracle, 3, 2)
	oldB119CastBlank(t, g, me, "Instant")
	oldB119CastBlank(t, g, me, "Sorcery")
	effectstest.KillCreature(t, g, me.ID, effectstest.B12Creature(g, me.ID, "Dead Bear", "Creature — Bear", 2, 2))
	if effectstest.HasEffectiveKeyword(t, g, wolv, "double strike") {
		t.Fatal("double strike with two instant and sorcery cards (and a creature card)")
	}
	third := oldB119CastBlank(t, g, me, "Instant")
	if !effectstest.HasEffectiveKeyword(t, g, wolv, "double strike") {
		t.Fatal("no double strike with three instant and sorcery cards")
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, third); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if effectstest.HasEffectiveKeyword(t, g, wolv, "double strike") {
		t.Error("double strike stayed after a spell left the graveyard")
	}
	// An opponent's graveyard does not count.
	for range 3 {
		effectstest.PushToGraveyard(g, opp, game.Card{Name: "Their Spell", TypeLine: "Instant"})
	}
	if effectstest.HasEffectiveKeyword(t, g, wolv, "double strike") {
		t.Error("an opponent's graveyard counted")
	}
}

func TestOldB119PhaseDolphinMakesAnotherAttackerUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dolphin := effectstest.B12Push(g, me.ID, "Phase Dolphin", "Creature — Elemental Whale", oldB119DolphinOracle, 1, 4)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Stay Home", "Creature — Bear", 2, 2)
	cantBeBlocked := func(id uuid.UUID) bool {
		return effectstest.GaLayered(t, g, id).Effective().Restrictions&game.CantBeBlocked != 0
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{dolphin, ally} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	for _, bad := range []uuid.UUID{dolphin, other} {
		ch := effectstest.PickTargetPrompt(t, g)
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("the trigger accepted %s (itself or a non-attacker)", bad)
		}
	}
	effectstest.AnswerPickTarget(t, g, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if !cantBeBlocked(ally) {
		t.Error("the other attacker can still be blocked")
	}
	if cantBeBlocked(dolphin) || cantBeBlocked(other) {
		t.Error("the Dolphin or a non-target became unblockable")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if cantBeBlocked(ally) {
		t.Error("unblockable lasted past the turn")
	}
}

func TestOldB119SnareTacticianTapsACreatureEachTimeYouCycle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Snare Tactician", "Creature — Human Soldier", oldB119TacticianOracle, 2, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)
	tapped := func(id uuid.UUID) bool { c := effectstest.GaLayered(t, g, id); return c.Tapped }
	effectstest.CycleFromHand(t, g, "Scrounging Skyray", "Creature — Bird", effectstest.ScroungingSkyrayOracle, "{C}{C}")
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Error("the trigger accepted my own creature")
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !tapped(theirs) || tapped(other) || tapped(mine) {
		t.Errorf("tapped: target=%v other=%v mine=%v, want the target only", tapped(theirs), tapped(other), tapped(mine))
	}
	effectstest.CycleFromHand(t, g, "Scrounging Skyray", "Creature — Bird", effectstest.ScroungingSkyrayOracle, "{C}{C}")
	effectstest.AnswerPickTarget(t, g, other)
	effectstest.PassPriorityAroundTable(t, g)
	if !tapped(other) {
		t.Error("the second cycle did not tap a second creature")
	}
}

func oldB119Library(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		ids[i] = c.InstanceID
		p.Library.PushTop(c)
	}
	return ids
}

// oldB119BottomInOrder answers the "rest on the bottom in any order" prompt
// with `order` (shallowest first) and returns the library's real bottom.
func oldB119BottomInOrder(t *testing.T, g *game.Game, p *game.Player, order ...uuid.UUID) {
	t.Helper()
	ch := effectstest.PutInLibraryChoiceFor(g, p.ID)
	if ch == nil {
		t.Fatal("no put_in_library prompt for the rest of the cards")
	}
	if err := g.ResolvePutInLibrary(ch.ID, p.ID, order, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := effectstest.LibraryBottomIDs(p, len(order)); !effectstest.SameIDs(got, order) {
		t.Errorf("bottom of the library %v, want %v", got, order)
	}
}

func oldB119In(z *game.Zone, id uuid.UUID) bool {
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestOldB119AnticipateTakesOneOfTheTopThreeAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := oldB119Library(me,
		game.Card{Name: "A", TypeLine: "Sorcery"},
		game.Card{Name: "B", TypeLine: "Sorcery"},
		game.Card{Name: "C", TypeLine: "Sorcery"},
		game.Card{Name: "D", TypeLine: "Sorcery"},
	)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Anticipate", "Instant", oldB119AnticipateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	oldB119BottomInOrder(t, g, me, ids[2], ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 || !oldB119In(me.Hand, ids[1]) {
		t.Errorf("hand %d -> %d, took the chosen card=%v", hand, got, oldB119In(me.Hand, ids[1]))
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != ids[3] {
		t.Error("the fourth card is not on top: more than three were looked at or the rest went elsewhere")
	}
	if me.Library.Size() != 3 || !oldB119In(me.Library, ids[0]) || !oldB119In(me.Library, ids[2]) {
		t.Errorf("library has %d cards, want the two rest cards under the fourth", me.Library.Size())
	}
}

func TestOldB119TentativeConnectionStealsUntapsAndHastesUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Menacing Goblin", "Creature — Goblin", "", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Tentative Connection", "Sorcery", oldB119ConnectionOracle, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, victim)
	if c.Controller != me.ID || c.Tapped {
		t.Errorf("controller=%v tapped=%v, want mine and untapped", c.Controller == me.ID, c.Tapped)
	}
	if !effectstest.HasEffectiveKeyword(t, g, victim, "haste") {
		t.Error("no haste")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if c := effectstest.GaLayered(t, g, victim); c.Controller != opp.ID {
		t.Error("control did not return at end of turn")
	}
}

func oldB119MenaceGoblin(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Menacing Goblin", TypeLine: "Creature — Goblin", Power: 2, Toughness: 2,
		Keywords: []string{"menace"}, Owner: owner, Controller: owner,
	})
}

func TestOldB119TentativeConnectionCostsThreeLessWithAMenaceCreature(t *testing.T) {
	cast := func(menace bool, float string) error {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if menace {
			oldB119MenaceGoblin(g, me.ID)
		} else {
			effectstest.B12Creature(g, me.ID, "Plain Goblin", "Creature — Goblin", 2, 2)
		}
		victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.ToMainForCost(t, g)
		id := effectstest.HandCardFull(me, "Tentative Connection", "Sorcery", "{3}{R}", oldB119ConnectionOracle, []string{"R"})
		effectstest.FloatMana(t, g, me, float)
		return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(victim), Strict: true})
	}
	if err := cast(true, "{R}"); err != nil {
		t.Errorf("with a menace creature {R} should pay for it: %v", err)
	}
	if err := cast(false, "{R}"); err == nil {
		t.Error("without a menace creature {R} alone paid for a {3}{R} spell")
	}
	if err := cast(false, "{R}{C}{C}{C}"); err != nil {
		t.Errorf("full price should work without a menace creature: %v", err)
	}
}

func TestOldB119LeadTheStampedeTakesAnyNumberOfCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := oldB119Library(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf"},
		game.Card{Name: "Elk", TypeLine: "Creature — Elk"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Sixth", TypeLine: "Creature — Bear"}, // not looked at
	)
	effectstest.CastCatalogSpell(t, g, "Lead the Stampede", "Sorcery", oldB119StampedeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0], ids[2])
	oldB119BottomInOrder(t, g, me, ids[4], ids[1], ids[3])
	effectstest.PassPriorityAroundTable(t, g)
	for _, i := range []int{0, 2} {
		if !oldB119In(me.Hand, ids[i]) {
			t.Errorf("creature %d is not in hand", i)
		}
	}
	for _, i := range []int{1, 3, 4} {
		if !oldB119In(me.Library, ids[i]) || me.Library.Cards[len(me.Library.Cards)-1].InstanceID == ids[i] {
			t.Errorf("rest card %d is not in the library below the top", i)
		}
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != ids[5] {
		t.Error("the sixth card is not on top any more")
	}
}

func TestOldB119LeadTheStampedeRefusesANonCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := oldB119Library(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
	)
	effectstest.CastCatalogSpell(t, g, "Lead the Stampede", "Sorcery", oldB119StampedeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if ch == nil {
		t.Fatal("no choose_cards prompt")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{ids[1]}); err == nil {
		t.Error("the noncreature card was accepted")
	}
}

func TestOldB119NightsquadCommandoMakesATokenOnlyAfterYouAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AvB16CastCreature(t, g, "Nightsquad Commando", "Creature — Human Soldier", oldB119CommandoOracle, 2, 3)
	if n := len(effectstest.S46Tokens(g, me.ID, "Human Soldier")); n != 0 {
		t.Fatalf("%d tokens before any attack", n)
	}
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.AvB16CastCreature(t, g, "Nightsquad Commando", "Creature — Human Soldier", oldB119CommandoOracle, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	toks := effectstest.S46Tokens(g, me.ID, "Human Soldier")
	if len(toks) != 1 {
		t.Fatalf("%d Human Soldier tokens, want 1", len(toks))
	}
	if p, tg := effectstest.PtOf(t, g, toks[0].InstanceID); p != 1 || tg != 1 {
		t.Errorf("the token is %d/%d, want 1/1", p, tg)
	}
}

func TestOldB119ValiantRescuerTriggersOnlyOnTheFirstOtherCycleEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Valiant Rescuer", "Creature — Human Soldier", oldB119RescuerOracle, 3, 1)
	cycle := func() {
		effectstest.CycleFromHand(t, g, "Scrounging Skyray", "Creature — Bird", effectstest.ScroungingSkyrayOracle, "{C}{C}")
		effectstest.PassPriorityAroundTable(t, g)
	}
	tokens := func() int { return len(effectstest.S46Tokens(g, me.ID, "Human Soldier")) }
	cycle()
	if tokens() != 1 {
		t.Fatalf("%d tokens after the first cycle, want 1", tokens())
	}
	cycle()
	if tokens() != 1 {
		t.Errorf("%d tokens after the second cycle, want still 1", tokens())
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	for g.Turn.ActiveSeat != 0 {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	cycle()
	if tokens() != 2 {
		t.Errorf("%d tokens on the next turn, want 2", tokens())
	}
}

func TestOldB119ValiantRescuerCyclesItselfWithoutATokenAndCountsForNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	resc, _ := effectstest.CycleFromHand(t, g, "Valiant Rescuer", "Creature — Human Soldier", oldB119RescuerOracle, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.S46Tokens(g, me.ID, "Human Soldier")); n != 0 {
		t.Errorf("%d tokens from cycling the Rescuer itself", n)
	}
	if !oldB119In(me.Graveyard, resc) {
		t.Error("the Rescuer is not in the graveyard")
	}
}

// oldB119SnapdaxTable seeds the boards of Mythos of Snapdax: each player has
// two creatures, an artifact and a land, and I also have an enchantment.
type oldB119SnapdaxBoards struct {
	g                     *game.Game
	me, opp               *game.Player
	a, b, art, ench, land uuid.UUID
	oa, ob, oart, oland   uuid.UUID
}

func oldB119SnapdaxTable(t *testing.T) oldB119SnapdaxBoards {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	return oldB119SnapdaxBoards{
		g: g, me: me, opp: opp,
		a:     effectstest.B12Creature(g, me.ID, "My Bear A", "Creature — Bear", 2, 2),
		b:     effectstest.B12Creature(g, me.ID, "My Bear B", "Creature — Bear", 2, 2),
		art:   effectstest.B12Creature(g, me.ID, "My Rock", "Artifact", 0, 0),
		ench:  effectstest.B12Creature(g, me.ID, "My Aura", "Enchantment", 0, 0),
		land:  effectstest.B12Creature(g, me.ID, "My Forest", "Basic Land — Forest", 0, 0),
		oa:    effectstest.B12Creature(g, opp.ID, "Their Bear A", "Creature — Bear", 2, 2),
		ob:    effectstest.B12Creature(g, opp.ID, "Their Bear B", "Creature — Bear", 2, 2),
		oart:  effectstest.B12Creature(g, opp.ID, "Their Rock", "Artifact", 0, 0),
		oland: effectstest.B12Creature(g, opp.ID, "Their Forest", "Basic Land — Forest", 0, 0),
	}
}

func oldB119CastSnapdax(t *testing.T, bt oldB119SnapdaxBoards, mana string) {
	t.Helper()
	effectstest.ToMainForCost(t, bt.g)
	id := effectstest.HandCardFull(bt.me, "Mythos of Snapdax", "Sorcery", "{2}{W}{W}", oldB119SnapdaxOracle, []string{"W"})
	effectstest.FloatMana(t, bt.g, bt.me, mana)
	if err := bt.g.CastSpell(bt.me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, bt.g)
}

func TestOldB119MythosOfSnapdaxEachPlayerKeepsOneOfEachType(t *testing.T) {
	bt := oldB119SnapdaxTable(t)
	oldB119CastSnapdax(t, bt, "{W}{W}{C}{C}")
	effectstest.AnswerChooseCards(t, bt.g, bt.me.ID, bt.a, bt.art, bt.ench)
	effectstest.AnswerChooseCards(t, bt.g, bt.opp.ID, bt.ob, bt.oart)
	effectstest.PassPriorityAroundTable(t, bt.g)
	for name, id := range map[string]uuid.UUID{"my A": bt.a, "my artifact": bt.art, "my enchantment": bt.ench, "my land": bt.land, "their B": bt.ob, "their artifact": bt.oart, "their land": bt.oland} {
		if !bt.g.Battlefield.Contains(id) {
			t.Errorf("%s was sacrificed", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"my B": bt.b, "their A": bt.oa} {
		if bt.g.Battlefield.Contains(id) {
			t.Errorf("%s survived", name)
		}
	}
}

func TestOldB119MythosOfSnapdaxLetsYouChooseForEveryPlayerWhenBlackAndRedWereSpent(t *testing.T) {
	bt := oldB119SnapdaxTable(t)
	oldB119CastSnapdax(t, bt, "{W}{W}{B}{R}")
	effectstest.AnswerChooseCards(t, bt.g, bt.me.ID, bt.a, bt.art, bt.ench)
	if effectstest.LatestChoiceOfKindFor(bt.g, game.PendingChoiceChooseCards, bt.opp.ID) != nil {
		t.Fatal("the opponent was asked to choose though I choose for them")
	}
	effectstest.AnswerChooseCards(t, bt.g, bt.me.ID, bt.oa, bt.oart)
	effectstest.PassPriorityAroundTable(t, bt.g)
	if !bt.g.Battlefield.Contains(bt.oa) || bt.g.Battlefield.Contains(bt.ob) || !bt.g.Battlefield.Contains(bt.oart) || !bt.g.Battlefield.Contains(bt.oland) {
		t.Error("the opponent's board is not what I kept for them")
	}
	if !bt.g.Battlefield.Contains(bt.a) || bt.g.Battlefield.Contains(bt.b) {
		t.Error("my board is not what I kept")
	}
}

func TestOldB119MythosOfSnapdaxNeedsBothBlackAndRedForTheChoice(t *testing.T) {
	bt := oldB119SnapdaxTable(t)
	oldB119CastSnapdax(t, bt, "{W}{W}{B}{C}")
	effectstest.AnswerChooseCards(t, bt.g, bt.me.ID, bt.a, bt.art, bt.ench)
	if effectstest.LatestChoiceOfKindFor(bt.g, game.PendingChoiceChooseCards, bt.opp.ID) == nil {
		t.Error("with only black spent the opponent should choose for themselves")
	}
}

func TestOldB119MythosOfIllunaCopiesAPermanentAndFightsWithRedGreen(t *testing.T) {
	cast := func(t *testing.T, mana string, victimSetup bool) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
		var victim uuid.UUID
		if victimSetup {
			victim = effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
		}
		effectstest.ToMainForCost(t, g)
		id := effectstest.HandCardFull(me, "Mythos of Illuna", "Sorcery", "{2}{U}{U}", oldB119IllunaOracle, []string{"U"})
		effectstest.FloatMana(t, g, me, mana)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(src), Strict: true}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		return g, me, src, victim
	}
	t.Run("plain copy", func(t *testing.T) {
		g, me, _, victim := cast(t, "{U}{U}{C}{C}", true)
		if n := len(effectstest.S46Tokens(g, me.ID, "Hill Giant")); n != 1 {
			t.Fatalf("%d Hill Giant tokens, want 1", n)
		}
		if effectstest.LatestPickTarget(g, me.ID) != nil || !g.Battlefield.Contains(victim) {
			t.Error("the plain copy fought")
		}
	})
	t.Run("red green copy fights", func(t *testing.T) {
		g, me, _, victim := cast(t, "{U}{U}{R}{G}", true)
		effectstest.AnswerPickTarget(t, g, victim)
		effectstest.PassPriorityAroundTable(t, g)
		toks := effectstest.S46Tokens(g, me.ID, "Hill Giant")
		if len(toks) != 1 {
			t.Fatalf("%d Hill Giant tokens, want 1", len(toks))
		}
		if g.Battlefield.Contains(victim) {
			t.Error("the 3/3 copy did not kill the 2/2 it fought")
		}
	})
	t.Run("red green copy may fight nothing", func(t *testing.T) {
		g, me, _, _ := cast(t, "{U}{U}{R}{G}", false)
		effectstest.B26DeclinePickTarget(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if n := len(effectstest.S46Tokens(g, me.ID, "Hill Giant")); n != 1 {
			t.Errorf("%d Hill Giant tokens, want 1", n)
		}
	})
}

func TestOldB119MythosOfIllunaCopyOfANoncreatureDoesNotFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Creature(g, opp.ID, "Mind Stone", "Artifact", 0, 0)
	effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, "Mythos of Illuna", "Sorcery", "{2}{U}{U}", oldB119IllunaOracle, []string{"U"})
	effectstest.FloatMana(t, g, me, "{U}{U}{R}{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(rock), Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.S46Tokens(g, me.ID, "Mind Stone")); n != 1 {
		t.Fatalf("%d Mind Stone tokens, want 1", n)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a noncreature copy asked for a fight target")
	}
}
