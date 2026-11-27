package game

// assigns_as_unblocked_mod.go — "<affected> gain 'You may have this creature
// assign its combat damage as though it weren't blocked'" until end of turn
// (Garruk, Savage Herald's -7, CR 510.1c). The same flag Pride of Lions'
// static sets (Characteristic.AssignsAsThoughUnblocked), written by a
// layer-6 scoped effect so it has a duration.
//
// The kind is an on-disk identity: written into snapshots by name, never
// renamed (additive within the schema version).
const ModAssignsAsThoughUnblocked ModKind = "assignsAsThoughUnblocked"

// AssignsAsThoughUnblockedMod is that grant (layer 6, CR 613.1f).
func AssignsAsThoughUnblockedMod() Mod { return Mod{Kind: ModAssignsAsThoughUnblocked} }

// ModMustAssignAsThoughUnblocked is "<affected> assign their combat damage
// this turn as though they weren't blocked" with no "you may" (Outmaneuver,
// CR 510.1c): Characteristic.MustAssignAsThoughUnblocked, so a blocked
// creature assigns all of its damage to what it attacks and is never asked.
// An on-disk identity: never renamed.
const ModMustAssignAsThoughUnblocked ModKind = "mustAssignAsThoughUnblocked"

// MustAssignAsThoughUnblockedMod is that grant (layer 6, CR 613.1f).
func MustAssignAsThoughUnblockedMod() Mod { return Mod{Kind: ModMustAssignAsThoughUnblocked} }
