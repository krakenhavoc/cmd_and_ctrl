package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cipher_test.go — CR 702.99 (#3867): a cipher spell may be exiled
// encoded on a creature its controller controls as it resolves; the
// creature then has "whenever this creature deals combat damage to a
// player, you may copy the encoded card and cast the copy without
// paying its mana cost"; the encoding ends when the creature leaves the
// battlefield, and the card stays in exile.

const (
	lastThoughtsOracle   = "649d983a-e861-4893-8ae8-fb0a22cfad0a"
	traitDoctoringOracle = "3b5b926b-ee8f-4c7a-8737-1de7d18ca60d"
)

// cipherResolveToPrompt passes priority until `chooser` has an option
// pick open, or the stack is empty.
func cipherResolveToPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 16; i++ {
		if c := latestOptionPickFor(g, chooser); c != nil {
			return c
		}
		if stackFullyEmpty(g) {
			return nil
		}
		if err := g.PassPriority(); err != nil {
			return latestOptionPickFor(g, chooser)
		}
	}
	return latestOptionPickFor(g, chooser)
}

// encodedOn is the creature `card` is encoded on, if any.
func encodedOn(g *game.Game, card uuid.UUID) (uuid.UUID, bool) {
	var ref game.PermissionCardRef
	var ok bool
	g.WithWriteLock(func() { ref, ok = g.EncodedOnForEffect(card) })
	return ref.ID, ok
}

// castAndEncode casts Last Thoughts and encodes it on `creature`.
func castAndEncode(t *testing.T, g *game.Game, me *game.Player, creature uuid.UUID) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Last Thoughts", "Sorcery", lastThoughtsOracle, nil)
	pick := cipherResolveToPrompt(t, g, me.ID)
	if pick == nil {
		t.Fatal("cipher asks whether to encode the card")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if len(o.Cards) == 1 && o.Cards[0] == creature {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("the creature is not offered: %+v", pick.PickOptions)
	}
	answerOptionPick(t, g, me.ID, idx)
	g.SettleResolution()
	return id
}

// Encoded: the card is exiled, not put into the graveyard, and it is
// encoded on the chosen creature. The spell's own text resolved first.
func TestCipherExilesTheCardEncodedOnTheChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := seedBear(g, me.ID)
	other := seedBear(g, me.ID)
	hand := me.Hand.Size()
	id := castAndEncode(t, g, me, other)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d: Last Thoughts draws one before it is encoded", hand, me.Hand.Size())
	}
	if !inExile(g, id) || me.Graveyard.Contains(id) {
		t.Fatal("an encoded card is exiled, not put into its owner's graveyard")
	}
	if on, ok := encodedOn(g, id); !ok || on != other {
		t.Errorf("encoded on %v (%v), want the creature chosen (%v), not %v", on, ok, other, bear)
	}
	// The creature has the ability, granted by the encoded card, and the
	// other creature does not: the view lists it from here.
	granted := game.GrantedAbilitiesOf(gaLayered(t, g, other))
	if len(granted) != 1 || granted[0].Source != id || !strings.Contains(granted[0].Text, "combat damage to a player") {
		t.Errorf("the encoded creature's granted abilities are %+v, want cipher's from the card", granted)
	}
	if n := len(game.GrantedAbilitiesOf(gaLayered(t, g, bear))); n != 0 {
		t.Errorf("the other creature has %d granted abilities, want 0", n)
	}
}

// Declined: the card goes to the graveyard as any sorcery does.
func TestCipherDeclinedPutsTheCardIntoTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedBear(g, me.ID)
	id := castCatalogSpell(t, g, "Last Thoughts", "Sorcery", lastThoughtsOracle, nil)
	pick := cipherResolveToPrompt(t, g, me.ID)
	if pick == nil {
		t.Fatal("cipher asks whether to encode the card")
	}
	answerOptionPick(t, g, me.ID, len(pick.PickOptions)-1)
	g.SettleResolution()
	if !me.Graveyard.Contains(id) || inExile(g, id) {
		t.Error("a card not encoded goes to its owner's graveyard")
	}
}

