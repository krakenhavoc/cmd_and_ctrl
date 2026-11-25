package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b020_cards_test.go — Secret Lair Drop: Creatures with keywords (S82
// older sets, batch 020, #5462). Each test plays the real catalog entry.

const (
	oldB20Smasher   = "17c651c7-da4e-45e9-9a64-fad3e873123d"
	oldB20Towershel = "45154dc6-ff0a-4e99-8528-689757d1778e"
	oldB20Fleet     = "52d349e7-4bd2-468a-ab26-df8cd262f6f6"
	oldB20Kira      = "60808c8f-5257-451b-a9b6-389ba51598d5"
	oldB20Dawnbreak = "6fa7229d-18b6-43d8-aa57-a56cdb815767"
	oldB20Feather   = "aa219936-661b-4ccb-8741-78b70cff2b1a"
	oldB20Amy       = "b2d597b4-f9ba-44b0-a947-0c2ecda2d308"
	oldB20Vaevictis = "c42ce2e9-4dfe-468a-8f41-49e187cb91d4"
	oldB20Nathan    = "ed69b907-272c-4a87-a943-06aaad976fb1"
)

func TestOldB20EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB20Smasher:   "Reality Smasher",
		oldB20Towershel: "Meandering Towershell",
		oldB20Fleet:     "Fleet Swallower",
		oldB20Kira:      "Kira, Great Glass-Spinner",
		oldB20Dawnbreak: "Dawnbreak Reclaimer",
		oldB20Feather:   "Feather, the Redeemed",
		oldB20Amy:       "Amy Rose",
		oldB20Vaevictis: "Vaevictis Asmadi, the Dire",
		oldB20Nathan:    "Nathan Drake, Treasure Hunter",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB20Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tgh int) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, name, typeLine, oracle, p, tgh)
}

func TestOldB20RealitySmasherCountersASpellUnlessItsControllerDiscards(t *testing.T) {
	for _, tc := range []struct {
		name        string
		hand        int
		discard     bool
		wantCounter bool
	}{
		{"discarding saves the spell", 2, true, false},
		{"declining counters it", 2, false, true},
		{"no card to discard counters it", 0, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			smasher := oldB20Push(g, me.ID, "Reality Smasher", "Creature — Eldrazi", oldB20Smasher, 5, 5)
			opp.Hand.Cards = nil
			var keep uuid.UUID
			for i := 0; i < tc.hand; i++ {
				c := game.NewCard("Chaff", opp.ID)
				opp.Hand.PushTop(c)
				keep = c.InstanceID
			}
			spell := effectstest.Batch01OpponentCasts(t, g, opp, "Zap", effectstest.LightningBoltOracle, "{0}",
				[]game.TargetRef{{Kind: game.TargetCard, ID: smasher}})
			effectstest.PassPriorityAroundTable(t, g)
			if tc.hand > 0 {
				if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
					t.Fatal("the spell's controller was not asked to discard")
				}
				if tc.discard {
					effectstest.AnswerDiscard(t, g, opp.ID, keep)
				} else {
					effectstest.AnswerDiscard(t, g, opp.ID)
				}
				effectstest.PassPriorityAroundTable(t, g)
			}
			countered := opp.Graveyard.Contains(spell) && effectstest.FindBattlefieldCardByID(g, smasher).DamageMarked == 0
			if countered != tc.wantCounter {
				t.Errorf("spell countered = %v, want %v (damage on Smasher %d)", countered, tc.wantCounter,
					effectstest.FindBattlefieldCardByID(g, smasher).DamageMarked)
			}
		})
	}
}

func TestOldB20RealitySmasherIgnoresItsOwnersSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	smasher := oldB20Push(g, me.ID, "Reality Smasher", "Creature — Eldrazi", oldB20Smasher, 5, 5)
	me.Hand.PushTop(game.NewCard("Chaff", me.ID))
	effectstest.CastCatalogSpell(t, g, "Zap", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: smasher}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Fatal("its own controller's spell was asked to discard")
	}
	if effectstest.FindBattlefieldCardByID(g, smasher).DamageMarked != 3 {
		t.Error("the spell was countered")
	}
}

func TestOldB20FleetSwallowerMillsHalfRoundedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fleet := oldB20Push(g, me.ID, "Fleet Swallower", "Creature — Fish", oldB20Fleet, 6, 6)
	opp.Library.Cards = nil
	for i := 0; i < 41; i++ {
		opp.Library.PushTop(game.NewCard("Filler", opp.ID))
	}
	g.WithWriteLock(func() { g.Turn.Step = game.StepBeginCombat })
	effectstest.DeclareAttack(t, g, opp.ID, fleet)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != 20 {
		t.Errorf("library = %d, want 20 (41 minus half rounded up, 21)", got)
	}
	if got := len(opp.Graveyard.Cards); got != 21 {
		t.Errorf("graveyard = %d, want 21", got)
	}
}

