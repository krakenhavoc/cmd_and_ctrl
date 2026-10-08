package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// daynight_a_cards_test.go — the first half of the day/night card list
// (#2586, slice daynight-a), on real imported cards for the werewolves
// (so the importer's per-face daybound / nightbound stamping is on the
// road under test) and on pushed catalogue permanents for the rest.

const (
	dnDay   = "Daybound (If a player casts no spells during their own turn, it becomes night next turn.)"
	dnNight = "Nightbound (If a player casts at least two spells during their own turn, it becomes day next turn.)"

	dnaArlinn       = "f227ce07-7e96-4a36-ab7c-9be6e777d649"
	dnaAvabruck     = "730be0d6-2612-44d6-9d36-e1fc6510c6bd"
	dnaBallista     = "b6811c31-fcd3-4d00-89d5-cf974575a87c"
	dnaBaneblade    = "4f5da665-e880-4325-a9e2-6fc4ca1d807a"
	dnaBrutalCathar = "1ed2d8e0-462b-468e-8fd3-1f3c6d99fb8a"
	dnaBurlyBreaker = "4cbfb898-e1ca-4e34-9cb2-11ba50272984"
	dnaSanctifier   = "6885aa03-fbfe-4c2e-a785-0865788574b4"
	dnaChildOfPack  = "a9c5b155-7c09-4100-9679-8fca2b5e9222"
	dnaCollector    = "6950ab0f-f448-4ed1-9dd4-868cc79a5982"
	dnaFangblade    = "dbd22a65-4ccb-4435-ae27-03a47a86d630"
	dnaDawnguard    = "2788cc1e-18b6-4506-be4c-f4056600e33c"
	dnaTrespasser   = "0bbd6cad-9b6f-45a6-9f2e-d7b4853586ae"
	dnaHookhand     = "52def237-0374-4ab0-8c01-d0d00aaa324f"
	dnaHoundTamer   = "e9208fc2-616d-4c32-bd66-76a8bf85a6b5"
	dnaPiper        = "cede233b-4e27-4738-8099-c8e46862ba96"
	dnaLoner        = "6e0b3317-394d-42cc-a350-cb5ce051787a"
	dnaKessig       = "54f8acb1-58f3-49d1-bff8-c1b578245936"
	dnaLambholt     = "f3c104e2-470b-429f-a047-a21edc2adb3b"
)

// dnaWerewolf is one transforming werewolf: both faces' names, type
// lines, printed text and P/T, as the dump has them, and the keywords
// (beyond daybound and nightbound) each face is expected to carry.
type dnaWerewolf struct {
	oracle                  string
	front, back             string
	frontType, backType     string
	cost                    string
	frontText, backText     string
	fp, ft, bp, bt          string
	colors                  []string
	extraKeywords           []string // the dump's top-level keywords
	frontKeywords, backKeys []string // expected on each face, beyond day/night
}

func (w dnaWerewolf) row() cards.Card {
	return werewolfRow(w.oracle, w.front, w.frontType, w.cost, w.frontText, w.back, w.backType, w.backText,
		w.fp, w.ft, w.bp, w.bt, w.colors, w.extraKeywords...)
}

