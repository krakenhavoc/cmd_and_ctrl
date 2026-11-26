package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 429 (Magic 2013): Elderscale Wurm, Touch of the Eternal,
// Bountiful Harvest, Void Stalker, Courtly Provocateur, Sands of Delirium,
// Ring of Evos Isle, Ring of Xathrid, Ring of Kalonia, Mindclaw Shaman,
// Primal Clay.

const (
	oldB429Wurm        = "1e81454a-ce30-4e41-9d09-3256ae30efc2"
	oldB429Touch       = "23ecb4bb-b295-46f5-9fd8-a4c66ce80c23"
	oldB429Harvest     = "2dc94a25-4e35-4277-8eba-b8d0b293ff7b"
	oldB429Stalker     = "2313eef6-5414-462c-bf2f-f3283786735f"
	oldB429Provocateur = "3752277a-5b3c-4902-90a1-63acdac66dc9"
	oldB429Sands       = "c2cc67bf-59c5-4e06-93a3-dcaa1bf276ba"
	oldB429Evos        = "4548f379-2647-4446-9e56-e7ef0662793a"
	oldB429Xathrid     = "ceddad0d-94ad-459a-b109-f979ffc26283"
	oldB429Kalonia     = "d4344e5e-148b-4df9-88e4-ca91afcc32e0"
	oldB429Shaman      = "018985f5-6a67-4f2f-855c-efa7254ad3f8"
	oldB429Clay        = "dccf2ca8-8c87-41c9-8373-351859396d05"
)

func TestOldB429EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB429Wurm: "Elderscale Wurm", oldB429Touch: "Touch of the Eternal", oldB429Harvest: "Bountiful Harvest",
		oldB429Stalker: "Void Stalker", oldB429Provocateur: "Courtly Provocateur", oldB429Sands: "Sands of Delirium",
		oldB429Evos: "Ring of Evos Isle", oldB429Xathrid: "Ring of Xathrid", oldB429Kalonia: "Ring of Kalonia",
		oldB429Shaman: "Mindclaw Shaman", oldB429Clay: "Primal Clay",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

func oldB429SetLife(g *game.Game, p *game.Player, life int) {
	g.WithWriteLock(func() { p.Life = life })
}

func oldB429Hit(g *game.Game, victim *game.Player, n int) {
	effectstest.GptB1DamagePlayer(g, victim.ID, n)
}

// --- Elderscale Wurm ---------------------------------------------------------

func TestOldB429WurmEntersAndSetsLifeToSevenOnlyWhenBelow(t *testing.T) {
	for _, tc := range []struct{ before, want int }{{3, 7}, {7, 7}, {20, 20}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		oldB429SetLife(g, me, tc.before)
		effectstest.B12PlayFromHand(t, g, "Elderscale Wurm", "Creature — Wurm", oldB429Wurm, game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != tc.want {
			t.Errorf("life %d on entering: %d, want %d", tc.before, me.Life, tc.want)
		}
	}
}

func TestOldB429WurmFloorsDamageAtSevenWhileAtSevenOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Elderscale Wurm", "Creature — Wurm", oldB429Wurm, false)
	oldB429SetLife(g, me, 10)
	oldB429Hit(g, me, 2)
	if me.Life != 8 {
		t.Fatalf("2 damage from 10: life %d, want 8", me.Life)
	}
	oldB429Hit(g, me, 5)
	if me.Life != 7 {
		t.Fatalf("5 damage from 8: life %d, want 7", me.Life)
	}
	oldB429Hit(g, me, 3)
	if me.Life != 7 {
		t.Errorf("3 damage at 7: life %d, want 7", me.Life)
	}
	oldB429SetLife(g, me, 6)
	oldB429Hit(g, me, 2)
	if me.Life != 4 {
		t.Errorf("2 damage from 6, below 7: life %d, want 4", me.Life)
	}
}

func TestOldB429WurmDoesNotFloorLifeLossThatIsNotDamageOrOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Elderscale Wurm", "Creature — Wurm", oldB429Wurm, false)
	oldB429SetLife(g, opp, 10)
	oldB429Hit(g, opp, 8)
	if opp.Life != 2 {
		t.Errorf("an opponent took 8 from 10: life %d, want 2", opp.Life)
	}
	oldB429SetLife(g, me, 10)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -6) })
	if me.Life != 4 {
		t.Errorf("losing 6 life from 10 (not damage): life %d, want 4", me.Life)
	}
}

