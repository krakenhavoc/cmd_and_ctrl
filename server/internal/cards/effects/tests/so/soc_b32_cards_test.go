package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b32_cards_test.go — Secrets of Strixhaven (sos), batch 32 (#4456).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	socB32PoxPlagueOracle    = "c505d304-9428-4532-87a2-167c69f16969"
	socB32OffenseOracle      = "99678cc9-59c8-4358-b847-8c38ffb4a050"
	socB32ScrollsmithOracle  = "a02f9f42-4c3a-4a97-8667-68ba4617f193"
	socB32ResearchOracle     = "51af3da5-0970-436e-baa6-3171889a40e0"
	socB32CharmOracle        = "c3c25a2c-f71c-4cb7-9117-0bcda9248bfd"
	socB32PrismariOracle     = "258f5045-3aa7-4b95-a15f-3172577294a8"
	socB32ProcrastinateOrcl  = "6f8fa68e-e961-4ae1-b27d-22d87e90f26b"
	socB32GazeOracle         = "a7fe7621-9bb1-4d0d-8b50-c179c685d79d"
	socB32PterafractylOracle = "cca03b76-dd3c-4228-b97a-f7fa20819b2e"
	socB32PullOracle         = "c5373026-8b43-42a7-aa69-a1b0f1619795"
	socB32BearType           = "Creature — Bear"
)

func TestSocB32EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB32PoxPlagueOracle:    "Pox Plague",
		socB32OffenseOracle:      "Practiced Offense",
		socB32ScrollsmithOracle:  "Practiced Scrollsmith",
		socB32ResearchOracle:     "Primary Research",
		socB32CharmOracle:        "Prismari Charm",
		socB32PrismariOracle:     "Prismari, the Inspiration",
		socB32ProcrastinateOrcl:  "Procrastinate",
		socB32GazeOracle:         "Proctor's Gaze",
		socB32PterafractylOracle: "Pterafractyl",
		socB32PullOracle:         "Pull from the Grave",
	}
	for oracle, name := range want {
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

// socB32Settle passes priority to an empty stack, answering a target pick
// addressed to chooser with `pick` (a targeted trigger with one legal target
// may still open one).
func socB32Settle(t *testing.T, g *game.Game, chooser, pick uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		if effectstest.LatestPickTarget(g, chooser) != nil && pick != uuid.Nil {
			effectstest.PickCard(t, g, chooser, pick)
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
}

// --- Pox Plague ---------------------------------------------------------------

func TestSocB32PoxPlagueHalvesLifeHandAndPermanentsRoundingDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Turn.ActiveSeat
	lives := []int{20, 15, 9, 1}
	wantLife := []int{10, 8, 5, 1}
	hands := []int{7, 5, 2, 1}
	discards := []int{3, 2, 1, 0}
	permanents := []int{5, 4, 3, 1}
	owed := []int{2, 2, 1, 0}

	ids := make([][]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		p.Life = lives[i]
		p.Hand.Cards = p.Hand.Cards[:hands[i]]
		for j := 0; j < permanents[i]; j++ {
			ids[i] = append(ids[i], effectstest.PushVanillaCreature(g, p.ID, "Fodder", 1, 1))
		}
	}

	effectstest.CastCatalogSpell(t, g, "Pox Plague", "Sorcery", socB32PoxPlagueOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != wantLife[i] {
			t.Errorf("seat %d life %d, want %d (half of %d, rounded down)", i, p.Life, wantLife[i], lives[i])
		}
	}

	// The discards come next, in APNAP order, each seat owing half its hand.
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		p := g.Seats[seat]
		if discards[seat] == 0 {
			if effectstest.DiscardChoiceFor(g, p.ID) != nil {
				t.Fatalf("seat %d (a hand of %d) was asked to discard", seat, hands[seat])
			}
			continue
		}
		if got := effectstest.DiscardOwed(g, p.ID); got != discards[seat] {
			t.Fatalf("seat %d owes %d discards, want %d", seat, got, discards[seat])
		}
		for _, other := range ids {
			for _, id := range other {
				if !g.Battlefield.Contains(id) {
					t.Fatal("a permanent was sacrificed before the discards finished")
				}
			}
		}
		effectstest.DiscardFromHand(t, g, p.ID)
	}
	for i, p := range g.Seats {
		if p.Hand.Size() != hands[i]-discards[i] {
			t.Errorf("seat %d hand %d, want %d", i, p.Hand.Size(), hands[i]-discards[i])
		}
	}

	// Then the permanents, chosen in APNAP order and leaving together.
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		p := g.Seats[seat]
		prompt := effectstest.SacrificeChoiceFor(g, p.ID)
		if owed[seat] == 0 {
			if prompt != nil {
				t.Fatalf("seat %d (%d permanents) was asked to sacrifice", seat, permanents[seat])
			}
			continue
		}
		if prompt == nil {
			t.Fatalf("seat %d was not asked to sacrifice", seat)
		}
		if prompt.ChooseMax != owed[seat] {
			t.Fatalf("seat %d owes %d permanents, want %d", seat, prompt.ChooseMax, owed[seat])
		}
		effectstest.AnswerSacrifice(t, g, p.ID, ids[seat][:owed[seat]]...)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := range g.Seats {
		gone := 0
		for _, id := range ids[i] {
			if !g.Battlefield.Contains(id) {
				gone++
			}
		}
		if gone != owed[i] {
			t.Errorf("seat %d lost %d permanents, want %d", i, gone, owed[i])
		}
		for _, id := range ids[i][:owed[i]] {
			if g.Battlefield.Contains(id) {
				t.Errorf("seat %d: a chosen permanent survived", i)
			}
		}
	}
}

