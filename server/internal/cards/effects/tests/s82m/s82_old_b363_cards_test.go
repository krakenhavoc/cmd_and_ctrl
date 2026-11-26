package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 363 (Conspiracy, #5557): the dethrone Rogues, Basandra,
// Heckling Fiends, Wood Sage and Wind Dancer.

type oldB363Card struct{ oracle, name, typeLine string }

var oldB363Cards = []oldB363Card{
	{"7a67cb39-6b53-4f2b-a52c-3861119eba12", "Marchesa's Emissary", "Creature — Human Rogue"},
	{"9a133991-8a7a-4bfd-b0ae-d76433450c03", "Grenzo's Cutthroat", "Creature — Goblin Rogue"},
	{"9f05fab1-398a-4bd1-824d-6b1373dc9a94", "Marchesa's Infiltrator", "Creature — Human Rogue"},
	{"c3e11b61-896e-4a6b-9b33-286313d4a5e2", "Marchesa's Smuggler", "Creature — Human Rogue"},
	{"f4b8e856-17a9-44db-bce6-312a6bd2af32", "Enraged Revolutionary", "Creature — Human Warrior"},
	{"4a480736-4cdc-4e3a-a8cb-d1c726982f9f", "Basandra, Battle Seraph", "Legendary Creature — Angel"},
	{"4b8fb8dc-f1c5-4994-8a58-4849e7e85539", "Heckling Fiends", "Creature — Devil"},
	{"6fde7bb3-2539-411f-8733-36ce9595dfd5", "Wood Sage", "Creature — Human Druid"},
	{"e9903d75-f71d-45d0-ac5f-955c0c227e02", "Wind Dancer", "Creature — Faerie"},
}

const (
	oldB363Emissary   = "7a67cb39-6b53-4f2b-a52c-3861119eba12"
	oldB363Cutthroat  = "9a133991-8a7a-4bfd-b0ae-d76433450c03"
	oldB363Infiltrate = "9f05fab1-398a-4bd1-824d-6b1373dc9a94"
	oldB363Smuggler   = "c3e11b61-896e-4a6b-9b33-286313d4a5e2"
	oldB363Basandra   = "4a480736-4cdc-4e3a-a8cb-d1c726982f9f"
	oldB363Fiends     = "4b8fb8dc-f1c5-4994-8a58-4849e7e85539"
	oldB363WoodSage   = "6fde7bb3-2539-411f-8733-36ce9595dfd5"
	oldB363WindDancer = "e9903d75-f71d-45d0-ac5f-955c0c227e02"
)

func TestOldB363EveryCardIsRegisteredFull(t *testing.T) {
	for _, c := range oldB363Cards {
		spec, ok := effects.Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		if spec.Name != c.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", c.name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB363DethroneTriggersOnlyAgainstTheLifeLeader(t *testing.T) {
	for _, c := range oldB363Cards {
		switch c.oracle {
		case oldB363Basandra, oldB363Fiends, oldB363WoodSage, oldB363WindDancer:
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := effectstest.PushCatalogPermanent(g, me.ID, c.name, c.typeLine, c.oracle, false)
			effectstest.DeclareAttack(t, g, opp.ID, id) // everyone is tied for the most life
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.PlusOnes(g, id); n != 1 {
				t.Errorf("dethrone put %d counters at the life leader, want 1", n)
			}
		})
		t.Run(c.name+" trailing", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := effectstest.PushCatalogPermanent(g, me.ID, c.name, c.typeLine, c.oracle, false)
			opp.Life = 5
			effectstest.DeclareAttack(t, g, opp.ID, id)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.PlusOnes(g, id); n != 0 {
				t.Errorf("dethrone put %d counters against a trailing player, want 0", n)
			}
		})
	}
}

func TestOldB363PrintedKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for oracle, kw := range map[string]string{
		oldB363Emissary: "hexproof", oldB363Cutthroat: "first strike", oldB363Basandra: "flying", oldB363WindDancer: "flying",
	} {
		id := effectstest.PushCatalogPermanent(g, me.ID, "Test", "Creature — Test", oracle, false)
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("%s lacks %s", oracle, kw)
		}
	}
}