func TestOldB20KiraCountersOnlyTheFirstSpellOrAbilityEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kira := oldB20Push(g, me.ID, "Kira, Great Glass-Spinner", "Legendary Creature — Spirit", oldB20Kira, 2, 2)
	bear := oldB20Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 9)
	g.BumpLayerVersionForTest()
	first := effectstest.Batch01OpponentCasts(t, g, opp, "Zap", effectstest.LightningBoltOracle, "{0}",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(first) || effectstest.FindBattlefieldCardByID(g, bear).DamageMarked != 0 {
		t.Fatalf("the first spell was not countered (damage %d)", effectstest.FindBattlefieldCardByID(g, bear).DamageMarked)
	}
	second := effectstest.Batch01OpponentCasts(t, g, opp, "Zap", effectstest.LightningBoltOracle, "{0}",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardByID(g, bear).DamageMarked != 3 {
		t.Errorf("the second spell was countered too (damage %d)", effectstest.FindBattlefieldCardByID(g, bear).DamageMarked)
	}
	_ = second
	// Another creature of mine gets its own first targeting, Kira herself too.
	third := effectstest.Batch01OpponentCasts(t, g, opp, "Zap", effectstest.LightningBoltOracle, "{0}",
		[]game.TargetRef{{Kind: game.TargetCard, ID: kira}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(third) {
		t.Error("Kira's own first targeting was not countered")
	}
}

func TestOldB20FeatherReturnsATargetingSpellToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB20Push(g, me.ID, "Feather, the Redeemed", "Legendary Creature — Angel", oldB20Feather, 3, 4)
	bear := oldB20Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 9)
	spell := effectstest.CastCatalogSpell(t, g, "Zap", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Fatal("the spell did not go to exile as it resolved")
	}
	effectstest.AdvanceThroughEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(spell) {
		t.Error("the spell did not return to hand at the next end step")
	}
}

func TestOldB20FeatherIgnoresASpellThatDoesNotTargetYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB20Push(g, me.ID, "Feather, the Redeemed", "Legendary Creature — Angel", oldB20Feather, 3, 4)
	theirs := oldB20Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 2, 9)
	spell := effectstest.CastCatalogSpell(t, g, "Zap", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(spell) {
		t.Error("a spell aimed at an opponent's creature was exiled")
	}
	spell2 := effectstest.CastCatalogSpell(t, g, "Zap", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(spell2) {
		t.Error("a spell aimed at a player was exiled")
	}
}

func TestOldB20FeatherCounteredSpellGoesToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB20Push(g, me.ID, "Feather, the Redeemed", "Legendary Creature — Angel", oldB20Feather, 3, 4)
	bear := oldB20Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 9)
	spell := effectstest.CastCatalogSpell(t, g, "Zap", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	g.WithWriteLock(func() { _ = g.CounterTargetForEffect(spell) })
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(spell) || g.Exile.Contains(spell) {
		t.Error("a countered spell did not go to the graveyard")
	}
}

func TestOldB20AmyRoseAttachesAnEquipmentAndPumpsAnotherAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	amy := oldB20Push(g, me.ID, "Amy Rose", "Legendary Creature — Hedgehog Warrior", oldB20Amy, 3, 3)
	friend := oldB20Push(g, me.ID, "Friend", "Creature — Bear", "", 2, 2)
	equip := effectstest.SeedEquipment(g, me.ID, "Plain Sword", "")
	effectstest.DeclareAttack(t, g, opp.ID, amy, friend)
	effectstest.CloudDeckPickTargets(t, g, me.ID, equip)
	effectstest.CloudDeckPickTargets(t, g, me.ID, friend)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, equip); c.AttachedTo.ID != amy {
		t.Errorf("the Equipment is attached to %v, want Amy Rose", c.AttachedTo.ID)
	}
	if got := effectstest.LayeredCard(t, g, friend).CurrentPower(); got != 5 {
		t.Errorf("the other attacker's power = %d, want 5 (2 +3)", got)
	}
	if got := effectstest.LayeredCard(t, g, friend).CurrentToughness(); got != 2 {
		t.Errorf("toughness = %d, want 2: the bonus is +X/+0", got)
	}
}

