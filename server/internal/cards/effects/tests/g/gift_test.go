package testsg

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the announcement ----------------------------------------------

// TestGiftAnnouncementIsValidated pins CR 702.174a's cost at announce:
// the gift is paid by naming an OPPONENT still in the game, a gift
// index with nobody named is refused, and so is a recipient sent with
// no promise.
func TestGiftAnnouncementIsValidated(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, gone := g.Seats[0], g.Seats[1], g.Seats[2]
	gone.Eliminated = true
	effectstest.AdvanceToMain(t, g)
	idx := effectstest.GiftIndex(t, effectstest.GiftDawnsTruceOracle)
	cases := []struct {
		name   string
		params game.CastSpellParams
	}{
		{"gift with no opponent", game.CastSpellParams{OptionalCosts: []int{idx}}},
		{"gift to yourself", game.CastSpellParams{OptionalCosts: []int{idx}, GiftOpponent: me.ID}},
		{"gift to a player who left", game.CastSpellParams{OptionalCosts: []int{idx}, GiftOpponent: gone.ID}},
		{"gift to nobody at the table", game.CastSpellParams{OptionalCosts: []int{idx}, GiftOpponent: uuid.New()}},
		{"a recipient with no promise", game.CastSpellParams{GiftOpponent: opp.ID}},
	}
	for _, tc := range cases {
		id := effectstest.PushCatalogHandCard(me, "Dawn's Truce", "Instant", effectstest.GiftDawnsTruceOracle)
		if err := g.CastSpell(me.ID, id, tc.params); !errors.Is(err, game.ErrInvalidParam) {
			t.Errorf("%s: got %v, want ErrInvalidParam", tc.name, err)
		}
		if !me.Hand.Contains(id) {
			t.Errorf("%s: a refused cast must leave the card in hand", tc.name)
		}
	}
	id := effectstest.PushCatalogHandCard(me, "Dawn's Truce", "Instant", effectstest.GiftDawnsTruceOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{OptionalCosts: []int{idx}, GiftOpponent: opp.ID}); err != nil {
		t.Fatalf("a gift promised to an opponent: %v", err)
	}
	item := g.StackItemForEffect(id)
	if item == nil || item.Paid.GiftOpponent != opp.ID || !item.Paid.GiftPromised() {
		t.Fatalf("the stack item must record the promise to %v, got %+v", opp.ID, item)
	}
}

// --- Dawn's Truce (the spell half, CR 702.174e / j) ----------------

func TestGiftDawnsTrucePromisedDrawsForTheOpponentAndAddsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Guy", 2, 2)
	effectstest.AdvanceToMain(t, g)
	oppHand, myHand := opp.Hand.Size(), me.Hand.Size()

	effectstest.MustCastWithGift(t, g, "Dawn's Truce", "Instant", effectstest.GiftDawnsTruceOracle, nil, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Hand.Size() != oppHand+1 {
		t.Errorf("the promised opponent draws a card: hand %d → %d", oppHand, opp.Hand.Size())
	}
	if me.Hand.Size() != myHand {
		t.Errorf("the caster draws nothing: hand %d → %d", myHand, me.Hand.Size())
	}
	abilities := effectstest.EffectiveAbilities(t, g, mine)
	if !effects.EotHasAbility(abilities, "hexproof") || !effects.EotHasAbility(abilities, "indestructible") {
		t.Errorf("a promised Truce gives hexproof AND indestructible, got %v", abilities)
	}
	if got := effectstest.PlayerAbilities(g, me); !effectstest.HasPlayerAbility(got, effects.KeywordHexproof) {
		t.Errorf("the caster still gains hexproof, got %v", got)
	}
}

