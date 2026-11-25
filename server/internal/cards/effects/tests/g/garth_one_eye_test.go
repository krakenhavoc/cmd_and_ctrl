package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// garth_one_eye_test.go — #6155: "{T}: Choose a card name that hasn't been
// chosen from among Disenchant, Braingeyser, Terror, Shivan Dragon,
// Regrowth, and Black Lotus. Create a copy of the card with the chosen
// name. You may cast the copy." (CR 707.13, 707.12, 707.10a).

const oracleGarthOneEye = "d8c4312e-03ae-43f3-b92c-4adb6875534c"

func garthTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	garth := effectstest.PushCatalogPermanent(g, me.ID, "Garth One-Eye", "Legendary Creature — Human Wizard", oracleGarthOneEye, false)
	return g, me, garth
}

// garthActivate untaps Garth, activates it and resolves the ability up to
// its name pick, returning the labels offered.
func garthActivate(t *testing.T, g *game.Game, me *game.Player, garth uuid.UUID) []string {
	t.Helper()
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == garth {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, garth, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Garth: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		return nil
	}
	var labels []string
	for _, o := range pick.PickOptions {
		labels = append(labels, o.Label)
	}
	return labels
}

// garthOfferedCopy is the one copy on offer in the open resolution_cast
// prompt, with its name.
func garthOfferedCopy(t *testing.T, g *game.Game, me *game.Player) (uuid.UUID, string) {
	t.Helper()
	c := effectstest.ResolutionCastOfferFor(g, me.ID)
	if c == nil || len(c.DiscardOptions) != 1 {
		t.Fatalf("no single-copy resolution_cast offer: %+v", c)
	}
	if !g.ResolutionPaused() {
		t.Fatal("the offer is not inside the resolution (CR 707.12)")
	}
	id := c.DiscardOptions[0]
	card, zone, ok := g.ResolutionCastCandidateLocked(me.ID, id)
	if !ok || zone != game.ZoneOutside || !card.CardCopy {
		t.Fatalf("the offered copy is %+v in %q; want a copy of a card outside the game", card.Name, zone)
	}
	return id, card.Name
}

// TestGarthOneEyeCastsABlackLotusCopy — choosing Black Lotus creates a
// copy outside the game, cast while the ability resolves; a copy of an
// artifact spell becomes a token as it resolves (CR 707.10f).
func TestGarthOneEyeCastsABlackLotusCopy(t *testing.T) {
	g, me, garth := garthTable(t)
	labels := garthActivate(t, g, me, garth)
	if len(labels) != 6 {
		t.Fatalf("the first activation offers %v, want all six names", labels)
	}
	effectstest.BlbSTPickOption(t, g, me.ID, "Black Lotus")
	copyID, name := garthOfferedCopy(t, g, me)
	if name != "Black Lotus" {
		t.Fatalf("the copy is %q", name)
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, copyID, game.CastSpellParams{FromZone: string(game.ZoneOutside)}); err != nil {
		t.Fatalf("cast the copy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var lotus *game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Black Lotus" {
			c := c
			lotus = &c
		}
	}
	if lotus == nil || !lotus.IsToken() || lotus.Controller != me.ID {
		t.Fatalf("no Black Lotus token under my control: %+v", lotus)
	}
	if n := len(me.SideboardCards()); n != 0 {
		t.Errorf("%d cards left outside the game", n)
	}
}

// TestGarthOneEyeOffersOnlyUnchosenNames — the second activation offers the
// five names left; declining the cast leaves no copy behind (CR 707.10a).
func TestGarthOneEyeOffersOnlyUnchosenNames(t *testing.T) {
	g, me, garth := garthTable(t)
	garthActivate(t, g, me, garth)
	effectstest.BlbSTPickOption(t, g, me.ID, "Black Lotus")
	garthOfferedCopy(t, g, me)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(me.SideboardCards()); n != 0 {
		t.Fatalf("a declined copy is still outside the game (%d cards)", n)
	}

	labels := garthActivate(t, g, me, garth)
	if len(labels) != 5 {
		t.Fatalf("the second activation offers %v, want five names", labels)
	}
	for _, l := range labels {
		if l == "Black Lotus" {
			t.Fatal("Black Lotus was offered again")
		}
	}
	effectstest.BlbSTPickOption(t, g, me.ID, "Terror")
	if _, name := garthOfferedCopy(t, g, me); name != "Terror" {
		t.Fatalf("the copy is %q", name)
	}
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.StackMeta) != 0 || len(me.SideboardCards()) != 0 {
		t.Error("a declined Terror copy was cast or stayed")
	}
}

// TestGarthOneEyeAllSixChosenDoesNothing — with every name chosen there is
// nothing to choose.
func TestGarthOneEyeAllSixChosenDoesNothing(t *testing.T) {
	g, me, garth := garthTable(t)
	g.WithWriteLock(func() {
		for _, n := range []string{"Disenchant", "Braingeyser", "Terror", "Shivan Dragon", "Regrowth", "Black Lotus"} {
			g.AddNoteForEffect(garth, n)
		}
	})
	if labels := garthActivate(t, g, me, garth); labels != nil {
		t.Fatalf("offered %v with every name chosen", labels)
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil || len(me.SideboardCards()) != 0 {
		t.Error("a copy was made with every name chosen")
	}
}

// TestGarthOneEyeShivanDragonCopyPaysItsCosts — "You still pay its costs":
// with no mana the Shivan Dragon copy cannot be cast, and a pool that pays
// {4}{R}{R} casts it as a creature spell that resolves to a token.
func TestGarthOneEyeShivanDragonCopyPaysItsCosts(t *testing.T) {
	g, me, garth := garthTable(t)
	garthActivate(t, g, me, garth)
	effectstest.BlbSTPickOption(t, g, me.ID, "Shivan Dragon")
	copyID, _ := garthOfferedCopy(t, g, me)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, copyID, game.CastSpellParams{FromZone: string(game.ZoneOutside)}); err == nil {
		t.Fatal("the Shivan Dragon copy was cast with no mana")
	}
	effectstest.DcMana(t, g, me.ID, "{R}{R}{C}{C}{C}{C}")
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, copyID, game.CastSpellParams{FromZone: string(game.ZoneOutside)}); err != nil {
		t.Fatalf("cast the paid copy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Shivan Dragon" && c.IsToken() {
			found = true
		}
	}
	if !found {
		t.Error("no Shivan Dragon token")
	}
}

// TestGarthOneEyeBraingeyserCopyDrawsX — the Braingeyser copy is the
// catalog's Braingeyser: cast with X = 2 at its own cost, target player
// draws two.
func TestGarthOneEyeBraingeyserCopyDrawsX(t *testing.T) {
	g, me, garth := garthTable(t)
	opp := g.Seats[1]
	for i := 0; i < 3; i++ {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	}
	hand := opp.Hand.Size()
	garthActivate(t, g, me, garth)
	effectstest.BlbSTPickOption(t, g, me.ID, "Braingeyser")
	copyID, _ := garthOfferedCopy(t, g, me)
	effectstest.DcMana(t, g, me.ID, "{U}{U}{C}{C}")
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, copyID, game.CastSpellParams{
		FromZone: string(game.ZoneOutside),
		XValue:   2,
		Targets:  []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast the Braingeyser copy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 2 {
		t.Errorf("the target drew %d, want 2", got)
	}
}