// With no creature to encode it on, nothing is asked.
func TestCipherWithNoCreatureAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	seedBear(g, opp.ID) // an opponent's creature is not "a creature you control"
	id := castCatalogSpell(t, g, "Last Thoughts", "Sorcery", lastThoughtsOracle, nil)
	if pick := cipherResolveToPrompt(t, g, me.ID); pick != nil {
		t.Fatalf("asked %q with no creature to encode it on", pick.Reason)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the card goes to the graveyard")
	}
}

// The encoded creature deals combat damage to a player: its controller
// may cast a copy of the card free. The copy draws, ceases to exist as
// it resolves (it is not a card, so it is never asked to be encoded),
// and the original stays exiled and encoded, ready for the next hit.
func TestCipherCombatDamageOffersAFreeCopy(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := seedBear(g, me.ID)
	id := castAndEncode(t, g, me, bear)

	// Noncombat damage does not trigger it.
	b27Damage(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if n := len(cardCopies(g)); n != 0 {
		t.Fatalf("%d copies after noncombat damage, want 0", n)
	}

	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	cp := copyOf(t, g, "Last Thoughts")
	if cp.InstanceID == id {
		t.Fatal("the offer is a copy, not the encoded card")
	}
	hand := me.Hand.Size()
	if err := castCopy(g, me.ID, cp.InstanceID, true); err != nil {
		t.Fatalf("the copy is cast without paying its mana cost: %v", err)
	}
	if pick := cipherResolveToPrompt(t, g, me.ID); pick != nil {
		t.Fatalf("a copy is not represented by a card, yet cipher asked %q", pick.Reason)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d: the copy draws one", hand, me.Hand.Size())
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left, want 0", n)
	}
	if on, ok := encodedOn(g, id); !ok || on != bear {
		t.Error("the original stays exiled and encoded on the creature")
	}

	// And again on the next hit.
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	copyOf(t, g, "Last Thoughts")
}

// CR 702.99c: the creature leaves the battlefield — the card stays
// exiled but is no longer encoded, and the creature back on the
// battlefield is a new object with nothing encoded on it.
func TestCipherCreatureLeavingEndsTheEncoding(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := seedBear(g, me.ID)
	id := castAndEncode(t, g, me, bear)
	var err error
	g.WithWriteLock(func() { err = g.BounceToHandForEffect(bear) })
	if err != nil {
		t.Fatalf("bounce: %v", err)
	}
	if !inExile(g, id) {
		t.Fatal("the card stays in exile")
	}
	if _, ok := encodedOn(g, id); ok {
		t.Error("the card is no longer encoded once the creature has left")
	}
	// A creature that was never chosen, and the bear's new object if it
	// returns, deal combat damage: nothing is offered.
	other := seedBear(g, me.ID)
	dealCombatDamageToPlayer(g, other, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies offered after the encoding ended, want 0", n)
	}
}

// CR 702.99c: a change of controller does not end the encoding, and the
// ability is the creature's, so its new controller casts the copy.
func TestCipherFollowsTheCreaturesController(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	third := g.Seats[(g.Turn.ActiveSeat+2)%4]
	bear := seedBear(g, me.ID)
	id := castAndEncode(t, g, me, bear)
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.Nil, bear, opp.ID, game.IndefiniteDuration(), "test — gain control")
		g.RecomputeLayersIfStaleLocked()
	})
	if on, ok := encodedOn(g, id); !ok || on != bear {
		t.Fatal("the card stays encoded on the creature under its new controller")
	}
	dealCombatDamageToPlayer(g, bear, third.ID, 2)
	passPriorityAroundTable(t, g)
	cp := copyOf(t, g, "Last Thoughts")
	if cp.Owner != opp.ID {
		t.Errorf("the copy is offered to %v, want the creature's controller %v", cp.Owner, opp.ID)
	}
}