var dnaWerewolves = []dnaWerewolf{
	{dnaAvabruck, "Avabruck Caretaker", "Hollowhenge Huntmaster", "Creature — Human Werewolf", "Creature — Werewolf", "{4}{G}{G}",
		"Hexproof\nAt the beginning of combat on your turn, put two +1/+1 counters on another target creature you control.\n" + dnDay,
		"Hexproof\nOther permanents you control have hexproof.\nAt the beginning of combat on your turn, put two +1/+1 counters on each creature you control.\n" + dnNight,
		"4", "4", "6", "6", []string{"G"}, []string{"Hexproof"}, []string{"hexproof"}, []string{"hexproof"}},
	{dnaBallista, "Ballista Watcher", "Ballista Wielder", "Creature — Human Soldier Werewolf", "Creature — Werewolf", "{2}{R}{R}",
		"{2}{R}, {T}: This creature deals 1 damage to any target.\n" + dnDay,
		"{2}{R}: This creature deals 1 damage to any target. A creature dealt damage this way can't block this turn.\n" + dnNight,
		"4", "3", "5", "5", []string{"R"}, nil, nil, nil},
	{dnaBaneblade, "Baneblade Scoundrel", "Baneclaw Marauder", "Creature — Human Rogue Werewolf", "Creature — Werewolf", "{3}{B}",
		"Whenever this creature becomes blocked, each creature blocking it gets -1/-1 until end of turn.\n" + dnDay,
		"Whenever this creature becomes blocked, each creature blocking it gets -1/-1 until end of turn.\nWhenever a creature blocking this creature dies, that creature's controller loses 1 life.\n" + dnNight,
		"4", "3", "5", "4", []string{"B"}, nil, nil, nil},
	{dnaBrutalCathar, "Brutal Cathar", "Moonrage Brute", "Creature — Human Soldier Werewolf", "Creature — Werewolf", "{2}{W}",
		"Whenever this creature enters or transforms into Brutal Cathar, exile target creature an opponent controls until this creature leaves the battlefield.\n" + dnDay,
		"First strike\nWard—Pay 3 life.\n" + dnNight,
		"2", "2", "3", "3", []string{"W"}, []string{"First strike", "Ward"}, nil, []string{"first strike"}},
	{dnaBurlyBreaker, "Burly Breaker", "Dire-Strain Demolisher", "Creature — Human Werewolf", "Creature — Werewolf", "{3}{G}{G}",
		"Ward {1}\n" + dnDay, "Ward {3}\n" + dnNight,
		"6", "5", "8", "7", []string{"G"}, []string{"Ward"}, nil, nil},
	{dnaChildOfPack, "Child of the Pack", "Savage Packmate", "Creature — Human Werewolf", "Creature — Werewolf", "{R}{R}{R}{G}",
		"{2}{R}{G}: Create a 2/2 green Wolf creature token.\n" + dnDay,
		"Trample\nOther creatures you control get +1/+0.\n" + dnNight,
		"2", "5", "5", "5", []string{"R", "G"}, []string{"Trample"}, nil, []string{"trample"}},
	{dnaFangblade, "Fangblade Brigand", "Fangblade Eviscerator", "Creature — Human Werewolf", "Creature — Werewolf", "{3}{R}",
		"{1}{R}: This creature gets +1/+0 and gains first strike until end of turn.\n" + dnDay,
		"{1}{R}: This creature gets +1/+0 and gains first strike until end of turn.\n{4}{R}: Creatures you control get +2/+0 until end of turn.\n" + dnNight,
		"3", "4", "4", "5", []string{"R"}, nil, nil, nil},
	{dnaTrespasser, "Graveyard Trespasser", "Graveyard Glutton", "Creature — Human Werewolf", "Creature — Werewolf", "{2}{B}",
		"Ward—Discard a card.\nWhenever this creature enters or attacks, exile up to one target card from a graveyard. If a creature card was exiled this way, each opponent loses 1 life and you gain 1 life.\n" + dnDay,
		"Ward—Discard a card.\nWhenever this creature enters or attacks, exile up to two target cards from graveyards. For each creature card exiled this way, each opponent loses 1 life and you gain 1 life.\n" + dnNight,
		"3", "3", "4", "4", []string{"B"}, []string{"Ward"}, nil, nil},
	{dnaHookhand, "Hookhand Mariner", "Riphook Raider", "Creature — Human Werewolf", "Creature — Werewolf", "{G}{G}{G}{G}",
		dnDay, "This creature can't be blocked by creatures with power 2 or less.\n" + dnNight,
		"4", "4", "6", "4", []string{"G"}, nil, nil, nil},
	{dnaHoundTamer, "Hound Tamer", "Untamed Pup", "Creature — Human Werewolf", "Creature — Werewolf", "{2}{G}",
		"Trample\n{3}{G}: Put a +1/+1 counter on target creature.\n" + dnDay,
		"Trample\nOther Wolves and Werewolves you control have trample.\n{3}{G}: Put a +1/+1 counter on target creature.\n" + dnNight,
		"3", "3", "4", "4", []string{"G"}, []string{"Trample"}, []string{"trample"}, []string{"trample"}},
	{dnaPiper, "Howlpack Piper", "Wildsong Howler", "Creature — Human Werewolf", "Creature — Werewolf", "{G}{G}{G}{G}",
		"This spell can't be countered.\n{1}{G}, {T}: You may put a creature card from your hand onto the battlefield. If it's a Wolf or Werewolf, untap this creature. Activate only as a sorcery.\n" + dnDay,
		"Whenever this creature enters or transforms into Wildsong Howler, look at the top six cards of your library. You may reveal a creature card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.\n" + dnNight,
		"2", "2", "4", "4", []string{"G"}, nil, nil, nil},
	{dnaLoner, "Ill-Tempered Loner", "Howlpack Avenger", "Creature — Human Werewolf", "Creature — Werewolf", "{2}{R}{R}",
		"Whenever this creature is dealt damage, it deals that much damage to any target.\n{1}{R}: This creature gets +2/+0 until end of turn.\n" + dnDay,
		"Whenever a permanent you control is dealt damage, this creature deals that much damage to any target.\n{1}{R}: This creature gets +2/+0 until end of turn.\n" + dnNight,
		"3", "3", "4", "4", []string{"R"}, nil, nil, nil},
	{dnaKessig, "Kessig Naturalist", "Lord of the Ulvenwald", "Creature — Human Werewolf", "Creature — Werewolf", "{R}{G}",
		"Whenever this creature attacks, add {R} or {G}. Until end of turn, you don't lose this mana as steps and phases end.\n" + dnDay,
		"Other Wolves and Werewolves you control get +1/+1.\nWhenever this creature attacks, add {R} or {G}. Until end of turn, you don't lose this mana as steps and phases end.\n" + dnNight,
		"2", "2", "3", "3", []string{"R", "G"}, nil, nil, nil},
	{dnaLambholt, "Lambholt Raconteur", "Lambholt Ravager", "Creature — Human Werewolf", "Creature — Werewolf", "{3}{R}",
		"Whenever you cast a noncreature spell, this creature deals 1 damage to each opponent.\n" + dnDay,
		"Whenever you cast a noncreature spell, this creature deals 2 damage to each opponent.\n" + dnNight,
		"2", "4", "4", "4", []string{"R"}, nil, nil, nil},
}

