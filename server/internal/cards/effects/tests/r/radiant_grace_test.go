package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/r"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// radiant_grace_test.go — Radiant Grace // Radiant Restraints (#5677):
// "return this card to the battlefield transformed under your control
// attached to target opponent". The rta* helpers are shared with
// vengeful_strangler_test.go, which also pins the engine door
// (Game.ReturnFromGraveyardTransformedAttachedForEffect).

func radiantGraceRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Radiant Grace // Radiant Restraints",
		Layout:   "transform",
		TypeLine: "Enchantment — Aura // Enchantment — Aura Curse",
		OracleID: uuid.MustParse(cardsr.RadiantGraceOracleID),
		CardFaces: []cards.CardFace{
			{Name: "Radiant Grace", TypeLine: "Enchantment — Aura", ManaCost: "{W}", Colors: []string{"W"}},
			{Name: "Radiant Restraints", TypeLine: "Enchantment — Aura Curse", Colors: []string{"W"}},
		},
	}
}

// rtaGraceOn puts Radiant Grace onto the battlefield attached to host,
// the only creature there at the time, so the CR 303.4f choice is
// forced.
func rtaGraceOn(t *testing.T, g *game.Game, p *game.Player, host uuid.UUID) uuid.UUID {
	t.Helper()
	grace := effectstest.ImportToBattlefield(t, g, radiantGraceRow(), p)
	if c := effectstest.AdCard(t, g, grace); !c.IsAttachedTo(host) {
		t.Fatalf("setup: Radiant Grace is attached to %+v", c.AttachedTo)
	}
	return grace
}

func rtaEntersTapped(t *testing.T, g *game.Game, p *game.Player) bool {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Entering Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	var err error
	g.WithWriteLock(func() { _, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{}) })
	if err != nil {
		t.Fatalf("put a creature onto the battlefield: %v", err)
	}
	return effectstest.AdCard(t, g, id).Tapped
}

// --- Radiant Grace ------------------------------------------------

// The front face: +1/+0 and vigilance. The creature dies, the trigger
// targets an opponent (never its controller), and the card comes back
// as Radiant Restraints attached to that player with nobody asked what
// it enchants.
func TestRadiantGraceReturnsAsRadiantRestraintsOnTheTargetOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.RtaBear(g, me.ID, "My Bear")
	grace := rtaGraceOn(t, g, me, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("enchanted creature gets +1/+0: power %d, want 3", got)
	}
	if c := effectstest.AdCard(t, g, bear); c.Effective().Toughness != 2 || !effectstest.HasKeywordOnBattlefield(t, g, bear, "vigilance") {
		t.Errorf("enchanted creature is 3/2 with vigilance: toughness %d", c.Effective().Toughness)
	}

	effectstest.RtaDestroy(t, g, bear)
	prompt := effectstest.RtaWaitForTarget(t, g, me.ID)
	if effectstest.HasID(prompt.PickTargetPlayers, me.ID) {
		t.Error("the trigger offers its own controller: it says target opponent")
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("targeting yourself was accepted")
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	if effectstest.AdZone(g, grace) != game.ZoneGraveyard || effectstest.TriggerOnStack(g, grace) == nil {
		t.Fatalf("the Aura is in the graveyard while its trigger waits on the stack, is in %v", effectstest.AdZone(g, grace))
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.AuraHostChoiceFor(g, me.ID) != nil {
		t.Fatal("the effect names the player: nothing is chosen as it enters")
	}
	c := effectstest.AdCard(t, g, grace)
	if effectstest.AdZone(g, grace) != game.ZoneBattlefield || c.ActiveFace != 1 || c.Name != "Radiant Restraints" {
		t.Fatalf("returned as %q face %d in %v, want Radiant Restraints on the battlefield", c.Name, c.ActiveFace, effectstest.AdZone(g, grace))
	}
	if c.AttachedTo != (game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}) {
		t.Errorf("attached to %+v, want the target opponent", c.AttachedTo)
	}
	if c.Controller != me.ID {
		t.Errorf("controller %v, want the trigger's controller", c.Controller)
	}
	if !c.HasSubtype("Curse") || c.Tapped {
		t.Errorf("an untapped Aura Curse: type %q tapped %v", c.TypeLine, c.Tapped)
	}
}