// The encoding and the open question both survive a restore.
func TestCipherEncodingSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := seedBear(g, me.ID)
	castCatalogSpell(t, g, "Last Thoughts", "Sorcery", lastThoughtsOracle, nil)
	if cipherResolveToPrompt(t, g, me.ID) == nil {
		t.Fatal("cipher asks")
	}
	// The question is keyed, so the table waiting on it is a restore point.
	restored := restoreRoundTrip(t, g, true)
	rme := restored.Seats[restored.Turn.ActiveSeat]
	answerOptionPick(t, restored, rme.ID, 0)
	restored.SettleResolution()
	var id uuid.UUID
	for _, c := range restored.Exile.Cards {
		if c.Name == "Last Thoughts" {
			id = c.InstanceID
		}
	}
	if id == uuid.Nil {
		t.Fatal("the restored table encodes the card")
	}
	again := restoreRoundTrip(t, restored, true)
	if on, ok := encodedOn(again, id); !ok || on != bear {
		t.Error("the encoding survives the snapshot")
	}
	dealCombatDamageToPlayer(again, bear, opp.ID, 2)
	passPriorityAroundTable(t, again)
	copyOf(t, again, "Last Thoughts")
}

// Trait Doctoring: the text change asks both of its words first; the
// cipher question is the spell's last instruction.
func TestTraitDoctoringChangesTextThenAsksCipher(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := seedBear(g, me.ID)
	theirs := twPut(g, opp.ID, "Green Bear", "Creature — Bear", []string{"G"})
	id := castCatalogSpell(t, g, "Trait Doctoring", "Sorcery", traitDoctoringOracle, twTargetRef(theirs))
	first := cipherResolveToPrompt(t, g, me.ID)
	if first == nil || strings.Contains(first.Reason, "cipher") {
		t.Fatalf("the text change is asked first; got %+v", first)
	}
	answerOptionPick(t, g, me.ID, twGreen)
	if p := latestOptionPickFor(g, me.ID); p == nil || strings.Contains(p.Reason, "cipher") {
		t.Fatalf("the replacing word is asked second; got %+v", p)
	}
	answerOptionPick(t, g, me.ID, twBlue)
	pick := latestOptionPickFor(g, me.ID)
	if pick == nil || !strings.Contains(pick.Reason, "cipher") {
		t.Fatalf("cipher is asked last; got %+v", pick)
	}
	answerOptionPick(t, g, me.ID, 0)
	g.SettleResolution()
	if on, ok := encodedOn(g, id); !ok || on != mine {
		t.Error("Trait Doctoring is encoded on the caster's creature")
	}
	if changes := twTextChanges(t, g, theirs); len(changes) == 0 {
		t.Error("the bear's colour word was changed")
	}
}

// The creature's ability is a layer-6 grant at the moment of encoding
// (CR 613.7): a "loses all abilities" newer than that takes it away,
// and one older than it does not (CR 613.6).
func TestCipherGrantAndAbilityRemovalFollowTimestamps(t *testing.T) {
	for _, removalFirst := range []bool{true, false} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		bear := seedBear(g, me.ID)
		lose := func() {
			g.WithWriteLock(func() {
				g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(bear),
					[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities")
				g.RecomputeLayersIfStaleLocked()
			})
		}
		if removalFirst {
			lose()
		}
		castAndEncode(t, g, me, bear)
		if !removalFirst {
			lose()
		}
		dealCombatDamageToPlayer(g, bear, opp.ID, 2)
		passPriorityAroundTable(t, g)
		offered := len(cardCopies(g)) > 0
		if offered != removalFirst {
			t.Errorf("removal first = %v: copy offered = %v, want %v", removalFirst, offered, removalFirst)
		}
	}
}
