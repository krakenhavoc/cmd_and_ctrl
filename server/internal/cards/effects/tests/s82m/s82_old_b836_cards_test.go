package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB836Heal          = "013c42a4-d555-46c9-9d57-b50d6b59ae9e"
	oldB836RayOfErasure  = "29d828fa-a813-48f6-8dcb-43f6047b88a3"
	oldB836ForceVoid     = "469bdca3-d373-4420-9266-921b6605f327"
	oldB836TouchOfDeath  = "52a768eb-16eb-4643-b8be-7a4b5b32f28b"
	oldB836Infuse        = "6035303e-22b3-4b64-8160-1f46ae3dad55"
	oldB836AltarOfBone   = "607a3f5f-4a36-4077-b5cf-40ffd52fa580"
	oldB836LightningBlow = "6b7bbdb7-bdd5-4eb1-912e-f2f810772ac8"
	oldB836BlessedWine   = "8079007c-a07a-4d71-bb08-e1dd215297de"
)

var oldB836Cards = map[string]string{
	oldB836Heal: "Heal", oldB836RayOfErasure: "Ray of Erasure", oldB836ForceVoid: "Force Void",
	oldB836TouchOfDeath: "Touch of Death", oldB836Infuse: "Infuse", oldB836AltarOfBone: "Altar of Bone",
	oldB836LightningBlow: "Lightning Blow", oldB836BlessedWine: "Blessed Wine",
}

func TestOldB836EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB836Cards {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB836DrawsAtTheNextUpkeep checks the card's delayed draw: nothing now, one
// card for the caster at the beginning of the next turn's upkeep.
func oldB836DrawsAtTheNextUpkeep(t *testing.T, g *game.Game, me *game.Player, name string) {
	t.Helper()
	before := me.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("%s: hand %d -> %d at the next upkeep, want +1", name, before, got)
	}
}

func oldB836Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestOldB836BlessedWineGainsALifeAndDrawsLater(t *testing.T) {
	g, me, _ := oldB836Table(t)
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Blessed Wine", "Instant", oldB836BlessedWine, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
	if me.Hand.Size() != hand {
		t.Error("drew immediately")
	}
	oldB836DrawsAtTheNextUpkeep(t, g, me, "Blessed Wine")
}

func TestOldB836TouchOfDeathDrainsAndDrawsLater(t *testing.T) {
	g, me, opp := oldB836Table(t)
	life, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Touch of Death", "Sorcery", oldB836TouchOfDeath,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != life+1 {
		t.Errorf("life: opponent %d -> %d, me %d -> %d; want -1 and +1", oppLife, opp.Life, life, me.Life)
	}
	oldB836DrawsAtTheNextUpkeep(t, g, me, "Touch of Death")
}

func TestOldB836RayOfErasureMillsAndDrawsLater(t *testing.T) {
	g, me, opp := oldB836Table(t)
	lib := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Ray of Erasure", "Instant", oldB836RayOfErasure,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-1 {
		t.Errorf("library %d -> %d, want -1", lib, opp.Library.Size())
	}
	oldB836DrawsAtTheNextUpkeep(t, g, me, "Ray of Erasure")
}

func TestOldB836LightningBlowAndInfuse(t *testing.T) {
	g, me, _ := oldB836Table(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Blow", "Instant", oldB836LightningBlow, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("no first strike")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	effectstest.CastCatalogSpell(t, g, "Infuse", "Instant", oldB836Infuse, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, bear) {
		t.Error("the Bear stayed tapped")
	}
	before := me.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d -> %d, want +2 (one per spell)", before, got)
	}
}

func TestOldB836HealPreventsOneAndDrawsLater(t *testing.T) {
	g, me, opp := oldB836Table(t)
	effectstest.CastCatalogSpell(t, g, "Heal", "Instant", oldB836Heal, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Giant", 5, 5)
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, me.ID, 3); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want 1 of 3 prevented", life, me.Life)
	}
	oldB836DrawsAtTheNextUpkeep(t, g, me, "Heal")
}

func TestOldB836ForceVoidCountersUnlessOneIsPaid(t *testing.T) {
	g, me, _ := oldB836Table(t)
	dummy := effectstest.CastCatalogSpell(t, g, "Dummy", "Sorcery", "", nil)
	effectstest.CastCatalogSpell(t, g, "Force Void", "Instant", oldB836ForceVoid, effectstest.CardRefs(dummy))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AnswerPayUnless(t, g, me.ID, false) == nil {
		t.Fatal("no pay-or-be-countered prompt")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(dummy); z == nil || z.Kind != game.ZoneGraveyard {
		t.Error("the unpaid spell was not countered")
	}
	oldB836DrawsAtTheNextUpkeep(t, g, me, "Force Void")
}

func TestOldB836AltarOfBoneSacrificesACreatureAndFetchesOne(t *testing.T) {
	g, me, _ := oldB836Table(t)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Target Dragon", TypeLine: "Creature — Dragon", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Altar of Bone", TypeLine: "Sorcery", OracleID: oldB836AltarOfBone, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{G}{W}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast without sacrificing a creature")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Target Dragon")
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, fodder) {
		t.Error("the creature was not sacrificed")
	}
	got := false
	for _, c := range me.Hand.Cards {
		if c.Name == "Target Dragon" {
			got = true
		}
	}
	if !got || me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d (the Dragon in hand: %v), want the creature tutored", hand, me.Hand.Size(), got)
	}
}