func TestSocB32PoxPlagueCountsEveryPermanentNotJustCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var all []uuid.UUID
	all = append(all, effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2))
	all = append(all, effectstest.SocB17Perm(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0))
	all = append(all, effectstest.SocB17Perm(g, me.ID, "Rock", "Artifact", "", "{1}", 0, 0))
	all = append(all, effectstest.SocB17Perm(g, me.ID, "Ring", "Enchantment", "", "{1}", 0, 0))
	all = append(all, effectstest.SocB17Perm(g, me.ID, "Token Land", "Basic Land — Island", "", "", 0, 0))
	// Everyone else has an empty hand and no permanents.
	for _, p := range g.Seats {
		p.Hand.Cards = nil
	}
	effectstest.CastCatalogSpell(t, g, "Pox Plague", "Sorcery", socB32PoxPlagueOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.SacrificeChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("five permanents owe two sacrifices, but nobody was asked")
	}
	if prompt.ChooseMax != 2 || len(prompt.ChooseCards) != 5 {
		t.Fatalf("prompt owes %d of %d candidates, want 2 of 5", prompt.ChooseMax, len(prompt.ChooseCards))
	}
	effectstest.AnswerSacrifice(t, g, me.ID, all[1], all[2])
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(all[1]) || g.Battlefield.Contains(all[2]) {
		t.Error("the land and the artifact the player chose were not sacrificed")
	}
	if !g.Battlefield.Contains(all[0]) || !g.Battlefield.Contains(all[3]) || !g.Battlefield.Contains(all[4]) {
		t.Error("a permanent the player did not choose was sacrificed")
	}
}

func TestSocB32PoxPlagueLeavesALonePermanentAndALoneCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	solo := effectstest.PushVanillaCreature(g, me.ID, "Solo", 1, 1)
	for _, p := range g.Seats {
		p.Hand.Cards = p.Hand.Cards[:1]
	}
	effectstest.CastCatalogSpell(t, g, "Pox Plague", "Sorcery", socB32PoxPlagueOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if effectstest.DiscardChoiceFor(g, p.ID) != nil {
			t.Errorf("%s was asked to discard half of one card", p.Name)
		}
		if p.Hand.Size() != 1 && p.ID != me.ID {
			t.Errorf("%s hand %d, want 1", p.Name, p.Hand.Size())
		}
	}
	if !g.Battlefield.Contains(solo) {
		t.Error("half of one permanent, rounded down, is none, but the lone creature was sacrificed")
	}
}

// --- Practiced Offense ----------------------------------------------------------

func socB32OffenseTargets(player, creature uuid.UUID) []game.TargetRef {
	return []game.TargetRef{
		{Kind: game.TargetPlayer, ID: player, Slot: 0},
		{Kind: game.TargetCard, ID: creature, Slot: 1},
	}
}

