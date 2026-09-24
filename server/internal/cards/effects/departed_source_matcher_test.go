package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// departed_source_matcher_test.go is the card-level proof for #1430:
// four catalog damage replacements that test only the source's
// CONTROLLER or its TYPE — not colour, which #1417 already fixed —
// still read `g.LookupCardForEffect(ev.DamageSource)` at apply time
// instead of `effects.damageSourceCharacteristics`, the event's
// last-known snapshot (CR 608.2h). A source that changed controller
// or type on its way out of the battlefield (a stolen creature's dies
// damage, an animated permanent) was judged by its NEW zone's card,
// not by what it was when it dealt the damage.
//
// Each test drives the registered Spec's AppliesTo closure directly
// (the pattern TestOjerAxonilLeavesCombatDamageAlone already uses in
// vivi_and_ojer_test.go), because the fix is entirely inside that
// closure and a manually-built ReplacementEvent can say precisely
// what the "current zone" card and the "last known" snapshot each
// claim — the same two answers a real stolen-and-killed creature
// would give, without the ceremony of staging the theft for real.
// Every scenario builds a real, contradicting current-zone card at
// DamageSource: the pre-fix code would read IT and get the wrong
// answer, which is exactly what the back-out table below checks.

// b1430DamageEvent builds a RepEventDamage event naming `dealer` as
// the source, `lki` as its last-known characteristics (the identity
// it had when it dealt the damage), and `target` as the recipient.
func b1430DamageEvent(dealer, target uuid.UUID, amount int, combat bool, lki *game.Characteristic) *game.ReplacementEvent {
	return &game.ReplacementEvent{
		Kind:           game.RepEventDamage,
		DamageSource:   dealer,
		DamageTarget:   target,
		DamageAmount:   amount,
		IsCombatDamage: combat,
		SourceLKI:      lki,
	}
}

// --- Angrath's Marauders: damageSourceControlledBy -------------------

// TestAngrathsMaraudersReadsTheDamageEventsLastKnownController: a
// source that dealt its damage while the Marauders' controller
// controlled it, and has since gone home to its owner (a stolen
// creature's "when this dies" trigger), still doubles — even though
// the current-zone card the pre-fix code read belongs to the
// opponent.
func TestAngrathsMaraudersReadsTheDamageEventsLastKnownController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	srcID := pushCatalogPermanent(g, me.ID, "Angrath's Marauders", "Creature — Human Pirate", angrathsMaraudersOracle, false)
	src := b12Card(t, g, srcID)
	spec, ok := Lookup(angrathsMaraudersOracle)
	if !ok || len(spec.Replacements) != 1 {
		t.Fatalf("Angrath's Marauders declares one replacement effect, got %d", len(spec.Replacements))
	}

	// The current-zone card: back under its OWNER (the opponent), the
	// way a stolen creature's graveyard card is.
	dealer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Returned Card", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})

	mineWhenItDealt := b1430DamageEvent(dealer, opp.ID, 1, false, &game.Characteristic{Controller: me.ID})
	theirsWhenItDealt := b1430DamageEvent(dealer, me.ID, 1, false, &game.Characteristic{Controller: opp.ID})
	liveFallback := b1430DamageEvent(srcID, opp.ID, 1, false, nil)

	var appliesMine, appliesTheirs, appliesLive bool
	g.WithWriteLock(func() {
		appliesMine = spec.Replacements[0].AppliesTo(mineWhenItDealt, g, &src)
		appliesTheirs = spec.Replacements[0].AppliesTo(theirsWhenItDealt, g, &src)
		appliesLive = spec.Replacements[0].AppliesTo(liveFallback, g, &src)
	})
	if !appliesMine {
		t.Error("a source I controlled when it dealt the damage doubles, even though its current-zone " +
			"card (the graveyard, under its owner) does not (#1430, CR 608.2h)")
	}
	if appliesTheirs {
		t.Error("a source the opponent controlled when it dealt the damage must not double, whatever the current zone says")
	}
	if !appliesLive {
		t.Error("a live source with no SourceLKI still falls back to a current-zone lookup and doubles")
	}
}

// --- Fated Firepower (the batch 28 helper): b28DamageSourceControlledBy

// TestFatedFirepowerReadsTheDamageEventsLastKnownController mirrors
// the Marauders proof for b28DamageSourceControlledBy, Fated
// Firepower's "a source you control" gate.
func TestFatedFirepowerReadsTheDamageEventsLastKnownController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	srcID := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fated Firepower", TypeLine: "Enchantment",
		OracleID: b28FatedFirepowerOracle, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{"fire": 3},
	})
	src := b12Card(t, g, srcID)
	spec, ok := Lookup(b28FatedFirepowerOracle)
	if !ok || len(spec.Replacements) != 1 {
		t.Fatalf("Fated Firepower declares one replacement effect, got %d", len(spec.Replacements))
	}

	dealer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Returned Card", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})

	mineWhenItDealt := b1430DamageEvent(dealer, opp.ID, 2, false, &game.Characteristic{Controller: me.ID})
	theirsWhenItDealt := b1430DamageEvent(dealer, opp.ID, 2, false, &game.Characteristic{Controller: opp.ID})
	liveFallback := b1430DamageEvent(srcID, opp.ID, 2, false, nil)

	var appliesMine, appliesTheirs, appliesLive bool
	g.WithWriteLock(func() {
		appliesMine = spec.Replacements[0].AppliesTo(mineWhenItDealt, g, &src)
		appliesTheirs = spec.Replacements[0].AppliesTo(theirsWhenItDealt, g, &src)
		appliesLive = spec.Replacements[0].AppliesTo(liveFallback, g, &src)
	})
	if !appliesMine {
		t.Error("a source I controlled when it dealt the damage should add the fire counters, even though " +
			"its current-zone card belongs to the opponent (#1430, CR 608.2h)")
	}
	if appliesTheirs {
		t.Error("a source the opponent controlled when it dealt the damage must not add my fire counters")
	}
	if !appliesLive {
		t.Error("a live source with no SourceLKI still falls back to a current-zone lookup")
	}
}

