package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b425_cards_test.go — Return to Ravnica: graveyard recursion, mana and
// ramp, token makers, enchantments (S82 older sets, batch 425, #5566). Each test
// plays the real catalog entry through the engine.

const (
	oldB425Spiral    = "5f928e54-3027-4f2f-9fef-d63b9ca416aa"
	oldB425Decoy     = "9fb2eb66-d33b-40e8-9d0d-06d73650a9ac"
	oldB425Zealot    = "ab9f03b4-eb67-4709-904d-abf4d491f0ee"
	oldB425Find      = "e99d6c5c-b029-4cd1-bd7a-6f8ea91eb4a7"
	oldB425Rakdos    = "3986b284-f943-4570-ba26-3419910feeef"
	oldB425Selesnya  = "d4ecee6d-b23f-48e7-b7e5-347027225709"
	oldB425Golgari   = "ec970eba-230a-4c41-874a-4a74abf3214e"
	oldB425Armada    = "33598888-0367-4085-9415-f4e21da08354"
	oldB425Captain   = "3c75234b-e367-46da-8454-be4315f9b9de"
	oldB425Currency  = "427ba80d-cfdf-4651-a1eb-89a2125a9217"
	oldB425Feud      = "fc5d0e18-2d10-47df-9742-2ea6c6d44431"
	oldB425Creature  = "Creature — Bear"
	oldB425Abandon   = "5404c6a5-399d-4237-b03c-41b6a1b79050"
	oldB425Enchament = "Enchantment"
)

func TestOldB425EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB425Spiral:   "Psychic Spiral",
		oldB425Decoy:    "Golgari Decoy",
		oldB425Zealot:   "Ash Zealot",
		oldB425Find:     "Treasured Find",
		oldB425Rakdos:   "Rakdos Keyrune",
		oldB425Selesnya: "Selesnya Keyrune",
		oldB425Golgari:  "Golgari Keyrune",
		oldB425Armada:   "Armada Wurm",
		oldB425Captain:  "Precinct Captain",
		oldB425Currency: "Conjured Currency",
		oldB425Feud:     "Guild Feud",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Keyrunes ---------------------------------------------------------------

var oldB425Keyrunes = []struct {
	name, oracle string
	colours      []string
	subtype      string
	power, tough int
	keyword      string
}{
	{"Rakdos Keyrune", oldB425Rakdos, []string{"B", "R"}, "Devil", 3, 1, "first strike"},
	{"Selesnya Keyrune", oldB425Selesnya, []string{"G", "W"}, "Wolf", 3, 3, ""},
	{"Golgari Keyrune", oldB425Golgari, []string{"B", "G"}, "Insect", 2, 2, "deathtouch"},
}

func TestOldB425KeyrunesTapForEitherColour(t *testing.T) {
	for _, tc := range oldB425Keyrunes {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			rune := effectstest.S58p6Push(g, me.ID, tc.name, "Artifact", tc.oracle, "{3}", 0, 0)
			for _, colour := range tc.colours {
				me.ManaPool.EmptyPool()
				effectstest.FindBattlefieldCardByID(g, rune).Tapped = false
				if err := g.ActivateManaAbility(me.ID, rune, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
					t.Fatalf("tap for %s: %v", colour, err)
				}
				if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != colour {
					t.Errorf("pool = %v, want one %s", me.ManaPool.Units(), colour)
				}
			}
		})
	}
}

