package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB080Flood    = "02393093-e970-40a0-848b-9cd758450cbb"
	oldB080Workshop = "319ee221-bace-41e9-b386-6b528bd780c1"
	oldB080Goliath  = "498707db-258f-423a-9ff6-9e294fce84d6"
	oldB080Deluge   = "56de4cdc-4db0-48ec-8a16-f66086507233"
	oldB080Ghen     = "6ac59552-9562-492c-aea4-4d64bad58e3e"
	oldB080Colfenor = "90dd050d-f58d-4590-9fc7-af624ee4a13f"
	oldB080Slayer   = "964e20e2-2947-4d8a-a0a2-b8b5fde27ff8"
	oldB080Caller   = "b9d75969-15ac-4509-b95c-a762fc2b99f9"
)

func oldB080Enter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.OracleID, c.Power, c.Toughness = typeLine, oracle, p, tough
	owner.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("enter %s: %v", name, err)
	}
	return c.InstanceID
}

func oldB080Grave(p *game.Player, name, typeLine string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID})
	return id
}

func TestOldB080EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB080Flood, oldB080Workshop, oldB080Goliath, oldB080Deluge, oldB080Ghen, oldB080Colfenor, oldB080Slayer, oldB080Caller} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestOldB080FloodOfRecollectionReturnsAnInstantOrSorceryAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bolt := oldB080Grave(me, "Old Bolt", "Instant", 0, 0)
	bear := oldB080Grave(me, "Old Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Flood of Recollection", "Sorcery", oldB080Flood, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Fatal("Flood targeted a creature card")
	}
	flood := effectstest.CastCatalogSpell(t, g, "Flood of Recollection", "Sorcery", oldB080Flood, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant was not returned to hand")
	}
	if !g.Exile.Contains(flood) || me.Graveyard.Contains(flood) {
		t.Error("Flood of Recollection did not exile itself")
	}
}

func TestOldB080WorkshopAssistantReturnsAnotherArtifactWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	assistant := oldB080Enter(t, g, me, "Workshop Assistant", "Artifact Creature — Construct", oldB080Workshop, 1, 2)
	rock := oldB080Grave(me, "Old Rock", "Artifact", 0, 0)
	effectstest.KillCreature(t, g, me.ID, assistant)
	effectstest.AnswerPickTargetIfOpen(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the artifact card was not returned")
	}
	if me.Hand.Contains(assistant) || !me.Graveyard.Contains(assistant) {
		t.Error("the Assistant returned itself; the trigger says another artifact card")
	}
}

func TestOldB080SkaabGoliathExilesTwoCreatureCardsAsItIsCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	a := oldB080Grave(me, "Dead A", "Creature — Bear", 2, 2)
	b := oldB080Grave(me, "Dead B", "Creature — Bear", 2, 2)
	land := oldB080Grave(me, "Dead Land", "Land", 0, 0)
	id := effectstest.SocB17Hand(me, "Skaab Goliath", "Creature — Zombie Giant", "{5}{U}", oldB080Goliath)
	effectstest.FloatMana(t, g, me, "{U}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, ExileIDs: []uuid.UUID{a, land}}); err == nil {
		t.Fatal("cast with a land card as one of the exiled cards")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no cards exiled")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, ExileIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Error("the two creature cards were not exiled")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("Skaab Goliath did not resolve")
	}
}

func TestOldB080GhenSacrificesAnEnchantmentToReturnAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ghen := effectstest.B21Push(g, me.ID, "Ghen, Arcanum Weaver", "Legendary Creature — Human Wizard", oldB080Ghen, 2, 3)
	effectstest.Unsick(g, ghen)
	fodder := effectstest.B21Push(g, me.ID, "Fodder", "Enchantment", "", 0, 0)
	back := oldB080Grave(me, "Old Enchantment", "Enchantment", 0, 0)
	effectstest.FloatMana(t, g, me, "{R}{W}{B}")
	effectstest.B16Activate(t, g, me.ID, ghen, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: back}}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the enchantment was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(back) {
		t.Error("the enchantment card did not return to the battlefield")
	}
}

