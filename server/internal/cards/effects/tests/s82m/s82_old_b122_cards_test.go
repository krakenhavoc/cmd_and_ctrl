package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander 2020, OLD batch 122: graveyard recursion and cycling.

const (
	oldB122Herald       = "62855d00-4a6b-45a5-a1c0-2e7699e821cb"
	oldB122EverAfter    = "86be501c-6f05-4ce0-adb0-7adb135bfb93"
	oldB122Sarcophagus  = "a051dee0-60c8-4f58-84cb-55460c097115"
	oldB122Phoenix      = "a86710da-e498-4ca1-832c-8ba4abd6fa46"
	oldB122Soul         = "b5ba8155-520b-4053-be31-2744b77fb862"
	oldB122Otrimi       = "c5054364-2a19-4b70-869b-94d48b7b12c0"
	oldB122Gladiator    = "4da2425d-2118-4116-855d-2ef80693a7b7" // creature with cycling
	oldB122Miscalc      = "9078db26-3a54-4a11-9aae-e45c38da7226" // instant with cycling
	oldB122Snatcher     = "ee8fb1d0-db86-489f-8d23-4238b2e8eba4" // creature with mutate
	oldB122GladiatorTL  = "Creature — Zombie Gladiator"
	oldB122CreatureLine = "Creature — Bear"
)

func TestOldB122CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB122Herald: "Herald of the Forgotten", oldB122EverAfter: "Ever After", oldB122Sarcophagus: "Abandoned Sarcophagus",
		oldB122Phoenix: "Spellpyre Phoenix", oldB122Soul: "Soul of Innistrad", oldB122Otrimi: "Otrimi, the Ever-Playful",
	} {
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

// oldB122AnswerTargets answers the open target prompt (when there is one) and
// lets the stack settle.
func oldB122AnswerTargets(t *testing.T, g *game.Game, chooser uuid.UUID, cards ...uuid.UUID) {
	t.Helper()
	if effectstest.LatestPickTarget(g, chooser) != nil {
		effectstest.AnswerTriggerTargets(t, g, chooser, cards...)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB122HeraldReturnsEveryCyclingPermanentWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	glad := effectstest.GraveyardCardOf(g, me, "Undead Gladiator", oldB122GladiatorTL, "{4}{B}", oldB122Gladiator)
	bear := effectstest.GraveyardCardOf(g, me, "Grave Bear", oldB122CreatureLine, "{1}{G}", "")
	spell := effectstest.GraveyardCardOf(g, me, "Miscalculation", "Instant", "{1}{U}", oldB122Miscalc)
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Herald of the Forgotten", TypeLine: "Creature — Cat Beast", OracleID: oldB122Herald, Power: 6, Toughness: 6}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	oldB122AnswerTargets(t, g, me.ID, glad)
	if !g.Battlefield.Contains(glad) {
		t.Error("the cycling creature card stayed in the graveyard")
	}
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(spell) {
		t.Error("a card without cycling, or an instant, came back")
	}
}

func TestOldB122HeraldDoesNothingWhenNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	glad := effectstest.GraveyardCardOf(g, me, "Undead Gladiator", oldB122GladiatorTL, "{4}{B}", oldB122Gladiator)
	herald := effectstest.GraveyardCardOf(g, me, "Herald of the Forgotten", "Creature — Cat Beast", "{6}{W}{W}", oldB122Herald)
	effectstest.B30Reanimate(t, g, herald, me.ID)
	oldB122AnswerTargets(t, g, me.ID, glad)
	if !g.Battlefield.Contains(herald) {
		t.Fatal("the Herald was not put onto the battlefield")
	}
	if g.Battlefield.Contains(glad) {
		t.Error("an uncast Herald returned a card")
	}
}

func TestOldB122EverAfterReturnsTwoBlackZombiesAndTucksItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.GraveyardCardOf(g, me, "Grave Bear", oldB122CreatureLine, "{1}{G}", "")
	b := effectstest.GraveyardCardOf(g, me, "Grave Elk", "Creature — Elk", "{2}{G}", "")
	c := effectstest.GraveyardCardOf(g, me, "Grave Ape", "Creature — Ape", "{2}{G}", "")
	spell := effectstest.CastCatalogSpell(t, g, "Ever After", "Sorcery", oldB122EverAfter, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !g.Battlefield.Contains(id) {
			t.Fatalf("%v did not return", id)
		}
		card := effectstest.LayeredCard(t, g, id)
		if !card.HasSubtype("Zombie") || !card.HasColor("B") {
			t.Errorf("%s is not a black Zombie (%s)", card.Name, card.TypeLine)
		}
	}
	if g.Battlefield.Contains(c) {
		t.Error("a third creature returned")
	}
	if me.Graveyard.Contains(spell) {
		t.Error("Ever After is in the graveyard")
	}
	lib := me.Library.Cards
	if len(lib) == 0 || lib[0].InstanceID != spell {
		t.Error("Ever After is not on the bottom of its owner's library")
	}
}

