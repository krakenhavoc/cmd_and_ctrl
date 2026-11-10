package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdrB25TimelessDragonOracle = "a5ae93e9-0385-47fd-9d8f-306e9a6b7fe6"
	cmdrB25ChampionOfWitsOracle = "3fd8ba0c-75ec-43ea-8706-f71e6cf1d236"
)

// cmdrB25Copy finds a battlefield card named `name` that `owner` controls
// and that is not `not` (the card the copy was made from).
func cmdrB25Copy(g *game.Game, owner uuid.UUID, name string, not uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == name && c.Controller == owner && c.InstanceID != not {
			return c
		}
	}
	return nil
}

// cmdrB25Eternalize eternalizes the card from the graveyard (ability index
// `idx`) paying `mana`, and returns the token.
func cmdrB25Eternalize(t *testing.T, g *game.Game, name, typeLine, oracle string, p, tough int, idx int, mana string) (*game.Card, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := pushCatalogGraveyardCard(me, name, typeLine, oracle, p, tough)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("eternalize %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Errorf("%s: the eternalized card is not in exile", name)
	}
	tok := cmdrB25Copy(g, me.ID, name, id)
	if tok == nil {
		t.Fatalf("%s: no eternalized token", name)
	}
	return tok, id
}

func TestCmdrB25TimelessDragonPlainscyclingNeedsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := pushCatalogHandCard(me, "Timeless Dragon", "Creature — Dragon", cmdrB25TimelessDragonOracle)
	if err := hobB4Activate(t, g, me.ID, id, 0, "{C}", nil); err == nil {
		t.Fatal("plainscycling {2} was paid with one mana")
	}
}

func TestCmdrB25TimelessDragonEternalizesIntoABlackFourFourZombieDragon(t *testing.T) {
	g := newCatalogGame(t)
	tok, _ := cmdrB25Eternalize(t, g, "Timeless Dragon", "Creature — Dragon", cmdrB25TimelessDragonOracle, 5, 5, 1, "{C}{C}{W}{W}")
	if tok.Power != 4 || tok.Toughness != 4 {
		t.Errorf("token %d/%d, want 4/4", tok.Power, tok.Toughness)
	}
	if len(tok.Colors) != 1 || tok.Colors[0] != "B" || tok.ManaCost != "" {
		t.Errorf("token colours %v cost %q, want black with no cost", tok.Colors, tok.ManaCost)
	}
	if tok.TypeLine != "Token Creature — Zombie Dragon" {
		t.Errorf("token type line %q", tok.TypeLine)
	}
}

// castChampion casts Champion of Wits free and resolves the spell, leaving
// the enters trigger's prompt open.
func cmdrB25CastChampion(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	if _, err := msc17Cast(t, g, me, game.Card{
		Name: "Champion of Wits", TypeLine: "Creature — Snake Wizard", OracleID: cmdrB25ChampionOfWitsOracle,
		Power: 2, Toughness: 1,
	}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Champion of Wits: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestCmdrB25ChampionOfWitsDrawsItsPowerThenDiscardsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	msc18Library(me, 8)
	cmdrB25CastChampion(t, g, me)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatal("no \"draw?\" prompt on entering")
	}
	hand := me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d after drawing its power (2), want %d", me.Hand.Size(), hand+2)
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after discarding two, want %d", me.Hand.Size(), hand)
	}
}

func TestCmdrB25ChampionOfWitsDeclineDrawsAndDiscardsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	msc18Library(me, 8)
	cmdrB25CastChampion(t, g, me)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("no prompt")
	}
	hand := me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand || discardChoiceFor(g, me.ID) != nil {
		t.Errorf("declining changed the hand (%d, want %d) or asked for a discard", me.Hand.Size(), hand)
	}
}

func TestCmdrB25ChampionOfWitsEternalizedTokenDrawsFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	msc18Library(me, 8)
	hand := me.Hand.Size()
	tok, _ := cmdrB25Eternalize(t, g, "Champion of Wits", "Creature — Snake Wizard", cmdrB25ChampionOfWitsOracle, 2, 1, 0, "{C}{C}{C}{C}{C}{U}{U}")
	if tok.Power != 4 || tok.Toughness != 4 {
		t.Errorf("token %d/%d, want 4/4", tok.Power, tok.Toughness)
	}
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("the token's enters trigger asked nothing")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand+4 {
		t.Errorf("hand %d, want %d (drew four)", me.Hand.Size(), hand+4)
	}
}
