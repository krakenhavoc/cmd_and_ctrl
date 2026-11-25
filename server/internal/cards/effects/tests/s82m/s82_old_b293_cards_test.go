package testss82m

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b293_cards_test.go — Shadows over Innistrad cards, S82 OLD batch
// 293, played through the real catalog entries.

const (
	oldB293PuncturingLight   = "15e71226-feb0-4870-a3ea-d8604ca5b3dc"
	oldB293StructuralDist    = "bf0b6800-8e6b-4ccd-8e4d-96e61342f981"
	oldB293Grafstone         = "4c702b96-3288-435d-9154-5b7454419896"
	oldB293RottenheartGhoul  = "f9535aa3-0c07-4fad-8dc9-9838a9ebd304"
	oldB293HowlpackWolf      = "a2fb7d03-d6bf-4716-9ca4-bd7dbcf9ae66"
	oldB293SilburlindSnapper = "eb2fe2f7-60f7-44d3-8d4a-4e102012435d"
	oldB293BreakneckRider    = "51685f85-06ad-4de2-a7a6-06e6f59a1cbb"
	oldB293RunawayCarriage   = "67c16435-54ef-4786-85cd-cacca4f06670"
	oldB293Vessel            = "c5a4d485-ab9c-48bb-866f-448b1f7707cc"
)

var oldB293Rider = effectstest.IwCard{Oracle: oldB293BreakneckRider, Front: "Breakneck Rider", FrontType: "Creature — Human Scout Werewolf", Cost: "{1}{R}{R}",
	FrontText: effectstest.IwFrontText, Fp: "3", Ft: "3", Back: "Neck Breaker",
	BackText: "Attacking creatures you control get +1/+0 and have trample.\n" + effectstest.IwBackText, Bp: "4", Bt: "3", Colors: []string{"R"}}

func TestOldB293CardsAreFullWithNoCaveats(t *testing.T) {
	for id, name := range map[string]string{
		oldB293PuncturingLight: "Puncturing Light", oldB293StructuralDist: "Structural Distortion",
		oldB293Grafstone: "Corrupted Grafstone", oldB293RottenheartGhoul: "Rottenheart Ghoul",
		oldB293HowlpackWolf: "Howlpack Wolf", oldB293SilburlindSnapper: "Silburlind Snapper",
		oldB293BreakneckRider: "Breakneck Rider", oldB293BreakneckRider + "#1": "Neck Breaker",
		oldB293RunawayCarriage: "Runaway Carriage", oldB293Vessel: "Vessel of Volatility",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB293PuncturingLightKillsOnlyASmallAttackerOrBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, small, big)
	cast := func(target uuid.UUID, strict bool) error {
		id := effectstest.HandCardOf(g, me, "Puncturing Light", "Instant", "", oldB293PuncturingLight)
		return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(target), Strict: strict})
	}
	if err := cast(big, true); err == nil {
		t.Error("targeted an attacker with power 4")
	}
	if err := cast(idle, true); err == nil {
		t.Error("targeted a creature that is neither attacking nor blocking")
	}
	if err := cast(small, true); err != nil {
		t.Fatalf("CastSpell on a small attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the small attacker survived")
	}
}

func TestOldB293StructuralDistortionExilesAndHitsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest", "", false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Structural Distortion", "Sorcery", oldB293StructuralDist, effectstest.CardRefs(bear)); err == nil {
		t.Error("targeted a nonartifact creature")
	}
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Structural Distortion", "Sorcery", oldB293StructuralDist, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land stayed on the battlefield")
	}
	if zone, _ := effectstest.CardWhere(g, opp, land); zone != "exile" {
		t.Errorf("the land is in %s, want exile", zone)
	}
	if opp.Life != before-2 {
		t.Errorf("life %d, want %d", opp.Life, before-2)
	}
}

func TestOldB293GrafstoneEntersTappedAndOffersGraveyardColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	stone := effectstest.SeedPermanentWithOracle(g, me.ID, "Corrupted Grafstone", "Artifact", oldB293Grafstone)
	if err := g.ActivateManaAbility(me.ID, stone, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility with an empty graveyard: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Fatalf("mana from an empty graveyard: %v", me.ManaPool)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == stone {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Red Thing", TypeLine: "Creature — Goblin", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	if err := g.ActivateManaAbility(me.ID, stone, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "R" {
		t.Errorf("pool = %v, want one {R}", me.ManaPool)
	}
}

func TestOldB293RottenheartGhoulMakesTheTargetPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := effectstest.B12Push(g, me.ID, "Rottenheart Ghoul", "Creature — Zombie", oldB293RottenheartGhoul, 2, 4)
	spare := effectstest.HandCard(opp, "Spare", "Sorcery")
	before := opp.Hand.Size()
	effectstest.KillCreature(t, g, me.ID, ghoul)
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("pick target: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, spare)
	if got := opp.Hand.Size(); got != before-1 {
		t.Errorf("hand is %d, want %d", got, before-1)
	}
}