// --- Gratuitous Violence ----------------------------------------------

// TestGratuitousViolenceReadsTheDamageEventsLastKnownControllerAndType:
// Gratuitous Violence tests BOTH the source's controller and that it
// is a creature. A departed source that was a creature I controlled
// when it dealt the damage still doubles, even though its
// current-zone card is neither (an opponent's noncreature).
func TestGratuitousViolenceReadsTheDamageEventsLastKnownControllerAndType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	srcID := pushCatalogPermanent(g, me.ID, "Gratuitous Violence", "Enchantment", b13GratuitousViolenceOracle, false)
	src := b12Card(t, g, srcID)
	spec, ok := Lookup(b13GratuitousViolenceOracle)
	if !ok || len(spec.Replacements) != 1 {
		t.Fatalf("Gratuitous Violence declares one replacement effect, got %d", len(spec.Replacements))
	}

	// Current zone: an opponent's noncreature artifact — wrong
	// controller AND wrong type.
	dealer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Returned Rock", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID,
	})

	mineACreatureWhenItDealt := b1430DamageEvent(dealer, opp.ID, 1, false,
		&game.Characteristic{Controller: me.ID, Types: []string{"Creature"}})
	mineButNotACreature := b1430DamageEvent(dealer, opp.ID, 1, false,
		&game.Characteristic{Controller: me.ID, Types: []string{"Artifact"}})
	aCreatureButTheirs := b1430DamageEvent(dealer, opp.ID, 1, false,
		&game.Characteristic{Controller: opp.ID, Types: []string{"Creature"}})
	liveFallback := b1430DamageEvent(srcID, opp.ID, 1, false, nil)

	var appliesMineCreature, appliesMineNoncreature, appliesTheirsCreature, appliesLive bool
	g.WithWriteLock(func() {
		appliesMineCreature = spec.Replacements[0].AppliesTo(mineACreatureWhenItDealt, g, &src)
		appliesMineNoncreature = spec.Replacements[0].AppliesTo(mineButNotACreature, g, &src)
		appliesTheirsCreature = spec.Replacements[0].AppliesTo(aCreatureButTheirs, g, &src)
		// Gratuitous Violence is itself an Enchantment, not a
		// creature, so its own damage (e.g. from an effect naming
		// it) must not double — the live fallback proves the type
		// gate still runs when there is no snapshot.
		appliesLive = spec.Replacements[0].AppliesTo(liveFallback, g, &src)
	})
	if !appliesMineCreature {
		t.Error("a creature I controlled when it dealt the damage doubles, even though its current-zone " +
			"card is an opponent's noncreature (#1430, CR 608.2h)")
	}
	if appliesMineNoncreature {
		t.Error("a source that was NOT a creature when it dealt the damage must not double, whatever its type became later")
	}
	if appliesTheirsCreature {
		t.Error("a creature the opponent controlled when it dealt the damage must not double")
	}
	if appliesLive {
		t.Error("Gratuitous Violence itself is not a creature, so its own live-fallback damage must not double")
	}
}

// --- Inquisitor's Flail -------------------------------------------------

// TestInquisitorsFlailReadsTheDamageEventsLastKnownType: the second
// clause ("if another creature would deal combat damage to equipped
// creature") tests only the dealer's TYPE. A dealer that was a
// creature when it dealt the damage still doubles the damage coming
// back, even though its current-zone card (returned to combat's
// simultaneous SBA sweep as, say, an animated permanent whose
// creature-granting effect ended) is not one.
func TestInquisitorsFlailReadsTheDamageEventsLastKnownType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	flailID := seedEquipment(g, me.ID, "Inquisitor's Flail", b39InquisitorsFlailOracle)
	equipTo(t, g, me.ID, flailID, bear)
	src := b12Card(t, g, flailID)

	spec, ok := Lookup(b39InquisitorsFlailOracle)
	if !ok || len(spec.Replacements) != 2 {
		t.Fatalf("Inquisitor's Flail declares two replacement effects, got %d", len(spec.Replacements))
	}
	inbound := spec.Replacements[1]

	// Current zone: a noncreature artifact (the animation that made it
	// a creature during combat has since ended).
	dealer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Formerly Animated", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})

	wasACreature := b1430DamageEvent(dealer, bear, 1, true, &game.Characteristic{Types: []string{"Creature"}})
	wasNotACreature := b1430DamageEvent(dealer, bear, 1, true, &game.Characteristic{Types: []string{"Artifact"}})
	notCombat := b1430DamageEvent(dealer, bear, 1, false, &game.Characteristic{Types: []string{"Creature"}})

	var appliesCreature, appliesNoncreature, appliesNoncombat bool
	g.WithWriteLock(func() {
		appliesCreature = inbound.AppliesTo(wasACreature, g, &src)
		appliesNoncreature = inbound.AppliesTo(wasNotACreature, g, &src)
		appliesNoncombat = inbound.AppliesTo(notCombat, g, &src)
	})
	if !appliesCreature {
		t.Error("a dealer that WAS a creature when it dealt the damage doubles the damage coming back, " +
			"even though its current-zone card is a plain artifact (#1430, CR 608.2h)")
	}
	if appliesNoncreature {
		t.Error("a dealer that was NOT a creature when it dealt the damage must not double, whatever it looks like now")
	}
	if appliesNoncombat {
		t.Error("noncombat damage is never doubled by the Flail, snapshot or not")
	}
}