// TestGiftGoesToAPlayerWhoLeftIsSkippedButStillPromised is CR 800.4a
// against CR 702.174k: a recipient who left the game gets nothing — a
// Food made for a player who is gone would be a permanent nobody is
// playing — and the spell's "if the gift was promised" branch still
// applies, because promising is the declaration, not the delivery.
func TestGiftGoesToAPlayerWhoLeftIsSkippedButStillPromised(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.MustCastWithGift(t, g, "Valley Rally", "Instant", effectstest.GiftValleyRallyOracle, effectstest.CardTarget(mine), opp.ID)
	if err := g.Concede(opp.ID); err != nil {
		t.Fatalf("concede: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if effects.IsToken(c) && c.Name == effectstest.GiftTestFoodTokenKeyName {
			t.Errorf("a player who left the game receives no gift, but a Food exists under %v", c.Controller)
		}
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, mine), "first strike") {
		t.Error("the gift was still promised, so the first strike still applies")
	}
}

// TestGiftCopyOfAPromisedSpellGivesAgain is CR 707.10: a copy copies
// the choices made when the spell was cast, the gift's opponent among
// them, so a copied promised Truce draws the same opponent a second
// card.
func TestGiftCopyOfAPromisedSpellGivesAgain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.MustCastWithGift(t, g, "Dawn's Truce", "Instant", effectstest.GiftDawnsTruceOracle, nil, opp.ID)
	oppHand := opp.Hand.Size()
	if err := g.CopySpellForEffect(id, me.ID, false, nil); err != nil {
		t.Fatalf("copy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != oppHand+2 {
		t.Errorf("the original and its copy each give a card: hand %d → %d, want +2", oppHand, opp.Hand.Size())
	}
}

// --- Long River's Pull (the clause swap, CR 702.174m) --------------

func TestGiftLongRiversPullPromisedCountersAnySpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", effectstest.GiftTestLightningOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if _, err := effectstest.CastWithGift(t, g, "Long River's Pull", "Instant", effectstest.GiftLongRiversOracle, effectstest.CardTarget(bolt), uuid.Nil); err == nil {
		t.Fatal("unpromised, a noncreature spell is not a legal target")
	}
	oppHand := opp.Hand.Size()
	effectstest.MustCastWithGift(t, g, "Long River's Pull", "Instant", effectstest.GiftLongRiversOracle, effectstest.CardTarget(bolt), opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the promised Pull counters the Bolt")
	}
	if opp.Hand.Size() != oppHand+1 {
		t.Errorf("the promised opponent draws: hand %d → %d", oppHand, opp.Hand.Size())
	}
}

// TestGiftFizzledSpellGivesNothing is CR 702.174j's last sentence: a
// spell that leaves the stack without resolving gives no gift. The
// Pull's only target resolves first, so the Pull fizzles.
func TestGiftFizzledSpellGivesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", effectstest.GiftTestLightningOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	pull := effectstest.MustCastWithGift(t, g, "Long River's Pull", "Instant", effectstest.GiftLongRiversOracle, effectstest.CardTarget(bolt), opp.ID)
	// Counter the Bolt before the Pull resolves, as a response would.
	if err := g.CounterTargetForEffect(bolt); err != nil {
		t.Fatalf("move the Bolt away: %v", err)
	}
	oppHand := opp.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pull) {
		t.Fatal("the Pull should have fizzled into its owner's graveyard")
	}
	if opp.Hand.Size() != oppHand {
		t.Errorf("a fizzled gift spell gives nothing: hand %d → %d", oppHand, opp.Hand.Size())
	}
}

// --- Into the Flood Maw (a token gift, before the other effects) ---

