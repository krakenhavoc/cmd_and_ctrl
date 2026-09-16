package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voracious Hydra — Creature — Hydra {X}{G}{G}, 0/1 (EDHREC rank
// 2734):
//
//	"Trample
//	 This creature enters with X +1/+1 counters on it.
//	 When this creature enters, choose one —
//	 • Double the number of +1/+1 counters on this creature.
//	 • This creature fights target creature you don't control."
//
// Green's X-cost removal spell that leaves a body behind. Trample
// rides PrintedKeywords; the X counters go on as the spell resolves
// (Goldvein Hydra's posture, declared below), so the ETB trigger
// sees a Hydra that already has them.
//
// The modal ETB has no mode prompt of its own — a triggered ability
// cannot carry a "choose one" — so the choice is made through the
// trigger's target prompt instead: the trigger targets "up to one
// creature you don't control", and picking a creature is the fight
// mode while picking none is the doubling mode
// (b25DoubleCountersOrFightChosen). It is declared TWICE for the
// reason Hazel's Brewmaster gives: the engine drops a targeted
// trigger whose legal set is empty, which would drop the doubling
// with it, so the targeted entry fires only while an opponent
// controls a creature and the untargeted doubling-only entry fires
// only when none does. Exactly one applies to any entry, and exactly
// one mode ever happens. A fight target that became illegal in
// response counters the trigger (CR 608.2b) — no fight AND no
// doubling, which is what the printed card does when its chosen
// mode's target is gone. The fight is b10Fight (both amounts read
// before either lands); the doubling goes through AddCounter, so a
// counter doubler applies to it, as printed.
//
// Two declared simplifications, both weaker than printed:
//
//   - The X +1/+1 counters are placed as the spell resolves, a beat
//     before the card enters, so a "whenever you put counters on a
//     permanent" payoff does not see them (Goldvein Hydra's gap).
//   - The mode is chosen through the target prompt rather than a
//     mode picker: choose a creature to fight it, choose none to
//     double the counters.
func init() {
	Register(Spec{
		OracleID:     "ff8f5a4b-112a-425e-b489-7ee26d1d9fb3",
		Name:         "Voracious Hydra",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The X +1/+1 counters are put on the Hydra as the spell resolves, a beat before it enters, so effects that watch you put counters on a permanent don't see them.",
			"The mode is chosen through the target prompt: pick a creature you don't control to fight it, or pick no target to double the Hydra's +1/+1 counters instead.",
		},
		PrintedKeywords: []string{"trample"},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: ctx.X()}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return b06SelfETB(ev, source, lki, g) && b25OpponentControlsACreature(g, source.Controller)
				},
				Targets: TargetCreature("up to one target creature you don't control to fight (none: double the counters)", OpponentControls()).WithCount(0, 1),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, b25VoraciousHydraLabel, b25DoubleCountersOrFightChosen)
				},
			},
			On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return b06SelfETB(ev, source, lki, g) && !b25OpponentControlsACreature(g, source.Controller)
			}, b25VoraciousHydraLabel, b25DoubleCountersOrFightChosen),
		},
	})
}