// dnaTable seats the active player with the werewolf cast at neither day
// nor night (so it becomes day and the creature stays on its front).
func dnaTable(t *testing.T, w dnaWerewolf) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := importAndCast(t, g, w.row(), me)
	monSettle(t, g)
	return g, me, opp, id
}

// dnaPower is current power, counters included (effectivePower is the
// layer result alone).
func dnaPower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return c.CurrentPower()
}

func dnaNight(g *game.Game) { g.WithWriteLock(func() { g.BecomeNightForEffect() }) }
func dnaDay(g *game.Game)   { g.WithWriteLock(func() { g.BecomeDayForEffect() }) }

// Every werewolf in the slice turns over with the designation, its two
// faces are the right cards, and each face's printed keywords ride.
func TestDayNightAWerewolvesTurnOverAndKeepTheirKeywords(t *testing.T) {
	for _, w := range dnaWerewolves {
		t.Run(w.front, func(t *testing.T) {
			spec, ok := Lookup(w.oracle)
			if !ok || spec.Name != w.front {
				t.Fatalf("front face is not registered as %q: %+v", w.front, spec)
			}
			back, ok := Lookup(w.oracle + "#1")
			if !ok || back.Name != w.back {
				t.Fatalf("back face is not registered as %q: %+v", w.back, back)
			}
			if !slices.Contains(spec.PrintedKeywords, "daybound") || !slices.Contains(back.PrintedKeywords, "nightbound") {
				t.Errorf("keywords: front %v back %v, want daybound / nightbound", spec.PrintedKeywords, back.PrintedKeywords)
			}
			g, _, _, id := dnaTable(t, w)
			if c, _ := battlefieldCard(g, id); c.Name != w.front || c.ActiveFace != 0 {
				t.Fatalf("by day: %s face %d, want %s", c.Name, c.ActiveFace, w.front)
			}
			for _, k := range w.frontKeywords {
				if !slices.Contains(effectiveAbilities(t, g, id), k) {
					t.Errorf("front lacks %q: %v", k, effectiveAbilities(t, g, id))
				}
			}
			dnaNight(g)
			if c, _ := battlefieldCard(g, id); c.Name != w.back || c.ActiveFace != 1 {
				t.Fatalf("by night: %s face %d, want %s", c.Name, c.ActiveFace, w.back)
			}
			for _, k := range w.backKeys {
				if !slices.Contains(effectiveAbilities(t, g, id), k) {
					t.Errorf("back lacks %q: %v", k, effectiveAbilities(t, g, id))
				}
			}
			dnaDay(g)
			if c, _ := battlefieldCard(g, id); c.Name != w.front {
				t.Errorf("back by day: %s, want %s", c.Name, w.front)
			}
		})
	}
}

