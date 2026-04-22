package game

// layer_listener.go ships the built-in `Listener` that drives the S16
// continuous-effect engine's invalidation. Auto-registered by
// NewGame so every game starts with the layer system listening for
// the events that change which static abilities are active or what
// they apply to.
//
// Today (sub-PR 2) the listener bumps `g.layerVersion` on:
//   - EventZoneMove where either OldZone or NewZone is the
//     battlefield — covers ETB and LTB universally without having
//     to audit every emission site individually.
//   - EventCounterPlaced — counters change layer 7d inputs
//     (CurrentPower / CurrentToughness delegation) and Tarmogoyf-
//     style CDA inputs (graveyard-counter changes etc.).
//
// Things this listener INTENTIONALLY does NOT bump on yet:
//   - Step / phase advance. "Until end of turn" continuous effects
//     are out of scope for S16; when they arrive in a later sprint,
//     advance the version inside the step-advance helper directly
//     (no event for it today; cleanest is a direct call rather than
//     a new event kind for one consumer).
//   - Control changes. Mind Control / aura attach is deferred to
//     S17. When the first "creatures you control" predicate that
//     can flip mid-game lands, add an EventControlChanged kind +
//     bump here.
//
// EventETB and EventLTB are also covered by EventZoneMove (every
// zone change emits both), so the listener doesn't double-bump on
// those — the ZoneMove path is the single source of truth.

// layerVersionBump is the `Listener` that auto-installs into every
// game to invalidate the layer engine's cached resolution when
// events occur that could change static-ability active-set or
// applies-to inputs.
type layerVersionBump struct{}

// OnEvent fires synchronously inside the EmitEvent path under the
// game's write lock. Cheap atomic increment per relevant event;
// no allocation. Falls through silently for events that don't
// affect continuous effects.
func (layerVersionBump) OnEvent(g *Game, ev Event) {
	switch ev.Kind {
	case EventZoneMove:
		if ev.OldZone == ZoneBattlefield || ev.NewZone == ZoneBattlefield {
			g.layerVersion.Add(1)
		}
	case EventCounterPlaced:
		g.layerVersion.Add(1)
	}
}
