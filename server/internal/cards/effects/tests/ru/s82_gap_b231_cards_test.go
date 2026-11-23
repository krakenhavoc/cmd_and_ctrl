package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander Legends: Battle for Baldur's Gate, GAP batch 231 (S82, #5948):
// the sorceries and Elminster.

const (
	gapElminsterOracle          = "56773795-bd8b-4ac4-8f2c-a9d6275989bb"
	gapArmsOfHadarOracle        = "c15e5cb8-b94c-4157-9327-410e66606b82"
	gapAtrociousExperimentOracl = "9f1c5040-ac46-4e8f-a5ff-8205e13a8893"
	gapCloudkillOracle          = "1941239b-e131-4bde-b493-1c8c805b6a3c"
	gapEldritchPactOracle       = "281ea9b6-055a-4274-90b3-c74311a6f863"
	gapMyrkulsEdictOracle       = "933afaa3-3b23-454c-9465-8f2ba9652b4f"
	gapSummonUndeadOracle       = "c86ff0e3-5d6e-4981-a2da-8da70b9daeb0"
	gapThwartTheGraveOracle     = "b545e26f-2034-409c-80cc-51b72e75e6e4"
	gapBarroomBrawlOracle       = "d1ce7fcf-42f8-4b35-9ba1-55033c65c43e"
	gapDruidicRitualOracle      = "79edf527-2de1-433a-9cac-c91f461b1ebf"
	gapMonsterManualOracle      = "6d4e7943-1fdf-4085-8af0-658badbd0cb8"
)

func TestS82GapB231CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapElminsterOracle:            "Elminster",
		gapArmsOfHadarOracle:          "Arms of Hadar",
		gapAtrociousExperimentOracl:   "Atrocious Experiment",
		gapCloudkillOracle:            "Cloudkill",
		gapEldritchPactOracle:         "Eldritch Pact",
		gapMyrkulsEdictOracle:         "Myrkul's Edict",
		gapSummonUndeadOracle:         "Summon Undead",
		gapThwartTheGraveOracle:       "Thwart the Grave",
		gapBarroomBrawlOracle:         "Barroom Brawl",
		gapDruidicRitualOracle:        "Druidic Ritual",
		gapMonsterManualOracle:        "Monster Manual",
		gapMonsterManualOracle + "#1": "Zoological Study",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

// gapB231Cast puts the card in the active seat's hand and casts it strictly for free.
func gapB231Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	effectstest.D5681Spell(g, me, id, name, typeLine, "", oracle)
	params.Strict = true
	return g.CastSpell(me.ID, id, params)
}

func gapB231Library(p *game.Player, n int, typeLine string) {
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Library Card", TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	}
}

func TestGapArmsOfHadarShrinksTheTargetPlayersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, other := effectstest.BlbB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	third := effectstest.B12Creature(g, other.ID, "Third", "Creature — Bear", 3, 3)
	if err := gapB231Cast(t, g, "Arms of Hadar", "Sorcery", gapArmsOfHadarOracle, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.B12Creature(g, opp.ID, "Late", "Creature — Bear", 3, 3)
	for id, want := range map[uuid.UUID]int{mine: 3, theirs: 1, third: 3, late: 3} {
		if _, got := effectstest.MkmB1PT(t, g, id); got != want {
			t.Errorf("toughness = %d, want %d", got, want)
		}
	}
}

func TestGapAtrociousExperimentMillsDrawsAndLoses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	_, opp, _ := effectstest.BlbB1Seats(g)
	gapB231Library(opp, 6, "Creature — Bear")
	grave, hand, life := opp.Graveyard.Size(), opp.Hand.Size(), opp.Life
	if err := gapB231Cast(t, g, "Atrocious Experiment", "Sorcery", gapAtrociousExperimentOracl, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - grave; got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
	if got := opp.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if got := life - opp.Life; got != 2 {
		t.Errorf("lost %d life, want 2", got)
	}
}