// Hookhand Mariner's back face can't be blocked by power 2 or less.
func TestRiphookRaiderCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := b12Push(g, me.ID, "Riphook Raider", "Creature — Werewolf", dnaHookhand+"#1", 6, 4)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	declareAttack(t, g, opp.ID, raider)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, raider); err == nil {
		t.Error("a power-2 creature blocked Riphook Raider")
	}
	if err := g.DeclareBlocker(big, raider); err != nil {
		t.Errorf("a power-3 creature could not block Riphook Raider: %v", err)
	}
}

// Lambholt: 1 damage by day, 2 by night, to each opponent, per
// noncreature spell, and not for a creature spell.
func TestLambholtPingsEachOpponentForNoncreatureSpells(t *testing.T) {
	w := dnaWerewolves[len(dnaWerewolves)-1]
	g, _, _, _ := dnaTable(t, w)
	before := []int{g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	for i, want := range before {
		if got := g.Seats[i+1].Life; got != want-1 {
			t.Errorf("seat %d life = %d, want %d after one noncreature spell by day", i+1, got, want-1)
		}
	}
	dnaNight(g)
	mid := g.Seats[1].Life
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if got := g.Seats[1].Life; got != mid-2 {
		t.Errorf("life = %d, want %d: Lambholt Ravager deals 2", got, mid-2)
	}
	mid = g.Seats[1].Life
	castCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := g.Seats[1].Life; got != mid {
		t.Errorf("a creature spell cost the opponent %d life", mid-got)
	}
}

// Fangblade Brigand's pump, and the Eviscerator's team pump.
func TestFangbladeFirebreathingAndTheEvisceratorsTeamPump(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	front := b12Push(g, me.ID, "Fangblade Brigand", "Creature — Human Werewolf", dnaFangblade, 3, 4)
	back := b12Push(g, me.ID, "Fangblade Eviscerator", "Creature — Werewolf", dnaFangblade+"#1", 4, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	floatMana(t, g, me, "{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, front, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, front); got != 4 {
		t.Errorf("Brigand power = %d, want 4", got)
	}
	if !slices.Contains(effectiveAbilities(t, g, front), "first strike") {
		t.Error("Brigand did not gain first strike")
	}

	floatMana(t, g, me, "{R}{R}{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, back, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("Bear power = %d, want 4", got)
	}
	if got := effectivePower(t, g, back); got != 6 {
		t.Errorf("Eviscerator power = %d, want 6", got)
	}
	opp := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	if got := effectivePower(t, g, opp); got != 2 {
		t.Errorf("an opposing creature got the pump: %d", got)
	}
}

// Hound Tamer's counter, and Untamed Pup's trample grant to OTHER Wolves
// and Werewolves.
func TestHoundTamerCountersAndUntamedPupGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pup := b12Push(g, me.ID, "Untamed Pup", "Creature — Werewolf", dnaHoundTamer+"#1", 4, 4)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Wolf", "Creature — Wolf", 2, 2)
	if !slices.Contains(effectiveAbilities(t, g, wolf), "trample") {
		t.Error("your Wolf has no trample")
	}
	if slices.Contains(effectiveAbilities(t, g, bear), "trample") || slices.Contains(effectiveAbilities(t, g, theirs), "trample") {
		t.Error("trample reached a Bear or an opponent's Wolf")
	}
	floatMana(t, g, me, "{G}{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, pup, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := dnaPower(t, g, bear); got != 3 {
		t.Errorf("Bear power = %d, want 3 after a counter", got)
	}
}

// Child of the Pack makes a Wolf; Savage Packmate's +1/+0 is for OTHER
// creatures you control only.
func TestChildOfThePackWolvesAndSavagePackmateAnthem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	child := b12Push(g, me.ID, "Child of the Pack", "Creature — Human Werewolf", dnaChildOfPack, 2, 5)
	floatMana(t, g, me, "{R}{R}{R}{G}")
	if err := g.ActivateCatalogAbility(me.ID, child, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Wolf"); n != 1 {
		t.Fatalf("Wolf tokens = %d, want 1", n)
	}

	packmate := b12Push(g, me.ID, "Savage Packmate", "Creature — Werewolf", dnaChildOfPack+"#1", 5, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("your Bear power = %d, want 3", got)
	}
	if got := effectivePower(t, g, packmate); got != 5 {
		t.Errorf("Packmate power = %d, want 5 (it says other)", got)
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("an opposing Bear got the anthem: %d", got)
	}
}