// The back face: creatures the enchanted player controls enter tapped,
// cards and tokens alike; nobody else's do.
func TestRadiantRestraintsTapsTheEnchantedPlayersEnteringCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := effectstest.RtaBear(g, me.ID, "My Bear")
	rtaGraceOn(t, g, me, bear)
	effectstest.RtaDestroy(t, g, bear)
	effectstest.RtaWaitForTarget(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if !rtaEntersTapped(t, g, opp) {
		t.Error("the enchanted player's creature entered untapped")
	}
	if rtaEntersTapped(t, g, me) {
		t.Error("the Curse's controller's creature entered tapped")
	}
	if rtaEntersTapped(t, g, other) {
		t.Error("another opponent's creature entered tapped")
	}

	var tokens []uuid.UUID
	var err error
	g.WithWriteLock(func() {
		tokens, err = g.CreateTokensForEffect(opp.ID, effects.TokenCard("1/1 green Saproling"), 1, game.TokenEntryOptions{})
	})
	if err != nil || len(tokens) != 1 {
		t.Fatalf("create a token: %v %v", tokens, err)
	}
	if !effectstest.AdCard(t, g, tokens[0]).Tapped {
		t.Error("the enchanted player's creature token entered untapped")
	}

	// A noncreature permanent is not a creature.
	land := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: land, Name: "Wastes", TypeLine: "Basic Land", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() { _, err = g.PutFromHandOntoBattlefieldForEffect(land, game.HandEntryOptions{}) })
	if err != nil {
		t.Fatal(err)
	}
	if effectstest.AdCard(t, g, land).Tapped {
		t.Error("the enchanted player's land entered tapped")
	}
}

// CR 400.7: a card that left the graveyard while the trigger waited is
// not the card the trigger returns.
func TestRadiantGraceThatLeftTheGraveyardDoesNotReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.RtaBear(g, me.ID, "My Bear")
	grace := rtaGraceOn(t, g, me, bear)
	effectstest.RtaDestroy(t, g, bear)
	effectstest.RtaWaitForTarget(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(grace); err != nil {
			t.Fatalf("exile it in response: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(grace) {
		t.Errorf("a card exiled in response came back: it is in %v", effectstest.AdZone(g, grace))
	}
}

// The trigger and its target are a restore point: the restored game
// resolves it the same way, and the Curse it lands still works.
func TestRadiantGraceTriggerSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.RtaBear(g, me.ID, "My Bear")
	grace := rtaGraceOn(t, g, me, bear)
	effectstest.RtaDestroy(t, g, bear)
	effectstest.RtaWaitForTarget(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	if effectstest.TriggerOnStack(g, grace) == nil {
		t.Fatal("setup: the trigger is not on the stack")
	}

	r := effectstest.RestoreRoundTrip(t, g, true)
	effectstest.PassPriorityAroundTable(t, r)
	c := effectstest.AdCard(t, r, grace)
	if effectstest.AdZone(r, grace) != game.ZoneBattlefield || c.Name != "Radiant Restraints" ||
		c.AttachedTo != (game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}) {
		t.Fatalf("restored game: %q in %v attached to %+v", c.Name, effectstest.AdZone(r, grace), c.AttachedTo)
	}
	// And the attached Curse is itself state a restore carries.
	r2 := effectstest.RestoreRoundTrip(t, r, true)
	if !rtaEntersTapped(t, r2, r2.Seats[1]) {
		t.Error("after a second restore the enchanted player's creature entered untapped")
	}
	if c := effectstest.AdCard(t, r2, grace); c.ActiveFace != 1 || c.AttachedTo.ID != opp.ID {
		t.Errorf("after a second restore: face %d attached to %+v", c.ActiveFace, c.AttachedTo)
	}
}
