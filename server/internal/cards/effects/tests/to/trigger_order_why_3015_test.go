package teststo

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_order_why_3015_test.go — #3015: the CR 603.3b prompt says why
// it asked. Each board below is one the independence check already
// asks on (trigger_order_reason_3013_test.go and the tests it points
// at); the prompt's sentence is the reason said in the player's terms,
// naming the card. The wording of every reason, including the ones a
// board here does not reach, is pinned in game/trigger_order_why_test.go.

func wantWhy(t *testing.T, ch *game.PendingChoice, want string) {
	t.Helper()
	if ch == nil {
		t.Fatal("no trigger-order prompt")
	}
	if ch.TriggerOrderWhy != want {
		t.Errorf("the prompt says\n got %q\nwant %q", ch.TriggerOrderWhy, want)
	}
}

func TestTheOrderPromptSaysWhyOnARealBoard(t *testing.T) {
	t.Run("reactor: a lifegain payoff watches the gain", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		marwynAndRiders(g, me.ID)
		effectstest.PushCatalogPermanent(g, me.ID, "Ajani's Pridemate", "Creature — Cat Soldier", effectstest.B24AjanisPridemateOracle, false)
		effectstest.B20CastCreature(t, g, me, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
		wantWhy(t, effectstest.SettleUntilOrderPrompt(t, g), "Ajani's Pridemate watches life totals.")
	})
	t.Run("reactor: a loss watcher beside a gain-only batch is named by family, not direction", func(t *testing.T) {
		// Exquisite Blood only cares about life LOST, and the batch
		// (Riders' gain, Marwyn's counter) only gains. The declared
		// watch is the event kind, which does not say which, so the
		// sentence must not say "gain" or "loss" about the watcher.
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		marwynAndRiders(g, me.ID)
		effectstest.PushCatalogPermanent(g, me.ID, "Exquisite Blood", "Enchantment", effectstest.B04ExquisiteBloodOracle, false)
		effectstest.B20CastCreature(t, g, me, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
		ch := effectstest.SettleUntilOrderPrompt(t, g)
		wantWhy(t, ch, "Exquisite Blood watches life totals.")
		if strings.Contains(ch.TriggerOrderWhy, "gain") || strings.Contains(ch.TriggerOrderWhy, "loss") {
			t.Errorf("a watcher's direction is not known, but the sentence claims one: %q", ch.TriggerOrderWhy)
		}
	})
	t.Run("reactor: a token the batch makes watches entering", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.PushCatalogPermanent(g, me.ID, "Evolve Token Probe", "Enchantment", effectstest.EvolveTokenProbe, false)
		effectstest.PushCatalogPermanent(g, me.ID, "Independence Token Probe", "Enchantment", effectstest.IndependenceTokenProbe, false)
		effectstest.CastNoncreature(t, g)
		wantWhy(t, effectstest.SettleUntilOrderPrompt(t, g), "A token these triggers create watches permanents entering the battlefield.")
	})
	t.Run("reactor: a draw watcher", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Chasm Skulker", "Creature — Squid Horror", chasmSkulkerOracle, false)
		ch := enterTwo(t, g, "Llanowar Visionary", "Creature — Elf Druid Shaman", llanowarVisionaryOracle,
			"Cauldron Familiar", "Creature — Bat", effectstest.CauldronFamiliarOracle)
		wantWhy(t, ch, "Chasm Skulker watches cards being drawn.")
	})
	t.Run("undescribed: a trigger the game cannot read", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Independence Token Probe", "Enchantment", effectstest.IndependenceTokenProbe, false)
		effectstest.PushCatalogPermanent(g, me.ID, "Independence Counting Probe", "Creature — Construct", effectstest.IndependenceCountingProbe, false)
		effectstest.CastCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
		wantWhy(t, effectstest.SettleUntilOrderPrompt(t, g),
			"Independence Counting Probe's trigger is handled by the players, so the game can't tell whether the order matters.")
	})
	t.Run("object tie: Riders reads the Elf Marwyn puts a counter on", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B21Push(g, me.ID, "Wolverine Riders", "Creature — Elf Warrior", effectstest.B27WolverineRidersOracle, 4, 4)
		effectstest.ToMain(t, g)
		marwyn := effectstest.PushCatalogPermanent(g, me.ID, "Marwyn, the Nurturer", "Legendary Creature — Elf Druid", effectstest.B10MarwynOracle, false)
		elf := effectstest.B21Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
		enterTogether(g, me.ID, marwyn, elf)
		wantWhy(t, effectstest.SettleUntilOrderPrompt(t, g), "One of these triggers changes Marwyn, which another looks at.")
	})
	t.Run("characteristic: a lord's static could read what the counter changes", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		marwynAndRiders(g, me.ID)
		effectstest.PushCatalogPermanent(g, me.ID, "Elvish Archdruid", "Creature — Elf Druid", effectstest.ElvishArchdruidOracle, false)
		effectstest.B20CastCreature(t, g, me, "Llanowar Elves", "Creature — Elf Druid", "", 1, 1)
		wantWhy(t, effectstest.SettleUntilOrderPrompt(t, g), "What one of these triggers does can change Elvish Archdruid's bonus, which another trigger counts.")
	})
	t.Run("characteristic: two gains beside a life-keyed anthem", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B12Push(g, me.ID, "Righteous Valkyrie", "Creature — Angel Cleric", effectstest.B18RighteousValkyrieOracle, 2, 4)
		effectstest.B12Push(g, me.ID, "Trostani, Selesnya's Voice", "Legendary Creature — Dryad", effectstest.TrostaniOracle, 2, 5)
		effectstest.ToMain(t, g)
		angel := effectstest.B21Push(g, me.ID, "An Angel", "Creature — Angel", "", 2, 2)
		enterTogether(g, me.ID, angel)
		wantWhy(t, effectstest.SettleUntilOrderPrompt(t, g),
			"The life gain from one trigger can change Righteous Valkyrie's bonus, which another trigger counts.")
	})
	t.Run("characteristic: copies of one trigger (the identical-batch case)", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B12Push(g, me.ID, "Righteous Valkyrie", "Creature — Angel Cleric", effectstest.B18RighteousValkyrieOracle, 2, 4)
		effectstest.ToMain(t, g)
		a := effectstest.B21Push(g, me.ID, "Small Angel", "Creature — Angel", "", 1, 1)
		b := effectstest.B21Push(g, me.ID, "Big Angel", "Creature — Angel", "", 4, 4)
		enterTogether(g, me.ID, a, b)
		wantWhy(t, effectstest.SettleUntilOrderPrompt(t, g),
			"The first life gain can change Righteous Valkyrie's bonus, which another trigger counts.")
	})
	t.Run("library: a draw beside a scry", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		ch := enterTwo(t, g, "Llanowar Visionary", "Creature — Elf Druid Shaman", llanowarVisionaryOracle,
			"Faerie Seer", "Creature — Faerie Wizard", effectstest.FaerieSeerOracle)
		wantWhy(t, ch, "These triggers both work on the same library or graveyard.")
	})
	t.Run("elimination: the damage can end the player Ugin exiles from", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushColoredVivi(g, me.ID)
		effectstest.PushCatalogWalker(g, me.ID, "Ugin, Eye of the Storms", effectstest.UginEyeOfTheStormsOracle, 7)
		victim := effectstest.PushUginVictim(g, opp.ID, "A Green Bear", []string{"G"})
		effectstest.ToMain(t, g)
		opp.Life = 1
		castColorlessSorcery(t, g)
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.PickCard(t, g, me.ID, victim)
		wantWhy(t, effectstest.TriggerOrderPromptFor(g, me.ID), opp.Name+" could leave the game partway through these triggers, which changes what the rest do.")
	})
	t.Run("rules: a pump on a card whose statics are closures", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		pushSokka(g, me.ID)
		effectstest.PushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", effectstest.FirebrandArcherOracle, false)
		effectstest.CastNoncreature(t, g)
		wantWhy(t, effectstest.SettleUntilOrderPrompt(t, g), "Sokka's trigger changes the rules the other triggers resolve under.")
	})
	t.Run("always: the seat asked for every prompt", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if err := g.SetTriggerOrderPreference(me.ID, game.TriggerOrderAlways); err != nil {
			t.Fatal(err)
		}
		pushMentor(g, me.ID)
		effectstest.CastNoncreature(t, g)
		wantWhy(t, effectstest.SettleUntilOrderPrompt(t, g), "You've set your triggers to ask every time.")
	})
}
