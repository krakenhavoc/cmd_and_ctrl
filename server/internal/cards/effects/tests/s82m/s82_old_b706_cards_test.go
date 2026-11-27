package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b706_cards_test.go — S82 OLD batch 706 (#5621, Urza's Destiny):
// the Seers, Scents, Thran Foundry, Flicker, Wake of Destruction, Iridescent
// Drake and Storage Matrix. Each test plays the real catalog entry.

const (
	oldB706Foundry   = "88079144-8d8e-4f98-9bee-3762540c94ca"
	oldB706Nightseer = "9bcf66d0-8a4d-4c57-b4de-4cc35399465e"
	oldB706IvySeer   = "f61591d5-0f2a-48cd-a091-863d2e93e019"
	oldB706Flicker   = "19ea326f-45ff-4bc0-a957-7d3f22dd8c6f"
	oldB706Wake      = "562f7f99-ee67-4a14-b5e2-1a4ae710726d"
	oldB706Jasmine   = "3a532c2b-a08c-46ab-86f7-c6dee0d6985f"
	oldB706JasSeer   = "d934a7f0-eb46-4a2c-b0fd-e3071e23da4a"
	oldB706BrineSeer = "64c88c94-3a9d-4fd4-b58c-4d2167c2b8d5"
	oldB706Brine     = "c22b5968-adea-4f74-877d-8ac2ee452769"
	oldB706Rofellos  = "7c309759-e607-4bad-b615-f7c673e508ac"
	oldB706Drake     = "f8cf31db-6c33-4bfb-a24c-f22644ccd377"
	oldB706Matrix    = "03136fe1-17e2-4b1c-860a-dc6373c8fdb5"

	oldB706Pacifism   = "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330"
	oldB706WildGrowth = "706ae742-1807-44b7-a4fa-f2e26f61519a"
)

func TestOldB706EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB706Foundry: "Thran Foundry", oldB706Nightseer: "Nightshade Seer", oldB706IvySeer: "Ivy Seer",
		oldB706Flicker: "Flicker", oldB706Wake: "Wake of Destruction", oldB706Jasmine: "Scent of Jasmine",
		oldB706JasSeer: "Jasmine Seer", oldB706BrineSeer: "Brine Seer", oldB706Brine: "Scent of Brine",
		oldB706Rofellos: "Rofellos's Gift", oldB706Drake: "Iridescent Drake", oldB706Matrix: "Storage Matrix",
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

func oldB706Hand(g *game.Game, p *game.Player, cards ...game.Card) []uuid.UUID {
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		p.Hand.Cards = nil
		for _, c := range cards {
			c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
			p.Hand.PushTop(c)
			ids = append(ids, c.InstanceID)
		}
	})
	return ids
}

func oldB706Colored(name, color string) game.Card {
	return game.Card{Name: name, TypeLine: "Sorcery", Colors: []string{color}}
}

// --- Thran Foundry -----------------------------------------------------------

func TestOldB706ThranFoundryShufflesTheTargetsGraveyardIntoTheirLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foundry := effectstest.PushCatalogPermanent(g, me.ID, "Thran Foundry", "Artifact", oldB706Foundry, false)
	effectstest.PushGraveyardCardTyped(opp, "Old A", "Creature — Bear")
	effectstest.PushGraveyardCardTyped(opp, "Old B", "Sorcery")
	effectstest.PushGraveyardCardTyped(me, "Mine", "Sorcery")
	lib, mine := opp.Library.Size(), me.Graveyard.Size()
	effectstest.FloatForTest(g, me, "C")
	effectstest.Pr7Activate(t, g, me.ID, foundry, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.P8Player(opp.ID)})
	if opp.Graveyard.Size() != 0 || opp.Library.Size() != lib+2 {
		t.Errorf("target's graveyard %d, library %d (was %d): want everything shuffled in", opp.Graveyard.Size(), opp.Library.Size(), lib)
	}
	if me.Graveyard.Size() != mine {
		t.Error("the Foundry's controller's graveyard changed")
	}
	if g.Battlefield.Contains(foundry) {
		t.Error("the Foundry was not exiled as a cost")
	}
}

