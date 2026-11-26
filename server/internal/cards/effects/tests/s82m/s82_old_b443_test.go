package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dark Ascension batch 443: werewolves, Loyal Cathar and the rest.

const (
	oldB443Elder     = "12330f42-0e3f-40e1-86f3-57a34403d143"
	oldB443Counter   = "16de96d0-a58e-42a7-9a13-45358542a246"
	oldB443Altar     = "171a3ac4-0c9c-46ac-b273-089d566cca05"
	oldB443Warden    = "5d66c2b9-6011-4168-b468-0bde50bac37e"
	oldB443Captive   = "7f59252f-3a37-4a22-938c-001f3208bd94"
	oldB443Hermit    = "8287306b-47d3-4e95-aa39-551f949d687c"
	oldB443Seance    = "a375b935-c02b-44bc-a942-37c5de58b1e4"
	oldB443Woods     = "a354b7f3-bfd0-47c6-bb16-1a5f8e26467d"
	oldB443Cathar    = "bf28a2ed-cad7-48a7-9dec-06e43af73e9c"
	oldB443Aristo    = "cd2a87c4-d1c3-4a9e-b2df-e6fe251421c4"
	oldB443Archangel = "fc734a75-93bf-448b-b347-810ae835fe49"
)

var (
	oldB443ElderCard = effectstest.IwCard{Oracle: oldB443Elder, Front: "Lambholt Elder", FrontType: "Creature — Human Werewolf", Cost: "{2}{G}",
		FrontText: effectstest.IwFrontText, Fp: "1", Ft: "2",
		Back: "Silverpelt Werewolf", BackText: "Whenever this creature deals combat damage to a player, draw a card.\n" + effectstest.IwBackText, Bp: "4", Bt: "5",
		Colors: []string{"G"}}
	oldB443CaptiveCard = effectstest.IwCard{Oracle: oldB443Captive, Front: "Wolfbitten Captive", FrontType: "Creature — Human Werewolf", Cost: "{G}",
		FrontText: "{1}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn.\n" + effectstest.IwFrontText, Fp: "1", Ft: "1",
		Back: "Krallenhorde Killer", BackText: "{3}{G}: This creature gets +4/+4 until end of turn. Activate only once each turn.\n" + effectstest.IwBackText, Bp: "2", Bt: "2",
		Colors: []string{"G"}}
	oldB443HermitCard = effectstest.IwCard{Oracle: oldB443Hermit, Front: "Hinterland Hermit", FrontType: "Creature — Human Werewolf", Cost: "{1}{R}",
		FrontText: effectstest.IwFrontText, Fp: "2", Ft: "1",
		Back: "Hinterland Scourge", BackText: "This creature must be blocked if able.\n" + effectstest.IwBackText, Bp: "3", Bt: "2",
		Colors: []string{"R"}}
)

func TestOldB443CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB443Elder: "Lambholt Elder", oldB443Elder + "#1": "Silverpelt Werewolf",
		oldB443Captive: "Wolfbitten Captive", oldB443Captive + "#1": "Krallenhorde Killer",
		oldB443Hermit: "Hinterland Hermit", oldB443Hermit + "#1": "Hinterland Scourge",
		oldB443Cathar: "Loyal Cathar", oldB443Cathar + "#1": "Unhallowed Cathar",
		oldB443Aristo: "Falkenrath Aristocrat", oldB443Counter: "Counterlash", oldB443Seance: "Séance",
		oldB443Woods: "Lost in the Woods", oldB443Altar: "Altar of the Lost",
		oldB443Archangel: "Archangel's Light", oldB443Warden: "Warden of the Wall",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q, completeness %v, caveats %v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB443LambholtElderTurnsOverOnAQuietTurnAndDrawsOnCombatDamage(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[1]
	id := effectstest.IwPut(t, g, oldB443ElderCard, me)
	effectstest.IwNextUpkeep(t, g, nil)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatal("a turn with no spells did not turn the Elder over")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 4 || tough != 5 {
		t.Errorf("Silverpelt Werewolf is %d/%d, want 4/5", p, tough)
	}
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, id, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
	effectstest.IwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 2})
	if effectstest.IwFace(g, id) != 0 {
		t.Error("a player casting two spells did not turn it back")
	}
}

