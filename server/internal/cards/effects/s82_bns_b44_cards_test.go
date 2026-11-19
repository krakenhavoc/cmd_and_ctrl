package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	bns44DivineFavor     = "bffa18c3-642c-4239-b425-238599007976"
	bns44DreadfulApathy  = "9c41b4ad-a73f-4274-9b71-9c9834910e05"
	bns44IndomitableWill = "cca60afe-b044-4401-8322-170aa015873c"
	bns44Malfunction     = "d2f98983-f395-4166-aa5a-309b4769baab"
	bns44RecumbentBliss  = "aaea9b83-3db5-4cdb-8e45-5c00a102c3ba"
	bns44SerrasEmbrace   = "6d6ba936-4a15-4c40-aaa6-71605fb732d1"
	bns44Slimebind       = "87aa2cbe-841c-40ff-9b1c-b259b4cc631e"
	bns44SmokeShroud     = "a621c5f6-c1c6-4bc9-88b1-f459efa07a39"
	bns44StarlitMantle   = "8a7643e4-b824-441a-bbe6-349fc7a3e11e"
	bns44Torment         = "4ba5bcee-fc95-4d05-965e-b1762529ec5f"
	bns44Eidolon         = "4c9b8b82-7f61-4dc0-aeac-fba5dc115e91"
)

// bns44Aura casts the Aura on `host` and lets it resolve.
func bns44Aura(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := castCatalogSpell(t, g, name, auraTypeLine, oracle, cardRefs(host))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatalf("%s did not resolve", name)
	}
	return aura
}

func bns44Bear(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	return pushVanillaCreature(g, owner, "Bear", 2, 2)
}

func TestBNS44DivineFavorGainsLifeAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := bns44Bear(t, g, me.ID)
	life := me.Life
	bns44Aura(t, g, "Divine Favor", bns44DivineFavor, bear)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 5 {
		t.Errorf("enchanted creature is %d/%d, want 3/5", p, tt)
	}
}

func TestBNS44IndomitableWillHasFlashAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := bns44Bear(t, g, me.ID)
	advanceTo(t, g, game.StepEnd)
	if err := castCatalogSpellErrStrict(t, g, "Indomitable Will", auraTypeLine, bns44IndomitableWill, cardRefs(bear)); err != nil {
		t.Fatalf("flash cast in the end step: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want 3/4", p, tt)
	}
}

func TestBNS44SerrasEmbraceGivesFlyingAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	bear := bns44Bear(t, g, g.Seats[0].ID)
	bns44Aura(t, g, "Serra's Embrace", bns44SerrasEmbrace, bear)
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want 4/4", p, tt)
	}
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("missing %s", kw)
		}
	}
}

func TestBNS44SlimebindShrinksPowerAtFlash(t *testing.T) {
	g := newCatalogGame(t)
	bear := bns44Bear(t, g, g.Seats[1].ID)
	advanceTo(t, g, game.StepEnd)
	if err := castCatalogSpellErrStrict(t, g, "Slimebind", auraTypeLine, bns44Slimebind, cardRefs(bear)); err != nil {
		t.Fatalf("flash cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != -2 || tt != 2 {
		t.Errorf("enchanted creature is %d/%d, want -2/2", p, tt)
	}
}

func TestBNS44StarlitMantleOnlyEnchantsYourCreatureAndGivesHexproofUntilEOT(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := bns44Bear(t, g, me.ID)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Starlit Mantle", auraTypeLine, bns44StarlitMantle, cardRefs(theirs)); err == nil {
		t.Fatal("enchant creature you control accepted an opponent's creature")
	}
	bns44Aura(t, g, "Starlit Mantle", bns44StarlitMantle, bear)
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3", p, tt)
	}
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("no hexproof after it entered")
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("hexproof outlasted the turn")
	}
}

func TestBNS44MalfunctionTapsAndKeepsAnArtifactOrCreatureTapped(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	rock := pushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	land := pushCatalogPermanent(g, opp.ID, "Plains", "Basic Land — Plains", "", false)
	if err := castCatalogSpellErrStrict(t, g, "Malfunction", auraTypeLine, bns44Malfunction, cardRefs(land)); err == nil {
		t.Fatal("a land was a legal enchant target")
	}
	bns44Aura(t, g, "Malfunction", bns44Malfunction, rock)
	if tapped, ok := battlefieldCardTapped(g, rock); !ok || !tapped {
		t.Fatal("the enchanted artifact was not tapped")
	}
	endTurn(t, g)
	advanceToUpkeepOfSeat(t, g, 1)
	advanceTo(t, g, game.StepDraw)
	if tapped, _ := battlefieldCardTapped(g, rock); !tapped {
		t.Error("the enchanted artifact untapped in its controller's untap step")
	}
}

func TestBNS44RecumbentBlissStopsAttackAndGainsLifeOnYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := bns44Bear(t, g, opp.ID)
	bns44Aura(t, g, "Recumbent Bliss", bns44RecumbentBliss, bear)
	endTurn(t, g)
	advanceToUpkeepOfSeat(t, g, 0)
	life := me.Life
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

func TestBNS44DreadfulApathyCantAttackAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := bns44Bear(t, g, me.ID)
	aura := bns44Aura(t, g, "Dreadful Apathy", bns44DreadfulApathy, bear)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err == nil {
		t.Error("the enchanted creature attacked")
	}
	std51Pay(g, opp, "{2}{W}")
	if err := g.ActivateCatalogAbility(opp.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("an opponent activated the ability")
	}
	std51Pay(g, me, "{2}{W}")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Error("the enchanted creature was not exiled")
	}
}

func bns44TormentTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Torment of Scarabs", "Enchantment — Aura Curse", bns44Torment,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	advanceToUpkeepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	return g, me, opp
}

func TestBNS44TormentOfScarabsCanLoseLife(t *testing.T) {
	g, _, opp := bns44TormentTable(t)
	life := opp.Life
	answerOptionPick(t, g, opp.ID, 0)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
}

func TestBNS44TormentOfScarabsCanSacrificeANonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Torment of Scarabs", "Enchantment — Aura Curse", bns44Torment,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	_ = me
	victim := pushVanillaCreature(g, opp.ID, "Fodder", 1, 1)
	endTurn(t, g)
	advanceToUpkeepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	life := opp.Life
	answerOptionPick(t, g, opp.ID, 1)
	if c := latestChooseCardsFor(g, opp.ID); c != nil {
		if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{victim}); err != nil {
			t.Fatalf("choose sacrifice: %v", err)
		}
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not sacrificed")
	}
	if opp.Life != life {
		t.Errorf("life %d, want %d: they sacrificed instead", opp.Life, life)
	}
}

func TestBNS44TormentOfScarabsCanDiscard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Torment of Scarabs", "Enchantment — Aura Curse", bns44Torment,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	advanceToUpkeepOfSeat(t, g, 1)
	card := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: card, Name: "Spare", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	passPriorityAroundTable(t, g)
	life := opp.Life
	answerOptionPick(t, g, opp.ID, 1) // no nonland permanent: option 1 is the discard
	if c := latestChooseCardsFor(g, opp.ID); c != nil {
		if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{card}); err != nil {
			t.Fatalf("choose discard: %v", err)
		}
	}
	if !zoneHas(opp.Graveyard, card) {
		t.Error("the card was not discarded")
	}
	if opp.Life != life {
		t.Errorf("life %d, want %d", opp.Life, life)
	}
}

func TestBNS44EidolonOfAstralWindsConstellation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	eidolon := pushCatalogPermanent(g, me.ID, "Eidolon of Astral Winds", "Enchantment Creature — Spirit", bns44Eidolon, false)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: eidolon, Actor: me.ID}) })
	pickTriggerTarget(t, g, me.ID, eidolon)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, eidolon); p != 4 || tt != 4 {
		t.Errorf("Eidolon is %d/%d, want 4/4", p, tt)
	}
	if !hasEffectiveKeyword(t, g, eidolon, "flying") {
		t.Error("no flying")
	}
	endTurn(t, g)
	if p, tt := ptOf(t, g, eidolon); p != 1 || tt != 1 {
		t.Errorf("after the turn Eidolon is %d/%d, want its own 1/1 back", p, tt)
	}
}

func TestBNS44SmokeShroudReturnsAttachedToANinja(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	shroud := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: shroud, Name: "Smoke Shroud", TypeLine: auraTypeLine,
		OracleID: bns44SmokeShroud, Owner: me.ID, Controller: me.ID})
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shroud) {
		t.Fatal("a non-Ninja brought it back")
	}
	ninja := pushCatalogPermanent(g, me.ID, "Test Ninja", "Creature — Human Ninja", "", false)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: ninja, Actor: me.ID}) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(shroud) {
		t.Fatal("Smoke Shroud did not return")
	}
	if host := attachmentHostOf(t, g, shroud); host.Kind != game.TargetCard || host.ID != ninja {
		t.Errorf("attached to %+v, want the Ninja", host)
	}
	if !hasEffectiveKeyword(t, g, ninja, "flying") {
		t.Error("the Ninja lacks flying")
	}
	_ = bear
}

func TestBNS44EidolonOfAstralWindsTriggersOnAnotherEnchantmentYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	eidolon := pushCatalogPermanent(g, me.ID, "Eidolon of Astral Winds", "Enchantment Creature — Spirit", bns44Eidolon, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	enter := func(owner uuid.UUID, typeLine string) {
		id := pushCatalogPermanent(g, owner, "Thing", typeLine, "", false)
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: owner}) })
	}
	enter(opp.ID, "Enchantment") // an opponent's enchantment: no trigger
	enter(me.ID, "Artifact")     // not an enchantment: no trigger
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("triggered on something that is not another enchantment you control")
	}
	enter(me.ID, "Enchantment")
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("Bear is %d/%d, want 4/4", p, tt)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("Bear has no flying")
	}
	if p, _ := ptOf(t, g, eidolon); p != 1 {
		t.Errorf("the Eidolon was not the target but is %d power", p)
	}
}
