package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB078PryingEyes = "06b60766-a1fb-4ed0-93f6-8423008f75b4"
	oldB078Herald     = "1b3d7411-7a25-4353-83c0-d09525a4fa08"
	oldB078Galestrike = "3006d992-2714-4264-a4c3-953b67dc7347"
	oldB078Wyleth     = "35bef4d3-4473-4265-86b0-f1393c82fd1d"
	oldB078Lore       = "47a9df32-7588-4430-a40e-56a9041e0360"
	oldB078Zedruu     = "514179c0-a50b-4565-8e71-9dad256edd85"
	oldB078Golem      = "75269a31-3819-4410-8bb7-10de16810a5e"
	oldB078Magus      = "ff922641-142a-4774-be3e-36a5183fce2b"
)

func oldB078Enter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.OracleID, c.Power, c.Toughness = typeLine, oracle, p, tough
	owner.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("enter %s: %v", name, err)
	}
	return c.InstanceID
}

func TestOldB078EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB078PryingEyes, oldB078Herald, oldB078Galestrike, oldB078Wyleth, oldB078Lore, oldB078Zedruu, oldB078Golem, oldB078Magus} {
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

func TestOldB078PryingEyesDrawsFourThenDiscardsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 8)
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Prying Eyes", "Instant", oldB078PryingEyes, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The test harness puts Prying Eyes into the hand and the cast takes it out.
	if got := me.Hand.Size(); got != before+4 {
		t.Fatalf("hand after the four draws = %d, want %d", got, before+4)
	}
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil || c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Fatalf("no \"discard two cards\" prompt: %+v", c)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand = %d, want %d", got, before+2)
	}
}

func TestOldB078GalestrikeBouncesATappedCreatureAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 4)
	tapped := effectstest.PushVanillaCreature(g, opp.ID, "Tapped Bear", 2, 2)
	untapped := effectstest.PushVanillaCreature(g, opp.ID, "Untapped Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tapped {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Galestrike", "Instant", oldB078Galestrike, []game.TargetRef{{Kind: game.TargetCard, ID: untapped}}); err == nil {
		t.Fatal("Galestrike targeted an untapped creature")
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Galestrike", "Instant", oldB078Galestrike, []game.TargetRef{{Kind: game.TargetCard, ID: tapped}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) || !opp.Hand.Contains(tapped) {
		t.Error("the tapped creature was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature moved")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d (drew one)", got, hand+1)
	}
}

func TestOldB078FlamekinHeraldGivesCommandersCascadeAndNoOneElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCheapLibrary(g, me, 6)
	effectstest.PushCatalogPermanent(g, me.ID, "Flamekin Herald", "Creature — Elemental Wizard", oldB078Herald, false)

	effectstest.CastWithCost(t, g, "Ordinary Spell", "Sorcery", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 0 {
		t.Fatalf("a non-commander spell got %d cascade offers, want 0", got)
	}

	cmdr := game.Card{InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Elf", ManaCost: "{3}{G}",
		Owner: me.ID, Controller: me.ID, IsCommander: true, Power: 3, Toughness: 3}
	me.Hand.PushTop(cmdr)
	if err := g.CastSpell(me.ID, cmdr.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast commander: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Fatalf("a commander spell got %d cascade offers, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
}

func TestOldB078WylethDrawsForEachAuraAndEquipmentAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wyleth := effectstest.B21Push(g, me.ID, "Wyleth, Soul of Steel", "Legendary Creature — Human Warrior", oldB078Wyleth, 2, 2)
	effectstest.Unsick(g, wyleth)
	sword := effectstest.B12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	pacifism := effectstest.B12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	other := effectstest.B12Permanent(g, me.ID, "Helmet", "Artifact")
	for _, id := range []uuid.UUID{sword, pacifism} {
		var err error
		g.WithWriteLock(func() { err = g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: wyleth}) })
		if err != nil {
			t.Fatalf("attach: %v", err)
		}
	}
	_ = other
	effectstest.Msc18Library(me, 6)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, wyleth)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want +2 (one Equipment and one Aura attached)", hand, got)
	}
}