func TestGiftIntoTheFloodMawFishArrivesTappedBeforeTheBounce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if _, err := effectstest.CastWithGift(t, g, "Into the Flood Maw", "Instant", effectstest.GiftFloodMawOracle, effectstest.CardTarget(rock), uuid.Nil); err == nil {
		t.Fatal("unpromised, only a creature is a legal target")
	}
	from := len(g.Events)
	effectstest.MustCastWithGift(t, g, "Into the Flood Maw", "Instant", effectstest.GiftFloodMawOracle, effectstest.CardTarget(rock), opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(rock) {
		t.Fatal("the promised Maw returns the artifact to its owner's hand")
	}
	n, tapped := effectstest.CountTokensNamed(g, opp.ID, "Fish")
	if n != 1 || tapped != 1 {
		t.Errorf("the opponent gets one TAPPED Fish: %d Fish, %d tapped", n, tapped)
	}
	if mine, _ := effectstest.CountTokensNamed(g, me.ID, "Fish"); mine != 0 {
		t.Error("the caster gets no Fish")
	}
	// CR 702.174j: the gift happens before any other spell ability.
	tokenAt, bounceAt := -1, -1
	for i, ev := range g.Events[from:] {
		if ev.Kind == game.EventTokenCreated && tokenAt < 0 {
			tokenAt = i
		}
		if ev.Kind == game.EventZoneMove && ev.CardID == rock && bounceAt < 0 {
			bounceAt = i
		}
	}
	if tokenAt < 0 || bounceAt < 0 || tokenAt > bounceAt {
		t.Errorf("the Fish (event %d) must be created before the bounce (event %d)", tokenAt, bounceAt)
	}
}

// --- Wear Down / Peerless Recycling (a wider count) ----------------

func TestGiftWearDownPromisedDestroysTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	aura := effectstest.B12Permanent(g, opp.ID, "Aura", "Enchantment")
	if _, err := effectstest.CastWithGift(t, g, "Wear Down", "Sorcery", effectstest.GiftWearDownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: aura}}, uuid.Nil); err == nil {
		t.Fatal("unpromised, Wear Down takes one target")
	}
	effectstest.MustCastWithGift(t, g, "Wear Down", "Sorcery", effectstest.GiftWearDownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: aura}}, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(aura) {
		t.Error("the promised Wear Down destroys both")
	}
}

// TestGiftPromisedClauseSurvivesARestore: the clause a spell was
// announced under is catalog data a restored game re-derives
// (castTargetSpecForItem), so a promised Long River's Pull restored
// from a snapshot mid-stack must still be judged under "target spell"
// at resolution (CR 608.2b) — under the printed "target creature
// spell" its Bolt would be illegal and the Pull would fizzle.
func TestGiftPromisedClauseSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", effectstest.GiftTestLightningOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.MustCastWithGift(t, g, "Long River's Pull", "Instant", effectstest.GiftLongRiversOracle, effectstest.CardTarget(bolt), opp.ID)

	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round game.GameSnapshot
	if err := json.Unmarshal(blob, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := round.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	life := restored.Seats[0].Life
	effectstest.PassPriorityAroundTable(t, restored)
	if restored.Seats[0].Life != life {
		t.Errorf("the restored promised Pull must still counter the Bolt: life %d → %d", life, restored.Seats[0].Life)
	}
}

func TestGiftPeerlessRecyclingPromisedReturnsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B17GraveyardCard(me, "Dead Signet", "Artifact", "{2}")
	b := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	oppHand := opp.Hand.Size()
	effectstest.MustCastWithGift(t, g, "Peerless Recycling", "Instant", effectstest.GiftPeerlessOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the promised Recycling returns both permanent cards")
	}
	if opp.Hand.Size() != oppHand+1 {
		t.Errorf("the opponent draws: hand %d → %d", oppHand, opp.Hand.Size())
	}
}

// --- Valley Rally (a Food, and a clause only the promise has) ------