func TestOldB425KeyrunesBecomeArtifactCreaturesUntilEndOfTurn(t *testing.T) {
	for _, tc := range oldB425Keyrunes {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			rune := effectstest.S58p6Push(g, me.ID, tc.name, "Artifact", tc.oracle, "{3}", 0, 0)
			effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
			if effectstest.S58p6View(t, g, rune).IsCreature() {
				t.Fatal("an idle Keyrune is a creature")
			}
			effectstest.NgheAnimate(t, g, me, rune, 0, tc.colours...)
			c := effectstest.S58p6View(t, g, rune)
			if !c.IsCreature() || !c.IsArtifact() {
				t.Fatalf("creature=%v artifact=%v, want both", c.IsCreature(), c.IsArtifact())
			}
			if !slices.Contains(c.Effective().Subtypes, tc.subtype) {
				t.Errorf("subtypes %v, want %s", c.Effective().Subtypes, tc.subtype)
			}
			if !slices.Equal(c.EffectiveColors(), tc.colours) {
				t.Errorf("colours %v, want %v", c.EffectiveColors(), tc.colours)
			}
			if c.CurrentPower() != tc.power || c.CurrentToughness() != tc.tough {
				t.Errorf("%d/%d, want %d/%d", c.CurrentPower(), c.CurrentToughness(), tc.power, tc.tough)
			}
			if tc.keyword != "" && !effectstest.HasEffectiveKeyword(t, g, rune, tc.keyword) {
				t.Errorf("no %s", tc.keyword)
			}
			effectstest.AdvanceToNextSeatsTurn(t, g)
			c = effectstest.S58p6View(t, g, rune)
			if c.IsCreature() || !c.IsArtifact() {
				t.Errorf("after the turn: creature=%v artifact=%v, want an artifact only", c.IsCreature(), c.IsArtifact())
			}
		})
	}
}

// --- Golgari Decoy ----------------------------------------------------------

func TestOldB425GolgariDecoyLuresEveryAbleBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	decoy := effectstest.ReqCreature(g, me.ID, "Golgari Decoy", oldB425Decoy, 2, 2, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := effectstest.ReqCreature(g, opp.ID, "Wall B", "", 0, 4)
	effectstest.ReqToBlockers(t, g, decoy)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Wall A must block Golgari Decoy if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := effectstest.ReqBlock(t, g, decoy, a, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both blocking: %v", err)
	}
}

func TestOldB425GolgariDecoyScavenges(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB425Creature, 2, 2)
	decoy, _ := effectstest.ActivateFromGraveyard(t, g, "Golgari Decoy", "Creature — Elf Rogue", oldB425Decoy, 2, 2, "{C}{C}{C}{G}{G}",
		game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr7Counters(g, bear); got != 2 {
		t.Errorf("Bear has %d +1/+1 counters, want 2 (the Decoy's power)", got)
	}
	if effectstest.B12ZoneOf(g, decoy) != game.ZoneExile {
		t.Error("scavenge exiles the card as a cost")
	}
}

// --- Ash Zealot -------------------------------------------------------------

func TestOldB425AshZealotHasFirstStrikeAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Ash Zealot", "Creature — Human Warrior", oldB425Zealot, true)
	for _, kw := range []string{"first strike", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

func TestOldB425AshZealotHitsAPlayerWhoCastsFromAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Ash Zealot", "Creature — Human Warrior", oldB425Zealot, 2, 2)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", oldB425Creature, 2, 2)
	life, oppLife := me.Life, opp.Life
	effectstest.MidB03Flash(t, g, "Abandon the Post", "Sorcery", oldB425Abandon, effectstest.CardTarget(foe))
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - me.Life; got != 3 {
		t.Errorf("the caster lost %d life, want 3", got)
	}
	if opp.Life != oppLife {
		t.Errorf("the Zealot's controller lost %d life, want 0", oppLife-opp.Life)
	}
}

func TestOldB425AshZealotIgnoresASpellCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Ash Zealot", "Creature — Human Warrior", oldB425Zealot, 2, 2)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("a spell cast from hand cost %d life", life-me.Life)
	}
}

// --- Treasured Find ---------------------------------------------------------

func TestOldB425TreasuredFindReturnsACardAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Dead Bear", TypeLine: oldB425Creature, Owner: me.ID, Controller: me.ID})
	spell := effectstest.CastCatalogSpell(t, g, "Treasured Find", "Sorcery", oldB425Find,
		[]game.TargetRef{{Kind: game.TargetCard, ID: card}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) {
		t.Error("the targeted card did not return to hand")
	}
	if effectstest.B12ZoneOf(g, spell) != game.ZoneExile {
		t.Errorf("Treasured Find is in %v, want exile", effectstest.B12ZoneOf(g, spell))
	}
}

func TestOldB425TreasuredFindOnlyTargetsYourOwnGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Bear", TypeLine: oldB425Creature, Owner: opp.ID, Controller: opp.ID})
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Treasured Find", "Sorcery", oldB425Find,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Error("an opponent's graveyard card was a legal target")
	}
}

