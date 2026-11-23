package testscr

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

var dnaWerewolves = []effectstest.DnaWerewolf{
	{Oracle: effectstest.DnaAvabruck, Front: "Avabruck Caretaker", Back: "Hollowhenge Huntmaster", FrontType: "Creature — Human Werewolf", BackType: "Creature — Werewolf", Cost: "{4}{G}{G}",
		FrontText: "Hexproof\nAt the beginning of combat on your turn, put two +1/+1 counters on another target creature you control.\n" + effectstest.DnDay,
		BackText:  "Hexproof\nOther permanents you control have hexproof.\nAt the beginning of combat on your turn, put two +1/+1 counters on each creature you control.\n" + effectstest.DnNight,
		Fp:        "4", Ft: "4", Bp: "6", Bt: "6", Colors: []string{"G"}, ExtraKeywords: []string{"Hexproof"}, FrontKeywords: []string{"hexproof"}, BackKeys: []string{"hexproof"}},
	{Oracle: effectstest.DnaBallista, Front: "Ballista Watcher", Back: "Ballista Wielder", FrontType: "Creature — Human Soldier Werewolf", BackType: "Creature — Werewolf", Cost: "{2}{R}{R}",
		FrontText: "{2}{R}, {T}: This creature deals 1 damage to any target.\n" + effectstest.DnDay,
		BackText:  "{2}{R}: This creature deals 1 damage to any target. A creature dealt damage this way can't block this turn.\n" + effectstest.DnNight,
		Fp:        "4", Ft: "3", Bp: "5", Bt: "5", Colors: []string{"R"}, ExtraKeywords: nil, FrontKeywords: nil, BackKeys: nil},
	{Oracle: effectstest.DnaBaneblade, Front: "Baneblade Scoundrel", Back: "Baneclaw Marauder", FrontType: "Creature — Human Rogue Werewolf", BackType: "Creature — Werewolf", Cost: "{3}{B}",
		FrontText: "Whenever this creature becomes blocked, each creature blocking it gets -1/-1 until end of turn.\n" + effectstest.DnDay,
		BackText:  "Whenever this creature becomes blocked, each creature blocking it gets -1/-1 until end of turn.\nWhenever a creature blocking this creature dies, that creature's controller loses 1 life.\n" + effectstest.DnNight,
		Fp:        "4", Ft: "3", Bp: "5", Bt: "4", Colors: []string{"B"}, ExtraKeywords: nil, FrontKeywords: nil, BackKeys: nil},
	{Oracle: effectstest.DnaBrutalCathar, Front: "Brutal Cathar", Back: "Moonrage Brute", FrontType: "Creature — Human Soldier Werewolf", BackType: "Creature — Werewolf", Cost: "{2}{W}",
		FrontText: "Whenever this creature enters or transforms into Brutal Cathar, exile target creature an opponent controls until this creature leaves the battlefield.\n" + effectstest.DnDay,
		BackText:  "First strike\nWard—Pay 3 life.\n" + effectstest.DnNight,
		Fp:        "2", Ft: "2", Bp: "3", Bt: "3", Colors: []string{"W"}, ExtraKeywords: []string{"First strike", "Ward"}, FrontKeywords: nil, BackKeys: []string{"first strike"}},
	{Oracle: effectstest.DnaBurlyBreaker, Front: "Burly Breaker", Back: "Dire-Strain Demolisher", FrontType: "Creature — Human Werewolf", BackType: "Creature — Werewolf", Cost: "{3}{G}{G}",
		FrontText: "Ward {1}\n" + effectstest.DnDay, BackText: "Ward {3}\n" + effectstest.DnNight,
		Fp: "6", Ft: "5", Bp: "8", Bt: "7", Colors: []string{"G"}, ExtraKeywords: []string{"Ward"}, FrontKeywords: nil, BackKeys: nil},
	{Oracle: effectstest.DnaChildOfPack, Front: "Child of the Pack", Back: "Savage Packmate", FrontType: "Creature — Human Werewolf", BackType: "Creature — Werewolf", Cost: "{R}{R}{R}{G}",
		FrontText: "{2}{R}{G}: Create a 2/2 green Wolf creature token.\n" + effectstest.DnDay,
		BackText:  "Trample\nOther creatures you control get +1/+0.\n" + effectstest.DnNight,
		Fp:        "2", Ft: "5", Bp: "5", Bt: "5", Colors: []string{"R", "G"}, ExtraKeywords: []string{"Trample"}, FrontKeywords: nil, BackKeys: []string{"trample"}},
	{Oracle: effectstest.DnaFangblade, Front: "Fangblade Brigand", Back: "Fangblade Eviscerator", FrontType: "Creature — Human Werewolf", BackType: "Creature — Werewolf", Cost: "{3}{R}",
		FrontText: "{1}{R}: This creature gets +1/+0 and gains first strike until end of turn.\n" + effectstest.DnDay,
		BackText:  "{1}{R}: This creature gets +1/+0 and gains first strike until end of turn.\n{4}{R}: Creatures you control get +2/+0 until end of turn.\n" + effectstest.DnNight,
		Fp:        "3", Ft: "4", Bp: "4", Bt: "5", Colors: []string{"R"}, ExtraKeywords: nil, FrontKeywords: nil, BackKeys: nil},
	{Oracle: effectstest.DnaTrespasser, Front: "Graveyard Trespasser", Back: "Graveyard Glutton", FrontType: "Creature — Human Werewolf", BackType: "Creature — Werewolf", Cost: "{2}{B}",
		FrontText: "Ward—Discard a card.\nWhenever this creature enters or attacks, exile up to one target card from a graveyard. If a creature card was exiled this way, each opponent loses 1 life and you gain 1 life.\n" + effectstest.DnDay,
		BackText:  "Ward—Discard a card.\nWhenever this creature enters or attacks, exile up to two target cards from graveyards. For each creature card exiled this way, each opponent loses 1 life and you gain 1 life.\n" + effectstest.DnNight,
		Fp:        "3", Ft: "3", Bp: "4", Bt: "4", Colors: []string{"B"}, ExtraKeywords: []string{"Ward"}, FrontKeywords: nil, BackKeys: nil},
	{Oracle: effectstest.DnaHookhand, Front: "Hookhand Mariner", Back: "Riphook Raider", FrontType: "Creature — Human Werewolf", BackType: "Creature — Werewolf", Cost: "{G}{G}{G}{G}",
		FrontText: effectstest.DnDay, BackText: "This creature can't be blocked by creatures with power 2 or less.\n" + effectstest.DnNight,
		Fp: "4", Ft: "4", Bp: "6", Bt: "4", Colors: []string{"G"}, ExtraKeywords: nil, FrontKeywords: nil, BackKeys: nil},
	{Oracle: effectstest.DnaHoundTamer, Front: "Hound Tamer", Back: "Untamed Pup", FrontType: "Creature — Human Werewolf", BackType: "Creature — Werewolf", Cost: "{2}{G}",
		FrontText: "Trample\n{3}{G}: Put a +1/+1 counter on target creature.\n" + effectstest.DnDay,
		BackText:  "Trample\nOther Wolves and Werewolves you control have trample.\n{3}{G}: Put a +1/+1 counter on target creature.\n" + effectstest.DnNight,
		Fp:        "3", Ft: "3", Bp: "4", Bt: "4", Colors: []string{"G"}, ExtraKeywords: []string{"Trample"}, FrontKeywords: []string{"trample"}, BackKeys: []string{"trample"}},
	{Oracle: effectstest.DnaPiper, Front: "Howlpack Piper", Back: "Wildsong Howler", FrontType: "Creature — Human Werewolf", BackType: "Creature — Werewolf", Cost: "{G}{G}{G}{G}",
		FrontText: "This spell can't be countered.\n{1}{G}, {T}: You may put a creature card from your hand onto the battlefield. If it's a Wolf or Werewolf, untap this creature. Activate only as a sorcery.\n" + effectstest.DnDay,
		BackText:  "Whenever this creature enters or transforms into Wildsong Howler, look at the top six cards of your library. You may reveal a creature card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.\n" + effectstest.DnNight,
		Fp:        "2", Ft: "2", Bp: "4", Bt: "4", Colors: []string{"G"}, ExtraKeywords: nil, FrontKeywords: nil, BackKeys: nil},
	{Oracle: effectstest.DnaLoner, Front: "Ill-Tempered Loner", Back: "Howlpack Avenger", FrontType: "Creature — Human Werewolf", BackType: "Creature — Werewolf", Cost: "{2}{R}{R}",
		FrontText: "Whenever this creature is dealt damage, it deals that much damage to any target.\n{1}{R}: This creature gets +2/+0 until end of turn.\n" + effectstest.DnDay,
		BackText:  "Whenever a permanent you control is dealt damage, this creature deals that much damage to any target.\n{1}{R}: This creature gets +2/+0 until end of turn.\n" + effectstest.DnNight,
		Fp:        "3", Ft: "3", Bp: "4", Bt: "4", Colors: []string{"R"}, ExtraKeywords: nil, FrontKeywords: nil, BackKeys: nil},
	{Oracle: effectstest.DnaKessig, Front: "Kessig Naturalist", Back: "Lord of the Ulvenwald", FrontType: "Creature — Human Werewolf", BackType: "Creature — Werewolf", Cost: "{R}{G}",
		FrontText: "Whenever this creature attacks, add {R} or {G}. Until end of turn, you don't lose this mana as steps and phases end.\n" + effectstest.DnDay,
		BackText:  "Other Wolves and Werewolves you control get +1/+1.\nWhenever this creature attacks, add {R} or {G}. Until end of turn, you don't lose this mana as steps and phases end.\n" + effectstest.DnNight,
		Fp:        "2", Ft: "2", Bp: "3", Bt: "3", Colors: []string{"R", "G"}, ExtraKeywords: nil, FrontKeywords: nil, BackKeys: nil},
	{Oracle: effectstest.DnaLambholt, Front: "Lambholt Raconteur", Back: "Lambholt Ravager", FrontType: "Creature — Human Werewolf", BackType: "Creature — Werewolf", Cost: "{3}{R}",
		FrontText: "Whenever you cast a noncreature spell, this creature deals 1 damage to each opponent.\n" + effectstest.DnDay,
		BackText:  "Whenever you cast a noncreature spell, this creature deals 2 damage to each opponent.\n" + effectstest.DnNight,
		Fp:        "2", Ft: "4", Bp: "4", Bt: "4", Colors: []string{"R"}, ExtraKeywords: nil, FrontKeywords: nil, BackKeys: nil},
}