func TestGiftValleyRallyTargetsOnlyWhenPromised(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	other := effectstest.PushVanillaCreature(g, me.ID, "My Other Bear", 2, 2)

	if _, err := effectstest.CastWithGift(t, g, "Valley Rally", "Instant", effectstest.GiftValleyRallyOracle, effectstest.CardTarget(mine), uuid.Nil); err == nil {
		t.Fatal("unpromised, Valley Rally has no target to take (CR 702.174m)")
	}
	if _, err := effectstest.CastWithGift(t, g, "Valley Rally", "Instant", effectstest.GiftValleyRallyOracle, nil, opp.ID); err == nil {
		t.Fatal("promised, Valley Rally must name its target creature")
	}
	effectstest.MustCastWithGift(t, g, "Valley Rally", "Instant", effectstest.GiftValleyRallyOracle, effectstest.CardTarget(mine), opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, mine); got != 4 {
		t.Errorf("creatures you control get +2/+0: power %d, want 4", got)
	}
	if got := effectstest.EffectivePower(t, g, other); got != 4 {
		t.Errorf("every creature you control gets +2/+0: power %d, want 4", got)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, mine), "first strike") {
		t.Error("the target gains first strike")
	}
	if effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, other), "first strike") {
		t.Error("only the target gains first strike")
	}
	if n, _ := effectstest.CountTokensNamed(g, opp.ID, effectstest.GiftTestFoodTokenKeyName); n != 1 {
		t.Errorf("the opponent gets a Food, got %d", n)
	}
}

func TestGiftValleyRallyUnpromisedIsAPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.MustCastWithGift(t, g, "Valley Rally", "Instant", effectstest.GiftValleyRallyOracle, nil, uuid.Nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, mine); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	if effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, mine), "first strike") {
		t.Error("no first strike without the promise")
	}
	if n, _ := effectstest.CountTokensNamed(g, opp.ID, effectstest.GiftTestFoodTokenKeyName); n != 0 {
		t.Errorf("no Food without the promise, got %d", n)
	}
}

func TestGiftScrapshooterPromisedGivesOnEntryAndDestroys(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	oppHand, myHand := opp.Hand.Size(), me.Hand.Size()
	id := effectstest.MustCastWithGift(t, g, "Scrapshooter", "Creature — Raccoon Archer", effectstest.GiftScrapshooterOracle, nil, opp.ID)
	effectstest.ResolveCreatureSpell(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Scrapshooter should be on the battlefield")
	}
	if c, _ := g.LookupCardForEffect(id); !c.GiftPromised() || c.Provenance.GiftOpponent != opp.ID {
		t.Fatalf("the permanent must remember the promise (CR 400.7d), got %+v", c.Provenance)
	}
	effectstest.PickTriggerTarget(t, g, me.ID, rock)
	// #1529: the gift draw and the destroy trigger are one batch.
	effectstest.AnswerTriggerOrderLastQueuedFirst(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the promised Scrapshooter destroys the artifact")
	}
	if opp.Hand.Size() != oppHand+1 {
		t.Errorf("the opponent draws on entry: hand %d → %d", oppHand, opp.Hand.Size())
	}
	if me.Hand.Size() != myHand {
		t.Errorf("the caster draws nothing: hand %d → %d", myHand, me.Hand.Size())
	}
}

func TestGiftScrapshooterUnpromisedTriggersNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	oppHand := opp.Hand.Size()
	id := effectstest.MustCastWithGift(t, g, "Scrapshooter", "Creature — Raccoon Archer", effectstest.GiftScrapshooterOracle, nil, uuid.Nil)
	effectstest.ResolveCreatureSpell(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TriggerOnStack(g, id) != nil {
		t.Fatal("an unpromised Scrapshooter puts no trigger on the stack (CR 603.4)")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || opp.Hand.Size() != oppHand {
		t.Error("nothing happens without the promise")
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, id), "reach") {
		t.Error("Scrapshooter has reach")
	}
}

// --- declaration guards --------------------------------------------

func TestGiftRegisterRefusesAHandRolledGiftCost(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a gift cost in OptionalCosts must panic at Register")
		}
	}()
	effects.Register(effects.Spec{
		OracleID: "test-hand-rolled-gift",
		Name:     "Hand-Rolled Gift",
		OptionalCosts: []game.AdditionalCost{{
			Optional: true, Key: game.GiftKey, ChoosesOpponent: true, Label: "Gift a card",
		}},
	})
}

func TestGiftRegisterRefusesAGiftBuiltByHand(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a Gift with no effect must panic at Register")
		}
	}()
	effects.Register(effects.Spec{OracleID: "test-empty-gift", Name: "Empty Gift", Gift: &effects.Gift{Label: "Gift a card"}})
}
