package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB079Invoker   = "17ed18cf-33e5-4861-8a36-8f7bb91971f9"
	oldB079Captain   = "4f4ec0fa-fcaf-4dc2-9c67-ee7881f5f77d"
	oldB079Crew      = "8e0f93d7-7208-4902-bd92-e7bf66d35fef"
	oldB079Will      = "a32860dd-e8df-4a3c-bdc2-948dcd07cf9a"
	oldB079Furnace   = "af6d6844-c612-4731-86da-59a8fa02956b"
	oldB079Gnostro   = "bff55e1a-730c-4349-9ccb-18c97f46496d"
	oldB079Blim      = "c7cf9574-e2cf-4137-a224-73f76c5d11d7"
	oldB079Sacrifice = "Sacrificed Rock"
)

func oldB079Enter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.OracleID, c.Power, c.Toughness = typeLine, oracle, p, tough
	owner.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("enter %s: %v", name, err)
	}
	return c.InstanceID
}

func TestOldB079EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB079Invoker, oldB079Captain, oldB079Crew, oldB079Will, oldB079Furnace, oldB079Gnostro, oldB079Blim} {
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

func TestOldB079ValakutInvokerPaysEightForThreeDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	invoker := effectstest.B21Push(g, me.ID, "Valakut Invoker", "Creature — Human Shaman", oldB079Invoker, 2, 3)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, invoker, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("activated with seven mana")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, invoker, 0, game.ActivateAbilityParams{Strict: true, Targets: target})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, opp.Life)
	}
}

func TestOldB079LightningRigCrewPingsEachOpponentAndUntapsOnAPirateSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	crew := effectstest.B21Push(g, me.ID, "Lightning-Rig Crew", "Creature — Goblin Pirate", oldB079Crew, 0, 5)
	effectstest.Unsick(g, crew)
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Life
	}
	effectstest.B16Activate(t, g, me.ID, crew, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := before[s.ID]
		if s.ID != me.ID {
			want--
		}
		if s.Life != want {
			t.Errorf("seat life %d, want %d", s.Life, want)
		}
	}
	if !effectstest.B16Tapped(t, g, crew) {
		t.Fatal("the Crew did not tap")
	}
	effectstest.CastWithCost(t, g, "Plain Sorcery", "Sorcery", "{1}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, crew) {
		t.Error("a non-Pirate spell untapped the Crew")
	}
	effectstest.CastWithCost(t, g, "Pirate Ship", "Creature — Pirate", "{1}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, crew) {
		t.Error("a Pirate spell did not untap the Crew")
	}
}

func TestOldB079EmberwildeCaptainMakesYouMonarchAndPunishesAnAttackingOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	// Seat 0 (the opponent) is the active player and attacks seat 1.
	oldB079Enter(t, g, me, "Emberwilde Captain", "Creature — Djinn Pirate", oldB079Captain, 4, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want me", g.Monarch)
	}
	attacker := effectstest.B21Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	other := effectstest.B21Push(g, opp.ID, "Their Other Bear", "Creature — Bear", "", 2, 2)
	effectstest.Unsick(g, attacker)
	effectstest.Unsick(g, other)
	effectstest.Msc18Library(opp, 6)
	effectstest.D2dSetHand(opp, 3)
	life := opp.Life
	effectstest.DeclareAttack(t, g, me.ID, attacker, other)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("attacker's life %d → %d, want -3 (three cards in hand, one trigger for two attackers)", life, opp.Life)
	}
}

func TestOldB079EmberwildeCaptainDoesNothingWithoutTheCrown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB079Enter(t, g, me, "Emberwilde Captain", "Creature — Djinn Pirate", oldB079Captain, 4, 2)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.Monarch = opp.ID })
	attacker := effectstest.B21Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.Unsick(g, attacker)
	effectstest.D2dSetHand(opp, 3)
	life := opp.Life
	effectstest.DeclareAttack(t, g, me.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life %d → %d, want unchanged while I am not the monarch", life, opp.Life)
	}
}

func TestOldB079FurnaceCelebrationPaysTwoWhenYouSacrificeAnotherPermanent(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		oldB079Enter(t, g, me, "Furnace Celebration", "Enchantment", oldB079Furnace, 0, 0)
		rock := effectstest.B21Push(g, me.ID, oldB079Sacrifice, "Artifact", "", 0, 0)
		victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 4)
		effectstest.BroSac(t, g, me.ID, rock)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPickTargetIfOpen(t, g, victim)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.B09AddColorless(me, 2)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 2
		}
		if got := effectstest.DamageOn(t, g, victim); got != want {
			t.Errorf("pay=%v: damage = %d, want %d", pay, got, want)
		}
	}
}

