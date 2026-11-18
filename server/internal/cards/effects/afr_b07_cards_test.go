package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	afrB07Trelasarra  = "cd10ebc9-9f6c-4da2-9a86-263902002ebc"
	afrB07Unicorn     = "be030639-cbc9-416b-b0ca-1a76439791d8"
	afrB07Roper       = "61f3e964-6bea-4543-992e-3e36404bf628"
	afrB07Bulette     = "4e690b82-9877-41f2-9dcd-cffdb6e3cd64"
	afrB07Shessra     = "fb29737c-2dfa-4139-988b-29dc8b829c08"
	afrB07DeathPriest = "af231505-bd2e-4499-9fd5-f761dde2ecef"
	afrB07Swarming    = "93ab7462-ceef-4bc7-a9fe-8af8ceb78e1c"
	afrB07PurpleWorm  = "179cc807-5150-4513-88b2-73cb05009b7d"
	afrB07Wanderer    = "614feadd-b74f-4d9c-a0a3-3a63724956af"
	afrB07Ebondeath   = "1a6fce11-e77c-41f1-9ff7-9a93b18071db"
	afrB07Book        = "046add85-952f-46ac-974a-463d2a7d3fd4"
	afrB07Wight       = "0a07a676-890f-49a2-9a2d-17d7b1dc6868"
)

func TestAfrB07EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		afrB07Trelasarra:  "Trelasarra, Moon Dancer",
		afrB07Unicorn:     "Celestial Unicorn",
		afrB07Roper:       "Lurking Roper",
		afrB07Bulette:     "Bulette",
		afrB07Shessra:     "Shessra, Death's Whisper",
		afrB07DeathPriest: "Death-Priest of Myrkul",
		afrB07Swarming:    "Skeletal Swarming",
		afrB07PurpleWorm:  "Purple Worm",
		afrB07Wanderer:    "Grim Wanderer",
		afrB07Ebondeath:   "Ebondeath, Dracolich",
		afrB07Book:        "The Book of Exalted Deeds",
		afrB07Wight:       "Wight",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func afrB07Gain(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, n) })
}

func TestAfrB07LifeGainGrowsUnicornOncePerGain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	unicorn := b12Push(g, me.ID, "Celestial Unicorn", "Creature — Unicorn", afrB07Unicorn, 3, 2)
	afrB07Gain(g, opp, 3)
	passPriorityAroundTable(t, g)
	if plusOnes(g, unicorn) != 0 {
		t.Fatal("an opponent's life gain is not yours")
	}
	afrB07Gain(g, me, 3)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, unicorn); got != 1 {
		t.Errorf("Celestial Unicorn has %d counters after one 3-life gain, want 1", got)
	}
}

func TestAfrB07LifeGainGrowsTrelasarraAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trel := b12Push(g, me.ID, "Trelasarra, Moon Dancer", "Legendary Creature — Elf Cleric", afrB07Trelasarra, 2, 2)
	afrB07Gain(g, me, 3)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, trel); got != 1 {
		t.Errorf("Trelasarra has %d counters after one 3-life gain, want 1", got)
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil {
		t.Error("Trelasarra's trigger did not scry 1")
	}
}

func TestAfrB07LurkingRoperUntapsOnLifeGainOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	roper := b12Push(g, me.ID, "Lurking Roper", "Creature — Horror", afrB07Roper, 4, 5)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(roper) })
	afrB07Gain(g, opp, 2)
	passPriorityAroundTable(t, g)
	if !isTapped(g, roper) {
		t.Fatal("an opponent's life gain untapped the Roper")
	}
	afrB07Gain(g, me, 2)
	passPriorityAroundTable(t, g)
	if isTapped(g, roper) {
		t.Error("gaining life did not untap the Roper")
	}
}

func TestAfrB07BuletteGrowsOnlyAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bul := b12Push(g, me.ID, "Bulette", "Creature — Beast", afrB07Bulette, 3, 3)
		bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		if died {
			killCreature(t, g, opp.ID, bear)
		}
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		want := 0
		if died {
			want = 1
		}
		if got := plusOnes(g, bul); got != want {
			t.Errorf("a creature died %v: %d counters, want %d", died, got, want)
		}
	}
}