func TestOldB122SoulOfInnistradReturnsThreeFromEitherZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	soul := effectstest.B12Push(g, me.ID, "Soul of Innistrad", "Creature — Avatar", oldB122Soul, 6, 6)
	var ids []uuid.UUID
	for _, n := range []string{"One", "Two", "Three", "Four"} {
		ids = append(ids, effectstest.GraveyardCardOf(g, me, "Grave "+n, oldB122CreatureLine, "{1}", ""))
	}
	effectstest.ApaMana(me, "C", "C", "C", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, soul, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(ids[0], ids[1], ids[2])}); err != nil {
		t.Fatalf("activate from the battlefield: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range ids {
		if want := i < 3; me.Hand.Contains(id) != want {
			t.Errorf("creature %d in hand = %v, want %v", i, !want, want)
		}
	}
	// And from the graveyard, exiling itself as the cost.
	effectstest.ApaMana(me, "C", "C", "C", "B", "B")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(soul) })
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, soul, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(ids[3])}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[3]) || me.Graveyard.Contains(soul) {
		t.Error("the graveyard ability did not exile the Soul and return the creature")
	}
}

func TestOldB122SpellpyrePhoenixReturnsACyclingSpellAndItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spell := effectstest.GraveyardCardOf(g, me, "Miscalculation", "Instant", "{1}{U}", oldB122Miscalc)
	other := effectstest.GraveyardCardOf(g, me, "Plain Instant", "Instant", "{1}{U}", "")
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Spellpyre Phoenix", TypeLine: "Creature — Phoenix", OracleID: oldB122Phoenix, Power: 4, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	oldB122AnswerTargets(t, g, me.ID, spell)
	if !me.Hand.Contains(spell) || me.Hand.Contains(other) {
		t.Error("the cycling instant was not the card returned")
	}
}

func TestOldB122SpellpyrePhoenixComesBackAfterTwoCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	phoenix := effectstest.GraveyardCardOf(g, me, "Spellpyre Phoenix", "Creature — Phoenix", "{3}{R}{R}", oldB122Phoenix)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventCycle, Actor: me.ID})
	})
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginEndStep, Actor: me.ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(phoenix) {
		t.Fatal("one cycle brought it back")
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventCycle, Actor: me.ID})
		g.EmitEvent(game.Event{Kind: game.EventBeginEndStep, Actor: me.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(phoenix) {
		t.Error("two cycles did not bring it back at the end step")
	}
}

func TestOldB122OtrimiReturnsACreatureCardWithMutate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snatcher := effectstest.GraveyardCardOf(g, me, "Souvenir Snatcher", "Creature — Drake", "{3}{U}{U}", oldB122Snatcher)
	plain := effectstest.GraveyardCardOf(g, me, "Grave Bear", oldB122CreatureLine, "{1}{G}", "")
	otrimi := effectstest.B12Push(g, me.ID, "Otrimi, the Ever-Playful", "Legendary Creature — Nightmare Beast", oldB122Otrimi, 6, 6)
	effectstest.DealCombatDamageToPlayer(g, otrimi, opp.ID, 6)
	oldB122AnswerTargets(t, g, me.ID, snatcher)
	if !me.Hand.Contains(snatcher) || me.Hand.Contains(plain) {
		t.Error("the creature card with mutate was not the one returned")
	}
}

func TestOldB122AbandonedSarcophagusOpensCyclingSpellsAndExilesTheUncycled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Abandoned Sarcophagus", oldB122Sarcophagus, "Artifact", 0, 0))
	glad := effectstest.GraveyardCardOf(g, me, "Undead Gladiator", oldB122GladiatorTL, "{0}", oldB122Gladiator)
	plain := effectstest.GraveyardCardOf(g, me, "Plain Instant", "Instant", "{0}", "")
	if !effectstest.EnumeratedFor(g, me.ID, glad) {
		t.Error("a cycling card in the graveyard is not offered")
	}
	if effectstest.EnumeratedFor(g, me.ID, plain) {
		t.Error("a card without cycling is offered")
	}
	// A creature with cycling that dies is exiled; one without goes to the graveyard.
	dier := effectstest.B12Push(g, me.ID, "Undead Gladiator", oldB122GladiatorTL, oldB122Gladiator, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dier); _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(dier) || !g.Exile.Contains(dier) {
		t.Error("a dying card with cycling was not exiled instead")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("a dying card without cycling did not reach the graveyard")
	}
}

func TestOldB122AbandonedSarcophagusLetsACycledCardReachTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Abandoned Sarcophagus", oldB122Sarcophagus, "Artifact", 0, 0))
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Undead Gladiator", TypeLine: oldB122GladiatorTL, OracleID: oldB122Gladiator, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("a cycled card was exiled instead of reaching the graveyard")
	}
}