func TestOldB706ThranFoundryCannotBeActivatedWithoutMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	foundry := effectstest.PushCatalogPermanent(g, me.ID, "Thran Foundry", "Artifact", oldB706Foundry, false)
	if err := g.ActivateCatalogAbility(me.ID, foundry, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.P8Player(me.ID)}); err == nil {
		t.Error("the Foundry was activated with no mana to pay {1}")
	}
	if !g.Battlefield.Contains(foundry) {
		t.Error("a refused activation exiled the Foundry")
	}
}

// --- Flicker -----------------------------------------------------------------

func TestOldB706FlickerBlinksANontokenPermanentAndRefusesATokenAndSendsItHome(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 2}
			}
		}
	})
	token := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Token", "Token Creature — Goblin", "", false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "WWWW")
	if _, err := effectstest.CastFromHandErr(g, me, "Flicker", "Sorcery", "{1}{W}", oldB706Flicker,
		game.CastSpellParams{Strict: true, Targets: effectstest.Pr7bTargets(token)}); err == nil {
		t.Error("Flicker targeted a token")
	}
	effectstest.CastFromHandForTest(t, g, me, "Flicker", "Sorcery", "{1}{W}", oldB706Flicker,
		game.CastSpellParams{Strict: true, Targets: effectstest.Pr7bTargets(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Their Bear" {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatal("the Bear did not come back")
	}
	if back.Controller != opp.ID || back.Owner != opp.ID {
		t.Errorf("the Bear is controlled by %v, want its owner", back.Controller)
	}
	if back.Counters[game.CounterPlusOne] != 0 {
		t.Error("the returned Bear kept its counters: it must be a new object")
	}
}

// --- Wake of Destruction -----------------------------------------------------

func TestOldB706WakeOfDestructionDestroysEveryLandWithTheTargetsName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	f1 := effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	f2 := effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	f3 := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	island := effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	bear := effectstest.B12Creature(g, opp.ID, "Forest Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RRRRRR")
	if _, err := effectstest.CastFromHandErr(g, me, "Wake of Destruction", "Sorcery", "{3}{R}{R}{R}", oldB706Wake,
		game.CastSpellParams{Strict: true, Targets: effectstest.Pr7bTargets(bear)}); err == nil {
		t.Error("Wake of Destruction targeted a creature")
	}
	effectstest.CastFromHandForTest(t, g, me, "Wake of Destruction", "Sorcery", "{3}{R}{R}{R}", oldB706Wake,
		game.CastSpellParams{Strict: true, Targets: effectstest.Pr7bTargets(f1)})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{f1, f2, f3} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a Forest survived Wake of Destruction")
		}
	}
	if !g.Battlefield.Contains(island) {
		t.Error("the Island was destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a nonland permanent was destroyed")
	}
}

// --- Iridescent Drake --------------------------------------------------------

func oldB706CastDrake(t *testing.T, g *game.Game, me *game.Player, pickAura uuid.UUID) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UUUU")
	drake := effectstest.CastFromHandForTest(t, g, me, "Iridescent Drake", "Creature — Drake", "{3}{U}", oldB706Drake, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if pickAura != uuid.Nil {
		if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
			effectstest.PickCard(t, g, me.ID, pickAura)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	return drake
}

func TestOldB706IridescentDrakePutsAnAuraFromAnyGraveyardOntoItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	aura := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: aura, Name: "Pacifism", OracleID: oldB706Pacifism, TypeLine: "Enchantment — Aura",
			ManaCost: "{1}{W}", Owner: opp.ID, Controller: opp.ID})
	})
	drake := oldB706CastDrake(t, g, me, aura)
	c, ok := effectstest.BattlefieldCard(g, aura)
	if !ok {
		t.Fatal("the Aura was not put onto the battlefield")
	}
	if c.Controller != me.ID {
		t.Errorf("the Aura is controlled by %v, want the Drake's controller", c.Controller)
	}
	if !c.IsAttachedTo(drake) {
		t.Error("the Aura is not attached to the Drake")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, drake, "flying") {
		t.Error("the Drake lacks flying")
	}
}

func TestOldB706IridescentDrakeLeavesAnAuraThatCannotEnchantItInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aura := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: aura, Name: "Wild Growth", OracleID: oldB706WildGrowth, TypeLine: "Enchantment — Aura",
			ManaCost: "{G}", Owner: me.ID, Controller: me.ID})
	})
	oldB706CastDrake(t, g, me, aura)
	if g.Battlefield.Contains(aura) {
		t.Error("an Aura that enchants only a land was put onto the Drake")
	}
	if !me.Graveyard.Contains(aura) {
		t.Error("the Aura left the graveyard")
	}
}