// Lord of the Ulvenwald: other Wolves and Werewolves you control only.
func TestLordOfTheUlvenwaldPumpsOtherWolvesAndWerewolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lord := b12Push(g, me.ID, "Lord of the Ulvenwald", "Creature — Werewolf", dnaKessig+"#1", 3, 3)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if got := effectivePower(t, g, wolf); got != 3 {
		t.Errorf("Wolf power = %d, want 3", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("Bear power = %d, want 2", got)
	}
	if got := effectivePower(t, g, lord); got != 3 {
		t.Errorf("the Lord pumped itself: %d", got)
	}
}

// Kessig Naturalist adds a kept mana when it attacks.
func TestKessigNaturalistAddsKeptManaOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kessig := b12Push(g, me.ID, "Kessig Naturalist", "Creature — Human Werewolf", dnaKessig, 2, 2)
	declareAttack(t, g, opp.ID, kessig)
	passPriorityAroundTable(t, g)
	if b10ResolveAllManaPicks(t, g, me.ID, "R") != 1 {
		t.Fatal("the attack trigger raised no {R}-or-{G} pick")
	}
	advanceTo(t, g, game.StepCombatDamage)
	if got := len(me.ManaPool); got != 1 {
		t.Errorf("mana pool = %d after the combat steps, want the kept {R}", got)
	}
}

// Avabruck Caretaker puts two counters on ANOTHER creature; the back
// face puts two on each creature and hands the others hexproof.
func TestAvabruckCaretakerCountersAndHollowhengeHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	caretaker := b12Push(g, me.ID, "Avabruck Caretaker", "Creature — Human Werewolf", dnaAvabruck, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := dnaPower(t, g, bear); got != 4 {
		t.Errorf("Bear power = %d, want 4", got)
	}
	if got := dnaPower(t, g, caretaker); got != 4 {
		t.Errorf("the Caretaker put counters on itself: %d", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	huntmaster := b12Push(g2, me2.ID, "Hollowhenge Huntmaster", "Creature — Werewolf", dnaAvabruck+"#1", 6, 6)
	bear2 := b12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	relic := b12Permanent(g2, me2.ID, "Relic", "Artifact")
	theirs := b12Creature(g2, g2.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{bear2, relic} {
		if !slices.Contains(effectiveAbilities(t, g2, id), "hexproof") {
			t.Errorf("%s has no hexproof", id)
		}
	}
	if slices.Contains(effectiveAbilities(t, g2, theirs), "hexproof") {
		t.Error("an opponent's creature got hexproof")
	}
	advanceTo(t, g2, game.StepBeginCombat)
	passPriorityAroundTable(t, g2)
	if got := dnaPower(t, g2, bear2); got != 4 {
		t.Errorf("Bear power = %d, want 4", got)
	}
	if got := dnaPower(t, g2, huntmaster); got != 8 {
		t.Errorf("Huntmaster power = %d, want 8: it counts as a creature you control", got)
	}
	if got := dnaPower(t, g2, theirs); got != 2 {
		t.Errorf("an opposing creature got counters: %d", got)
	}
}

// Baneblade Scoundrel shrinks every blocker once, however many.
func TestBaneclawMarauderShrinksBlockersAndDrainsWhenOneDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marauder := b12Push(g, me.ID, "Baneclaw Marauder", "Creature — Werewolf", dnaBaneblade+"#1", 5, 4)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	chump := b12Creature(g, opp.ID, "Chump", "Creature — Bear", 1, 1)
	declareAttack(t, g, opp.ID, marauder)
	advanceTo(t, g, game.StepDeclareBlockers)
	life := opp.Life
	if err := g.DeclareBlocker(wall, marauder); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareBlocker(chump, marauder); err != nil {
		t.Fatal(err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if got := effectiveToughness(t, g, wall); got != 3 {
		t.Errorf("Wall toughness = %d, want 3", got)
	}
	if _, ok := battlefieldCard(g, chump); ok {
		t.Error("the 1/1 blocker survived -1/-1")
	}
	if opp.Life != life-1 {
		t.Errorf("opponent life = %d, want %d: the dead blocker's controller loses 1", opp.Life, life-1)
	}
}

// Ballista Wielder: the ping, and the "can't block" rider only when
// damage was dealt.
func TestBallistaWielderPingsAndStopsTheCreatureBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	wielder := b12Push(g, me.ID, "Ballista Wielder", "Creature — Werewolf", dnaBallista+"#1", 5, 5)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 3, 3)
	other := b12Creature(g, opp.ID, "Other", "Creature — Bear", 3, 3)

	floatMana(t, g, me, "{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, wielder, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: blocker}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err == nil {
		t.Error("the creature the Wielder pinged blocked")
	}
	if err := g.DeclareBlocker(other, attacker); err != nil {
		t.Errorf("a creature that was not pinged could not block: %v", err)
	}
}

func TestBallistaWatcherPingsAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	watcher := b12Push(g, me.ID, "Ballista Watcher", "Creature — Human Werewolf", dnaBallista, 4, 3)
	floatMana(t, g, me, "{R}{R}{R}")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, watcher, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life = %d, want %d", opp.Life, life-1)
	}
}

