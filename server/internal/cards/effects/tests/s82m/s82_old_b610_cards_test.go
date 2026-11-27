package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b610_cards_test.go — Eighth Edition, damage (S82 OLD batch 610,
// #5607). Every test plays the real catalog entry through the engine.

const (
	oldB610DingusEgg      = "1973f1e9-aa14-49dc-bafe-be17b30ba288"
	oldB610SpittingSpider = "24ddcf3c-a4f3-4373-9483-5ce6e50d6150"
	oldB610SuddenImpact   = "2bf65e7d-6d31-4ad6-93cd-623612d3175d"
	oldB610VexingArcanix  = "2fd5b7e5-496c-4f6d-85e0-256afb7544d1"
	oldB610Javelineer     = "649300e2-c491-435d-92d5-f1b88f0e2e2e"
	oldB610NoblePurpose   = "84528204-faf2-45cb-bce0-365b8adf5339"
	oldB610SeismicAssault = "8ad4f2fe-6d98-4279-a331-3817d40ae46d"
	oldB610LightningBlast = "91fd731c-e076-4f2d-9f22-872880c3cc3d"
)

func TestS82OldB610EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB610DingusEgg: "Dingus Egg", oldB610SpittingSpider: "Spitting Spider",
		oldB610SuddenImpact: "Sudden Impact", oldB610VexingArcanix: "Vexing Arcanix",
		oldB610Javelineer: "Elite Javelineer", oldB610NoblePurpose: "Noble Purpose",
		oldB610SeismicAssault: "Seismic Assault", oldB610LightningBlast: "Lightning Blast",
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

func TestS82OldB610LightningBlastDealsFourToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Blast", "Instant", oldB610LightningBlast, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("player: life %d → %d, want 4 damage", life, opp.Life)
	}
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.CastCatalogSpell(t, g, "Lightning Blast", "Instant", oldB610LightningBlast, []game.TargetRef{{Kind: game.TargetCard, ID: wall}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, wall); got != 4 {
		t.Errorf("creature: took %d, want 4", got)
	}
	_ = me
}

func TestS82OldB610SuddenImpactDealsDamageEqualToTheirHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.D2dSetHand(opp, 5)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Sudden Impact", "Instant", oldB610SuddenImpact, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("life %d → %d, want 5 damage for their five cards", life, opp.Life)
	}
}

func TestS82OldB610SeismicAssaultDiscardsALandForTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	assault := effectstest.PushCatalogPermanent(g, me.ID, "Seismic Assault", "Enchantment", oldB610SeismicAssault, false)
	land := effectstest.PushCatalogHandCard(me, "Mountain", "Basic Land — Mountain", "")
	spell := effectstest.PushCatalogHandCard(me, "Shock", "Instant", "")
	if err := g.ActivateCatalogAbility(me.ID, assault, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.PlayerRefs(opp.ID), DiscardIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("paid with a nonland discard")
	}
	before := opp.Life
	effectstest.B16Activate(t, g, me.ID, assault, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID), DiscardIDs: []uuid.UUID{land}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("dealt %d, want 2", got)
	}
	if effectstest.HandHasID(me, land) || !effectstest.HandHasID(me, spell) {
		t.Error("the land was not the card discarded")
	}
}

func TestS82OldB610DingusEggHurtsTheLandsControllerWhateverTheLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Dingus Egg", "Artifact", oldB610DingusEgg, false)
	mine := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	creature := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	myLife, theirLife := me.Life, opp.Life
	effectstest.B25Destroy(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 || me.Life != myLife {
		t.Errorf("their land: lives me %d→%d, them %d→%d; want only them −2", myLife, me.Life, theirLife, opp.Life)
	}
	effectstest.B25Destroy(g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-2 {
		t.Errorf("my own land: life %d → %d, want −2", myLife, me.Life)
	}
	effectstest.B25Destroy(g, creature)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 {
		t.Errorf("a creature dying hurt them: %d → %d", theirLife, opp.Life)
	}
}

func TestS82OldB610SpittingSpiderSacrificesALandToHitEveryFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spider := effectstest.B12Push(g, me.ID, "Spitting Spider", "Creature — Spider", oldB610SpittingSpider, 3, 5)
	if !effectstest.HasKeywordOnBattlefield(t, g, spider, "reach") {
		t.Error("no reach")
	}
	land := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 2, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, spider, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if got := effectstest.DamageMarkedOn(g, bird); got != 1 {
		t.Errorf("the flyer took %d, want 1", got)
	}
	if got := effectstest.DamageMarkedOn(g, bear); got != 0 {
		t.Errorf("the ground creature took %d, want 0", got)
	}
	if got := effectstest.DamageMarkedOn(g, spider); got != 0 {
		t.Errorf("the Spider took %d, want 0", got)
	}
	// With no land to sacrifice it cannot be activated.
	if err := g.ActivateCatalogAbility(me.ID, spider, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no land to sacrifice")
	}
}