func TestAfrB07ShessraEndStepPaysTwoLifeForACard(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		b12Push(g, me.ID, "Shessra, Death's Whisper", "Legendary Creature — Human Elf Warlock", afrB07Shessra, 1, 3)
		bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		if died {
			killCreature(t, g, opp.ID, bear)
		}
		life, hand := me.Life, me.Hand.Size()
		advanceTo(t, g, game.StepEnd)
		if !died {
			passPriorityAroundTable(t, g)
			if latestConfirmFor(g, me.ID) != nil || me.Life != life || me.Hand.Size() != hand {
				t.Error("no creature died, yet Shessra asked or paid")
			}
			continue
		}
		passPriorityAroundTable(t, g)
		answerMayChoice(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
		if me.Life != life-2 || me.Hand.Size() != hand+1 {
			t.Errorf("life %d→%d, hand %d→%d; want -2 life, +1 card", life, me.Life, hand, me.Hand.Size())
		}
	}
}

func TestAfrB07ShessraEntersMakesATargetCreatureBlock(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Shessra, Death's Whisper", "Legendary Creature — Human Elf Warlock", afrB07Shessra, 1, 3)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !mkmB1MustBlock(t, g, bear) {
		t.Error("the target has no blocks-if-able requirement")
	}
}

func TestAfrB07DeathPriestAnthemAndEndStepSkeleton(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Death-Priest of Myrkul", "Creature — Tiefling Cleric", afrB07DeathPriest, 2, 2)
	zombie := b12Creature(g, me.ID, "Zed", "Creature — Zombie", 2, 2)
	vamp := b12Creature(g, me.ID, "Vlad", "Creature — Vampire", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Zed", "Creature — Zombie", 2, 2)
	for id, want := range map[uuid.UUID]int{zombie: 3, vamp: 3, bear: 2, theirs: 2} {
		if got := effectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
	advanceToMain(t, g)
	killCreature(t, g, me.ID, bear)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	floatMana(t, g, me, "{B}")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Skeleton" && c.Controller == me.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Skeleton tokens after paying {1}, want 1", n)
	}
}

func TestAfrB07SkeletalSwarmingBuffsSkeletonsAndMakesTappedTokens(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		b12Push(g, me.ID, "Skeletal Swarming", "Enchantment", afrB07Swarming, 0, 0)
		s1 := b12Creature(g, me.ID, "Skel A", "Creature — Skeleton", 1, 1)
		s2 := b12Creature(g, me.ID, "Skel B", "Creature — Skeleton", 1, 1)
		b12Creature(g, opp.ID, "Their Skel", "Creature — Skeleton", 1, 1)
		bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		if got := effectivePower(t, g, s1); got != 2 {
			t.Fatalf("a Skeleton with one other Skeleton has power %d, want 2", got)
		}
		assertKeywords(t, g, s2, "trample")
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(s1); _ = g.TapTargetForEffect(s2) })
		advanceToMain(t, g)
		if died {
			killCreature(t, g, opp.ID, bear)
		}
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		tokens := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Skeleton" && c.Controller == me.ID {
				tokens++
				if !c.Tapped {
					t.Error("the Skeleton token did not enter tapped")
				}
			}
		}
		want := 1
		if died {
			want = 2
		}
		if tokens != want {
			t.Errorf("a creature died %v: %d tokens, want %d", died, tokens, want)
		}
	}
}

func TestAfrB07PurpleWormCostsTwoLessAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		worm := trekB13Card(g, me.Hand, me, "Purple Worm", "Creature — Worm", "{5}{G}{G}", afrB07PurpleWorm, 8, 7)
		bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		if died {
			b27Kill(g, bear)
			g.RunStateChecksForTest()
		}
		err := trekB13Cast(t, g, me, worm, "{G}{G}{G}{G}{G}", game.CastSpellParams{})
		if died && err != nil {
			t.Errorf("a creature died, but five mana could not cast the Worm: %v", err)
		}
		if !died && err == nil {
			t.Error("five mana cast the Worm with no creature having died")
		}
	}
}

func TestAfrB07GrimWandererOnlyAfterADeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	w := trekB13Card(g, me.Hand, me, "Grim Wanderer", "Creature — Goblin Warlock", "{1}{B}", afrB07Wanderer, 5, 3)
	if err := trekB13Cast(t, g, me, w, "{B}{B}", game.CastSpellParams{}); err == nil {
		t.Fatal("cast with no creature having died")
	}
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b27Kill(g, bear)
	g.RunStateChecksForTest()
	if err := trekB13Cast(t, g, me, w, "", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast after a death: %v", err)
	}
}

func TestAfrB07EbondeathGraveyardCastIgnoresItsOwnName(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	eb := trekB13Card(g, me.Graveyard, me, "Ebondeath, Dracolich", "Legendary Creature — Zombie Dragon", "{2}{B}{B}", afrB07Ebondeath, 5, 2)
	floatMana(t, g, me, "{B}{B}{B}{B}")
	if err := g.CastSpell(me.ID, eb, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("cast from the graveyard with no creature having died")
	}
	copyOf := b12Creature(g, opp.ID, "Ebondeath, Dracolich", "Legendary Creature — Zombie Dragon", 5, 2)
	b27Kill(g, copyOf)
	g.RunStateChecksForTest()
	if err := g.CastSpell(me.ID, eb, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("a creature named Ebondeath, Dracolich dying opened the graveyard")
	}
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b27Kill(g, bear)
	g.RunStateChecksForTest()
	if err := g.CastSpell(me.ID, eb, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast after a Bear died: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !isTapped(g, eb) {
		t.Error("Ebondeath did not enter tapped")
	}
}

func TestAfrB07WightReplacesADamagedCreatureWithATappedZombie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wight := b12Push(g, me.ID, "Wight", "Creature — Zombie Soldier", afrB07Wight, 3, 2)
	hit := b12Creature(g, opp.ID, "Hit Bear", "Creature — Bear", 2, 3)
	dealDamageTo(t, g, wight, hit, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hit) })
	passPriorityAroundTable(t, g)
	zombies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && c.Controller == me.ID {
			zombies++
			if !c.Tapped {
				t.Error("the Zombie token did not enter tapped")
			}
		}
	}
	if zombies != 1 {
		t.Errorf("%d Zombie tokens, want 1", zombies)
	}
	if opp.Graveyard.Contains(hit) {
		t.Error("the dead creature card was not exiled")
	}
}

func TestAfrB07BookMakesAnAngelThenGrantsItCantLose(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	book := b12Push(g, me.ID, "The Book of Exalted Deeds", "Legendary Artifact — Book", afrB07Book, 0, 0)
	afrB07Gain(g, me, 2)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	angels := func() (out []uuid.UUID) {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Angel" && c.Controller == me.ID {
				out = append(out, c.InstanceID)
			}
		}
		return
	}
	if len(angels()) != 0 {
		t.Fatal("two life gained made an Angel")
	}
	_ = book
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	b12Push(g2, me2.ID, "The Book of Exalted Deeds", "Legendary Artifact — Book", afrB07Book, 0, 0)
	afrB07Gain(g2, me2, 3)
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	n := 0
	for _, c := range g2.Battlefield.Cards {
		if c.Name == "Angel" && c.Controller == me2.ID {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d Angels after gaining 3, want 1", n)
	}
	_ = opp
}

func TestAfrB07BookActivationGivesTheAngelTheGate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	book := b12Push(g, me.ID, "The Book of Exalted Deeds", "Legendary Artifact — Book", afrB07Book, 0, 0)
	angel := pushCatalogPermanent(g, me.ID, "Angel", "Creature — Angel", "", false)
	advanceToMain(t, g)
	floatMana(t, g, me, "{W}{W}{W}")
	b16Activate(t, g, me.ID, book, 0, game.ActivateAbilityParams{Targets: cardRefs(angel), Strict: true})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, angel, "enlightened"); got != 1 {
		t.Fatalf("%d enlightened counters, want 1", got)
	}
	if g.Battlefield.Contains(book) {
		t.Error("the Book was not exiled")
	}
	g.WithWriteLock(func() { me.Life = 0 })
	g.RunStateChecksForTest()
	if g.PlayerByIDForEffect(me.ID).Eliminated {
		t.Error("the Angel's controller lost at 0 life with the gate on the table")
	}
}