// dnaTable seats the active player with the werewolf cast at neither day
// nor night (so it becomes day and the creature stays on its front).
func dnaTable(t *testing.T, w effectstest.DnaWerewolf) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.ImportAndCast(t, g, w.Row(), me)
	effectstest.MonSettle(t, g)
	return g, me, opp, id
}

func dnaNight(g *game.Game) { g.WithWriteLock(func() { g.BecomeNightForEffect() }) }
func dnaDay(g *game.Game)   { g.WithWriteLock(func() { g.BecomeDayForEffect() }) }

// Every werewolf in the slice turns over with the designation, its two
// faces are the right cards, and each face's printed keywords ride.
func TestDayNightAWerewolvesTurnOverAndKeepTheirKeywords(t *testing.T) {
	for _, w := range dnaWerewolves {
		t.Run(w.Front, func(t *testing.T) {
			spec, ok := effects.Lookup(w.Oracle)
			if !ok || spec.Name != w.Front {
				t.Fatalf("front face is not registered as %q: %+v", w.Front, spec)
			}
			back, ok := effects.Lookup(w.Oracle + "#1")
			if !ok || back.Name != w.Back {
				t.Fatalf("back face is not registered as %q: %+v", w.Back, back)
			}
			if !slices.Contains(spec.PrintedKeywords, "daybound") || !slices.Contains(back.PrintedKeywords, "nightbound") {
				t.Errorf("keywords: front %v back %v, want daybound / nightbound", spec.PrintedKeywords, back.PrintedKeywords)
			}
			g, _, _, id := dnaTable(t, w)
			if c, _ := effectstest.BattlefieldCard(g, id); c.Name != w.Front || c.ActiveFace != 0 {
				t.Fatalf("by day: %s face %d, want %s", c.Name, c.ActiveFace, w.Front)
			}
			for _, k := range w.FrontKeywords {
				if !slices.Contains(effectstest.EffectiveAbilities(t, g, id), k) {
					t.Errorf("front lacks %q: %v", k, effectstest.EffectiveAbilities(t, g, id))
				}
			}
			dnaNight(g)
			if c, _ := effectstest.BattlefieldCard(g, id); c.Name != w.Back || c.ActiveFace != 1 {
				t.Fatalf("by night: %s face %d, want %s", c.Name, c.ActiveFace, w.Back)
			}
			for _, k := range w.BackKeys {
				if !slices.Contains(effectstest.EffectiveAbilities(t, g, id), k) {
					t.Errorf("back lacks %q: %v", k, effectstest.EffectiveAbilities(t, g, id))
				}
			}
			dnaDay(g)
			if c, _ := effectstest.BattlefieldCard(g, id); c.Name != w.Front {
				t.Errorf("back by day: %s, want %s", c.Name, w.Front)
			}
		})
	}
}

