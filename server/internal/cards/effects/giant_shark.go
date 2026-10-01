package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Giant Shark — Creature — Shark {5}{U}, 4/4:
//
//	"This creature can't attack unless defending player controls an
//	 Island.
//	 Whenever this creature blocks or becomes blocked by a creature that
//	 has been dealt damage this turn, this creature gets +2/+0 and gains
//	 trample until end of turn.
//	 When you control no Islands, sacrifice this creature."
//
// ADR 0107 (#1858, #1879). The attack restriction is §2's (CR 508.1c,
// 508.5, 508.5a) and the sacrifice §1's CR 603.8 state trigger; both read
// one PermanentQuery.
//
// The block trigger is per creature (CR 509.3b, 509.3d): one event per
// blocker–attacker pair, so the Shark blocked by two damaged creatures
// triggers twice and gets +4/+0. "Dealt damage this turn" is
// DealtDamageThisTurn (ADR 0107 PR 5), read off the turn's damage events
// and wiped by a re-entry (CR 400.7). It is a fact about the other
// creature as blockers are declared (CR 509.3f), not an intervening "if".
//
// No simplification.
func init() {
	q := QuerySubtype("Island")
	damaged := DealtDamageThisTurn()
	Register(Spec{
		OracleID:     "44a10a63-be9c-4f1d-aad6-b5337112bda5",
		Name:         "Giant Shark",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				var other uuid.UUID
				switch source.InstanceID {
				case ev.CardID:
					other = ev.Target
				case ev.Target:
					other = ev.CardID
				default:
					return false
				}
				c, ok := g.LookupCardForEffect(other)
				return ok && damaged(g, source.Controller, c)
			}, "Giant Shark — +2/+0 and trample until end of turn", giantSharkPump),
			WhenYouControlNo(q, "Giant Shark — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}

// giantSharkPump is the block trigger's effect on the Shark itself.
func giantSharkPump(g *game.Game, item *game.StackItem) error {
	if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: "Giant Shark — +2/+0"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"trample"}, Label: "Giant Shark — trample"}.Apply(ctx)
}
