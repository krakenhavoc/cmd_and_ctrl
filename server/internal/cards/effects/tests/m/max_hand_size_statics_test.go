package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func mhPush(g *game.Game, controller uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, controller, name, typeLine, oracle, false)
}

func mhJin(g *game.Game, controller uuid.UUID) uuid.UUID {
	return mhPush(g, controller, "Jin-Gitaxias, Core Augur", "Legendary Creature — Phyrexian Praetor", effectstest.MhJinGitaxiasOracle)
}

func mhNullProfusion(g *game.Game, controller uuid.UUID) uuid.UUID {
	return mhPush(g, controller, "Null Profusion", "Enchantment", effectstest.MhNullProfusionOracle)
}

func mhPrice(g *game.Game, controller uuid.UUID) uuid.UUID {
	return mhPush(g, controller, "Price of Knowledge", "Enchantment", effectstest.MhPriceOfKnowledgeOracle)
}

func mhSpellbook(g *game.Game, controller uuid.UUID) uuid.UUID {
	return mhPush(g, controller, "Spellbook", "Artifact — Book", effectstest.SpellbookOracle)
}

func mhTower(g *game.Game, controller uuid.UUID) uuid.UUID {
	return mhPush(g, controller, "Reliquary Tower", "Land", effectstest.ReliquaryTowerOracl)
}

// The Null Profusion ruling (2009-10-01), in the first order: Spellbook,
// then Null Profusion, is a maximum of two.
func TestSpellbookThenNullProfusionIsTwo(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mhSpellbook(g, me.ID)
	mhNullProfusion(g, me.ID)
	if got := effectstest.MhMax(g, me); got != 2 {
		t.Errorf("Spellbook, then Null Profusion: %d, want 2", got)
	}
}

// The other order: Null Profusion, then Spellbook, is no maximum.
func TestNullProfusionThenSpellbookIsNoMaximum(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mhNullProfusion(g, me.ID)
	mhSpellbook(g, me.ID)
	if got := effectstest.MhMax(g, me); got != game.NoMaxHandSize {
		t.Errorf("Null Profusion, then Spellbook: %d, want no maximum", got)
	}
}

// Jin-Gitaxias: each opponent at zero, its controller at seven (the
// 2017-11-17 rulings).
func TestJinGitaxiasReducesEachOpponentNotItsController(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mhJin(g, me.ID)
	for _, p := range g.Seats {
		want := 0
		if p.ID == me.ID {
			want = game.DefaultMaxHandSize
		}
		if got := effectstest.MhMax(g, p); got != want {
			t.Errorf("%s: %d, want %d", p.Name, got, want)
		}
	}
}

// Two Jin-Gitaxias make 7 - 7 - 7 = -7. The view reports zero (never
// the -1 that means "no maximum"), and the opponent's cleanup discards
// the whole hand (CR 107.1b, CR 514.1).
func TestTwoJinGitaxiasIsZeroInTheViewAndAtCleanup(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mhJin(g, other.ID)
	mhJin(g, g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)].ID)

	view := protocol.ViewOfGameFor(g, active.ID.String())
	for _, s := range view.Seats {
		if s.ID == active.ID.String() && s.MaxHandSize != 0 {
			t.Errorf("view max_hand_size = %d, want 0", s.MaxHandSize)
		}
	}

	effectstest.FillHandTo(t, g, active, 3)
	effectstest.RunCleanup(t, g)
	if n, hand := g.DiscardPending[active.ID], active.Hand.Size(); n != hand || hand == 0 {
		t.Errorf("cleanup discard %d of a %d-card hand, want all of it", n, hand)
	}
}

// Jin-Gitaxias, then the opponent's Reliquary Tower: no maximum.
func TestJinGitaxiasThenReliquaryTowerIsNoMaximum(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mhJin(g, me.ID)
	mhTower(g, opp.ID)
	if got := effectstest.MhMax(g, opp); got != game.NoMaxHandSize {
		t.Errorf("Jin-Gitaxias, then Reliquary Tower: %d, want no maximum", got)
	}
}

// Reliquary Tower, then Jin-Gitaxias: still no maximum, because a
// reduction leaves "no maximum" unbounded.
func TestReliquaryTowerThenJinGitaxiasIsNoMaximum(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mhTower(g, opp.ID)
	mhJin(g, me.ID)
	if got := effectstest.MhMax(g, opp); got != game.NoMaxHandSize {
		t.Errorf("Reliquary Tower, then Jin-Gitaxias: %d, want no maximum", got)
	}
}

// Price of Knowledge lifts every player's maximum, its controller's
// included, and an opponent's under an EARLIER Jin-Gitaxias.
func TestJinGitaxiasThenPriceOfKnowledgeIsNoMaximumForEveryone(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mhJin(g, me.ID)
	mhPrice(g, opp.ID)
	for _, p := range g.Seats {
		if got := effectstest.MhMax(g, p); got != game.NoMaxHandSize {
			t.Errorf("%s under Jin-Gitaxias then Price of Knowledge: %d, want no maximum", p.Name, got)
		}
	}
}