// --- Psychic Spiral ---------------------------------------------------------

func TestOldB425PsychicSpiralShufflesYourGraveyardAndMillsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead", TypeLine: oldB425Creature, Owner: me.ID, Controller: me.ID})
	}
	effectstest.SeedLibrary(opp, "A", "B", "C", "D", "E", "F")
	mine, theirs, theirGrave := me.Library.Size(), opp.Library.Size(), opp.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Psychic Spiral", "Instant", oldB425Spiral, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 0 && me.Graveyard.Size() != 1 {
		t.Errorf("my graveyard holds %d cards, want only the resolved spell at most", me.Graveyard.Size())
	}
	if got := me.Library.Size() - mine; got != 3 {
		t.Errorf("my library grew by %d, want 3", got)
	}
	if got := theirs - opp.Library.Size(); got != 3 {
		t.Errorf("the target milled %d, want 3", got)
	}
	if got := opp.Graveyard.Size() - theirGrave; got != 3 {
		t.Errorf("the target's graveyard grew by %d, want 3", got)
	}
}

func TestOldB425PsychicSpiralCanTargetItsOwnController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 2; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead", TypeLine: oldB425Creature, Owner: me.ID, Controller: me.ID})
	}
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	lib := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Psychic Spiral", "Instant", oldB425Spiral, []game.TargetRef{effectstest.PlayerTarget(me.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	// Two cards shuffled in, two milled back out.
	if me.Library.Size() != lib {
		t.Errorf("library %d → %d, want unchanged (2 in, 2 milled)", lib, me.Library.Size())
	}
}

// --- Armada Wurm ------------------------------------------------------------

func TestOldB425ArmadaWurmMakesAFiveFiveTramplingWurm(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Armada Wurm", "Creature — Wurm", oldB425Armada, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTokensNamed2(g, me.ID, "Wurm"); got != 1 {
		t.Fatalf("%d Wurm tokens, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Wurm" {
			if c.Power != 5 || c.Toughness != 5 || !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "trample") {
				t.Errorf("token is %d/%d, want 5/5 with trample", c.Power, c.Toughness)
			}
		}
	}
}

// --- Precinct Captain -------------------------------------------------------

func TestOldB425PrecinctCaptainMakesASoldierWhenItHitsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cap := effectstest.ReqCreature(g, me.ID, "Precinct Captain", oldB425Captain, 2, 2, "haste")
	effectstest.BrAttack(t, g, cap)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MidB03Tokens(g, me.ID, "Soldier"); len(got) != 1 {
		t.Fatalf("%d Soldier tokens, want 1", len(got))
	} else if got[0].Power != 1 || got[0].Toughness != 1 {
		t.Errorf("Soldier is %d/%d, want 1/1", got[0].Power, got[0].Toughness)
	}
}

// --- Conjured Currency ------------------------------------------------------

func oldB425Perm(g *game.Game, owner, controller uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: oldB425Creature,
		Power: 1, Toughness: 1, Owner: owner, Controller: controller})
}

func TestOldB425ConjuredCurrencyExchangesWithAPermanentYouNeitherOwnNorControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	cur := effectstest.SeedEnchantment(g, oldB425Currency, "Conjured Currency", oldB425Enchament, me.ID)
	mine := oldB425Perm(g, me.ID, me.ID, "Mine")
	stolen := oldB425Perm(g, me.ID, opp.ID, "Stolen From Me") // I own it, they control it
	borrowed := oldB425Perm(g, opp.ID, other.ID, "Borrowed")  // neither mine
	theirs := oldB425Perm(g, opp.ID, opp.ID, "Theirs")        // neither mine
	effectstest.BeginUpkeep(t, g, me.ID)
	pc := effectstest.LatestPickTarget(g, me.ID)
	if pc == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range pc.PickTargetCards {
		if id == mine || id == stolen || id == cur {
			t.Fatalf("a permanent I own or control (%s) was offered", id)
		}
	}
	effectstest.CloudDeckPickTargets(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, cur); c.Controller != opp.ID {
		t.Errorf("Conjured Currency is controlled by %v, want the opponent", c.Controller)
	}
	if c := effectstest.FindBattlefieldCardByID(g, theirs); c.Controller != me.ID {
		t.Errorf("the target is controlled by %v, want me", c.Controller)
	}
	if c := effectstest.FindBattlefieldCardByID(g, borrowed); c.Controller != other.ID {
		t.Error("the unrelated permanent moved")
	}
}

