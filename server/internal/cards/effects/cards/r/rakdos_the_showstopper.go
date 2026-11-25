package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakdos, the Showstopper — Legendary Creature — Demon {4}{B}{R}, 6/6
// (Ravnica: Clue Edition):
//
//	"Flying, trample
//	 When Rakdos enters, flip a coin for each creature that isn't a Demon,
//	 Devil, or Imp. Destroy each creature whose coin comes up tails."
//
// The creatures are counted as the ability resolves, one uncalled flip each
// (CR 705.1, FlipCoinsForEffect: nobody wins or loses), and every creature
// whose flip is tails is destroyed together. Rakdos is a Demon, so it is
// never flipped for. Regeneration and indestructible apply as for any
// destruction. A declared partial sweep (ADR 0126). Under Krark's Thumb
// the flipper keeps heads for their own creatures and tails for the rest
// (#6152).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "03a4b997-4738-41a9-933f-e1f765e3a75a",
		Name:            "Rakdos, the Showstopper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rakdos, the Showstopper — flip a coin for each creature that isn't a Demon, Devil, or Imp; destroy each creature whose coin comes up tails",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					flipped := effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.Not(effects.AnySubtype("Demon", "Devil", "Imp"))))
					if len(flipped) == 0 {
						return nil
					}
					// The face the flipper wants for each coin, read only when
					// Krark's Thumb has them flip several and keep one (#6152):
					// their own creature heads, everyone else's tails.
					prefer := make([]bool, len(flipped))
					for i, c := range flipped {
						prefer[i] = c.Controller == item.Controller
					}
					faces, err := g.FlipCoinsPreferringForEffect(game.RandomDraw{Player: item.Controller, Source: item.SourceCardID}, prefer)
					if err != nil {
						return err
					}
					tails := map[uuid.UUID]bool{}
					for i, c := range flipped {
						if !faces[i] {
							tails[c.InstanceID] = true
						}
					}
					if len(tails) == 0 {
						return nil
					}
					return effects.DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return tails[c.InstanceID]
					}}.Apply(ctx)
				}),
		},
	})
}