func TestSocB32PracticedOffenseCountersEveryCreatureOfThePlayerAndChoosesDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirA := effectstest.B12Creature(g, opp.ID, "Their A", socB32BearType, 2, 2)
	theirB := effectstest.B12Creature(g, opp.ID, "Their B", socB32BearType, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", socB32BearType, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Practiced Offense", "Sorcery", socB32OffenseOracle, socB32OffenseTargets(opp.ID, mine))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{theirA, theirB} {
		if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("the targeted player's creature has %d counters, want 1", n)
		}
	}
	if n := effectstest.CountersOn(g, mine, game.CounterPlusOne); n != 0 {
		t.Errorf("your own creature got %d counters though you targeted the opponent", n)
	}
	c := effectstest.FindBattlefieldCardByID(g, mine)
	if c == nil || !game.HasKeyword(c, "double strike") || game.HasKeyword(c, "lifelink") {
		t.Error("your creature should have double strike and not lifelink")
	}
	// The keyword only lasts until end of turn.
	effectstest.AdvanceToMainOf(t, g, 1)
	if c := effectstest.FindBattlefieldCardByID(g, mine); c == nil || game.HasKeyword(c, "double strike") {
		t.Error("double strike outlasted the turn")
	}
}

func TestSocB32PracticedOffenseCanChooseLifelinkAndTargetTheCountersOnYourself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", socB32BearType, 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Other", socB32BearType, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", socB32BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Practiced Offense", "Sorcery", socB32OffenseOracle, socB32OffenseTargets(me.ID, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, other} {
		if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("your creature has %d counters, want 1", n)
		}
	}
	if n := effectstest.CountersOn(g, theirs, game.CounterPlusOne); n != 0 {
		t.Errorf("the opponent's creature got %d counters, want 0", n)
	}
	if c := effectstest.FindBattlefieldCardByID(g, theirs); c == nil || !game.HasKeyword(c, "lifelink") || game.HasKeyword(c, "double strike") {
		t.Error("the targeted creature (not necessarily yours) should have lifelink and not double strike")
	}
}

func TestSocB32PracticedOffenseStillCountersWhenTheCreatureTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", socB32BearType, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", socB32BearType, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Practiced Offense", "Sorcery", socB32OffenseOracle, socB32OffenseTargets(opp.ID, mine))
	effectstest.ExileCreature(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, theirs, game.CounterPlusOne); n != 1 {
		t.Errorf("the player's creature has %d counters, want 1: the other clause is still legal", n)
	}
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("asked for a keyword with no creature to give it to")
	}
}

func TestSocB32PracticedOffenseFlashbackCostsOneWhiteAndExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", socB32BearType, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", socB32BearType, 1, 1)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Practiced Offense", TypeLine: "Sorcery", ManaCost: "{2}{W}",
		OracleID: socB32OffenseOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "flashback",
		Strict: true, Targets: socB32OffenseTargets(opp.ID, mine)}); err == nil {
		t.Fatal("flashback with one mana was accepted, but it costs {1}{W}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "W", 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "flashback",
		Strict: true, Targets: socB32OffenseTargets(opp.ID, mine)}); err != nil {
		t.Fatalf("flash it back: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, theirs, game.CounterPlusOne); n != 1 {
		t.Errorf("counters %d, want 1", n)
	}
	if me.Graveyard.Contains(id) {
		t.Error("the flashed-back card returned to the graveyard instead of being exiled")
	}
}

// --- Practiced Scrollsmith --------------------------------------------------------

func socB32Scrollsmith() game.Card {
	return game.Card{Name: "Practiced Scrollsmith", TypeLine: "Creature — Dwarf Cleric", OracleID: socB32ScrollsmithOracle, Power: 3, Toughness: 2}
}

func TestSocB32ScrollsmithExilesANoncreatureNonlandCardAndLetsYouCastItThroughYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spell := effectstest.GraveyardCardOf(g, me, "Dead Sorcery", "Sorcery", "{R}", "")
	creature := effectstest.GraveyardCardOf(g, me, "Dead Bear", socB32BearType, "{1}{G}", "")
	land := effectstest.GraveyardCardOf(g, me, "Dead Island", "Basic Land — Island", "", "")
	effectstest.Msc17Play(t, g, me, socB32Scrollsmith())
	socB32Settle(t, g, me.ID, spell)

	if z := effectstest.ZnrZoneOf(g, spell); z != game.ZoneExile {
		t.Fatalf("the sorcery is in %v, want exile", z)
	}
	if !me.Graveyard.Contains(creature) || !me.Graveyard.Contains(land) {
		t.Error("the creature card or the land card was exiled, but only a noncreature, nonland card is a legal target")
	}
	perm := effectstest.ExiledPermission(g, spell)
	if perm.Player != me.ID || !perm.CastOnly {
		t.Errorf("permission holder %v castOnly %v, want the controller and cast only", perm.Player, perm.CastOnly)
	}
	// It is cast for its printed cost, from exile, this turn.
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: string(game.ZoneExile), Strict: true}); err == nil {
		t.Fatal("cast it with an empty pool; the printed cost {R} must still be paid")
	}
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: string(game.ZoneExile), Strict: true}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(spell) {
		t.Error("the cast sorcery did not resolve into the graveyard")
	}
}