// Hookhand Mariner's back face can't be blocked by power 2 or less.
func TestRiphookRaiderCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := effectstest.B12Push(g, me.ID, "Riphook Raider", "Creature — Werewolf", effectstest.DnaHookhand+"#1", 6, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, raider)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
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
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, want := range before {
		if got := g.Seats[i+1].Life; got != want-1 {
			t.Errorf("seat %d life = %d, want %d after one noncreature spell by day", i+1, got, want-1)
		}
	}
	dnaNight(g)
	mid := g.Seats[1].Life
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.Seats[1].Life; got != mid-2 {
		t.Errorf("life = %d, want %d: Lambholt Ravager deals 2", got, mid-2)
	}
	mid = g.Seats[1].Life
	effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.Seats[1].Life; got != mid {
		t.Errorf("a creature spell cost the opponent %d life", mid-got)
	}
}

// Fangblade Brigand's pump, and the Eviscerator's team pump.
func TestFangbladeFirebreathingAndTheEvisceratorsTeamPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	front := effectstest.B12Push(g, me.ID, "Fangblade Brigand", "Creature — Human Werewolf", effectstest.DnaFangblade, 3, 4)
	back := effectstest.B12Push(g, me.ID, "Fangblade Eviscerator", "Creature — Werewolf", effectstest.DnaFangblade+"#1", 4, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, front, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, front); got != 4 {
		t.Errorf("Brigand power = %d, want 4", got)
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, front), "first strike") {
		t.Error("Brigand did not gain first strike")
	}

	effectstest.FloatMana(t, g, me, "{R}{R}{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, back, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("Bear power = %d, want 4", got)
	}
	if got := effectstest.EffectivePower(t, g, back); got != 6 {
		t.Errorf("Eviscerator power = %d, want 6", got)
	}
	opp := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	if got := effectstest.EffectivePower(t, g, opp); got != 2 {
		t.Errorf("an opposing creature got the pump: %d", got)
	}
}