func TestOldB080ColfenorRefusesEqualToughnessAndTheDeadCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB080Enter(t, g, me, "Colfenor, the Last Yew", "Legendary Creature — Treefolk Shaman", oldB080Colfenor, 3, 7)
	victim := effectstest.B21Push(g, me.ID, "Mid Bear", "Creature — Bear", "", 3, 4)
	small := oldB080Grave(me, "Small Dead", "Creature — Rat", 1, 3)
	same := oldB080Grave(me, "Same Dead", "Creature — Rat", 1, 4)
	effectstest.KillCreature(t, g, me.ID, victim)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: same}); err == nil {
		t.Error("a creature card with equal toughness was a legal target")
	}
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: victim}); err == nil {
		t.Error("the creature that died was a legal target")
	}
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: small}); err != nil {
		t.Errorf("the lesser-toughness card was refused: %v", err)
	}
}

func TestOldB080ColfenorPicksTheLesserToughnessCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB080Enter(t, g, me, "Colfenor, the Last Yew", "Legendary Creature — Treefolk Shaman", oldB080Colfenor, 3, 7)
	victim := effectstest.B21Push(g, me.ID, "Mid Bear", "Creature — Bear", "", 3, 4)
	small := oldB080Grave(me, "Small Dead", "Creature — Rat", 1, 3)
	effectstest.KillCreature(t, g, me.ID, victim)
	effectstest.AnswerPickTargetIfOpen(t, g, small)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(small) {
		t.Error("the lesser-toughness creature card was not returned")
	}
}

func TestOldB080IroncladSlayerMayReturnAnAuraOrEquipment(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		sword := oldB080Grave(me, "Old Sword", "Artifact — Equipment", 0, 0)
		oldB080Grave(me, "Old Rock", "Artifact", 0, 0)
		oldB080Enter(t, g, me, "Ironclad Slayer", "Creature — Human Warrior", oldB080Slayer, 3, 2)
		effectstest.AnswerPickTargetIfOpen(t, g, sword)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Contains(sword) != yes {
			t.Errorf("answer %v: Equipment in hand = %v", yes, me.Hand.Contains(sword))
		}
	}
}

func TestOldB080NullCallerExilesACreatureCardForATappedZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	caller := effectstest.B21Push(g, me.ID, "Null Caller", "Creature — Vampire Shaman", oldB080Caller, 2, 4)
	dead := oldB080Grave(me, "Dead Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{B}{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, caller, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{dead}})
	if !g.Exile.Contains(dead) {
		t.Error("the creature card was not exiled as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	var zombie uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Zombie" && c.Controller == me.ID {
				zombie = c.InstanceID
			}
		}
	})
	if zombie == uuid.Nil {
		t.Fatal("no Zombie token")
	}
	if !effectstest.B16Tapped(t, g, zombie) || effectstest.EffectivePower(t, g, zombie) != 2 {
		t.Error("the token is not a tapped 2/2")
	}
}

func TestOldB080MnemonicDelugeMakesThreeFreeCopiesAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Their Cantrip", TypeLine: "Sorcery", ManaCost: "{2}{U}", Owner: opp.ID, Controller: opp.ID})
	deluge := effectstest.CastCatalogSpell(t, g, "Mnemonic Deluge", "Sorcery", oldB080Deluge, []game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) {
		t.Error("the card was not exiled")
	}
	if !g.Exile.Contains(deluge) {
		t.Error("Mnemonic Deluge did not exile itself")
	}
	copies := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.Name == "Their Cantrip" && c.InstanceID != spell {
				copies++
			}
		}
		for _, c := range me.Graveyard.Cards {
			if c.Name == "Their Cantrip" {
				copies++
			}
		}
	})
	if copies != 3 {
		t.Errorf("copies = %d, want 3", copies)
	}
}
