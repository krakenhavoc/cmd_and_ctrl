package testsso

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func speedSliceARestricted(t *testing.T, g *game.Game, id uuid.UUID, r game.Restriction) bool {
	t.Helper()
	var restricted, found bool
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				restricted, found = game.Restricted(&g.Battlefield.Cards[i], r), true
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return restricted
}

func speedSliceAOpponent(g *game.Game) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+effectstest.SpeedSliceAOpponentSeatIdx)%len(g.Seats)]
}

func TestSpeedSliceACardsDeclareStartYourEngines(t *testing.T) {
	for _, oracle := range []string{
		effectstest.HowlsquadHeavyOracle, effectstest.MendicantCoreOracle, effectstest.HazoretGodseekerOracle, effectstest.MomentumBreakerOracle,
		effectstest.NestingBotOracle, effectstest.KickoffCelebrationsOracle, effectstest.ZahurGlorysPastOracle, effectstest.GasGuzzlerOracle,
		effectstest.EndriderCatalyzerOracle, effectstest.FarFortuneEndBossOracle, effectstest.SamutDrivingForceOracle, effectstest.PointTheWayOracle,
		effectstest.HourOfVictoryOracle, effectstest.GastalThrillseekerOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		found := false
		for _, kw := range spec.PrintedKeywords {
			if kw == game.KeywordStartYourEngines {
				found = true
			}
		}
		if !found {
			t.Errorf("%s does not declare start your engines!", spec.Name)
		}
	}
}