func TestOldB20DawnbreakReclaimerReturnsOneCreatureFromEachGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB20Push(g, me.ID, "Dawnbreak Reclaimer", "Creature — Angel", oldB20Dawnbreak, 5, 5)
	theirs := effectstest.MkmMP2Grave(opp, "Their Ogre", "Creature — Ogre")
	effectstest.MkmMP2Grave(opp, "Their Island", "Basic Land — Island")
	mine := effectstest.MkmMP2Grave(me, "My Bear", "Creature — Bear")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("the controller was not asked to choose")
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != theirs {
		t.Fatalf("the first choice offered %v, want only the opponent's creature card", c.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, theirs)
	c = effectstest.LatestChooseCardsFor(g, opp.ID)
	if c == nil {
		t.Fatal("the opponent was not asked to choose")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, mine)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	for id, owner := range map[uuid.UUID]uuid.UUID{theirs: opp.ID, mine: me.ID} {
		card := effectstest.FindBattlefieldCardByID(g, id)
		if card == nil {
			t.Fatalf("%s did not return", id)
		}
		if card.Controller != owner {
			t.Errorf("%s is controlled by %v, want its owner %v", card.Name, card.Controller, owner)
		}
	}
}

func TestOldB20DawnbreakReclaimerMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB20Push(g, me.ID, "Dawnbreak Reclaimer", "Creature — Angel", oldB20Dawnbreak, 5, 5)
	theirs := effectstest.MkmMP2Grave(opp, "Their Ogre", "Creature — Ogre")
	mine := effectstest.MkmMP2Grave(me, "My Bear", "Creature — Bear")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, theirs)
	effectstest.AnswerChooseCards(t, g, opp.ID, mine)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	if effectstest.FindBattlefieldCardByID(g, theirs) != nil || effectstest.FindBattlefieldCardByID(g, mine) != nil {
		t.Error("a card returned after the controller declined")
	}
}

func TestOldB20VaevictisSacrificesOnePermanentPerPlayerAndRevealsForEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	vae := oldB20Push(g, me.ID, "Vaevictis Asmadi, the Dire", "Legendary Creature — Elder Dragon", oldB20Vaevictis, 6, 6)
	var victims []uuid.UUID
	victims = append(victims, oldB20Push(g, me.ID, "My Rock", "Artifact", "", 0, 0))
	for _, p := range g.Seats[1:] {
		victims = append(victims, oldB20Push(g, p.ID, "Their Rock", "Artifact", "", 0, 0))
	}
	// A permanent card on top of each library, except the second seat's land... all permanents.
	tops := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		top := game.NewCard("Revealed Bear", p.ID)
		top.TypeLine = "Creature — Bear"
		p.Library.PushTop(top)
		tops[p.ID] = top.InstanceID
	}
	// The third seat's top card is an instant: revealed, not put onto the battlefield.
	g.Seats[2].Library.Cards = g.Seats[2].Library.Cards[:len(g.Seats[2].Library.Cards)-1]
	instant := game.NewCard("Revealed Instant", g.Seats[2].ID)
	instant.TypeLine = "Instant"
	g.Seats[2].Library.PushTop(instant)
	tops[g.Seats[2].ID] = instant.InstanceID
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, vae)
	for i, v := range victims {
		p := g.Seats[i]
		if i == 0 {
			effectstest.CloudDeckPickTargets(t, g, me.ID, v)
		} else {
			effectstest.CloudDeckPickTargets(t, g, me.ID, v)
		}
		_ = p
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, v := range victims {
		if g.Battlefield.Contains(v) {
			t.Errorf("%s was not sacrificed", v)
		}
	}
	for _, p := range g.Seats {
		top := tops[p.ID]
		onBattlefield := g.Battlefield.Contains(top)
		if want := p != g.Seats[2]; onBattlefield != want {
			t.Errorf("%s's revealed card on the battlefield = %v, want %v", p.Name, onBattlefield, want)
		}
		if c := effectstest.FindBattlefieldCardByID(g, top); c != nil && c.Controller != p.ID {
			t.Errorf("%s's revealed card is under %v", p.Name, c.Controller)
		}
	}
}

func TestOldB20NathanDrakeCastsAnExiledSpellWithAnyColorOfMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	nathan := oldB20Push(g, me.ID, "Nathan Drake, Treasure Hunter", "Legendary Creature — Human Rogue", oldB20Nathan, 3, 2)
	// Their top card is a {R} instant; mine a land.
	their := game.NewCard("Their Shock", opp.ID)
	their.TypeLine = "Instant"
	their.ManaCost = "{R}"
	opp.Library.PushTop(their)
	land := game.NewCard("My Island", me.ID)
	land.TypeLine = "Basic Land — Island"
	me.Library.PushTop(land)
	effectstest.DeclareAttack(t, g, opp.ID, nathan)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(their.InstanceID) || !g.Exile.Contains(land.InstanceID) {
		t.Fatal("the top cards were not exiled")
	}
	prompt := effectstest.KeywordCastPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no cast offer")
	}
	// A blue mana pays {R} for a spell I don't own.
	effectstest.FloatMana(t, g, me, "{U}")
	effectstest.AnswerMayCastPrompt(t, g, me.ID, true)
	if !g.Stack.Contains(their.InstanceID) {
		t.Error("the opponent's spell was not cast with blue mana")
	}
}