// Hound Tamer's counter, and Untamed Pup's trample grant to OTHER Wolves
// and Werewolves.
func TestHoundTamerCountersAndUntamedPupGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	pup := effectstest.B12Push(g, me.ID, "Untamed Pup", "Creature — Werewolf", effectstest.DnaHoundTamer+"#1", 4, 4)
	wolf := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Wolf", "Creature — Wolf", 2, 2)
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, wolf), "trample") {
		t.Error("your Wolf has no trample")
	}
	if slices.Contains(effectstest.EffectiveAbilities(t, g, bear), "trample") || slices.Contains(effectstest.EffectiveAbilities(t, g, theirs), "trample") {
		t.Error("trample reached a Bear or an opponent's Wolf")
	}
	effectstest.FloatMana(t, g, me, "{G}{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, pup, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DnaPower(t, g, bear); got != 3 {
		t.Errorf("Bear power = %d, want 3 after a counter", got)
	}
}

// Child of the Pack makes a Wolf; Savage Packmate's +1/+0 is for OTHER
// creatures you control only.
func TestChildOfThePackWolvesAndSavagePackmateAnthem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	child := effectstest.B12Push(g, me.ID, "Child of the Pack", "Creature — Human Werewolf", effectstest.DnaChildOfPack, 2, 5)
	effectstest.FloatMana(t, g, me, "{R}{R}{R}{G}")
	if err := g.ActivateCatalogAbility(me.ID, child, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Wolf"); n != 1 {
		t.Fatalf("Wolf tokens = %d, want 1", n)
	}

	packmate := effectstest.B12Push(g, me.ID, "Savage Packmate", "Creature — Werewolf", effectstest.DnaChildOfPack+"#1", 5, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("your Bear power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, packmate); got != 5 {
		t.Errorf("Packmate power = %d, want 5 (it says other)", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("an opposing Bear got the anthem: %d", got)
	}
}

// Lord of the Ulvenwald: other Wolves and Werewolves you control only.
func TestLordOfTheUlvenwaldPumpsOtherWolvesAndWerewolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lord := effectstest.B12Push(g, me.ID, "Lord of the Ulvenwald", "Creature — Werewolf", effectstest.DnaKessig+"#1", 3, 3)
	wolf := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if got := effectstest.EffectivePower(t, g, wolf); got != 3 {
		t.Errorf("Wolf power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("Bear power = %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, lord); got != 3 {
		t.Errorf("the Lord pumped itself: %d", got)
	}
}

// Kessig Naturalist adds a kept mana when it attacks.
func TestKessigNaturalistAddsKeptManaOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kessig := effectstest.B12Push(g, me.ID, "Kessig Naturalist", "Creature — Human Werewolf", effectstest.DnaKessig, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, kessig)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B10ResolveAllManaPicks(t, g, me.ID, "R") != 1 {
		t.Fatal("the attack trigger raised no {R}-or-{G} pick")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if got := len(me.ManaPool); got != 1 {
		t.Errorf("mana pool = %d after the combat steps, want the kept {R}", got)
	}
}

// Avabruck Caretaker puts two counters on ANOTHER creature; the back
// face puts two on each creature and hands the others hexproof.
func TestAvabruckCaretakerCountersAndHollowhengeHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	caretaker := effectstest.B12Push(g, me.ID, "Avabruck Caretaker", "Creature — Human Werewolf", effectstest.DnaAvabruck, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DnaPower(t, g, bear); got != 4 {
		t.Errorf("Bear power = %d, want 4", got)
	}
	if got := effectstest.DnaPower(t, g, caretaker); got != 4 {
		t.Errorf("the Caretaker put counters on itself: %d", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	huntmaster := effectstest.B12Push(g2, me2.ID, "Hollowhenge Huntmaster", "Creature — Werewolf", effectstest.DnaAvabruck+"#1", 6, 6)
	bear2 := effectstest.B12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	relic := effectstest.B12Permanent(g2, me2.ID, "Relic", "Artifact")
	theirs := effectstest.B12Creature(g2, g2.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{bear2, relic} {
		if !slices.Contains(effectstest.EffectiveAbilities(t, g2, id), "hexproof") {
			t.Errorf("%s has no hexproof", id)
		}
	}
	if slices.Contains(effectstest.EffectiveAbilities(t, g2, theirs), "hexproof") {
		t.Error("an opponent's creature got hexproof")
	}
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.DnaPower(t, g2, bear2); got != 4 {
		t.Errorf("Bear power = %d, want 4", got)
	}
	if got := effectstest.DnaPower(t, g2, huntmaster); got != 8 {
		t.Errorf("Huntmaster power = %d, want 8: it counts as a creature you control", got)
	}
	if got := effectstest.DnaPower(t, g2, theirs); got != 2 {
		t.Errorf("an opposing creature got counters: %d", got)
	}
}

// Baneblade Scoundrel shrinks every blocker once, however many.
func TestBaneclawMarauderShrinksBlockersAndDrainsWhenOneDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marauder := effectstest.B12Push(g, me.ID, "Baneclaw Marauder", "Creature — Werewolf", effectstest.DnaBaneblade+"#1", 5, 4)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	chump := effectstest.B12Creature(g, opp.ID, "Chump", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, marauder)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	life := opp.Life
	if err := g.DeclareBlocker(wall, marauder); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareBlocker(chump, marauder); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveToughness(t, g, wall); got != 3 {
		t.Errorf("Wall toughness = %d, want 3", got)
	}
	if _, ok := effectstest.BattlefieldCard(g, chump); ok {
		t.Error("the 1/1 blocker survived -1/-1")
	}
	if opp.Life != life-1 {
		t.Errorf("opponent life = %d, want %d: the dead blocker's controller loses 1", opp.Life, life-1)
	}
}

// Ballista Wielder: the ping, and the "can't block" rider only when
// damage was dealt.
func TestBallistaWielderPingsAndStopsTheCreatureBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wielder := effectstest.B12Push(g, me.ID, "Ballista Wielder", "Creature — Werewolf", effectstest.DnaBallista+"#1", 5, 5)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 3, 3)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 3, 3)

	effectstest.FloatMana(t, g, me, "{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, wielder, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: blocker}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err == nil {
		t.Error("the creature the Wielder pinged blocked")
	}
	if err := g.DeclareBlocker(other, attacker); err != nil {
		t.Errorf("a creature that was not pinged could not block: %v", err)
	}
}

func TestBallistaWatcherPingsAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	watcher := effectstest.B12Push(g, me.ID, "Ballista Watcher", "Creature — Human Werewolf", effectstest.DnaBallista, 4, 3)
	effectstest.FloatMana(t, g, me, "{R}{R}{R}")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, watcher, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life = %d, want %d", opp.Life, life-1)
	}
}

// Brutal Cathar exiles on entering, returns the creature when it leaves,
// and exiles again when it turns back to its front face.
func TestBrutalCatharExilesOnEnteringAndOnTurningBackToDay(t *testing.T) {
	var w effectstest.DnaWerewolf
	for _, x := range dnaWerewolves {
		if x.Oracle == effectstest.DnaBrutalCathar {
			w = x
		}
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.B12Creature(g, opp.ID, "First", "Creature — Bear", 2, 2)
	second := effectstest.B12Creature(g, opp.ID, "Second", "Creature — Bear", 2, 2)
	cathar := effectstest.ImportAndCast(t, g, w.Row(), me)
	effectstest.MonSettle(t, g)
	effectstest.AnswerPickTarget(t, g, first)
	effectstest.MonSettle(t, g)
	if _, ok := effectstest.BattlefieldCard(g, first); ok {
		t.Fatal("the first creature was not exiled")
	}

	dnaNight(g)
	effectstest.MonSettle(t, g)
	if c, _ := effectstest.BattlefieldCard(g, cathar); c.Name != "Moonrage Brute" {
		t.Fatalf("night: %s", c.Name)
	}
	if _, ok := effectstest.BattlefieldCard(g, first); ok {
		t.Fatal("turning over returned the exiled creature: it should stay until Cathar leaves")
	}
	dnaDay(g)
	effectstest.MonSettle(t, g)
	effectstest.AnswerPickTarget(t, g, second)
	effectstest.MonSettle(t, g)
	if _, ok := effectstest.BattlefieldCard(g, second); ok {
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
	var w effectstest.DnaWerewolf
	for _, x := range dnaWerewolves {
		if x.Oracle == effectstest.DnaBrutalCathar {
			w = x
		}
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationNight })
	effectstest.ImportAndCast(t, g, w.Row(), me)
	effectstest.MonSettle(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); !ok {
		t.Error("Moonrage Brute entering exiled a creature")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("a prompt is open: %+v", g.PendingChoices)
	}
}

// Graveyard Trespasser: a creature card exiled drains, a noncreature
// card does not.
func TestGraveyardTrespasserDrainsOnlyForACreatureCard(t *testing.T) {
	var w effectstest.DnaWerewolf
	for _, x := range dnaWerewolves {
		if x.Oracle == effectstest.DnaTrespasser {
			w = x
		}
	}
	for _, tc := range []struct {
		name, typeLine string
		drain          bool
	}{{"Bear", "Creature — Bear", true}, {"Shock", "Instant", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			card := effectstest.PushGraveyardPermanent(opp, tc.name, tc.typeLine, "{1}")
			life, myLife := opp.Life, me.Life
			effectstest.ImportAndCast(t, g, w.Row(), me)
			effectstest.MonSettle(t, g)
			effectstest.AnswerPickTarget(t, g, card)
			effectstest.MonSettle(t, g)
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
	if spec, _ := effects.Lookup(effectstest.DnaPiper); !spec.CantBeCountered {
		t.Error("Howlpack Piper can be countered")
	}
	for _, tc := range []struct {
		typeLine string
		untapped bool
	}{{"Creature — Werewolf", true}, {"Creature — Bear", false}} {
		t.Run(tc.typeLine, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			piper := effectstest.B12Push(g, me.ID, "Howlpack Piper", "Creature — Human Werewolf", effectstest.DnaPiper, 2, 2)
			card := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: card, Name: "Hand Creature", TypeLine: tc.typeLine, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
			effectstest.FloatMana(t, g, me, "{G}{G}")
			if err := g.ActivateCatalogAbility(me.ID, piper, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			pick := effectstest.LatestChooseCards(g, me.ID)
			if pick == nil {
				t.Fatal("no pick of a creature card from hand")
			}
			if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{card}); err != nil {
				t.Fatal(err)
			}
			if _, ok := effectstest.BattlefieldCard(g, card); !ok {
				t.Fatal("the creature did not enter")
			}
			tapped, _ := effectstest.BattlefieldCardTapped(g, piper)
			if tapped == tc.untapped {
				t.Errorf("Piper tapped = %v, want %v", tapped, !tc.untapped)
			}
		})
	}
}

// Wildsong Howler, cast at night, looks at six and may take a creature.
func TestWildsongHowlerDigsWhenItEnters(t *testing.T) {
	var w effectstest.DnaWerewolf
	for _, x := range dnaWerewolves {
		if x.Oracle == effectstest.DnaPiper {
			w = x
		}
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationNight })
	bear := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.ImportAndCast(t, g, w.Row(), me)
	effectstest.MonSettle(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loner := effectstest.B12Push(g, me.ID, "Ill-Tempered Loner", "Creature — Human Werewolf", effectstest.DnaLoner, 3, 3)
	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, loner, 2); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.MonSettle(t, g)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.MonSettle(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life = %d, want %d", opp.Life, life-2)
	}
}

// Howlpack Avenger reflects damage dealt to ANY permanent you control.
func TestHowlpackAvengerReflectsDamageToOtherPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Howlpack Avenger", "Creature — Werewolf", effectstest.DnaLoner+"#1", 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, theirs, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.MonSettle(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatal("damage to an opponent's creature triggered the Avenger")
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, bear, 2); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.MonSettle(t, g)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.MonSettle(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life = %d, want %d", opp.Life, life-2)
	}
}