func oldB293WolfTable(t *testing.T, extras func(g *game.Game, attacker, defender uuid.UUID)) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	attacker, defender := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, attacker.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	wolf := effectstest.B12Push(g, defender.ID, "Howlpack Wolf", "Creature — Wolf", oldB293HowlpackWolf, 3, 3)
	extras(g, attacker.ID, defender.ID)
	effectstest.BrAttack(t, g, bear)
	return g, wolf, bear
}

func TestOldB293HowlpackWolfCantBlockAloneOrBesideNonWolves(t *testing.T) {
	for name, extras := range map[string]func(g *game.Game, attacker, defender uuid.UUID){
		"alone":         func(*game.Game, uuid.UUID, uuid.UUID) {},
		"beside an elf": func(g *game.Game, _, d uuid.UUID) { effectstest.B12Creature(g, d, "Elf", "Creature — Elf", 1, 1) },
		"their own wolf": func(g *game.Game, a, _ uuid.UUID) {
			effectstest.B12Creature(g, a, "Their Wolf", "Creature — Wolf", 1, 1)
		},
	} {
		t.Run(name, func(t *testing.T) {
			g, wolf, bear := oldB293WolfTable(t, extras)
			effectstest.BrRefusal(t, g.DeclareBlocker(wolf, bear), game.BlockReasonCantBlockAttacker)
		})
	}
}

func TestOldB293HowlpackWolfBlocksBesideAnotherWolfOrWerewolf(t *testing.T) {
	for name, typeLine := range map[string]string{"wolf": "Creature — Wolf", "werewolf": "Creature — Werewolf"} {
		t.Run(name, func(t *testing.T) {
			g, wolf, bear := oldB293WolfTable(t, func(g *game.Game, _, d uuid.UUID) {
				effectstest.B12Creature(g, d, "Pack Mate", typeLine, 1, 1)
			})
			if err := g.DeclareBlocker(wolf, bear); err != nil {
				t.Errorf("block beside another %s: %v", name, err)
			}
		})
	}
}

func TestOldB293SilburlindSnapperAttacksOnlyAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snapper := effectstest.B12Push(g, me.ID, "Silburlind Snapper", "Creature — Turtle", oldB293SilburlindSnapper, 6, 6)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	effectstest.WantAttackRefused(t, g, snapper, opp.ID, "hasn't cast a noncreature spell this turn")
	g.WithWriteLock(func() {
		g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{me.ID: {Total: 1}}
	})
	effectstest.WantAttackRefused(t, g, snapper, opp.ID, "hasn't cast a noncreature spell this turn")
	g.WithWriteLock(func() {
		g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{me.ID: {Total: 1, Noncreature: 1}}
	})
	if err := effectstest.DeclareResult(g, snapper, opp.ID); err != nil {
		t.Errorf("attack after a noncreature spell: %v", err)
	}
}

func TestOldB293NeckBreakerBoostsAndTramplesAttackers(t *testing.T) {
	g, me := effectstest.IwTable(t)
	effectstest.IwPutBack(t, g, oldB293Rider, me)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p := effectstest.PowerOf(t, g, bear); p != 2 {
		t.Errorf("idle bear power %d, want 2", p)
	}
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, bear)
	if p := effectstest.PowerOf(t, g, bear); p != 3 {
		t.Errorf("attacking bear power %d, want 3", p)
	}
	effectstest.AssertKeywords(t, g, bear, "trample")
}

func TestOldB293BreakneckRiderTurnsOverAfterAQuietTurnAndBack(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, oldB293Rider, me)
	effectstest.IwNextUpkeep(t, g, nil)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatalf("did not transform after a turn with no spells")
	}
	effectstest.IwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 2})
	if effectstest.IwFace(g, id) != 0 {
		t.Errorf("did not turn back after a player cast two spells")
	}
}

func TestOldB293RunawayCarriageIsSacrificedAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	car := effectstest.B12Push(g, me.ID, "Runaway Carriage", "Artifact Creature — Construct", oldB293RunawayCarriage, 5, 6)
	effectstest.DeclareAttack(t, g, opp.ID, car)
	effectstest.PassPriorityAroundTable(t, g)
	if zone, _ := effectstest.CardWhere(g, me, car); zone != "battlefield" {
		t.Fatalf("carriage is in %s before end of combat", zone)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if zone, _ := effectstest.CardWhere(g, me, car); zone != "graveyard" {
		t.Errorf("carriage is in %s after end of combat, want the graveyard", zone)
	}
}

func TestOldB293VesselOfVolatilityPaysTwoForFourRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	v := effectstest.SeedPermanentWithOracle(g, me.ID, "Vessel of Volatility", "Enchantment", oldB293Vessel)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.ActivateManaAbility(me.ID, v, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if g.Battlefield.Contains(v) {
		t.Error("the Vessel was not sacrificed")
	}
	if me.ManaPool.Total() != 4 {
		t.Errorf("pool %v, want four mana", me.ManaPool)
	}
}
