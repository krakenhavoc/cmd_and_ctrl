package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

const synthesisPodOracle = "55e42539-3fe2-4c84-9978-c071d6fc6aa0"

// synthesisPodTable is a main-phase table with the Pod untapped under the
// active player.
func synthesisPodTable(t *testing.T) (g *game.Game, me, opp *game.Player, pod uuid.UUID) {
	t.Helper()
	g, me, opp = exileCostTable(t)
	pod = pushCatalogPermanent(g, me.ID, "Synthesis Pod", "Artifact", synthesisPodOracle, false)
	return g, me, opp, pod
}

// synthesisPodCast puts a spell with the given mana cost on the stack
// under `seat`'s control, X announced as x.
func synthesisPodCast(t *testing.T, g *game.Game, seat *game.Player, name, manaCost string, x int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		seat.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: manaCost,
			Owner: seat.ID, Controller: seat.ID,
		})
	})
	if err := g.CastSpell(seat.ID, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// synthesisPodLibrary replaces p's library with one card per mana cost,
// the first on top, and returns their IDs in that order.
func synthesisPodLibrary(g *game.Game, p *game.Player, manaCosts ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(manaCosts))
	g.WithWriteLock(func() {
		p.Library.Cards = nil
		for i := len(manaCosts) - 1; i >= 0; i-- {
			ids[i] = uuid.New()
			typeLine := "Sorcery"
			if manaCosts[i] == "" {
				typeLine = "Land"
			}
			p.Library.PushTop(game.Card{
				InstanceID: ids[i], Name: "Library " + manaCosts[i], TypeLine: typeLine, ManaCost: manaCosts[i],
				Owner: p.ID, Controller: p.ID,
			})
		}
	})
	return ids
}

// synthesisPodActivate floats the ability's mana and activates it, strict,
// naming `exile` to the cost and the opponent as the target.
func synthesisPodActivate(t *testing.T, g *game.Game, me, opp *game.Player, pod uuid.UUID, exile ...uuid.UUID) error {
	t.Helper()
	floatMana(t, g, me, "{U}{U}")
	return g.ActivateCatalogAbility(me.ID, pod, 0, game.ActivateAbilityParams{
		ExileIDs: exile,
		Targets:  []game.TargetRef{playerTarget(opp.ID)},
		Strict:   true,
	})
}

// synthesisPodResolve passes priority until the Pod's ability has either
// offered its free cast or left the stack.
func synthesisPodResolve(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for i := 0; i < 16 && resolutionCastOfferFor(g, me.ID) == nil && !stackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
}

// The cost takes a spell the activator controls and nothing else: not an
// opponent's spell, not a card in a hand, not nothing, and not two.
// The spell is exiled without resolving; the reveal stops on the first
// card whose mana value is the spell's plus one, that card is exiled, the
// rest stay in the library, and the controller casts the card for free.
func TestSynthesisPodExilesASpellYouControlAndCastsTheCardItFinds(t *testing.T) {
	g, me, opp, pod := synthesisPodTable(t)
	spell := synthesisPodCast(t, g, me, "Two Drop", "{1}{R}", 0)
	other := synthesisPodCast(t, g, me, "Other Spell", "{R}", 0)
	var inHand uuid.UUID
	g.WithWriteLock(func() { inHand = pushHandCardWithManaCost(me, "Held", "Instant", "{1}{R}") })
	lib := synthesisPodLibrary(g, opp, "{1}", "", "{3}{G}", "{2}{U}", "{1}{B}{B}")

	var options []uuid.UUID
	g.WithWriteLock(func() {
		options = g.ExileCostOptionsForEffect(me.ID, pod, ExileASpellYouControl().ExileCards)
	})
	if len(options) != 2 {
		t.Fatalf("the walk offers %v, want the activator's two spells", options)
	}
	for name, ids := range map[string][]uuid.UUID{
		"a card in hand":       {inHand},
		"no spell":             nil,
		"two spells":           {spell, other},
		"the same spell twice": {spell, spell},
		"the Pod itself":       {pod},
	} {
		if err := synthesisPodActivate(t, g, me, opp, pod, ids...); err == nil {
			t.Fatalf("%s paid \"exile a spell you control\"", name)
		}
	}
	if err := synthesisPodActivate(t, g, me, opp, pod, spell); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Stack.Contains(spell) || g.StackMeta[spell] != nil || !g.Exile.Contains(spell) {
		t.Fatal("the spell paid to the cost is not exiled off the stack")
	}
	if !g.Stack.Contains(other) {
		t.Fatal("the spell that was not named left the stack")
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventCounterSpell {
			t.Fatalf("the exiled spell was countered (%+v)", ev)
		}
	}

	synthesisPodResolve(t, g, me)
	hit := lib[3] // {2}{U}: mana value 3 = 1 + the spell's 2
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordExileUntil, hit)
	if !g.Exile.Contains(hit) {
		t.Fatal("the revealed card was not exiled")
	}
	if opp.Library.Size() != 4 {
		t.Fatalf("the opponent's library holds %d cards, want the 4 that were not exiled", opp.Library.Size())
	}
	for _, id := range []uuid.UUID{lib[0], lib[1], lib[2], lib[4]} {
		if !opp.Library.Contains(id) {
			t.Fatal("a card revealed on the way, or below the hit, left the library")
		}
	}
	if err := castThroughResolutionCast(t, g, me.ID, hit, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast of the exiled card: %v", err)
	}
	if item := g.StackMeta[hit]; item == nil || item.Controller != me.ID {
		t.Fatal("the exiled card is not a spell the Pod's controller controls")
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(hit) {
		t.Error("the resolved spell did not go to its owner's graveyard")
	}
	if !g.Exile.Contains(spell) {
		t.Error("the spell paid to the cost left exile")
	}
}

