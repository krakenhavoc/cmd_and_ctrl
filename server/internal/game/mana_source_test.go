package game

import (
	"testing"

	"github.com/google/uuid"
)

// mana_source_test.go — #1212: a mana token remembers WHAT made it,
// and the record follows the spell onto the permanent it becomes.
//
// The case the whole file exists for is the first one below: a
// Treasure pays for a spell by sacrificing itself, so every lookup
// through ManaToken.Source answers "nothing" by the time the spell
// resolves. If the snapshot is not taken before the cost is paid there
// is no later moment that can take it.

// pushTreasureFor seeds a real-shaped Treasure token: a tap-and-
// sacrifice mana ability offering all five colours, which is a
// multi-option slot and therefore queues a colour pick the way the
// catalog's Treasure template does.
func pushTreasureFor(g *Game, owner *Player) uuid.UUID {
	return pushIntrinsicPermanent(g, owner, "Treasure", "Token Artifact — Treasure",
		[]ManaAbilityShape{{TapCost: true, SacrificeCost: true, Produced: "{W|U|B|R|G}"}}, nil)
}

// pushPetalFor seeds a Lotus Petal-shaped source: the same
// tap-and-sacrifice cost with a single printed colour, so it mints
// straight into the pool instead of queueing a pick. The other half of
// the sacrifice path.
func pushPetalFor(g *Game, owner *Player, color string) uuid.UUID {
	return pushIntrinsicPermanent(g, owner, "Lotus Petal", "Artifact",
		[]ManaAbilityShape{{TapCost: true, SacrificeCost: true, Produced: "{" + color + "}"}}, nil)
}

// pushSnowIslandFor seeds a Snow-supertyped Island.
func pushSnowIslandFor(g *Game, owner *Player) uuid.UUID {
	return pushBattlefieldForTest(g, owner.ID, "Snow-Covered Island", "Basic Snow Land — Island", "")
}

// answerManaPick answers the one queued colour pick.
func answerManaPick(t *testing.T, g *Game, p *Player, color string) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c.Kind != PendingChoiceMana {
			continue
		}
		if err := g.ResolveManaChoice(c.ID, p.ID, color); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
		return
	}
	t.Fatal("no colour pick queued")
}

// tokenFrom finds the pool token a given source minted.
func tokenFrom(p *Player, source uuid.UUID) (ManaToken, bool) {
	for _, tok := range p.ManaPool {
		if tok.Source == source {
			return tok, true
		}
	}
	return ManaToken{}, false
}

// --- the snapshot --------------------------------------------------

// The headline. A Treasure is in a graveyard — and a Treasure TOKEN
// has ceased to exist (CR 111.7) — before its mana ever reaches the
// pool, so the kinds have to be read before the cost is paid. The pick
// it queues is answered a beat later still, which is why the snapshot
// rides the PendingChoice as well as the token.
func TestTreasureManaRemembersItWasATreasureAfterTheTreasureIsGone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	treasure := pushTreasureFor(g, me)

	if err := g.ActivateManaAbility(me.ID, treasure, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility on the Treasure: %v", err)
	}
	// The Treasure paid for its own ability and is gone before the
	// colour is even chosen.
	if g.Battlefield.Contains(treasure) {
		t.Fatal("the Treasure survived its own sacrifice cost; this test proves nothing")
	}
	var pick *PendingChoice
	for _, c := range g.PendingChoices {
		if c.Kind == PendingChoiceMana {
			pick = c
		}
	}
	if pick == nil {
		t.Fatal("no colour pick queued by the Treasure")
	}
	if !pick.ManaSourceKinds.Has(ManaSourceTreasure) {
		t.Fatalf("the queued pick lost the source snapshot: %b", pick.ManaSourceKinds)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "B"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}

	tok, ok := tokenFrom(me, treasure)
	if !ok {
		t.Fatalf("the Treasure minted nothing: %+v", me.ManaPool)
	}
	if !tok.SourceKinds.Has(ManaSourceTreasure) {
		t.Errorf("token = %+v, want it to remember it came from a Treasure", tok)
	}
	// A Treasure is an artifact, and a reader that had to derive that
	// from the subtype would be a reader that gets it wrong.
	if !tok.SourceKinds.Has(ManaSourceArtifact) {
		t.Errorf("token = %+v, want the artifact bit too", tok)
	}
	if tok.SourceKinds.Has(ManaSourceLand) {
		t.Errorf("token = %+v, a Treasure is not a land", tok)
	}
}

