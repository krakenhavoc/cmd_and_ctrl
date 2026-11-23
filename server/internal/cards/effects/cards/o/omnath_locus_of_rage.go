package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Omnath, Locus of Rage — Legendary Creature — Elemental,
// {3}{R}{R}{G}{G}, 5/5 (EDHREC rank 945):
//
//	"Landfall — Whenever a land you control enters, create a 5/5 red
//	 and green Elemental creature token.
//	 Whenever Omnath or another Elemental you control dies, Omnath
//	 deals 3 damage to any target."
//
// The landfall commander: every land is a 5/5, and every 5/5 that
// dies (or Omnath himself) is a Lightning Bolt with an extra point.
// Two triggers. Landfall is Tireless Provisioner's condition with a
// token. The dies trigger fires for Omnath's own death (the LTB
// harvest finds him in the graveyard and hands the ability his
// battlefield characteristics) and for any other Elemental you
// control — an Elemental as it last existed on the battlefield (CR
// 603.10a, #1679) — the tokens his own first ability makes are Elementals,
// which is the whole card — and it targets "any target", picked
// when it goes on the stack. The damage source is Omnath whether he
// is on the battlefield or in the graveyard, and he is red, so
// Torbran adds to it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1816eede-c5bd-49df-958f-a3af64cb2932",
		Name:         "Omnath, Locus of Rage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Omnath, Locus of Rage — create a 5/5 Elemental", game.LandfallPayoff{Tokens: 1}, effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 red and green Elemental"), N: 1})),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if effects.CardDied(ev, source) {
						return true
					}
					dead, ok := effects.DiedCreature(ev, g)
					return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && effects.LeftAsSubtype(ev, dead, source.TextSubtype("Elemental"))
				},
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
				Key:     "Omnath, Locus of Rage — 3 damage to any target",
				Effect:  effects.DamageToFirstTarget(3),
			},
		},
	})
}
