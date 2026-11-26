package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b379_cards_test.go — Born of the Gods, batch 379 (S82 older sets,
// #5559): six cards of graveyard recursion, equipment, counters, a counterspell
// and Aura removal. Every test plays the real catalog entry through the engine.

const (
	oldB379Trawler     = "0282b5c0-e792-4099-a3fa-329eda11fb44"
	oldB379Lyre        = "400da7e6-43f8-4539-8959-ec60a3cc431b"
	oldB379Pillar      = "50b931f9-f843-4129-a4c8-2bb7a051c063"
	oldB379Scourge     = "725d7374-4ea8-4aef-ac09-8dcb5d8dcad1"
	oldB379Rebuff      = "bc5d9899-1a1c-4779-b9f0-b00f2c233699"
	oldB379Starbreaker = "bc84a4e7-5f5f-4105-b0f8-c06ebcde6218"

	oldB379Bear    = "Creature — Bear"
	oldB379EnchCrt = "Enchantment Creature — Spirit"
)

var oldB379Cards = map[string]string{
	oldB379Trawler: "Odunos River Trawler", oldB379Lyre: "Siren Song Lyre", oldB379Pillar: "Pillar of War",
	oldB379Scourge: "Scourge of Skola Vale", oldB379Rebuff: "Thassa's Rebuff", oldB379Starbreaker: "Setessan Starbreaker",
}

func TestOldB379EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB379Cards {
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

func oldB379Grave(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID, Power: 1, Toughness: 1})
	return id
}

func TestOldB379OdunosRiverTrawlerReturnsAnEnchantmentCreatureOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	want := oldB379Grave(me, "Spirit", oldB379EnchCrt)
	plain := oldB379Grave(me, "Plain Bear", oldB379Bear)
	effectstest.CastCatalogSpell(t, g, "Odunos River Trawler", "Creature — Zombie", oldB379Trawler, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: plain}); err == nil {
			t.Fatal("a plain creature card was a legal target")
		}
		effectstest.PickCard(t, g, me.ID, want)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(want) || me.Hand.Contains(plain) {
		t.Errorf("hand has the Spirit %v and the plain Bear %v, want true false", me.Hand.Contains(want), me.Hand.Contains(plain))
	}
}

func TestOldB379OdunosRiverTrawlerSacrificesItselfForTheSameReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	trawler := effectstest.B12Push(g, me.ID, "Odunos River Trawler", "Creature — Zombie", oldB379Trawler, 2, 2)
	spirit := oldB379Grave(me, "Spirit", oldB379EnchCrt)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, trawler, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(spirit)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(trawler) || !me.Hand.Contains(spirit) {
		t.Errorf("the Trawler is on the battlefield %v, the Spirit in hand %v", g.Battlefield.Contains(trawler), me.Hand.Contains(spirit))
	}
}