func TestOldB078DemonicLoreDrawsThreeAndCostsTwoLifePerCardAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc18Library(me, 6)
	hand := me.Hand.Size()
	oldB078Enter(t, g, me, "Demonic Lore", "Enchantment", oldB078Lore, 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Fatalf("hand %d → %d, want +3", hand, got)
	}
	life := me.Life
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if want := life - 2*me.Hand.Size(); me.Life != want {
		t.Errorf("life = %d, want %d (2 per card in hand)", me.Life, want)
	}
}

func TestOldB078ZedruuCountsPermanentsYouOwnThatOpponentsControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	effectstest.B12Push(g, me.ID, "Zedruu the Greathearted", "Legendary Creature — Minotaur Monk", oldB078Zedruu, 2, 4)
	gift := effectstest.B21Push(g, me.ID, "My Gift", "Artifact", "", 0, 0)
	effectstest.B21Push(g, me.ID, "Mine Still", "Artifact", "", 0, 0)
	effectstest.B21Push(g, opp.ID, "Their Own", "Artifact", "", 0, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == gift {
				g.Battlefield.Cards[i].Controller = opp.ID
			}
		}
	})
	effectstest.Msc18Library(me, 6)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	life, hand := me.Life, me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d → %d, hand %d → %d, want +1 each (one permanent I own is theirs)", life, me.Life, hand, me.Hand.Size())
	}
}

func TestOldB078ZedruuGivesAPermanentToTargetOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	zedruu := oldB078Enter(t, g, me, "Zedruu the Greathearted", "Legendary Creature — Minotaur Monk", oldB078Zedruu, 2, 4)
	rock := effectstest.B21Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	effectstest.FloatMana(t, g, me, "{U}{R}{W}")
	effectstest.B16Activate(t, g, me.ID, zedruu, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: rock}}})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.BattlefieldCardCopy(t, g, rock); c.Controller != opp.ID {
		t.Errorf("controller = %v, want the opponent", c.Controller)
	}
}

func TestOldB078MagusSacrificesItselfAndAnotherGreenCreatureToFetchOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	magus := effectstest.B21Push(g, me.ID, "Magus of the Order", "Creature — Human Wizard", oldB078Magus, 3, 3, "G")
	effectstest.Unsick(g, magus)
	green := effectstest.B21Push(g, me.ID, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	red := effectstest.B21Push(g, me.ID, "Red Bear", "Creature — Bear", "", 2, 2, "R")
	fetched := effectstest.SeedLibraryTop(me, "Fetched Elf", "Creature — Elf", "{G}")
	g.WithWriteLock(func() {
		for i := range me.Library.Cards {
			if me.Library.Cards[i].InstanceID == fetched {
				me.Library.Cards[i].Colors = []string{"G"}
			}
		}
	})
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, magus, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{red}}); err == nil {
		t.Fatal("sacrificed a non-green creature for the cost")
	}
	effectstest.B16Activate(t, g, me.ID, magus, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{green}})
	if g.Battlefield.Contains(magus) || g.Battlefield.Contains(green) {
		t.Error("the Magus and the other green creature were not both sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Fetched Elf")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fetched) {
		t.Error("the green creature card was not put onto the battlefield")
	}
}

func TestOldB078HowlingGolemMakesEveryPlayerDrawWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	golem := effectstest.B21Push(g, me.ID, "Howling Golem", "Artifact Creature — Golem", oldB078Golem, 2, 3)
	effectstest.Unsick(g, golem)
	for _, s := range g.Seats {
		effectstest.Msc18Library(s, 5)
	}
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, golem)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != mine+1 || opp.Hand.Size() != theirs+1 {
		t.Errorf("hands %d→%d and %d→%d, want +1 each", mine, me.Hand.Size(), theirs, opp.Hand.Size())
	}
}