// Brutal Cathar exiles on entering, returns the creature when it leaves,
// and exiles again when it turns back to its front face.
func TestBrutalCatharExilesOnEnteringAndOnTurningBackToDay(t *testing.T) {
	var w dnaWerewolf
	for _, x := range dnaWerewolves {
		if x.oracle == dnaBrutalCathar {
			w = x
		}
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := b12Creature(g, opp.ID, "First", "Creature — Bear", 2, 2)
	second := b12Creature(g, opp.ID, "Second", "Creature — Bear", 2, 2)
	cathar := importAndCast(t, g, w.row(), me)
	monSettle(t, g)
	answerPickTarget(t, g, first)
	monSettle(t, g)
	if _, ok := battlefieldCard(g, first); ok {
		t.Fatal("the first creature was not exiled")
	}

	dnaNight(g)
	monSettle(t, g)
	if c, _ := battlefieldCard(g, cathar); c.Name != "Moonrage Brute" {
		t.Fatalf("night: %s", c.Name)
	}
	if _, ok := battlefieldCard(g, first); ok {
		t.Fatal("turning over returned the exiled creature: it should stay until Cathar leaves")
	}
	dnaDay(g)
	monSettle(t, g)
	answerPickTarget(t, g, second)
	monSettle(t, g)
	if _, ok := battlefieldCard(g, second); ok {
		t.Error("transforming into Brutal Cathar did not exile a second creature")
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(cathar); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	back := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == opp.ID && c.IsCreature() {
			back++
		}
	}
	if back != 2 {
		t.Errorf("creatures returned = %d, want both exiled creatures back (as new objects)", back)
	}
}

// Cast at night, Brutal Cathar enters as Moonrage Brute and exiles
// nothing.
func TestBrutalCatharCastAtNightExilesNothing(t *testing.T) {
	var w dnaWerewolf
	for _, x := range dnaWerewolves {
		if x.oracle == dnaBrutalCathar {
			w = x
		}
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationNight })
	importAndCast(t, g, w.row(), me)
	monSettle(t, g)
	if _, ok := battlefieldCard(g, victim); !ok {
		t.Error("Moonrage Brute entering exiled a creature")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("a prompt is open: %+v", g.PendingChoices)
	}
}