// The other order: Price of Knowledge first, then Jin-Gitaxias. "No
// maximum" is unbounded, and a reduction after it leaves it unbounded,
// so every player still has no maximum.
func TestPriceOfKnowledgeThenJinGitaxiasIsNoMaximumForEveryone(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mhPrice(g, opp.ID)
	mhJin(g, me.ID)
	for _, p := range g.Seats {
		if got := effectstest.MhMax(g, p); got != game.NoMaxHandSize {
			t.Errorf("%s under Price of Knowledge then Jin-Gitaxias: %d, want no maximum", p.Name, got)
		}
	}
}

// Finale of Revelation's grant, then a Null Profusion: two. The grant
// has the timestamp of its resolution (CR 613.7b), so a later "your
// maximum hand size is two" applies after it.
func TestFinaleOfRevelationThenNullProfusionIsTwo(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Finale of Revelation", TypeLine: "Sorcery",
		OracleID: effectstest.FinaleOfRevelationOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{XValue: 10}); err != nil {
		t.Fatalf("CastSpell Finale of Revelation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.ChooseCardsChoiceFor(g, caster.ID); pick != nil {
		effectstest.AnswerChooseCards(t, g, caster.ID)
	}
	if caster.MaxHandSize != game.NoMaxHandSize || caster.MaxHandSizeAt == 0 {
		t.Fatalf("setup: Finale's grant = %d at %d, want a stamped no maximum", caster.MaxHandSize, caster.MaxHandSizeAt)
	}
	if got := effectstest.MhMax(g, caster); got != game.NoMaxHandSize {
		t.Fatalf("setup: after Finale, %d, want no maximum", got)
	}
	mhNullProfusion(g, caster.ID)
	if got := effectstest.MhMax(g, caster); got != 2 {
		t.Errorf("Finale of Revelation, then Null Profusion: %d, want 2", got)
	}
}

// A Null Profusion that has lost all its abilities gives nothing
// (CR 613.1f): the maximum is seven again.
func TestNullProfusionThatLostItsAbilitiesGivesNothing(t *testing.T) {
	effectstest.MhTickingClock(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	np := mhNullProfusion(g, me.ID)
	if got := effectstest.MhMax(g, me); got != 2 {
		t.Fatalf("setup: Null Profusion alone: %d, want 2", got)
	}
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(np),
			[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities") {
			t.Fatal("setup: the removal registered nothing")
		}
	})
	if got := effectstest.MhMax(g, me); got != game.DefaultMaxHandSize {
		t.Errorf("Null Profusion without its abilities: %d, want 7", got)
	}
}

// Null Profusion's other two lines: the draw step is skipped, and a
// land played or a spell cast draws a card.
func TestNullProfusionDrawsWhenYouPlayACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	mhNullProfusion(g, me.ID)

	library := me.Library.Size()
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{}); err != nil {
		t.Fatalf("play Swamp: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if drawn := library - me.Library.Size(); drawn != 1 {
		t.Errorf("playing a land drew %d, want 1", drawn)
	}

	library = me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if drawn := library - me.Library.Size(); drawn != 1 {
		t.Errorf("casting a spell drew %d, want 1", drawn)
	}

	// An opponent playing a land draws nobody anything.
	oppSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	opp := g.Seats[oppSeat]
	effectstest.AdvanceToStepOf(t, g, oppSeat, game.StepPrecombatMain)
	library = me.Library.Size()
	oppLand := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: oppLand, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, oppLand, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent plays Swamp: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if drawn := library - me.Library.Size(); drawn != 0 {
		t.Errorf("an opponent's land drew %d, want 0", drawn)
	}
}

// "Skip your draw step": the controller's next turn draws nothing in
// its draw step.
func TestNullProfusionSkipsYourDrawStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mhNullProfusion(g, me.ID)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	library := me.Library.Size()
	effectstest.AdvanceToStepOf(t, g, 0, game.StepPrecombatMain)
	if drawn := library - me.Library.Size(); drawn != 0 {
		t.Errorf("drew %d in a skipped draw step, want 0", drawn)
	}
}

// Jin-Gitaxias draws seven at the beginning of its controller's end
// step.
func TestJinGitaxiasDrawsSevenAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mhJin(g, me.ID)
	library := me.Library.Size()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if drawn := library - me.Library.Size(); drawn != 7 {
		t.Errorf("drew %d at the end step, want 7", drawn)
	}
}

// Price of Knowledge's trigger: at each OPPONENT's upkeep, damage equal
// to that player's hand, counted on resolution; never at its
// controller's own upkeep.
func TestPriceOfKnowledgeDamagesEachOpponentByHandSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mhPrice(g, me.ID)

	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	hand := opp.Hand.Size()
	life := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if lost := life - opp.Life; lost != hand {
		t.Errorf("opponent with %d cards lost %d, want %d", hand, lost, hand)
	}

	// Its controller's own upkeep deals nothing.
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	life = me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("Price of Knowledge's controller lost %d at their own upkeep", life-me.Life)
	}
}