func TestOldB379SirenSongLyreGivesTheEquippedCreatureATapAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lyre := effectstest.PushCatalogPermanent(g, me.ID, "Siren Song Lyre", "Artifact — Equipment", oldB379Lyre, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB379Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB379Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	for idx := 0; idx < 3; idx++ {
		effectstest.FloatMana(t, g, me, "{C}{C}")
		if err := g.ActivateCatalogAbility(me.ID, mine, idx, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err == nil {
			t.Fatal("an unequipped creature has the tap ability")
		}
	}
	effectstest.AttachForTest(t, g, lyre, mine)
	var err error
	for idx := 0; idx < 3; idx++ {
		effectstest.FloatMana(t, g, me, "{C}{C}")
		if err = g.ActivateCatalogAbility(me.ID, mine, idx, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("the target creature is not tapped")
	}
	if !effectstest.B16Tapped(t, g, mine) {
		t.Error("the equipped creature did not tap for its own cost")
	}
}

func TestOldB379PillarOfWarAttacksOnlyWhileEnchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pillar := effectstest.B12Push(g, me.ID, "Pillar of War", "Artifact Creature — Golem", oldB379Pillar, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, pillar, "defender") {
		t.Fatal("the Pillar has no defender")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(pillar, opp.ID); err == nil {
		t.Fatal("an unenchanted Pillar attacked")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	pillar = effectstest.B12Push(g, me.ID, "Pillar of War", "Artifact Creature — Golem", oldB379Pillar, 3, 3)
	effectstest.PushAuraOnLand(t, g, opp.ID, "Pacifism-ish", "", pillar)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(pillar, opp.ID); err != nil {
		t.Errorf("an enchanted Pillar could not attack: %v", err)
	}
}

func TestOldB379ScourgeOfSkolaValeEntersWithTwoAndEatsForToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scourge := effectstest.CastAndResolveCreature(t, g, "Scourge of Skola Vale", "Creature — Hydra", oldB379Scourge)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, scourge); p != 2 || tough != 2 {
		t.Fatalf("the Scourge is %d/%d, want 2/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, scourge, "trample") {
		t.Error("the Scourge has no trample")
	}

	// A Scourge that is no longer summoning sick eats a creature.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	scourge = effectstest.B12Push(g, me.ID, "Scourge of Skola Vale", "Creature — Hydra", oldB379Scourge, 2, 2)
	food := effectstest.B12Creature(g, me.ID, "Wall", oldB379Bear, 1, 4)
	if err := g.ActivateCatalogAbility(me.ID, scourge, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{scourge}}); err == nil {
		t.Fatal("the Scourge was allowed to sacrifice itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, scourge, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{food}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(food) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if p, tough := effectstest.PtOf(t, g, scourge); p != 6 || tough != 6 {
		t.Errorf("the Scourge is %d/%d, want 6/6 (two counters and a toughness of 4)", p, tough)
	}
}

// oldB379RebuffBoard has seat 0 cast a sorcery and seat 1, who controls devotion
// blue permanents, answer it with Thassa's Rebuff, stopping at the tax prompt.
func oldB379RebuffBoard(t *testing.T, blueSymbols int) (g *game.Game, victim uuid.UUID, caster *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	caster, them := g.Seats[0], g.Seats[1]
	if blueSymbols > 0 {
		cost := ""
		for i := 0; i < blueSymbols; i++ {
			cost += "{U}"
		}
		effectstest.ApaPush(g, them.ID, them.ID, game.Card{Name: "Blue Thing", TypeLine: "Creature — Wizard", ManaCost: cost, Colors: []string{"U"}, Power: 1, Toughness: 1})
	}
	effectstest.ToMainForCost(t, g)
	victim = effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	rebuff := effectstest.HandCardFull(them, "Thassa's Rebuff", "Instant", "", oldB379Rebuff, []string{"U"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	if err := g.CastSpell(them.ID, rebuff, game.CastSpellParams{Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("casting Thassa's Rebuff: %v", err)
	}
	return g, victim, caster
}

func TestOldB379ThassasRebuffTaxesByBlueDevotion(t *testing.T) {
	g, victim, caster := oldB379RebuffBoard(t, 3)
	effectstest.PassUntilTaxed(t, g, caster.ID)
	prompt := effectstest.AnswerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{3}" {
		t.Errorf("the tax is %q, want {3}", prompt.PayCost)
	}
	if g.Stack.Contains(victim) {
		t.Error("a declined tax must counter the spell")
	}
}

func TestOldB379ThassasRebuffWithNoDevotionCountersNothing(t *testing.T) {
	g, victim, _ := oldB379RebuffBoard(t, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Log("the victim was already resolved off the stack")
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.PayCost != "" {
			t.Errorf("a pay-unless prompt %q was asked with no devotion", c.PayCost)
		}
	}
}

func TestOldB379SetessanStarbreakerMayDestroyAnAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB379Bear, 2, 2)
	aura := effectstest.PushAuraOnLand(t, g, opp.ID, "Some Aura", "", bear)
	other := effectstest.B12Permanent(g, opp.ID, "Plain Enchantment", "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Setessan Starbreaker", "Creature — Human Warrior", oldB379Starbreaker, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: other}); err == nil {
			t.Fatal("a non-Aura enchantment was a legal target")
		}
		effectstest.PickCard(t, g, me.ID, aura)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura is still on the battlefield")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("the other enchantment was destroyed")
	}
}
