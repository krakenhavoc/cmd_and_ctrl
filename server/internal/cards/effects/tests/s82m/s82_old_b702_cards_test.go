package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b702_cards_test.go — Starter 1999 cards, S82 OLD batch 702 (#5620),
// played through the real catalog entries.

const (
	o702DarkOffering    = "102356a9-0de5-4034-b121-034da5deb4f2"
	o702PathOfPeace     = "b7593cf8-4dcb-473b-a2ef-180fffe66738"
	o702HollowDogs      = "16aa2e9e-bfe8-4b40-a1db-bb1da2ede849"
	o702GoblinGeneral   = "5fab0d9c-ccf7-4c6a-94b4-959be139dad8"
	o702SylvanBasilisk  = "4f348e7a-fe13-4621-bd88-5d2fb5b597f2"
	o702SylvanYeti      = "6493960b-4bbb-4ef0-8efb-abb122b1697d"
	o702DakmorPlague    = "a6cc516d-5a4b-455a-ab16-367d64215527"
	o702JaggedLightning = "b5d5cfc0-17e0-42d1-b5bd-9f27b11191f9"
	o702RenewingTouch   = "1bd5fd97-f898-4e17-9e8f-5fa93b9eef71"
	o702PsychicTransfer = "5554d422-6513-4c7e-866d-9d5f0403fc80"
	o702IngeniousThief  = "73ea2949-5812-478d-8f09-00743ce4d40f"
	o702Bargain         = "a0dd88f6-6e36-40ce-bac2-a0db2b0117b6"
)

func TestS82OldB702CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o702DarkOffering: "Dark Offering", o702PathOfPeace: "Path of Peace", o702HollowDogs: "Hollow Dogs",
		o702GoblinGeneral: "Goblin General", o702SylvanBasilisk: "Sylvan Basilisk", o702SylvanYeti: "Sylvan Yeti",
		o702DakmorPlague: "Dakmor Plague", o702JaggedLightning: "Jagged Lightning", o702RenewingTouch: "Renewing Touch",
		o702PsychicTransfer: "Psychic Transfer", o702IngeniousThief: "Ingenious Thief", o702Bargain: "Bargain",
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

func TestS82OldB702DarkOfferingDestroysANonblackCreatureAndGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Black Bear", TypeLine: "Creature — Bear",
		Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	if !effectstest.B03CastRefused(t, g, "Dark Offering", "Sorcery", o702DarkOffering,
		game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(black)}) {
		t.Fatal("Dark Offering targeted a black creature")
	}
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Dark Offering", "Sorcery", o702DarkOffering, effectstest.CardRefs(green))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the nonblack creature survived")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestS82OldB702PathOfPeaceGivesTheOwnerTheLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID})
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Path of Peace", "Sorcery", o702PathOfPeace, effectstest.CardRefs(stolen))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(stolen) {
		t.Error("the creature survived")
	}
	if opp.Life != theirs+4 {
		t.Errorf("the owner's life %d, want %d", opp.Life, theirs+4)
	}
	if me.Life != mine {
		t.Errorf("the controller's life changed to %d", me.Life)
	}
}

func TestS82OldB702HollowDogsGetsPlusTwoWhenAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dog := effectstest.B12Push(g, me.ID, "Hollow Dogs", "Creature — Phyrexian Zombie Dog", o702HollowDogs, 3, 3)
	effectstest.BrAttack(t, g, dog)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, dog); p != 5 {
		t.Errorf("attacking Dogs have power %d, want 5", p)
	}
}

func TestS82OldB702GoblinGeneralPumpsGoblinsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	general := effectstest.B12Push(g, me.ID, "Goblin General", "Creature — Goblin Warrior", o702GoblinGeneral, 1, 1)
	gob := effectstest.B12Creature(g, me.ID, "Other Goblin", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	effectstest.BrAttack(t, g, general)
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{general: 2, gob: 2, bear: 2, theirs: 1} {
		if p, _ := effectstest.PtOf(t, g, id); p != want {
			t.Errorf("power %d, want %d", p, want)
		}
	}
}

func TestS82OldB702SylvanBasiliskDestroysEveryBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	basilisk := effectstest.B12Push(g, me.ID, "Sylvan Basilisk", "Creature — Basilisk", o702SylvanBasilisk, 2, 4)
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, opp.ID, "Blocker B", "Creature — Bear", 1, 9)
	effectstest.BrAttack(t, g, basilisk)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareBlocker(id, basilisk); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a blocker survived the Basilisk's trigger")
	}
	if !g.Battlefield.Contains(basilisk) {
		t.Error("the Basilisk died")
	}
}