func TestGastalThrillseekerPingsAnOpponentAndGainsLife(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	opp := speedSliceAOpponent(g)
	lifeMe, lifeOpp := me.Life, opp.Life
	id := effectstest.CastCatalogSpell(t, g, "Gastal Thrillseeker", "Creature — Lizard Berserker", effectstest.GastalThrillseekerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := lifeOpp - opp.Life; got != 1 {
		t.Errorf("opponent lost %d, want 1", got)
	}
	if got := me.Life - lifeMe; got != 1 {
		t.Errorf("you gained %d, want 1", got)
	}
	// The ping made an opponent lose life on your turn: speed 1 from the
	// keyword, then 2 from CR 702.179d.
	if me.Speed != 2 {
		t.Errorf("speed = %d, want 2", me.Speed)
	}
	abilities := effectstest.EffectiveAbilities(t, g, id)
	if effects.EotHasAbility(abilities, "deathtouch") || effects.EotHasAbility(abilities, "haste") {
		t.Error("deathtouch or haste below max speed")
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	abilities = effectstest.EffectiveAbilities(t, g, id)
	if !effects.EotHasAbility(abilities, "deathtouch") || !effects.EotHasAbility(abilities, "haste") {
		t.Errorf("no deathtouch and haste at max speed: %v", abilities)
	}
}

func TestNestingBotLeavesAServoAndIsBiggerAtMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	id := effectstest.PushSpeedCard(t, g, me, "Nesting Bot", "Artifact Creature — Robot", effectstest.NestingBotOracle, 1, 1)
	if p := effectstest.EffectivePower(t, g, id); p != 1 {
		t.Errorf("power below max speed = %d, want 1", p)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if p := effectstest.EffectivePower(t, g, id); p != 2 {
		t.Errorf("power at max speed = %d, want 2", p)
	}
	if tt := effectstest.EffectiveToughness(t, g, id); tt != 1 {
		t.Errorf("toughness at max speed = %d, want 1", tt)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Servo", me.ID); got != 1 {
		t.Errorf("Servo tokens = %d, want 1", got)
	}
}

func TestEndriderCatalyzerAddsRedRedOnlyAtMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	id := effectstest.PushSpeedCard(t, g, me, "Endrider Catalyzer", "Creature — Human Warrior", effectstest.EndriderCatalyzerOracle, 3, 1)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	effectstest.ActivateManaFor(t, g, me.ID, id, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.Total(); got != 2 {
		t.Errorf("pool holds %d mana, want 2", got)
	}
	for _, m := range me.ManaPool.Units() {
		if m.Color != "R" {
			t.Errorf("mana colour %q, want R", m.Color)
		}
	}
}

func TestKickoffCelebrationsRummagesOnEntryAndGrantsHasteAtMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	effectstest.CastCatalogSpell(t, g, "Kickoff Celebrations", "Enchantment", effectstest.KickoffCelebrationsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Declining the discard draws nothing.
	hand := me.Hand.Size()
	if c := effectstest.DiscardChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerDiscard(t, g, me.ID)
	}
	if me.Hand.Size() != hand {
		t.Errorf("declined: hand %d, want %d", me.Hand.Size(), hand)
	}
	col := effectstest.PushCatalogPermanent(g, me.ID, "Kickoff Celebrations", "Enchantment", effectstest.KickoffCelebrationsOracle, false)
	bear := effectstest.PushSpeedCard(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, col, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := g.ActivateCatalogAbility(me.ID, col, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(col) {
		t.Error("the enchantment was not sacrificed")
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, bear), "haste") {
		t.Error("the creature did not gain haste")
	}
}

func TestKickoffCelebrationsDiscardsToDrawTwo(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	effectstest.CastCatalogSpell(t, g, "Kickoff Celebrations", "Enchantment", effectstest.KickoffCelebrationsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	pitch := me.Hand.Cards[0].InstanceID
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d (one discarded, two drawn)", got, hand+1)
	}
}

func TestPointTheWayFetchesSpeedBasicsTapped(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	way := effectstest.PushSpeedCard(t, g, me, "Point the Way", "Enchantment", effectstest.PointTheWayOracle, 0, 0)
	var forests []uuid.UUID
	for i := 0; i < 4; i++ {
		forests = append(forests, effectstest.PushLibraryCardForTest(me, game.Card{
			InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: me.ID, Controller: me.ID,
		}))
	}
	g.SetSpeedForTest(me.ID, 3)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{G}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, way, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c == nil || c.SearchMax != 3 {
		t.Fatalf("search prompt max = %v, want 3 (speed 3)", c != nil)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, forests[:3]...)
	effectstest.PassPriorityAroundTable(t, g)
	lands := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && c.Controller == me.ID {
			lands++
			if !c.Tapped {
				t.Error("a fetched land entered untapped")
			}
		}
	}
	if lands != 3 {
		t.Errorf("fetched %d lands, want 3 (speed 3)", lands)
	}
	if g.Battlefield.Contains(way) {
		t.Error("Point the Way was not sacrificed")
	}
}

func TestHourOfVictoryTutorNeedsMaxSpeedAndSorcerySpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	effectstest.CastCatalogSpell(t, g, "Hour of Victory", "Enchantment", effectstest.HourOfVictoryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Zombie", me.ID); got != 1 {
		t.Fatalf("Zombie tokens = %d, want 1", got)
	}
	var hour uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == effectstest.HourOfVictoryOracle {
			hour = c.InstanceID
		}
	}
	needle := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Needle", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, hour, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	if !g.Battlefield.Contains(hour) {
		t.Fatal("a refused activation paid its sacrifice")
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := g.ActivateCatalogAbility(me.ID, hour, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
	if g.Battlefield.Contains(hour) {
		t.Error("Hour of Victory was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, needle)
	if !me.Hand.Contains(needle) {
		t.Error("the searched card is not in hand")
	}
}

func TestMomentumBreakerEdictsAndDiscardsForTheOpponentWhoCant(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	a := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	c := g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)]
	aBear := effectstest.PushCatalogPermanent(g, a.ID, "A Bear", "Creature — Bear", "", false)
	// b has only a Vehicle that is not a creature; c has nothing.
	bVehicle := effectstest.PushCatalogPermanent(g, b.ID, "B Cart", "Artifact — Vehicle", "", false)
	handB, handC := b.Hand.Size(), c.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Momentum Breaker", "Enchantment", effectstest.MomentumBreakerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, a.ID, aBear)
	effectstest.AnswerSacrifice(t, g, b.ID, bVehicle)
	if g.Battlefield.Contains(aBear) || g.Battlefield.Contains(bVehicle) {
		t.Error("an opponent kept the permanent they should have sacrificed")
	}
	effectstest.DiscardFromHand(t, g, c.ID)
	if got := c.Hand.Size(); got != handC-1 {
		t.Errorf("the opponent who couldn't sacrifice: hand %d, want %d", got, handC-1)
	}
	if got := b.Hand.Size(); got != handB {
		t.Errorf("an opponent who sacrificed discarded: hand %d, want %d", got, handB)
	}
	if me.Speed != 1 {
		t.Errorf("speed = %d, want 1", me.Speed)
	}
}