func TestSocB32ScrollsmithPermissionLastsThroughYourNextTurnAndNoLonger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spell := effectstest.GraveyardCardOf(g, me, "Dead Sorcery", "Sorcery", "{R}", "")
	effectstest.Msc17Play(t, g, me, socB32Scrollsmith())
	socB32Settle(t, g, me.ID, spell)

	effectstest.NextTurnOfMine(t, g)
	if perm := effectstest.ExiledPermission(g, spell); !effectstest.PermissionLive(g, perm, me.ID) {
		t.Fatal("the permission was gone at your next turn; it lasts until the end of that turn")
	}
	effectstest.NextTurnOfMine(t, g)
	if perm := effectstest.ExiledPermission(g, spell); effectstest.PermissionLive(g, perm, me.ID) {
		t.Error("the permission outlived the end of your next turn")
	}
	if z := effectstest.ZnrZoneOf(g, spell); z != game.ZoneExile {
		t.Errorf("the unused card is in %v, want it left in exile", z)
	}
}

func TestSocB32ScrollsmithWithNoLegalCardExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	creature := effectstest.GraveyardCardOf(g, me, "Dead Bear", socB32BearType, "{1}{G}", "")
	land := effectstest.GraveyardCardOf(g, me, "Dead Island", "Basic Land — Island", "", "")
	theirs := effectstest.GraveyardCardOf(g, opp, "Their Sorcery", "Sorcery", "{R}", "")
	effectstest.Msc17Play(t, g, me, socB32Scrollsmith())
	socB32Settle(t, g, me.ID, uuid.Nil)
	if !me.Graveyard.Contains(creature) || !me.Graveyard.Contains(land) || !opp.Graveyard.Contains(theirs) {
		t.Error("a card that is not a noncreature, nonland card in your graveyard was exiled")
	}
}

func TestSocB32ScrollsmithHasFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.Msc17Play(t, g, me, socB32Scrollsmith())
	socB32Settle(t, g, me.ID, uuid.Nil)
	if c := effectstest.FindBattlefieldCardByID(g, id); c == nil || !game.HasKeyword(c, "first strike") {
		t.Error("Practiced Scrollsmith should have first strike")
	}
}

// --- Primary Research -----------------------------------------------------------------

func socB32Research() game.Card {
	return game.Card{Name: "Primary Research", TypeLine: "Enchantment", OracleID: socB32ResearchOracle}
}

func TestSocB32PrimaryResearchReturnsAPermanentCardOfManaValueThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	small := effectstest.GraveyardCardOf(g, me, "Small Bear", socB32BearType, "{1}{G}", "")
	big := effectstest.GraveyardCardOf(g, me, "Big Bear", socB32BearType, "{3}{G}", "")
	land := effectstest.GraveyardCardOf(g, me, "Dead Island", "Basic Land — Island", "", "")
	spell := effectstest.GraveyardCardOf(g, me, "Dead Instant", "Instant", "{U}", "")
	effectstest.Msc17Play(t, g, me, socB32Research())
	socB32Settle(t, g, me.ID, small)
	if !g.Battlefield.Contains(small) {
		t.Error("the mana value 2 creature card did not return to the battlefield")
	}
	for _, id := range []uuid.UUID{big, land, spell} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("a card that is not a nonland permanent card with mana value 3 or less left the graveyard")
		}
	}
}

func TestSocB32PrimaryResearchWithOnlyIllegalCardsReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	big := effectstest.GraveyardCardOf(g, me, "Big Bear", socB32BearType, "{3}{G}", "")
	land := effectstest.GraveyardCardOf(g, me, "Dead Island", "Basic Land — Island", "", "")
	effectstest.Msc17Play(t, g, me, socB32Research())
	socB32Settle(t, g, me.ID, uuid.Nil)
	if g.Battlefield.Contains(big) || g.Battlefield.Contains(land) {
		t.Error("a mana value 4 creature or a land card came back")
	}
}

