package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shilgengar, Sire of Famine — Legendary Creature — Elder Demon {3}{B}{B}, 6/6
// (Modern Horizons 3):
//
//	"Flying
//	 Sacrifice another creature: Create a Blood token. If you sacrificed an Angel
//	 this way, create a number of Blood tokens equal to its toughness instead.
//	 {W/B}{W/B}{W/B}, Sacrifice six Blood tokens: Return each creature card from
//	 your graveyard to the battlefield with a finality counter on it. Those
//	 creatures are Vampires in addition to their other types."
//
// The Angel check and its toughness are the sacrificed creature's last-known
// information (CR 608.2h).
func init() {
	Register(Spec{
		OracleID:        "71ac4d8c-5b58-4ea8-996b-f0289717c4bb",
		Name:            "Shilgengar, Sire of Famine",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Tokens: 1},
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			{
				Label:   "Sacrifice another creature: Create a Blood token. If you sacrificed an Angel this way, create a number of Blood tokens equal to its toughness instead.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Tokens: 1},
				Cost:    SacrificeAnotherN(1, "another creature", Creature()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := 1
					if info, ok := ctx.SacrificedPermanent(); ok && gapB50SacrificedAngel(info) {
						n = max(info.Toughness, 0)
					}
					if n == 0 {
						return nil
					}
					return CreateToken{Controller: item.Controller, Template: BloodToken(), N: n}.Apply(ctx)
				},
			},
			{
				Label:   "{W/B}{W/B}{W/B}, Sacrifice six Blood tokens: Return each creature card from your graveyard to the battlefield with a finality counter on it. Those creatures are Vampires in addition to their other types.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost: Plus(ManaCost("{W/B}{W/B}{W/B}"),
					SacrificeN(6, "six Blood tokens", Subtype("Blood"), IsTokenPredicate())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					p := ctx.PlayerByID(item.Controller)
					if p == nil || p.Graveyard == nil {
						return nil
					}
					var ids []uuid.UUID
					for _, c := range p.Graveyard.Cards {
						if c.IsCreature() && !c.IsToken() {
							ids = append(ids, c.InstanceID)
						}
					}
					if len(ids) == 0 {
						return nil
					}
					return putCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, ids, game.ZoneEntryOptions{
						Controller:         item.Controller,
						EntersWithCounters: map[string]int{game.CounterFinality: 1},
						EntersAs: &game.EntersAs{
							AddTypes:    []string{"Creature"},
							AddSubtypes: ctx.TextSubtypes("Vampire"),
							Label:       "Shilgengar — a Vampire in addition to its other types",
						},
					}, game.TailRef{}, 0)
				},
			},
		},
	})
}
