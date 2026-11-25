package cardsh

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Haktos the Unscarred — Legendary Creature — Human Warrior {R}{R}{W}{W},
// 6/1 (Theros Beyond Death, #5514):
//
//	"Haktos attacks each combat if able.
//	 As Haktos enters, choose 2, 3, or 4 at random.
//	 Haktos has protection from each mana value other than the chosen number."
//
// The choice is a random selection among three fixed ids, one per number,
// through the engine's keyed pick stream (ChooseAtRandomForEffect), made as
// Haktos enters (CR 614.12) and stored in Card.ChosenNumber. "Protection from
// each mana value other than N" is CR 702.16i shorthand for one quality per
// value; over the non-negative integers it is exactly protection from mana
// value N-1 or less and from mana value N+1 or greater, both qualities the
// grammar already has (CR 702.16a over CR 202.3). Read off the stored number
// every layer pass, so it is gone when Haktos is, and a copy chooses anew.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "932a86e5-9640-4ef5-8264-2ce97bde33e2",
		Name:         "Haktos the Unscarred",
		Completeness: effects.CompletenessFull,
		AsEnters:     haktosChooseAtRandom,
		Static: []game.StaticAbility{
			effects.AttacksEachCombat(),
			{
				Label: "Haktos has protection from each mana value other than the chosen number.",
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source != nil && target.InstanceID == source.InstanceID && source.ChosenNumber > 0
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					n := source.ChosenNumber
					effects.AppendKeywordsTo(c, []string{
						fmt.Sprintf("protection from mana value %d or less", n-1),
						fmt.Sprintf("protection from mana value %d or greater", n+1),
					})
				},
			},
		},
	})
}

// haktosNumbers are the three numbers Haktos may choose, each named by a
// fixed id so the engine's random pick (which selects ids) can choose one.
var haktosNumbers = map[uuid.UUID]int{
	uuid.NewSHA1(uuid.NameSpaceOID, []byte("haktos-the-unscarred/2")): 2,
	uuid.NewSHA1(uuid.NameSpaceOID, []byte("haktos-the-unscarred/3")): 3,
	uuid.NewSHA1(uuid.NameSpaceOID, []byte("haktos-the-unscarred/4")): 4,
}

func haktosChooseAtRandom(card *game.Card, ctx *effects.Context) error {
	ids := []uuid.UUID{
		uuid.NewSHA1(uuid.NameSpaceOID, []byte("haktos-the-unscarred/2")),
		uuid.NewSHA1(uuid.NameSpaceOID, []byte("haktos-the-unscarred/3")),
		uuid.NewSHA1(uuid.NameSpaceOID, []byte("haktos-the-unscarred/4")),
	}
	pick := ctx.Game.ChooseAtRandomForEffect(game.RandomDraw{Player: card.Controller, Source: card.InstanceID}, ids, 1)
	if len(pick) == 1 {
		ctx.Game.SetChosenNumberForEffect(card.InstanceID, haktosNumbers[pick[0]])
	}
	return nil
}