func TestOldB20NathanDrakeDoesNotBendManaForSpellsYouOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB20Push(g, me.ID, "Nathan Drake, Treasure Hunter", "Legendary Creature — Human Rogue", oldB20Nathan, 3, 2)
	own := effectstest.HandCardFull(me, "My Shock", "Instant", "{R}", effectstest.LightningBoltOracle, []string{"R"})
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, own, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}}); err == nil {
		t.Fatal("blue mana paid {R} for a spell I own")
	}
}

func TestOldB20MeanderingTowershellExilesOnAttackAndReturnsAttackingOnYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	shell := oldB20Push(g, me.ID, "Meandering Towershell", "Creature — Turtle", oldB20Towershel, 5, 9)
	if !effectstest.HasKeywordOnBattlefield(t, g, shell, "islandwalk") {
		t.Error("no islandwalk")
	}
	effectstest.DeclareAttack(t, g, opp.ID, shell)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(shell) || g.Battlefield.Contains(shell) {
		t.Fatal("the Towershell was not exiled by its attack trigger")
	}
	// The other players' turns do not return it.
	for i := 0; i < 3; i++ {
		effectstest.EndTurn(t, g)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Exile.Contains(shell) {
			t.Fatalf("returned on another player's turn %d", i+1)
		}
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, 0, false, "my next turn")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.PassPriorityAroundTable(t, g)
	pc := effectstest.LatestOptionPickFor(g, me.ID)
	if pc == nil {
		t.Fatal("no prompt for what it attacks")
	}
	if err := g.ResolveOptionPick(pc.ID, me.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var c *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Meandering Towershell" {
			c = &g.Battlefield.Cards[i]
		}
	}
	if c == nil {
		t.Fatal("the Towershell did not return to the battlefield")
	}
	if !c.Tapped || c.AttackingTarget == uuid.Nil || c.Controller != me.ID {
		t.Errorf("returned tapped=%v attacking=%v controller=%v, want tapped and attacking under my control", c.Tapped, c.AttackingTarget, c.Controller)
	}
}

func TestOldB20DawnbreakReclaimerAsksWhichOpponentsGraveyardWhenSeveralHaveCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	oldB20Push(g, me.ID, "Dawnbreak Reclaimer", "Creature — Angel", oldB20Dawnbreak, 5, 5)
	effectstest.MkmMP2Grave(opp, "Ogre", "Creature — Ogre")
	pick := effectstest.MkmMP2Grave(third, "Elf", "Creature — Elf")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, third)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil || len(c.ChooseCards) != 1 || c.ChooseCards[0] != pick {
		t.Fatalf("the card choice = %+v, want only the third player's Elf", c)
	}
}

// Nathan Drake's second half: mana bends for an activation of a permanent
// its controller controls but doesn't own, and only that.
func TestOldB20NathanDrakeCoversActivationsOfPermanentsYouDontOwn(t *testing.T) {
	spec, ok := effects.Lookup(oldB20Nathan)
	if !ok || len(spec.AnyColorSpend) != 1 {
		t.Fatalf("Nathan Drake's any-colour grant: %v %d", ok, len(spec.AnyColorSpend))
	}
	covers := spec.AnyColorSpend[0].Covers
	me, opp := uuid.New(), uuid.New()
	stolen := game.Card{InstanceID: uuid.New(), Name: "Stolen", TypeLine: "Artifact", Owner: opp, Controller: me}
	own := game.Card{InstanceID: uuid.New(), Name: "Own", TypeLine: "Artifact", Owner: me, Controller: me}
	given := game.Card{InstanceID: uuid.New(), Name: "Given", TypeLine: "Artifact", Owner: me, Controller: opp}
	for _, tc := range []struct {
		name  string
		card  game.Card
		payer uuid.UUID
		want  bool
	}{
		{"a permanent I control but don't own", stolen, me, true},
		{"a permanent I own and control", own, me, false},
		{"a permanent an opponent controls", given, me, false},
	} {
		ctx := game.ManaSpendForAbility(tc.card).PaidBy(tc.payer)
		if got := covers(ctx); got != tc.want {
			t.Errorf("%s: covered = %v, want %v", tc.name, got, tc.want)
		}
	}
	// A spell cast: covered only when the payer doesn't own it.
	foreign := game.ManaSpendContext{Purpose: game.SpendPurposeCast, SpellOwner: opp, Payer: me}
	owned := game.ManaSpendContext{Purpose: game.SpendPurposeCast, SpellOwner: me, Payer: me}
	if !covers(foreign) || covers(owned) {
		t.Errorf("spell casts: foreign %v owned %v, want true false", covers(foreign), covers(owned))
	}
}