func TestS82OldB610EliteJavelineerShootsAnAttackerWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	jav := effectstest.B12Push(g, opp.ID, "Elite Javelineer", "Creature — Human Soldier", oldB610Javelineer, 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	other := effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := effectstest.MbBlock(g, jav, attacker); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.B04WaitForPick(t, g, opp.ID)
	effectstest.PickCard(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, attacker); got != 1 {
		t.Errorf("the blocked attacker took %d from the trigger, want 1", got)
	}
	if got := effectstest.DamageMarkedOn(g, other); got != 0 {
		t.Errorf("a creature that is not attacking took %d, want 0", got)
	}
}

func TestS82OldB610NoblePurposeGainsWhatYourCreaturesDealInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Noble Purpose", "Enchantment", oldB610NoblePurpose, false)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 3, 3)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, big, small)
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5 (3 and 2 dealt)", life, me.Life)
	}
}

func TestS82OldB610NoblePurposeIgnoresTheBlockersDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Noble Purpose", "Enchantment", oldB610NoblePurpose, false)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Giant", 3, 5)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 4, 4)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := effectstest.MbBlock(g, blocker, attacker); err != nil {
		t.Fatal(err)
	}
	life, oppLife := me.Life, opp.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3: only the creature I control counts, not the blocker's 4", life, me.Life)
	}
	if opp.Life != oppLife {
		t.Errorf("opponent life changed %d → %d", oppLife, opp.Life)
	}
}

// oldB610Arcanix puts the Arcanix on the active seat's side with {3} floating
// and the target player's top card set.
func oldB610Arcanix(t *testing.T, g *game.Game, top string) (*game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	arcanix := effectstest.PushCatalogPermanent(g, me.ID, "Vexing Arcanix", "Artifact", oldB610VexingArcanix, false)
	opp.Library.Cards = nil
	for i := 0; i < 3; i++ {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	}
	topID := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: topID, Name: top, TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	effectstest.B06AddMana(me, "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, arcanix, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	return me, opp, arcanix, topID
}

func oldB610Name(t *testing.T, g *game.Game, chooser uuid.UUID, name string) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceCardName {
			if c.Chooser != chooser {
				t.Fatalf("the name prompt asks %v, want the target player %v", c.Chooser, chooser)
			}
			if err := g.ResolveCardNameChoice(c.ID, c.Chooser, name); err != nil {
				t.Fatalf("name: %v", err)
			}
			return
		}
	}
	t.Fatal("no card-name prompt")
}

func TestS82OldB610VexingArcanixMatchPutsTheCardIntoTheirHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _, top := oldB610Arcanix(t, g, "Counterspell")
	life := opp.Life
	oldB610Name(t, g, opp.ID, "counterspell")
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, opp, top); z != "hand" {
		t.Errorf("a right guess: the card is in %s, want the hand", z)
	}
	if opp.Life != life {
		t.Errorf("a right guess cost %d life", life-opp.Life)
	}
}

func TestS82OldB610VexingArcanixMissPutsItIntoTheGraveyardAndHurtsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _, top := oldB610Arcanix(t, g, "Counterspell")
	life := opp.Life
	oldB610Name(t, g, opp.ID, "Opt")
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, opp, top); z != "graveyard" {
		t.Errorf("a wrong guess: the card is in %s, want the graveyard", z)
	}
	if opp.Life != life-2 {
		t.Errorf("a wrong guess: life %d → %d, want 2 damage", life, opp.Life)
	}
}
