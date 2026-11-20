package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glowing One — Creature — Zombie Mutant {2}{G}, 2/2:
//
//	"Deathtouch
//	 Whenever this creature deals combat damage to a player, they get
//	 four rad counters.
//	 Whenever a player mills a nonland card, you gain 1 life."
//
// #2042. The second trigger fires once per nonland card that reached a
// graveyard (the engine's EventMill is per card), whoever milled it and
// whatever made them: a spell, or the rad counters' own trigger
// (CR 728.1). A card a replacement sent to exile instead was not milled
// and fires nothing.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b30eae2a-bc1c-45a3-93ae-e1c30d26797c",
		Name:            "Glowing One",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Glowing One — they get four rad counters",
				effects.DamagedPlayerGetsRadCounters(4)),
			effects.On(game.EventMill, effects.ANonlandCardWasMilled, "Glowing One — you gain 1 life",
				effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