func TestSocB32PrimaryResearchDrawsAtYourEndStepOnlyIfACardLeftYourGraveyard(t *testing.T) {
	// The card its own trigger returns left the graveyard, so it draws.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	small := effectstest.GraveyardCardOf(g, me, "Small Bear", socB32BearType, "{1}{G}", "")
	effectstest.Msc17Play(t, g, me, socB32Research())
	socB32Settle(t, g, me.ID, small)
	before := me.Hand.Size()
	effectstest.AdvanceToStepOf(t, g, 0, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d at the end step, want %d: a card left the graveyard this turn", got, before+1)
	}

	// Nothing left: no draw.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.Msc17Play(t, g2, me2, socB32Research())
	socB32Settle(t, g2, me2.ID, uuid.Nil)
	before2 := me2.Hand.Size()
	effectstest.AdvanceToStepOf(t, g2, 0, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := me2.Hand.Size(); got != before2 {
		t.Errorf("hand %d at the end step, want %d: no card left the graveyard", got, before2)
	}
}

func TestSocB32PrimaryResearchCountsAnyWayACardLeavesAndIgnoresOpponentTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc17Play(t, g, me, socB32Research())
	socB32Settle(t, g, me.ID, uuid.Nil)
	exiled := effectstest.GraveyardCardOf(g, me, "Exiled Spell", "Instant", "{U}", "")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneExile}, exiled); err != nil {
		t.Fatalf("exile from the graveyard: %v", err)
	}
	// An opponent's end step is not yours.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepEnd)
	oppHand := g.Seats[1].Hand.Size()
	myHand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != myHand || g.Seats[1].Hand.Size() != oppHand {
		t.Error("Primary Research drew at an opponent's end step")
	}
}

// --- Prismari Charm ------------------------------------------------------------------

func TestSocB32PrismariCharmSurveilsTwoThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c1 := effectstest.TopLibraryCard(me, "Bottom", "Sorcery")
	c2 := effectstest.TopLibraryCard(me, "Second", "Sorcery")
	c3 := effectstest.TopLibraryCard(me, "Top", "Sorcery")
	hand := me.Hand.Size()
	effectstest.CastModal(t, g, "Prismari Charm", "Instant", socB32CharmOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("surveil looked at %d cards, want 2", len(c.ScryCards))
	}
	if me.Hand.Size() != hand {
		t.Fatal("drew before the surveil was answered")
	}
	// Put the top card into the graveyard and keep the second on top: it is the card drawn.
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{c3}, []uuid.UUID{c2}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(c3) {
		t.Error("the binned card is not in the graveyard")
	}
	if !me.Hand.Contains(c2) || me.Hand.Contains(c1) || me.Hand.Size() != hand+1 {
		t.Error("the card drawn should be the one kept on top by the surveil, and only one")
	}
}

func TestSocB32PrismariCharmDealsOneToEachOfTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.PushCostedPermanentForTest(g, opp.ID, "Bear", socB32BearType, "{1}{G}")
	life := opp.Life
	effectstest.CastModal(t, g, "Prismari Charm", "Instant", socB32CharmOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.DamageMarkedOn(g, bear); d != 1 {
		t.Errorf("the creature has %d damage, want 1", d)
	}
	if opp.Life != life-1 {
		t.Errorf("the player's life %d -> %d, want a loss of exactly 1", life, opp.Life)
	}
}

func TestSocB32PrismariCharmOneTargetAndNeverThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	effectstest.CastModal(t, g, "Prismari Charm", "Instant", socB32CharmOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("one target: life %d -> %d, want -1", life, opp.Life)
	}
	a := effectstest.PushCostedPermanentForTest(g, opp.ID, "A", socB32BearType, "{1}{G}")
	b := effectstest.PushCostedPermanentForTest(g, opp.ID, "B", socB32BearType, "{1}{G}")
	c := effectstest.PushCostedPermanentForTest(g, opp.ID, "C", socB32BearType, "{1}{G}")
	me := g.Seats[g.Turn.ActiveSeat]
	if _, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Prismari Charm", TypeLine: "Instant", OracleID: socB32CharmOracle},
		game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardRefs(a, b, c)}); err == nil {
		t.Error("three targets were accepted for \"one or two targets\"")
	}
}

