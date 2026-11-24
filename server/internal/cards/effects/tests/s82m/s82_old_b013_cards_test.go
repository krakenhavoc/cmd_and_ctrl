package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b013_cards_test.go — Secret Lair Drop, "Graveyard recursion"
// (S82 OLD batch 13, #5462): each card driven through a real cast or
// activation.

const (
	oldB013SecretsOracle   = "b9720c63-5aa7-4fb8-8d71-8f55d8b1337b"
	oldB013GaeasOracle     = "bd5478df-3854-404b-8634-9debcef69b99"
	oldB013GlissaOracle    = "cafacba9-195c-44ea-ba80-3e489274759c"
	oldB013DregscapeOracle = "d1678eed-a335-454c-9fa1-749317500a36"
	oldB013BalthorOracle   = "dfd8e6c2-e425-41fa-af55-985271f771b3"
	oldB013TorturedOracle  = "fa9dfcf7-069c-4441-9ea0-133de47ecada"
	oldB013CoffinOracle    = "c19ec6aa-aadd-4767-b695-384fd07b4ded"
	oldB013BrunaOracle     = "e8f213eb-58db-4fab-bd0b-2acfd0a86ac3"
	oldB013CroakingOracle  = "142b2f24-49f6-4ef8-9731-9ae2e513f796"
)

func TestS82OldB013AllAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{oldB013SecretsOracle, oldB013GaeasOracle, oldB013GlissaOracle,
		oldB013DregscapeOracle, oldB013BalthorOracle, oldB013TorturedOracle, oldB013CoffinOracle, oldB013BrunaOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Secrets of the Dead -----------------------------------------------------

func TestS82OldB013SecretsOfTheDeadDrawsOnAGraveyardCastOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedPermanentWithOracle(g, me.ID, "Secrets of the Dead", "Enchantment", oldB013SecretsOracle)
	effectstest.AuraSeedLibrary(me, 5)
	// A cast from hand does not draw.
	before := effectstest.HandCount(g, me)
	effectstest.CastCatalogSpell(t, g, "Croaking Counterpart", "Sorcery", oldB013CroakingOracle,
		effectstest.B16TargetCard(effectstest.B12Creature(g, me.ID, "Bear0", "Creature — Bear", 2, 2)))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != before {
		t.Errorf("a hand cast changed the hand: %d, want %d", got, before)
	}
	// A cast from the graveyard does.
	id := effectstest.SeedGraveyardCard(t, g, "Croaking Counterpart", "Sorcery", oldB013CroakingOracle)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != before+1 {
		t.Errorf("hand = %d, want %d after a graveyard cast", got, before+1)
	}
}

// --- Gaea's Blessing ---------------------------------------------------------

func TestS82OldB013GaeasBlessingShufflesUpToThreeCardsThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var theirs []uuid.UUID
	for _, n := range []string{"T1", "T2", "T3", "T4"} {
		theirs = append(theirs, effectstest.PushGraveyardCardTyped(opp, n, "Creature — Bear"))
	}
	effectstest.AuraSeedLibrary(me, 3)
	lib, hand := opp.Library.Size(), effectstest.HandCount(g, me)
	effectstest.CastCatalogSpell(t, g, "Gaea's Blessing", "Sorcery", oldB013GaeasOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: theirs[0], Slot: 1}, {Kind: game.TargetCard, ID: theirs[1], Slot: 1}, {Kind: game.TargetCard, ID: theirs[2], Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != lib+3 || !opp.Graveyard.Contains(theirs[3]) || opp.Graveyard.Contains(theirs[0]) {
		t.Errorf("library %d (want %d); the three picked cards should be in it and the fourth still in the graveyard", opp.Library.Size(), lib+3)
	}
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand = %d, want %d (the caster draws)", got, hand+1)
	}
	// A fourth target card is refused.
	if err := effectstest.CastCatalogSpellErr(t, g, "Gaea's Blessing", "Sorcery", oldB013GaeasOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: theirs[0], Slot: 1}, {Kind: game.TargetCard, ID: theirs[1], Slot: 1},
		{Kind: game.TargetCard, ID: theirs[2], Slot: 1}, {Kind: game.TargetCard, ID: theirs[3], Slot: 1},
	}); err == nil {
		t.Error("four target cards were accepted")
	}
}

func TestS82OldB013GaeasBlessingShufflesTheGraveyardWhenMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	other := effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	effectstest.SpmB17Top(me, game.Card{Name: "Gaea's Blessing", TypeLine: "Sorcery", OracleID: oldB013GaeasOracle})
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(me.ID, 1, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(other) || me.Graveyard.Size() != 0 {
		t.Errorf("the graveyard holds %d cards after the Blessing was milled, want 0", me.Graveyard.Size())
	}
}

// --- Glissa, the Traitor -----------------------------------------------------