// --- Touch of the Eternal ----------------------------------------------------

func TestOldB429TouchSetsLifeToThePermanentCountAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Touch of the Eternal", "Enchantment", oldB429Touch, false)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Elk", "Creature — Elk", 2, 2)
	effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	oldB429SetLife(g, me, 40)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID {
			n++
		}
	}
	if n < 3 || me.Life != n {
		t.Errorf("life %d with %d permanents, want life equal to the count", me.Life, n)
	}
}

func TestOldB429TouchDoesNothingOnAnotherPlayersUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Touch of the Eternal", "Enchantment", oldB429Touch, false)
	oldB429SetLife(g, me, 40)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 40 {
		t.Errorf("life %d on an opponent's upkeep, want 40", me.Life)
	}
}

// --- Bountiful Harvest -------------------------------------------------------

func TestOldB429HarvestGainsOneLifePerLandYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	}
	effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	before := me.Life
	effectstest.B12PlayFromHand(t, g, "Bountiful Harvest", "Sorcery", oldB429Harvest, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("life %d, want %d (three lands of yours)", me.Life, before+3)
	}
}

// --- Void Stalker ------------------------------------------------------------

func TestOldB429VoidStalkerPutsItselfAndTheTargetOnTopOfTheirLibraries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stalker := effectstest.PushCatalogPermanent(g, me.ID, "Void Stalker", "Creature — Elemental", oldB429Stalker, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, stalker, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12ZoneOf(g, stalker); got != game.ZoneLibrary {
		t.Errorf("Void Stalker is in %q, want its owner's library", got)
	}
	if got := effectstest.B12ZoneOf(g, theirs); got != game.ZoneLibrary {
		t.Errorf("the target is in %q, want its owner's library", got)
	}
	if !me.Library.Contains(stalker) || !opp.Library.Contains(theirs) {
		t.Error("a card went to the wrong library")
	}
}

func TestOldB429VoidStalkerNeedsATargetAndTheTapAndMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stalker := effectstest.PushCatalogPermanent(g, me.ID, "Void Stalker", "Creature — Elemental", oldB429Stalker, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, stalker, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Error("activated with no mana")
	}
	_ = g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, stalker, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no target")
	}
}

// --- Courtly Provocateur -----------------------------------------------------

func TestOldB429ProvocateurForcesAnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	prov := effectstest.PushCatalogPermanent(g, me.ID, "Courtly Provocateur", "Creature — Human Wizard", oldB429Provocateur, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, prov, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.FaMustAttack(t, g) {
		t.Error("the target creature was not required to attack")
	}
}

func TestOldB429ProvocateurForcesABlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	prov := effectstest.PushCatalogPermanent(g, me.ID, "Courtly Provocateur", "Creature — Human Wizard", oldB429Provocateur, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, prov, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if reqs := effectstest.LayeredCard(t, g, bear).Effective().BlockRequirements; len(reqs) != 1 || reqs[0].Kind != game.BlockRequirementBlocks {
		t.Errorf("block requirements %v, want one \"blocks if able\"", reqs)
	}
	if reqs := effectstest.LayeredCard(t, g, bear).Effective().AttackRequirements; len(reqs) != 0 {
		t.Errorf("attack requirements %v, want none", reqs)
	}
}

func TestOldB429ProvocateurAbilitiesAreTapAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	prov := effectstest.PushCatalogPermanent(g, me.ID, "Courtly Provocateur", "Creature — Human Wizard", oldB429Provocateur, true)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, prov, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Error("a summoning-sick Provocateur tapped for an ability")
	}
}

// --- Sands of Delirium -------------------------------------------------------

func TestOldB429SandsMillsXCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sands := effectstest.PushCatalogPermanent(g, me.ID, "Sands of Delirium", "Artifact", oldB429Sands, false)
	effectstest.SeedCheapLibrary(g, opp, 10)
	before := len(opp.Library.Cards)
	_ = g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sands, 0, game.ActivateAbilityParams{Strict: true, XValue: 3,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - len(opp.Library.Cards); got != 3 {
		t.Errorf("milled %d, want 3", got)
	}
	if len(opp.Graveyard.Cards) != 3 {
		t.Errorf("graveyard %d, want 3", len(opp.Graveyard.Cards))
	}
}

