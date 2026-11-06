package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ecl_6_cards_test.go — Lorwyn Eclipsed, sixth batch (#4458). Each test
// plays the real catalog entry and asserts what its text says.

const (
	ecl6FowlknightOracle = "4e0ba9d5-4bd6-49a0-ae00-70818738d0ba"
	ecl6StomplingOracle  = "60206d4e-4301-4d7a-a68e-46a2ad9fc407"
	ecl6GiantfallOracle  = "77e5e7e3-b97b-4953-9df4-1aee509fbf2a"
	ecl6EmbraceOracle    = "5eb8bb1d-7567-4c24-913b-463b5fe463c4"
	ecl6GifterOracle     = "e67a820b-6249-478e-91db-9827e44aa60a"
	ecl6GlamermiteOracle = "a7b0d897-38c0-4f90-81e8-3c5b8d31a607"
	ecl6BairnOracle      = "6a7229a8-1634-4278-ab50-0ec2b51edac1"
	ecl6GoatnapOracle    = "060e7187-0c7d-4022-9a1f-36a6f75314ca"
	ecl6NomadOracle      = "79dfe2eb-82b9-4dae-95c8-58fdf9006bc7"
	ecl6ShifterOracle    = "1eebc02a-9a64-49b0-8322-c5ce671457ce"
)

func TestEcl6EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		ecl6FowlknightOracle: "Gallant Fowlknight",
		ecl6StomplingOracle:  "Gangly Stompling",
		ecl6GiantfallOracle:  "Giantfall",
		ecl6EmbraceOracle:    "Gilt-Leaf's Embrace",
		ecl6GifterOracle:     "Glamer Gifter",
		ecl6GlamermiteOracle: "Glamermite",
		ecl6BairnOracle:      "Glister Bairn",
		ecl6GoatnapOracle:    "Goatnap",
		ecl6NomadOracle:      "Goldmeadow Nomad",
		ecl6ShifterOracle:    "Graveshifter",
	}
	for oracle, name := range want {
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

// --- shared helpers ----------------------------------------------------------

func ecl6Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// ecl6Cast puts a catalog card with printed P/T into p's hand and casts it
// where the cursor stands, without the strict mana gate.
func ecl6Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, power, toughness int, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
		})
	})
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// ecl6ResolveSpell passes priority until the stack is empty or a prompt
// opens, leaving the prompt for the caller.
func ecl6ResolveSpell(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 12 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				return
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

func ecl6NextTurn(t *testing.T, g *game.Game) {
	t.Helper()
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
}

// --- Gallant Fowlknight --------------------------------------------------------

func TestEcl6GallantFowlknightPumpsTheTeamAndGivesKithkinFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	kithkin := b12Creature(g, me.ID, "Kithkin", "Creature — Kithkin Soldier", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Kithkin", "Creature — Kithkin", 2, 2)
	advanceToMain(t, g)
	knight := ecl6Cast(t, g, me, "Gallant Fowlknight", "Creature — Kithkin Knight", ecl6FowlknightOracle, 3, 4, game.CastSpellParams{})
	passPriorityAroundTable(t, g)

	for _, tc := range []struct {
		name         string
		id           uuid.UUID
		power, tough int
		firstStrike  bool
	}{
		{"the Fowlknight", knight, 4, 4, true},
		{"a Kithkin you control", kithkin, 2, 1, true},
		{"a Bear you control", bear, 3, 2, false},
		{"an opponent's Kithkin", theirs, 2, 2, false},
	} {
		p, tt := ptOf(t, g, tc.id)
		fs := hasEffectiveKeyword(t, g, tc.id, "first strike")
		if p != tc.power || tt != tc.tough || fs != tc.firstStrike {
			t.Errorf("%s: %d/%d first strike %v; want %d/%d first strike %v", tc.name, p, tt, fs, tc.power, tc.tough, tc.firstStrike)
		}
	}

	ecl6NextTurn(t, g)
	if p, _ := ptOf(t, g, bear); p != 2 || hasEffectiveKeyword(t, g, kithkin, "first strike") {
		t.Error("the pump or the first strike outlasted the turn")
	}
}

// --- Gangly Stompling ------------------------------------------------------------

func TestEcl6GanglyStomplingHasTrampleAndIsEveryCreatureType(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	advanceToMain(t, g)
	id := ecl6Cast(t, g, me, "Gangly Stompling", "Creature — Shapeshifter", ecl6StomplingOracle, 4, 2, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, id, "trample") {
		t.Error("Gangly Stompling has no trample")
	}
	c := layeredCard(t, g, id)
	for _, sub := range []string{"Kithkin", "Goat", "Elf"} {
		if !c.HasSubtype(sub) {
			t.Errorf("Gangly Stompling is not a %s", sub)
		}
	}
}

// --- Giantfall -----------------------------------------------------------------

func TestEcl6GiantfallBitesWithTheFirstTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	biter := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	victim := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 2, 3)
	mine := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 2, 2)
	advanceToMain(t, g)

	// The second clause is a creature an opponent controls.
	if _, err := castModalPaying(t, g, "Giantfall", "Instant", ecl6GiantfallOracle, []int{0},
		cardRefs(biter, mine), nil, nil, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("biting my own creature: err = %v, want ErrIllegalTarget", err)
	}
	if _, err := castModalPaying(t, g, "Giantfall", "Instant", ecl6GiantfallOracle, []int{0},
		cardRefs(biter, victim), nil, nil, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !opp.Graveyard.Contains(victim) {
		t.Error("3 damage did not kill the 2/3")
	}
	if !g.Battlefield.Contains(biter) {
		t.Error("the biter took damage back; it is one-sided")
	}
}