func TestMomentumBreakerSacrificeGainsLifeEqualToSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	mb := effectstest.PushSpeedCard(t, g, me, "Momentum Breaker", "Enchantment", effectstest.MomentumBreakerOracle, 0, 0)
	g.SetSpeedForTest(me.ID, 3)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, mb, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 3 {
		t.Errorf("gained %d, want 3 (speed 3)", got)
	}
}

func TestGasGuzzlerEntersTappedAndDrawsOnlyAtMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	effectstest.CastCatalogSpell(t, g, "Gas Guzzler", "Creature — Vampire Rogue", effectstest.GasGuzzlerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var guzzler uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == effectstest.GasGuzzlerOracle {
			guzzler = c.InstanceID
			if !c.Tapped {
				t.Error("Gas Guzzler entered untapped")
			}
		}
	}
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Bear", "", false)
	params := game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, guzzler, 0, params); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := g.ActivateCatalogAbility(me.ID, guzzler, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{guzzler}}); err == nil {
		t.Fatal("Gas Guzzler sacrificed itself to its own ability")
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, guzzler, 0, params); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the fodder was not sacrificed")
	}
}

func TestZahurSurveilsOncePerTurnAndMakesZombiesAtMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	zahur := effectstest.PushSpeedCard(t, g, me, "Zahur, Glory's Past", "Legendary Creature — Zombie Cat Warrior", effectstest.ZahurGlorysPastOracle, 3, 2)
	a := effectstest.PushCatalogPermanent(g, me.ID, "Fodder A", "Creature — Bear", "", false)
	b := effectstest.PushCatalogPermanent(g, me.ID, "Fodder B", "Creature — Bear", "", false)
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, zahur, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a}}); err != nil {
		t.Fatalf("first surveil: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, zahur, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{b}}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("second surveil: err = %v, want ErrConditionNotMet", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSurveil {
			if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
				t.Fatalf("ResolveSurveil: %v", err)
			}
			break
		}
	}
	// Below max speed a dying nontoken creature makes nothing.
	zombies := effectstest.CountOnBattlefield(g, "Zombie", me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(b) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Zombie", me.ID) - zombies; got != 0 {
		t.Errorf("below max speed made %d Zombies", got)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	token := effectstest.PushCatalogPermanent(g, me.ID, "Zombie", "Token Creature — Zombie", "", false)
	zombies = effectstest.CountOnBattlefield(g, "Zombie", me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Zombie", me.ID) - zombies + 1; got != 0 {
		t.Errorf("a dying token made a Zombie (net %d)", got)
	}
	nontoken := effectstest.PushCatalogPermanent(g, me.ID, "Nontoken Bear", "Creature — Bear", "", false)
	seen := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		seen[c.InstanceID] = true
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(nontoken) })
	effectstest.PassPriorityAroundTable(t, g)
	made := 0
	for _, c := range g.Battlefield.Cards {
		if seen[c.InstanceID] || c.Name != "Zombie" || c.Controller != me.ID {
			continue
		}
		made++
		if !c.Tapped {
			t.Error("the new Zombie entered untapped")
		}
		if c.Power != 2 || c.Toughness != 2 {
			t.Errorf("the new Zombie is %d/%d, want 2/2", c.Power, c.Toughness)
		}
	}
	if made != 1 {
		t.Errorf("at max speed a nontoken creature dying made %d Zombies, want 1", made)
	}
}

func TestHazoretCantAttackOrBlockBelowMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	opp := speedSliceAOpponent(g)
	hz := effectstest.PushSpeedCard(t, g, me, "Hazoret, Godseeker", "Legendary Creature — God", effectstest.HazoretGodseekerOracle, 5, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hz, opp.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Fatalf("below max speed: err = %v, want ErrCantAttack", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := g.DeclareAttacker(hz, opp.ID); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
}

func TestHazoretCantBlockBelowMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	hz := effectstest.PushSpeedCard(t, g, me, "Hazoret, Godseeker", "Legendary Creature — God", effectstest.HazoretGodseekerOracle, 5, 3)
	if !speedSliceARestricted(t, g, hz, game.CantBlock) {
		t.Error("Hazoret can block below max speed")
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if speedSliceARestricted(t, g, hz, game.CantBlock) {
		t.Error("Hazoret still can't block at max speed")
	}
}

func TestHazoretMakesASmallCreatureUnblockable(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	hz := effectstest.PushSpeedCard(t, g, me, "Hazoret, Godseeker", "Legendary Creature — God", effectstest.HazoretGodseekerOracle, 5, 3)
	small := effectstest.PushSpeedCard(t, g, me, "Small Bear", "Creature — Bear", "", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, hz, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(small)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !speedSliceARestricted(t, g, small, game.CantBeBlocked) {
		t.Error("the target can still be blocked")
	}
}

func TestHazoretCantTargetABigCreature(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	hz := effectstest.PushSpeedCard(t, g, me, "Hazoret, Godseeker", "Legendary Creature — God", effectstest.HazoretGodseekerOracle, 5, 3)
	big := effectstest.PushSpeedCard(t, g, me, "Big Bear", "Creature — Bear", "", 3, 3)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, hz, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(big)}); err == nil {
		t.Error("a power 3 creature was a legal target")
	}
}

func TestHowlsquadHeavyMakesAHastyGoblinThatMustAttack(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	heavy := effectstest.PushSpeedCard(t, g, me, "Howlsquad Heavy", "Creature — Goblin Mercenary", effectstest.HowlsquadHeavyOracle, 2, 3)
	_ = heavy
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && c.Controller == me.ID {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("no Goblin token at the beginning of combat")
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, token), "haste") {
		t.Error("the Goblin token has no haste from Howlsquad Heavy")
	}
	if !effectstest.HasRequirementRecord(g) {
		t.Error("the token has no attack requirement")
	}
	if effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, heavy), "haste") {
		t.Error("Howlsquad Heavy gave itself haste ('other Goblins')")
	}
}

func TestHowlsquadHeavyManaAtMaxSpeedCountsGoblins(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	heavy := effectstest.PushSpeedCard(t, g, me, "Howlsquad Heavy", "Creature — Goblin Mercenary", effectstest.HowlsquadHeavyOracle, 2, 3)
	effectstest.PushSpeedCard(t, g, me, "Goblin A", "Creature — Goblin", "", 1, 1)
	effectstest.PushSpeedCard(t, g, me, "Goblin B", "Creature — Goblin", "", 1, 1)
	effectstest.PushSpeedCard(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	if err := g.ActivateManaAbility(me.ID, heavy, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("below max speed: err = %v, want ErrConditionNotMet", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	effectstest.ActivateManaFor(t, g, me.ID, heavy, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("pool holds %d mana, want 3 (three Goblins)", got)
	}
}

func TestFarFortunePingsOnAttackAndAddsOneAtMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	opp := speedSliceAOpponent(g)
	ff := effectstest.PushSpeedCard(t, g, me, "Far Fortune, End Boss", "Legendary Creature — Human Mercenary", effectstest.FarFortuneEndBossOracle, 4, 5)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	_ = ff
	life := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		life[s.ID] = s.Life
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if s.ID == me.ID {
			continue
		}
		if got := life[s.ID] - s.Life; got != 1 {
			t.Errorf("%s lost %d to the attack trigger below max speed, want 1", s.Name, got)
		}
	}
	// At max speed every damage source you control deals one more to an
	// opponent, Far Fortune's own trigger included.
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	before := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(ff, opp.ID, 2); err != nil {
			t.Fatalf("DealDamageForEffect: %v", err)
		}
	})
	if got := before - opp.Life; got != 3 {
		t.Errorf("2 damage at max speed dealt %d, want 3", got)
	}
	own := me.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(ff, me.ID, 2); err != nil {
			t.Fatalf("DealDamageForEffect: %v", err)
		}
	})
	if got := own - me.Life; got != 2 {
		t.Errorf("2 damage to yourself dealt %d, want 2", got)
	}
}

