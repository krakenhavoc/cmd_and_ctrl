package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stensian Sanguinist // Exsanguinate — Creature — Vampire Cleric {1}{B},
// 2/2, with the prepare spell Exsanguinate, Sorcery {X}{B}{B} (Secrets of
// Strixhaven Commander, #4834):
//
//	"Whenever you attack, target creature gains deathtouch until end of
//	 turn. Whenever that creature deals combat damage to a player this
//	 combat, this creature becomes prepared. (While it's prepared, you may
//	 cast a copy of its spell. Doing so unprepares it.)"
//
//	Exsanguinate: "Each opponent loses X life. You gain life equal to the
//	 life lost this way."
//
// A preparation card (CR 722): the permanent under the bare oracle ID and
// its prepare spell under "#1". "Whenever you attack" is one trigger per
// declaration however many creatures attack (CR 603.2c, OncePerBatch). The
// second sentence is a repeating delayed triggered ability the first
// creates as it resolves (CR 603.7a), bound to the targeted creature as the
// object it is then (CR 400.7) and ending with the combat phase
// (whenThatCreatureHitsAPlayer, as Love on the Battlefield's watch): a
// creature that hits a player in an extra combat later this turn does not
// prepare the Sanguinist. A creature that is already prepared does not
// become prepared again, and one that left the battlefield is not there to
// (BecomePrepared).
//
// The spell is Exsanguinate's body (exsanguinateResolve), cast for X with
// the copy's X chosen as it is cast.
//
// No simplification.
func init() {
	const oracle = "25877c41-39a4-4cc3-ac4b-8f3dd06d579b"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Stensian Sanguinist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(OncePerBatch(On(game.EventAttack,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				},
				"Stensian Sanguinist — target creature gains deathtouch until end of turn; this creature becomes prepared when it deals combat damage to a player this combat",
				stensianSanguinistResolve)),
				TargetCreature("target creature")),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Exsanguinate",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve:    exsanguinateResolve(),
	})
}

// stensianSanguinistBecomesPreparedBody is "this creature becomes
// prepared", the delayed trigger's body. An on-disk key: never renamed,
// never reused.
var stensianSanguinistBecomesPreparedBody = game.SimpleDelayedBody("prepare/stensian-sanguinist-becomes-prepared",
	func(g *game.Game, item *game.StackItem) error {
		return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
	})

// stensianSanguinistResolve gives the target deathtouch and watches it for
// combat damage to a player until the combat ends.
func stensianSanguinistResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (GrantKeywordUntilEOT{
		Target:   target,
		Keywords: []string{"deathtouch"},
		Label:    "Stensian Sanguinist — deathtouch until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	until, ok := DurationUntilEndOfCombat(ctx)
	if !ok {
		return nil
	}
	return whenThatCreatureHitsAPlayer(ctx, target, "Stensian Sanguinist — this creature becomes prepared",
		false, until, stensianSanguinistBecomesPreparedBody)
}
