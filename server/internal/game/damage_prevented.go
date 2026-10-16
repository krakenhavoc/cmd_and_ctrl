package game

import "github.com/google/uuid"

// damage_prevented.go — CR 615.13 (#2047): "whenever damage is
// prevented" triggers, and the one event they watch.
//
// THE RULE. CR 615.13: "Some triggered abilities trigger when damage that
// would be dealt is prevented. Such an ability triggers each time a
// prevention effect is applied to one or more simultaneous damage events
// and prevents some or all of that damage." Selfless Squire: "Whenever
// damage that would be dealt to you is prevented, put that many +1/+1
// counters on this creature."
//
// THE UNIT is one APPLICATION of one prevention effect to the
// simultaneous damage events of one damage instance (ADR 0108 PR 0,
// damage_instance.go) — the unit the follow-up queue already groups a
// CR 615.5 additional effect by. Three attackers' damage to you,
// prevented by one Fog, is one application and one trigger for 3+3+3;
// the same damage prevented partly by a charged shield and partly by a
// "prevent all" is two effects and two triggers. A prevention effect
// that prevents nothing — one applied to damage that can't be prevented
// (CR 615.12), or a half shield on 1 damage — is not an application
// CR 615.13 counts, and announces nothing.
//
// ONE EVENT PER RECIPIENT. The announcement is keyed on the effect, the
// instance and the RECIPIENT, so an application that covered damage to
// you and to your creatures is one event for you and one for each
// creature. Every printed CR 615.13 trigger names its recipient ("dealt
// to you"), so each sees exactly one event per application; a trigger
// that names no recipient, or two, would count an application once per
// recipient, and no card prints one.
//
// HOW IT IS MEASURED. Every prevention effect — a scoped shield, a
// catalog static, protection's built-in (CR 702.16e), a shield counter
// (CR 122.1c) — is applied through runReplaceLocked, which reads the
// event's damage before and after the effect's Replace: what Replace took
// off, all of it when it cancelled the event, is what this application
// prevented (TestDamageReplacementsDeclareWhetherTheyPrevent holds every
// damage replacement that reduces the damage to declaring Prevention).
// Under "damage can't be prevented" the apply loop settles a prevention
// effect without running its Replace (settleUnpreventableLocked), so it
// is never measured and never announced: Leyline of Punishment and
// Skullcrack switch the trigger off with the prevention, as will every
// later "damage can't be prevented" (Spider-Punk, #4044).
//
// WHEN IT IS EMITTED. The announcement joins the follow-up queue as an
// entry of its own (PreventionFollowUp.Announce), owed for the instance,
// so the amounts of the instance's events add up before anything sees
// them, and it is emitted where a follow-up runs: as the instance ends,
// or at ADR 0107's flush points (as a player would next receive priority,
// before the state-based actions, and as a new event batch opens). A
// creature killed by the same combat damage is still on the battlefield
// to trigger. The queue is in the snapshot, so a table paused inside the
// instance keeps what it owes.

// announcePreventionLocked owes CR 615.13's "damage was prevented" for one
// application of the prevention effect `a` to `ev`, which prevented
// `prevented` of the `damage` the event carried as the effect met it.
// Nothing is owed when nothing was prevented.
//
// THE KEY names the effect by something the snapshot can carry: a scoped
// shield by its record (Seq) and mod (Slot); a catalog static by its
// object (Source, Epoch, CR 400.7) and slot, as its follow-up is; an
// engine built-in (protection, a shield counter) by a negative slot of
// its own. The unit is always the recipient.
//
// Caller must hold g.mu (write).
func (g *Game) announcePreventionLocked(ev *ReplacementEvent, a activeReplacement, target uuid.UUID, prevented, damage int) {
	if ev == nil || prevented <= 0 {
		return
	}
	f := g.newPreventionFollowUpLocked(ev, "", prevented, damage)
	f.Announce = true
	f.Target = target
	f.Unit = target
	f.Label = a.effect.Label
	switch seq, mod, scoped := decodeScopedReplacementID(a.id); {
	case scoped:
		f.Seq, f.Slot = seq, mod
		if i, ok := g.scopedEffectIndexBySeqLocked(seq); ok {
			e := g.ScopedEffects[i]
			f.Controller, f.Source = e.Controller, e.Source.ID
			if f.Label == "" {
				f.Label = e.Label
			}
		}
	case a.source != nil:
		f.Static = true
		f.Source, f.Epoch, f.Slot = a.source.InstanceID, a.source.ObjectEpoch, a.identity.slot
		if _, slot, ok := decodeCatalogReplacementID(a.id); ok {
			// The packed slot, which a redirecting or self-acting
			// effect's zero identity does not carry.
			f.Slot = slot
		}
		f.Controller = a.source.Controller
	default:
		f.Static = true
		f.Slot = builtinAnnounceSlot(a.id)
	}
	g.owePreventionFollowUpLocked(f)
}

// builtinAnnounceSlot is the Slot an engine-owned prevention effect's
// announcement is keyed on: below builtinFollowUpSlot, one per built-in
// (or test-injected) registration, from its ID, which is minted by index
// when the game is built and so the same after a restore. A catalog slot
// is never negative, so it never meets a static's own key.
func builtinAnnounceSlot(id ReplacementEffectID) int {
	switch {
	case id >= builtinReplacementIDBase:
		return builtinFollowUpSlot - 1 - 2*int(id-builtinReplacementIDBase)
	case id >= testReplacementIDBase:
		return builtinFollowUpSlot - 2 - 2*int(id-testReplacementIDBase)
	}
	return builtinFollowUpSlot - 1
}

// emitDamagePreventedLocked emits one owed announcement as
// EventDamagePrevented: Target the recipient the damage would have been
// dealt to, Amount the total this application prevented from it in the
// instance ("that many"), Source the first source whose damage it
// prevented, Actor and Colors that source's controller and colours as
// they were when the damage would have been dealt, Combat whether it was
// combat damage.
//
// Caller must hold g.mu (write).
func (g *Game) emitDamagePreventedLocked(f PreventionFollowUp) {
	if f.Prevented <= 0 {
		return
	}
	g.EmitEvent(Event{
		Kind:           EventDamagePrevented,
		Source:         f.DamageSource,
		Target:         f.Target,
		Amount:         f.Prevented,
		Actor:          f.SourceController,
		Colors:         copyStrings(f.SourceColors),
		Combat:         f.Combat,
		LastKnownTypes: copyStrings(f.SourceTypes),
		DamageInstance: f.Instance,
	})
}