func TestSamutPumpsOthersAndDiscountsNoncreatureSpellsBySpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	samut := effectstest.PushSpeedCard(t, g, me, "Samut, the Driving Force", "Legendary Creature — Human Warrior Cleric", effectstest.SamutDrivingForceOracle, 4, 5)
	bear := effectstest.PushSpeedCard(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	opp := speedSliceAOpponent(g)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	if me.Speed != 1 {
		t.Fatalf("speed = %d, want 1", me.Speed)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("Bear power at speed 1 = %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, samut); p != 4 {
		t.Errorf("Samut pumped itself: power %d, want 4", p)
	}
	g.SetSpeedForTest(me.ID, 3)
	if p := effectstest.EffectivePower(t, g, bear); p != 5 {
		t.Errorf("Bear power at speed 3 = %d, want 5", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's creature was pumped: power %d", p)
	}
	// {2}{R} less three generic: the reduction stops at the generic part.
	if got := effectstest.PriceInHand(t, g, me, "Shock", "Instant", "{2}{R}"); got != 1 {
		t.Errorf("noncreature spell at speed 3: %d, want 1", got)
	}
	g.SetSpeedForTest(me.ID, 1)
	if got := effectstest.PriceInHand(t, g, me, "Shock", "Instant", "{2}{R}"); got != 2 {
		t.Errorf("noncreature spell at speed 1: %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("creature spell: %d, want 2 (no discount)", got)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Shock", "Instant", "{2}{R}"); got != 3 {
		t.Errorf("an opponent's spell: %d, want 3", got)
	}
}

func TestMendicantCoreCountsArtifactsAndCopiesArtifactSpellsAtMaxSpeed(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	core := effectstest.PushSpeedCard(t, g, me, "Mendicant Core, Guidelight", "Legendary Artifact Creature — Robot", effectstest.MendicantCoreOracle, 0, 3)
	if p := effectstest.EffectivePower(t, g, core); p != 1 {
		t.Errorf("power with just itself = %d, want 1", p)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	if p := effectstest.EffectivePower(t, g, core); p != 2 {
		t.Errorf("power with two artifacts = %d, want 2", p)
	}
	// Below max speed an artifact spell is not copied.
	effectstest.CastCatalogSpell(t, g, "Plain Trinket", "Artifact", "", nil)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("asked to copy below max speed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	trinkets := effectstest.CountOnBattlefield(g, "Plain Trinket", me.ID)
	if trinkets != 1 {
		t.Fatalf("Plain Trinket permanents = %d, want 1", trinkets)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	effectstest.CastCatalogSpell(t, g, "Second Trinket", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("not asked to pay {1} at max speed")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Second Trinket", me.ID); got != 2 {
		t.Errorf("Second Trinket permanents = %d, want 2 (the spell and its token copy)", got)
	}
}

func TestMendicantCoreDeclinedCopyMakesNothing(t *testing.T) {
	g, me := effectstest.SpeedGame(t)
	effectstest.PushSpeedCard(t, g, me, "Mendicant Core, Guidelight", "Legendary Artifact Creature — Robot", effectstest.MendicantCoreOracle, 0, 3)
	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	effectstest.CastCatalogSpell(t, g, "Trinket", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Trinket", me.ID); got != 1 {
		t.Errorf("Trinket permanents = %d, want 1", got)
	}
}