// --- Storage Matrix ----------------------------------------------------------

func TestOldB706StorageMatrixLetsAPlayerUntapOnlyTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	matrix := effectstest.PushCatalogPermanent(g, me.ID, "Storage Matrix", "Artifact", oldB706Matrix, false)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{land, bear, rock} {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Tapped = true
				}
			}
		}
	})
	asked := 0
	for i := 0; i < 400 && asked == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		for _, c := range g.PendingChoices {
			if c == nil || c.Kind != game.PendingChoiceOptionPick || c.Chooser != me.ID {
				continue
			}
			idx := -1
			for j, o := range c.PickOptions {
				if o.Label == "Land" {
					idx = j
				}
			}
			if idx < 0 {
				continue
			}
			if len(c.PickOptions) != 3 {
				t.Fatalf("options = %+v, want artifact, creature, land", c.PickOptions)
			}
			if err := g.ResolveOptionPick(c.ID, me.ID, idx); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			asked++
			break
		}
		if asked == 0 {
			// An opponent's untap step is asked the same question: answer it with the first option.
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser != me.ID {
					if err := g.ResolveOptionPick(c.ID, c.Chooser, 0); err != nil {
						t.Fatalf("opponent's pick: %v", err)
					}
				}
			}
		}
	}
	if asked == 0 {
		t.Fatal("the controller was never asked at their untap step")
	}
	tapped := func(id uuid.UUID) bool { c, _ := effectstest.BattlefieldCard(g, id); return c.Tapped }
	if tapped(land) {
		t.Error("the chosen type (land) did not untap")
	}
	if !tapped(bear) || !tapped(rock) {
		t.Error("a permanent of another type untapped")
	}
	_ = matrix
}

// --- Scent of Jasmine and Jasmine Seer ---------------------------------------

func TestOldB706ScentOfJasmineGainsTwoLifePerWhiteCardRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := oldB706Hand(g, me, oldB706Colored("White A", "W"), oldB706Colored("White B", "W"), oldB706Colored("White C", "W"), oldB706Colored("Red D", "R"))
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "W")
	life := me.Life
	effectstest.CastFromHandForTest(t, g, me, "Scent of Jasmine", "Instant", "{W}", oldB706Jasmine, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.ChooseCardsPrompt(t, g).ChooseCards); n != 3 {
		t.Fatalf("offered %d cards to reveal, want the 3 white ones", n)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0], ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life = %d, want %d after revealing two white cards", me.Life, life+4)
	}
}

func TestOldB706JasmineSeerActivatesForTheSameLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seer := effectstest.PushCatalogPermanent(g, me.ID, "Jasmine Seer", "Creature — Human Wizard", oldB706JasSeer, false)
	ids := oldB706Hand(g, me, oldB706Colored("White A", "W"), oldB706Colored("White B", "W"))
	effectstest.FloatForTest(g, me, "WWW")
	life := me.Life
	effectstest.Pr7Activate(t, g, me.ID, seer, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.AnswerChooseCards(t, g, me.ID, ids...)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life = %d, want %d", me.Life, life+4)
	}
	if c, _ := effectstest.BattlefieldCard(g, seer); !c.Tapped {
		t.Error("the Seer was not tapped for the cost")
	}
}

func TestOldB706JasmineSeerWithNoWhiteCardsAsksNothingAndGainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seer := effectstest.PushCatalogPermanent(g, me.ID, "Jasmine Seer", "Creature — Human Wizard", oldB706JasSeer, false)
	oldB706Hand(g, me, oldB706Colored("Red A", "R"))
	effectstest.FloatForTest(g, me, "WWW")
	life := me.Life
	effectstest.Pr7Activate(t, g, me.ID, seer, 0, game.ActivateAbilityParams{Strict: true})
	if effectstest.PendingOfKind(g, game.PendingChoiceChooseCards) != nil {
		t.Error("asked to reveal with no white card in hand")
	}
	if me.Life != life {
		t.Errorf("life = %d, want %d", me.Life, life)
	}
}

// --- Nightshade Seer and Ivy Seer --------------------------------------------