func TestOldB079GnostroXIsTheSpellsYouCastThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	gnostro := effectstest.B21Push(g, me.ID, "Gnostro, Voice of the Crags", "Legendary Creature — Chimera", oldB079Gnostro, 3, 3)
	effectstest.Unsick(g, gnostro)
	effectstest.CastWithCost(t, g, "First Spell", "Sorcery", "{1}", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastWithCost(t, g, "Second Spell", "Sorcery", "{1}", "")
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, gnostro, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{2}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2 for two spells cast", life, me.Life)
	}
}

func TestOldB079GnostroDealsXToATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	gnostro := effectstest.B21Push(g, me.ID, "Gnostro, Voice of the Crags", "Legendary Creature — Chimera", oldB079Gnostro, 3, 3)
	effectstest.Unsick(g, gnostro)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 3)
	effectstest.CastWithCost(t, g, "First Spell", "Sorcery", "{1}", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastWithCost(t, g, "Second Spell", "Sorcery", "{1}", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastWithCost(t, g, "Third Spell", "Sorcery", "{1}", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16Activate(t, g, me.ID, gnostro, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{1}, Targets: effectstest.CardTarget(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("three damage did not kill a 2/3")
	}
}

func TestOldB079KamahlsWillOneModeWithoutACommanderBothWithOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	land := effectstest.B21Push(g, me.ID, "My Forest", "Basic Land — Forest", "", 0, 0)
	bear := effectstest.B21Push(g, me.ID, "My Bear", "Creature — Bear", "", 3, 3)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Giant", 5, 4)
	both := []game.TargetRef{{Kind: game.TargetCard, ID: land, Mode: 0}, {Kind: game.TargetCard, ID: victim, Mode: 1}}
	if err := effectstest.B22TryModal(g, oldB079Will, []int{0, 1}, both); err == nil {
		t.Fatal("chose both modes without a commander")
	}
	effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 1, 1)
	if err := effectstest.B22TryModal(g, oldB079Will, []int{0, 1}, both); err != nil {
		t.Fatalf("both modes with a commander: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The Forest is a 1/1 haste, vigilance, indestructible Elemental land creature now:
	// 3 (Bear) + 1 (Commander) + 1 (Forest) = 5 damage, against 4 toughness.
	if g.Battlefield.Contains(victim) {
		t.Error("the victim survived damage from the Bear, the Commander and the animated land")
	}
	if effectstest.EffectivePower(t, g, land) != 1 || !effectstest.HasEffectiveKeyword(t, g, land, "haste") ||
		!effectstest.HasEffectiveKeyword(t, g, land, "indestructible") || !effectstest.HasEffectiveKeyword(t, g, land, "vigilance") {
		t.Error("the land is not a 1/1 with vigilance, indestructible and haste")
	}
	_ = bear
}

func TestOldB079BlimGivesAPermanentThenEachPlayerLosesAndDiscardsPerPermanentTheyDontOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	blim := effectstest.B21Push(g, me.ID, "Blim, Comedic Genius", "Legendary Creature — Imp", oldB079Blim, 4, 3)
	effectstest.Unsick(g, blim)
	gift := effectstest.B21Push(g, me.ID, "Gift", "Artifact", "", 0, 0)
	stolen := effectstest.B21Push(g, opp.ID, "Stolen", "Artifact", "", 0, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == stolen {
				g.Battlefield.Cards[i].Controller = me.ID
			}
		}
	})
	effectstest.Msc18Library(me, 4)
	effectstest.D2dSetHand(me, 4)
	effectstest.D2dSetHand(opp, 4)
	myLife, oppLife := me.Life, opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, blim)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.AnswerPickTargetIfOpen(t, g, gift)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.BattlefieldCardCopy(t, g, gift); c.Controller != opp.ID {
		t.Fatalf("Gift controller = %v, want the damaged player", c.Controller)
	}
	// I control "Stolen" (not mine): 1. The damaged player controls "Gift" (not theirs): 1.
	// Blim dealt 4 combat damage to the opponent, so their life is 4 lower before the loss.
	if me.Life != myLife-1 || opp.Life != oppLife-4-1 {
		t.Errorf("life me %d → %d, opp %d → %d, want -1 and -5", myLife, me.Life, oppLife, opp.Life)
	}
	for _, p := range []*game.Player{me, opp} {
		if effectstest.DiscardOwed(g, p.ID) != 1 {
			t.Errorf("a player owes %d discards, want 1", effectstest.DiscardOwed(g, p.ID))
		}
	}
}
