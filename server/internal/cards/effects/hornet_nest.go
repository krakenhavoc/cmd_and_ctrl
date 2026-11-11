package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hornet Nest — Creature — Insect {2}{G}, 0/2 (Tarkir: Dragonstorm
// Commander):
//
//	"Defender
//	 Whenever this creature is dealt damage, create that many 1/1 green
//	 Insect creature tokens with flying and deathtouch."
//
// Barbed Servitor's damage trigger: it fires even when the damage is
// lethal (the printed ruling), and "that many" is the damage the Nest was
// dealt as one occurrence (AtBatchEnd, ThatMuchDamage), read off the
// triggering event. The tokens are made for the Nest's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "123d2b6e-114e-4767-860d-50790382ca4a",
		Name:            "Hornet Nest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			},
			Key:        "Hornet Nest — create that many 1/1 green Insect creature tokens with flying and deathtouch",
			AtBatchEnd: ThatMuchDamage(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
					return nil
				}
				return CreateToken{
					Template: TokenCard("1/1 green Insect with flying and deathtouch"),
					N:        item.Trigger.Event.Amount,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