func TestGapEldritchPactDrawsAndLosesPerGraveyardCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	gapB231Library(opp, 8, "Creature — Bear")
	for i := 0; i < 3; i++ {
		effectstest.PushGraveyardCard(g, opp, "Old Card")
	}
	hand, life := opp.Hand.Size(), opp.Life
	if err := gapB231Cast(t, g, "Eldritch Pact", "Sorcery", gapEldritchPactOracle, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
	if got := life - opp.Life; got != 3 {
		t.Errorf("lost %d, want 3", got)
	}
	_ = me
}

func TestGapCloudkillShrinksEveryCreatureByTheGreatestCommanderManaValue(t *testing.T) {
	for _, zone := range []string{"none", "battlefield", "command"} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp, _ := effectstest.BlbB1Seats(g)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 5, 5)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 5, 5)
		want := 5
		switch zone {
		case "battlefield":
			c := game.Card{InstanceID: uuid.New(), Name: "Leader", TypeLine: "Legendary Creature — Bear", ManaCost: "{1}{G}{G}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID, IsCommander: true}
			g.WithWriteLock(func() { g.Battlefield.PushTop(c) })
			want = 2
		case "command":
			g.WithWriteLock(func() {
				me.Command.PushTop(game.Card{InstanceID: uuid.New(), Name: "Leader", TypeLine: "Legendary Creature — Bear", ManaCost: "{2}{G}{G}", Owner: me.ID, Controller: me.ID, IsCommander: true})
			})
			want = 1
		}
		if err := gapB231Cast(t, g, "Cloudkill", "Sorcery", gapCloudkillOracle, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		for _, id := range []uuid.UUID{mine, theirs} {
			if _, got := effectstest.MkmB1PT(t, g, id); got != want {
				t.Errorf("commander in %s: toughness %d, want %d", zone, got, want)
			}
		}
	}
}

func TestGapMyrkulsEdictByTheRoll(t *testing.T) {
	for _, tc := range []struct {
		natural int
		sacs    int // opponents expected to lose a creature
	}{{5, 1}, {15, 3}, {20, 3}} {
		edict := uuid.New()
		g := effectstest.D5681Table(t, edict, tc.natural)
		me, opp, other := effectstest.BlbB1Seats(g)
		third := g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)]
		small := map[uuid.UUID]uuid.UUID{}
		big := map[uuid.UUID]uuid.UUID{}
		for _, p := range []*game.Player{opp, other, third} {
			small[p.ID] = effectstest.B12Creature(g, p.ID, "Small", "Creature — Bear", 1, 1)
			big[p.ID] = effectstest.B12Creature(g, p.ID, "Big", "Creature — Bear", 4, 4)
		}
		effectstest.D5681Spell(g, me, edict, "Myrkul's Edict", "Sorcery", "", gapMyrkulsEdictOracle)
		if err := g.CastSpell(me.ID, edict, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if tc.natural < 10 {
			effectstest.AnswerChoosePlayer(t, g, me.ID, other)
			effectstest.PassPriorityAroundTable(t, g)
		}
		for i := 0; i < 8; i++ {
			for _, p := range []*game.Player{opp, other, third} {
				if c := effectstest.SacrificeChoiceFor(g, p.ID); c != nil {
					effectstest.AnswerSacrifice(t, g, p.ID, small[p.ID])
					effectstest.PassPriorityAroundTable(t, g)
				}
			}
		}
		lost := 0
		for _, p := range []*game.Player{opp, other, third} {
			bigGone := !g.Battlefield.Contains(big[p.ID])
			smallGone := !g.Battlefield.Contains(small[p.ID])
			switch {
			case tc.natural == 20 && !bigGone:
				t.Errorf("natural 20: a player kept their greatest-power creature")
			case tc.natural < 20 && bigGone:
				t.Errorf("natural %d: a greatest-power creature was sacrificed by choice", tc.natural)
			}
			if bigGone || smallGone {
				lost++
			}
		}
		if lost != tc.sacs {
			t.Errorf("natural %d: %d opponents sacrificed, want %d", tc.natural, lost, tc.sacs)
		}
		if tc.natural < 10 && g.Battlefield.Contains(small[other.ID]) {
			t.Errorf("natural %d: the chosen opponent kept the creature", tc.natural)
		}
	}
}

func TestGapSummonUndeadReturnsACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gapB231Library(me, 6, "Creature — Bear")
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := gapB231Cast(t, g, "Summon Undead", "Sorcery", gapSummonUndeadOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	yes := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if yes == nil {
		t.Fatal("no mill question")
	}
	if err := g.ResolveConfirm(yes.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	grave := me.Graveyard.Size()
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if prompt == nil {
		t.Fatal("no creature pick offered")
	}
	if grave != 4 {
		t.Errorf("graveyard = %d, want 4 (one Bear plus three milled)", grave)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, bear); c == nil || c.Controller != me.ID {
		t.Errorf("the Bear is not on the battlefield under the caster's control: %+v", c)
	}
}

func TestGapDruidicRitualDeclinedTakesACreatureAndALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gapB231Library(me, 6, "Creature — Bear")
	bear, land, spell := uuid.New(), uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Dead Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Dead Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	if err := gapB231Cast(t, g, "Druidic Ritual", "Sorcery", gapDruidicRitualOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	yes := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if yes == nil {
		t.Fatal("no mill question")
	}
	if err := g.ResolveConfirm(yes.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	for i := 0; i < 4; i++ {
		p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		if p == nil {
			break
		}
		picks := []uuid.UUID{}
		for _, id := range p.ChooseCards {
			if id == bear || id == land {
				picks = append(picks, id)
			}
		}
		if err := g.ResolveChooseCards(p.ID, me.ID, picks); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !me.Hand.Contains(bear) || !me.Hand.Contains(land) {
		t.Errorf("the Bear and the Forest are not both in hand")
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the instant left the graveyard")
	}
}

func TestGapZoologicalStudyMillsFiveAndTakesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gapB231Library(me, 8, "Instant")
	bear := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: bear, Name: "Milled Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	old := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if err := gapB231Cast(t, g, "Zoological Study", "Sorcery — Adventure", gapMonsterManualOracle+"#1", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if p == nil {
		t.Fatal("no pick offered")
	}
	for _, id := range p.ChooseCards {
		if id == old {
			t.Error("a creature already in the graveyard was offered")
		}
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the milled creature is not in hand")
	}
}

func TestGapMonsterManualPutsACreatureFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	book := effectstest.PushCatalogPermanent(g, me.ID, "Monster Manual", "Artifact — Book", gapMonsterManualOracle, false)
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Hand Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if p == nil {
		t.Fatal("no put offered")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, bear) {
		t.Error("the creature card is not on the battlefield")
	}
}