// The single-colour half of the same path: a Lotus Petal mints
// straight into the pool, still after its own sacrifice.
func TestASacrificedSingleColourSourceStillRecordsItsKinds(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	petal := pushPetalFor(g, me, "G")

	if err := g.ActivateManaAbility(me.ID, petal, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if g.Battlefield.Contains(petal) {
		t.Fatal("the Petal survived its own sacrifice cost")
	}
	tok, ok := tokenFrom(me, petal)
	if !ok {
		t.Fatalf("the Petal minted nothing: %+v", me.ManaPool)
	}
	if !tok.SourceKinds.Has(ManaSourceArtifact) {
		t.Errorf("token = %+v, want the artifact bit", tok)
	}
	if tok.SourceKinds.Has(ManaSourceTreasure) {
		t.Errorf("token = %+v, a Lotus Petal is not a Treasure", tok)
	}
}

// Snow and land, off a basic, through the ordinary tap path.
func TestASnowLandsManaRecordsSnowAndLand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	island := pushSnowIslandFor(g, me)

	if err := g.ActivateManaAbility(me.ID, island, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	tok, ok := tokenFrom(me, island)
	if !ok {
		t.Fatalf("the Island minted nothing: %+v", me.ManaPool)
	}
	if !tok.SourceKinds.Has(ManaSourceSnow) {
		t.Errorf("token = %+v, want the snow bit", tok)
	}
	if !tok.SourceKinds.Has(ManaSourceLand) {
		t.Errorf("token = %+v, want the land bit", tok)
	}
	if tok.SourceKinds.Has(ManaSourceCreature) {
		t.Errorf("token = %+v, an Island is not a creature", tok)
	}
}

// A mana creature records the creature bit — Inga and Esika's "three
// or more mana from creatures".
func TestAManaCreaturesManaRecordsCreature(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	birds := pushIntrinsicPermanent(g, me, "Birds of Paradise", "Creature — Bird",
		[]ManaAbilityShape{{TapCost: true, Produced: "{G}"}}, nil)
	// CR 302.6: it has to have been here since the turn began.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == birds {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}

	if err := g.ActivateManaAbility(me.ID, birds, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	tok, ok := tokenFrom(me, birds)
	if !ok {
		t.Fatalf("the Birds minted nothing: %+v", me.ManaPool)
	}
	if !tok.SourceKinds.Has(ManaSourceCreature) {
		t.Errorf("token = %+v, want the creature bit", tok)
	}
}

// --- the record a spell reads --------------------------------------

// A mixed payment: the record's per-symbol counts, its total and its
// source kinds all match what was actually spent, and none of them is
// recomputed from a board that has moved on.
func TestTheRecordMatchesAMixedPayment(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	treasure := pushTreasureFor(g, me)
	if err := g.ActivateManaAbility(me.ID, treasure, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	answerManaPick(t, g, me, "B")
	// Two more black mana from lands, so the payment mixes a Treasure
	// with sources that are not one.
	swamp := pushBattlefieldForTest(g, me.ID, "Swamp", "Basic Land — Swamp", "")
	if err := g.ActivateManaAbility(me.ID, swamp, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	swamp2 := pushBattlefieldForTest(g, me.ID, "Swamp", "Basic Land — Swamp", "")
	if err := g.ActivateManaAbility(me.ID, swamp2, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}

	item := castForTest(t, g, me, "{2}{B}", CastSpellParams{Strict: true})
	spent := item.Paid.Spent()
	if !spent.Known() {
		t.Fatal("a strict cast recorded an unknown payment")
	}
	if got := spent.Total(); got != 3 {
		t.Fatalf("Total() = %d, want 3", got)
	}
	if got := spent.Count("B"); got != 3 {
		t.Errorf("Count(\"B\") = %d, want 3", got)
	}
	if got := spent.ColorCount(); got != 1 {
		t.Errorf("ColorCount() = %d, want 1", got)
	}
	if !spent.FromTreasure() {
		t.Error("FromTreasure() = false; the Treasure's mana paid for this spell")
	}
	if got := spent.CountFrom(ManaSourceTreasure); got != 1 {
		t.Errorf("CountFrom(Treasure) = %d, want 1 — Marut counts exactly this", got)
	}
	if got := spent.CountFrom(ManaSourceLand); got != 2 {
		t.Errorf("CountFrom(Land) = %d, want 2", got)
	}
	if spent.Snow() {
		t.Error("Snow() = true off two Swamps and a Treasure")
	}
}

// A payment the engine waived answers in the weaker-than-printed
// direction for the SOURCE questions too, the way ADR 0068 §3 already
// requires of the colour ones. Hired Hexblade at a permissive table
// draws no card.
func TestAWaivedPaymentClaimsNoSources(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	item := castForTest(t, g, me, "{1}{B}", CastSpellParams{})
	spent := item.Paid.Spent()
	if spent.Known() {
		t.Fatal("a permissive cast did not record OnPaper")
	}
	if spent.FromTreasure() || spent.Snow() {
		t.Error("an unrecorded payment claimed a source it cannot know about")
	}
	if got := spent.CountFrom(ManaSourceLand); got != 0 {
		t.Errorf("CountFrom(Land) = %d off an unrecorded payment, want 0", got)
	}
	if spent.None() {
		t.Error("an unrecorded payment read as \"no mana was spent\" — the #259 direction")
	}
}

// --- the record a PERMANENT reads ----------------------------------

// CR 400.7d: the tokens travel onto the permanent the spell becomes,
// so "when this creature enters, if mana from a Treasure was spent to
// cast it" has something to read after the item is gone.
func TestAnEnteringPermanentRemembersTheManaThatPaidForIt(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	treasure := pushTreasureFor(g, me)
	if err := g.ActivateManaAbility(me.ID, treasure, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	answerManaPick(t, g, me, "B")

	c := NewCard("Hexblade-shaped Creature", me.ID)
	c.TypeLine = "Creature — Elf Warlock"
	c.ManaCost = "{B}"
	c.Controller = me.ID
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			t.Fatalf("resolveTopOfStackLocked: %v", err)
		}
	})

	prov := g.CastProvenanceForEffect(c.InstanceID)
	spent := prov.Spent()
	if got := spent.Total(); got != 1 {
		t.Fatalf("the permanent remembers %d mana, want 1 (provenance %+v)", got, prov)
	}
	if !spent.FromTreasure() {
		t.Error("the permanent forgot its mana came from a Treasure")
	}
	if got := spent.Count("B"); got != 1 {
		t.Errorf("Count(\"B\") = %d on the permanent, want 1", got)
	}

	// CR 400.7: the record dies with the object. A permanent that
	// leaves and comes back was not cast with anything.
	if err := g.SacrificePermanent(me.ID, c.InstanceID); err != nil {
		t.Fatalf("SacrificePermanent: %v", err)
	}
	if got := g.CastProvenanceForEffect(c.InstanceID); got.Any() {
		t.Errorf("the record survived the zone change: %+v", got)
	}
}

// The OnPaper bit crosses the entry too, and it has to. Without it a
// waived payment and a genuinely free cast are the same empty slice on
// the permanent, and ADR 0068 §3's distinction — the whole reason
// OnPaper exists — would survive the stack and die at the
// battlefield. A permissive table's Hired Hexblade draws no card
// (weaker), and a permissive table's "if NO mana was spent to cast it"
// does not fire (also weaker, and the other direction).
func TestAPermanentFromAWaivedCastRemembersThatItWasWaived(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	c := NewCard("Hexblade-shaped Creature", me.ID)
	c.TypeLine = "Creature — Elf Warlock"
	c.ManaCost = "{1}{B}"
	c.Controller = me.ID
	me.Hand.PushTop(c)
	// No Strict: the human default, where the engine never sees the
	// mana.
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			t.Fatalf("resolveTopOfStackLocked: %v", err)
		}
	})

	spent := g.CastProvenanceForEffect(c.InstanceID).Spent()
	if spent.Known() {
		t.Fatal("the permanent claims to know a payment the engine waived")
	}
	if spent.FromTreasure() {
		t.Error("the permanent claims Treasure mana off an unrecorded payment")
	}
	if spent.None() {
		t.Error("the permanent reads as \"no mana was spent\" — the #259 direction")
	}
}

