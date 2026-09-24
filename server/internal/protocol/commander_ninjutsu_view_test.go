package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// commander_ninjutsu_view_test.go — the wire half of #1278. Nothing
// on the wire is new: stampZoneAbilities has walked the command zone
// since #1221 and ships the rows on `zone_abilities`, owner-only. What
// is new is a card whose ability functions THERE, so these pin that a
// commander-ninjutsu row reaches its owner off the command zone with
// the unblocked-attacker picker, and reaches nobody else.

// seatCommanderNinjutsuCommander puts a commander-ninjutsu card in a
// seat's command zone. The command zone is public, so every seat
// knows the card — as game.go marks every commander at game start.
func seatCommanderNinjutsuCommander(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	knownBy := map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		knownBy[s.ID] = true
	}
	p.Command.PushTop(game.Card{
		KnownBy:     knownBy,
		InstanceID:  id,
		Name:        "Yuriko, the Tiger's Shadow",
		TypeLine:    "Legendary Creature — Human Ninja",
		OracleID:    "00000000-0000-0000-0000-00000000cnj1",
		Power:       1,
		Toughness:   3,
		Owner:       p.ID,
		Controller:  p.ID,
		IsCommander: true,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Commander ninjutsu {U}{B}",
			Cost: game.AbilityCost{
				Mana: "{U}{B}",
				ReturnToHand: &game.ReturnToHandCost{
					Count: 1,
					Filter: &game.TargetSpec{
						Mode:  "permanent",
						Label: "an unblocked attacker you control",
						Zones: []game.ZoneKind{game.ZoneBattlefield},
						CardOK: func(g *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
							return c.IsCreature() && g.UnblockedAttackerForEffect(c.InstanceID)
						},
						Min: 1, Max: 1,
					},
					Label: "an unblocked attacker you control",
				},
			},
			Zones: []game.ZoneKind{game.ZoneHand, game.ZoneCommand},
		}},
	})
	return id
}

// commandCardView finds a card in a seat's command zone on a filtered
// view.
func commandCardView(t *testing.T, v GameView, seat int, id uuid.UUID) CardView {
	t.Helper()
	for _, c := range v.Seats[seat].Command.Cards {
		if c.InstanceID == id.String() {
			return c
		}
	}
	t.Fatalf("card %s not in seat %d's command zone", id, seat)
	return CardView{}
}

func TestCommanderNinjutsuRowRidesTheCommandZoneForItsOwnerOnly(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	yuriko := seatCommanderNinjutsuCommander(g, me)
	attacker := seatCreatureFor(g, me.ID, "Sneaky Rat")
	g.WithWriteLock(func() {
		g.Turn.Step = game.StepDeclareBlockers
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				g.Battlefield.Cards[i].AttackingTarget = them.ID
			}
		}
	})
	g.BumpLayerVersionForTest()
	// #1279: an attacker is unblocked only once its defending player has
	// DECLARED — here, declared no blocks beyond what is staged.
	if err := g.FinishBlocks(them.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}

	mine := FilterViewFor(ViewOfGame(g), me.ID.String())
	c := commandCardView(t, mine, 0, yuriko)
	if len(c.ZoneAbilities) != 1 {
		t.Fatalf("zone_abilities = %+v, want the one commander ninjutsu row", c.ZoneAbilities)
	}
	ab := c.ZoneAbilities[0]
	if ab.ReturnOptions == nil || len(ab.ReturnOptions.Cards) != 1 || ab.ReturnOptions.Cards[0] != attacker.String() {
		t.Errorf("return_options = %+v, want just the unblocked attacker %s", ab.ReturnOptions, attacker)
	}

	theirs := FilterViewFor(ViewOfGame(g), them.ID.String())
	if rows := commandCardView(t, theirs, 0, yuriko).ZoneAbilities; len(rows) != 0 {
		t.Errorf("another seat sees %d ability rows on my commander, want none — the rows are the owner's (CR 108.4)", len(rows))
	}
}