func TestSocB32PrismariCharmBouncesANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.PushCostedPermanentForTest(g, opp.ID, "Bear", socB32BearType, "{1}{G}")
	land := effectstest.SocB17Perm(g, opp.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	me := g.Seats[g.Turn.ActiveSeat]
	if _, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Prismari Charm", TypeLine: "Instant", OracleID: socB32CharmOracle},
		game.CastSpellParams{Modes: []int{2}, Targets: effectstest.CardRefs(land)}); err == nil {
		t.Error("a land was accepted as the target of \"nonland permanent\"")
	}
	effectstest.CastModal(t, g, "Prismari Charm", "Instant", socB32CharmOracle, []int{2}, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the Bear should be back in its owner's hand")
	}
}

// --- Prismari, the Inspiration ----------------------------------------------------------

// socB32CastPull casts Pull from the Grave (no targets: each resolution gains 2).
func socB32CastPull(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Pull from the Grave", "Sorcery", socB32PullOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB32PrismariGivesYourInstantsAndSorceriesStorm(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Prismari, the Inspiration", "Legendary Creature — Elder Dragon", socB32PrismariOracle, 7, 7)
	start := me.Life
	// Storm counts the spells cast before it: 0, 1, then 2 copies.
	socB32CastPull(t, g)
	if me.Life != start+2 {
		t.Fatalf("the first spell: life %d, want %d (no copies)", me.Life, start+2)
	}
	socB32CastPull(t, g)
	if me.Life != start+2+4 {
		t.Fatalf("the second spell: life %d, want %d (the spell and one copy)", me.Life, start+6)
	}
	socB32CastPull(t, g)
	if me.Life != start+2+4+6 {
		t.Fatalf("the third spell: life %d, want %d (the spell and two copies)", me.Life, start+12)
	}
}

func TestSocB32WithoutPrismariThereAreNoCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	start := me.Life
	socB32CastPull(t, g)
	socB32CastPull(t, g)
	if me.Life != start+4 {
		t.Errorf("life %d, want %d: no storm without Prismari", me.Life, start+4)
	}
}

func TestSocB32PrismariOnlyGivesStormToYourOwnSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Prismari, the Inspiration", "Legendary Creature — Elder Dragon", socB32PrismariOracle, 7, 7)
	start := me.Life
	socB32CastPull(t, g)
	socB32CastPull(t, g)
	if me.Life != start+4 {
		t.Errorf("life %d, want %d: an opponent's Prismari gives my spells no storm", me.Life, start+4)
	}
}

func TestSocB32PrismariDoesNotGiveStormToACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Prismari, the Inspiration", "Legendary Creature — Elder Dragon", socB32PrismariOracle, 7, 7)
	socB32CastPull(t, g)
	if _, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Bear", TypeLine: socB32BearType, Power: 2, Toughness: 2}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast a creature: %v", err)
	}
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered {
			t.Errorf("a trigger %q is on the stack for a creature spell", item.Label)
		}
	}
}

func TestSocB32PrismariHasWardPayFiveLife(t *testing.T) {
	g, me, opp, dragon := effectstest.WardTable(t, "Prismari, the Inspiration", "Legendary Creature — Elder Dragon", socB32PrismariOracle)
	spell := effectstest.CastAtWardedCreature(t, g, opp, dragon)
	life := opp.Life
	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if effectstest.ConfirmChoiceFor(g, me.ID) != nil {
		t.Error("the dragon's controller must not be asked to pay")
	}
	if prompt.LifeCost != 5 {
		t.Errorf("prompt LifeCost = %d, want 5", prompt.LifeCost)
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dragon) || !opp.Graveyard.Contains(spell) {
		t.Error("declining the ward should counter the spell and leave the dragon")
	}
	if opp.Life != life {
		t.Errorf("declining cost %d life", life-opp.Life)
	}
}

func TestSocB32PrismariHasFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Prismari, the Inspiration", "Legendary Creature — Elder Dragon", socB32PrismariOracle, 7, 7)
	if c := effectstest.FindBattlefieldCardByID(g, id); c == nil || !game.HasKeyword(c, "flying") {
		t.Error("Prismari, the Inspiration should have flying")
	}
}