// An opponent's spell is not "a spell you control", and an ability on the
// stack is not a spell.
func TestSynthesisPodCantExileAnOpponentsSpellOrAnAbility(t *testing.T) {
	g, me, opp, pod := synthesisPodTable(t)
	theirs := synthesisPodCast(t, g, me, "Their Spell", "{1}{R}", 0)
	g.WithWriteLock(func() {
		// The opponent gains control of the spell (CR 112.2: the
		// controller is the stack item's).
		g.StackMeta[theirs].Controller = opp.ID
	})
	synthesisPodLibrary(g, opp, "{2}{U}")
	if err := synthesisPodActivate(t, g, me, opp, pod, theirs); err == nil {
		t.Fatal("an opponent's spell paid \"exile a spell you control\"")
	}
	var options []uuid.UUID
	g.WithWriteLock(func() {
		options = g.ExileCostOptionsForEffect(me.ID, pod, ExileASpellYouControl().ExileCards)
	})
	if len(options) != 0 {
		t.Fatalf("the walk offers %v with no spell of the activator's on the stack", options)
	}
	if !g.Stack.Contains(theirs) {
		t.Fatal("a refused activation moved the spell")
	}

	// With the Pod's own ability on the stack, the ability is not a pick.
	mine := synthesisPodCast(t, g, me, "My Spell", "{R}", 0)
	if err := synthesisPodActivate(t, g, me, opp, pod, mine); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for id, item := range g.StackMeta {
		if item == nil || item.Kind == game.StackItemSpell {
			continue
		}
		g.WithWriteLock(func() {
			options = g.ExileCostOptionsForEffect(me.ID, uuid.New(), ExileASpellYouControl().ExileCards)
		})
		for _, o := range options {
			if o == id {
				t.Fatal("the walk offers an ability on the stack")
			}
		}
	}
}

// "The exiled spell's mana value" is the spell's as it last existed on
// the stack, X counted (CR 202.3e, CR 608.2h), and the payment survives
// a restore before the ability resolves.
func TestSynthesisPodReadsTheExiledSpellsXAcrossARestore(t *testing.T) {
	g, me, opp, pod := synthesisPodTable(t)
	spell := synthesisPodCast(t, g, me, "X Spell", "{X}{R}", 3)
	// {1}{U} would be the hit if X read as 0 in exile; {4}{G} is 1 + 4.
	lib := synthesisPodLibrary(g, opp, "{1}{U}", "{4}{G}", "{4}{B}")
	if err := synthesisPodActivate(t, g, me, opp, pod, spell); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if c, ok := g.LookupCardForEffect(spell); !ok || c.ManaValue() != 1 {
		t.Fatalf("the exiled card should read mana value 1 in exile (found %v)", ok)
	}

	restored := restoreRoundTrip(t, g, true)
	rme := restored.PlayerByIDForEffect(me.ID)
	ropp := restored.PlayerByIDForEffect(opp.ID)
	synthesisPodResolve(t, restored, rme)
	mustResolutionCastOffer(t, restored, me.ID, game.MayCastKeywordExileUntil, lib[1])
	if !restored.Exile.Contains(lib[1]) || !ropp.Library.Contains(lib[0]) || !ropp.Library.Contains(lib[2]) {
		t.Fatal("the restored ability did not exile the card with mana value 5 alone")
	}

	// Declining leaves the card in exile.
	declineResolutionCast(t, restored, me.ID)
	passPriorityAroundTable(t, restored)
	if !restored.Exile.Contains(lib[1]) {
		t.Error("a declined card left exile")
	}
}