func TestOldB443WolfbittenCaptivePumpsOncePerTurn(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, oldB443CaptiveCard, me)
	effectstest.FloatMana(t, g, me, "{C}{G}{C}{G}")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 3 || tough != 3 {
		t.Fatalf("pumped Captive is %d/%d, want 3/3", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second activation in one turn was accepted")
	}
	effectstest.IwNextUpkeep(t, g, nil)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatal("did not turn over")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{G}")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 6 || tough != 6 {
		t.Errorf("pumped Killer is %d/%d, want 6/6", p, tough)
	}
}

func TestOldB443HinterlandScourgeMustBeBlockedIfAble(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	id := effectstest.IwPutBack(t, g, oldB443HermitCard, me)
	wall := effectstest.ReqCreature(g, opp.ID, "Wall", "", 0, 4)
	effectstest.ReqToBlockers(t, g, id)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Hinterland Scourge must be blocked if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := effectstest.ReqBlock(t, g, id, wall); err != nil {
		t.Fatal(err)
	}
}

func oldB443CatharRow() cards.Card {
	return cards.Card{
		ID: uuid.New(), Name: "Loyal Cathar // Unhallowed Cathar", Layout: "transform",
		TypeLine: "Creature — Human Soldier // Creature — Zombie Soldier", OracleID: uuid.MustParse(oldB443Cathar),
		CardFaces: []cards.CardFace{
			{Name: "Loyal Cathar", TypeLine: "Creature — Human Soldier", ManaCost: "{W}{W}", Colors: []string{"W"}, Power: "2", Toughness: "2",
				OracleText: "Vigilance\nWhen this creature dies, return it to the battlefield transformed under your control at the beginning of the next end step."},
			{Name: "Unhallowed Cathar", TypeLine: "Creature — Zombie Soldier", Colors: []string{"W"}, Power: "2", Toughness: "1",
				OracleText: "This creature can't block."},
		},
	}
}

func TestOldB443LoyalCatharReturnsTransformedAtTheNextEndStep(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.ImportToHand(oldB443CatharRow(), me)
	c, _ := effectstest.CardInZone(me.Hand, id)
	me.Hand.Remove(id)
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Fatal("it came back at once")
	}
	for i := 0; i < 12 && g.Turn.Step != game.StepEnd; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	var back game.Card
	for _, bc := range g.Battlefield.Cards {
		if bc.Controller == me.ID && bc.ActiveFace == 1 && bc.Name == "Unhallowed Cathar" {
			back = bc
		}
	}
	if back.InstanceID == uuid.Nil {
		t.Fatal("no Unhallowed Cathar under your control at the end step")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, back.InstanceID); p != 2 || tough != 1 {
		t.Errorf("Unhallowed Cathar is %d/%d, want 2/1", p, tough)
	}
}

func TestOldB443FalkenrathAristocratSacrificesForIndestructibleAndAHumanCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	vamp := effectstest.B12Push(g, me.ID, "Falkenrath Aristocrat", "Creature — Vampire Noble", oldB443Aristo, 4, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human", 1, 1)
	effectstest.B16Activate(t, g, me.ID, vamp, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, vamp) != 0 || !effectstest.EffectiveAbilitiesContain(t, g, vamp, "indestructible") {
		t.Fatalf("a Bear: counters %d, indestructible %v", effectstest.PlusOnes(g, vamp), effectstest.EffectiveAbilitiesContain(t, g, vamp, "indestructible"))
	}
	effectstest.B16Activate(t, g, me.ID, vamp, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{human}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, vamp) != 1 {
		t.Errorf("a Human gave %d counters, want 1", effectstest.PlusOnes(g, vamp))
	}
}

func TestOldB443CounterlashCountersAndOffersAFreeSpellOfASharedType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Big Bolt", effectstest.LightningBoltOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	hand := func(name, typeLine, cost string) uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID,
			Power: 2, Toughness: 2})
		return id
	}
	myInstant := hand("My Instant", "Instant", "{1}{R}")
	myBear := hand("My Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Counterlash", "Instant", oldB443Counter,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the spell was not countered")
	}
	offer := effectstest.ResolutionCastOfferFor(g, me.ID)
	if offer == nil {
		t.Fatal("no free-cast offer after the counter")
	}
	cast := func(id uuid.UUID) error {
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost})
	}
	if err := cast(myBear); err == nil {
		t.Error("a creature spell was cast free off a countered instant")
	}
	if err := cast(myInstant); err != nil {
		t.Fatalf("cast the instant free: %v", err)
	}
}