// --- Procrastinate ------------------------------------------------------------------------

func socB32Procrastinate(t *testing.T, g *game.Game, me *game.Player, x int, target uuid.UUID) error {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	return effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Procrastinate", TypeLine: "Sorcery", ManaCost: "{X}{U}", OracleID: socB32ProcrastinateOrcl},
		"U", x+1, game.CastSpellParams{XValue: x, Targets: effectstest.CardRefs(target)})
}

func TestSocB32ProcrastinateTapsTheCreatureAndPutsTwiceXStunCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", socB32BearType, 2, 2)
	if err := socB32Procrastinate(t, g, me, 3, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardByID(g, bear)
	if c == nil || !c.Tapped {
		t.Error("the creature should be tapped")
	}
	if n := effectstest.CountersOn(g, bear, game.CounterStun); n != 6 {
		t.Errorf("stun counters %d, want twice X = 6", n)
	}
}

func TestSocB32ProcrastinateStunCountersKeepItTappedThroughUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", socB32BearType, 2, 2)
	if err := socB32Procrastinate(t, g, me, 1, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToMainOf(t, g, 1)
	c := effectstest.FindBattlefieldCardByID(g, bear)
	if c == nil || !c.Tapped {
		t.Error("the stunned creature untapped in its controller's untap step")
	}
	if n := effectstest.CountersOn(g, bear, game.CounterStun); n != 1 {
		t.Errorf("stun counters %d after one untap step, want 2 - 1 = 1", n)
	}
}

func TestSocB32ProcrastinateAtXZeroStillTapsAndAnAlreadyTappedCreatureStillGetsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", socB32BearType, 2, 2)
	if err := socB32Procrastinate(t, g, me, 0, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, bear); c == nil || !c.Tapped {
		t.Error("at X = 0 the creature should still be tapped")
	}
	if n := effectstest.CountersOn(g, bear, game.CounterStun); n != 0 {
		t.Errorf("stun counters %d at X = 0, want 0", n)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	tapped := effectstest.B12Creature(g2, opp2.ID, "Tapped Bear", socB32BearType, 2, 2)
	g2.WithWriteLock(func() {
		for i := range g2.Battlefield.Cards {
			if g2.Battlefield.Cards[i].InstanceID == tapped {
				g2.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := socB32Procrastinate(t, g2, me2, 2, tapped); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if n := effectstest.CountersOn(g2, tapped, game.CounterStun); n != 4 {
		t.Errorf("an already tapped creature got %d stun counters, want 4", n)
	}
}

// --- Proctor's Gaze --------------------------------------------------------------------------

func socB32GazeLibrary(me *game.Player) (forest, instant uuid.UUID) {
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Shock", "Instant"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Bolt", "Instant"),
	)
	return ids[1], ids[0]
}

func TestSocB32ProctorsGazeBouncesThenFetchesATappedBasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest, _ := socB32GazeLibrary(me)
	bear := effectstest.PushCostedPermanentForTest(g, opp.ID, "Bear", socB32BearType, "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Proctor's Gaze", "Instant", socB32GazeOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the Bear should be back in its owner's hand")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	if card, ok := effectstest.BattlefieldCard(g, forest); !ok || !card.Tapped {
		t.Errorf("the Forest should be on the battlefield tapped: %+v ok=%v", card, ok)
	}
}

func TestSocB32ProctorsGazeWithNoTargetStillSearches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest, _ := socB32GazeLibrary(me)
	effectstest.CastCatalogSpell(t, g, "Proctor's Gaze", "Instant", socB32GazeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	if _, ok := effectstest.BattlefieldCard(g, forest); !ok {
		t.Error("with up to zero targets the search still happens")
	}
}

func TestSocB32ProctorsGazeRefusesALandTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.SocB17Perm(g, opp.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	if err := effectstest.CastCatalogSpellErr(t, g, "Proctor's Gaze", "Instant", socB32GazeOracle, effectstest.CardRefs(land)); err == nil {
		t.Error("a land was accepted as the target of \"nonland permanent\"")
	}
}

func TestSocB32ProctorsGazeDoesNothingIfItsOnlyTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest, _ := socB32GazeLibrary(me)
	bear := effectstest.PushCostedPermanentForTest(g, opp.ID, "Bear", socB32BearType, "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Proctor's Gaze", "Instant", socB32GazeOracle, effectstest.CardRefs(bear))
	effectstest.ExileCreature(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("the spell searched though its only target was illegal")
	}
	if _, ok := effectstest.BattlefieldCard(g, forest); ok {
		t.Error("the Forest was fetched by a spell that fizzled")
	}
}

