package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abattoir Ghoul — {3}{B} Creature — Zombie 3/2 (Shadows over Innistrad via
// Duel Decks: Blessed vs. Cursed, S82 OLD b294):
//
//	"First strike
//	 Whenever a creature dealt damage by this creature this turn dies, you
//	 gain life equal to that creature's toughness."
//
// Sengir Vampire's watcher over every creature this Ghoul damaged earlier in
// the turn; the life is the toughness the creature last had on the
// battlefield (CR 608.2h), read as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2b211adb-de44-4468-b65d-907a09aa7e9d",
		Name:            "Abattoir Ghoul",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && effects.DealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Abattoir Ghoul — gain life equal to its toughness",
				effects.Do(effects.GainLifeEqualToSubject{Stat: effects.SubjectToughness})),
		},
	})
}
