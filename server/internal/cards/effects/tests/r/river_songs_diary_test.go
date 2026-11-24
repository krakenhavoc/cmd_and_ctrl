package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const riverSongsDiaryOracle = "60a63eb4-b0f7-4930-9bf6-afd42261d4d0"

func TestRiverSongsDiaryIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(riverSongsDiaryOracle)
	if !ok || spec.Name != "River Song's Diary" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// An instant cast from hand is exiled as it resolves, linked to the Diary.
func TestRiverSongsDiaryExilesAnInstantCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	diary := effectstest.PushCatalogPermanent(g, opp.ID, "River Song's Diary", "Artifact — Book", riverSongsDiaryOracle, false)
	spell := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.Cmdr2S41InExile(g, spell)
	if c == nil {
		t.Fatal("the resolved instant was not exiled")
	}
	if c.ExiledWith.ID != diary {
		t.Errorf("the exiled spell is linked to %v, want the Diary", c.ExiledWith.ID)
	}
	if effectstest.InGraveyard(me, spell) {
		t.Error("the spell is also in the graveyard")
	}
}

// A sorcery cast from a graveyard is not cast from a hand: it goes to the
// graveyard as usual.
func TestRiverSongsDiaryIgnoresASpellNotCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "River Song's Diary", "Artifact — Book", riverSongsDiaryOracle, false)
	id := effectstest.GraveyardCardOf(g, me, "Grave Sorcery", "Sorcery", "{R}", "")
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{Player: me.ID, Zone: game.ZoneGraveyard})
	})
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InGraveyard(me, id) {
		t.Fatalf("the sorcery cast from a graveyard is in %q, want the graveyard", effectstest.E2Zone(g, id))
	}
}

// riverSongsDiaryLinked exiles n sorcery cards of `p`'s linked to the Diary.
func riverSongsDiaryLinked(t *testing.T, g *game.Game, p *game.Player, diary uuid.UUID, n int) []uuid.UUID {
	t.Helper()
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		var epoch int
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == diary {
				epoch = c.ObjectEpoch
			}
		}
		for i := 0; i < n; i++ {
			c := game.NewCard("Linked Sorcery", p.ID)
			c.TypeLine, c.ManaCost = "Sorcery", "{5}{R}"
			p.Graveyard.PushTop(c)
			if err := g.ExileLinkedForEffect(game.ObjectRef{ID: diary, Epoch: epoch}, c.InstanceID); err != nil {
				t.Fatalf("exile: %v", err)
			}
			ids = append(ids, c.InstanceID)
		}
	})
	return ids
}

// diaryUpkeepOffer plays to the controller's next upkeep and passes until
// a resolution_cast offer opens, or gives up.
func diaryUpkeepOffer(t *testing.T, g *game.Game, me *game.Player, seat int) *game.PendingChoice {
	t.Helper()
	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	for i := 0; i < 8 && effectstest.ResolutionCastOfferFor(g, me.ID) == nil; i++ {
		if g.Turn.Step != game.StepUpkeep {
			break
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	return effectstest.ResolutionCastOfferFor(g, me.ID)
}

// With four cards exiled with the Diary, its controller's upkeep offers a
// free cast of one of them, chosen at random.
func TestRiverSongsDiaryUpkeepCastsOneOfFourExiledCardsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	diary := effectstest.PushCatalogPermanent(g, me.ID, "River Song's Diary", "Artifact — Book", riverSongsDiaryOracle, false)
	ids := riverSongsDiaryLinked(t, g, me, diary, 4)
	if diaryUpkeepOffer(t, g, me, seat) == nil {
		t.Fatal("no free cast was offered with four cards exiled with the Diary")
	}
	cast := uuid.Nil
	for _, id := range ids {
		if effectstest.CastThroughResolutionCast(t, g, me.ID, id, game.CastSpellParams{Strict: true}) == nil {
			cast = id
			break
		}
	}
	if cast == uuid.Nil {
		t.Fatal("none of the exiled cards could be cast free through the offer")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.Cmdr2S41InExile(g, cast) != nil {
		t.Error("the cast card is still in exile")
	}
}

// With only three, the upkeep trigger does not trigger.
func TestRiverSongsDiaryUpkeepNeedsFourExiledCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	diary := effectstest.PushCatalogPermanent(g, me.ID, "River Song's Diary", "Artifact — Book", riverSongsDiaryOracle, false)
	riverSongsDiaryLinked(t, g, me, diary, 3)
	if diaryUpkeepOffer(t, g, me, seat) != nil {
		t.Fatal("a free cast was offered with three cards exiled with the Diary")
	}
}