func TestS82OldB013GlissaReturnsAnArtifactWhenAnOpponentsCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Glissa, the Traitor", "Legendary Creature — Phyrexian Zombie Elf", oldB013GlissaOracle, 3, 3)
	rock := effectstest.GraveyardCardOf(g, me, "Plain Rock", "Artifact", "{1}", "")
	// Your own creature dying does not trigger it.
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	effectstest.B15Destroy(t, g, mine)
	if me.Hand.Contains(rock) || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("Glissa triggered on a creature its controller controls")
	}
	// An opponent's does.
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the artifact card was not returned to hand")
	}
}

// --- Dregscape Sliver --------------------------------------------------------

func TestS82OldB013DregscapeSliverGivesSliverCardsInYourGraveyardUnearth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sliver := effectstest.ZgGraveyardCard(me, "Plain Sliver", "Creature — Sliver")
	bear := effectstest.ZgGraveyardCard(me, "Plain Bear", "Creature — Bear")
	theirs := effectstest.ZgGraveyardCard(opp, "Their Sliver", "Creature — Sliver")
	src := effectstest.B12Push(g, me.ID, "Dregscape Sliver", "Creature — Sliver", oldB013DregscapeOracle, 2, 2)
	for _, tc := range []struct {
		id   uuid.UUID
		name string
		want int
	}{{sliver, "Sliver", 1}, {bear, "non-Sliver", 0}, {theirs, "opponent's Sliver", 0}} {
		rows, _ := effectstest.ZgRows(g, tc.id)
		if len(rows) != tc.want {
			t.Errorf("%s: rows = %v, want %d", tc.name, effectstest.ZgLabels(rows), tc.want)
		}
	}
	effectstest.ZgLeaveBattlefield(t, g, src)
	if rows, _ := effectstest.ZgRows(g, sliver); len(rows) != 0 {
		t.Errorf("after the Sliver left, rows = %v, want none", effectstest.ZgLabels(rows))
	}
}

func TestS82OldB013DregscapeSliverGrantedUnearthReturnsASliver(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	sliver := effectstest.ZgGraveyardCard(me, "Plain Sliver", "Creature — Sliver")
	effectstest.B12Push(g, me.ID, "Dregscape Sliver", "Creature — Sliver", oldB013DregscapeOracle, 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sliver, 0, game.ActivateAbilityParams{
		Ref: "grant:dregscape-sliver/unearth:0:0", Strict: true,
	}); err != nil {
		t.Fatalf("activate the granted unearth: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sliver) {
		t.Fatal("the Sliver did not return")
	}
}

// --- Balthor the Defiled -----------------------------------------------------

func TestS82OldB013BalthorPumpsMinionsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	balthor := effectstest.B12Push(g, me.ID, "Balthor the Defiled", "Legendary Creature — Zombie Dwarf", oldB013BalthorOracle, 2, 2)
	minion := effectstest.B12Creature(g, opp.ID, "Their Minion", "Creature — Minion", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if got := effectstest.ApaLive(g, minion).CurrentPower(); got != 2 {
		t.Errorf("an opponent's Minion has power %d, want 2", got)
	}
	if got := effectstest.ApaLive(g, bear).CurrentPower(); got != 2 {
		t.Errorf("a Bear has power %d, want 2", got)
	}
	if got := effectstest.ApaLive(g, balthor).CurrentPower(); got != 2 {
		t.Errorf("Balthor has power %d, want 2 (he is not a Minion)", got)
	}
}

func TestS82OldB013BalthorReturnsBlackAndRedCreatureCardsForEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	balthor := effectstest.B12Push(g, me.ID, "Balthor the Defiled", "Legendary Creature — Zombie Dwarf", oldB013BalthorOracle, 2, 2)
	black := effectstest.GraveyardCardOf(g, me, "Black Bear", "Creature — Bear", "{1}{B}", "")
	red := effectstest.GraveyardCardOf(g, opp, "Red Bear", "Creature — Bear", "{1}{R}", "")
	green := effectstest.GraveyardCardOf(g, me, "Green Bear", "Creature — Bear", "{1}{G}", "")
	blackNon := effectstest.GraveyardCardOf(g, opp, "Black Rock", "Artifact", "{1}{B}", "")
	effectstest.FloatMana(t, g, me, "{B}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, balthor, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(balthor) {
		t.Error("Balthor was not exiled as a cost")
	}
	for name, id := range map[string]uuid.UUID{"black": black, "red": red} {
		c, ok := effectstest.BattlefieldCardByID(g, id)
		if !ok {
			t.Errorf("the %s creature card did not return", name)
			continue
		}
		if want := map[string]uuid.UUID{"black": me.ID, "red": opp.ID}[name]; c.Controller != want {
			t.Errorf("the %s creature came back under %v, want its owner", name, c.Controller)
		}
	}
	if g.Battlefield.Contains(green) || g.Battlefield.Contains(blackNon) {
		t.Error("a green creature card or a noncreature card returned")
	}
}

