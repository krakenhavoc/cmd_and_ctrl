package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brackwater Elemental — Creature — Elemental {2}{U}, 4/4 (Conflux):
//
//	"When this creature attacks or blocks, sacrifice it at the beginning of the
//	 next end step.
//	 Unearth {2}{U} ({2}{U}: Return this card from your graveyard to the
//	 battlefield. It gains haste. Exile it at the beginning of the next end step
//	 or if it would leave the battlefield. Unearth only as a sorcery.)"
//
// Mardu Blazebringer's attack-or-block trigger (one printed ability, one
// registered trigger; it triggers once however many attackers it blocks,
// CR 509.3a) over a delayed sacrifice at the beginning of the next end step
// instead of end of combat. The sacrifice happens only to the permanent that
// attacked or blocked (CR 400.7). Unearth is the shared ability: an unearthed
// Elemental is exiled at the next end step by unearth itself, whether or not
// its own sacrifice trigger got there first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8af4d61-1aed-415d-8d8e-7384194ee3e2",
		Name:         "Brackwater Elemental",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
			}, "Brackwater Elemental — sacrifice it at the beginning of the next end step",
				effects.OldB507SacrificeThisAtTheNextEndStep("Brackwater Elemental — sacrifice it")),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{2}{U}")},
	})
}
