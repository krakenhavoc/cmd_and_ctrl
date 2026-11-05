package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kang, Suspicious Captain — Legendary Creature — Klingon Warrior Officer
// {3}{R}, 4/4 (Star Trek Commander, #4974):
//
//	"When Kang enters, choose target creature an opponent controls. Kang
//	 and that creature are goaded for as long as you control Kang. (They
//	 attack each combat if able and attack a player other than you if
//	 able.)
//	 Whenever one or more goaded creatures attack, put a +1/+1 counter on
//	 target creature you control."
//
// The goad is a continuous effect from the resolved trigger (Hot
// Pursuit's ModGoad record, #2733), not the act of goading: it has no
// "until your next turn" end (CR 701.15a), only "for as long as you
// control Kang" (CR 611.2b), so it ends the moment Kang leaves or changes
// control, for both creatures. Both are goaded by you (CR 701.15b): Kang
// attacks each combat if able, and the stolen-from opponent's creature
// attacks one of your opponents if able. The duration never starts if
// Kang is gone or not yours as the trigger resolves, and an illegal
// target means the ability does nothing at all (CR 608.2b), Kang's own
// goad included.
//
// The second trigger watches every player's attack, not only yours: one
// trigger per declaration however many goaded creatures attack
// (OncePerBatch), and any goad counts — a spell's, an Aura's, Kang's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ce8bc83-3cda-4ac5-ad9a-2a2a327c5193",
		Name:         "Kang, Suspicious Captain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Kang, Suspicious Captain — Kang and target creature an opponent controls are goaded while you control Kang", kangGoadsBoth),
				TargetCreature("target creature an opponent controls", OpponentControls()),
			),
			OncePerBatch(Targeting(
				On(game.EventAttack, trekB37GoadedCreatureAttacked,
					"Kang, Suspicious Captain — put a +1/+1 counter on target creature you control",
					PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control", YouControl()),
			)),
		},
	})
}

// kangGoadsBoth goads Kang and the target, by Kang's controller, for as
// long as that player controls Kang.
func kangGoadsBoth(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	you := ctx.Controller()
	dur, ok := DurationWhileYouControlSource(ctx, ctx.Source(), you)
	if !ok {
		return nil
	}
	for _, id := range []uuid.UUID{ctx.Source(), target} {
		if err := (ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.GoadMod(you)},
			Duration: dur,
			Label:    "Kang, Suspicious Captain — goaded for as long as you control Kang",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