func TestS82OldB702SylvanYetiPowerIsTheCardsInYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	yeti := effectstest.CastCreatureCard(t, g, "Sylvan Yeti", "Creature — Yeti", o702SylvanYeti, 0, 4)
	if p := effectstest.EffectivePowerWithCounters(t, g, yeti); p != effectstest.HandSize(me) {
		t.Errorf("power %d, want hand size %d", p, effectstest.HandSize(me))
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, yeti); p != effectstest.HandSize(me) {
		t.Errorf("power %d after a draw, want %d", p, effectstest.HandSize(me))
	}
}

func TestS82OldB702DakmorPlagueHitsEveryCreatureAndPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 4, 4)
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Dakmor Plague", "Sorcery", o702DakmorPlague, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !g.Battlefield.Contains(big) {
		t.Errorf("3 damage: small survives %v, big survives %v", g.Battlefield.Contains(small), g.Battlefield.Contains(big))
	}
	if me.Life != mine-3 || opp.Life != theirs-3 {
		t.Errorf("life %d / %d, want %d / %d", me.Life, opp.Life, mine-3, theirs-3)
	}
}

func TestS82OldB702JaggedLightningNeedsTwoTargetsAndHitsBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 3)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 4)
	if !effectstest.B03CastRefused(t, g, "Jagged Lightning", "Sorcery", o702JaggedLightning,
		game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(a)}) {
		t.Fatal("Jagged Lightning was cast with one target")
	}
	effectstest.CastCatalogSpell(t, g, "Jagged Lightning", "Sorcery", o702JaggedLightning, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("the 2/3 survived 3 damage")
	}
	if !g.Battlefield.Contains(b) {
		t.Error("the 2/4 died to 3 damage")
	}
}

func TestS82OldB702RenewingTouchShufflesChosenCreatureCardsIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c1 := effectstest.SsCard(me.Graveyard, me.ID, "Dead Bear", "Creature — Bear", 2, 2)
	c2 := effectstest.SsCard(me.Graveyard, me.ID, "Dead Elf", "Creature — Elf", 1, 1)
	spell := effectstest.SsCard(me.Graveyard, me.ID, "Dead Spell", "Sorcery", 0, 0)
	if !effectstest.B03CastRefused(t, g, "Renewing Touch", "Sorcery", o702RenewingTouch,
		game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(spell)}) {
		t.Fatal("Renewing Touch targeted a noncreature card")
	}
	effectstest.CastCatalogSpell(t, g, "Renewing Touch", "Sorcery", o702RenewingTouch, effectstest.CardRefs(c1, c2))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{c1, c2} {
		if !me.Library.Contains(id) || me.Graveyard.Contains(id) {
			t.Error("a chosen creature card did not go to the library")
		}
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the untargeted card left the graveyard")
	}
}

func TestS82OldB702PsychicTransferExchangesOnlyWithinFive(t *testing.T) {
	for _, tc := range []struct {
		mine, theirs  int
		wantMine, opp int
	}{{20, 17, 17, 20}, {20, 25, 25, 20}, {20, 26, 20, 26}, {20, 14, 20, 14}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		g.WithWriteLock(func() { me.Life, opp.Life = tc.mine, tc.theirs })
		effectstest.CastCatalogSpell(t, g, "Psychic Transfer", "Sorcery", o702PsychicTransfer, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != tc.wantMine || opp.Life != tc.opp {
			t.Errorf("%d vs %d: got %d / %d, want %d / %d", tc.mine, tc.theirs, me.Life, opp.Life, tc.wantMine, tc.opp)
		}
	}
}

func TestS82OldB702IngeniousThiefLooksAtTheTargetsHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	card := effectstest.SsCard(opp.Hand, opp.ID, "Secret", "Sorcery", 0, 0)
	effectstest.CastCreatureCard(t, g, "Ingenious Thief", "Creature — Human Rogue", o702IngeniousThief, 1, 1)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	var known bool
	g.WithWriteLock(func() {
		for _, c := range opp.Hand.Cards {
			if c.InstanceID == card {
				known = c.KnownBy[me.ID]
			}
		}
	})
	if !known {
		t.Error("the controller did not see the target's hand")
	}
	spec, _ := effects.Lookup(o702IngeniousThief)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flying" {
		t.Errorf("printed keywords %v, want flying", spec.PrintedKeywords)
	}
}

func TestS82OldB702BargainDrawsTheOpponentACardAndGainsSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if !effectstest.B03CastRefused(t, g, "Bargain", "Sorcery", o702Bargain,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}}) {
		t.Fatal("Bargain targeted its caster")
	}
	hand, life := effectstest.HandSize(opp), me.Life
	effectstest.CastCatalogSpell(t, g, "Bargain", "Sorcery", o702Bargain, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(opp) != hand+1 {
		t.Errorf("the opponent's hand %d, want %d", effectstest.HandSize(opp), hand+1)
	}
	if me.Life != life+7 {
		t.Errorf("life %d, want %d", me.Life, life+7)
	}
}