// --- The Rings ---------------------------------------------------------------

func oldB429Ring(t *testing.T, name, oracle, color, offColor string) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	match := effectstest.B12Creature(g, me.ID, "Matching", "Creature — Bear", 2, 2)
	off := effectstest.B12Creature(g, me.ID, "Off", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case match:
				g.Battlefield.Cards[i].Colors = []string{color}
			case off:
				g.Battlefield.Cards[i].Colors = []string{offColor}
			}
		}
	})
	ring := effectstest.PushCatalogPermanent(g, me.ID, name, "Artifact — Equipment", oracle, false)
	effectstest.BrAttach(t, g, ring, match)
	return g, me, ring, match, off
}

func oldB429RingCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.LayeredCard(t, g, id).Counters[game.CounterPlusOne]
}

func oldB429RingUpkeep(t *testing.T, name, oracle, color, offColor string) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
	g, me, ring, match, off := oldB429Ring(t, name, oracle, color, offColor)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB429RingCounters(t, g, match); got != 1 {
		t.Errorf("%s: a creature of its colour has %d counters, want 1", name, got)
	}
	effectstest.BrAttach(t, g, ring, off)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB429RingCounters(t, g, off); got != 0 {
		t.Errorf("%s: a creature of another colour has %d counters, want 0", name, got)
	}
	return g, me, ring, match, off
}

func TestOldB429RingOfEvosIsleGrowsABlueHostAndGivesHexproofUntilEndOfTurn(t *testing.T) {
	g, me, ring, match, _ := oldB429RingUpkeep(t, "Ring of Evos Isle", oldB429Evos, "U", "R")
	effectstest.BrAttach(t, g, ring, match)
	if effectstest.HasEffectiveKeyword(t, g, match, "hexproof") {
		t.Fatal("hexproof before the ability")
	}
	_ = g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, match, "hexproof") {
		t.Error("the equipped creature did not gain hexproof")
	}
}

func TestOldB429RingOfXathridGrowsABlackHostAndRegeneratesIt(t *testing.T) {
	g, me, ring, match, _ := oldB429RingUpkeep(t, "Ring of Xathrid", oldB429Xathrid, "B", "R")
	effectstest.BrAttach(t, g, ring, match)
	_ = g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Damage(t, g, match, match, 10)
	g.RunStateChecksForTest()
	if got := effectstest.B12ZoneOf(g, match); got != game.ZoneBattlefield {
		t.Errorf("the regenerating creature is in %q, want the battlefield", got)
	}
}

func TestOldB429RingOfXathridDoesNothingUnattached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ring := effectstest.PushCatalogPermanent(g, me.ID, "Ring of Xathrid", "Artifact — Equipment", oldB429Xathrid, false)
	_ = g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB429RingOfKaloniaGivesTrampleAndGrowsAGreenHost(t *testing.T) {
	g, _, _, match, off := oldB429RingUpkeep(t, "Ring of Kalonia", oldB429Kalonia, "G", "R")
	if effectstest.HasEffectiveKeyword(t, g, match, "trample") {
		t.Error("an unequipped creature has trample")
	}
	if !effectstest.HasEffectiveKeyword(t, g, off, "trample") {
		t.Error("the equipped creature lacks trample")
	}
}

func TestOldB429RingsOnlyGrowOnYourOwnUpkeep(t *testing.T) {
	g, _, _, match, _ := oldB429Ring(t, "Ring of Evos Isle", oldB429Evos, "U", "R")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB429RingCounters(t, g, match); got != 0 {
		t.Errorf("%d counters on an opponent's upkeep, want 0", got)
	}
}

// --- Mindclaw Shaman ---------------------------------------------------------