func TestEcl6GiantfallDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := ecl6Seats(g)
	rock := pushArtifactForTest(g, opp.ID, "Mind Stone")
	advanceToMain(t, g)
	if _, err := castModalPaying(t, g, "Giantfall", "Instant", ecl6GiantfallOracle, []int{1},
		cardRefs(rock), nil, nil, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
}

// --- Gilt-Leaf's Embrace -------------------------------------------------------

func TestEcl6GiltLeafsEmbraceFlashesInForTrampleIndestructibleAndTwoPower(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	// Flash: cast at the beginning of combat, outside a main phase.
	advanceTo(t, g, game.StepBeginCombat)
	addPool(t, g, me.ID, "{C}{C}{G}")
	castFromHandForTest(t, g, me, "Gilt-Leaf's Embrace", "Enchantment — Aura", "{2}{G}", ecl6EmbraceOracle,
		game.CastSpellParams{Targets: cardRefs(bear), Strict: true})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 2 {
		t.Errorf("enchanted Bear = %d/%d, want 4/2", p, tt)
	}
	for _, kw := range []string{"trample", "indestructible"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the enters trigger gave no %s", kw)
		}
	}
	b27Kill(g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("an indestructible creature was destroyed")
	}

	ecl6NextTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "trample") || hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("trample or indestructible outlasted the turn")
	}
	if p, _ := ptOf(t, g, bear); p != 4 {
		t.Errorf("the +2/+0 ended with the turn: power %d", p)
	}
}

// --- Glamer Gifter ---------------------------------------------------------------

func TestEcl6GlamerGifterMakesAnotherCreatureAFourFourOfEveryType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	elf := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	advanceToMain(t, g)
	gifter := ecl6Cast(t, g, me, "Glamer Gifter", "Creature — Faerie Wizard", ecl6GifterOracle, 1, 2, game.CastSpellParams{})
	ecl6ResolveSpell(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if hasID(pick.PickTargetCards, gifter) || !hasID(pick.PickTargetCards, elf) {
		t.Fatalf("offered %v: want the other creature and not the Gifter", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, elf)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, elf); p != 4 || tt != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tt)
	}
	if c := layeredCard(t, g, elf); !c.HasSubtype("Goblin") || !c.HasSubtype("Kithkin") {
		t.Error("the target did not gain all creature types")
	}
	if p, tt := ptOf(t, g, gifter); p != 1 || tt != 2 {
		t.Errorf("the Gifter is %d/%d, want its own 1/2", p, tt)
	}

	ecl6NextTurn(t, g)
	if p, tt := ptOf(t, g, elf); p != 1 || tt != 1 {
		t.Errorf("after the turn the target is %d/%d, want 1/1", p, tt)
	}
	if layeredCard(t, g, elf).HasSubtype("Goblin") {
		t.Error("all creature types outlasted the turn")
	}
}