// --- Tortured Existence ------------------------------------------------------

func TestS82OldB013TorturedExistenceDiscardsACreatureToReturnOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	src := effectstest.SeedPermanentWithOracle(g, me.ID, "Tortured Existence", "Enchantment", oldB013TorturedOracle)
	dead := effectstest.GraveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	spell := effectstest.HandCardForTest(me, "Lightning Bolt", "Instant", "")
	bear := effectstest.HandCardForTest(me, "Hand Bear", "Creature — Bear", "")
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.B16TargetCard(dead), DiscardIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("discarded a noncreature card")
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.B16TargetCard(dead), DiscardIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card was not returned to hand")
	}
	if effectstest.ZoneOf(g, bear) != game.ZoneGraveyard {
		t.Errorf("the discarded creature is in %q", effectstest.ZoneOf(g, bear))
	}
}

// --- Coffin Queen ------------------------------------------------------------

// oldB013Reanimate has the Queen put a card from the opponent's graveyard onto
// the battlefield and returns the Queen and the creature.
func oldB013Reanimate(t *testing.T, g *game.Game) (queen, creature uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	queen = effectstest.B12Push(g, me.ID, "Coffin Queen", "Creature — Zombie Wizard", oldB013CoffinOracle, 1, 1)
	creature = effectstest.GraveyardCardOf(g, opp, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.FloatMana(t, g, me, "{B}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, queen, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(creature)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCardByID(g, creature)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the creature did not come back under my control: %+v", c)
	}
	if q, _ := effectstest.BattlefieldCardByID(g, queen); !q.Tapped {
		t.Fatal("the Queen was not tapped for her cost")
	}
	return queen, creature
}

func TestS82OldB013CoffinQueenReanimatesThenExilesItWhenSheUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	queen, creature := oldB013Reanimate(t, g)
	effectstest.RtUntap(g, queen)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(creature) {
		t.Error("the creature stayed after the Queen untapped")
	}
	if effectstest.ZoneOf(g, creature) == game.ZoneGraveyard {
		t.Error("the creature went to a graveyard, not exile")
	}
}

func TestS82OldB013CoffinQueenExilesItWhenYouLoseControlOfHer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	queen, creature := oldB013Reanimate(t, g)
	effectstest.StealForTest(t, g, queen, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(creature) {
		t.Error("the creature stayed after the Queen was stolen")
	}
	_ = me
}

func TestS82OldB013CoffinQueenKeepsItWhileSheStaysTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, creature := oldB013Reanimate(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(creature) {
		t.Error("the creature left while the Queen stayed tapped")
	}
}

func TestS82OldB013CoffinQueenMayChooseNotToUntap(t *testing.T) {
	spec, ok := effects.Lookup(oldB013CoffinOracle)
	if !ok || len(spec.UntapOptOuts) != 1 {
		t.Fatalf("Coffin Queen has %d untap opt-outs, want 1", len(spec.UntapOptOuts))
	}
}

// --- Bruna, Light of Alabaster -----------------------------------------------

func TestS82OldB013BrunaAttachesAurasFromEverywhereWhenSheAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bruna := effectstest.B12Push(g, me.ID, "Bruna, Light of Alabaster", "Legendary Creature — Angel", oldB013BrunaOracle, 5, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	onField := effectstest.B15Aura(g, opp.ID, bear)
	inHand := effectstest.HandCardForTest(me, "Hand Aura", "Enchantment — Aura", "")
	inYard := effectstest.PushGraveyardCardTyped(me, "Yard Aura", "Enchantment — Aura")
	notAura := effectstest.HandCardForTest(me, "Hand Bear", "Creature — Bear", "")
	effectstest.DeclareAttack(t, g, opp.ID, bruna)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, onField)
	effectstest.AnswerChooseCards(t, g, me.ID, inHand)
	effectstest.AnswerChooseCards(t, g, me.ID, inYard)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"the Aura already on the battlefield": onField, "the Aura card from hand": inHand, "the Aura card from the graveyard": inYard} {
		c, ok := effectstest.BattlefieldCardByID(g, id)
		if !ok || !c.IsAttachedTo(bruna) {
			t.Errorf("%s is not attached to Bruna", name)
		}
	}
	if g.Battlefield.Contains(notAura) {
		t.Error("a non-Aura card entered")
	}
}

func TestS82OldB013BrunaMayChooseNoneOfThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bruna := effectstest.B12Push(g, me.ID, "Bruna, Light of Alabaster", "Legendary Creature — Angel", oldB013BrunaOracle, 5, 5)
	inHand := effectstest.HandCardForTest(me, "Hand Aura", "Enchantment — Aura", "")
	effectstest.DeclareAttack(t, g, opp.ID, bruna)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(inHand) {
		t.Error("the Aura entered though none was chosen")
	}
}
