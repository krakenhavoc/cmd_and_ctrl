package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aatchik, Emerald Radian — Legendary Creature — Insect Druid {3}{B}{B}{G},
// 3/3 (Aetherdrift, #4464):
//
//	"When Aatchik enters, create a 1/1 green Insect creature token for
//	 each artifact and/or creature card in your graveyard.
//	 Whenever another Insect you control dies, put a +1/+1 counter on
//	 Aatchik. Each opponent loses 1 life."
//
// The enters trigger counts as it resolves (Ceti Eel's count): a card
// that is both an artifact and a creature is one card, and a card that
// left the graveyard in response is not counted. The death trigger is
// Idol of False Gods' "another <type> you control dies", read as the
// Insect last existed on the battlefield (CR 603.10a), so an Insect token
// counts and a creature that was an Insect only through an effect does
// too. The counter goes on Aatchik only while it is still the same
// object; the life loss happens either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "405c4a7b-dbb3-48a1-9205-7dbe6e2ad363",
		Name:         "Aatchik, Emerald Radian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aatchik — create a 1/1 green Insect for each artifact and/or creature card in your graveyard",
				Do(Instruction(func(ctx *Context) error {
					n := dftABArtifactOrCreatureCardsInGraveyard(ctx.Game, ctx.Controller())
					if n == 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 green Insect"), N: n}.Apply(ctx)
				}))),
			On(game.EventLTB, dftABAnotherInsectYouControlDied,
				"Aatchik — put a +1/+1 counter on Aatchik; each opponent loses 1 life",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, EachOpponentLosesLife{N: 1})),
		},
	})
}

// dftABArtifactOrCreatureCardsInGraveyard counts the artifact and/or
// creature cards in a player's graveyard.
func dftABArtifactOrCreatureCardsInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsArtifact() || c.IsCreature() {
			n++
		}
	}
	return n
}

// dftABAnotherInsectYouControlDied is "Whenever another Insect you
// control dies": an Insect permanent the source's controller controlled,
// other than the source, put into a graveyard from the battlefield.
func dftABAnotherInsectYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := g.LookupCardForEffect(ev.CardID)
	return ok && leftUnderControlOf(ev, dead) == source.Controller && leftAsSubtype(ev, dead, source.TextSubtype("Insect"))
}