// Burly Breaker, Moonrage Brute and Gavony Dawnguard carry their ward.
func TestDayNightAWardCostsAreDeclared(t *testing.T) {
	for _, key := range []string{effectstest.DnaBurlyBreaker, effectstest.DnaBurlyBreaker + "#1", effectstest.DnaBrutalCathar + "#1", effectstest.DnaDawnguard} {
		spec, ok := effects.Lookup(key)
		if !ok || len(spec.Triggered) == 0 {
			t.Errorf("%s has no triggered ability (ward)", key)
		}
	}
}

// The day/night payoffs: neither -> day on entering, then a flip fires.
func TestComponentCollectorTapsOrUntapsOnAFlip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dnCast(t, g, "Component Collector", "Creature — Homunculus", effectstest.DnaCollector)
	if !g.IsDay() {
		t.Fatal("it did not become day")
	}
	target := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	dnaNight(g)
	effectstest.MonSettle(t, g)
	effectstest.AnswerPickTarget(t, g, target)
	effectstest.MonSettle(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // tap it
	if tapped, _ := effectstest.BattlefieldCardTapped(g, target); !tapped {
		t.Error("the Relic was not tapped")
	}
}

func TestComponentCollectorCannotTargetALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	dnCast(t, g, "Component Collector", "Creature — Homunculus", effectstest.DnaCollector)
	island := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	dnaNight(g)
	effectstest.MonSettle(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && slices.Contains(c.PickTargetCards, island) {
			t.Errorf("the land was offered as a target: %v", c.PickTargetCards)
		}
	}
}