// The base size is layer 7b: a +1/+1 counter still applies on top.
func TestEcl6GlamerGifterSetsBaseSizeUnderACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	advanceToMain(t, g)
	ecl6Cast(t, g, me, "Glamer Gifter", "Creature — Faerie Wizard", ecl6GifterOracle, 1, 2, game.CastSpellParams{})
	ecl6ResolveSpell(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 5 || tt != 5 {
		t.Errorf("the countered Bear is %d/%d, want 5/5", p, tt)
	}
}

// --- Glamermite ------------------------------------------------------------------

func ecl6GlamermiteEnters(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	ecl6Cast(t, g, me, "Glamermite", "Creature — Faerie Rogue", ecl6GlamermiteOracle, 2, 2, game.CastSpellParams{})
	ecl6ResolveSpell(t, g)
	mode := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt from the enters trigger: %+v", g.PendingChoices)
	}
	return mode
}

func TestEcl6GlamermiteTapsAndUntapsATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	mode := ecl6GlamermiteEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("choosing tap: %v", err)
	}
	if err := sgAnswerTargets(t, g, theirs); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, theirs); !tapped {
		t.Error("the targeted creature should be tapped")
	}

	mode = ecl6GlamermiteEnters(t, g, me)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("choosing untap: %v", err)
	}
	if err := sgAnswerTargets(t, g, theirs); err != nil {
		t.Fatalf("target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, theirs); tapped {
		t.Error("the targeted creature should be untapped")
	}
}

// --- Glister Bairn ---------------------------------------------------------------

func TestEcl6GlisterBairnPumpsAnotherCreatureByTheColorsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bairn := b27Push(g, me.ID, "Glister Bairn", "Creature — Ouphe", ecl6BairnOracle, "{2}{G/U}{G/U}{G/U}", 1, 4, "G", "U")
	red := b27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	b27Push(g, me.ID, "Second Red", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	b27Push(g, me.ID, "Rock", "Artifact", "", "{2}", 0, 0)
	theirs := b27Push(g, opp.ID, "Their Black", "Creature — Zombie", "", "{1}{B}", 2, 2, "B")

	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no Glister Bairn trigger at the beginning of combat")
	}
	if hasID(pick.PickTargetCards, bairn) || hasID(pick.PickTargetCards, theirs) || !hasID(pick.PickTargetCards, red) {
		t.Fatalf("offered %v: want another creature you control only", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, red)
	passPriorityAroundTable(t, g)
	// Green, blue and red among my permanents: X is 3. The opponent's black
	// creature and the colorless artifact add nothing.
	if p, tt := ptOf(t, g, red); p != 5 || tt != 5 {
		t.Errorf("the target is %d/%d, want 5/5", p, tt)
	}
	if p, tt := ptOf(t, g, bairn); p != 1 || tt != 4 {
		t.Errorf("the Bairn is %d/%d, want its own 1/4", p, tt)
	}
}

func TestEcl6GlisterBairnDoesNotTriggerOnAnOpponentsCombat(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	b27Push(g, opp.ID, "Glister Bairn", "Creature — Ouphe", ecl6BairnOracle, "{2}{G/U}{G/U}{G/U}", 1, 4, "G", "U")
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	if latestPickTarget(g, opp.ID) != nil {
		t.Error("Glister Bairn triggered at another player's beginning of combat")
	}
}

// --- Goatnap ---------------------------------------------------------------------