func oldB429Hand(p *game.Player, typeLines ...string) []uuid.UUID {
	p.Hand.Cards = nil
	var ids []uuid.UUID
	for i, tl := range typeLines {
		c := game.NewCard("Their Card "+string(rune('A'+i)), p.ID)
		c.TypeLine = tl
		c.ManaCost = "{2}{G}"
		c.Power, c.Toughness = 2, 2
		p.Hand.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func oldB429Enter(t *testing.T, g *game.Game, me, opp *game.Player) {
	t.Helper()
	effectstest.B12PlayFromHand(t, g, "Mindclaw Shaman", "Creature — Lizard Shaman", oldB429Shaman, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB429MindclawShamanRevealsTheHandAndCastsAnInstantOrSorceryFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	hand := oldB429Hand(opp, "Creature — Bear", "Instant", "Land", "Sorcery")
	oldB429Enter(t, g, me, opp)
	offer := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, me.ID)
	if offer == nil {
		t.Fatalf("no resolution_cast offer: %+v", g.PendingChoices)
	}
	for _, id := range offer.DiscardOptions {
		if id == hand[0] || id == hand[2] {
			t.Errorf("a creature or land was offered: %v", offer.DiscardOptions)
		}
	}
	if len(offer.DiscardOptions) != 2 {
		t.Fatalf("offer %v, want the instant and the sorcery", offer.DiscardOptions)
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, hand[0], game.CastSpellParams{FromZone: "hand"}); err == nil {
		t.Fatal("a creature card was cast")
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, hand[3], game.CastSpellParams{FromZone: "hand"}); err != nil {
		t.Fatalf("cast the sorcery for free: %v", err)
	}
	if got := effectstest.B12ZoneOf(g, hand[3]); got != game.ZoneStack {
		t.Errorf("the cast card is in %q, want the stack", got)
	}
}

func TestOldB429MindclawShamanMayDeclineAndTheHandStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	hand := oldB429Hand(opp, "Instant", "Land")
	oldB429Enter(t, g, me, opp)
	offer := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, me.ID)
	if offer == nil {
		t.Fatalf("no resolution_cast offer")
	}
	if err := g.ResolvePendingChoice(offer.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	for _, id := range hand {
		if got := effectstest.B12ZoneOf(g, id); got != game.ZoneHand {
			t.Errorf("a card is in %q, want their hand", got)
		}
	}
}

func TestOldB429MindclawShamanWithNothingToCastOffersNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	oldB429Hand(opp, "Creature — Bear", "Land")
	oldB429Enter(t, g, me, opp)
	if offer := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, me.ID); offer != nil {
		t.Errorf("an offer with no instant or sorcery in hand: %+v", offer)
	}
}

// --- Primal Clay -------------------------------------------------------------

func oldB429MakeClay(t *testing.T, pick int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	id := effectstest.PushHandCardWithManaCost(me, "Primal Clay", "Artifact Creature — Shapeshifter", "")
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == id {
			me.Hand.Cards[i].OracleID = oldB429Clay
			me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 0, 0
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, pick)
	effectstest.PassPriorityAroundTable(t, g)
	return g, id
}

func TestOldB429PrimalClayBecomesTheChosenForm(t *testing.T) {
	for _, tc := range []struct {
		pick       int
		power, tgh int
		flying     bool
		wall       bool
		defender   bool
	}{
		{0, 3, 3, false, false, false},
		{1, 2, 2, true, false, false},
		{2, 1, 6, false, true, true},
	} {
		g, id := oldB429MakeClay(t, tc.pick)
		if p, th := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != tc.power || th != tc.tgh {
			t.Errorf("form %d: %d/%d, want %d/%d", tc.pick, p, th, tc.power, tc.tgh)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, id, "flying"); got != tc.flying {
			t.Errorf("form %d: flying = %v", tc.pick, got)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, id, "defender"); got != tc.defender {
			t.Errorf("form %d: defender = %v", tc.pick, got)
		}
		subs := effectstest.EffectiveSubtypes(t, g, id)
		if got := slices.Contains(subs, "Wall"); got != tc.wall {
			t.Errorf("form %d: Wall = %v (subtypes %v)", tc.pick, got, subs)
		}
		if !slices.Contains(subs, "Shapeshifter") {
			t.Errorf("form %d lost the Shapeshifter type (subtypes %v)", tc.pick, subs)
		}
		for _, ty := range []string{"Artifact", "Creature"} {
			if !effectstest.HasEffectiveType(t, g, id, ty) {
				t.Errorf("form %d lost the type %s", tc.pick, ty)
			}
		}
	}
}

var _ = effects.Lookup