func TestOldB443SeanceMakesAnExiledSpiritCopyAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Séance", "Enchantment", oldB443Seance, 0, 0)
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Grave Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	start := g.Turn.ActiveSeat
	for i := 0; i < 40 && (g.Turn.Step != game.StepUpkeep || g.Turn.ActiveSeat == start); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestTriggerPromptFor(g, me.ID); p != nil {
		if err := g.ResolveTriggerPrompt(p.ID, me.ID, true); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bear) {
		t.Fatal("the card is still in the graveyard")
	}
	var token game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Grave Bear" && effects.IsToken(c) && c.Controller == me.ID {
			token = c
		}
	}
	if token.InstanceID == uuid.Nil {
		t.Fatal("no token copy of Grave Bear")
	}
	if !token.HasSubtype("Spirit") || !token.HasSubtype("Bear") {
		t.Errorf("token type line %q, want a Bear Spirit", token.TypeLine)
	}
	for i := 0; i < 12 && g.Turn.Step != game.StepEnd; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token.InstanceID) {
		t.Error("the token was not exiled at the beginning of the next end step")
	}
}

func TestOldB443LostInTheWoodsRemovesAnAttackerOnlyForAForest(t *testing.T) {
	for _, forest := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, opp.ID, "Lost in the Woods", "Enchantment", oldB443Woods, 0, 0)
		atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
		top := effectstest.SeedLibraryTop(opp, "Plains", "Basic Land — Plains", "")
		if forest {
			top = effectstest.SeedLibraryTop(opp, "Forest", "Basic Land — Forest", "")
		}
		libBefore := opp.Library.Size()
		effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
		effectstest.DeclareAttack(t, g, opp.ID, atk)
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.FindBattlefieldCardForTest(g, atk)
		if removed := c.AttackingTarget == uuid.Nil; removed != forest {
			t.Errorf("forest=%v: removed from combat = %v", forest, removed)
		}
		if opp.Library.Size() != libBefore || opp.Library.Contains(top) == false {
			t.Errorf("forest=%v: the revealed card should stay in the library (bottom)", forest)
		}
		if got, _ := opp.Library.Top(); got.InstanceID == top {
			t.Errorf("forest=%v: the revealed card is still on top", forest)
		}
	}
}

func TestOldB443ArchangelsLightGainsTwoPerGraveyardCardThenShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 5; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	life, lib := me.Life, me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Archangel's Light", "Sorcery", oldB443Archangel, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+10 {
		t.Errorf("life %d → %d, want +10", life, me.Life)
	}
	if me.Graveyard.Size() != 1 || me.Library.Size() != lib+5 {
		t.Errorf("graveyard %d, library %d → %d: want only the spell left in the graveyard and +5 in the library", me.Graveyard.Size(), lib, me.Library.Size())
	}
}

func TestOldB443WardenOfTheWallIsAGargoyleOnOtherPlayersTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Warden of the Wall", "Artifact", oldB443Warden, 0, 0)
	if c := effectstest.LayeredCard(t, g, id); c.IsCreature() {
		t.Fatal("a creature on its controller's own turn")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	c := effectstest.LayeredCard(t, g, id)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Gargoyle") {
		t.Fatalf("on another player's turn it is %q", c.TypeLine)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 2 || tough != 3 {
		t.Errorf("%d/%d, want 2/3", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, id, "flying") {
		t.Error("no flying")
	}
}

func TestOldB443AltarOfTheLostMakesRestrictedMana(t *testing.T) {
	spec, ok := effects.Lookup(oldB443Altar)
	if !ok || len(spec.ManaAbilities) != 1 {
		t.Fatalf("Altar of the Lost: %v, %d mana abilities", ok, len(spec.ManaAbilities))
	}
	ab := spec.ManaAbilities[0]
	if len(ab.Restrictions) != 2 || ab.Restrictions[0] != game.ManaRestrictSpellFlashback || ab.Restrictions[1] != game.ManaRestrictCastFromGraveyard {
		t.Errorf("restrictions %v", ab.Restrictions)
	}
}
