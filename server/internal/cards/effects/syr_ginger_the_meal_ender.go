package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Syr Ginger, the Meal Ender — Legendary Artifact Creature — Food Knight
// {2}, 3/1 (Wilds of Eldraine, #4471):
//
//	"Syr Ginger has trample, hexproof, and haste as long as an opponent
//	 controls a planeswalker.
//	 Whenever another artifact you control is put into a graveyard from
//	 the battlefield, put a +1/+1 counter on Syr Ginger and scry 1.
//	 {2}, {T}, Sacrifice Syr Ginger: You gain life equal to its power."
//
// The keywords are Goblin Tomb Raider's conditional self-grant
// (whileYouHave), read each recompute, so they come and go with the
// opposing planeswalker. The trigger is Donatello's "artifact you
// control put into a graveyard from the battlefield", read by the
// controller it left under, minus Syr Ginger itself. The life is
// Hellhole Flailer's last-known power, as Syr Ginger was sacrificed.
//
// No simplification.
func init() {
	whilePlaneswalker := whileYouHave(syrGingerOpponentHasAPlaneswalker, selfOnly)
	var grants []game.StaticAbility
	for _, kw := range []string{"trample", "hexproof", "haste"} {
		s := KeywordGrant(whilePlaneswalker, kw)
		s.Label = "Has " + kw + " as long as an opponent controls a planeswalker."
		grants = append(grants, s)
	}
	Register(Spec{
		OracleID:     "2eb16535-718f-4b88-92bc-96429d796327",
		Name:         "Syr Ginger, the Meal Ender",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Static:       grants,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventLTB},
			AppliesTo: syrGingerAnotherArtifactWentToTheYard,
			Key:       "Syr Ginger, the Meal Ender — +1/+1 counter and scry 1",
			Effect:    Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, Scry{N: 1}),
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice Syr Ginger: You gain life equal to its power.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				if !ok || info.Power <= 0 {
					return nil
				}
				return GainLife{Amount: info.Power}.Apply(ctx)
			},
		}},
	})
}

// syrGingerOpponentHasAPlaneswalker is "an opponent controls a
// planeswalker", by effective types.
func syrGingerOpponentHasAPlaneswalker(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you && c.IsPlaneswalker() {
			return true
		}
	}
	return false
}

// syrGingerAnotherArtifactWentToTheYard is "whenever another artifact
// you control is put into a graveyard from the battlefield".
func syrGingerAnotherArtifactWentToTheYard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID || !b23ArtifactPutIntoGraveyardFromBattlefield(ev, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && leftUnderControlOf(ev, c) == source.Controller
}