// --- Pterafractyl ------------------------------------------------------------------------------

func socB32CastPterafractyl(t *testing.T, g *game.Game, me *game.Player, x int) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.PayMana(t, g, me, "{G}{U}")
	if x > 0 {
		effectstest.FillPool(me, x)
	}
	id, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Pterafractyl", TypeLine: "Creature — Dinosaur Fractal", ManaCost: "{X}{G}{U}",
		OracleID: socB32PterafractylOracle, Power: 1, Toughness: 0, PrintedPTKnown: true}, game.CastSpellParams{XValue: x, Strict: true})
	if err != nil {
		t.Fatalf("cast Pterafractyl for X=%d: %v", x, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestSocB32PterafractylEntersWithXCountersAndGainsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	id := socB32CastPterafractyl(t, g, me, 3)
	if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 3 {
		t.Errorf("+1/+1 counters %d, want X = 3", n)
	}
	if p, tough := effectstest.SocB14PT(t, g, id); p != 4 || tough != 3 {
		t.Errorf("Pterafractyl is %d/%d, want 1/0 plus three counters = 4/3", p, tough)
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	if c := effectstest.FindBattlefieldCardByID(g, id); c == nil || !game.HasKeyword(c, "flying") {
		t.Error("Pterafractyl should have flying")
	}
}

func TestSocB32PterafractylAtXZeroDiesButStillGainsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	id := socB32CastPterafractyl(t, g, me, 0)
	if g.Battlefield.Contains(id) {
		t.Error("a 1/0 stayed on the battlefield")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2: the enters trigger still resolves", life, me.Life)
	}
}

// --- Pull from the Grave ------------------------------------------------------------------------

func socB32Pull(t *testing.T, g *game.Game, me *game.Player, ids ...uuid.UUID) error {
	t.Helper()
	return effectstest.CastCatalogSpellErr(t, g, "Pull from the Grave", "Sorcery", socB32PullOracle, effectstest.CardRefs(ids...))
}

func TestSocB32PullFromTheGraveReturnsTwoCreatureCardsAndGainsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.GraveyardCardOf(g, me, "Dead Bear", socB32BearType, "{1}{G}", "")
	b := effectstest.GraveyardCardOf(g, me, "Dead Elf", "Creature — Elf", "{G}", "")
	c := effectstest.GraveyardCardOf(g, me, "Dead Spell", "Sorcery", "{1}", "")
	life := me.Life
	if err := socB32Pull(t, g, me, a, b); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two creature cards did not return to hand")
	}
	if !me.Graveyard.Contains(c) {
		t.Error("the sorcery card left the graveyard")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestSocB32PullFromTheGraveWithNoTargetStillGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	if err := socB32Pull(t, g, me); err != nil {
		t.Fatalf("cast with no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2 with up to zero targets", life, me.Life)
	}
}

func TestSocB32PullFromTheGraveRefusesNoncreatureOpponentAndThreeTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := effectstest.GraveyardCardOf(g, me, "Dead Spell", "Sorcery", "{1}", "")
	theirs := effectstest.GraveyardCardOf(g, opp, "Their Bear", socB32BearType, "{1}{G}", "")
	a := effectstest.GraveyardCardOf(g, me, "A", socB32BearType, "{1}", "")
	b := effectstest.GraveyardCardOf(g, me, "B", socB32BearType, "{1}", "")
	c := effectstest.GraveyardCardOf(g, me, "C", socB32BearType, "{1}", "")
	for _, bad := range [][]uuid.UUID{{spell}, {theirs}, {a, b, c}} {
		if err := socB32Pull(t, g, me, bad...); err == nil {
			t.Errorf("targets %v were accepted", bad)
		}
	}
}

func TestSocB32PullFromTheGraveDoesNothingWhenEveryTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.GraveyardCardOf(g, me, "Dead Bear", socB32BearType, "{1}{G}", "")
	life := me.Life
	if err := socB32Pull(t, g, me, a); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneExile}, a); err != nil {
		t.Fatalf("exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d -> %d: a spell whose only target is gone does nothing", life, me.Life)
	}
}