func TestOldB425ConjuredCurrencyMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cur := effectstest.SeedEnchantment(g, oldB425Currency, "Conjured Currency", oldB425Enchament, me.ID)
	theirs := oldB425Perm(g, opp.ID, opp.ID, "Theirs")
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.CloudDeckPickTargets(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, cur); c.Controller != me.ID {
		t.Error("declining the may still exchanged control")
	}
}

// --- Guild Feud -------------------------------------------------------------

func oldB425Top(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 3, Toughness: 3,
		Owner: p.ID, Controller: p.ID})
	return id
}

func oldB425FeudUpkeep(t *testing.T, g *game.Game, me, opp *game.Player) {
	t.Helper()
	effectstest.BeginUpkeep(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB425GuildFeudBothPutACreatureAndTheyFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEnchantment(g, oldB425Feud, "Guild Feud", oldB425Enchament, me.ID)
	oppRest1 := oldB425Top(opp, "Opp Sorcery", "Sorcery")
	oppRest2 := oldB425Top(opp, "Opp Land", "Basic Land — Forest")
	oppBear := oldB425Top(opp, "Opp Bear", oldB425Creature)
	myRest1 := oldB425Top(me, "My Sorcery", "Sorcery")
	myBear := oldB425Top(me, "My Bear", oldB425Creature)
	myRest2 := oldB425Top(me, "My Land", "Basic Land — Forest")
	// Opponent's three are revealed first.
	oldB425FeudUpkeep(t, g, me, opp)
	c := effectstest.LatestChooseCardsFor(g, opp.ID)
	if c == nil {
		t.Fatal("the opponent was not offered a creature card")
	}
	if !effectstest.HasID(c.ChooseCards, oppBear) || effectstest.HasID(c.ChooseCards, oppRest1) || effectstest.HasID(c.ChooseCards, oppRest2) {
		t.Errorf("offered %v, want only the creature card", c.ChooseCards)
	}
	if c.ChooseMin != 0 {
		t.Error("\"may put\": declining must be legal")
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{oppBear}); err != nil {
		t.Fatalf("opponent's pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c = effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("I was not offered a creature card")
	}
	if !effectstest.HasID(c.ChooseCards, myBear) || effectstest.HasID(c.ChooseCards, myRest1) || effectstest.HasID(c.ChooseCards, myRest2) {
		t.Errorf("offered %v, want only the creature card", c.ChooseCards)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{myBear}); err != nil {
		t.Fatalf("my pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{oppRest1, oppRest2} {
		if !opp.Graveyard.Contains(id) {
			t.Errorf("the opponent's rest %s did not go to their graveyard", id)
		}
	}
	for _, id := range []uuid.UUID{myRest1, myRest2} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("my rest %s did not go to my graveyard", id)
		}
	}
	// 3/3 fights 3/3: both die.
	if effects.OnBattlefield(g, oppBear) || effects.OnBattlefield(g, myBear) {
		t.Error("the two creatures did not fight each other to death")
	}
}

func TestOldB425GuildFeudNoFightWhenTheOpponentPutsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedEnchantment(g, oldB425Feud, "Guild Feud", oldB425Enchament, me.ID)
	oppRest := oldB425Top(opp, "Opp Land", "Basic Land — Forest")
	oldB425Top(opp, "Opp Land 2", "Basic Land — Forest")
	oldB425Top(opp, "Opp Land 3", "Basic Land — Forest")
	myBear := oldB425Top(me, "My Bear", oldB425Creature)
	oldB425FeudUpkeep(t, g, me, opp)
	// No creature among theirs: no prompt for them, mine opens at once.
	if effectstest.LatestChooseCardsFor(g, opp.ID) != nil {
		t.Fatal("the opponent was asked with no creature card to put")
	}
	if !opp.Graveyard.Contains(oppRest) {
		t.Error("the opponent's cards did not go to their graveyard")
	}
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("I was not offered my creature")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{myBear}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, myBear) {
		t.Error("my creature fought nothing and should stay on the battlefield")
	}
}
