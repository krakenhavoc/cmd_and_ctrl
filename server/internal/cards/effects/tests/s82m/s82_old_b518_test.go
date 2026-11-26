package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b518_test.go — Shards of Alara, instants and sorceries, counterspells
// and bounce, Equipment, removal, planeswalkers: S82 OLD batch 518 (#5588), each
// card driven through the engine.

const (
	b518Tortoise = "a3bf84b3-b28a-4531-95d9-d2967ade670b"
	b518Ultimate = "a9d58cad-c1c2-4bd2-80f2-da9ce60801df"
	b518Punish   = "0fc56124-bdcd-4f4f-886f-0fafcf8003bc"
	b518Sigil    = "3aa3526b-bd06-417f-af29-5a9ff0844923"
	b518Razer    = "90646d77-9b35-4f87-94ee-255c9f150456"
	b518Ajani    = "f953afdf-96f7-4851-994c-1617e8f5e5c9"
)

func TestS82OldB518EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b518Tortoise: "Tortoise Formation", b518Ultimate: "Brilliant Ultimatum", b518Punish: "Punish Ignorance",
		b518Sigil: "Sigil of Distinction", b518Razer: "Realm Razer", b518Ajani: "Ajani Vengeant",
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

func TestS82OldB518TortoiseFormationGivesYourCreaturesShroudUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Tortoise Formation", "Instant", b518Tortoise, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "shroud") {
		t.Error("my creature has no shroud")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "shroud") {
		t.Error("an opponent's creature gained shroud")
	}
	later := effectstest.B12Creature(g, me.ID, "Later", "Creature — Bear", 2, 2)
	if effectstest.HasKeywordOnBattlefield(t, g, later, "shroud") {
		t.Error("a creature that arrived after resolution gained shroud")
	}
	effectstest.DcCleanup(g)
	if effectstest.HasKeywordOnBattlefield(t, g, mine, "shroud") {
		t.Error("shroud outlived the turn")
	}
}

func TestS82OldB518PunishIgnoranceCountersAndSwingsSixLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	myLife, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Punish Ignorance", "Instant", b518Punish, effectstest.B16TargetCard(bolt))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	if me.Life != myLife+3 || opp.Life != oppLife-3 {
		t.Errorf("life: me %d → %d (want +3), opponent %d → %d (want −3)", myLife, me.Life, oppLife, opp.Life)
	}
}

func TestS82OldB518SigilOfDistinctionEntersWithXCountersAndEquipCostsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sigil := effectstest.B12PlayFromHand(t, g, "Sigil of Distinction", "Artifact — Equipment", b518Sigil, game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, sigil, game.CounterCharge); got != 2 {
		t.Fatalf("charge counters = %d, want 2", got)
	}
	effectstest.RavB2MustActivate(t, g, me.ID, sigil, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, sigil, game.CounterCharge); got != 1 {
		t.Fatalf("charge counters after equipping = %d, want 1", got)
	}
	if c := effectstest.LayeredCard(t, g, bear); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("equipped creature is %d/%d, want 3/3 (one counter left)", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestS82OldB518SigilOfDistinctionCantEquipWithNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sigil := effectstest.PushCatalogPermanent(g, me.ID, "Sigil of Distinction", "Artifact — Equipment", b518Sigil, false)
	if err := g.ActivateCatalogAbility(me.ID, sigil, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("equipped with no charge counter to remove")
	}
}

func TestS82OldB518RealmRazerExilesEveryLandAndReturnsThemTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Permanent(g, me.ID, "My Forest", "Basic Land — Forest")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	razer := effectstest.CastCatalogSpell(t, g, "Realm Razer", "Creature — Beast", b518Razer, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if z := effectstest.ZoneOfCard(g, id); z != game.ZoneExile {
			t.Fatalf("a land is in %v, want exile", z)
		}
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a nonland permanent was exiled")
	}
	effectstest.B27Kill(g, razer)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	// The lands return as new objects (CR 400.7): look them up by name.
	for name, owner := range map[string]uuid.UUID{"My Forest": me.ID, "Their Island": opp.ID} {
		found := false
		for _, c := range g.Battlefield.Cards {
			if c.Name != name {
				continue
			}
			found = true
			if !c.Tapped {
				t.Errorf("%s returned untapped", name)
			}
			if c.Controller != owner {
				t.Errorf("%s returned under %v, want its owner", name, c.Controller)
			}
		}
		if !found {
			t.Errorf("%s did not return", name)
		}
	}
}