func TestCelestusSanctifierPutsOneOfTheTopTwoInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dnCast(t, g, "Celestus Sanctifier", "Creature — Human Cleric", effectstest.DnaSanctifier)
	if !g.IsDay() {
		t.Fatal("it did not become day")
	}
	a, b := uuid.New(), uuid.New()
	me.Library.PushTop(game.Card{InstanceID: b, Name: "B", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: a, Name: "A", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	dnaNight(g)
	effectstest.MonSettle(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dnCast(t, g, "Gavony Dawnguard", "Creature — Human Soldier", effectstest.DnaDawnguard)
	small, big := uuid.New(), uuid.New()
	me.Library.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{5}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: small, Name: "Small", TypeLine: "Creature — Elf", ManaCost: "{2}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	dnaNight(g)
	effectstest.MonSettle(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	arlinn := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: arlinn, Name: "Arlinn, the Moon's Fury", TypeLine: "Legendary Planeswalker — Arlinn",
		OracleID: effectstest.DnaArlinn + "#1", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	if err := g.ActivateCatalogAbility(me.ID, arlinn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, arlinn); got != 5 {
		t.Errorf("power = %d, want 5", got)
	}
	abilities := effectstest.EffectiveAbilities(t, g, arlinn)
	for _, k := range []string{"trample", "indestructible", "haste"} {
		if !slices.Contains(abilities, k) {
			t.Errorf("missing %s: %v", k, abilities)
		}
	}
	if types := effectstest.EffectiveTypes(t, g, arlinn); !slices.Contains(types, "Creature") || !slices.Contains(types, "Planeswalker") {
		t.Errorf("types = %v, want a planeswalker creature", types)
	}
}

func TestArlinnThePacksHopeMakesWolvesAndGrantsFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	arlinn := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: arlinn, Name: "Arlinn, the Pack's Hope", TypeLine: "Legendary Planeswalker — Arlinn",
		OracleID: effectstest.DnaArlinn, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	if err := g.ActivateCatalogAbility(me.ID, arlinn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Wolf"); n != 2 {
		t.Errorf("Wolf tokens = %d, want 2", n)
	}
	// The +1 counter half is pinned in cav_g1_test.go.
	if spec, _ := effects.Lookup(effectstest.DnaArlinn); spec.Completeness != effects.CompletenessFull {
		t.Errorf("Arlinn is complete: %v", spec.Caveats)
	}
}