func TestGapThwartTheGraveCostsLessPerPartyMemberAndReturnsBothCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Healer", "Creature — Human Cleric", 1, 1)
	effectstest.B12Creature(g, me.ID, "Thief", "Creature — Human Rogue", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	bear, cleric, rogueCard := uuid.New(), uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: cleric, Name: "Dead Cleric", TypeLine: "Creature — Human Cleric", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: rogueCard, Name: "Dead Rogue", TypeLine: "Creature — Human Rogue", Owner: me.ID, Controller: me.ID})
	id := effectstest.HandCardFull(me, "Thwart the Grave", "Sorcery", "{4}{B}{B}", gapThwartTheGraveOracle, []string{"B"})
	// Two party members (a Cleric and a Rogue): {4}{B}{B} costs {2}{B}{B}.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{G}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: bear, Slot: 0}, {Kind: game.TargetCard, ID: cleric, Slot: 1}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err == nil {
		t.Fatal("cast for three mana with a party of two")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("cast for {2}{B}{B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, bear) || !effectstest.OnBattlefieldByID(g, cleric) {
		t.Error("both targets did not return")
	}
	if !me.Graveyard.Contains(rogueCard) {
		t.Error("an untargeted card left the graveyard")
	}
	// The second clause only takes a party type.
	again := effectstest.HandCardFull(me, "Thwart the Grave", "Sorcery", "{4}{B}{B}", gapThwartTheGraveOracle, []string{"B"})
	other := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: other, Name: "Dead Bear Two", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	bad := []game.TargetRef{{Kind: game.TargetCard, ID: rogueCard, Slot: 0}, {Kind: game.TargetCard, ID: other, Slot: 1}}
	if err := g.CastSpell(me.ID, again, game.CastSpellParams{Targets: bad}); err == nil {
		t.Error("the second clause took a Bear")
	}
}

func TestGapBarroomBrawlFightsTheLeftOpponentAndThatPlayerMayCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	left := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if got := g.NearestOpponentForEffect(me.ID, game.AttackDirectionLeft); got != left.ID {
		left = g.PlayerByIDForEffect(got)
	}
	right := g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, left.ID, "Theirs", "Creature — Bear", 1, 5)
	far := effectstest.B12Creature(g, right.ID, "Far", "Creature — Bear", 1, 5)
	if err := gapB231Cast(t, g, "Barroom Brawl", "Sorcery", gapBarroomBrawlOracle, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: far, Slot: 1}},
	}); err == nil {
		t.Fatal("Barroom Brawl targeted a creature of an opponent who is not to the left")
	}
	if err := gapB231Cast(t, g, "Barroom Brawl", "Sorcery", gapBarroomBrawlOracle, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, theirs).DamageMarked; got != 3 {
		t.Errorf("their creature has %d damage, want 3", got)
	}
	if got := effectstest.FindCardAnywhere(t, g, mine).DamageMarked; got != 1 {
		t.Errorf("my creature has %d damage, want 1", got)
	}
	q := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if q == nil || q.Chooser != left.ID {
		t.Fatalf("copy question %+v, want it asked of the player to the left", q)
	}
}

func TestGapElminsterScryMakesTheNextInstantCheaperAndLooksTwoPerPlus(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	walker := effectstest.GuffbWalker(g, me.ID, "Elminster", "Legendary Planeswalker — Elminster", gapElminsterOracle, 5)
	gapB231Library(me, 6, "Creature — Bear")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry) == nil {
		t.Fatal("no scry 2 prompt")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	// The scry of two made the next instant or sorcery cost {2} less: a {2}{R} spell costs {R}.
	id := effectstest.HandCardFull(me, "Test Bolt", "Instant", "{2}{R}", "", []string{"R"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {R} after scrying 2: %v", err)
	}
	// The promise is spent by that spell.
	again := effectstest.HandCardFull(me, "Test Bolt Two", "Instant", "{2}{R}", "", []string{"R"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.CastSpell(me.ID, again, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a second instant was also cheaper")
	}
}

func TestGapElminsterMinusThreeMakesFaerieDragonsForTheExiledCardsManaValue(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	walker := effectstest.GuffbWalker(g, me.ID, "Elminster", "Legendary Planeswalker — Elminster", gapElminsterOracle, 5)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Card", TypeLine: "Creature — Bear", ManaCost: "{2}{G}{G}", Owner: me.ID, Controller: me.ID})
	effectstest.B16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(top); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the top card is in %v, want exile", z)
	}
	dragons := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == "Faerie Dragon" && c.Controller == me.ID {
			dragons++
			if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Error("a Faerie Dragon has no flying")
			}
		}
	}
	if dragons != 4 {
		t.Errorf("Faerie Dragons = %d, want 4", dragons)
	}
}