func TestOldB363InfiltratorDrawsOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AuraSeedLibrary(me, 3)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Marchesa's Infiltrator", "Creature — Human Rogue", oldB363Infiltrate, false)
	hand := len(me.Hand.Cards)
	effectstest.DealCombatDamageToPlayer(g, id, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, len(me.Hand.Cards))
	}
}

func TestOldB363SmugglerGivesHasteAndMakesATargetUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	smuggler := effectstest.PushCatalogPermanent(g, me.ID, "Marchesa's Smuggler", "Creature — Human Rogue", oldB363Smuggler, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FillPool(me, 3)
	if err := effectstest.InvB1Activate(g, me.ID, smuggler, 0, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Error("targeted a creature the controller doesn't control")
	}
	effectstest.FillPoolColored(me, "U", 1)
	effectstest.FillPoolColored(me, "R", 1)
	if err := effectstest.InvB1Activate(g, me.ID, smuggler, 0, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the bear did not gain haste")
	}
	if effectstest.AuraRestrictions(t, g, bear)&game.CantBeBlocked == 0 {
		t.Error("the bear can be blocked")
	}
}

func TestOldB363WindDancerGrantsFlyingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dancer := effectstest.PushCatalogPermanent(g, me.ID, "Wind Dancer", "Creature — Faerie", oldB363WindDancer, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := effectstest.InvB1Activate(g, me.ID, dancer, 0, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the bear did not gain flying")
	}
	if err := effectstest.InvB1Activate(g, me.ID, dancer, 0, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Error("activated again while tapped")
	}
}

func TestOldB363ForcedAttackFromBasandraAndTheFiends(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, mana string
	}{
		{"Basandra, Battle Seraph", oldB363Basandra, "Legendary Creature — Angel", "R"},
		{"Heckling Fiends", oldB363Fiends, "Creature — Devil", "RRR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me := g.Seats[seat]
			src := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			bear := effectstest.SeedCreature(g, "Grizzly Bears", me.ID)
			effectstest.FillPoolColored(me, "R", len(tc.mana))
			if err := effectstest.InvB1Activate(g, me.ID, src, 0, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
			var re *game.AttackRequirementError
			if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != bear {
				t.Errorf("pass with the target at home = %v, want a refusal naming it", err)
			}
		})
	}
}

func TestOldB363BasandraStopsEveryPlayerCastingSpellsDuringCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Basandra, Battle Seraph", "Legendary Creature — Angel", oldB363Basandra, false)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.PlayerTarget(g.Seats[1].ID)}); err != nil {
		t.Fatalf("a spell in the main phase was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.PlayerTarget(g.Seats[1].ID)}); err == nil {
		t.Error("a spell was cast during combat with Basandra on the battlefield")
	}
}

func TestOldB363WoodSageTakesEveryCreatureCardWithTheNameAndGravesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sage := effectstest.PushCatalogPermanent(g, me.ID, "Wood Sage", "Creature — Human Druid", oldB363WoodSage, false)
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		game.Card{Name: "Runeclaw Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deeper", TypeLine: "Creature — Bear"},
	)
	hand, grave := len(me.Hand.Cards), len(me.Graveyard.Cards)
	if err := effectstest.InvB1Activate(g, me.ID, sage, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceCardName && c.Chooser == me.ID {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatal("no card name prompt")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Grizzly Bears"); err != nil {
		t.Fatalf("name: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - hand; got != 2 {
		t.Errorf("%d cards to hand, want the two Grizzly Bears", got)
	}
	if got := len(me.Graveyard.Cards) - grave; got != 2 {
		t.Errorf("%d cards to the graveyard, want the other two revealed", got)
	}
	if len(me.Library.Cards) != 1 {
		t.Errorf("the library has %d cards, want the one beyond the top four", len(me.Library.Cards))
	}
	_ = uuid.Nil
}