func TestOldB706NightshadeSeerShrinksByTheBlackCardsRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seer := effectstest.PushCatalogPermanent(g, me.ID, "Nightshade Seer", "Creature — Human Wizard", oldB706Nightseer, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 4, 4)
	ids := oldB706Hand(g, me, oldB706Colored("Black A", "B"), oldB706Colored("Black B", "B"), oldB706Colored("Red C", "R"))
	effectstest.FloatForTest(g, me, "BBB")
	effectstest.Pr7Activate(t, g, me.ID, seer, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.Pr7bTargets(victim)})
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0], ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, victim); p != 2 || tough != 2 {
		t.Errorf("Victim is %d/%d, want 2/2 after revealing two black cards", p, tough)
	}
}

func TestOldB706IvySeerPumpsByTheGreenCardsRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seer := effectstest.PushCatalogPermanent(g, me.ID, "Ivy Seer", "Creature — Elf Wizard", oldB706IvySeer, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ids := oldB706Hand(g, me, oldB706Colored("Green A", "G"), oldB706Colored("Green B", "G"), oldB706Colored("Green C", "G"))
	effectstest.FloatForTest(g, me, "GGG")
	effectstest.Pr7Activate(t, g, me.ID, seer, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.Pr7bTargets(bear)})
	effectstest.AnswerChooseCards(t, g, me.ID, ids...)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("Bear is %d/%d, want 5/5 after revealing three green cards", p, tough)
	}
}

// --- Scent of Brine and Brine Seer -------------------------------------------

func TestOldB706ScentOfBrineCountersUnlessThePlayerPaysOnePerBlueCardRevealed(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ids := oldB706Hand(g, opp, oldB706Colored("Blue A", "U"), oldB706Colored("Blue B", "U"))
		effectstest.ToMainForCost(t, g)
		victim := effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		effectstest.CastFromSeat(t, g, opp, "Scent of Brine", "Instant", oldB706Brine, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, opp.ID, ids...)
		if pay {
			effectstest.B06AddMana(me, "C", "C")
		}
		effectstest.PassUntilTaxed(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		_, resolved := effectstest.BattlefieldCard(g, victim)
		if pay && !resolved {
			t.Error("the spell was countered although its controller paid {2}")
		}
		if !pay && (resolved || !me.Graveyard.Contains(victim)) {
			t.Error("the spell was not countered")
		}
	}
}

func TestOldB706BrineSeerCountersUnlessThePlayerPays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seer := effectstest.PushCatalogPermanent(g, opp.ID, "Brine Seer", "Creature — Human Wizard", oldB706BrineSeer, false)
	ids := oldB706Hand(g, opp, oldB706Colored("Blue A", "U"))
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.FloatForTest(g, opp, "UUU")
	if err := g.ActivateCatalogAbility(opp.ID, seer, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("activate Brine Seer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, ids...)
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if _, resolved := effectstest.BattlefieldCard(g, victim); resolved || !me.Graveyard.Contains(victim) {
		t.Error("the spell was not countered")
	}
}

// --- Rofellos's Gift ---------------------------------------------------------

func TestOldB706RofellossGiftReturnsOneEnchantmentPerGreenCardRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	e1 := effectstest.PushGraveyardCardTyped(me, "Ench One", "Enchantment")
	e2 := effectstest.PushGraveyardCardTyped(me, "Ench Two", "Enchantment — Aura")
	creature := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	ids := oldB706Hand(g, me, oldB706Colored("Green A", "G"), oldB706Colored("Green B", "G"))
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "G")
	hand := me.Hand.Size()
	effectstest.CastFromHandForTest(t, g, me, "Rofellos's Gift", "Sorcery", "{G}", oldB706Rofellos, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0], ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsPrompt(t, g)
	if len(prompt.ChooseCards) != 2 {
		t.Fatalf("offered %d graveyard cards, want the 2 enchantments", len(prompt.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, e1, e2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand = %d, want %d (the Gift left, two enchantments came back)", me.Hand.Size(), hand+2)
	}
	if me.Graveyard.Contains(e1) || me.Graveyard.Contains(e2) {
		t.Error("an enchantment stayed in the graveyard")
	}
	if !me.Graveyard.Contains(creature) {
		t.Error("the creature card left the graveyard")
	}
}