// Graveyard Trespasser: a creature card exiled drains, a noncreature
// card does not.
func TestGraveyardTrespasserDrainsOnlyForACreatureCard(t *testing.T) {
	var w dnaWerewolf
	for _, x := range dnaWerewolves {
		if x.oracle == dnaTrespasser {
			w = x
		}
	}
	for _, tc := range []struct {
		name, typeLine string
		drain          bool
	}{{"Bear", "Creature — Bear", true}, {"Shock", "Instant", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			card := pushGraveyardPermanent(opp, tc.name, tc.typeLine, "{1}")
			life, myLife := opp.Life, me.Life
			importAndCast(t, g, w.row(), me)
			monSettle(t, g)
			answerPickTarget(t, g, card)
			monSettle(t, g)
			if opp.Graveyard.Contains(card) {
				t.Fatal("the card was not exiled")
			}
			wantOpp, wantMe := life, myLife
			if tc.drain {
				wantOpp, wantMe = life-1, myLife+1
			}
			if opp.Life != wantOpp || me.Life != wantMe {
				t.Errorf("life: opp %d me %d, want %d / %d", opp.Life, me.Life, wantOpp, wantMe)
			}
		})
	}
}

// Howlpack Piper can't be countered, and its put untaps it only for a
// Wolf or Werewolf.
func TestHowlpackPiperPutsACreatureAndUntapsForAWerewolf(t *testing.T) {
	if spec, _ := Lookup(dnaPiper); !spec.CantBeCountered {
		t.Error("Howlpack Piper can be countered")
	}
	for _, tc := range []struct {
		typeLine string
		untapped bool
	}{{"Creature — Werewolf", true}, {"Creature — Bear", false}} {
		t.Run(tc.typeLine, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			piper := b12Push(g, me.ID, "Howlpack Piper", "Creature — Human Werewolf", dnaPiper, 2, 2)
			card := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: card, Name: "Hand Creature", TypeLine: tc.typeLine, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
			floatMana(t, g, me, "{G}{G}")
			if err := g.ActivateCatalogAbility(me.ID, piper, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			pick := latestChooseCards(g, me.ID)
			if pick == nil {
				t.Fatal("no pick of a creature card from hand")
			}
			if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{card}); err != nil {
				t.Fatal(err)
			}
			if _, ok := battlefieldCard(g, card); !ok {
				t.Fatal("the creature did not enter")
			}
			tapped, _ := battlefieldCardTapped(g, piper)
			if tapped == tc.untapped {
				t.Errorf("Piper tapped = %v, want %v", tapped, !tc.untapped)
			}
		})
	}
}

// Wildsong Howler, cast at night, looks at six and may take a creature.
func TestWildsongHowlerDigsWhenItEnters(t *testing.T) {
	var w dnaWerewolf
	for _, x := range dnaWerewolves {
		if x.oracle == dnaPiper {
			w = x
		}
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationNight })
	bear := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	importAndCast(t, g, w.row(), me)
	monSettle(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no dig prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(bear) {
		t.Error("the creature did not reach the hand")
	}
}

// Ill-Tempered Loner reflects the damage it is dealt at any target.
func TestIllTemperedLonerReflectsDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loner := b12Push(g, me.ID, "Ill-Tempered Loner", "Creature — Human Werewolf", dnaLoner, 3, 3)
	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, loner, 2); err != nil {
			t.Fatal(err)
		}
	})
	monSettle(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	monSettle(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life = %d, want %d", opp.Life, life-2)
	}
}

// Howlpack Avenger reflects damage dealt to ANY permanent you control.
func TestHowlpackAvengerReflectsDamageToOtherPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Howlpack Avenger", "Creature — Werewolf", dnaLoner+"#1", 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, theirs, 1); err != nil {
			t.Fatal(err)
		}
	})
	monSettle(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatal("damage to an opponent's creature triggered the Avenger")
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, bear, 2); err != nil {
			t.Fatal(err)
		}
	})
	monSettle(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	monSettle(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life = %d, want %d", opp.Life, life-2)
	}
}