func TestS82OldB518AjaniVengeantPlusOneKeepsAPermanentTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ajani := effectstest.GuffbWalker(g, me.ID, "Ajani Vengeant", "Legendary Planeswalker — Ajani", b518Ajani, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ArnTap(t, g, bear)
	effectstest.B16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, ajani, game.CounterLoyalty); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if !effectstest.ArnTapped(t, g, bear) {
		t.Error("the Bear untapped in its controller's next untap step")
	}
}

func TestS82OldB518AjaniVengeantMinusTwoDamagesAndGains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ajani := effectstest.GuffbWalker(g, me.ID, "Ajani Vengeant", "Legendary Planeswalker — Ajani", b518Ajani, 3)
	life, oppLife := me.Life, opp.Life
	effectstest.B16Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 || me.Life != life+3 {
		t.Errorf("life: me %d → %d (want +3), opponent %d → %d (want −3)", life, me.Life, oppLife, opp.Life)
	}
	if got := effectstest.CounterOn(g, ajani, game.CounterLoyalty); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

func TestS82OldB518AjaniVengeantMinusSevenDestroysOnlyTheTargetsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ajani := effectstest.GuffbWalker(g, me.ID, "Ajani Vengeant", "Legendary Planeswalker — Ajani", b518Ajani, 7)
	mine := effectstest.B12Permanent(g, me.ID, "My Forest", "Basic Land — Forest")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	theirs2 := effectstest.B12Permanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.B16Activate(t, g, me.ID, ajani, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(theirs2) {
		t.Error("one of the target player's lands survived")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(rock) {
		t.Error("a land of another player, or a nonland, was destroyed")
	}
}

func b518Seed(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

// b518Ultimatum runs Brilliant Ultimatum with the opponent splitting the land and
// the Wurm from three sorceries, and the caster taking pile `index`. It reports
// whether the land was offered and played and whether each card could then be
// cast without paying.
func b518Ultimatum(t *testing.T, index int) (landPlayed, wurmCast, fillerCast bool) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := b518Seed(me,
		game.Card{Name: "Filler One", TypeLine: "Sorcery", ManaCost: "{4}"},
		game.Card{Name: "Filler Two", TypeLine: "Sorcery", ManaCost: "{4}"},
		game.Card{Name: "Filler Three", TypeLine: "Sorcery", ManaCost: "{4}"},
		game.Card{Name: "Wurm", TypeLine: "Creature — Wurm", ManaCost: "{6}{G}", Power: 6, Toughness: 6},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	forest, wurm, filler := ids[4], ids[3], ids[0]
	effectstest.CastCatalogSpell(t, g, "Brilliant Ultimatum", "Sorcery", b518Ultimate, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids {
		if z := effectstest.ZoneOfCard(g, id); z != game.ZoneExile {
			t.Fatalf("a top card is in %v, want exile", z)
		}
	}
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	split := effectstest.LatestRevealPickFor(g, opp.ID)
	if split == nil {
		t.Fatalf("the opponent separates the piles: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(split.ID, opp.ID, []uuid.UUID{forest, wurm}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, index)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceMayCast, me.ID); c != nil {
		if err := g.ResolveMayCast(c.ID, me.ID, true); err != nil {
			t.Fatalf("ResolveMayCast: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		landPlayed = landPlayed || c.Name == "Forest"
	}
	err := g.CastSpell(me.ID, wurm, game.CastSpellParams{FromZone: "exile", Strict: true})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
		for _, c := range g.Battlefield.Cards {
			wurmCast = wurmCast || c.Name == "Wurm"
		}
	}
	err = g.CastSpell(me.ID, filler, game.CastSpellParams{FromZone: "exile", Strict: true})
	if err == nil {
		fillerCast = true
	}
	return landPlayed, wurmCast, fillerCast
}

func TestS82OldB518BrilliantUltimatumPlaysTheChosenPilesLandAndCastsItsSpellsFree(t *testing.T) {
	l0, w0, f0 := b518Ultimatum(t, 0)
	l1, w1, f1 := b518Ultimatum(t, 1)
	// One pile is the land and the Wurm, the other the three sorceries: taking
	// one plays the land and casts the Wurm free, taking the other casts a
	// sorcery free and does neither.
	switch {
	case l0 && w0 && !f0 && !l1 && !w1 && f1:
	case l1 && w1 && !f1 && !l0 && !w0 && f0:
	default:
		t.Errorf("pile 0: land=%v wurm=%v filler=%v; pile 1: land=%v wurm=%v filler=%v; want the two piles to split exactly", l0, w0, f0, l1, w1, f1)
	}
}