// A permanent that was never cast has the zero record, and every
// question about it answers the weaker way — which is what CR 400.7d
// says about a reanimated or tokened permanent.
func TestAPermanentThatWasNeverCastRemembersNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushIntrinsicPermanent(g, me, "Bear", "Creature — Bear", nil, nil)

	spent := g.CastProvenanceForEffect(id).Spent()
	if spent.Total() != 0 || spent.FromTreasure() {
		t.Errorf("a permanent that was never cast claims a payment: %+v", spent)
	}
	if !spent.None() {
		t.Error("None() = false for a permanent that was never cast")
	}
}

// --- the auto-tapper's ordering hint --------------------------------

// The wish reorders and never filters: the wished source is the one
// that gets tapped when the planner has a free choice, and the plan is
// the same size either way.
//
// Inga and Esika's "three or more mana from creatures" is the case.
// Without the hint a Sol Ring (colourless, generic tier 0) always
// wins the generic pip over a mana creature (monocoloured, tier 2),
// which is the right instinct for every card but the twenty that read
// their payment back.
//
// DECLARED LIMIT, and it is the reason this test is not about a
// Treasure. The auto-tapper refuses every SACRIFICE-cost mana ability
// outright (autoTapAbilityFor: `!a.TapCost || a.SacrificeCost`), so a
// Treasure is not an auto-tap source at all and no ordering hint can
// reach it. Treasure mana is floated by hand, and the RECORD is what
// the cards read either way.
func TestTheAutoTapperPrefersAWishedSourceWithoutChangingPayability(t *testing.T) {
	build := func(t *testing.T) (*Game, *Player, uuid.UUID, uuid.UUID) {
		t.Helper()
		g := newActiveGame(t)
		me := g.Seats[0]
		birds := pushIntrinsicPermanent(g, me, "Birds of Paradise", "Creature — Bird",
			[]ManaAbilityShape{{TapCost: true, Produced: "{G}"}}, nil)
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == birds {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
		solring := pushIntrinsicPermanent(g, me, "Sol Ring", "Artifact",
			[]ManaAbilityShape{{TapCost: true, Produced: "{C}"}}, nil)
		pushBattlefieldForTest(g, me.ID, "Swamp", "Basic Land — Swamp", "")
		return g, me, birds, solring
	}
	cost, err := ParseCost("{1}{B}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}

	// No wish: the Sol Ring takes the generic pip.
	g, me, _, solring := build(t)
	var plain tapPlan
	var ok bool
	g.WithWriteLock(func() { plain, ok = g.autoTapPreferringLocked(me.ID, cost, 0, nil, 0) })
	if !ok {
		t.Fatal("no plan for {1}{B} off a Swamp, a Sol Ring and a Birds of Paradise")
	}
	if !plain.hasCard(solring) {
		t.Errorf("plan = %v, want the Sol Ring when nothing is wished for", plain.cardIDs())
	}

	// With the wish: the creature takes it instead, and the plan is
	// still a plan of the same size.
	var birds uuid.UUID
	g, me, birds, solring = build(t)
	var wished tapPlan
	g.WithWriteLock(func() {
		wished, ok = g.autoTapPreferringLocked(me.ID, cost, 0, nil, ManaSourceCreature)
	})
	if !ok {
		t.Fatal("the wish made a payable cost unpayable")
	}
	if !wished.hasCard(birds) {
		t.Errorf("plan = %v, want the creature when the spell reads creature mana", wished.cardIDs())
	}
	if wished.hasCard(solring) {
		t.Errorf("plan = %v, want the creature INSTEAD of the Sol Ring", wished.cardIDs())
	}
	if len(wished) != len(plain) {
		t.Errorf("the wish changed the plan's SIZE: %v vs %v", wished.cardIDs(), plain.cardIDs())
	}
}

// The same hint between two sources the planner would otherwise rank
// equal: a snow land and a plain one both pay {B}, and only the wish
// tells them apart.
func TestTheWishBreaksATieBetweenEquallyGoodSources(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushBattlefieldForTest(g, me.ID, "Swamp", "Basic Land — Swamp", "")
	snow := pushBattlefieldForTest(g, me.ID, "Snow-Covered Swamp", "Basic Snow Land — Swamp", "")

	cost, err := ParseCost("{B}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	g.WithWriteLock(func() {
		plan, ok := g.autoTapPreferringLocked(me.ID, cost, 0, nil, ManaSourceSnow)
		if !ok {
			t.Fatal("no plan for {B} off two Swamps")
		}
		if len(plan) != 1 {
			t.Fatalf("plan = %v, want one land", plan.cardIDs())
		}
		if !plan.hasCard(snow) {
			t.Errorf("plan = %v, want the snow Swamp", plan.cardIDs())
		}
	})
}

// The bound that makes the hint safe to put in front of the legal-move
// enumerator: a board with no wished source pays exactly as it did,
// and a cost that cannot be paid is not made payable by wishing.
func TestAWishChangesNeitherPayabilityNorAWishlessBoard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushBattlefieldForTest(g, me.ID, "Swamp", "Basic Land — Swamp", "")
	pushBattlefieldForTest(g, me.ID, "Swamp", "Basic Land — Swamp", "")

	payable, err := ParseCost("{1}{B}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	unpayable, err := ParseCost("{4}{B}{B}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	g.WithWriteLock(func() {
		plain, okPlain := g.autoTapPreferringLocked(me.ID, payable, 0, nil, 0)
		wished, okWished := g.autoTapPreferringLocked(me.ID, payable, 0, nil, ManaSourceTreasure)
		if okPlain != okWished {
			t.Fatalf("payability differed: %v vs %v", okPlain, okWished)
		}
		if len(plain) != len(wished) {
			t.Errorf("a wish nobody on this board satisfies changed the plan: %v vs %v",
				plain.cardIDs(), wished.cardIDs())
		}
		if _, ok := g.autoTapPreferringLocked(me.ID, unpayable, 0, nil, ManaSourceTreasure); ok {
			t.Error("a wish made an unpayable cost payable")
		}
	})
}

// --- CR 707.10: what a copy inherits --------------------------------

// A copy of a spell records a REAL zero for the mana, and the reason
// is worth stating once because the two halves of CR 707.10 point
// different ways.
//
// CR 707.10b copies "the choices made when casting" — the modes, the
// value of X, and whether an optional additional cost was paid — so a
// copy of a KICKED spell is kicked, and the engine carries that
// (spell_copy.go, ADR 0073 §5). WHICH MANA PAID is not one of those
// choices: mana is not an object, nothing was spent to cast the copy,
// and the ruling under CR 707.10 on Dawnglow Infusion says so in as
// many words. It is not a characteristic either, so CR 707.2 does not
// reach it.
//
// So the ruling this engine takes, and ADR 0068 §4 took before it: a
// copy inherits the kicker and not the mana. A copied Hired Hexblade
// draws no card, a copied converge spell converges for nothing, and a
// copied Vexing Bauble trigger counters the copy.
func TestACopyOfASpellInheritsTheKickerButNotTheManaSpent(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	treasure := pushTreasureFor(g, me)
	if err := g.ActivateManaAbility(me.ID, treasure, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	answerManaPick(t, g, me, "B")
	floatMana(me, "B")

	item := castForTest(t, g, me, "{B}{B}", CastSpellParams{Strict: true})
	g.mu.Lock()
	// The original was kicked, so the half that DOES travel has
	// something to travel with.
	item.Paid.OptionalCosts = []int{0}
	_ = g.CopySpellForEffect(item.ID, me.ID, false, nil)
	var copyItem *StackItem
	for id, it := range g.StackMeta {
		if id != item.ID {
			copyItem = it
		}
	}
	g.mu.Unlock()
	if copyItem == nil {
		t.Fatal("no copy was created")
	}

	if !item.Paid.Spent().FromTreasure() {
		t.Fatal("the ORIGINAL lost its Treasure mana; this test proves nothing")
	}
	spent := copyItem.Paid.Spent()
	if spent.Total() != 0 {
		t.Errorf("the copy inherited %d mana; nothing was spent to cast a copy (CR 707.10)", spent.Total())
	}
	if spent.FromTreasure() {
		t.Error("the copy reports Treasure mana it never spent")
	}
	if !spent.None() {
		t.Error("None() = false on a copy — a real zero is the answer, not an unknown")
	}
	if copyItem.Paid.OptionalCostTimes(0) != 1 {
		t.Error("the copy lost the KICKER, which CR 707.10b does copy")
	}
}