// Burly Breaker, Moonrage Brute and Gavony Dawnguard carry their ward.
func TestDayNightAWardCostsAreDeclared(t *testing.T) {
	for _, key := range []string{dnaBurlyBreaker, dnaBurlyBreaker + "#1", dnaBrutalCathar + "#1", dnaDawnguard} {
		spec, ok := Lookup(key)
		if !ok || len(spec.Triggered) == 0 {
			t.Errorf("%s has no triggered ability (ward)", key)
		}
	}
}

// The day/night payoffs: neither -> day on entering, then a flip fires.
func TestComponentCollectorTapsOrUntapsOnAFlip(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dnCast(t, g, "Component Collector", "Creature — Homunculus", dnaCollector)
	if !g.IsDay() {
		t.Fatal("it did not become day")
	}
	target := b12Permanent(g, opp.ID, "Relic", "Artifact")
	dnaNight(g)
	monSettle(t, g)
	answerPickTarget(t, g, target)
	monSettle(t, g)
	answerOptionPick(t, g, me.ID, 0) // tap it
	if tapped, _ := battlefieldCardTapped(g, target); !tapped {
		t.Error("the Relic was not tapped")
	}
}

func TestComponentCollectorCannotTargetALand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	dnCast(t, g, "Component Collector", "Creature — Homunculus", dnaCollector)
	island := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	dnaNight(g)
	monSettle(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && slices.Contains(c.PickTargetCards, island) {
			t.Errorf("the land was offered as a target: %v", c.PickTargetCards)
		}
	}
}

func TestCelestusSanctifierPutsOneOfTheTopTwoInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dnCast(t, g, "Celestus Sanctifier", "Creature — Human Cleric", dnaSanctifier)
	if !g.IsDay() {
		t.Fatal("it did not become day")
	}
	a, b := uuid.New(), uuid.New()
	me.Library.PushTop(game.Card{InstanceID: b, Name: "B", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: a, Name: "A", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	dnaNight(g)
	monSettle(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 2 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("pick = %+v, want exactly one of the top two", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{b}); err != nil {
		t.Fatal(err)
	}
	if !me.Graveyard.Contains(b) {
		t.Error("the chosen card is not in the graveyard")
	}
	if !me.Library.Contains(a) {
		t.Error("the other card left the library")
	}
}

func TestGavonyDawnguardTakesACreatureOfManaValueThreeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dnCast(t, g, "Gavony Dawnguard", "Creature — Human Soldier", dnaDawnguard)
	small, big := uuid.New(), uuid.New()
	me.Library.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{5}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: small, Name: "Small", TypeLine: "Creature — Elf", ManaCost: "{2}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	dnaNight(g)
	monSettle(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no dig prompt")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != small {
		t.Fatalf("offered %v, want only the mana value 2 creature", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{small}); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(small) {
		t.Error("the creature did not reach the hand")
	}
}

// Arlinn: the back face's loyalty abilities, and the animation.
func TestArlinnTheMoonsFuryAnimatesAndAddsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	arlinn := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: arlinn, Name: "Arlinn, the Moon's Fury", TypeLine: "Legendary Planeswalker — Arlinn",
		OracleID: dnaArlinn + "#1", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	if err := g.ActivateCatalogAbility(me.ID, arlinn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, arlinn); got != 5 {
		t.Errorf("power = %d, want 5", got)
	}
	abilities := effectiveAbilities(t, g, arlinn)
	for _, k := range []string{"trample", "indestructible", "haste"} {
		if !slices.Contains(abilities, k) {
			t.Errorf("missing %s: %v", k, abilities)
		}
	}
	if types := effectiveTypes(t, g, arlinn); !slices.Contains(types, "Creature") || !slices.Contains(types, "Planeswalker") {
		t.Errorf("types = %v, want a planeswalker creature", types)
	}
}

func TestArlinnThePacksHopeMakesWolvesAndGrantsFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	arlinn := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: arlinn, Name: "Arlinn, the Pack's Hope", TypeLine: "Legendary Planeswalker — Arlinn",
		OracleID: dnaArlinn, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	if err := g.ActivateCatalogAbility(me.ID, arlinn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Wolf"); n != 2 {
		t.Errorf("Wolf tokens = %d, want 2", n)
	}
	spec, _ := Lookup(dnaArlinn)
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Errorf("Arlinn keeps one caveat for the missing counter half: %v", spec.Caveats)
	}
}