// A library with no card of the right mana value is revealed whole and
// shuffled; nothing is exiled and nothing is offered.
func TestSynthesisPodWithNoMatchingCardOnlyShuffles(t *testing.T) {
	g, me, opp, pod := synthesisPodTable(t)
	spell := synthesisPodCast(t, g, me, "Two Drop", "{1}{R}", 0)
	lib := synthesisPodLibrary(g, opp, "{1}", "", "{3}{G}", "{4}")
	if err := synthesisPodActivate(t, g, me, opp, pod, spell); err != nil {
		t.Fatalf("activate: %v", err)
	}
	exiled := g.Exile.Size()
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a free cast was offered with no card found")
	}
	if g.Exile.Size() != exiled || opp.Library.Size() != len(lib) {
		t.Fatalf("exile %d (want %d), library %d (want %d)", g.Exile.Size(), exiled, opp.Library.Size(), len(lib))
	}
	shuffled, revealed := false, false
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Target == opp.ID && ev.Label == "shuffle" {
			shuffled = true
		}
		if ev.Kind == game.EventRevealCards {
			revealed = true
		}
	}
	if !shuffled || !revealed {
		t.Errorf("revealed %v, shuffled %v: the whole library is revealed, then shuffled", revealed, shuffled)
	}
}

// A copy of a spell is a spell you control: it pays the cost, ceases to
// exist, and its mana value is the copied spell's.
func TestSynthesisPodExilesACopyOfASpell(t *testing.T) {
	g, me, opp, pod := synthesisPodTable(t)
	spell := synthesisPodCast(t, g, me, "Two Drop", "{1}{R}", 0)
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(spell, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
	var copyID uuid.UUID
	for id, item := range g.StackMeta {
		if id != spell && item != nil && item.Kind == game.StackItemSpell {
			copyID = id
		}
	}
	if copyID == uuid.Nil {
		t.Fatal("no copy on the stack")
	}
	lib := synthesisPodLibrary(g, opp, "{1}", "{2}{U}")
	if err := synthesisPodActivate(t, g, me, opp, pod, copyID); err != nil {
		t.Fatalf("activate with the copy: %v", err)
	}
	if g.Stack.Contains(copyID) || g.Exile.Contains(copyID) {
		t.Fatal("the copy did not cease to exist")
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("the original spell left the stack")
	}
	synthesisPodResolve(t, g, me)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordExileUntil, lib[1])
}

// Register refuses the stack form on a mana ability and with an open
// count.
func TestRegisterRefusesAMisplacedExileASpellCost(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	mustPanic(t, "exiles a spell from the stack", func() {
		Register(Spec{
			OracleID: "exile-spell-cost-test-mana",
			Name:     "Spell Exile Mana",
			ManaAbilities: []ManaAbility{{
				Cost: ManaAbilityCost{Tap: true, ExileCards: ExileASpellYouControl().ExileCards}, Produced: "{C}", Label: "Add {C}",
			}},
		})
	})
	mustPanic(t, "exiles spells from the stack with an open count or a link", func() {
		Register(Spec{
			OracleID: "exile-spell-cost-test-link",
			Name:     "Spell Exile Link",
			Activated: []ActivatedAbility{{
				Label:  "Exile a spell you control: do nothing",
				Cost:   game.AbilityCost{ExileCards: &game.ExileCost{N: 1, Label: "a spell you control", From: game.ZoneStack, LinkToSource: true}},
				Effect: noop,
			}},
		})
	})
}

// The view ships the stack as the cost's zone with the viewer's own
// spell as the one option, and the enumerator offers the activation
// naming that spell.
func TestSynthesisPodViewAndEnumeratorNameTheSpell(t *testing.T) {
	g, me, opp, pod := synthesisPodTable(t)
	synthesisPodLibrary(g, opp, "{2}{U}")
	floatMana(t, g, me, "{U}{U}")

	row := func() protocol.ActivatedAbilityView {
		activatedRowOfMarkKnown(g)
		v := protocol.ViewOfGameFor(g, me.ID.String())
		for _, cv := range v.Battlefield.Cards {
			if cv.InstanceID == pod.String() && len(cv.ActivatedAbilities) > 0 {
				return cv.ActivatedAbilities[0]
			}
		}
		t.Fatal("no activated row on the Pod")
		return protocol.ActivatedAbilityView{}
	}
	if r := row(); r.ExileCostN != 1 || r.ExileCostZone != "stack" || len(r.ExileCostOptions) != 0 {
		t.Errorf("row %+v, want one spell from the stack and no option yet", r)
	}
	move := func() []string {
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type != legal.TypeActivateAbility || m.Source != pod {
				continue
			}
			var p struct {
				ExileIDs []string `json:"exile_ids"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatal(err)
			}
			return p.ExileIDs
		}
		return nil
	}
	if ids := move(); ids != nil {
		t.Errorf("the activation is offered with no spell on the stack (%v)", ids)
	}

	spell := synthesisPodCast(t, g, me, "Two Drop", "{1}{R}", 0)
	floatMana(t, g, me, "{U}{U}")
	if r := row(); len(r.ExileCostOptions) != 1 || r.ExileCostOptions[0] != spell.String() {
		t.Errorf("row options %v, want the viewer's spell %s", r.ExileCostOptions, spell)
	}
	if ids := move(); len(ids) != 1 || ids[0] != spell.String() {
		t.Errorf("the enumerator pays with %v, want the spell %s", ids, spell)
	}
}