func TestEcl6GoatnapStealsUntapsAndHastesAndPumpsAGoat(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		power          int
	}{
		{"a Goat", "Creature — Goat", 4},
		{"not a Goat", "Creature — Bear", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := ecl6Seats(g)
			target := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Victim", TypeLine: tc.typeLine,
				Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID, Tapped: true,
			})
			advanceToMain(t, g)
			ecl6Cast(t, g, me, "Goatnap", "Sorcery", ecl6GoatnapOracle, 0, 0,
				game.CastSpellParams{Targets: cardRefs(target)})
			passPriorityAroundTable(t, g)
			if got := controllerOf(t, g, target); got != me.ID {
				t.Fatal("Goatnap did not gain control of the creature")
			}
			if tapped, _ := battlefieldCardTapped(g, target); tapped {
				t.Error("the stolen creature was not untapped")
			}
			if !hasEffectiveKeyword(t, g, target, "haste") {
				t.Error("the stolen creature has no haste")
			}
			if p, tt := ptOf(t, g, target); p != tc.power || tt != 1 {
				t.Errorf("the stolen creature is %d/%d, want %d/1", p, tt, tc.power)
			}
			ecl6NextTurn(t, g)
			if got := controllerOf(t, g, target); got != opp.ID {
				t.Error("control did not return at the end of the turn")
			}
			if p, _ := ptOf(t, g, target); p != 1 {
				t.Errorf("the +3/+0 outlasted the turn: power %d", p)
			}
		})
	}
}

// A changeling is a Goat.
func TestEcl6GoatnapPumpsAChangeling(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	target := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shapeshifter", TypeLine: "Creature — Shapeshifter",
		Keywords: []string{game.KeywordChangeling},
		Power:    2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	advanceToMain(t, g)
	ecl6Cast(t, g, me, "Goatnap", "Sorcery", ecl6GoatnapOracle, 0, 0,
		game.CastSpellParams{Targets: cardRefs(target)})
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, target); p != 5 {
		t.Errorf("the stolen changeling has power %d, want 5", p)
	}
}

// --- Goldmeadow Nomad --------------------------------------------------------------

func TestEcl6GoldmeadowNomadExilesItselfForAKithkin(t *testing.T) {
	g := newCatalogGame(t)
	nomad, me := activateFromGraveyard(t, g, "Goldmeadow Nomad", "Creature — Kithkin Scout", ecl6NomadOracle, 1, 2, "{W}",
		game.ActivateAbilityParams{})
	if me.Graveyard.Contains(nomad) {
		t.Fatal("the Nomad was not exiled as part of the cost")
	}
	passPriorityAroundTable(t, g)
	if got := ecl1Named(g, me.ID, "Kithkin"); got != 1 {
		t.Fatalf("Kithkin tokens = %d, want 1", got)
	}
	k := layeredCard(t, g, findBattlefieldByName(g, "Kithkin"))
	if k.CurrentPower() != 1 || k.CurrentToughness() != 1 || !k.HasColor("G") || !k.HasColor("W") {
		t.Errorf("the token is %d/%d colors %v, want a 1/1 green and white", k.CurrentPower(), k.CurrentToughness(), k.Colors)
	}
}

func TestEcl6GoldmeadowNomadActivatesOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	advanceTo(t, g, game.StepBeginCombat)
	nomad := pushCatalogGraveyardCard(me, "Goldmeadow Nomad", "Creature — Kithkin Scout", ecl6NomadOracle, 1, 2)
	addPool(t, g, me.ID, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, nomad, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the Nomad activated outside a main phase")
	}
	if !me.Graveyard.Contains(nomad) {
		t.Error("a refused activation exiled the Nomad")
	}
}

// --- Graveshifter ----------------------------------------------------------------

func TestEcl6GraveshifterReturnsACreatureCardOnAYes(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	bear := pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	spell := batch01GraveyardCard(me, "Old Spell", "Sorcery")
	advanceToMain(t, g)
	shifter := ecl6Cast(t, g, me, "Graveshifter", "Creature — Shapeshifter", ecl6ShifterOracle, 2, 2, game.CastSpellParams{})
	ecl6ResolveSpell(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if !hasID(pick.PickTargetCards, bear) || hasID(pick.PickTargetCards, spell) {
		t.Fatalf("offered %v: want the creature card and not the sorcery", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, bear)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || me.Graveyard.Contains(bear) {
		t.Error("the creature card did not return to hand")
	}
	if !layeredCard(t, g, shifter).HasSubtype("Goat") {
		t.Error("Graveshifter is not every creature type")
	}
}

func TestEcl6GraveshifterMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	bear := pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	advanceToMain(t, g)
	ecl6Cast(t, g, me, "Graveshifter", "Creature — Shapeshifter", ecl6ShifterOracle, 2, 2, game.CastSpellParams{})
	ecl6ResolveSpell(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("a declined trigger returned the card")
	}
}
