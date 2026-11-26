package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Startled Awake // Persistent Nightmare — {2}{U}{U} Sorcery // Creature —
// Nightmare 1/1 (Shadows over Innistrad, #5543):
//
//	Front: "Target opponent mills thirteen cards.
//	        {3}{U}{U}: Put this card from your graveyard onto the battlefield
//	        transformed. Activate only as a sorcery."
//	Back:  "Skulk (This creature can't be blocked by creatures with greater
//	        power.)
//	        When this creature deals combat damage to a player, return it to
//	        its owner's hand."
//
// The spell is a sorcery front face, the mill the ordinary target-player
// mill. The graveyard ability is Garland, Knight of Cornelia's: it only
// works from the graveyard, at sorcery speed, and does nothing if the card
// has left it by the time it resolves. The back face is a new object
// (CR 400.7), and its trigger returns that object to its owner's hand if it
// is still the same permanent.
//
// No simplification.
func init() {
	const oracle = "99d42ae8-8cee-449b-812b-98a81a593077"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Startled Awake",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerMillsEffect(13)(ctx.Game, item)
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{U}{U}: Put this card from your graveyard onto the battlefield transformed. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{U}{U}"),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.PutThisFromYourGraveyardOntoTheBattlefieldTransformed,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Persistent Nightmare",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"skulk"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Persistent Nightmare — return it to its owner's hand",
				effects.ReturnThisPermanentToOwnersHand),
		},
	})
}
